package game

import "testing"

// evolve_test.go — the engine half of #1805 in isolation: the token is
// canonical and cumulative, each instance on the ability list is one
// trigger (CR 702.100d), and the comparison is CR 702.100a's with
// CR 702.100c's noncreature rule. The rule end to end (entries, the
// re-check on resolution, grants, "evolves") is tested in
// cards/effects/evolve_test.go, where the catalog is wired.

func TestEvolveIsCanonicalAndCumulative(t *testing.T) {
	if kw, ok := CanonicalKeyword("Evolve"); !ok || kw != KeywordEvolve {
		t.Errorf("CanonicalKeyword(\"Evolve\") = %q, %v", kw, ok)
	}
	if !KeywordIsCumulative(KeywordEvolve) {
		t.Error("evolve is not cumulative; CR 702.100d says each instance triggers separately")
	}
	if got := AppendKeywordAbility([]string{KeywordEvolve}, KeywordEvolve); len(got) != 2 {
		t.Errorf("a second evolve grant was deduped: %v", got)
	}
}

func TestEvolveTriggersCountInstancesOnTheAbilityList(t *testing.T) {
	printed := Card{Name: "Cloudfin Raptor", TypeLine: "Creature — Bird Mutant", Keywords: []string{"flying", KeywordEvolve}}
	trigs := TriggersForCard(printed)
	if len(trigs) != 1 {
		t.Fatalf("a printed evolve with no catalog entry has %d triggers, want 1", len(trigs))
	}
	if trigs[0].Keyword != KeywordEvolve {
		t.Errorf("the trigger is named %q, want %q", trigs[0].Keyword, KeywordEvolve)
	}

	// Printed plus a layer-6 grant (Tyranid Prime) is two instances and
	// two triggers.
	eff := printed.printedCharacteristic()
	eff.Abilities = AppendKeywordAbility(eff.Abilities, KeywordEvolve)
	granted := printed
	granted.effective = &eff
	if n := EvolveCount(&granted); n != 2 {
		t.Errorf("printed + granted evolve = %d instances, want 2", n)
	}
	if n := len(TriggersForCard(granted)); n != 2 {
		t.Errorf("printed + granted evolve = %d triggers, want 2", n)
	}

	// Prowess and evolve on one object are counted apart.
	both := Card{Name: "Both", TypeLine: "Creature — Test", Keywords: []string{KeywordProwess, KeywordEvolve}}
	if ProwessCount(&both) != 1 || EvolveCount(&both) != 1 || len(TriggersForCard(both)) != 2 {
		t.Errorf("prowess %d, evolve %d, triggers %d; want 1, 1, 2",
			ProwessCount(&both), EvolveCount(&both), len(TriggersForCard(both)))
	}

	// Losing all abilities, or being face down, leaves no evolve.
	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d evolve triggers", n)
	}
	faceDown := printed
	faceDown.FaceDown, faceDown.FaceDownKind = true, FaceDownManifested
	if n := len(TriggersForCard(faceDown)); n != 0 {
		t.Errorf("a face-down creature has %d evolve triggers", n)
	}
}

// TestEvolveComparison is CR 702.100a — greater power OR greater
// toughness, each compared with its own kind — and CR 702.100c.
func TestEvolveComparison(t *testing.T) {
	cases := []struct {
		name            string
		enteredCreature bool
		ep, et          int
		selfCreature    bool
		sp, st          int
		want            bool
	}{
		{"greater power only", true, 3, 1, true, 2, 2, true},
		{"greater toughness only", true, 1, 3, true, 2, 2, true},
		{"both greater", true, 3, 3, true, 2, 2, true},
		{"equal", true, 2, 2, true, 2, 2, false},
		{"smaller", true, 1, 1, true, 2, 2, false},
		{"power is not compared with toughness", true, 3, 1, true, 3, 2, false},
		{"toughness is not compared with power", true, 1, 3, true, 2, 3, false},
		{"noncreature entering", false, 9, 9, true, 0, 1, false},
		{"noncreature evolving", true, 9, 9, false, 0, 1, false},
	}
	for _, c := range cases {
		if got := evolveComparisonHolds(c.enteredCreature, c.ep, c.et, c.selfCreature, c.sp, c.st); got != c.want {
			t.Errorf("%s: %d/%d entering vs %d/%d = %v, want %v", c.name, c.ep, c.et, c.sp, c.st, got, c.want)
		}
	}
}

// TestVanillaEvolveCreaturesNeedNoCatalogEntry — "When NOT to add a
// catalog entry" (docs/adding-cards.md): with evolve canonical, a
// creature whose only lines are evolve and other enforced keywords
// (Cloudfin Raptor, Shambleshark, …) prints no rule the engine leaves
// undone, so the importer does not flag it. A card with any other line
// (Fathom Mage) still is.
func TestVanillaEvolveCreaturesNeedNoCatalogEntry(t *testing.T) {
	const reminder = " (Whenever a creature you control enters, if that creature has greater power or toughness than this creature, put a +1/+1 counter on this creature.)"
	for _, text := range []string{
		"Flying\nEvolve" + reminder,
		"Flash (You may cast this spell any time you could cast an instant.)\nEvolve" + reminder,
		"Evolve" + reminder,
	} {
		if printsUnhandledRules(text) {
			t.Errorf("a keyword-only evolve creature reads as printing unhandled rules: %q", text)
		}
	}
	if !printsUnhandledRules("Evolve" + reminder + "\nWhenever a +1/+1 counter is put on this creature, you may draw a card.") {
		t.Error("Fathom Mage's second line was not flagged")
	}
}
