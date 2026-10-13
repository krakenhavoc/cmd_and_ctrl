package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward_discard_test.go — "Ward—Discard a card" (#3172): Graveyard
// Trespasser, Graveyard Glutton, Gideon the Oathless.
//
// ward_test.go pins what every ward shares and ward_life_sacrifice_test.go
// the other non-mana costs. These pin what a discard cost changes: the
// payer picks the card from their own hand, an empty hand cannot pay and
// is countered with no prompt, and paying actually puts the card in the
// graveyard before the spell goes on to resolve.

var wardDiscardCards = []struct {
	name, typeLine, oracle string
}{
	{"Graveyard Trespasser", "Creature — Human Werewolf", "0bbd6cad-9b6f-45a6-9f2e-d7b4853586ae"},
	{"Graveyard Glutton", "Creature — Werewolf", "0bbd6cad-9b6f-45a6-9f2e-d7b4853586ae#1"},
	{"Gideon the Oathless", "Legendary Creature — Human Mercenary", "7afa530f-ea51-4bee-af3b-59b83ea5cb25"},
}

// boltWardedCreature casts a Lightning Bolt at the warded creature and
// returns the spell's ID. The Glutton is a 4/4 and survives it, so a
// test reads whether the bolt resolved off the damage on the creature.
func boltWardedCreature(t *testing.T, g *game.Game, caster *game.Player, victim uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("a warded permanent is a LEGAL target — announce must succeed: %v", err)
	}
	return id
}

// boltResolved is true when the Lightning Bolt dealt its damage: the
// creature either died or carries 3 marked damage.
func boltResolved(g *game.Game, victim uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(victim)
	if !ok || !g.Battlefield.Contains(victim) {
		return true
	}
	return c.DamageMarked == 3
}

func fillHand(p *game.Player, names ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range names {
		c := game.NewCard(n, p.ID)
		c.TypeLine = "Sorcery"
		p.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// Declining counters the spell and discards nothing.
func TestDiscardWardDeclinedCountersTheSpell(t *testing.T) {
	for _, w := range wardDiscardCards {
		t.Run(w.name, func(t *testing.T) {
			g, _, opp, warded := wardTable(t, w.name, w.typeLine, w.oracle)
			hand := fillHand(opp, "Filler A", "Filler B")
			bolt := boltWardedCreature(t, g, opp, warded)

			prompt := passUntilConfirmFor(t, g, opp.ID)
			if prompt == nil {
				t.Fatal("no ward prompt for the spell's controller")
			}
			if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			passPriorityAroundTable(t, g)

			if !opp.Graveyard.Contains(bolt) {
				t.Error("the countered spell should be in its owner's graveyard")
			}
			for _, id := range hand {
				if !opp.Hand.Contains(id) {
					t.Error("declining a discard ward must not discard anything")
				}
			}
			if boltResolved(g, warded) {
				t.Error("a countered Lightning Bolt must not damage the creature")
			}
		})
	}
}

// Paying is a two-link chain — the confirm, then the caster's own
// pick from their own hand — and the bolt then resolves.
func TestDiscardWardPaidDiscardsTheChosenCardAndTheSpellResolves(t *testing.T) {
	for _, w := range wardDiscardCards {
		t.Run(w.name, func(t *testing.T) {
			g, me, opp, warded := wardTable(t, w.name, w.typeLine, w.oracle)
			hand := fillHand(opp, "Filler A", "Filler B")
			boltWardedCreature(t, g, opp, warded)

			prompt := passUntilConfirmFor(t, g, opp.ID)
			if prompt == nil {
				t.Fatal("no ward prompt for the spell's controller")
			}
			if confirmChoiceFor(g, me.ID) != nil {
				t.Error("the ward permanent's controller must not be asked to pay")
			}
			if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			if discardChoiceFor(g, me.ID) != nil {
				t.Error("the ward permanent's controller must not be asked to discard")
			}
			if n := discardOwed(g, opp.ID); n != 1 {
				t.Fatalf("the payer owes %d cards, want 1", n)
			}
			// The payer chooses: pitch the SECOND card, not the top one.
			answerDiscard(t, g, opp.ID, hand[1])
			passPriorityAroundTable(t, g)

			if !opp.Graveyard.Contains(hand[1]) {
				t.Error("the chosen card should be in the graveyard")
			}
			if !opp.Hand.Contains(hand[0]) {
				t.Error("only the chosen card goes")
			}
			if !boltResolved(g, warded) {
				t.Error("the ward was paid — the Lightning Bolt should have resolved")
			}
		})
	}
}

// A caster with an empty hand cannot pay: the spell is countered and
// nobody is asked (CR 118.3).
func TestDiscardWardCountersWithoutAPromptWhenTheHandIsEmpty(t *testing.T) {
	g, _, opp, warded := wardTable(t, "Gideon the Oathless", "Legendary Creature — Human Mercenary", wardDiscardCards[2].oracle)
	for opp.Hand.Size() > 0 {
		if _, err := opp.Hand.PopTop(); err != nil {
			t.Fatal(err)
		}
	}
	bolt := boltWardedCreature(t, g, opp, warded)

	if prompt := passUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatalf("a caster with no cards was offered a discard payment: %q", prompt.Reason)
	}
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) || boltResolved(g, warded) {
		t.Error("a ward the caster cannot pay must counter the spell")
	}
}

// Your own spell on your own warded creature does not trigger it.
func TestDiscardWardIgnoresItsControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	warded := pushCatalogPermanent(g, me.ID, "Gideon the Oathless", "Legendary Creature — Human Mercenary", wardDiscardCards[2].oracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	fillHand(me, "Filler A")
	boltWardedCreature(t, g, me, warded)
	if prompt := passUntilConfirmFor(t, g, me.ID); prompt != nil {
		t.Fatalf("a ward asked its own controller to pay: %q", prompt.Reason)
	}
}

// The cards are Full: the ward was the only declared gap.
func TestDiscardWardCardsAreComplete(t *testing.T) {
	for _, w := range wardDiscardCards {
		spec, ok := Lookup(w.oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want CompletenessFull with no caveats, got %v %v", w.name, spec.Completeness, spec.Caveats)
		}
	}
}
