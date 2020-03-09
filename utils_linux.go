package main

import (
	"fmt"
	"net"
	"sort"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func getVethHost() net.Interface {
	// list root ns's interfaces before going into net ns
	iface0, _ := net.Interfaces()

	// sort by ifindex
	sort.SliceStable(iface0, func(i, j int) bool {
		return iface0[i].Index < iface0[j].Index
	})

	// XXX: need better detection of veth host-side pair
	lastIf := iface0[len(iface0)-1]

	return lastIf
}

func getVethInfo(spec *specs.Spec) (*lklIfInfo, error) {
	ifInfo := new(lklIfInfo)

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

			logrus.Warnf("ifaddr= %s, ipnet=%s", ifaddr, ipNet)
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
	vethHostIf := getVethHost()
	disableTxCsumOffloadForRawsock(vethHostIf.Name)

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

	logrus.Debugf("nspath= %s", netnsPath)
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
