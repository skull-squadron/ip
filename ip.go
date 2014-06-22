package ip

import (
  "bytes"
  "errors"
  "net"
  "strings"
)

// conforms to interface net.Addr
type IP struct {
  IP   net.IP
  Mask net.IPMask
  Zone string // used with IPv4 and IPv6 to indicate interface name
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
    err = errors.New("IP may contain only one '%': <ip>%<zone>")
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

  _, ipnet, _ := net.ParseCIDR(s)

  if ipnet != nil { // we're done
    r.IP = ipnet.IP
    r.Mask = ipnet.Mask
    return
  }

  r.IP = net.ParseIP(s)
  if r.IP == nil {
    err = errors.New("Bad IPv4/v6 IP addr or network")
    return
  }

  return
}

func (n IP) Equal(n2 IP) bool {
  return n.EqualZone(n2.Zone) &&
    n.IP.Equal(n2.IP) &&
    bytes.Compare(n.Mask, n2.Mask) == 0
}

func (n IP) IsIPv6() bool {
  return !n.IsIPv4()
}

func (n IP) IsIPv4() bool {
  return n.IP.To4() != nil
}

func (n IP) IsNetwork() bool {
  for _, v := range n.Mask {
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

func (n IP) IPNet() *net.IPNet {
  return &net.IPNet{IP: n.IP, Mask: n.Mask}
}

func (n IP) IPAddr() net.IPAddr {
  return net.IPAddr{IP: n.IP, Zone: n.Zone}
}

// iface: nil = any interface
func (n IP) ContainsWithInterface(ip net.IP, iface *net.Interface) bool {
  return n.EqualInterface(iface) && (n.IP.Equal(ip) || n.IPNet().Contains(ip))
}

// any interface is allowed
func (n IP) Contains(ip net.IP) bool {
  return n.ContainsWithInterface(ip, nil /* any interface */)
}

func (n IP) Network() string {
  return "ip+net+zone"
}

func (n IP) String() (s string) {
  if n.Mask == nil {
    s = n.IP.String()
  } else {
    s = n.IPNet().String()
  }
  if n.HasZone() {
    s += ZoneSep + n.Zone
  }
  return
}
