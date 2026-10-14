package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soul_exchange_test.go — #3402: "exile a creature you control" as a
// SPELL's additional cost, and the Thrull rider that reads what it took.

const soulExchangeOracle = "366ea3b4-5e7e-4693-9f4b-1e549972609e"

func castSoulExchange(t *testing.T, g *game.Game, target uuid.UUID, exile ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Soul Exchange", "Sorcery", "{B}{B}", soulExchangeOracle, game.CastSpellParams{
		Targets:           []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		ExilePermanentIDs: exile,
	})
}

func pushTypedCreature(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// Exiling a Thrull pays the cost, returns the creature, and puts a +2/+2
// counter on it. The Thrull is exiled, not sacrificed: nothing died.
func TestSoulExchangeWithAThrullPutsACounterOnTheReturnedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thrull := pushTypedCreature(g, me.ID, "Blood Pet", "Creature — Thrull")
	dead := pushGraveyardCardForTest(me, "Dead Bear")

	spell, err := castSoulExchange(t, g, dead, thrull)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Exile.Contains(thrull) || me.Graveyard.Contains(thrull) {
		t.Fatal("the Thrull was not exiled as the cost")
	}
	paid := g.StackMeta[spell].Paid
	if len(paid.Exiled) != 1 || paid.Exiled[0] != thrull || len(paid.ExiledObjects) != 1 {
		t.Fatalf("Paid = %+v, want the Thrull recorded as exiled", paid)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, dead)
	if !ok || c.Controller != me.ID {
		t.Fatal("the target did not return to the battlefield under your control")
	}
	if got := c.Counters["+2/+2"]; got != 1 {
		t.Errorf("+2/+2 counters = %d, want 1 for an exiled Thrull", got)
	}
}

// Any other creature pays the cost and the reanimation still happens,
// with no counter.
func TestSoulExchangeWithANonThrullPutsNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushTypedCreature(g, me.ID, "Bear", "Creature — Bear")
	dead := pushGraveyardCardForTest(me, "Dead Bear")

	if _, err := castSoulExchange(t, g, dead, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, dead)
	if !ok {
		t.Fatal("the target did not return")
	}
	if got := c.Counters["+2/+2"]; got != 0 {
		t.Errorf("+2/+2 counters = %d, want none for a Bear", got)
	}
}

// CR 608.2h: the Thrull is read as it last existed, so a creature that was
// a Thrull only through an effect still counts, and one that stopped being
// one does not.
func TestSoulExchangeReadsTheExiledCreatureAsItLastExisted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	changeling := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mirror Shape", TypeLine: "Creature — Shapeshifter",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Keywords: []string{game.KeywordChangeling},
	})
	dead := pushGraveyardCardForTest(me, "Dead Bear")
	if _, err := castSoulExchange(t, g, dead, changeling); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, dead)
	if !ok || c.Counters["+2/+2"] != 1 {
		t.Errorf("a changeling is a Thrull: counters = %v", c.Counters)
	}
}

// The cost must be paid: naming nothing, an opponent's creature, a
// non-creature, or two permanents is refused, and nothing moves.
func TestSoulExchangeRefusesAnIllegalExileCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushTypedCreature(g, me.ID, "Bear", "Creature — Bear")
	other := pushTypedCreature(g, me.ID, "Bear Two", "Creature — Bear")
	theirs := pushTypedCreature(g, opp.ID, "Ogre", "Creature — Ogre")
	artifact := pushTypedCreature(g, me.ID, "Rock", "Artifact")
	dead := pushGraveyardCardForTest(me, "Dead Bear")

	for _, tc := range []struct {
		name string
		ids  []uuid.UUID
	}{
		{"nothing named", nil},
		{"an opponent's creature", []uuid.UUID{theirs}},
		{"a noncreature", []uuid.UUID{artifact}},
		{"two creatures", []uuid.UUID{mine, other}},
		{"the same creature twice", []uuid.UUID{mine, mine}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := castSoulExchange(t, g, dead, tc.ids...)
			if err == nil {
				t.Fatal("cast accepted an illegal exile cost")
			}
			if !errors.Is(err, game.ErrInvalidParam) && !errors.Is(err, game.ErrCardCallerMismatch) && !errors.Is(err, game.ErrIllegalTarget) {
				t.Errorf("err = %v, want a refusal of the payment", err)
			}
			if g.Exile.Contains(mine) || g.Exile.Contains(theirs) {
				t.Error("a refused cast exiled something")
			}
		})
	}
}

// exile_permanent_ids on a spell with no such cost is refused rather than
// ignored, as an unexpected sacrifice_ids is.
func TestSoulExchangeFamilyRefusesAnExilePickWithoutTheCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushTypedCreature(g, me.ID, "Bear", "Creature — Bear")
	_, err := castWithTapParams(t, g, "Altar's Reap", "Instant", "{1}{B}", "6a125750-2b8c-4f9d-8173-ac8d14c91ddb",
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{bear}, ExilePermanentIDs: []uuid.UUID{bear}})
	if err == nil {
		t.Fatal("cast accepted exile_permanent_ids for a spell with no exile cost")
	}
}

// A leaves-the-battlefield watcher sees the exile and a dies watcher does
// not: the cost exiles.
func TestSoulExchangeExilesRatherThanSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mender := pushCircuitMender(g, me.ID)
	zulaport := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	_ = zulaport
	dead := pushGraveyardCardForTest(me, "Dead Bear")
	life := lifeOf(g, g.Seats[1].ID)
	if _, err := castSoulExchange(t, g, dead, mender); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if got := lifeOf(g, g.Seats[1].ID); got != life {
		t.Errorf("Zulaport Cutthroat drained for an exile (opponent life %d → %d)", life, got)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("Circuit Mender's leaves-the-battlefield trigger drew %d cards, want 1", got)
	}
}
