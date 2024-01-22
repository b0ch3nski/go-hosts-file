package hosts

import (
	"bufio"
	"bytes"
	"io"
	"net/netip"
	"regexp"
	"strings"
)

const (
	maxAliasesPerLine = 9
	maxLineLength     = 255
)

var (
	rgxHostsFileLine = regexp.MustCompile(`(\S+)+`)
	rgxValidAlias    = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-.]*[a-zA-Z0-9]$`)
)

type strSet map[string]struct{}
type ipSet map[netip.Addr]struct{}

// Hosts is the representation of IP-to-Host and Host-to-IP mappings.
type Hosts struct {
	ipToAlias map[netip.Addr]strSet
	aliasToIp map[string]ipSet
}

// New creates empty `Hosts` instance.
func New() Hosts {
	return Hosts{ipToAlias: make(map[netip.Addr]strSet), aliasToIp: make(map[string]ipSet)}
}

// Len returns amount of mapped IP addresses.
func (h *Hosts) Len() int {
	return len(h.ipToAlias)
}

// GetAlias returns all aliases associated with specified IP address.
func (h *Hosts) GetAlias(ip netip.Addr) []string {
	als := h.ipToAlias[ip]
	res := make([]string, 0, len(als))
	for a := range als {
		res = append(res, a)
	}
	return res
}

// GetIP returns all IP addresses associated with specified alias.
func (h *Hosts) GetIP(alias string) []netip.Addr {
	ips := h.aliasToIp[alias]
	res := make([]netip.Addr, 0, len(ips))
	for ip := range ips {
		res = append(res, ip)
	}
	return res
}

// Add adds IP:[]Host mapping skipping invalid IPs and hosts aliases.
func (h *Hosts) Add(ip netip.Addr, alias ...string) {
	if !ip.IsValid() || len(alias) == 0 {
		return
	}
	if _, okIp := h.ipToAlias[ip]; !okIp {
		h.ipToAlias[ip] = make(strSet, len(alias))
	}

	for _, a := range alias {
		if !rgxValidAlias.MatchString(a) {
			continue
		}
		h.ipToAlias[ip][a] = struct{}{}

		if _, okA := h.aliasToIp[a]; !okA {
			h.aliasToIp[a] = make(ipSet, 1)
		}
		h.aliasToIp[a][ip] = struct{}{}
	}

	if len(h.ipToAlias[ip]) == 0 {
		delete(h.ipToAlias, ip)
	}
}

// DelByIP removes all aliases associated with specified IP address.
func (h *Hosts) DelByIP(ip netip.Addr) {
	for a := range h.ipToAlias[ip] {
		delete(h.aliasToIp, a)
	}
	delete(h.ipToAlias, ip)
}

// DelByAlias removes all IP addresses (and their aliases) associated with specified alias.
func (h *Hosts) DelByAlias(alias string) {
	for ip := range h.aliasToIp[alias] {
		h.DelByIP(ip)
	}
}

// Read appends hosts read from file using provided `io.Reader`.
func (h *Hosts) Read(reader io.Reader) error {
	bufRd := bufio.NewReader(reader)

	for {
		line, errRead := bufRd.ReadString('\n')
		if errRead != nil {
			if errRead == io.EOF {
				break
			}
			return errRead
		}

		// skip comments
		if idx := strings.IndexAny(line, `#;`); idx > -1 {
			line = line[0:idx]
		}

		if matchHosts := rgxHostsFileLine.FindAllString(line, -1); len(matchHosts) > 1 {
			ip, errParse := netip.ParseAddr(matchHosts[0])
			if errParse != nil {
				continue
			}
			h.Add(ip, matchHosts[1:]...)
		}
	}

	return nil
}

// Write writes all mappings from `Hosts` instance to hosts file using provided `io.Writer`.
func (h *Hosts) Write(writer io.Writer) error {
	bufWr := bufio.NewWriter(writer)

	for ip, aliases := range h.ipToAlias {
		addr := ip.String()
		lineLen := len(addr)
		aliasCount := 0

		bufWr.WriteString(addr)
		for alias := range aliases {
			if (aliasCount > 0 && aliasCount%maxAliasesPerLine == 0) || lineLen+len(alias)+1 > maxLineLength {
				bufWr.WriteString("\n")
				bufWr.WriteString(addr)
				lineLen = len(addr)
			}
			bufWr.WriteString(" ")
			bufWr.WriteString(alias)

			lineLen += len(alias) + 1 // space
			aliasCount++
		}
		bufWr.WriteString("\n")
	}

	return bufWr.Flush()
}

func (h *Hosts) String() string {
	var buf bytes.Buffer
	h.Write(&buf)
	return buf.String()
}
