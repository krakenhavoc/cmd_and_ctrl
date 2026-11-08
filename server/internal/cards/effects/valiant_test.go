package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// valiant_test.go — the valiant seam (#5230), driven through Heartfire
// Hero's counter (an untargeted valiant ability) and Mouse Trapper's tap (a
// targeted one, whose trigger waits on a target prompt before it is
// announced).

const (
	valiantHeartfireHeroOracle = "49e1e202-37bf-46ed-93d1-1adb13726039"
	valiantMouseTrapperOracle  = "5b0a4530-6e19-41a2-9ea6-e4c8cf85b6c4"
)

type valiantTable struct {
	g        *game.Game
	me, them *game.Player
	seat     int
	hero     uuid.UUID
}

// newValiantTable is a Heartfire Hero under the active player, in their
// precombat main phase.
func newValiantTable(t *testing.T) valiantTable {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	tb := valiantTable{g: g, seat: seat, me: g.Seats[seat], them: g.Seats[(seat+1)%len(g.Seats)]}
	tb.hero = b12Push(g, tb.me.ID, "Heartfire Hero", "Creature — Mouse Soldier", valiantHeartfireHeroOracle, 1, 1)
	return tb
}

// valiantTarget casts an uncatalogued instant for `who` with one target
// slot per id in `targets`, and does not resolve it.
func valiantTarget(t *testing.T, g *game.Game, who *game.Player, targets ...uuid.UUID) {
	t.Helper()
	spell := game.NewCard("Test Pump", who.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{W}"
	who.Hand.PushTop(spell)
	refs := make([]game.TargetRef, 0, len(targets))
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.CastSpell(who.ID, spell.InstanceID, game.CastSpellParams{Targets: refs}); err != nil {
		t.Fatalf("cast the targeting spell: %v", err)
	}
}

// valiantTriggersFrom counts the abilities of `source` waiting to resolve:
// on the stack or still queued.
func valiantTriggersFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source {
			n++
		}
	}
	return n
}

func TestValiantFiresTheFirstTimeYouTargetItThisTurn(t *testing.T) {
	tb := newValiantTable(t)
	valiantTarget(t, tb.g, tb.me, tb.hero)
	if n := valiantTriggersFrom(tb.g, tb.hero); n != 1 {
		t.Fatalf("%d valiant triggers on the stack, want 1", n)
	}
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, tb.hero); n != 1 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters, want 1", n)
	}
}

func TestValiantDoesNotFireASecondTimeInTheSameTurn(t *testing.T) {
	tb := newValiantTable(t)
	valiantTarget(t, tb.g, tb.me, tb.hero)
	passPriorityAroundTable(t, tb.g)
	valiantTarget(t, tb.g, tb.me, tb.hero)
	if n := valiantTriggersFrom(tb.g, tb.hero); n != 0 {
		t.Fatalf("the second spell this turn put %d valiant triggers on the stack", n)
	}
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, tb.hero); n != 1 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters after two spells in one turn, want 1", n)
	}
}

func TestValiantFiresAgainOnTheNextTurn(t *testing.T) {
	tb := newValiantTable(t)
	valiantTarget(t, tb.g, tb.me, tb.hero)
	passPriorityAroundTable(t, tb.g)
	toNextTurnMainOf(t, tb.g, (tb.seat+1)%len(tb.g.Seats))
	// An opponent's turn is still "each turn": the controller's instant then
	// is the first time this turn.
	valiantTarget(t, tb.g, tb.me, tb.hero)
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, tb.hero); n != 2 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters after one spell on each of two turns, want 2", n)
	}
}

func TestValiantIgnoresAnOpponentsSpell(t *testing.T) {
	tb := newValiantTable(t)
	valiantTarget(t, tb.g, tb.them, tb.hero)
	if n := valiantTriggersFrom(tb.g, tb.hero); n != 0 {
		t.Fatalf("an opponent's spell put %d valiant triggers on the stack", n)
	}
	passPriorityAroundTable(t, tb.g)
	// The opponent's targeting does not use up the turn's first time.
	valiantTarget(t, tb.g, tb.me, tb.hero)
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, tb.hero); n != 1 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters, want 1 (from its controller's spell only)", n)
	}
}

func TestValiantFiresOnceForOneSpellTargetingItInTwoSlots(t *testing.T) {
	tb := newValiantTable(t)
	valiantTarget(t, tb.g, tb.me, tb.hero, tb.hero)
	if n := valiantTriggersFrom(tb.g, tb.hero); n != 1 {
		t.Fatalf("one spell with two target slots on Heartfire Hero put %d valiant triggers on the stack, want 1", n)
	}
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, tb.hero); n != 1 {
		t.Fatalf("Heartfire Hero has %d +1/+1 counters, want 1", n)
	}
}

