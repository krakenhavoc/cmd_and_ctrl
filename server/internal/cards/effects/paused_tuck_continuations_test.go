package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_tuck_continuations_test.go — #783. Three catalog cards that
// kept going while a tuck was PAUSED on the CR 903.9 command-zone
// prompt, plus the God-Eternals' positioned landing.
//
// A library is a CR 903.9 destination, so every tuck can stop to ask
// its owner a question, and `TuckToLibraryForEffect` returns nil
// whichever happened. Card code that wrote the next instruction on the
// next line therefore ran it with the permanent still on the
// battlefield: Chaos Warp shuffled and revealed first, Aetherspouts
// scried a card its owner had no right to look at, and the God-Eternals
// reordered a library the card had not reached yet.
//
// The fix is one continuation (TuckToLibraryThenForEffect /
// TuckCardsToLibraryThenForEffect) and, for the God-Eternals, a
// positioned landing carried on the route.

// tuckCommanderOnBattlefield puts a commander creature onto the
// battlefield under `owner` and returns its ID.
func tuckCommanderOnBattlefield(t *testing.T, g *game.Game, owner *game.Player, name string) uuid.UUID {
	t.Helper()
	id := b12Creature(g, owner.ID, name, "Legendary Creature — Angel", 4, 4)
	markCommanderCard(t, g, owner, id)
	return id
}

// tuckRevealCount counts EventRevealCards for `actor` since `from`.
func tuckRevealCount(g *game.Game, actor uuid.UUID, from int) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventRevealCards && ev.Actor == actor {
			n++
		}
	}
	return n
}

// tuckEffectErrors lists the EventEffectError messages since `from`.
func tuckEffectErrors(g *game.Game, from int) []string {
	var out []string
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventEffectError {
			out = append(out, ev.ErrorMsg)
		}
	}
	return out
}

// --- Chaos Warp ----------------------------------------------------

// TestChaosWarpOnACommanderWaitsForTheCommandZoneAnswer is the
// ordering half of the bug: CR 608.2c runs the instructions in order,
// so "shuffles it into their library, THEN reveals the top card" may
// not reveal anything while the shuffle-in is still a question.
func TestChaosWarpOnACommanderWaitsForTheCommandZoneAnswer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Library.Cards = nil
	for i := 0; i < 4; i++ {
		plTop(opp, "Forest", "Basic Land — Forest", "")
	}
	cmd := tuckCommanderOnBattlefield(t, g, opp, "Their Commander")
	before := len(g.Events)

	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle, b16TargetCard(cmd))
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(cmd) {
		t.Fatal("the commander does not move until its owner answers")
	}
	if n := tuckRevealCount(g, opp.ID, before); n != 0 {
		t.Errorf("%d cards revealed with the CR 903.9 prompt still open", n)
	}
	if opp.Library.Size() != 4 {
		t.Errorf("library = %d, want the four it started with: nothing has been shuffled in yet", opp.Library.Size())
	}

	b21DeclineCommandZone(t, g, opp.ID)

	if g.Battlefield.Contains(cmd) {
		t.Error("a declined commander leaves the battlefield")
	}
	if n := tuckRevealCount(g, opp.ID, before); n != 1 {
		t.Errorf("after the answer, exactly one card is revealed; got %d", n)
	}
	if got := opp.Library.Size() + g.Battlefield.Size() - 1; got < 4 {
		t.Errorf("the warped commander and the four Forests are still accounted for; got %d", got)
	}
}

