package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// departed_ability_source_target_test.go is the card-level proof for
// #1429: when an ability resolves after its source has left the
// battlefield, the CR 608.2b target re-check tests protection against
// the source AS IT LAST EXISTED there (CR 608.2h), not against the
// source's card in the graveyard.
//
// The line is #1417's (departed_source_colour_test.go): a black-red
// Murderous Redcap's enter trigger targets an opponent's creature, and
// in response Cerulean Wisps turns the Redcap blue and it dies. Its
// graveyard card is black-red again. Then the opponent's Mother of
// Runes gives the targeted creature protection from a colour, and the
// trigger resolves.

// motherOfRunesProtectsInResponse has `opp` (not the active player)
// activate Mother of Runes on `creature`, answers its colour prompt
// with `color`, and stops with the Redcap trigger still on the stack.
func motherOfRunesProtectsInResponse(t *testing.T, g *game.Game, opp *game.Player, creature uuid.UUID, color string) {
	t.Helper()
	mother := pushCatalogPermanent(g, opp.ID, "Mother of Runes", "Creature — Human Cleric", motherOfRunesOracle, false)
	// The active player passes; the opponent gets priority.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.ActivateCatalogAbility(opp.ID, mother, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("activate Mother of Runes: %v", err)
	}
	for i := 0; i < 8 && pendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority over Mother of Runes: %v", err)
		}
	}
	answerColor(t, g, opp.ID, color)
	if c, _ := battlefieldCard(g, creature); !game.HasKeyword(&c, game.ProtectionFromColor(color)) {
		t.Fatalf("setup: Mother of Runes did not give the creature protection from %s", color)
	}
}

// redcapTriggerFizzled reports whether the Redcap's trigger was
// countered by game rules (CR 608.2b) rather than resolved.
func redcapTriggerFizzled(g *game.Game, redcap uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventFizzle && ev.Source == redcap {
			return true
		}
	}
	return false
}

// Protection from RED: the Redcap was blue as it last existed, so the
// target is still legal and the 2 damage lands. Reading the black-red
// graveyard card made the target illegal and the trigger fizzled.
func TestRedcapTriggerTurnedBlueThenKilledStillHitsACreatureGivenProtectionFromRed(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hill Giant", TypeLine: "Creature — Giant",
		Power: 3, Toughness: 5, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID,
	})
	redcap := redcapTurnedBlueAndKilled(t, g, game.TargetRef{Kind: game.TargetCard, ID: victim})
	motherOfRunesProtectsInResponse(t, g, opp, victim, "R")
	passPriorityAroundTable(t, g)

	if redcapTriggerFizzled(g, redcap) {
		t.Error("the Redcap trigger fizzled: its source was BLUE as it last existed, so a " +
			"pro-red target is still legal (CR 608.2b / 608.2h, #1429)")
	}
	c, ok := battlefieldCard(g, victim)
	if !ok {
		t.Fatal("the creature should still be there")
	}
	if c.DamageMarked != 2 {
		t.Errorf("creature has %d damage, want 2 from the departed blue Redcap", c.DamageMarked)
	}
}

// Protection from BLUE: the Redcap was blue as it last existed, so the
// target is illegal and the trigger is countered by game rules. Reading
// the black-red graveyard card let it resolve (its damage was then
// prevented by #1417, so only the fizzle tells the two apart).
func TestRedcapTriggerTurnedBlueThenKilledFizzlesAgainstProtectionFromBlue(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hill Giant", TypeLine: "Creature — Giant",
		Power: 3, Toughness: 5, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID,
	})
	redcap := redcapTurnedBlueAndKilled(t, g, game.TargetRef{Kind: game.TargetCard, ID: victim})
	motherOfRunesProtectsInResponse(t, g, opp, victim, "U")
	passPriorityAroundTable(t, g)

	if !redcapTriggerFizzled(g, redcap) {
		t.Error("the Redcap trigger resolved: its source was BLUE as it last existed, so a " +
			"pro-blue target is illegal and the trigger is countered by game rules " +
			"(CR 608.2b / 608.2h, #1429)")
	}
	if c, _ := battlefieldCard(g, victim); c.DamageMarked != 0 {
		t.Errorf("creature has %d damage, want 0", c.DamageMarked)
	}
}
