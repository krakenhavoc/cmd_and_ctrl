package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// explosion_of_riches_test.go is #4052: one yes/no question to each other
// player in turn order, and one damage trigger per card drawn this way,
// each choosing its opponent at random as it goes on the stack.

const eorOracle = "ffd469b4-413d-49f5-a873-e3f654240b31"

// eorAnswer answers `chooser`'s open "draw a card?" question (the option
// pick Explosion of Riches asked): yes draws, no does not.
func eorAnswer(t *testing.T, g *game.Game, chooser uuid.UUID, yes bool) {
	t.Helper()
	c := otOptionPick(t, g, chooser)
	index := 0
	if yes {
		index = 1
	}
	if err := g.ResolveOptionPick(c.ID, chooser, index); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
}

// eorNoQuestionFor fails if `player` has an open option pick.
func eorNoQuestionFor(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == player {
			t.Fatalf("%s was asked a question", player)
		}
	}
}

// eorDrawn is how many cards the player has drawn this turn.
func eorDrawn(g *game.Game, p *game.Player) int { return len(g.DrawnThisTurn[p.ID]) }

// eorSettle passes priority until the stack is empty. Two or more of the
// damage triggers at once, aimed at different opponents, are not alike, so
// their controller is asked for an order; any order will do.
func eorSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		passPriorityAroundTable(t, g)
		c := pendingOfKind(g, game.PendingChoiceTriggerOrder)
		if c == nil {
			return
		}
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	t.Fatal("the triggers never settled")
}

// You draw, then each other player in turn order may; every accepted draw
// is a trigger of five damage to an opponent of yours, and never to you.
func TestExplosionOfRichesAsksInTurnOrderAndHitsOpponentsPerCardDrawn(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	lives := map[uuid.UUID]int{me.ID: me.Life, a.ID: a.Life, b.ID: b.Life, c.ID: c.Life}

	castCatalogSpell(t, g, "Explosion of Riches", "Sorcery", eorOracle, nil)
	passPriorityAroundTable(t, g)

	// You drew already; the question goes to the seat after you first.
	// Two: the turn's draw step, which the harness walks through to reach
	// the main phase, and the spell's.
	if got := eorDrawn(g, me); got != 2 {
		t.Fatalf("the caster drew %d cards, want 2", got)
	}
	eorAnswer(t, g, a.ID, true)
	eorNoQuestionFor(t, g, c.ID) // the third seat waits its turn
	eorAnswer(t, g, b.ID, false)
	eorAnswer(t, g, c.ID, true)
	eorSettle(t, g)

	if eorDrawn(g, a) != 1 || eorDrawn(g, b) != 0 || eorDrawn(g, c) != 1 {
		t.Errorf("draws a=%d b=%d c=%d, want 1, 0, 1", eorDrawn(g, a), eorDrawn(g, b), eorDrawn(g, c))
	}
	if me.Life != lives[me.ID] {
		t.Errorf("the caster lost %d life: the damage is to opponents", lives[me.ID]-me.Life)
	}
	lost := 0
	for _, p := range []*game.Player{a, b, c} {
		lost += lives[p.ID] - p.Life
	}
	// Three cards were drawn this way (yours and two accepted): three
	// triggers, 15 damage between the opponents.
	if lost != 15 {
		t.Errorf("opponents lost %d life in all, want 15 (three triggers of 5)", lost)
	}
	for _, p := range []*game.Player{a, b, c} {
		if d := lives[p.ID] - p.Life; d%5 != 0 {
			t.Errorf("%s lost %d, not a multiple of 5", p.ID, d)
		}
	}
}

// Nobody accepts: your own draw still triggers once.
func TestExplosionOfRichesOwnDrawTriggersAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	total := a.Life + b.Life + c.Life
	castCatalogSpell(t, g, "Explosion of Riches", "Sorcery", eorOracle, nil)
	passPriorityAroundTable(t, g)
	eorAnswer(t, g, a.ID, false)
	eorAnswer(t, g, b.ID, false)
	eorAnswer(t, g, c.ID, false)
	eorSettle(t, g)
	if got := total - (a.Life + b.Life + c.Life); got != 5 {
		t.Errorf("opponents lost %d in all, want 5 (your own draw)", got)
	}
	if me.Life != game.StartingLife {
		t.Errorf("caster life %d", me.Life)
	}
}