// TestChaosWarpOnACommanderShufflesItInRatherThanOnTop is the outcome
// half, made deterministic by an EMPTY library: the only card the
// shuffle can have to reveal is the commander itself, and it is there
// only if the shuffle-in really happened before the reveal.
//
// With the old order — reveal first, commander lands on top afterwards
// — there is nothing to reveal and the commander ends up as the top
// card of a one-card library, which is exactly the caveat Chaos Warp
// used to carry.
func TestChaosWarpOnACommanderShufflesItInRatherThanOnTop(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Library.Cards = nil
	cmd := tuckCommanderOnBattlefield(t, g, opp, "Their Commander")
	before := len(g.Events)

	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle, b16TargetCard(cmd))
	passPriorityAroundTable(t, g)
	b21DeclineCommandZone(t, g, opp.ID)

	if n := tuckRevealCount(g, opp.ID, before); n != 1 {
		t.Fatalf("the shuffled-in commander is the top card and is revealed; got %d reveals", n)
	}
	if !g.Battlefield.Contains(cmd) {
		t.Error("the revealed card is a permanent card, so it comes back onto the battlefield")
	}
	if opp.Library.Size() != 0 {
		t.Errorf("library = %d, want 0 — its only card was revealed and put", opp.Library.Size())
	}
	if errs := tuckEffectErrors(g, before); len(errs) != 0 {
		t.Errorf("effect errors: %v", errs)
	}
}

// TestChaosWarpOnACommanderTakenToTheCommandZoneStillShufflesAndReveals
// — the answer is not a gate. "Shuffles it into their library, then
// reveals the top card" is one sentence about the library, not an "if
// you do", so a commander that goes to the command zone instead still
// leaves its owner shuffling and revealing.
func TestChaosWarpOnACommanderTakenToTheCommandZoneStillShufflesAndReveals(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Library.Cards = nil
	for i := 0; i < 4; i++ {
		plTop(opp, "Forest", "Basic Land — Forest", "")
	}
	cmd := tuckCommanderOnBattlefield(t, g, opp, "Their Commander")
	before := len(g.Events)

	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle, b16TargetCard(cmd))
	passPriorityAroundTable(t, g)
	b36AcceptCommandZone(t, g, opp.ID)

	if !opp.Command.Contains(cmd) {
		t.Fatal("the owner took the command zone")
	}
	if g.Battlefield.Contains(cmd) || opp.Library.Contains(cmd) {
		t.Error("the commander is in exactly one place")
	}
	if n := tuckRevealCount(g, opp.ID, before); n != 1 {
		t.Errorf("the reveal happens whichever way the question is answered; got %d", n)
	}
	if errs := tuckEffectErrors(g, before); len(errs) != 0 {
		t.Errorf("effect errors: %v", errs)
	}
}

// TestChaosWarpUndoAcrossTheCommanderPromptReplays — the undo contract
// the continuation signs: rewind into the open prompt, answer the other
// way, and the board follows that answer.
func TestChaosWarpUndoAcrossTheCommanderPromptReplays(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Library.Cards = nil
	cmd := tuckCommanderOnBattlefield(t, g, opp, "Their Commander")
	oppID := opp.ID

	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle, b16TargetCard(cmd))
	passPriorityAroundTable(t, g)
	promptOpen := g.Clone()

	b36AcceptCommandZone(t, g, oppID)
	if !g.Seats[1].Command.Contains(cmd) {
		t.Fatal("accepting puts the commander in the command zone")
	}

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Battlefield.Contains(cmd) {
		t.Fatal("the rewind puts the commander back on the battlefield, with the prompt open")
	}

	b21DeclineCommandZone(t, g, oppID)
	if g.Seats[1].Command.Contains(cmd) {
		t.Error("the replayed answer is the new one")
	}
	if !g.Battlefield.Contains(cmd) {
		t.Error("declining shuffles it in; it is then the only card revealed and comes back")
	}
}

// --- the God-Eternals ----------------------------------------------

