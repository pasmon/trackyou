package main

import "testing"

func TestShouldDetachForInteractiveLaunch(t *testing.T) {
	tests := []struct {
		name             string
		isInteractiveTTY bool
		detachMarker     string
		detachEnabled    string
		want             bool
	}{
		{
			name:             "interactive terminal with opt-in and no marker detaches",
			isInteractiveTTY: true,
			detachMarker:     "",
			detachEnabled:    "1",
			want:             true,
		},
		{
			name:             "interactive terminal without opt-in skips detach",
			isInteractiveTTY: true,
			detachMarker:     "",
			detachEnabled:    "",
			want:             false,
		},
		{
			name:             "interactive terminal with marker skips detach",
			isInteractiveTTY: true,
			detachMarker:     "1",
			detachEnabled:    "1",
			want:             false,
		},
		{
			name:             "interactive terminal with non-empty marker skips detach",
			isInteractiveTTY: true,
			detachMarker:     "0",
			detachEnabled:    "1",
			want:             false,
		},
		{
			name:             "non-interactive launch skips detach",
			isInteractiveTTY: false,
			detachMarker:     "",
			detachEnabled:    "1",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldDetachForInteractiveLaunch(tt.isInteractiveTTY, tt.detachMarker, tt.detachEnabled)
			if got != tt.want {
				t.Errorf("shouldDetachForInteractiveLaunch() = %v, want %v", got, tt.want)
			}
		})
	}
}
