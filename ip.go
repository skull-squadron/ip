package ip

import (
  "bytes"
  "errors"
  "net"
  "strings"
)

var ipv4InIPv6Prefix = []byte{
  0x00, 0x00, 0x00, 0x00,
  0x00, 0x00, 0x00, 0x00,
  0x00, 0x00, 0xFF, 0xFF,
  // A,    B,    C,    D
}

type IP struct {
  *net.IPNet
  Zone string
}

const (
  NoZone  = ""
  ZoneSep = "%"
)

func ParseZone(s string) (ip, zone string, err error) {
  parts := strings.Split(s, ZoneSep)
  if len(parts) != 1 && len(parts) != 2 {
    err = errors.New("Address/network may may contain only one '%': <addr>%<zone>")
    return
  }
  ip = parts[0]
  if len(parts) == 2 {
    zone = parts[1]
  } else {
    zone = NoZone
  }
  return
}

/* Parses:
   IPv4 addr w/o zone  4.5.6.7
   IPv4 addr w/ zone   1.2.3.4%lo0
   IPv4 net w/o zone   192.168.0.0/16
   IPv4 net w/ zone    192.168.0.0/16%eth0
   IPv6 addr w/o zone  ::1
   IPv6 addr w/ zone   ::1%eth0
   IPv6 net w/o zone   2001:DB8::/48
   IPv6 net w/ zone    2001:DB8::/48%eth0
*/
func Parse(s string) (r IP, err error) {
  s, r.Zone, err = ParseZone(s)
  if err != nil {
    return
  }

  _, r.IPNet, _ = net.ParseCIDR(s)

  if r.IPNet != nil { // we're done
    return
  }

  ip := net.ParseIP(s)
  if ip == nil {
    err = errors.New("Bad IPv4/v6 IP addr or network")
    return
  }

  r.IPNet = &net.IPNet{ip, net.CIDRMask(len(ip)*8, 0)}
  return
}

func (n IP) Compare(n2 IP) bool {
  if !n.CompareZone(n2.Zone) {
    return false
  }
  if bytes.Compare(n.IPNet.IP, n2.IPNet.IP) != 0 {
    return false
  }
  if bytes.Compare(n.IPNet.Mask, n2.IPNet.Mask) != 0 {
    return false
  }
  return true
}

func (n IP) IsIPv6() bool {
  return !n.IsIPv4()
}

func (n IP) IsIPv4() bool {
  if len(n.IPNet.IP) == 4 {
    return true
  }
  return bytes.Compare(n.IPNet.IP[:len(ipv4InIPv6Prefix)], ipv4InIPv6Prefix) == 0
}

func (n IP) IsNetwork() bool {
  for _, v := range n.IPNet.Mask {
    if v != 0xff {
      return true
    }
  }
  return false
}

func (n IP) HasZone() bool {
  return n.Zone != NoZone
}

func (n IP) CompareZone(zone string) bool {
  return zone == n.Zone
}

func (n IP) CompareZoneToInterface(iface *net.Interface) bool {
  return iface == nil || n.CompareZone(iface.Name)
}

// no zone = all interfaces
func (n IP) Interfaces() (ifaces []net.Interface, err error) {
  if !n.HasZone() {
    return net.Interfaces()
  }
  iface, err := net.InterfaceByName(n.Zone)
  ifaces = []net.Interface{*iface}
  return
}

// iface: nil = any interface
func (n IP) ContainsWithInterface(ip net.IP, iface *net.Interface) bool {
  return n.CompareZoneToInterface(iface) && n.IPNet.Contains(ip)
}

// any interface is allowed
func (n IP) Contains(ip net.IP) bool {
  return n.ContainsWithInterface(ip, nil /* any interface */)
}

func (n IP) Network() string {
  return "ip+net+zone"
}

func (n IP) String() string {
  if n.HasZone() {
    return (*n.IPNet).String() + ZoneSep + n.Zone
  }
  return (*n.IPNet).String()
}