// TestGodEternalCommanderReturnsThirdFromTopWithoutErroring is the
// God-Eternal half of #783. Oketra is commonly a commander, so its
// return tuck routinely pauses — and the old remove-and-reinsert then
// reached for a card that was still in the graveyard, logged an effect
// error, and left the God-Eternal on TOP once the owner declined.
func TestGodEternalCommanderReturnsThirdFromTopWithoutErroring(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oketra := b12Push(g, me.ID, "God-Eternal Oketra", "Legendary Creature — Zombie God", b22GodEternalOketraOracle, 3, 6)
	markCommanderCard(t, g, me, oketra)
	if me.Library.Size() < 3 {
		t.Fatalf("test needs a library of at least 3, got %d", me.Library.Size())
	}
	before := len(g.Events)

	// Dies. The graveyard is a CR 903.9 destination too, so the first
	// question is about the death; decline it so the card reaches the
	// graveyard and the dies-trigger fires.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(oketra) })
	b21DeclineCommandZone(t, g, me.ID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	// Second question: the tuck. Nothing has moved yet.
	if !me.Graveyard.Contains(oketra) {
		t.Fatal("the card waits in the graveyard until the tuck's question is answered")
	}
	b21DeclineCommandZone(t, g, me.ID)

	if me.Graveyard.Contains(oketra) {
		t.Fatal("a declined tuck puts the card into the library")
	}
	names := libraryTopNames(me, 3)
	if len(names) != 3 || names[2] != "God-Eternal Oketra" || names[0] == "God-Eternal Oketra" {
		t.Errorf("third from the top, got %v", names)
	}
	if errs := tuckEffectErrors(g, before); len(errs) != 0 {
		t.Errorf("the paused tuck logged effect errors: %v", errs)
	}
}

// TestGodEternalCommanderTakingTheCommandZoneStaysThere — the other
// answer, and the one that has to stay consistent: the God-Eternal's
// "put it into its owner's library third from the top" is a move to a
// CR 903.9 destination, so its owner may take the command zone instead
// and the card must not also be in a library.
func TestGodEternalCommanderTakingTheCommandZoneStaysThere(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oketra := b12Push(g, me.ID, "God-Eternal Oketra", "Legendary Creature — Zombie God", b22GodEternalOketraOracle, 3, 6)
	markCommanderCard(t, g, me, oketra)
	before := len(g.Events)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(oketra) })
	b21DeclineCommandZone(t, g, me.ID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	b36AcceptCommandZone(t, g, me.ID)

	if !me.Command.Contains(oketra) {
		t.Fatal("the owner took the command zone")
	}
	if me.Library.Contains(oketra) || me.Graveyard.Contains(oketra) {
		t.Error("the card is in exactly one place")
	}
	if errs := tuckEffectErrors(g, before); len(errs) != 0 {
		t.Errorf("effect errors: %v", errs)
	}
}

// --- Aetherspouts ---------------------------------------------------

// TestAetherspoutsScriesOnlyTheAttackersThatReachedTheLibrary — the
// scry is the tuck batch's continuation and counts what LANDED. An
// attacking commander whose owner takes the command zone was never put
// into a library (CR 400.7), so it is not among the cards its owner
// arranges — and nothing is arranged at all until the question is
// answered.
func TestAetherspoutsScriesOnlyTheAttackersThatReachedTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	raider := pushVanillaCreature(g, a.ID, "Raider", 2, 2)
	cmd := tuckCommanderOnBattlefield(t, g, a, "Their Commander")
	aangAdvanceToMain(t, g, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{raider, cmd} {
		if err := g.DeclareAttacker(id, b.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	castInPlace(t, g, me.ID, "Aetherspouts", b26AetherspoutsOracle)
	passPriorityAroundTable(t, g)

	if scryChoiceFor(g, a.ID) != nil {
		t.Fatal("nobody scries while the CR 903.9 prompt is open")
	}
	b36AcceptCommandZone(t, g, a.ID)

	if !a.Command.Contains(cmd) {
		t.Fatal("the owner took the command zone")
	}
	c := scryChoiceFor(g, a.ID)
	if c == nil {
		t.Fatal("the owner arranges the attackers that did reach their library")
	}
	if len(c.ScryCards) != 1 || c.ScryCards[0] != raider {
		t.Errorf("scry looks at %v, want only the attacker that landed in the library", c.ScryCards)
	}
}
