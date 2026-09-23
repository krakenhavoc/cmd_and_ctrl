package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emblems_test.go is the card-side of CR 114 (#623, ADR 0064): the
// two ultimates that now make real emblems, driven through the real
// loyalty path with the real catalog wired.

const teferiHeroOracle = "f2f165b6-ef0a-42ad-9352-ba68be8248b0"

// emblemsOf reads a seat's emblems off the projection the wire uses.
func emblemsOf(g *game.Game, playerID uuid.UUID) []game.EmblemView {
	var out []game.EmblemView
	g.ReadSnapshot(func() { out = g.EmblemsForPlayer(playerID) })
	return out
}

func effectivePTOf(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tough int
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				eff := c.Effective()
				p, tough, found = eff.Power, eff.Toughness, true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return p, tough
}

func hasKeywordOf(g *game.Game, id uuid.UUID, kw string) bool {
	has := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			for _, k := range c.Effective().Abilities {
				if k == kw {
					has = true
				}
			}
		}
	})
	return has
}

// TestElspethUltimateMakesAnEmblemThatPumpsYourCreatures is the
// worked example from ADR 0064: the −7 resolves, an emblem appears in
// the command zone, and from then on the controller's creatures are
// +2/+2 with flying and nobody else's are.
func TestElspethUltimateMakesAnEmblemThatPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWalkerForTest(g, owner.ID, "Elspeth, Sun's Champion", elspethSunsChampionOracle, 7)
	mine := pushCreatureToBattlefieldForTest(g, owner.ID, "Mine")
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Theirs")
	advanceToMainOf(t, g, seat)

	if p, tough := effectivePTOf(t, g, mine); p != 2 || tough != 2 {
		t.Fatalf("baseline P/T = %d/%d, want 2/2", p, tough)
	}

	// Index 2 is the −7; the +1 and the −3 are 0 and 1.
	if err := g.ActivateCatalogAbility(owner.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	passPriorityAroundTable(t, g)

	emblems := emblemsOf(g, owner.ID)
	if len(emblems) != 1 {
		t.Fatalf("owner has %d emblems after the −7, want 1", len(emblems))
	}
	if emblems[0].Label != "Elspeth, Sun's Champion emblem" {
		t.Errorf("emblem label = %q", emblems[0].Label)
	}
	if emblems[0].Text != "Creatures you control get +2/+2 and have flying." {
		t.Errorf("emblem text = %q", emblems[0].Text)
	}
	if p, tough := effectivePTOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my creature is %d/%d, want 4/4", p, tough)
	}
	if !hasKeywordOf(g, mine, "flying") {
		t.Error("my creature does not have flying")
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's creature is %d/%d — the emblem is not symmetric", p, tough)
	}
	if hasKeywordOf(g, theirs, "flying") {
		t.Error("the opponent's creature got flying from my emblem")
	}
	if len(emblemsOf(g, opp.ID)) != 0 {
		t.Error("the opponent got an emblem too")
	}
}

// TestElspethsEmblemOutlivesElspeth — the emblem is an object, not a
// continuous effect the planeswalker sources, so killing Elspeth
// changes nothing (CR 114).
func TestElspethsEmblemOutlivesElspeth(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	pw := pushWalkerForTest(g, owner.ID, "Elspeth, Sun's Champion", elspethSunsChampionOracle, 7)
	mine := pushCreatureToBattlefieldForTest(g, owner.ID, "Mine")
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	passPriorityAroundTable(t, g)

	// The −7 takes Elspeth to 0 loyalty, so the CR 704.5i state-based
	// action has already put her in the graveyard by now.
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == pw {
			t.Fatal("Elspeth is still on the battlefield at 0 loyalty")
		}
	}
	if len(emblemsOf(g, owner.ID)) != 1 {
		t.Fatal("the emblem left with Elspeth")
	}
	if p, tough := effectivePTOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my creature is %d/%d after Elspeth died, want 4/4", p, tough)
	}
}

