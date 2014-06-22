package ip

import (
  "net"
  "strings"
  "testing"
)

type parseCase struct {
  x         string
  ipv6, net bool
  zone      string
}

type stringCase struct {
  x string
}

var mustParse = []parseCase{
  {"4.5.6.7", false, false, ""},
  {"1.2.3.4%lo0", false, false, "lo0"},
  {"192.168.0.0/16", false, true, ""},
  {"192.168.0.0/16%eth0", false, true, "eth0"},
  {"::1", true, false, ""},
  {"::1%eth0", true, false, "eth0"},
  {"2001:DB8::/48", true, true, ""},
  {"2001:DB8::/48%eth0", true, true, "eth0"},
}
var stringCases = []stringCase{
  {"4.5.6.7"},
  {"1.2.3.4%lo0"},
  {"192.168.0.0/16"},
  {"192.168.0.0/16%eth0"},
  {"::1"},
  {"::1%eth0"},
  {"2001:DB8::/48"},
  {"2001:DB8::/48%eth0"},
}

var mustNotParse = []string{
  "",
  "a1.2.3.4%lo0",
  "%192.168.0.0/16",
  "192.168.0",
  "192.168.0.0.0/16%eth0",
  "%::1%",
  "::1z%eth0",
  "2001:DB8::/48%%",
  "2001:DB8::/48%eth0%",
}

func TestParse(t *testing.T) {
  for _, validCase := range mustParse {
    ipn, err := Parse(validCase.x)
    if err != nil {
      t.Errorf("Parse(\"%s\") failed (should parse)", validCase.x)
    }
    if validCase.ipv6 {
      if !ipn.IsIPv6() {
        t.Errorf("Parse(\"%s\") failed (should be ipv6)", validCase.x)
      }
    } else {
      if !ipn.IsIPv4() {
        t.Errorf("Parse(\"%s\") failed (should be ipv4)", validCase.x)
      }
    }
    if validCase.net {
      if !ipn.IsNetwork() {
        t.Errorf("Parse(\"%s\") failed (should be network addr)", validCase.x)
      }
    } else {
      if ipn.IsNetwork() {
        t.Errorf("Parse(\"%s\") failed (should not be network addr)", validCase.x)
      }
    }
    if !ipn.EqualZone(validCase.zone) {
      t.Errorf("Parse(\"%s\") failed (bad zone %s != %s) %s", validCase.x, ipn.Zone, validCase.zone)
    }
  }

  for _, invalid := range mustNotParse {
    if _, err := Parse(invalid); err == nil {
      t.Errorf("Parse(\"%s\") failed (should not parse)", invalid)
    }
  }
}

func TestContains(t *testing.T) {
  x, _ := Parse("80.0.0.0/8")
  if !x.Contains(net.ParseIP("80.1.2.3")) {
    t.Errorf("Contains() fails")
  }

  y, _ := Parse("80.1.2.3")
  if !y.Contains(net.ParseIP("80.1.2.3")) {
    t.Errorf("Contains() fails 2")
  }
}

func TestInterfaces(t *testing.T) {
  x, _ := Parse("80.0.0.0/8")
  ifaces, err := x.Interfaces()
  if err != nil {
    t.Errorf("Interfaces() failed, err=", err)
  }
  for _, iface := range ifaces {
    t.Logf("Interface: iface=", iface)
  }
}

func TestNetwork(t *testing.T) {
  x, _ := Parse("80.0.0.0/8")
  if x.Network() != "ip+net+zone" {
    t.Errorf("Network() failed")
  }
}

func TestString(t *testing.T) {
  for _, stringCase := range stringCases {
    j, _ := Parse(stringCase.x)
    if s := j.String(); strings.ToUpper(s) != strings.ToUpper(stringCase.x) {
      t.Errorf("String() failed (%s != %s)", s, stringCase.x)
    }
  }
}
