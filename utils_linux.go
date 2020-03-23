package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func getVethHost(spec *specs.Spec) *net.Interface {
	var netnsPath string

	if spec.Linux != nil {
		for _, v := range spec.Linux.Namespaces {
			if v.Type == specs.NetworkNamespace {
				netnsPath = v.Path
				break
			}
		}
	}
	// should look like '/var/run/netns/cni-6d46d1b2-c836-3b4e-87a0-88641242aa5e'
	if strings.Index(netnsPath, "/var/run/netns/") == -1 {
		return nil
	}

	origns, _ := netns.Get()
	defer origns.Close()

	netnsPath = strings.Replace(netnsPath, "/var/run/netns/", "", 1)
	if err != nil {
		logrus.Errorf("unable to get netns handle %s(%s)",
			netnsPath, err)
		return nil
	}

	nsh, err := netns.GetFromName(netnsPath)
	if err := netns.Set(nsh); err != nil {
		logrus.Errorf("unable to get set netns %s", err)
		return nil
	}

	// Look for the ifindex of veth pair of the host interface
	vEth, err := netlink.LinkByName("eth0")
	if err != nil {
		logrus.Errorf("unable to get guest veth %s", err)
		return nil
	}

	ifIndex, err := netlink.VethPeerIndex(&netlink.Veth{LinkAttrs: *vEth.Attrs()})
	if err != nil {
		logrus.Errorf("unable to get ifindex of veth pair %s", err)
		return nil
	}

	// Switch back to the original namespace
	netns.Set(origns)
	iface, err := net.InterfaceByIndex(ifIndex)
	if err != nil {
		logrus.Errorf("unable to get interface of veth pair (idx=%d)(%s)",
			ifIndex, err)
		return nil
	}

	return iface
}

// XXX: Only treat ipv4 information
func getVethInfo(spec *specs.Spec) (*lklIfInfo, error) {
	ifInfo := new(lklIfInfo)

	// default gateway
	v4gw, err := netlink.RouteGet(net.ParseIP("8.8.8.8"))
	if err != nil {
		return nil, fmt.Errorf("Could not determine single default route (got %v)",
			len(v4gw))
	}
	ifInfo.v4Gw = v4gw[0].Gw

	ifaces, _ := net.Interfaces()
	logrus.Debugf("ifaces= %+v", ifaces)
	for _, iface := range ifaces {
		if iface.Name != "eth0" {
			continue
		}

		allAddrs, _ := iface.Addrs()
		for _, ifaddr := range allAddrs {
			ipNet, ok := ifaddr.(*net.IPNet)
			if !ok {
				return nil, fmt.Errorf("address is not IPNet: %+v", ifaddr)
			}

			logrus.Infof("ifaddr= %s, ipnet=%s", ifaddr, ipNet)
			// XXX: We only treat IPv4 at the moment
			// may need to work for IPv6 support
			if ipNet.IP.To4() == nil {
				continue
			}

			ifInfo.ifAddrs = append(ifInfo.ifAddrs, *ipNet)
			ifInfo.ifName = "eth0"

			// Get the link for the interface.
			ifaceLink, err := netlink.LinkByName(iface.Name)
			if err != nil {
				return nil, fmt.Errorf("getting link for interface %q: %v", iface.Name, err)
			}

			// Steal IP address from NIC.
			r4addr, err := netlink.ParseAddr(ipNet.String())
			if err != nil {
				return nil, fmt.Errorf("parse address error %v: %v", iface.Name, err)
			}

			logrus.Debugf("r4addr= %s", r4addr)
			// XXX: delete only the main container (works fine but dunno ?)
			if spec.Process.Args[0] != "/pause" {
				if err := netlink.AddrDel(ifaceLink, r4addr); err != nil {
					return nil, fmt.Errorf("removing address %v from device %q: %v",
						iface.Name, ipNet, err)
				}
			}
			return ifInfo, nil
		}
	}

	return ifInfo, nil
}

func setupNetwork(spec *specs.Spec) (*lklIfInfo, error) {
	var netnsPath string

	// disable HW offload if raw socket
	vethHostIf := getVethHost(spec)
	// only pause container on k8s retures non-nil value
	if vethHostIf != nil {
		// XXX: this disable can be eliminated when vnethdr on raw sock
		// is implemented.
		disableTxCsumOffloadForRawsock(vethHostIf.Name)
	}

	if spec.Linux != nil {
		for _, v := range spec.Linux.Namespaces {
			if v.Type == specs.NetworkNamespace {
				netnsPath = v.Path
				break
			}
		}
	}

	// if there is no path, then runu assumes
	// it runs on docker (not on k8s)
	if netnsPath == "" {
		logrus.Infof("no netns detected: no addr configuration, skipped")
		return nil, nil
	}

	logrus.Infof("nspath= %s", netnsPath)
	nsh, err := netns.GetFromPath(netnsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get netns handle %s", err)
	}

	if err := netns.Set(nsh); err != nil {
		return nil, fmt.Errorf("unable to get set netns %s", err)
	}

	// now traverse in netns
	return getVethInfo(spec)
}
