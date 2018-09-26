// Windows doesn't have sleep.exe which we use for making the subprocesses stay
// alive for a certain amount of time.

// +build !windows

package main

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func startGoreman(ctx context.Context, t *testing.T, ch <-chan os.Signal, file []byte) {
	t.Helper()
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(file); err != nil {
		t.Fatal(err)
	}
	cfg := &config{
		Procfile: f.Name(),
	}
	if ch == nil {
		ch = notifyCh()
	}
	if err := start(ctx, ch, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestGoreman(t *testing.T) {
	var file = []byte(`
web1: sleep 0.1
web2: sleep 0.1
web3: sleep 0.1
web4: sleep 0.1
`)
	startGoreman(context.TODO(), t, nil, file)
}

func TestGoremanSignal(t *testing.T) {
	var file = []byte(`
web1: sleep 10
web2: sleep 10
web3: sleep 10
web4: sleep 10
`)
	now := time.Now()
	sc := make(chan os.Signal, 1)
	go func() {
		sc <- os.Interrupt
	}()
	startGoreman(context.TODO(), t, sc, file)
	if dur := time.Since(now); dur > 50*time.Millisecond {
		t.Errorf("test took too much time; should have canceled after 10ms, got %s", dur)
	}
}
