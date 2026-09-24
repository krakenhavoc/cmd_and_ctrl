package game

import "testing"

// prowess_test.go — the engine half of #706 in isolation: how many
// prowess triggers an object has, read off its ability list. The rule
// end to end (casts, pumps, cleanup, tokens, grants) is tested in
// cards/effects/prowess_test.go, where the catalog is wired.

func TestProwessTriggersCountInstancesOnTheAbilityList(t *testing.T) {
	printed := Card{Name: "Swiftspear", TypeLine: "Creature — Human Monk", Keywords: []string{"haste", KeywordProwess}}
	if n := len(TriggersForCard(printed)); n != 1 {
		t.Fatalf("a printed prowess with no catalog entry has %d triggers, want 1", n)
	}
	if got := TriggersForCard(printed)[0].Keyword; got != KeywordProwess {
		t.Errorf("the trigger is named %q, want %q (#1258)", got, KeywordProwess)
	}

	// Two instances on the effective list — a printed one plus a
	// layer-6 grant — are two triggers (CR 702.108b).
	eff := printed.printedCharacteristic()
	eff.Abilities = AppendKeywordAbility(eff.Abilities, KeywordProwess)
	granted := printed
	granted.effective = &eff
	if n := ProwessCount(&granted); n != 2 {
		t.Errorf("printed + granted prowess = %d instances, want 2", n)
	}
	if n := len(TriggersForCard(granted)); n != 2 {
		t.Errorf("printed + granted prowess = %d triggers, want 2", n)
	}

	// A creature that lost all abilities has an empty list and no
	// triggers; a face-down one has no text at all.
	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d prowess triggers", n)
	}
	faceDown := printed
	faceDown.FaceDown, faceDown.FaceDownKind = true, FaceDownManifested
	if n := len(TriggersForCard(faceDown)); n != 0 {
		t.Errorf("a face-down creature has %d prowess triggers", n)
	}
}

// TestProwessIsCumulativeAndCanonical — the two table facts the rest
// rests on: the importer admits the token, and a second grant is kept
// rather than deduped the way a second flying is.
func TestProwessIsCumulativeAndCanonical(t *testing.T) {
	if kw, ok := CanonicalKeyword("Prowess"); !ok || kw != KeywordProwess {
		t.Errorf("CanonicalKeyword(\"Prowess\") = %q, %v", kw, ok)
	}
	got := AppendKeywordAbility([]string{KeywordProwess}, KeywordProwess)
	if len(got) != 2 {
		t.Errorf("a second prowess grant was deduped: %v", got)
	}
	if got := AppendKeywordAbility([]string{"flying"}, "flying"); len(got) != 1 {
		t.Errorf("a second flying grant was kept: %v", got)
	}
}

// TestPrintedProwessTwiceIsTwoInstances — #1510, CR 702.108b: a card
// that PRINTS prowess twice ("Prowess, prowess" — Thor Odinson, Ruric
// Thar, Biomagus) has two instances, whether or not the object has a
// catalog entry. The deck importer's half of this (counting the
// repeat off the oracle line rather than taking Scryfall's `keywords`
// set at its word) is tested in deck.TestRepeatedPrintedCumulativeKeywordImportsAsTwoInstances;
// this is the engine's half — printedCharacteristic's merge of
// Card.Keywords with the catalog's own PrintedKeywords must not
// collapse a cumulative keyword's repeats back down to one the way it
// legitimately collapses a repeated "flying".
func TestPrintedProwessTwiceIsTwoInstances(t *testing.T) {
	// Off the battlefield, a card whose OWN Keywords already carries
	// prowess twice (what the deck importer now stamps for a doubled
	// oracle line) reports two triggers with no catalog entry at all.
	printedTwice := Card{
		Name:     "Thor Odinson",
		TypeLine: "Legendary Creature — God",
		Keywords: []string{"flying", "vigilance", KeywordProwess, KeywordProwess},
	}
	if n := len(TriggersForCard(printedTwice)); n != 2 {
		t.Fatalf("a card printing prowess twice has %d triggers, want 2", n)
	}
	if !HasKeyword(&printedTwice, "flying") || !HasKeyword(&printedTwice, "vigilance") {
		t.Errorf("the ordinary printed keywords were lost: %v", printedTwice.Keywords)
	}

	// On the battlefield: printedCharacteristic's merge (layer 0) is
	// what Effective() starts from, and it must keep both instances
	// rather than deduping the second "prowess" against the first the
	// way containsKeyword would have before #1510's mergePrintedKeywords.
	eff := printedTwice.printedCharacteristic()
	onBattlefield := printedTwice
	onBattlefield.effective = &eff
	if n := ProwessCount(&onBattlefield); n != 2 {
		t.Errorf("printedCharacteristic merge: %d instances of prowess, want 2 (%v)", n, eff.Abilities)
	}

	// A catalog entry that ALSO declares prowess once must not turn
	// the import's two into three — the merge takes the higher of the
	// two sources' own counts, never their sum.
	merged := mergePrintedKeywords([]string{KeywordProwess}, []string{"flying", "vigilance", KeywordProwess, KeywordProwess})
	count := 0
	for _, kw := range merged {
		if kw == KeywordProwess {
			count++
		}
	}
	if count != 2 {
		t.Errorf("catalog prowess x1 + imported prowess x2 merged to %d, want 2 (%v)", count, merged)
	}
}