// TestTeferiHeroUltimateMakesATriggeredEmblem is the triggered half:
// the −8's emblem watches the controller's draws and puts a targeted
// exile on the stack for each one.
func TestTeferiHeroUltimateMakesATriggeredEmblem(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWalkerForTest(g, owner.ID, "Teferi, Hero of Dominaria", teferiHeroOracle, 8)
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Victim")
	advanceToMainOf(t, g, seat)

	// Index 2 is the −8; 0 is the +1 and 1 is the −3.
	if err := g.ActivateCatalogAbility(owner.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	passPriorityAroundTable(t, g)

	emblems := emblemsOf(g, owner.ID)
	if len(emblems) != 1 {
		t.Fatalf("owner has %d emblems after the −8, want 1", len(emblems))
	}
	if emblems[0].Text != "Whenever you draw a card, exile target permanent an opponent controls." {
		t.Errorf("emblem text = %q", emblems[0].Text)
	}

	if err := g.DrawCard(owner.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	// The trigger is targeted, so the harvester queues the CR 603.3d
	// target pick before the item reaches the stack.
	pickCard(t, g, owner.ID, victim)
	passPriorityAroundTable(t, g)

	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == victim {
			t.Fatal("the opponent's permanent is still on the battlefield — the emblem's trigger did not exile it")
		}
	}
	if !g.Exile.Contains(victim) {
		t.Error("the opponent's permanent is not in exile")
	}
}

// TestTeferiHerosEmblemIgnoresAnOpponentsDraw — "whenever YOU draw",
// read off the emblem's controller, which is its owner (CR 114.5).
func TestTeferiHerosEmblemIgnoresAnOpponentsDraw(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWalkerForTest(g, owner.ID, "Teferi, Hero of Dominaria", teferiHeroOracle, 8)
	mine := pushCreatureToBattlefieldForTest(g, owner.ID, "Mine")
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−8: %v", err)
	}
	passPriorityAroundTable(t, g)

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)

	found := false
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == mine {
			found = true
		}
	}
	if !found {
		t.Error("the opponent's draw fired my emblem and exiled my own permanent")
	}
}

// TestWrennsUltimateStaysOmittedAndNamesItsBlocker is ADR 0064
// Decision 10. Emblems exist; retrace does not, and Wrenn's emblem
// grants nothing else, so the −7 stays unregistered and the caveat
// names #652 rather than emblems.
func TestWrennsUltimateStaysOmittedAndNamesItsBlocker(t *testing.T) {
	spec, ok := Lookup(wrennAndSixOracle)
	if !ok {
		t.Fatal("Wrenn and Six is not registered")
	}
	if spec.Emblem != nil {
		t.Error("Wrenn declares an Emblem — its only text is a retrace grant nothing can honour yet")
	}
	if len(spec.Activated) != 2 {
		t.Errorf("Wrenn declares %d abilities, want 2 (+1 and −1)", len(spec.Activated))
	}
	if spec.Completeness != CompletenessCaveats {
		t.Errorf("Completeness = %v, want CompletenessCaveats", spec.Completeness)
	}
	if len(spec.Caveats) != 1 {
		t.Fatalf("Caveats = %v, want exactly one", spec.Caveats)
	}
	if !strings.Contains(spec.Caveats[0], "#652") {
		t.Errorf("the caveat does not name the blocking seam: %q", spec.Caveats[0])
	}
	if strings.Contains(spec.Caveats[0], "emblem") && !strings.Contains(spec.Caveats[0], "retrace") {
		t.Errorf("the caveat still blames emblems: %q", spec.Caveats[0])
	}
}

// TestEveryEmblemSpecIsWellFormed is the catalog-wide guard on the
// new slot. effects.Register panics on a malformed one at boot, so
// this is the readable statement of the same rule plus the label
// convention the client renders.
func TestEveryEmblemSpecIsWellFormed(t *testing.T) {
	for _, spec := range All() {
		if spec.Emblem == nil {
			continue
		}
		if len(spec.Emblem.Static) == 0 && len(spec.Emblem.Triggered) == 0 &&
			len(spec.Emblem.UntapStep) == 0 && len(spec.Emblem.DrawStep) == 0 {
			t.Errorf("%s declares an emblem with no abilities", spec.Name)
		}
		if spec.Emblem.Label == "" || spec.Emblem.Text == "" {
			t.Errorf("%s declares an emblem with no label or no text", spec.Name)
		}
		if !strings.Contains(spec.Emblem.Label, "emblem") {
			t.Errorf("%s's emblem label %q does not say it is an emblem", spec.Name, spec.Emblem.Label)
		}
		// The emblem's def has to be reachable under the synthetic
		// key, or nothing the card creates will have any abilities.
		// #1315 widened the def with two turn-based-action slots
		// alongside Static/Triggered, so a card whose only abilities
		// are UntapStep/DrawStep (Teferi, Who Slows the Sunset) has
		// to be found through those two hooks too.
		key := game.EmblemKey(spec.OracleID)
		if len(game.CatalogStaticAbilities(key)) == 0 &&
			len(game.CatalogTriggers(key)) == 0 &&
			len(game.CatalogUntapStepPermissions(key)) == 0 &&
			len(game.CatalogDrawStepPermissions(key)) == 0 {
			t.Errorf("%s's emblem is not registered under %q", spec.Name, key)
		}
	}
}
