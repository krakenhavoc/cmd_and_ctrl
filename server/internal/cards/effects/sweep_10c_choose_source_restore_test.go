package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_10c_choose_source_restore_test.go — #4253 slice 10c: a
// choose_source prompt ("choose a source of damage", CR 609.7a) is a
// restore point. The shield, redirection or multiplier the card described
// rides the prompt as data, and the answer, given on the restored table,
// finishes it against the pinned source.

// Desperate Gambit asks for a source you control, then calls a coin: the
// restored answer opens the call, and the flip leaves the source with a
// doubled or a prevented next instance.
func TestDesperateGambitSourcePromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := p2Tough(g, me.ID, "Mine")
	theirs := p2Tough(g, opp.ID, "Theirs")
	castCatalogSpell(t, g, "Desperate Gambit", "Instant", p2DesperateOracle, nil)
	passPriorityAroundTable(t, g)

	g = restoreAtKeyedPrompt(t, g, "choose-source/desperate-gambit")
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pick := latestChoiceOfKind(g, game.PendingChoiceChooseSource)
	if pick == nil {
		t.Fatal("the restored table lost the choose_source prompt")
	}
	if !slices.Contains(pick.ChooseCards, mine) || slices.Contains(pick.ChooseCards, theirs) {
		t.Fatalf("restored candidates %v: want your sources only", pick.ChooseCards)
	}
	if err := g.ResolveChooseSource(pick.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	flip := latestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if flip == nil {
		t.Fatal("the restored answer did not open the coin call")
	}
	if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
		t.Fatalf("coin: %v", err)
	}
	won := false
	if flips := randomEvents(g, game.EventFlipCoin); len(flips) > 0 {
		won = flips[len(flips)-1].Won
	}
	start := lifeOf(g, opp.ID)
	p2Ping(t, g, mine, opp.ID, 3)
	want := start // prevented
	if won {
		want = start - 6 // doubled
	}
	if got := lifeOf(g, opp.ID); got != want {
		t.Errorf("won %v: life %d, want %d", won, got, want)
	}
}

// Circle of Protection: Red's prompt, restored, still offers only red
// sources, and the restored answer shields against the chosen one once.
func TestCircleOfProtectionSourcePromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cop := pushCatalogPermanent(g, me.ID, "Circle of Protection: Red", "Enchantment", pr7CoPRed, false)
	red := pr7Creature(g, opp.ID, "Red Dragon", 5, "R")
	green := pr7Creature(g, opp.ID, "Green Bear", 2, "G")
	pr7Activate(t, g, me.ID, cop, 0, game.ActivateAbilityParams{})

	g = restoreAtKeyedPrompt(t, g, "choose-source/shield")
	me = g.Seats[0]
	c := pr7SourcePrompt(t, g)
	if !pr7Offers(c, red) || pr7Offers(c, green) || c.Chooser != me.ID {
		t.Fatalf("restored prompt for %v offers %v: want the red creature and not the green one", c.Chooser, c.ChooseCards)
	}
	pr7Choose(t, g, me.ID, red)
	life := me.Life
	pr7Hit(t, g, green, me.ID, 2)
	pr7Hit(t, g, red, me.ID, 5)
	if me.Life != life-2 {
		t.Fatalf("life %d → %d, want only the green 2", life, me.Life)
	}
}

// Deflecting Palm's shield has a follow-up body (the damage prevented is
// dealt back): it is registered data on the restored prompt.
func TestDeflectingPalmSourcePromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 4, "G")
	castCatalogSpell(t, g, "Deflecting Palm", "Instant", pr7DeflectingPalm, nil)
	passPriorityAroundTable(t, g)

	g = restoreAtKeyedPrompt(t, g, "choose-source/shield")
	me, opp := g.Seats[0], g.Seats[1]
	pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 4)
	if me.Life != mine || opp.Life != theirs-4 {
		t.Errorf("me %d → %d (want unchanged), them %d → %d (want -4)", mine, me.Life, theirs, opp.Life)
	}
}

// Harm's Way is a redirection: 2 of the chosen source's damage goes to the
// target; the restored prompt carries the redirection and its target.
func TestHarmsWaySourcePromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	src := pr7Creature(g, opp.ID, "Dragon", 5, "R")
	castCatalogSpell(t, g, "Harm's Way", "Instant", pr9HarmsWay, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	g = restoreAtKeyedPrompt(t, g, "choose-source/shield")
	me, opp := g.Seats[0], g.Seats[1]
	pr7Choose(t, g, me.ID, src)
	mine, theirs := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 5)
	if me.Life != mine-3 || opp.Life != theirs-2 {
		t.Fatalf("life me %d (want %d), opponent %d (want %d)", me.Life, mine-3, opp.Life, theirs-2)
	}
}
