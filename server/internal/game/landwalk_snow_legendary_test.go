package game

import "testing"

// landwalk_snow_legendary_test.go — #6143: snow landwalk, legendary
// landwalk, snow <land type>walk and the granted "snow landwalk of the
// chosen type" family.
//
// CR 702.14c: "A creature with landwalk can't be blocked as long as the
// defending player controls at least one land with the specified land
// type (as in "islandwalk"), with the specified type or supertype (as
// in "artifact landwalk"), without the specified type or supertype (as
// in "nonbasic landwalk"), or with both the specified type or supertype
// and the specified subtype (as in "snow swampwalk")."

// TestSnowForestwalkNeedsASnowForest is Rime Dryad's: a non-snow Forest
// or a snow land of another type does not switch it on; a snow Forest
// does.
func TestSnowForestwalkNeedsASnowForest(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	dryad := pushWalker(g, me, "Rime Dryad", "snow forestwalk")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Forest", "Basic Land — Forest")
	pushLand(g, them, "Snow-Covered Island", "Basic Snow Land — Island")
	pushLand(g, me, "Snow-Covered Forest", "Basic Snow Land — Forest")
	attackAndStepToBlocks(t, g, them, dryad)

	if r := pairRefusal(g, dryad, blocker); !r.Legal() {
		t.Fatalf("defender has only a non-snow Forest and a snow Island: refused with %q", r.Reason)
	}
	pushLand(g, them, "Snow-Covered Forest", "Basic Snow Land — Forest")
	r := pairRefusal(g, dryad, blocker)
	if r.Reason != BlockReasonLandwalk || r.Source != dryad {
		t.Fatalf("defender controls a snow Forest: refusal = %+v, want landwalk", r)
	}
}

// TestSnowLandwalkReadsOnlyTheSupertype is Zombie Musher's: any snow
// land, whatever its type.
func TestSnowLandwalkReadsOnlyTheSupertype(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	musher := pushWalker(g, me, "Zombie Musher", "snow landwalk")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Swamp", "Basic Land — Swamp")
	attackAndStepToBlocks(t, g, them, musher)
	if r := pairRefusal(g, musher, blocker); !r.Legal() {
		t.Fatalf("no snow land: refused with %q", r.Reason)
	}
	pushLand(g, them, "Dark Depths", "Legendary Snow Land")
	if r := pairRefusal(g, musher, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("a snow land: reason = %q, want landwalk", r.Reason)
	}
}

// TestLegendaryLandwalkReadsTheLegendarySupertype is Livonya Silone's
// and Ayumi's.
func TestLegendaryLandwalkReadsTheLegendarySupertype(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	livonya := pushWalker(g, me, "Livonya Silone", "first strike", "legendary landwalk")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Island", "Basic Land — Island")
	pushLand(g, me, "Urborg, Tomb of Yawgmoth", "Legendary Land")
	attackAndStepToBlocks(t, g, them, livonya)
	if r := pairRefusal(g, livonya, blocker); !r.Legal() {
		t.Fatalf("no legendary land under the defender: refused with %q", r.Reason)
	}
	pushLand(g, them, "Mikokoro, Center of the Sea", "Legendary Land")
	if r := pairRefusal(g, livonya, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("a legendary land: reason = %q, want landwalk", r.Reason)
	}
}

// TestGrantedSnowLandwalkOfAChosenTypeLapses is Barbarian Guides': a
// granted "snow islandwalk" switches on against a snow Island until end
// of turn and is gone after; a chosen nonbasic type ("snow gatewalk")
// is read the same way through the parameterised token.
func TestGrantedSnowLandwalkOfAChosenTypeLapses(t *testing.T) {
	if got := SnowLandwalkToken("Island"); got != "snow islandwalk" {
		t.Fatalf("SnowLandwalkToken(Island) = %q", got)
	}
	if got := SnowLandwalkToken("Goblin"); got != "" {
		t.Fatalf("SnowLandwalkToken(Goblin) = %q, want \"\" (not a land type)", got)
	}
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	guided := pushWalker(g, me, "Guided")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Snow-Covered Island", "Basic Snow Land — Island")
	pushLand(g, them, "Snow Gate", "Snow Land — Gate")
	registerScopedEffectForTest(t, g, guided, []Mod{AddKeywordsMod(SnowLandwalkToken("Island"))}, g.UntilEndOfTurnDuration())
	attackAndStepToBlocks(t, g, them, guided)
	if r := pairRefusal(g, guided, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("granted snow islandwalk vs a snow Island: reason = %q, want landwalk", r.Reason)
	}
	g.WithWriteLock(func() {
		g.Turn.Seq++
		g.ClearEndOfTurnScopedStaticsLocked()
	})
	if r := pairRefusal(g, guided, blocker); !r.Legal() {
		t.Fatalf("after end of turn the grant lapsed: still refused with %q", r.Reason)
	}

	registerScopedEffectForTest(t, g, guided, []Mod{AddKeywordsMod(SnowLandwalkToken("Gate"))}, g.UntilEndOfTurnDuration())
	r := pairRefusal(g, guided, blocker)
	if r.Reason != BlockReasonLandwalk {
		t.Fatalf("granted snow gatewalk vs a snow Gate: reason = %q, want landwalk", r.Reason)
	}
	if !IsLandwalkToken("snow gatewalk") {
		t.Error("snow gatewalk is not read as a landwalk token")
	}
}

// TestSnowAndLegendaryLandwalkDescribe pins the illegal_block land.
func TestSnowAndLegendaryLandwalkDescribe(t *testing.T) {
	for tok, want := range map[string]string{
		"snow landwalk": "a snow land", "legendary landwalk": "a legendary land",
		"snow swampwalk": "a snow Swamp", "snow gatewalk": "a snow Gate",
	} {
		s, ok := landwalkRequirement(tok)
		if !ok || s.describe() != want {
			t.Errorf("%s describes as %q (ok %v), want %q", tok, s.describe(), ok, want)
		}
	}
	for _, tok := range []string{"snow goblinwalk", "legendary swampwalk", "snowwalk"} {
		if _, ok := landwalkRequirement(tok); ok {
			t.Errorf("%q parsed as landwalk", tok)
		}
	}
	all := LandwalkTokens()
	// The table, every non-basic land type's snow token, and (Traveler's
	// Cloak) the plain "<type>walk" of every type the table does not name:
	// the five basics and Desert are in it already.
	if len(all) != len(landwalkTokens)+len(LandTypes)-5+len(LandTypes)-6 {
		t.Errorf("LandwalkTokens has %d, want the table plus every non-basic land type's snow token and the plain tokens outside the table", len(all))
	}
}
