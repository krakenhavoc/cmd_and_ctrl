package game

import "testing"

// TestCardTypeHelpers covers the new S13.1 type-line predicates that
// the cast / resolve path uses to route spells. The helpers all do
// case-insensitive substring matches against TypeLine; an empty type
// line returns false everywhere so demo-seed cards default to "no
// type" rather than spuriously matching.
func TestCardTypeHelpers(t *testing.T) {
	cases := []struct {
		name        string
		typeLine    string
		isLand      bool
		isInstant   bool
		isSorcery   bool
		isArtifact  bool
		isEnchant   bool
		isPlanesw   bool
		isBattle    bool
		isCreature  bool
		isPermanent bool
	}{
		{
			name:        "Land",
			typeLine:    "Land",
			isLand:      true,
			isPermanent: true,
		},
		{
			name:        "Basic Land",
			typeLine:    "Basic Land — Forest",
			isLand:      true,
			isPermanent: true,
		},
		{
			name:        "Creature",
			typeLine:    "Creature — Human Wizard",
			isCreature:  true,
			isPermanent: true,
		},
		{
			name:        "Legendary Artifact Creature",
			typeLine:    "Legendary Artifact Creature — Golem",
			isCreature:  true,
			isArtifact:  true,
			isPermanent: true,
		},
		{
			name:        "Planeswalker",
			typeLine:    "Legendary Planeswalker — Jace",
			isPlanesw:   true,
			isPermanent: true,
		},
		{
			name:        "Battle",
			typeLine:    "Battle — Siege",
			isBattle:    true,
			isPermanent: true,
		},
		{
			name:        "Enchantment",
			typeLine:    "Enchantment",
			isEnchant:   true,
			isPermanent: true,
		},
		{
			name:        "Aura",
			typeLine:    "Enchantment — Aura",
			isEnchant:   true,
			isPermanent: true,
		},
		{
			name:      "Instant",
			typeLine:  "Instant",
			isInstant: true,
		},
		{
			name:      "Sorcery",
			typeLine:  "Sorcery",
			isSorcery: true,
		},
		{
			name:     "Empty type line",
			typeLine: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			card := Card{TypeLine: c.typeLine}
			if got := card.IsLand(); got != c.isLand {
				t.Errorf("IsLand(%q): got %v, want %v", c.typeLine, got, c.isLand)
			}
			if got := card.IsInstant(); got != c.isInstant {
				t.Errorf("IsInstant(%q): got %v, want %v", c.typeLine, got, c.isInstant)
			}
			if got := card.IsSorcery(); got != c.isSorcery {
				t.Errorf("IsSorcery(%q): got %v, want %v", c.typeLine, got, c.isSorcery)
			}
			if got := card.IsArtifact(); got != c.isArtifact {
				t.Errorf("IsArtifact(%q): got %v, want %v", c.typeLine, got, c.isArtifact)
			}
			if got := card.IsEnchantment(); got != c.isEnchant {
				t.Errorf("IsEnchantment(%q): got %v, want %v", c.typeLine, got, c.isEnchant)
			}
			if got := card.IsPlaneswalker(); got != c.isPlanesw {
				t.Errorf("IsPlaneswalker(%q): got %v, want %v", c.typeLine, got, c.isPlanesw)
			}
			if got := card.IsBattle(); got != c.isBattle {
				t.Errorf("IsBattle(%q): got %v, want %v", c.typeLine, got, c.isBattle)
			}
			if got := card.IsCreature(); got != c.isCreature {
				t.Errorf("IsCreature(%q): got %v, want %v", c.typeLine, got, c.isCreature)
			}
			if got := card.IsPermanent(); got != c.isPermanent {
				t.Errorf("IsPermanent(%q): got %v, want %v", c.typeLine, got, c.isPermanent)
			}
		})
	}
}

// TestCardTypeHelpersCaseInsensitive verifies the substring match
// folds case so weird capitalisations from third-party data sources
// (cockatrice, scryfall variants) still classify correctly.
func TestCardTypeHelpersCaseInsensitive(t *testing.T) {
	for _, tl := range []string{"INSTANT", "instant", "Instant", "iNsTaNt"} {
		c := Card{TypeLine: tl}
		if !c.IsInstant() {
			t.Errorf("IsInstant(%q) should be true", tl)
		}
	}
}
