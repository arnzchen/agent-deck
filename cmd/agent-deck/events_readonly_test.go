package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/asheshgoplani/agent-deck/internal/events"
)

func TestReadonlyFollowerUsesResolvedProfileAndNormalizedBus(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, "data"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	events.SetProfile("resolved-reader")
	t.Cleanup(func() { events.SetProfile("default") })
	// Seed the exact resolved profile independently of the default process bus.
	reader, err := events.OpenReader("resolved-reader")
	if err == nil {
		reader.Close()
		t.Fatal("absent reader must not create storage")
	}
	writer, err := events.OpenAt(filepath.Join(home, "data", "agent-deck", "bus", "resolved-reader"), events.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	if _, err := writer.Commit("session.turn", "child", map[string]string{"state": "completed"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"events", " EVENTS ", ""} {
		bus, err := openFollowerBus(name, true)
		if err != nil {
			t.Fatal(err)
		}
		if !bus.ReadOnly() || bus.Stats().Dir != writer.Stats().Dir {
			t.Fatal("follower must be read-only and use the resolved profile")
		}
		release := bus.Want("tmux.output")
		if matches, _ := filepath.Glob(filepath.Join(bus.Stats().Dir, "want", "*")); len(matches) != 0 {
			t.Fatal("read-only follower created a demand lease")
		}
		release()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		sub, err := bus.Subscribe(ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		select {
		case frame := <-sub.Frames():
			if frame.SessionID != "child" {
				t.Fatal("wrong profile event")
			}
		case <-ctx.Done():
			t.Fatal("seeded event not observed")
		}
		cancel()
		bus.Close()
	}
}
