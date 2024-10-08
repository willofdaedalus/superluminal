package server

import "testing"

func TestGetIPAdrr(t *testing.T) {
	expected := "192.168.1.204"
	got, _ := getIpAddr()

	if got != expected {
		t.Fatalf("expected %s got %s\n", expected, got)
	}
}
