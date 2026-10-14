package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_source_choice_test.go is #3604: which source the auto-tapper
// spends on a generic cost, and which land a filter land is fed from.
// Three agents at the table reported the same family of wrong picks:
// a Sol Ring tapped for a lone {1}, a Forest tapped for Skullclamp's
// {1} with a green spell still in hand, and a filter land that tapped a
// further land for a second {W} it could have made itself.

const solRingOracle = "6ad8011d-3471-4369-9d68-b264cc027487"

func sourceChoiceHook(oracleID string) []ManaAbilityShape {
	switch oracleID {
	case solRingOracle:
		return solRingHook(oracleID)
	case "zz-3604-foundry": // a shockland: {T}: Add {R} or {W}
		return []ManaAbilityShape{{TapCost: true, Produced: "{R|W}", Label: "Add {R} or {W}"}}
	case "zz-3604-prairie": // Rugged Prairie
		return []ManaAbilityShape{
			{TapCost: true, Produced: "{C}", Label: "Add {C}"},
			{TapCost: true, ManaCost: "{R/W}", Produced: "{R|W}{R|W}", Label: "{R/W}, {T}: Add {R}{R}, {R}{W}, or {W}{W}"},
		}
	}
	return nil
}

// choiceTable is an empty-handed game at seat 0's main phase.
func choiceTable(t *testing.T) (*Game, *Player) {
	t.Helper()
	withCatalogHook(t, sourceChoiceHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	p.Hand.Cards = nil
	return g, p
}

func planOf(t *testing.T, g *Game, p *Player, cost string) map[uuid.UUID]bool {
	t.Helper()
	plan, ok := g.AutoTapForCost(p.ID, costFor(t, cost), 0)
	if !ok {
		t.Fatalf("no plan for %s", cost)
	}
	return planSet(plan)
}

// Sol Ring for a lone {1} floats a {C} nobody asked for. With an Island
// beside it, the Island pays exactly and Sol Ring stays up for the next
// spell.
func TestAutoTapGenericPrefersASourceItUsesInFull(t *testing.T) {
	g, p := choiceTable(t)
	ring := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", solRingOracle)
	island := pushBattlefieldForTest(g, p.ID, "Island", "Basic Land — Island", "")

	got := planOf(t, g, p, "{1}")
	if !got[island] || got[ring] {
		t.Errorf("plan = %v, want the Island (%v) and not Sol Ring (%v)", got, island, ring)
	}
}

// The reported Abrade case: {1}{R} with a Mountain, a Sol Ring and a
// spare Island taps the Mountain and the Island, not Sol Ring.
func TestAutoTapAbradeDoesNotWasteSolRing(t *testing.T) {
	g, p := choiceTable(t)
	mountain := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	ring := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", solRingOracle)
	island := pushBattlefieldForTest(g, p.ID, "Island", "Basic Land — Island", "")

	got := planOf(t, g, p, "{1}{R}")
	if !got[mountain] || !got[island] || got[ring] {
		t.Errorf("plan = %v, want Mountain (%v) + Island (%v), not Sol Ring (%v)", got, mountain, island, ring)
	}
}

// The preference is a tiebreak and never a refusal: Sol Ring alone still
// pays a {1}, and pays a {2} in full with lands beside it.
func TestAutoTapSolRingStillPaysWhenItFits(t *testing.T) {
	g, p := choiceTable(t)
	ring := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", solRingOracle)
	if got := planOf(t, g, p, "{1}"); !got[ring] {
		t.Errorf("Sol Ring alone must pay {1}, plan = %v", got)
	}

	island := pushBattlefieldForTest(g, p.ID, "Island", "Basic Land — Island", "")
	pushBattlefieldForTest(g, p.ID, "Island", "Basic Land — Island", "")
	got := planOf(t, g, p, "{2}")
	if !got[ring] || got[island] || len(got) != 1 {
		t.Errorf("{2} should be Sol Ring alone, plan = %v", got)
	}
}

// The fit preference stays inside its group: a Treasure is a last-resort
// tier and is not cracked to save a Sol Ring's spare mana.
func TestAutoTapFitDoesNotCrackATreasure(t *testing.T) {
	g, p := choiceTable(t)
	ring := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", solRingOracle)
	treasure := pushTreasures(g, p, 1)[0]

	got := planOf(t, g, p, "{1}")
	if !got[ring] || got[treasure] {
		t.Errorf("plan = %v, want Sol Ring (%v), not the Treasure (%v)", got, ring, treasure)
	}
}

// Omnath: a Forest paying Skullclamp's {1} leaves no green for the second
// green spell. With a green spell in hand the Mountain pays; with a red
// one the Forest does.
func TestAutoTapGenericSparesTheColourTheHandNeeds(t *testing.T) {
	for _, tc := range []struct {
		name       string
		handCost   string
		wantForest bool
	}{
		{"a green spell waits", "{G}{G}", false},
		{"a red spell waits", "{1}{R}", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, p := choiceTable(t)
			forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
			mountain := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
			pushTypedCardToHandWithCost(p, "Waiting", "Instant", tc.handCost)

			got := planOf(t, g, p, "{1}")
			if got[forest] != tc.wantForest || got[mountain] == tc.wantForest {
				t.Errorf("plan = %v, forest=%v mountain=%v, want forest paid=%v", got, forest, mountain, tc.wantForest)
			}
		})
	}
}

