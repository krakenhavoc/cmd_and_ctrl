package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	departedDeckhandOracle    = "a620a765-97ba-4687-acfd-4dec7da75d9f"
	immaculateMagistrateProbe = "e6e38bd4-e6dc-400b-8e08-956726842dc4"
)

// A spell that targets the Deckhand kills it. The trigger fires at
// announce (CR 601.2c), so it is on the stack above the spell and
// resolves first — which is why the spell then finds no legal target.
func TestDepartedDeckhandSacrificesItselfToASpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deckhand := pushCatalogPermanent(g, me.ID, "Departed Deckhand",
		"Creature — Spirit Pirate", departedDeckhandOracle, false)

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
	if triggerOnStack(g, deckhand) == nil && len(g.PendingTriggers) == 0 {
		t.Fatalf("becoming a spell's target triggers at announce: %+v", g.StackMeta)
	}

	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	deckhand := pushCatalogPermanent(g, me.ID, "Departed Deckhand",
		"Creature — Spirit Pirate", departedDeckhandOracle, false)
	magistrate := pushCatalogPermanent(g, me.ID, "Immaculate Magistrate",
		"Creature — Elf Shaman", immaculateMagistrateProbe, false)

	if err := g.ActivateCatalogAbility(me.ID, magistrate, 0,
		game.ActivateAbilityParams{Targets: cardRefs(deckhand)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if triggerOnStack(g, deckhand) != nil {
		t.Error("an ability is not a spell — no sacrifice trigger")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(deckhand) {
		t.Error("the Deckhand survives an ability that targets it")
	}
}
