package game

import "testing"

// infect_wither_toxic_test.go pins the pure half of ADR 0056: the
// token grammar, the cumulative-append rule and the two result
// functions. The engine-level tests — the damage tail actually placing
// the counters — are in infect_wither_toxic_tail_test.go.
//
// Card fixtures come from cardWithAbilities (keywords_test.go), which
// is the battlefield shape: a populated `effective`, so the tokens
// read the way a granted keyword does.

// TestToxicValueGrammar pins CR 702.164's numbered keyword as ADR
// 0056 Decision 1 spells it. The refusals matter more than the
// accepts: every one of them is a string Scryfall can hand the
// importer, and each has to leave the card flagged unimplemented
// rather than stamped with a number nobody wrote.
func TestToxicValueGrammar(t *testing.T) {
	for in, want := range map[string]int{
		"toxic 1":   1,
		"toxic 4":   4,
		"Toxic 2":   2,
		"toxic 01":  1,
		" toxic 3 ": 3,
		"TOXIC 10":  10,
	} {
		got, ok := ToxicValue(in)
		if !ok || got != want {
			t.Errorf("ToxicValue(%q) = (%d, %v), want (%d, true)", in, got, ok, want)
		}
	}
	// A bare "toxic" is Scryfall's `keywords` array entry, and it
	// names no amount: accepting it would stamp a badge with nothing
	// behind it, the call ADR 0038 §6 made for "Hexproof from" and
	// keywords.go makes for a bare "protection".
	for _, in := range []string{
		"toxic",
		"toxic ",
		"toxic 0",
		"toxic 00",
		"toxic -1",
		"toxic +1",
		"toxic two",
		"toxic  2",
		"toxic 2 1",
		"toxic2",
		"toxic 1x",
		"toxic 1000",
		"infect",
		"",
	} {
		if got, ok := ToxicValue(in); ok {
			t.Errorf("ToxicValue(%q) = (%d, true), want not-toxic", in, got)
		}
	}
}

// TestCanonicalToxicTokenNormalises covers the minting half: what the
// importer will stamp into Characteristic.Abilities once the token
// joins the table. "Toxic 01" and "toxic 1" have to be the SAME
// string, or ToxicTotal counts them twice and a card grows a second
// badge.
func TestCanonicalToxicTokenNormalises(t *testing.T) {
	for in, want := range map[string]string{
		"Toxic 1":  "toxic 1",
		"toxic 01": "toxic 1",
		"TOXIC 4":  "toxic 4",
		" toxic 2": "toxic 2",
	} {
		got, ok := CanonicalToxicToken(in)
		if !ok || got != want {
			t.Errorf("CanonicalToxicToken(%q) = (%q, %v), want (%q, true)", in, got, ok, want)
		}
	}
	for _, in := range []string{"toxic", "Toxic", "toxic 0", "flying", ""} {
		if got, ok := CanonicalToxicToken(in); ok {
			t.Errorf("CanonicalToxicToken(%q) = (%q, true), want not-toxic", in, got)
		}
	}
}

// TestToxicTotalSumsEveryInstance is CR 702.164b: "a creature's total
// toxic value is the sum of all N values of toxic abilities that
// creature has". This is the one place in the engine where a repeated
// keyword means twice as much, and it is why toxic is read through
// this helper and never through HasKeyword.
func TestToxicTotalSumsEveryInstance(t *testing.T) {
	// A Rat that prints toxic 1 under Karumonix's "other Rats you
	// control have toxic 1": two instances, total 2.
	granted := cardWithAbilities("toxic 1", "flying", "toxic 1")
	if got := ToxicTotal(granted); got != 2 {
		t.Errorf("ToxicTotal(printed + granted toxic 1) = %d, want 2", got)
	}
	// Tyrranax Rex: one instance of toxic 4, alongside keywords that
	// are not toxic at all.
	rex := cardWithAbilities("trample", "haste", "toxic 4")
	if got := ToxicTotal(rex); got != 4 {
		t.Errorf("ToxicTotal(toxic 4) = %d, want 4", got)
	}
	// Ixhel's toxic 2 plus a granted toxic 1.
	if got := ToxicTotal(cardWithAbilities("toxic 2", "toxic 1")); got != 3 {
		t.Errorf("ToxicTotal(toxic 2 + toxic 1) = %d, want 3", got)
	}
	// No toxic at all, and a malformed line that names no amount.
	if got := ToxicTotal(cardWithAbilities("infect", "deathtouch", "toxic")); got != 0 {
		t.Errorf("ToxicTotal(no numbered toxic) = %d, want 0", got)
	}
	if got := ToxicTotal(nil); got != 0 {
		t.Errorf("ToxicTotal(nil) = %d, want 0", got)
	}
	// Off the battlefield the card's own printed keywords answer, the
	// same fallback HasKeyword documents. A toxic creature in a
	// graveyard is what a "return it to the battlefield" effect is
	// about to put back.
	inHand := &Card{Keywords: []string{"toxic 1", "toxic 2"}}
	if got := ToxicTotal(inHand); got != 3 {
		t.Errorf("ToxicTotal(off battlefield) = %d, want 3", got)
	}
}

