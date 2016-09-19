package mdns

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestServer_StartStop(t *testing.T) {
	s := makeService(t)
	z := []Zone{s}
	serv, err := NewServer(&Config{Zones: z})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer serv.Shutdown()
}

func TestServer_Lookup(t *testing.T) {
	s := makeServiceWithServiceName(t, "_foobar._tcp")
	z := []Zone{s}
	serv, err := NewServer(&Config{Zones: z})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer serv.Shutdown()

	entries := make(chan *ServiceEntry, 1)
	found := false
	go func() {
		select {
		case e := <-entries:
			if e.Name != "hostname._foobar._tcp.local." {
				t.Fatalf("bad: %v", e)
			}
			if e.Port != 80 {
				t.Fatalf("bad: %v", e)
			}
			if e.Info != "Local web server" {
				t.Fatalf("bad: %v", e)
			}
			found = true

		case <-time.After(80 * time.Millisecond):
			t.Fatalf("timeout")
		}
	}()

	params := &QueryParam{
		Services: []string{"_foobar._tcp"},
		Domain:   "local",
		Timeout:  50 * time.Millisecond,
		Entries:  entries,
	}
	err = Query(params)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !found {
		t.Fatalf("record not found")
	}
}

func TestServer_MultiLookup(t *testing.T) {
	z := []Zone{makeServiceWithServiceName(t, "_foobar._tcp"), makeServiceWithServiceName(t, "_barfoo._tcp")}
	serv, err := NewServer(&Config{Zones: z})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	defer serv.Shutdown()
	entries := make(chan *ServiceEntry, 2)
	var count uint64
	go func() {
		for {
			select {
			case e := <-entries:
				if !((e.Name == "hostname._foobar._tcp.local.") || (e.Name == "hostname._barfoo._tcp.local.")) {
					t.Fatalf("bad: %v", e)
				}
				if e.Port != 80 {
					t.Fatalf("bad: %v", e)
				}
				if e.Info != "Local web server" {
					t.Fatalf("bad: %v", e)
				}
				atomic.AddUint64(&count, 1)

			case <-time.After(10000 * time.Millisecond):
				t.Fatalf("timeout")
			}
		}
	}()
	params := &QueryParam{
		Services: []string{"_barfoo._tcp", "_foobar._tcp"},
		Domain:   "local",
		Timeout:  100 * time.Millisecond,
		Entries:  entries,
	}
	err = Query(params)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if count != 2 {
		t.Fatalf("both records not found")
	}
}
