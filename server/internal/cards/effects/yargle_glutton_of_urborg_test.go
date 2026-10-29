package effects

import "testing"

// A vanilla body is registered with no abilities (#4453).
func TestYargleGluttonOfUrborgIsARegisteredVanillaBody(t *testing.T) {
	spec, ok := Lookup("ed66cd31-958f-4b28-82a3-e04acc819afc")
	if !ok {
		t.Fatal("Yargle, Glutton of Urborg is not registered")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("completeness %s, want Full", spec.Completeness)
	}
	if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
		t.Error("a vanilla body declares no abilities")
	}
}