// TestAppendKeywordAbilityCumulativeVsRedundant pins ADR 0056
// Decision 1's split: a second haste is nothing (CR 702.10b and the
// rest of the redundant table), a second toxic is a bigger number
// (CR 702.164b). The catalog's grant helpers dedupe unconditionally
// today, which is correct for every keyword that exists and wrong for
// the first one that doesn't.
func TestAppendKeywordAbilityCumulativeVsRedundant(t *testing.T) {
	// Redundant: a second grant of the same token adds no badge.
	abilities := []string{"haste"}
	abilities = AppendKeywordAbility(abilities, "haste")
	if len(abilities) != 1 {
		t.Errorf("appending a second haste = %v, want one entry", abilities)
	}
	abilities = AppendKeywordAbility(abilities, KeywordInfect)
	abilities = AppendKeywordAbility(abilities, KeywordInfect)
	if got := len(abilities); got != 2 {
		t.Errorf("abilities = %v, want haste + one infect", abilities)
	}

	// Cumulative: every toxic instance stays, and ToxicTotal counts
	// them all.
	tox := AppendKeywordAbility(nil, "toxic 1")
	tox = AppendKeywordAbility(tox, "toxic 1")
	tox = AppendKeywordAbility(tox, "toxic 2")
	if len(tox) != 3 {
		t.Fatalf("toxic grants = %v, want three entries", tox)
	}
	if got := ToxicTotal(cardWithAbilities(tox...)); got != 4 {
		t.Errorf("ToxicTotal after three toxic grants = %d, want 4", got)
	}

	// An empty token is dropped rather than rendered as a blank badge.
	if got := AppendKeywordAbility([]string{"flying"}, ""); len(got) != 1 {
		t.Errorf("appending an empty token = %v, want it dropped", got)
	}
}

// TestSourceDamageResultTraits covers the reader: what the damage
// tail will snapshot off the source, beside deathtouch and lifelink.
func TestSourceDamageResultTraits(t *testing.T) {
	glistenerElf := SourceDamageResultTraits(cardWithAbilities("infect"))
	if !glistenerElf.Infect || glistenerElf.Wither || glistenerElf.ToxicTotal != 0 {
		t.Errorf("infect source = %+v, want infect only", glistenerElf)
	}
	if !glistenerElf.Any() {
		t.Error("an infect source must report Any()")
	}

	ramGang := SourceDamageResultTraits(cardWithAbilities("wither", "haste"))
	if ramGang.Infect || !ramGang.Wither {
		t.Errorf("wither source = %+v, want wither only", ramGang)
	}

	rex := SourceDamageResultTraits(cardWithAbilities("trample", "toxic 4"))
	if rex.ToxicTotal != 4 || rex.Infect || rex.Wither {
		t.Errorf("toxic 4 source = %+v, want ToxicTotal 4", rex)
	}

	// An ordinary creature changes nothing about its damage, and a
	// source the engine can no longer see carries no keywords — the
	// same answer deathtouch and lifelink give for a departed source
	// today.
	plain := SourceDamageResultTraits(cardWithAbilities("flying", "deathtouch", "lifelink"))
	if plain.Any() {
		t.Errorf("plain source = %+v, want no result keywords", plain)
	}
	if SourceDamageResultTraits(nil).Any() {
		t.Error("a source the engine cannot see must carry no result keywords")
	}
}