// CR 400.7: a Hero that left and came back is a new object that has not
// been targeted this turn. A bounce and a recast keep the card's instance
// ID (only an exile return mints a new one), so this is the case the
// zone-change reset in the scan is for.
func TestValiantFiresAgainForANewObjectTheSameTurn(t *testing.T) {
	tb := newValiantTable(t)
	valiantTarget(t, tb.g, tb.me, tb.hero)
	passPriorityAroundTable(t, tb.g)
	tb.g.WithWriteLock(func() {
		if err := tb.g.BounceToHandForEffect(tb.hero); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	if !tb.me.Hand.Contains(tb.hero) {
		t.Fatal("Heartfire Hero is not in its owner's hand under the same ID")
	}
	if err := tb.g.CastSpell(tb.me.ID, tb.hero, game.CastSpellParams{}); err != nil {
		t.Fatalf("recast Heartfire Hero: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	if !onBoard(tb.g, tb.hero) {
		t.Fatal("Heartfire Hero did not come back under the same ID")
	}
	if n := plusOneCounters(tb.g, tb.hero); n != 0 {
		t.Fatalf("the returned Hero kept %d counters", n)
	}
	valiantTarget(t, tb.g, tb.me, tb.hero)
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, tb.hero); n != 1 {
		t.Fatalf("the returned Hero has %d +1/+1 counters after its first targeting, want 1", n)
	}
}

// --- a targeted valiant ability --------------------------------------

func newMouseTrapperTable(t *testing.T) (valiantTable, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	tb := valiantTable{g: g, seat: seat, me: g.Seats[seat], them: g.Seats[(seat+1)%len(g.Seats)]}
	tb.hero = b12Push(g, tb.me.ID, "Mouse Trapper", "Creature — Mouse Soldier", valiantMouseTrapperOracle, 3, 2)
	return tb, tb.hero
}

func valiantPickTargetPrompts(g *game.Game) int {
	n := 0
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoicePickTarget {
			n++
		}
	}
	return n
}

// The trigger of a targeted valiant ability waits on its target prompt, so
// it has not been announced when the spell's second slot is: the gate must
// still say "already targeted".
func TestValiantTargetedAbilityFiresOnceForTwoSlots(t *testing.T) {
	tb, trapper := newMouseTrapperTable(t)
	// Two candidates, so the target is a real prompt and not an automatic pick.
	bear := b12Creature(tb.g, tb.them.ID, "Opposing Bear", "Creature — Bear", 2, 2)
	other := b12Creature(tb.g, tb.them.ID, "Other Bear", "Creature — Bear", 2, 2)
	valiantTarget(t, tb.g, tb.me, trapper, trapper)
	if n := valiantPickTargetPrompts(tb.g) + valiantTriggersFrom(tb.g, trapper); n != 1 {
		t.Fatalf("one two-slot spell made %d Mouse Trapper triggers (prompts plus items), want 1", n)
	}
	answerPickTarget(t, tb.g, bear)
	passPriorityAroundTable(t, tb.g)
	if !tappedNow(tb.g, bear) {
		t.Error("Mouse Trapper did not tap the opposing creature")
	}
	if tappedNow(tb.g, other) || tappedNow(tb.g, trapper) {
		t.Error("Mouse Trapper tapped something it did not target")
	}
}

// CR 603.3d: with no creature an opponent controls, the trigger is removed,
// but the Trapper has been targeted, so a later spell that turn does not
// trigger it either.
func TestValiantRemovedForNoTargetStillUsesTheTurn(t *testing.T) {
	tb, trapper := newMouseTrapperTable(t)
	valiantTarget(t, tb.g, tb.me, trapper)
	if n := valiantPickTargetPrompts(tb.g) + valiantTriggersFrom(tb.g, trapper); n != 0 {
		t.Fatalf("Mouse Trapper triggered %d times with nothing to tap", n)
	}
	passPriorityAroundTable(t, tb.g)
	b12Creature(tb.g, tb.them.ID, "Late Bear", "Creature — Bear", 2, 2)
	valiantTarget(t, tb.g, tb.me, trapper)
	if n := valiantPickTargetPrompts(tb.g) + valiantTriggersFrom(tb.g, trapper); n != 0 {
		t.Fatalf("Mouse Trapper triggered %d times on the second targeting of the turn", n)
	}
}
