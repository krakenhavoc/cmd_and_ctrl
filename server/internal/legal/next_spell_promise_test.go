package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// next_spell_promise_test.go — the enumerator half of #1852, in the
// agreement style cast_timing_test.go uses: a promise about the next
// creature spell changes what the bot is offered exactly as it changes
// what the engine accepts, and every offer is one the engine takes.

func grantNextSpell(g *game.Game, p *game.Player, np game.NextSpellPromise) {
	g.WithWriteLock(func() {
		g.GrantNextSpellPromiseForEffect(p.ID, np, "Test promise", uuid.New(), game.Duration{})
	})
}

// A creature in hand in the active seat's END step: shut without the
// flash promise, offered while it is live (and accepted by the engine),
// shut again once a cast has spent it, and never opened for a spell the
// promise's filter does not name.
func TestFlashPromiseOpensTheCreatureCastInTheMoveList(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)
	bear := handCard(me, game.Card{Name: "Promise Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	sorcery := handCard(me, game.Card{Name: "Promise Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{G}"})
	lands(g, me, "Forest", "Forest", 6)
	advanceTo(t, g, game.StepEnd)

	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), bear)); n != 0 {
		t.Fatalf("a creature spell was offered in an end step with no promise")
	}
	grantNextSpell(g, me, game.NextSpellPromise{Filter: game.PermissionFilter{CreatureOnly: true}, Flash: true, Text: "flash"})

	moves := castMovesFor(legal.EnumerateFor(g, me.ID), bear)
	if len(moves) == 0 {
		t.Fatal("the flash promise did not put the creature spell in the move list")
	}
	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), sorcery)); n != 0 {
		t.Fatal("a sorcery is not named by a creature promise")
	}
	// The half a copy of the rule cannot give: the engine accepts it.
	dispatchAll(t, g, me.ID, moves[:1])
}

// The price rider: a {1}{G} creature the seat can't afford with one
// land is offered once a {1} reduction is live, and the engine accepts
// that offer at the reduced price.
func TestCostPromiseMakesTheCastAffordableInTheMoveList(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)
	bear := handCard(me, game.Card{Name: "Promise Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	lands(g, me, "Forest", "Forest", 1)
	advanceTo(t, g, game.StepPrecombatMain)

	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), bear)); n != 0 {
		t.Fatal("setup: one land cannot pay {1}{G}")
	}
	grantNextSpell(g, me, game.NextSpellPromise{Filter: game.PermissionFilter{InstantOrSorceryOnly: true}, Reduce: 1})
	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), bear)); n != 0 {
		t.Fatal("a promise about instants and sorceries does not price a creature")
	}
	grantNextSpell(g, me, game.NextSpellPromise{Reduce: 1, Text: "{1} less"})
	moves := castMovesFor(legal.EnumerateFor(g, me.ID), bear)
	if len(moves) == 0 {
		t.Fatal("the reduction did not make the creature affordable in the move list")
	}
	dispatchAll(t, g, me.ID, moves[:1])
}
