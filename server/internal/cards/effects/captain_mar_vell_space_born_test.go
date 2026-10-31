package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captain Mar-Vell, Space-Born (#4570): "As long as an opponent has cast a
// spell this turn, you may cast spells as though they had flash." The
// condition (CastTimingRule.OpponentCastASpellThisTurn) is read live from
// the other seats' per-turn cast tallies. Seat 0 controls Mar-Vell and seat 1
// is the opponent whose turn it is, unless a test says otherwise.

const marVellName = "Captain Mar-Vell, Space-Born"

func marVell(t *testing.T, g *game.Game, who uuid.UUID) uuid.UUID {
	t.Helper()
	return hobPush(t, g, who, marVellName, "Legendary Creature — Kree Soldier Hero", 4, 4)
}

// marVellTimingOpen asks the engine whether p may begin to cast a spell of
// the given type line out of their hand right now. It is the question the
// cast gate, the bot's enumerator and the client's castable_here all ask.
func marVellTimingOpen(g *game.Game, p *game.Player, typeLine string) bool {
	var open bool
	g.WithWriteLock(func() {
		open = g.CastTimingOpenLocked(p.ID, game.Card{
			InstanceID: uuid.New(), Name: "Timing Probe", TypeLine: typeLine,
			Owner: p.ID, Controller: p.ID,
		}, game.ZoneHand, nil)
	})
	return open
}

// marVellOpponentCasts walks to seat 1's precombat main and has seat 1 cast
// an instant. The spell stays on the stack with seat 1 holding priority.
func marVellOpponentCasts(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToMainOf(t, g, 1)
	castFromHandForTest(t, g, g.Seats[1], "Opposing Bolt", "Instant", "", "", strict)
}

func marVellTally(g *game.Game, who uuid.UUID) int {
	var n int
	g.ReadSnapshot(func() { n = g.CastTallyFor(who).Total })
	return n
}

func TestMarVellRefusesASorceryAtInstantSpeedWhileNoOpponentHasCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marVell(t, g, me.ID)
	advanceTo(t, g, game.StepBeginCombat)
	if marVellTimingOpen(g, me, "Sorcery") {
		t.Error("a sorcery could begin to be cast at the beginning of combat with no opponent spell this turn")
	}
	id := handCardFull(me, "Mystery Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("a sorcery was cast at instant speed with no opponent spell this turn")
	}
	if marVellTimingOpen(g, me, "Creature — Bear") {
		t.Error("a creature could begin to be cast at instant speed with no opponent spell this turn")
	}
}

func TestMarVellLetsYouCastASorceryAndACreatureInResponseToAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marVell(t, g, me.ID)
	advanceToMainOf(t, g, 1)
	// Seat 1's turn, before they have cast anything: not seat 0's window.
	if marVellTimingOpen(g, me, "Sorcery") || marVellTimingOpen(g, me, "Creature — Bear") {
		t.Fatal("seat 0 could cast sorcery-speed spells on seat 1's turn before any opponent spell")
	}
	castFromHandForTest(t, g, g.Seats[1], "Opposing Bolt", "Instant", "", "", strict)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !marVellTimingOpen(g, me, "Sorcery") || !marVellTimingOpen(g, me, "Creature — Bear") {
		t.Fatal("seat 0 could not begin sorcery-speed spells after an opponent cast a spell")
	}
	sorcery := handCardFull(me, "Mystery Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("a sorcery could not be cast during an opponent's turn: %v", err)
	}
	bear := rfdHandCard(me, "Late Bear", "Creature — Bear", "", "", 2, 2)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("a creature could not be cast during an opponent's turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := hobB6BattlefieldCount(g, me.ID, "Late Bear"); got != 1 {
		t.Errorf("Late Bear is on the battlefield %d times, want 1", got)
	}
}

func TestMarVellLetsYouCastASorceryOnYourOwnTurnAtInstantSpeedAfterAnOpponentCasts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marVell(t, g, me.ID)
	advanceToMainOf(t, g, 0)
	// Seat 0's own instant, then seat 1 answers it: the opponent's spell is
	// what opens the window, and it stays open after both resolve.
	castFromHandForTest(t, g, me, "My Instant", "Instant", "", "", strict)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castFromHandForTest(t, g, g.Seats[1], "Opposing Bolt", "Instant", "", "", strict)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	if got := marVellTally(g, g.Seats[1].ID); got != 1 {
		t.Fatalf("the opponent's tally is %d, want 1", got)
	}
	id := handCardFull(me, "Mystery Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Errorf("an instant-speed sorcery was refused after an opponent cast a spell: %v", err)
	}
}

