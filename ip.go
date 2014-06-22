package ip

import (
  "bytes"
  "errors"
  "net"
  "strings"
)

// conforms to interface net.Addr
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
  switch len(parts) {
  case 1:
    ip, zone = s, NoZone
  case 2:
    ip, zone = parts[0], parts[1]
  default:
    err = errors.New("Address/network may may contain only one '%': <addr>%<zone>")
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

  r.IPNet = &net.IPNet{ip, nil}
  return
}

func (n IP) Equal(n2 IP) bool {
  return n.EqualZone(n2.Zone) &&
    n.IPNet.IP.Equal(n2.IPNet.IP) &&
    bytes.Compare(n.IPNet.Mask, n2.IPNet.Mask) == 0
}

func (n IP) IsIPv6() bool {
  return !n.IsIPv4()
}

func (n IP) IsIPv4() bool {
  return n.IPNet.IP.To4() != nil
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

func (n IP) EqualZone(zone string) bool {
  return zone == n.Zone
}

func (n IP) EqualInterface(iface *net.Interface) bool {
  return iface == nil || n.EqualZone(iface.Name)
}

// no zone = all interfaces
func (n IP) Interfaces() (ifaces []net.Interface) {
  if !n.HasZone() {
    ifaces, _ = net.Interfaces()
    return
  }
  iface, err := net.InterfaceByName(n.Zone)
  if err != nil {
    ifaces = []net.Interface{}
  } else {
    ifaces = []net.Interface{*iface}
  }
  return
}

// iface: nil = any interface
func (n IP) ContainsWithInterface(ip net.IP, iface *net.Interface) bool {
  return n.EqualInterface(iface) && (n.IPNet.Contains(ip) || n.IPNet.IP.Equal(ip))
}

// any interface is allowed
func (n IP) Contains(ip net.IP) bool {
  return n.ContainsWithInterface(ip, nil /* any interface */)
}

func (n IP) Network() string {
  return "ip+net+zone"
}

func (n IP) String() (s string) {
  if n.IPNet.Mask == nil {
    s = n.IPNet.IP.String()
  } else {
    s = n.IPNet.String()
  }
  if n.HasZone() {
    s += ZoneSep + n.Zone
  }
  return
}
