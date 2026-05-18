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
	restoreMainWindow(nil, nil, nil)
	restoreMainWindow(func() {}, nil, nil)
	restoreMainWindow(nil, func() {}, nil)
	restoreMainWindow(nil, nil, func() {})
}
