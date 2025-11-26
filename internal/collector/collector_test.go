package collector

import (
	"context"
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		opts    Options
		wantErr bool
	}{
		{
			name:   "basic collection",
			target: "example.com",
			opts: Options{
				EnablePorts: false,
				NoAgent:     true,
				Timeout:     5 * time.Second,
			},
			wantErr: false, // DNS is implemented, should succeed
		},
		{
			name:   "with ports",
			target: "example.com",
			opts: Options{
				EnablePorts: true,
				NoAgent:     false,
				Timeout:     10 * time.Second,
			},
			wantErr: false, // DNS is implemented, should succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := Collect(ctx, tt.target, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("Collect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// DNS is implemented, should not fail
			if err != nil && !tt.wantErr {
				t.Errorf("Collect() unexpected error = %v", err)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		item  int
		want  bool
	}{
		{
			name:  "contains item",
			slice: []int{1, 2, 3, 4, 5},
			item:  3,
			want:  true,
		},
		{
			name:  "does not contain item",
			slice: []int{1, 2, 3, 4, 5},
			item:  6,
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []int{},
			item:  1,
			want:  false,
		},
		{
			name:  "single item match",
			slice: []int{42},
			item:  42,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollect_Timeout(t *testing.T) {
	ctx := context.Background()
	opts := Options{
		EnablePorts: false,
		NoAgent:     true,
		Timeout:     1 * time.Millisecond, // Very short timeout
	}

	_, err := Collect(ctx, "example.com", opts)
	if err == nil {
		t.Error("Collect() expected timeout error")
	}
}
