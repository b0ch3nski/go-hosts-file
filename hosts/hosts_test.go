package hosts

import (
	"bytes"
	"io"
	"net/http"
	"net/netip"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const exampleInput1 = `
#comment
    # comment spaces
	#	comment tab and empty line

127.0.0.1 localhost
127.0.0.1 the-same the-same
127.0.0.1 the-same
192.168.1.1		tabs    spaces
192.168.1.2 tabs # some comment here
	  192.168.1.3 ;and here too - why not?
192.168.1.4 d01 d02 d03 d04 d05 d06 d07 d08 d09 d10 d11 d12 d13 d14 d15 d16 d17 d18 d19 d20 d21 d22 d23 d24 d25 d26 d27 d28 d29 d30 d31 d32 d33 d34 d35 d36 d37 d38 d39 d40 d41 d42 d43 d44 d45 d46 d47 d48 d49 d50
192.168.1.5 this-will-be-a-super-long-totally-invalid-domain-that-will-exceed-sane-amount-of-chars-1.com this-will-be-a-super-long-totally-invalid-domain-that-will-exceed-sane-amount-of-chars-2.com this-will-be-a-super-long-totally-invalid-domain-that-will-exceed-sane-amount-of-chars-3.com
`

const exampleInput2 = `
172.16.0.1 1bad.org totaly$%@wrong .looked.ok this.is.bad.too. good321
010.0.10.1 tabs
not-an-ip but-domain.ok
`

const benchHostListUrl = `https://raw.githubusercontent.com/StevenBlack/hosts/3.14.43/hosts`

var (
	ip_127_0_0_1   = netip.AddrFrom4([4]byte{127, 0, 0, 1})
	ip_192_168_1_1 = netip.AddrFrom4([4]byte{192, 168, 1, 1})
	ip_192_168_1_2 = netip.AddrFrom4([4]byte{192, 168, 1, 2})
	ip_192_168_1_3 = netip.AddrFrom4([4]byte{192, 168, 1, 3})
	ip_192_168_1_4 = netip.AddrFrom4([4]byte{192, 168, 1, 4})
	ip_192_168_1_5 = netip.AddrFrom4([4]byte{192, 168, 1, 5})
	ip_172_16_0_1  = netip.AddrFrom4([4]byte{172, 16, 0, 1})
)

func testCommon(t *testing.T, h *Hosts) {
	// no duplicated entries
	equalStrArr(t, []string{"localhost", "the-same"}, h.GetAlias(ip_127_0_0_1))
	equalStrArr(t, []string{"127.0.0.1"}, ipArrStr(h.GetIP("localhost")))

	// parsing tabs/spaces separators
	equalStrArr(t, []string{"tabs", "spaces"}, h.GetAlias(ip_192_168_1_1))

	// multiple IPs for one alias
	equalStrArr(t, []string{"192.168.1.1", "192.168.1.2"}, ipArrStr(h.GetIP("tabs")))

	// no entries with IP only
	equal(t, 0, len(h.GetAlias(ip_192_168_1_3)))

	// only valid aliases
	equalStrArr(t, []string{"good321"}, h.GetAlias(ip_172_16_0_1))
	equal(t, 0, len(h.GetIP("1bad.org")))
}

func TestHostsManipulation(t *testing.T) {
	// create example hosts file
	h := New()
	h.Add(ip_127_0_0_1, "localhost", "the-same", "the-same")
	h.Add(ip_192_168_1_1, "tabs")
	h.Add(ip_192_168_1_1, "spaces")
	h.Add(ip_192_168_1_2, "tabs")
	h.Add(ip_192_168_1_3)
	h.Add(ip_192_168_1_3, "", "")
	h.Add(ip_172_16_0_1, "1bad.org", "totaly$%@wrong", ".looked.ok", "this.is.bad.too.", "good321")

	// example has 4 valid entries
	equal(t, 4, h.Len())

	// common tests run
	testCommon(t, &h)

	// entry deleted by IP address
	h.DelByIP(ip_127_0_0_1)
	equal(t, 0, len(h.GetIP("localhost")))
	equal(t, 0, len(h.GetIP("the-same")))
	equal(t, 0, len(h.GetAlias(ip_127_0_0_1)))

	// entries deleted by alias
	h.DelByAlias("tabs")
	equal(t, 0, len(h.GetIP("tabs")))
	equal(t, 0, len(h.GetIP("spaces")))
	equal(t, 0, len(h.GetAlias(ip_192_168_1_1)))
	equal(t, 0, len(h.GetAlias(ip_192_168_1_2)))

	// only 1 entry left
	equal(t, 1, h.Len())
}

func TestReadWriteHostsFile(t *testing.T) {
	// read example hosts files
	h := New()
	if errRead1 := h.Read(strings.NewReader(exampleInput1)); errRead1 != nil {
		t.Fatal(errRead1)
	}
	if errRead2 := h.Read(strings.NewReader(exampleInput2)); errRead2 != nil {
		t.Fatal(errRead2)
	}

	// example has 6 valid entries
	equal(t, 6, h.Len())

	// common tests run
	testCommon(t, &h)

	// long lines parsed
	equalStrArr(t, []string{"192.168.1.4"}, ipArrStr(h.GetIP("d50")))
	equal(t, 50, len(h.GetAlias(ip_192_168_1_4)))
	equal(t, 3, len(h.GetAlias(ip_192_168_1_5)))

	// write parsed example as new hosts file
	var buf bytes.Buffer
	if errWrite := h.Write(&buf); errWrite != nil {
		t.Fatal(errWrite)
	}
	b := buf.Bytes()

	// basic output verification
	equal(t, 1, bytes.Count(b, []byte("127.0.0.1")))
	equal(t, 1, bytes.Count(b, []byte("the-same")))
	equal(t, 1, bytes.Count(b, []byte("192.168.1.1")))
	equal(t, 2, bytes.Count(b, []byte("tabs")))
	equal(t, 0, bytes.Count(b, []byte("192.168.1.3")))

	// verify if 50 aliases are split into 6 lines (max 9 per line)
	equal(t, 6, bytes.Count(b, []byte("192.168.1.4")))

	// verify if line exceeding character limit (255) is split into 2 lines
	equal(t, 2, bytes.Count(b, []byte("192.168.1.5")))

	// verify file format
	equal(t, 59, bytes.Count(b, []byte(" ")))
	equal(t, 12, bytes.Count(b, []byte("\n")))
}

func BenchmarkStevenBlackHosts(b *testing.B) {
	resp, errResp := http.Get(benchHostListUrl)
	if errResp != nil {
		b.Fatal(errResp)
	}
	defer resp.Body.Close()

	list, errList := io.ReadAll(resp.Body)
	if errList != nil {
		b.Fatal(errList)
	}
	b.Logf("Steven Black hosts bytes: %d", len(list))

	var h Hosts
	b.Run("read", func(bb *testing.B) {
		for i := 0; i < bb.N; i++ {
			h = New()
			if errRead := h.Read(bytes.NewReader(list)); errRead != nil {
				bb.Fatal(errRead)
			}
		}
	})
	b.Logf("Parsed entries: %d", len(h.ipToAlias[netip.IPv4Unspecified()]))

	b.Run("write", func(bb *testing.B) {
		for i := 0; i < bb.N; i++ {
			var buf bytes.Buffer
			if errWrite := h.Write(&buf); errWrite != nil {
				bb.Fatal(errWrite)
			}
		}
	})
}

func equal(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Not equal: \nexpected: %v\nactual  : %v", expected, actual)
	}
}

func equalStrArr(t *testing.T, expected, actual []string) {
	sort.Strings(expected)
	sort.Strings(actual)
	equal(t, expected, actual)
}

func ipArrStr(ips []netip.Addr) []string {
	res := make([]string, 0, len(ips))
	for _, ip := range ips {
		res = append(res, ip.String())
	}
	return res
}
