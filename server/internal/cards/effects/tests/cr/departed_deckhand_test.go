package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A spell that targets the Deckhand kills it. The trigger fires at
// announce (CR 601.2c), so it is on the stack above the spell and
// resolves first — which is why the spell then finds no legal target.
func TestDepartedDeckhandSacrificesItselfToASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deckhand := effectstest.PushCatalogPermanent(g, me.ID, "Departed Deckhand",
		"Creature — Spirit Pirate", effectstest.DepartedDeckhandOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Unsummon", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: deckhand}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if effectstest.TriggerOnStack(g, deckhand) == nil && len(g.PendingTriggers) == 0 {
		t.Fatalf("becoming a spell's target triggers at announce: %+v", g.StackMeta)
	}

	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(deckhand) {
		t.Error("the Deckhand sacrifices itself")
	}
	if !me.Graveyard.Contains(deckhand) {
		t.Error("and it goes to its owner's graveyard")
	}
}

// "A SPELL", not "a spell or ability". An activated ability that
// targets it leaves it alone — the narrower wording is the whole
// difference between this and every Phantasmal Image variant.
func TestDepartedDeckhandSurvivesAnAbilityThatTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	deckhand := effectstest.PushCatalogPermanent(g, me.ID, "Departed Deckhand",
		"Creature — Spirit Pirate", effectstest.DepartedDeckhandOracle, false)
	magistrate := effectstest.PushCatalogPermanent(g, me.ID, "Immaculate Magistrate",
		"Creature — Elf Shaman", effectstest.ImmaculateMagistrateProbe, false)

	if err := g.ActivateCatalogAbility(me.ID, magistrate, 0,
		game.ActivateAbilityParams{Targets: effectstest.CardRefs(deckhand)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if effectstest.TriggerOnStack(g, deckhand) != nil {
		t.Error("an ability is not a spell — no sacrifice trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(deckhand) {
		t.Error("the Deckhand survives an ability that targets it")
	}
}
