package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_lady_loki_test.go — Lady Loki, Agent of Chaos (#4455), played through
// the engine: a real cast, the real trigger, the real exile walk and the
// real free-cast offer.

type mscLokiTable struct {
	g    *game.Game
	me   *game.Player
	loki uuid.UUID
}

func mscLokiNew(t *testing.T) *mscLokiTable {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	loki := apaPush(g, me.ID, me.ID, game.Card{
		Name: "Lady Loki, Agent of Chaos", TypeLine: "Legendary Creature — God Sorcerer Villain",
		OracleID: mscWlLokiOracle, Power: 5, Toughness: 5,
	})
	return &mscLokiTable{g: g, me: me, loki: loki}
}

// library seeds the top of the library, first card = top.
func (tb *mscLokiTable) library(cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	tb.g.WithWriteLock(func() {
		for i := len(cards) - 1; i >= 0; i-- {
			c := cards[i]
			c.InstanceID = uuid.New()
			c.Owner, c.Controller = tb.me.ID, tb.me.ID
			ids[i] = c.InstanceID
			tb.me.Library.PushTop(c)
		}
	})
	return ids
}

// cast puts the spell in hand and casts it, paying with floating mana.
func (tb *mscLokiTable) cast(t *testing.T, name, typeLine, cost, mana string, x int) uuid.UUID {
	t.Helper()
	floatForTest(tb.g, tb.me, mana)
	id, err := castWithTapParams(t, tb.g, name, typeLine, cost, "", game.CastSpellParams{Strict: true, XValue: x})
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func (tb *mscLokiTable) resolveTrigger(t *testing.T) {
	t.Helper()
	passPriorityAroundTable(t, tb.g)
}

func (tb *mscLokiTable) opponentLife() []int {
	var out []int
	for _, p := range tb.g.Seats {
		if p.ID != tb.me.ID {
			out = append(out, p.Life)
		}
	}
	return out
}

func mscLokiExpectOpponents(t *testing.T, tb *mscLokiTable, lost int) {
	t.Helper()
	for i, life := range tb.opponentLife() {
		if life != 40-lost {
			t.Errorf("opponent %d has %d life, want %d", i+1, life, 40-lost)
		}
	}
}

func TestMscLokiExilesTheSpellAndDamagesByTheManaValueDifference(t *testing.T) {
	tb := mscLokiNew(t)
	g, me := tb.g, tb.me
	ids := tb.library(
		game.Card{Name: "Passed Over Land", TypeLine: "Land"},
		game.Card{Name: "Passed Over Land 2", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Hit", TypeLine: "Creature — Test", ManaCost: "{3}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Deeper", TypeLine: "Creature — Test", ManaCost: "{G}", Power: 1, Toughness: 1},
	)
	spell := tb.cast(t, "Test Sorcery", "Sorcery", "{1}{R}", "CR", 0)
	if g.CastTallyFor(me.ID).InstantSorceryVillain != 1 {
		t.Fatalf("the cast was not tallied")
	}
	tb.resolveTrigger(t)

	if !inExile(g, spell) {
		t.Error("the spell was not exiled")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == spell {
			t.Error("the spell went to the graveyard: it was countered, not exiled")
		}
	}
	for _, id := range ids[:3] {
		if !inExile(g, id) {
			t.Errorf("card %s of the walk is not in exile", id)
		}
	}
	if inExile(g, ids[3]) || !me.Library.Contains(ids[3]) {
		t.Error("the walk went past the first nonland card")
	}
	// {1}{R} is mana value 2, the hit is 4: 2 damage to each opponent.
	mscLokiExpectOpponents(t, tb, 2)
	if me.Life != 40 {
		t.Errorf("Lady Loki's controller has %d life, want 40", me.Life)
	}

	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, ids[2])
	if err := castThroughResolutionCast(t, g, me.ID, ids[2], game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the exiled card free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[2]) {
		t.Error("the free cast did not put the card onto the battlefield")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("the free cast cost mana: %d left", me.ManaPool.Total())
	}
}

func TestMscLokiDamageIsTheAbsoluteDifference(t *testing.T) {
	// The spell is the bigger one: mana value 5 against a mana value 1.
	tb := mscLokiNew(t)
	tb.library(game.Card{Name: "Cheap", TypeLine: "Creature — Test", ManaCost: "{R}", Power: 1, Toughness: 1})
	tb.cast(t, "Big Sorcery", "Sorcery", "{4}{R}", "CCCCR", 0)
	tb.resolveTrigger(t)
	mscLokiExpectOpponents(t, tb, 4)
}

func TestMscLokiEqualManaValuesDealNoDamage(t *testing.T) {
	tb := mscLokiNew(t)
	ids := tb.library(game.Card{Name: "Same", TypeLine: "Creature — Test", ManaCost: "{1}{G}", Power: 1, Toughness: 1})
	tb.cast(t, "Same Sorcery", "Sorcery", "{1}{R}", "CR", 0)
	tb.resolveTrigger(t)
	mscLokiExpectOpponents(t, tb, 0)
	mustResolutionCastOffer(t, tb.g, tb.me.ID, game.MayCastKeywordPaidCast, ids[0])
}

func TestMscLokiTheSpellsXCountsAndTheExiledCardsXIsZero(t *testing.T) {
	tb := mscLokiNew(t)
	tb.library(game.Card{Name: "X Card", TypeLine: "Sorcery", ManaCost: "{X}{R}", Power: 0, Toughness: 0})
	// {X}{R} with X=3 is mana value 4 on the stack; the exiled {X}{R} is 1.
	tb.cast(t, "X Spell", "Instant", "{X}{R}", "CCCR", 3)
	tb.resolveTrigger(t)
	mscLokiExpectOpponents(t, tb, 3)
}

func TestMscLokiOnlyTheFirstQualifyingSpellEachTurnTriggers(t *testing.T) {
	tb := mscLokiNew(t)
	g, me := tb.g, tb.me
	tb.library(
		game.Card{Name: "First Hit", TypeLine: "Creature — Test", ManaCost: "{2}", Power: 1, Toughness: 1},
		game.Card{Name: "Second Hit", TypeLine: "Creature — Test", ManaCost: "{6}", Power: 1, Toughness: 1},
	)
	// A plain creature spell first is not one of the three kinds.
	creature := tb.cast(t, "Plain Bear", "Creature — Bear", "{G}", "G", 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) || len(g.PendingChoices) != 0 || me.Library.Size() != 2 {
		t.Fatalf("a creature spell that is not a Villain triggered Lady Loki (library %d)", me.Library.Size())
	}
	tb.cast(t, "First Instant", "Instant", "{R}", "R", 0)
	tb.resolveTrigger(t)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 1 {
		t.Fatalf("the first instant did not walk the library once (library %d)", me.Library.Size())
	}
	mscLokiExpectOpponents(t, tb, 1)

	// A second instant, and a Villain creature spell after them, do not.
	tb.cast(t, "Second Instant", "Instant", "{R}", "R", 0)
	passPriorityAroundTable(t, g)
	tb.cast(t, "Second Villain", "Creature — Human Villain", "{B}", "B", 0)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 1 || len(g.PendingChoices) != 0 {
		t.Errorf("a later qualifying spell triggered Lady Loki again (library %d)", me.Library.Size())
	}
	mscLokiExpectOpponents(t, tb, 1)
}

func TestMscLokiAVillainCreatureSpellIsExiledAndNeverEnters(t *testing.T) {
	tb := mscLokiNew(t)
	g, me := tb.g, tb.me
	tb.library(game.Card{Name: "Hit", TypeLine: "Creature — Test", ManaCost: "{4}", Power: 1, Toughness: 1})
	villain := tb.cast(t, "Mean Villain", "Creature — Human Villain", "{1}{B}", "CB", 0)
	tb.resolveTrigger(t)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !inExile(g, villain) || g.Battlefield.Contains(villain) {
		t.Error("the Villain spell did not stay in exile")
	}
	mscLokiExpectOpponents(t, tb, 2)
}

func TestMscLokiDecliningTheFreeCastLeavesTheCardInExile(t *testing.T) {
	tb := mscLokiNew(t)
	g, me := tb.g, tb.me
	ids := tb.library(game.Card{Name: "Hit", TypeLine: "Creature — Test", ManaCost: "{2}", Power: 1, Toughness: 1})
	tb.cast(t, "Test Instant", "Instant", "{R}", "R", 0)
	tb.resolveTrigger(t)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !inExile(g, ids[0]) || g.Battlefield.Contains(ids[0]) {
		t.Error("the declined card is not waiting in exile")
	}
	mscLokiExpectOpponents(t, tb, 1)
}

func TestMscLokiALibraryOfLandsDealsNoDamageAndOffersNothing(t *testing.T) {
	tb := mscLokiNew(t)
	g, me := tb.g, tb.me
	ids := tb.library(
		game.Card{Name: "Land A", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Land B", TypeLine: "Basic Land — Forest"},
	)
	tb.cast(t, "Test Instant", "Instant", "{R}", "R", 0)
	tb.resolveTrigger(t)
	passPriorityAroundTable(t, g)
	for _, id := range ids {
		if !inExile(g, id) {
			t.Errorf("land %s was not exiled by the walk", id)
		}
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a free cast was offered with no nonland card")
	}
	mscLokiExpectOpponents(t, tb, 0)
}

func TestMscLokiTheRestStillHappensWhenTheSpellIsAlreadyGone(t *testing.T) {
	tb := mscLokiNew(t)
	g, me := tb.g, tb.me
	ids := tb.library(game.Card{Name: "Hit", TypeLine: "Creature — Test", ManaCost: "{4}", Power: 1, Toughness: 1})
	spell := tb.cast(t, "Test Sorcery", "Sorcery", "{1}{R}", "CR", 0)
	// In response the spell is countered: the trigger can't exile it.
	if err := g.CounterSpell(spell, &game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}); err != nil {
		t.Fatalf("counter: %v", err)
	}
	tb.resolveTrigger(t)
	if inExile(g, spell) {
		t.Error("a countered spell was exiled")
	}
	// Its mana value, as it last existed, is still 2: 4 - 2.
	mscLokiExpectOpponents(t, tb, 2)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, ids[0])
}

func TestMscLokiTheSpellNeverResolves(t *testing.T) {
	tb := mscLokiNew(t)
	g, me := tb.g, tb.me
	tb.library(game.Card{Name: "Hit", TypeLine: "Creature — Test", ManaCost: "{2}", Power: 1, Toughness: 1})
	// A burn spell whose resolution would hurt: Lady Loki's exile stops it.
	victim := apaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 1, Toughness: 4})
	floatForTest(g, me, "R")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", OracleID: shockOracleForCopies,
			Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	tb.resolveTrigger(t)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, victim); got != 0 {
		t.Errorf("Shock resolved and dealt %d", got)
	}
	if !inExile(g, id) {
		t.Error("Shock is not in exile")
	}
}
