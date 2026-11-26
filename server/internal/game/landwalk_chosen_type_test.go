package game

import "testing"

// TestGrantedLandwalkOfAChosenType is Traveler's Cloak's (S82 OLD batch 362,
// #5557): a granted "<land type>walk" for a type outside the table
// ("gatewalk") is read like the table's own tokens (CR 702.14c).
func TestGrantedLandwalkOfAChosenType(t *testing.T) {
	if got := LandwalkToken("Island"); got != "islandwalk" {
		t.Fatalf("LandwalkToken(Island) = %q", got)
	}
	if got := LandwalkToken("Gate"); got != "gatewalk" {
		t.Fatalf("LandwalkToken(Gate) = %q", got)
	}
	if got := LandwalkToken("Goblin"); got != "" {
		t.Fatalf("LandwalkToken(Goblin) = %q, want \"\"", got)
	}
	g := newActiveGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	walker := pushWalker(g, me, "Walker")
	blocker := pushWalker(g, them, "Blocker")
	pushLand(g, them, "Gateway", "Land — Gate")
	registerScopedEffectForTest(t, g, walker, []Mod{AddKeywordsMod(LandwalkToken("Gate"))}, g.UntilEndOfTurnDuration())
	attackAndStepToBlocks(t, g, them, walker)
	if r := pairRefusal(g, walker, blocker); r.Reason != BlockReasonLandwalk {
		t.Fatalf("granted gatewalk vs a Gate: reason = %q, want landwalk", r.Reason)
	}
	s, ok := landwalkRequirement("gatewalk")
	if !ok || s.describe() != "a Gate" {
		t.Errorf("gatewalk describes as %q (ok %v)", s.describe(), ok)
	}
	for _, tok := range []string{"goblinwalk", "walk", "snowwalk"} {
		if _, ok := landwalkRequirement(tok); ok {
			t.Errorf("%q parsed as landwalk", tok)
		}
	}
	found := false
	for _, tok := range LandwalkTokens() {
		if tok == "gatewalk" {
			found = true
		}
	}
	if !found {
		t.Error("LandwalkTokens omits gatewalk")
	}
}