// A commander in the command zone is part of "what the hand needs".
func TestAutoTapGenericSparesTheCommandersColour(t *testing.T) {
	g, p := choiceTable(t)
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	mountain := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	cmd := NewCard("Commander", p.ID)
	cmd.TypeLine = "Legendary Creature"
	cmd.ManaCost = "{2}{G}{G}"
	p.Command.PushTop(cmd)

	got := planOf(t, g, p, "{1}")
	if got[forest] || !got[mountain] {
		t.Errorf("plan = %v, want the Mountain (%v) so green (%v) stays up for the commander", got, mountain, forest)
	}
}

// Demand outranks fit: with a green spell waiting, a lone {1} is still
// paid by Sol Ring (one {C} floats) rather than by the Forest.
func TestAutoTapDemandOutranksFit(t *testing.T) {
	g, p := choiceTable(t)
	ring := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", solRingOracle)
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	pushTypedCardToHandWithCost(p, "Waiting", "Instant", "{G}")

	got := planOf(t, g, p, "{1}")
	if !got[ring] || got[forest] {
		t.Errorf("plan = %v, want Sol Ring (%v), not the Forest (%v)", got, ring, forest)
	}
}

// The reported Puresteel Paladin case: {W}{W} with a shockland, a filter
// land and a Mountain. The Mountain feeds the filter land, which makes
// {W}{W} itself; the shockland stays untapped and nothing floats.
func TestAutoTapFilterLandMakesBothOfTheSpellsPips(t *testing.T) {
	g, p := choiceTable(t)
	foundry := pushBattlefieldForTest(g, p.ID, "Sacred Foundry", "Land — Mountain Plains", "zz-3604-foundry")
	prairie := pushBattlefieldForTest(g, p.ID, "Rugged Prairie", "Land — Mountain Plains", "zz-3604-prairie")
	mountain := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")

	id := pushTypedCardToHandWithCost(p, "Puresteel Paladin", "Creature", "{W}{W}")
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	if !tappedForTest(g, prairie) || !tappedForTest(g, mountain) {
		t.Error("want Rugged Prairie and the Mountain tapped")
	}
	if tappedForTest(g, foundry) {
		t.Error("Sacred Foundry was tapped; the filter land should have made both {W}")
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool = %v, want nothing floating", p.ManaPool)
	}
}

// The case the reporter said already worked stays so: {1}{W}{W} with the
// filter land, a Mountain and two Plains is the three lands that make it,
// with nothing left over.
func TestAutoTapFilterLandBesideTwoPlainsStillPaysWithoutWaste(t *testing.T) {
	g, p := choiceTable(t)
	pushBattlefieldForTest(g, p.ID, "Rugged Prairie", "Land — Mountain Plains", "zz-3604-prairie")
	pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")
	pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")

	id := pushTypedCardToHandWithCost(p, "Council's Judgment", "Sorcery", "{1}{W}{W}")
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast refused: %v", err)
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool = %v, want nothing floating", p.ManaPool)
	}
}
