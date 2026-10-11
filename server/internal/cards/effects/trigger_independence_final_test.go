package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_final_test.go — ADR 0018's #3013 amendment,
// sub-PR 4, batch 6: the power variant of GainLifeEqualToSubject
// (Conclave Mentor) and the tapped-token step (CreateTokenAdvanced with
// only EntersTapped), each with a batch that skips and one that asks.

const (
	conclaveMentorFinalOracle = "a2fe5937-212c-4e71-8d6e-f408b38100aa"
	donAndresFinalOracle      = "060ef981-db05-436b-b1c1-f55071375344"
	kessigFinalOracle         = "eadd9559-6dcb-4b96-8c95-58abddd0e120"
)

func TestTheFinalBatchDeclaresFootprints(t *testing.T) {
	rows := game.CatalogTriggers(conclaveMentorFinalOracle)
	if len(rows) != 1 || len(rows[0].Footprint) != 1 ||
		rows[0].Footprint[0].Kind != game.FootprintGainLife || !rows[0].Footprint[0].AmountFromSubject {
		t.Errorf("Conclave Mentor: rows %+v, want a gain from the subject", rows)
	}
	rows = game.CatalogTriggers(donAndresFinalOracle)
	if len(rows) != 1 || len(rows[0].Footprint) != 1 || rows[0].Footprint[0].Kind != game.FootprintCreateToken {
		t.Errorf("Don Andres: rows %+v, want a token step", rows)
	}
}

// Conclave Mentor dying gains life equal to its last-known power, and
// beside Zulaport Cutthroat's drain (its own death trigger too) the
// order cannot show: no prompt.
func TestConclaveMentorAndADrainOnItsDeathDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mentor := b21Push(g, me.ID, "Conclave Mentor", "Creature — Centaur Cleric", conclaveMentorFinalOracle, 3, 3)
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportOracle, false)
	toMain(t, g)
	life := me.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mentor) })
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("Conclave Mentor and a drain asked for an order: %+v", ch)
	}
	if me.Life != life+3+1 {
		t.Errorf("life %d → %d, want +4 (its power and the drain)", life, me.Life)
	}
}

// A watcher of life gain (Ajani's Pridemate) sees the two gains in the
// order they resolve.
func TestConclaveMentorAndADrainStillAskBesideALifeWatcher(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mentor := b21Push(g, me.ID, "Conclave Mentor", "Creature — Centaur Cleric", conclaveMentorFinalOracle, 3, 3)
	pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportOracle, false)
	pushCatalogPermanent(g, me.ID, "Ajani's Pridemate", "Creature — Cat Soldier", b24AjanisPridemateOracle, false)
	toMain(t, g)

	before := promptsByReason(t, game.TriggerOrderReasonReactor)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mentor) })
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("a life watcher on the board must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
		t.Errorf("counted %v under reactor, want 1", got)
	}
}

// Don Andres (two tapped Treasures off a noncreature spell the caster
// doesn't own) and Kessig Flamebreather (1 damage off any noncreature
// spell) both trigger on one cast of an opponent's card.
func castStolenSpellBesideDonAndres(t *testing.T, g *game.Game) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Don Andres, the Renegade", "Legendary Creature — Human Pirate", donAndresFinalOracle, false)
	pushCatalogPermanent(g, me.ID, "Kessig Flamebreather", "Creature — Human Shaman", kessigFinalOracle, false)
	toMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Stolen Charm", TypeLine: "Instant",
		Owner: opp.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}

// treasureCount counts the Treasure tokens on the battlefield.
func treasureCount(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Treasure") {
			n++
		}
	}
	return n
}

func TestATappedTokenMakerAndADrainDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	castStolenSpellBesideDonAndres(t, g)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("a tapped token maker and a drain asked for an order: %+v", ch)
	}
	if n := treasureCount(g); n != 2 {
		t.Errorf("%d Treasures, want 2", n)
	}
}

func TestATappedTokenMakerAndADrainStillAskBesideAnEntryWatcher(t *testing.T) {
	g := newCatalogGame(t)
	pushCatalogPermanent(g, g.Seats[0].ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	before := promptsByReason(t, game.TriggerOrderReasonReactor)
	castStolenSpellBesideDonAndres(t, g)
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("an entry watcher on the board must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
		t.Errorf("counted %v under reactor, want 1", got)
	}
}
