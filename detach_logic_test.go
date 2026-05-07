package main

import "testing"

func TestShouldDetachForInteractiveLaunch(t *testing.T) {
	tests := []struct {
		name             string
		isInteractiveTTY bool
		detachMarker     string
		want             bool
	}{
		{
			name:             "interactive terminal without marker detaches",
			isInteractiveTTY: true,
			detachMarker:     "",
			want:             true,
		},
		{
			name:             "interactive terminal with marker skips detach",
			isInteractiveTTY: true,
			detachMarker:     "1",
			want:             false,
		},
		{
			name:             "interactive terminal with non-empty marker skips detach",
			isInteractiveTTY: true,
			detachMarker:     "0",
			want:             false,
		},
		{
			name:             "non-interactive launch skips detach",
			isInteractiveTTY: false,
			detachMarker:     "",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldDetachForInteractiveLaunch(tt.isInteractiveTTY, tt.detachMarker)
			if got != tt.want {
				t.Fatalf("shouldDetachForInteractiveLaunch() = %v, want %v", got, tt.want)
			}
		})
	}
}
