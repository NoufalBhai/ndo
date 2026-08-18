package config

import "testing"

func TestSettingsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Settings{Shell: "bash -c", Editor: "vim", Color: true}

	if err := SaveSettings(dir, want); err != nil {
		t.Fatalf("SaveSettings() error: %v", err)
	}
	got, err := LoadSettings(dir)
	if err != nil {
		t.Fatalf("LoadSettings() error: %v", err)
	}
	if got != want {
		t.Fatalf("LoadSettings() = %+v, want %+v", got, want)
	}
}

func TestLoadSettingsMissingFileIsNotError(t *testing.T) {
	dir := t.TempDir()
	got, err := LoadSettings(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != (Settings{}) {
		t.Fatalf("LoadSettings() = %+v, want zero value", got)
	}
}