// Each trigger picks its own opponent at random: over many tables all
// three opponents are hit and the caster never is.
func TestExplosionOfRichesTargetsVaryAcrossOpponents(t *testing.T) {
	hit := map[int]bool{}
	for seed := 0; seed < 40 && len(hit) < 3; seed++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(seed)))
		me := g.Seats[0]
		castCatalogSpell(t, g, "Explosion of Riches", "Sorcery", eorOracle, nil)
		passPriorityAroundTable(t, g)
		for _, p := range g.Seats[1:] {
			eorAnswer(t, g, p.ID, false)
		}
		eorSettle(t, g)
		if me.Life != game.StartingLife {
			t.Fatalf("seed %d: the caster was hit", seed)
		}
		for i, p := range g.Seats[1:] {
			if p.Life < game.StartingLife {
				hit[i+1] = true
			}
		}
	}
	if len(hit) != 3 {
		t.Errorf("only %d of 3 opponents were ever chosen: the target is not random", len(hit))
	}
}

// A player who has left the game is not asked, and a player a draw limit
// stops is not asked either.
func TestExplosionOfRichesSkipsPlayersWhoCannotDraw(t *testing.T) {
	g := newCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	g.WithWriteLock(func() { a.Eliminated = true })
	castCatalogSpell(t, g, "Explosion of Riches", "Sorcery", eorOracle, nil)
	passPriorityAroundTable(t, g)
	eorNoQuestionFor(t, g, a.ID)
	eorAnswer(t, g, b.ID, true)
	eorAnswer(t, g, g.Seats[3].ID, false)
	eorSettle(t, g)
	if eorDrawn(g, b) != 1 {
		t.Errorf("b drew %d, want 1", eorDrawn(g, b))
	}
}

// The question is a keyed continuation: a table waiting on an opponent's
// answer is a restore point, and the restored table finishes the card.
func TestExplosionOfRichesIsARestorePointWhileAsking(t *testing.T) {
	g := newCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	castCatalogSpell(t, g, "Explosion of Riches", "Sorcery", eorOracle, nil)
	passPriorityAroundTable(t, g)
	eorAnswer(t, g, a.ID, true)
	// b's question is open.
	restored := restoreRoundTrip(t, g, true)
	eorAnswer(t, restored, b.ID, true)
	eorAnswer(t, restored, restored.Seats[3].ID, false)
	eorSettle(t, restored)
	for _, id := range []uuid.UUID{a.ID, b.ID} {
		if n := len(restored.DrawnThisTurn[id]); n != 1 {
			t.Errorf("restored: %s drew %d, want 1", id, n)
		}
	}
	total := 0
	for _, p := range restored.Seats[1:] {
		total += game.StartingLife - p.Life
	}
	if total != 15 {
		t.Errorf("restored: opponents lost %d in all, want 15", total)
	}
}

// A draw that puts no card in a hand is not "a card drawn this way": a
// player with an empty library may say yes, draws nothing, and triggers
// nothing (they lose to the empty draw instead, CR 704.5b). The triggers
// are counted rather than their damage, because the player who loses can
// be the random target of a trigger already waiting.
func TestExplosionOfRichesAnEmptyLibraryDrawsNoCardAndTriggersNothing(t *testing.T) {
	g := newCatalogGame(t)
	a, b, c := g.Seats[1], g.Seats[2], g.Seats[3]
	g.WithWriteLock(func() { c.Library.Cards = nil })
	castCatalogSpell(t, g, "Explosion of Riches", "Sorcery", eorOracle, nil)
	passPriorityAroundTable(t, g)
	eorAnswer(t, g, a.ID, true)
	eorAnswer(t, g, b.ID, false)
	eorAnswer(t, g, c.ID, true)
	if eorDrawn(g, c) != 0 {
		t.Fatalf("the player with no library drew %d cards", eorDrawn(g, c))
	}
	triggers := 0
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.SourceCardID != uuid.Nil && strings.Contains(it.Label, "Explosion of Riches") {
				triggers++
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && strings.Contains(it.Label, "Explosion of Riches") {
				triggers++
			}
		}
	})
	if triggers != 2 {
		t.Errorf("%d damage triggers, want 2: your draw and one accepted draw, not the empty one", triggers)
	}
}
