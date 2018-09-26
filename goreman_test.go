package main

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var sleep string

func TestMain(m *testing.M) {
	var dir string
	var err error
	sleep, err = exec.LookPath("sleep")
	if err != nil {
		if runtime.GOOS == "windows" {
			code := `package main;import ("os";"strconv";"time");func main(){i,_:=strconv.ParseFloat(os.Args[1]);time.Sleep(time.Duration(i)*time.Second)}`
			dir, err := ioutil.TempDir("", "goreman-test")
			if err != nil {
				panic(err)
			}
			sleep = filepath.Join(dir, "sleep.exe")
			src := filepath.Join(dir, "sleep.go")
			err = ioutil.WriteFile(src, []byte(code), 0644)
			if err != nil {
				panic(err)
			}
			b, err := exec.Command("go", "build", "-o", sleep, src).CombinedOutput()
			if err != nil {
				panic(string(b))
			}
		} else {
			panic(err)
		}
		oldpath := os.Getenv("PATH")
		os.Setenv("PATH", dir+";"+oldpath)
		defer os.Setenv("PATH", oldpath)

	}
	r := m.Run()

	if dir != "" {
		os.RemoveAll(dir)
	}
	os.Exit(r)
}

func startGoreman(ctx context.Context, t *testing.T, ch <-chan os.Signal, file []byte) error {
	t.Helper()
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(file); err != nil {
		t.Fatal(err)
	}
	cfg := &config{
		ExitOnError: true,
		Procfile:    f.Name(),
	}
	if ch == nil {
		ch = notifyCh()
	}
	return start(ctx, ch, cfg)
}

func TestGoreman(t *testing.T) {
	var file = []byte(`
web1: sleep 0.1
web2: sleep 0.1
web3: sleep 0.1
web4: sleep 0.1
`)
	if err := startGoreman(context.TODO(), t, nil, file); err != nil {
		t.Fatal(err)
	}
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
	if err := startGoreman(context.TODO(), t, sc, file); err != nil {
		t.Fatal(err)
	}
	if dur := time.Since(now); dur > time.Second {
		t.Errorf("test took too much time; should have canceled after 1s, got %s", dur)
	}
}

func TestGoremanExitsOnError(t *testing.T) {
	var file = []byte(`
web1: sleep 10
web2: sleep 0.01 && foobarbangbazunknownproc
web3: sleep 10
web4: sleep 10
`)
	now := time.Now()
	// process 2 should exit which should trigger exit of entire program.
	if err := startGoreman(context.TODO(), t, nil, file); err == nil {
		t.Fatal("got nil err, should have received error")
	}
	if dur := time.Since(now); dur > time.Second {
		t.Errorf("test took too much time; should have canceled after 1s, got %s", dur)
	}
}

func TestGoremanExitsOnErrorOtherWay(t *testing.T) {
	var file = []byte(`
web1: sleep 10
web2: sleep 0.01 && exit 2
web3: sleep 10
web4: sleep 10
`)
	// process 2 should exit which should trigger exit of entire program.
	now := time.Now()
	if err := startGoreman(context.TODO(), t, nil, file); err == nil {
		t.Fatal("got nil err, should have received error")
	}
	if dur := time.Since(now); dur > time.Second {
		t.Errorf("test took too much time; should have canceled after 1s, got %s", dur)
	}
}

func TestGoremanStopProcDoesntStopOtherProcs(t *testing.T) {
	var file = []byte(`
web1: sleep 10
web2: sleep 10
web3: sleep 10
web4: sleep 10
`)
	goremanStopped := make(chan struct{}, 1)
	sc := make(chan os.Signal, 1)
	go func() {
		startGoreman(context.TODO(), t, sc, file)
		goremanStopped <- struct{}{}
	}()
	for {
		if procs == nil {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		proc, ok := procs["web2"]
		if !ok {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		proc.mu.Lock()
		cmd := proc.cmd
		proc.mu.Unlock()
		if cmd == nil {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		if err := stopProc("web2", nil); err != nil {
			t.Fatal(err)
		}
		break
	}
	select {
	case <-goremanStopped:
		t.Errorf("stopping web2 subprocess should not have stopped supervisor")
	case <-time.After(30 * time.Millisecond):
	}
	sc <- os.Interrupt
}
