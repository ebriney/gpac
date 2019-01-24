package gpac_test

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/darren/gpac"
)

func init() {
	var mux http.ServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Example")
	})
	go func() {
		log.Fatal(http.ListenAndServe("127.0.0.1:8080", &mux))
	}()

	go func() {
		log.Fatal(http.ListenAndServe("127.0.0.1:8081", &mux))
	}()
}

func TestParseProxy(t *testing.T) {
	proxy := "PROXY 127.0.0.1:8080; SOCKs 127.0.0.1:1080; Direct"

	proxies := gpac.ParseProxy(proxy)

	if len(proxies) != 3 {
		t.Error("Parse failed")
		return
	}

	if proxies[1].Type != "SOCKS" {
		t.Error("Should be SOCKS5")
	}

	if !proxies[2].IsDirect() {
		t.Error("Should be direct")
	}
}

func testProxyGet(t *testing.T, typ string) {
	t.Logf("Test proxy type: %s", typ)

	var p = gpac.Proxy{Type: typ, Address: "127.0.0.1:8080"}

	client := http.Client{
		Transport: &http.Transport{
			Proxy: p.Proxy(),
		},
	}

	resp, err := client.Get("http://127.0.0.1:8081")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(buf) != "Example" {
		t.Errorf("Response not expected: %s", string(buf))
	}
}

func testClientGet(t *testing.T, typ string) {
	t.Logf("Test Client proxy type: %s", typ)

	var p = gpac.Proxy{Type: typ, Address: "127.0.0.1:8080"}

	resp, err := p.Get("http://127.0.0.1:8081")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(buf) != "Example" {
		t.Errorf("Response not expected: %s", string(buf))
	}
}

func TestMultiProxyGet(t *testing.T) {
	//BUG: SOCKS5 seems not work
	knownTypes := []string{"DIRECT", "HTTP"}
	for _, typ := range knownTypes {
		testProxyGet(t, typ)
		testClientGet(t, typ)
	}
}