func TestMarVellYourOwnSpellsDoNotOpenTheWindow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marVell(t, g, me.ID)
	advanceToMainOf(t, g, 0)
	castFromHandForTest(t, g, me, "My Instant", "Instant", "", "", strict)
	passPriorityAroundTable(t, g)
	if got := marVellTally(g, me.ID); got != 1 {
		t.Fatalf("my tally is %d, want 1", got)
	}
	advanceTo(t, g, game.StepBeginCombat)
	if marVellTimingOpen(g, me, "Sorcery") {
		t.Error("my own spell opened the instant-speed window")
	}
	id := handCardFull(me, "Mystery Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("a sorcery was cast at instant speed after only my own spell")
	}
}

func TestMarVellWindowShutsWhenTheTurnEnds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marVell(t, g, me.ID)
	marVellOpponentCasts(t, g)
	passPriorityAroundTable(t, g)
	if !marVellTimingOpen(g, me, "Sorcery") {
		t.Fatal("the window was shut on the turn the opponent cast")
	}
	// Seat 0's next turn: nobody has cast anything yet.
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepBeginCombat); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepBeginCombat {
		t.Fatalf("never reached seat 0's beginning of combat (seat %d, %s)", g.Turn.ActiveSeat, g.Turn.Step)
	}
	if marVellTally(g, g.Seats[1].ID) != 0 {
		t.Fatal("the opponent's tally survived into the next turn")
	}
	if marVellTimingOpen(g, me, "Sorcery") {
		t.Error("the window stayed open into the next turn")
	}
}

func TestMarVellEnteringAfterTheOpponentsSpellStillCountsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marVellOpponentCasts(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if marVellTimingOpen(g, me, "Sorcery") {
		t.Fatal("the window was open before Mar-Vell was on the battlefield")
	}
	marVell(t, g, me.ID)
	if !marVellTimingOpen(g, me, "Sorcery") {
		t.Error("a Mar-Vell that entered after the opponent's spell did not see it")
	}
	id := handCardFull(me, "Mystery Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Errorf("a sorcery could not be cast in response with a late Mar-Vell: %v", err)
	}
}

func TestMarVellGivesNothingOnceItLeavesTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mv := marVell(t, g, me.ID)
	marVellOpponentCasts(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !marVellTimingOpen(g, me, "Sorcery") {
		t.Fatal("the window was shut with Mar-Vell on the battlefield")
	}
	hobDestroy(t, g, mv)
	if marVellTimingOpen(g, me, "Sorcery") {
		t.Error("the window stayed open after Mar-Vell left")
	}
}

func TestMarVellDoesNotCountAPlayerWhoHasLeftTheGame(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marVell(t, g, me.ID)
	marVellOpponentCasts(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !marVellTimingOpen(g, me, "Sorcery") {
		t.Fatal("the window was shut with an opponent's spell on the stack")
	}
	g.WithWriteLock(func() { g.Seats[1].Eliminated = true })
	if marVellTimingOpen(g, me, "Sorcery") {
		t.Error("a player who left the game still counted as an opponent who cast a spell")
	}
}

func TestMarVellOnlyHelpsItsController(t *testing.T) {
	g := newCatalogGame(t)
	marVell(t, g, g.Seats[0].ID)
	marVellOpponentCasts(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	// Seat 2 is an opponent of seat 0 as well, but the statement says "you".
	if marVellTimingOpen(g, g.Seats[2], "Sorcery") {
		t.Error("an opponent of Mar-Vell's controller got the flash grant")
	}
}

func TestMarVellTallySurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	marVell(t, g, g.Seats[0].ID)
	marVellOpponentCasts(t, g)
	passPriorityAroundTable(t, g)
	if got := marVellTally(g, g.Seats[1].ID); got != 1 {
		t.Fatalf("the opponent's tally is %d before the restore, want 1", got)
	}

	restored := restoreThroughJSON(t, g)
	if got := marVellTally(restored, restored.Seats[1].ID); got != 1 {
		t.Errorf("the opponent's tally is %d after the restore, want 1", got)
	}
	if !marVellTimingOpen(restored, restored.Seats[0], "Sorcery") {
		t.Error("the restored table does not let Mar-Vell's controller cast a sorcery at instant speed")
	}
	if marVellTimingOpen(restored, restored.Seats[2], "Sorcery") {
		t.Error("the restored table gave the flash grant to the wrong seat")
	}
}

func TestMarVellRestoreWithNoOpponentSpellStaysClosed(t *testing.T) {
	g := newCatalogGame(t)
	marVell(t, g, g.Seats[0].ID)
	advanceToMainOf(t, g, 1)
	restored := restoreThroughJSON(t, g)
	if marVellTimingOpen(restored, restored.Seats[0], "Sorcery") {
		t.Error("the restored table opened the window with no opponent spell this turn")
	}
}