// TestDamageResultOnCreature is CR 120.3d against CR 120.3e: counters
// or marked damage, never both, and nothing at all when the damage
// was prevented.
func TestDamageResultOnCreature(t *testing.T) {
	for _, tc := range []struct {
		name         string
		src          DamageResultSource
		amount       int
		wantCounters int
		wantMarked   int
	}{
		{"plain source marks damage", DamageResultSource{}, 3, 0, 3},
		{"infect puts -1/-1 counters", DamageResultSource{Infect: true}, 2, 2, 0},
		{"wither puts -1/-1 counters", DamageResultSource{Wither: true}, 2, 2, 0},
		// CR 702.90 and CR 702.80 name the same result. A source with
		// both does not get two sets.
		{"infect and wither together", DamageResultSource{Infect: true, Wither: true}, 4, 4, 0},
		// Toxic does nothing to a creature (the Pestilent Syphoner
		// and Karumonix rulings are explicit).
		{"toxic alone still marks damage", DamageResultSource{ToxicTotal: 3}, 2, 0, 2},
		{"infect with toxic still one set", DamageResultSource{Infect: true, ToxicTotal: 3}, 2, 2, 0},
		// Prevented damage has no result: no counters, no mark.
		{"prevented infect damage", DamageResultSource{Infect: true}, 0, 0, 0},
		{"negative amount", DamageResultSource{}, -2, 0, 0},
	} {
		counters, marked := tc.src.DamageToCreature(tc.amount)
		if counters != tc.wantCounters || marked != tc.wantMarked {
			t.Errorf("%s: DamageToCreature(%d) = (%d counters, %d marked), want (%d, %d)",
				tc.name, tc.amount, counters, marked, tc.wantCounters, tc.wantMarked)
		}
	}
}

// TestDamageResultOnPlayer is CR 120.3a / 120.3b / 120.3g: infect
// replaces the life loss with poison, toxic adds poison on top of it
// and only in combat, and wither does neither.
func TestDamageResultOnPlayer(t *testing.T) {
	for _, tc := range []struct {
		name       string
		src        DamageResultSource
		amount     int
		combat     bool
		wantPoison int
		wantLife   int
	}{
		{"plain combat damage", DamageResultSource{}, 3, true, 0, 3},
		{"infect instead of life", DamageResultSource{Infect: true}, 3, true, 3, 0},
		{"infect out of combat", DamageResultSource{Infect: true}, 2, false, 2, 0},
		// Wither says nothing about players (CR 702.80 is about
		// creatures), so Puncture Blast to the face is life loss.
		{"wither to a player is life loss", DamageResultSource{Wither: true}, 3, true, 0, 3},
		// Toxic is IN ADDITION to the damage, and combat only.
		{"toxic 1 in combat", DamageResultSource{ToxicTotal: 1}, 2, true, 1, 2},
		{"toxic granted twice sums", DamageResultSource{ToxicTotal: 2}, 2, true, 2, 2},
		{"toxic out of combat gives nothing", DamageResultSource{ToxicTotal: 3}, 2, false, 0, 2},
		// One placement, not two: the sum is what a halving or "+1"
		// replacement gets to see (ADR 0056 Decision 4).
		{"infect and toxic are one number", DamageResultSource{Infect: true, ToxicTotal: 1}, 2, true, 3, 0},
		// Prevented damage gives no poison — the Grafted Exoskeleton
		// ruling, and the reason the caller returns before this point
		// when nothing landed.
		{"prevented", DamageResultSource{Infect: true, ToxicTotal: 2}, 0, true, 0, 0},
	} {
		poison, life := tc.src.DamageToPlayer(tc.amount, tc.combat)
		if poison != tc.wantPoison || life != tc.wantLife {
			t.Errorf("%s: DamageToPlayer(%d, combat=%v) = (%d poison, %d life), want (%d, %d)",
				tc.name, tc.amount, tc.combat, poison, life, tc.wantPoison, tc.wantLife)
		}
	}
}

// TestToxicIsNotScaledByTheDamageAmount is the Pestilent Syphoner
// ruling as its own case, because it is the one thing about toxic a
// future doubler or halver could plausibly get wrong: "the counter
// count doesn't change when a damage replacement changes the damage".
func TestToxicIsNotScaledByTheDamageAmount(t *testing.T) {
	src := DamageResultSource{ToxicTotal: 1}
	doubled, doubledLife := src.DamageToPlayer(4, true)
	single, singleLife := src.DamageToPlayer(2, true)
	if doubled != 1 || single != 1 {
		t.Errorf("toxic 1 gave %d poison on 4 damage and %d on 2, want 1 and 1", doubled, single)
	}
	if doubledLife != 4 || singleLife != 2 {
		t.Errorf("life loss = %d and %d, want 4 and 2: toxic must not change the damage", doubledLife, singleLife)
	}
}
