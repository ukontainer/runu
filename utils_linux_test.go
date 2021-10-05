package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	nsName1      = "runu-test1"
	vethHost     = "runu-host"
	vethPeer     = "runu-peer"
	vethPeerLast = "eth0"
)

var (
	origNs netns.NsHandle
)

func createVethPair(t *testing.T) {
	veth := &netlink.Veth{LinkAttrs: netlink.LinkAttrs{Name: vethHost},
		PeerName: vethPeer}
	t.Logf("veth = %v", veth)

	// add link
	if err := netlink.LinkAdd(veth); err != nil {
		t.Fatal(err)
	}

}

func configGuest(t *testing.T, newns netns.NsHandle, v4Addr, v4Gw string) {
	vethG, err := netlink.LinkByName(vethPeer)
	if err != nil {
		t.Fatal(err)
	}
	if err := netlink.LinkSetNsFd(vethG, int(newns)); err != nil {
		t.Fatal(err)
	}

	// revert to newns
	if err := netns.Set(newns); err != nil {
		t.Fatal(err)
	}

	// rename veth guest
	if err := netlink.LinkSetName(vethG, vethPeerLast); err != nil {
		t.Fatal(err)
	}

	if err := netlink.LinkSetUp(vethG); err != nil {
		t.Fatal(err)
	}

	// assign IP address to veth guest
	// add ipv4 addr
	addr, _ := netlink.ParseAddr(v4Addr)
	if err := netlink.AddrAdd(vethG, addr); err != nil {
		t.Fatal(err)
	}
	t.Logf("IPv4 address: %v\n", addr)

	// add ipv4 default gw
	gw, err := netlink.ParseAddr(v4Gw)
	if err != nil {
		t.Fatal(err)
	}

	// XXX: calico case
	// if v4Gw and v4Addr are in same subnet, it adds
	// an onlink route in addition to default route
	if !addr.Contains(gw.IPNet.IP) {
		t.Logf("gw is not in v4addr mask (%s: %s)", addr, gw.IPNet.IP)

		route := netlink.Route{
			LinkIndex: vethG.Attrs().Index,
			Dst:       gw.IPNet,
		}
		if err := netlink.RouteAdd(&route); err != nil {
			t.Fatal(err)
		}
		t.Logf("Gateway(extra): %v\n", route)
	}
	route := netlink.Route{
		LinkIndex: vethG.Attrs().Index,
		Gw:        gw.IPNet.IP,
		Flags:     int(netlink.FLAG_ONLINK),
		Dst:       nil,
	}
	if err := netlink.RouteAdd(&route); err != nil {
		t.Logf("Failure/Gateway: %v\n", route)
		t.Fatalf("adding route failed: %s", err)
	}
	t.Logf("Gateway: %v\n", route)

	// print created interfaces
	ifaces, _ := net.Interfaces()
	t.Logf("Interfaces: %v\n", ifaces)

	addrs, err := netlink.AddrList(vethG, netlink.FAMILY_ALL)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Address: %v\n", addrs)

}

func createNetNs(t *testing.T, v4Addr, v4Gw string) {
	// Clean up previous names
	destroyNetNs(t)

	// create eth0 pair
	createVethPair(t)

	// store origin ns
	origns, err := netns.Get()
	if err != nil {
		t.Fatal(err)
	}

	origNs = origns

	// ip netns add NAME
	newns, err := netns.NewNamed(nsName1)
	if err != nil {
		t.Fatal(err)
	}

	parent, err := netns.Get()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("netns = %v", newns)
	t.Logf("parent netns = %v", parent)
	t.Logf("origin netns = %v", origns)

	// set ns to veth guest
	if err := netns.Set(origns); err != nil {
		t.Fatal(err)
	}

	// Link up
	vethH, err := netlink.LinkByName(vethHost)
	if err != nil {
		t.Fatal(err)
	}
	if err := netlink.LinkSetUp(vethH); err != nil {
		t.Fatal(err)
	}

	configGuest(t, newns, v4Addr, v4Gw)

}

func destroyNetNs(t *testing.T) {
	t.Log("Cleaning up the netns")

	ns, err := netns.Get()
	if err != nil {
		t.Fatal(err)
	}
	ns.Close()
	netns.DeleteNamed(nsName1)

	// restore origin ns
	netns.Set(origNs)

	// delete veth pair
	vethH, err := netlink.LinkByName(vethHost)
	if err != nil {
		t.Log(err)
		return
	}

	if err := netlink.LinkDel(vethH); err != nil {
		t.Fatal(err)
	}
}

func testSetupSpec(t *testing.T) *specs.Spec {
	spec := Example()

	t.Logf("spec = %v", spec.Linux)

	spec.Linux.Namespaces = []specs.LinuxNamespace{
		{
			Type: specs.NetworkNamespace,
			Path: fmt.Sprintf("/var/run/netns/%s", nsName1),
		},
	}

	return spec
}

func validateJson(t *testing.T, lklJson, v4Gw string) {
	var config lklConfig

	bytes, err := ioutil.ReadFile(lklJson)
	if err != nil {
		t.Fatalf("failed to read JSON file: %s (%s)",
			lklJson, err)
	}

	// decode json
	if err := json.Unmarshal(bytes, &config); err != nil {
		t.Fatalf("failed to decode JSON file: %s (%s)",
			lklJson, err)
	}

	t.Logf("%+v", config)
	gw, _ := netlink.ParseAddr(v4Gw)
	if config.V4Gateway != gw.IP.String() {
		t.Fatalf("gateway address is invalid (expected: %s, value: %s)", gw.IP.String(), config.V4Gateway)
	}
}

func testIPAddressAuto(t *testing.T, v4Addr, v4Gw string) {
	tmp, err := ioutil.TempFile("/tmp/", "lkl-json")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.Remove(tmp.Name())
	})
	jsonOut := tmp.Name()
	t.Logf("jsonOut = %v", jsonOut)

	createNetNs(t, v4Addr, v4Gw)

	spec := testSetupSpec(t)
	json, err := generateLklJsonFile("", &jsonOut, spec)
	if err != nil {
		t.Fatal(err)
	}

	validateJson(t, jsonOut, v4Gw)

	t.Logf("json = %+v", json)
}

func TestIPAddressAuto(t *testing.T) {
	t.Cleanup(func() {
		time.Sleep(time.Second * 0)
		destroyNetNs(t)
	})

	if testing.Verbose() {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// flannel case
	t.Log("==================================")
	testIPAddressAuto(t, "192.168.39.2/24", "192.168.39.1/24")
	// calico case
	t.Log("==================================")
	testIPAddressAuto(t, "192.168.39.2/32", "172.16.0.1/32")
}
