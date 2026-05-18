package main

import "testing"

func TestRestoreMainWindow_CallOrder(t *testing.T) {
	var calls []string
	restoreMainWindow(
		func() { calls = append(calls, "show") },
		func() { calls = append(calls, "center") },
		func() { calls = append(calls, "focus") },
	)

	want := []string{"show", "center", "focus"}
	if len(calls) != len(want) {
		t.Fatalf("expected %d calls, got %d", len(want), len(calls))
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("expected call %d to be %q, got %q", i, want[i], calls[i])
		}
	}
}

func TestRestoreMainWindow_AllowsNilCallbacks(t *testing.T) {
	assertNoPanic := func(name string, fn func()) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("expected no panic, got %v", r)
				}
			}()
			fn()
		})
	}

	assertNoPanic("all nil", func() {
		restoreMainWindow(nil, nil, nil)
	})
	assertNoPanic("only show", func() {
		restoreMainWindow(func() {}, nil, nil)
	})
	assertNoPanic("only center", func() {
		restoreMainWindow(nil, func() {}, nil)
	})
	assertNoPanic("only focus", func() {
		restoreMainWindow(nil, nil, func() {})
	})
}
