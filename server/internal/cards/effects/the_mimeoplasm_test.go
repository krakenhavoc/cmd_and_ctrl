package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const theMimeoplasmOracle = "55123455-d435-46d4-b0cd-0c1614343be2"

func TestTheMimeoplasmIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(theMimeoplasmOracle)
	if !ok || spec.Name != "The Mimeoplasm" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

func castTheMimeoplasm(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castGraveyardCopier(t, g, "The Mimeoplasm", "Legendary Creature — Ooze", theMimeoplasmOracle, 0, 0)
}

// mimeoAnswer answers the open copy_target prompt, which must carry
// `use`, with `pick`.
func mimeoAnswer(t *testing.T, g *game.Game, use string, pick uuid.UUID) *game.PendingChoice {
	t.Helper()
	p := waitForCopyPrompt(t, g)
	if p.CopyUse != use {
		t.Fatalf("prompt use = %q, want %q", p.CopyUse, use)
	}
	if err := g.ResolveCopyTarget(p.ID, p.Chooser, pick); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	return p
}

// mimeoSettle passes priority until the stack is empty or a question is
// open.
func mimeoSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32 && !stackFullyEmpty(g) && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// Two creature cards from two graveyards: both are exiled, The Mimeoplasm
// is a copy of the first and enters with +1/+1 counters equal to the
// second's power. The second question never offers the first card again,
// and a noncreature card is never offered.
func TestTheMimeoplasmExilesTwoCopiesOneAndTakesTheOthersPowerAsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)
	wurm := graveyardCreature(me, "Craw Wurm", "Wurm", 6, 4)
	bolt := graveyardCardOf(g, opp, "Lightning Bolt", "Instant", "{R}", "")

	ooze := castTheMimeoplasm(t, g)
	first := mimeoAnswer(t, g, game.CopyUseExileCopied, giant)
	if !slices.Contains(first.CopyOptions, wurm) || slices.Contains(first.CopyOptions, bolt) {
		t.Fatalf("first options = %v; want both creature cards and no instant", first.CopyOptions)
	}
	if inExile(g, giant) || onBattlefieldNamed(g, "Hill Giant") != 0 {
		t.Fatal("something happened before the second card was chosen")
	}
	second := mimeoAnswer(t, g, game.CopyUseExileOther, wurm)
	if slices.Contains(second.CopyOptions, giant) || !slices.Contains(second.CopyOptions, wurm) {
		t.Fatalf("second options = %v; want the Wurm and not the Giant again", second.CopyOptions)
	}
	mimeoSettle(t, g)

	if !inExile(g, giant) || !inExile(g, wurm) {
		t.Fatal("both creature cards should be exiled")
	}
	got := copyBattlefieldCard(t, g, ooze)
	if got.Name != "Hill Giant" || got.Counters[game.CounterPlusOne] != 6 {
		t.Fatalf("The Mimeoplasm = %q with %d counters, want a Hill Giant with 6", got.Name, got.Counters[game.CounterPlusOne])
	}
	if lc := layeredCard(t, g, ooze); lc.CurrentPower() != 9 || lc.CurrentToughness() != 9 || !lc.HasSubtype("Giant") {
		t.Errorf("The Mimeoplasm is %d/%d, want a 9/9 Giant", lc.CurrentPower(), lc.CurrentToughness())
	}
}

// Declining the first question, or the second, exiles nothing, and The
// Mimeoplasm enters as itself, with no counters.
func TestTheMimeoplasmDecliningEitherQuestionExilesNothing(t *testing.T) {
	for _, declineAt := range []int{1, 2} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)
		wurm := graveyardCreature(me, "Craw Wurm", "Wurm", 6, 4)

		ooze := castTheMimeoplasm(t, g)
		if declineAt == 1 {
			mimeoAnswer(t, g, game.CopyUseExileCopied, uuid.Nil)
		} else {
			mimeoAnswer(t, g, game.CopyUseExileCopied, giant)
			mimeoAnswer(t, g, game.CopyUseExileOther, uuid.Nil)
		}
		mimeoSettle(t, g)
		if copyPrompt(g) != nil {
			t.Fatalf("decline at %d: a question is still open", declineAt)
		}
		if !opp.Graveyard.Contains(giant) || !me.Graveyard.Contains(wurm) {
			t.Errorf("decline at %d: a creature card left its graveyard", declineAt)
		}
		if got := copyBattlefieldCard(t, g, ooze); got.Name != "The Mimeoplasm" || got.Counters[game.CounterPlusOne] != 0 {
			t.Errorf("decline at %d: entered as %q with %d counters, want itself with none", declineAt, got.Name, got.Counters[game.CounterPlusOne])
		}
	}
}

// The second answer may not name the first card again: that is not two
// cards, so nothing is exiled.
func TestTheMimeoplasmSameCardTwiceIsADecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)
	wurm := graveyardCreature(me, "Craw Wurm", "Wurm", 6, 4)

	castTheMimeoplasm(t, g)
	mimeoAnswer(t, g, game.CopyUseExileCopied, giant)
	mimeoAnswer(t, g, game.CopyUseExileOther, giant)
	mimeoSettle(t, g)
	if !opp.Graveyard.Contains(giant) || !me.Graveyard.Contains(wurm) || onBattlefieldNamed(g, "Hill Giant") != 0 {
		t.Fatal("naming one card twice exiled or copied something")
	}
}

// One creature card in all graveyards is not two: nothing is asked.
func TestTheMimeoplasmNeedsTwoCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)

	ooze := castTheMimeoplasm(t, g)
	mimeoSettle(t, g)
	if copyPrompt(g) != nil {
		t.Fatal("asked with one creature card in the graveyards")
	}
	if got := copyBattlefieldCard(t, g, ooze); !opp.Graveyard.Contains(giant) || got.Name != "The Mimeoplasm" {
		t.Fatalf("want the Giant untouched and The Mimeoplasm as itself, got %q", got.Name)
	}
}

// A table paused on the second question is a restore point: the restored
// table still knows the first card.
func TestTheMimeoplasmSecondQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	giant := graveyardCreature(opp, "Hill Giant", "Giant", 3, 3)
	wurm := graveyardCreature(me, "Craw Wurm", "Wurm", 6, 4)

	ooze := castTheMimeoplasm(t, g)
	mimeoAnswer(t, g, game.CopyUseExileCopied, wurm)
	waitForCopyPrompt(t, g)
	restored := restoreRoundTrip(t, g, true)
	mimeoAnswer(t, restored, game.CopyUseExileOther, giant)
	mimeoSettle(t, restored)
	got := copyBattlefieldCard(t, restored, ooze)
	if got.Name != "Craw Wurm" || got.Counters[game.CounterPlusOne] != 3 {
		t.Fatalf("restored: The Mimeoplasm = %q with %d counters, want a Craw Wurm with 3", got.Name, got.Counters[game.CounterPlusOne])
	}
	if !inExile(restored, giant) || !inExile(restored, wurm) {
		t.Fatal("restored: both creature cards should be exiled")
	}
}
