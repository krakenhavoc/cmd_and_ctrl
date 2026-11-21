package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_retarget_departed_source_test.go is the card-level proof for
// #1449: when Strionic Resonator copies a triggered ability whose
// source has left the battlefield, the copy's "you may choose new
// targets" prompt (CR 707.10c) judges the new targets against the
// source AS IT LAST EXISTED there (CR 608.2h), not against its card in
// the graveyard.
//
// The line is #1429's (departed_ability_source_target_test.go): a
// black-red Murderous Redcap's enter trigger targets an opponent's
// creature; in response Cerulean Wisps turns the Redcap blue and it
// dies, so its graveyard card is black-red again. Then the Redcap's
// controller copies the trigger with Strionic Resonator.

// resonatorCopiesTheRedcapTrigger activates Strionic Resonator on the
// Redcap's enter trigger and passes priority until the copy's
// re-target prompt opens. Returns the prompt.
func ResonatorCopiesTheRedcapTrigger(t *testing.T, g *game.Game, me *game.Player, redcap uuid.UUID) *game.PendingChoice {
	t.Helper()
	res := PushCatalogPermanent(g, me.ID, "Strionic Resonator", "Artifact", StrionicResonatorOracle, false)
	trig := TriggerOnStack(g, redcap)
	if trig == nil {
		t.Fatal("setup: the Redcap's enter trigger should be on the stack")
	}
	FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, res, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: trig.ID}},
	}); err != nil {
		t.Fatalf("activate Strionic Resonator: %v", err)
	}
	for i := 0; i < 8 && LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority over the Resonator: %v", err)
		}
	}
	p := LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the Resonator's copy opened no re-target prompt")
	}
	return p
}

// protectedCreatures seats an opponent's plain creature and one with
// protection from each of red and blue.
func ProtectedCreatures(g *game.Game, opp *game.Player) (plain, proRed, proBlue uuid.UUID) {
	seat := func(name string, keywords ...string) uuid.UUID {
		return PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Giant",
			Power: 3, Toughness: 5, Colors: []string{"W"}, Keywords: keywords,
			Owner: opp.ID, Controller: opp.ID,
		})
	}
	return seat("Plain Giant"), seat("Pro-Red Giant", "protection from red"), seat("Pro-Blue Giant", "protection from blue")
}
