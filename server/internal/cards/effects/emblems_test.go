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

// TestWrennsUltimateMakesAnEmblemThatGrantsRetrace is #2528. The −7
// resolves, an emblem appears whose whole text is a retrace grant, and
// from then on the owner's instants and sorceries in the graveyard —
// and only theirs — can be cast again for their printed cost plus a
// discarded land card, on any turn. Driven through the real loyalty path
// (ADR 0064 Decision 10 closed: the caveat is gone with the omission).
func TestWrennsUltimateMakesAnEmblemThatGrantsRetrace(t *testing.T) {
	spec, ok := Lookup(wrennAndSixOracle)
	if !ok {
		t.Fatal("Wrenn and Six is not registered")
	}
	if len(spec.Activated) != 3 || spec.Emblem == nil {
		t.Fatalf("Wrenn declares %d abilities and emblem %v, want 3 (+1, −1, −7) and the emblem", len(spec.Activated), spec.Emblem)
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Wrenn is %v with caveats %v, want Full", spec.Completeness, spec.Caveats)
	}

	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWalkerForTest(g, owner.ID, "Wrenn and Six", wrennAndSixOracle, 7)
	advanceToMainOf(t, g, seat)
	bolt := graveyardCardOf(g, owner, "Spent Bolt", "Instant", "{R}", "")
	theirs := graveyardCardOf(g, opp, "Their Bolt", "Instant", "{R}", "")
	creature := graveyardCardOf(g, owner, "Spent Bear", "Creature — Bear", "{1}{G}", "")
	land := uuid.New()
	owner.Hand.PushTop(game.Card{InstanceID: land, Name: "Hand Mountain", TypeLine: "Basic Land — Mountain",
		Owner: owner.ID, Controller: owner.ID, KnownBy: map[uuid.UUID]bool{owner.ID: true}})

	// Before the ultimate: nothing is castable from the graveyard.
	payMana(t, g, owner, "{R}")
	if err := retraceCast(g, owner, bolt, land); err == nil {
		t.Fatal("retrace was castable before the emblem existed")
	}

	if err := g.ActivateCatalogAbility(owner.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	passPriorityAroundTable(t, g)
	emblems := emblemsOf(g, owner.ID)
	if len(emblems) != 1 || emblems[0].Text != "Instant and sorcery cards in your graveyard have retrace." {
		t.Fatalf("emblems after the −7 = %+v", emblems)
	}
	advanceToMainOf(t, g, seat)

	// A permanent card is not covered; an opponent's graveyard is not mine.
	if err := retraceCast(g, owner, creature, land); err == nil {
		t.Error("the emblem granted retrace to a creature card")
	}
	if err := g.CastSpell(owner.ID, theirs, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: []uuid.UUID{land}, Strict: true,
	}); err == nil {
		t.Error("the emblem opened an opponent's graveyard")
	}

	// The reference cast: a spent Bolt is cast again for {R} and a land.
	payMana(t, g, owner, "{R}")
	if err := g.CastSpell(owner.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: []uuid.UUID{land}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("retrace under the emblem: %v", err)
	}
	if !owner.Graveyard.Contains(land) {
		t.Error("the retrace land was not discarded")
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
			len(spec.Emblem.UntapStep) == 0 && len(spec.Emblem.DrawStep) == 0 &&
			len(spec.Emblem.ActivationTimings) == 0 && len(spec.Emblem.CastRestrictions) == 0 &&
			len(spec.Emblem.LandPlayRestrictions) == 0 && len(spec.Emblem.GameEndGates) == 0 &&
			len(spec.Emblem.UntapCaps) == 0 && len(spec.Emblem.CastPermissions) == 0 {
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
		// to be found through those two hooks too. #1275 added the
		// activation-timing slot (Teferi, Temporal Archmage), whose
		// emblem is found through CatalogActivationTimings alone.
		// ADR 0109 §5 (#1899) added the four rule-gate slots (Narset
		// Transcendent's cast restriction, Dovin Baan's untap cap).
		key := game.EmblemKey(spec.OracleID)
		if len(game.CatalogStaticAbilities(key)) == 0 &&
			len(game.CatalogTriggers(key)) == 0 &&
			len(game.CatalogUntapStepPermissions(key)) == 0 &&
			len(game.CatalogDrawStepPermissions(key)) == 0 &&
			len(game.CatalogActivationTimings(key)) == 0 &&
			len(game.CatalogCastRestrictions(key)) == 0 &&
			len(game.CatalogLandPlayRestrictions(key)) == 0 &&
			len(game.CatalogGameEndGates(key)) == 0 &&
			len(game.CatalogUntapCaps(key)) == 0 &&
			len(game.CatalogCastPermissions(key)) == 0 {
			t.Errorf("%s's emblem is not registered under %q", spec.Name, key)
		}
	}
}
