package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// clash_cards_test.go — #5424: clash (CR 701.30) and the cards built on
// it, Hoarder's Greed and Marvo, Deep Operative.

const (
	hoardersGreedOracle = "60040023-7bf0-42e0-ab30-46a730675564"
	marvoOracle         = "317ab0b3-1163-4893-a881-4c397feb213d"
)

// clashTop puts a card with the given mana cost on top of p's library.
func clashTop(p *game.Player, name, typeLine, cost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = cost
	p.Library.PushTop(c)
	return c.InstanceID
}

// clashPrompts is the open option picks, in queue order.
func clashPrompts(g *game.Game) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			out = append(out, c)
		}
	}
	return out
}

// clashOnlyPrompt fails unless exactly one option pick is open and it is
// chooser's, and returns it.
func clashOnlyPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	open := clashPrompts(g)
	if len(open) != 1 {
		t.Fatalf("%d option picks are open, want 1", len(open))
	}
	if open[0].Chooser != chooser {
		t.Fatalf("the open question is %q for another player, want one for %s", open[0].Reason, chooser)
	}
	return open[0]
}

// clashChooseOpponent answers "clash with an opponent".
func clashChooseOpponent(t *testing.T, g *game.Game, chooser, opponent uuid.UUID) {
	t.Helper()
	c := clashOnlyPrompt(t, g, chooser)
	for i, o := range c.PickOptions {
		if o.Player == opponent {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("choose the opponent: %v", err)
			}
			return
		}
	}
	t.Fatalf("the question %q does not offer that opponent", c.Reason)
}

// clashDecide answers chooser's top-or-bottom question, which must be the
// only question open.
func clashDecide(t *testing.T, g *game.Game, chooser uuid.UUID, bottom bool) *game.PendingChoice {
	t.Helper()
	c := clashOnlyPrompt(t, g, chooser)
	if len(c.PickOptions) != 2 || !strings.Contains(c.PickOptions[1].Label, "bottom") {
		t.Fatalf("the question %q is not a clash decision: %+v", c.Reason, c.PickOptions)
	}
	index := 0
	if bottom {
		index = 1
	}
	if err := g.ResolveOptionPick(c.ID, chooser, index); err != nil {
		t.Fatalf("decide: %v", err)
	}
	return c
}

// clashWins counts the EventClashWon events each player has.
func clashWins(g *game.Game) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, ev := range g.Events {
		if ev.Kind == game.EventClashWon {
			out[ev.Actor]++
		}
	}
	return out
}

// clashedLabels is the Label of each EventClashed, in order, per player.
func clashedLabels(g *game.Game, player uuid.UUID) []string {
	var out []string
	for _, ev := range g.Events {
		if ev.Kind == game.EventClashed && ev.Actor == player {
			out = append(out, ev.Label)
		}
	}
	return out
}

func clashLibraryTop(p *game.Player) uuid.UUID {
	return p.Library.Cards[p.Library.Size()-1].InstanceID
}

func clashLibraryBottom(p *game.Player) uuid.UUID { return p.Library.Cards[0].InstanceID }

// castHoardersGreed casts the sorcery and resolves it up to its first
// question.
func castHoardersGreed(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Hoarder's Greed", "Sorcery", hoardersGreedOracle, nil)
	passPriorityAroundTable(t, g)
}

// TestHoardersGreedRepeatsWhileYouWin — a won clash repeats the whole
// process, opponent choice included, and a lost one ends it. The decisions
// are asked in APNAP order, one at a time, and the cards move once both
// are in (CR 701.30c).
func TestHoardersGreedRepeatsWhileYouWin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[2]
	advanceToMain(t, g)
	// My library, top down: two draws, a 5, two draws, a 1.
	low := clashTop(me, "My One", "Sorcery", "{B}")
	clashTop(me, "Draw Four", "Sorcery", "{1}")
	clashTop(me, "Draw Three", "Sorcery", "{1}")
	high := clashTop(me, "My Five", "Sorcery", "{3}{B}{B}")
	clashTop(me, "Draw Two", "Sorcery", "{1}")
	clashTop(me, "Draw One", "Sorcery", "{1}")
	theirs := clashTop(opp, "Their Two", "Sorcery", "{1}{U}")
	life, hand := me.Life, me.Hand.Size()

	castHoardersGreed(t, g)
	if me.Life != life-2 || me.Hand.Size() != hand+2 {
		t.Fatalf("after the first pass: life %d hand %d, want %d and %d", me.Life, me.Hand.Size(), life-2, hand+2)
	}
	clashChooseOpponent(t, g, me.ID, opp.ID)

	// Both revealed; nothing has moved; the active player decides first.
	if clashLibraryTop(me) != high || clashLibraryTop(opp) != theirs {
		t.Fatal("a revealed card moved before the decisions")
	}
	clashDecide(t, g, me.ID, true)
	if clashLibraryTop(me) != high {
		t.Fatal("my card moved before the other clasher decided (CR 701.30c)")
	}
	second := clashDecide(t, g, opp.ID, false)
	if !strings.Contains(second.Reason, "My Five") || !strings.Contains(second.Reason, "bottom") {
		t.Errorf("the later decision was not told the earlier one: %q", second.Reason)
	}
	if clashLibraryBottom(me) != high {
		t.Error("my card was not put on the bottom of my library")
	}
	if clashLibraryTop(opp) != theirs {
		t.Error("their card did not stay on top")
	}
	if got := clashWins(g); got[me.ID] != 1 || got[opp.ID] != 0 {
		t.Fatalf("clash wins = %v, want one for me", got)
	}

	// I won: the process again. A new opponent is chosen each time.
	if me.Life != life-4 || me.Hand.Size() != hand+4 {
		t.Fatalf("after the repeat: life %d hand %d, want %d and %d", me.Life, me.Hand.Size(), life-4, hand+4)
	}
	clashChooseOpponent(t, g, me.ID, opp.ID)
	if clashLibraryTop(me) != low {
		t.Fatal("the second clash did not reveal my next card")
	}
	clashDecide(t, g, me.ID, false)
	clashDecide(t, g, opp.ID, true)
	if clashLibraryBottom(opp) != theirs {
		t.Error("their card was not put on the bottom of their library")
	}

	// 1 against 2: they won, and the process stops.
	if got := clashWins(g); got[me.ID] != 1 || got[opp.ID] != 1 {
		t.Errorf("clash wins = %v, want one each", got)
	}
	if len(clashPrompts(g)) != 0 {
		t.Fatal("the process went on after a lost clash")
	}
	if me.Life != life-4 || me.Hand.Size() != hand+4 {
		t.Errorf("after the lost clash: life %d hand %d, want %d and %d", me.Life, me.Hand.Size(), life-4, hand+4)
	}
	if got := clashedLabels(g, me.ID); len(got) != 2 || got[0] != game.ClashPutOnBottom || got[1] != game.ClashKeptOnTop {
		t.Errorf("my clashes = %v, want bottom then top", got)
	}
	if got := clashedLabels(g, opp.ID); len(got) != 2 || got[0] != game.ClashKeptOnTop || got[1] != game.ClashPutOnBottom {
		t.Errorf("their clashes = %v, want top then bottom", got)
	}
}

// TestHoardersGreedATieIsNotAWin — equal mana values: nobody wins
// (CR 701.30d), so the process runs once.
func TestHoardersGreedATieIsNotAWin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	clashTop(me, "My Three", "Sorcery", "{2}{B}")
	clashTop(me, "Draw Two", "Sorcery", "{1}")
	clashTop(me, "Draw One", "Sorcery", "{1}")
	clashTop(opp, "Their Three", "Creature — Bear", "{1}{G}{G}")
	life, hand := me.Life, me.Hand.Size()

	castHoardersGreed(t, g)
	clashChooseOpponent(t, g, me.ID, opp.ID)
	clashDecide(t, g, me.ID, false)
	clashDecide(t, g, opp.ID, false)

	if got := clashWins(g); len(got) != 0 {
		t.Errorf("clash wins = %v, want none on a tie", got)
	}
	if len(clashPrompts(g)) != 0 || me.Life != life-2 || me.Hand.Size() != hand+2 {
		t.Errorf("a tie repeated the process: life %d hand %d", me.Life, me.Hand.Size())
	}
}

// TestHoardersGreedWithNoCardToReveal — a player with an empty library
// reveals nothing, is asked nothing and cannot win; the other clasher's
// card is then higher than every other card revealed (CR 701.30d).
func TestHoardersGreedWithNoCardToReveal(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	clashTop(me, "Draw Two", "Sorcery", "{1}")
	clashTop(me, "Draw One", "Sorcery", "{1}")
	theirs := clashTop(opp, "Their Land", "Land", "")
	life := me.Life

	castHoardersGreed(t, g)
	clashChooseOpponent(t, g, me.ID, opp.ID)
	// Only the player who revealed a card has one to place.
	clashDecide(t, g, opp.ID, true)

	if clashLibraryBottom(opp) != theirs {
		t.Error("their card was not put on the bottom")
	}
	if got := clashWins(g); got[opp.ID] != 1 || got[me.ID] != 0 {
		t.Errorf("clash wins = %v, want one for the only player who revealed a card", got)
	}
	if got := clashedLabels(g, me.ID); len(got) != 1 || got[0] != "" {
		t.Errorf("my clash = %v, want one with no card", got)
	}
	if len(clashPrompts(g)) != 0 || me.Life != life-2 {
		t.Errorf("the process repeated after a clash I could not win: life %d", me.Life)
	}
}

// TestHoardersGreedOpenDecisionIsARestorePoint — a table waiting on a
// clasher's decision is plain data: restored, the answers still move the
// cards, settle the winner and repeat the process.
func TestHoardersGreedOpenDecisionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	clashTop(me, "Draw Four", "Sorcery", "{1}")
	clashTop(me, "Draw Three", "Sorcery", "{1}")
	high := clashTop(me, "My Five", "Sorcery", "{3}{B}{B}")
	clashTop(me, "Draw Two", "Sorcery", "{1}")
	clashTop(me, "Draw One", "Sorcery", "{1}")
	clashTop(opp, "Their Two", "Sorcery", "{1}{U}")
	life := me.Life

	castHoardersGreed(t, g)
	clashChooseOpponent(t, g, me.ID, opp.ID)
	clashDecide(t, g, me.ID, true)

	restored := restoreRoundTrip(t, g, true)
	rme := restored.Seats[me.Seat]
	clashDecide(t, restored, opp.ID, false)
	if clashLibraryBottom(rme) != high {
		t.Error("restored: my card was not put on the bottom")
	}
	if got := clashWins(restored); got[me.ID] != 1 {
		t.Errorf("restored: clash wins = %v, want one for me", got)
	}
	if rme.Life != life-4 {
		t.Errorf("restored: life %d, want %d (the process repeated)", rme.Life, life-4)
	}
	clashOnlyPrompt(t, restored, me.ID)
}

// marvoTable seats Marvo for the active player, ready to attack, with the
// hand emptied.
func marvoTable(t *testing.T) (g *game.Game, me, opp *game.Player, marvo uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	marvo = pushCatalogPermanent(g, me.ID, "Marvo, Deep Operative", "Legendary Creature — Octopus Rogue", marvoOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, opp, marvo
}

// marvoAttacks declares Marvo attacking `target` and resolves its attack
// trigger up to the first clash decision.
func marvoAttacks(t *testing.T, g *game.Game, marvo, target uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(marvo, target); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

// TestMarvoClashesWithDefendingPlayerAndWinning — the attack clashes with
// the player attacked, nobody is chosen, and winning draws a card and then
// offers a free cast of a spell with mana value 8 or less, the card just
// drawn included, in the declare attackers step.
func TestMarvoClashesWithDefendingPlayerAndWinning(t *testing.T) {
	g, me, opp, marvo := marvoTable(t)
	drawn := clashTop(me, "Drawn Sorcery", "Sorcery", "{4}{U}{U}")
	high := clashTop(me, "My Six", "Sorcery", "{5}{B}")
	clashTop(opp, "Their Two", "Sorcery", "{1}{U}")
	eight := handCardFull(me, "Eight Drop", "Sorcery", "{6}{U}{U}", "", []string{"U"})
	nine := handCardFull(me, "Nine Drop", "Sorcery", "{7}{U}{U}", "", []string{"U"})
	land := handCardFull(me, "Island", "Basic Land — Island", "", "", nil)

	marvoAttacks(t, g, marvo, opp.ID)
	// No "which opponent" question: the first thing asked is my decision.
	clashDecide(t, g, me.ID, true)
	clashDecide(t, g, opp.ID, false)
	if clashLibraryBottom(me) != high {
		t.Fatal("my card was not put on the bottom")
	}
	if got := clashWins(g); got[me.ID] != 1 {
		t.Fatalf("clash wins = %v, want one for me", got)
	}

	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(drawn) {
		t.Fatal("winning the clash did not draw a card")
	}
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, eight, drawn)
	if hasID(offer.DiscardOptions, nine) || hasID(offer.DiscardOptions, land) {
		t.Fatalf("the offer is %v, want no spell above mana value 8 and no land", offer.DiscardOptions)
	}
	// Control: the nine-drop is not castable free.
	if err := g.CastSpell(me.ID, nine, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a spell with mana value 9 was cast free")
	}
	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("step %s, want declare attackers", g.Turn.Step)
	}
	if err := g.CastSpell(me.ID, eight, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast of a sorcery with mana value 8: %v", err)
	}
	if !g.Stack.Contains(eight) {
		t.Fatal("the free spell is not on the stack")
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("the offer outlived its one cast")
	}
}

// TestMarvoLosingOrTyingTheClashDoesNothing — the second ability reads the
// winner: a clash the defending player won, and a tie, draw nothing.
func TestMarvoLosingOrTyingTheClashDoesNothing(t *testing.T) {
	for _, tc := range []struct {
		name, theirCost string
		theirWins       int
	}{
		{"lost", "{2}{U}{U}", 1},
		{"tied", "{1}{U}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, marvo := marvoTable(t)
			clashTop(me, "My Two", "Sorcery", "{1}{B}")
			clashTop(opp, "Theirs", "Sorcery", tc.theirCost)
			handCardFull(me, "Cheap Spell", "Sorcery", "{U}", "", []string{"U"})
			hand := me.Hand.Size()

			marvoAttacks(t, g, marvo, opp.ID)
			clashDecide(t, g, me.ID, false)
			clashDecide(t, g, opp.ID, false)
			passPriorityAroundTable(t, g)

			if got := clashWins(g); got[me.ID] != 0 || got[opp.ID] != tc.theirWins {
				t.Errorf("clash wins = %v", got)
			}
			if me.Hand.Size() != hand || resolutionCastOfferFor(g, me.ID) != nil {
				t.Errorf("a clash I did not win drew a card or offered a cast (hand %d, want %d)", me.Hand.Size(), hand)
			}
		})
	}
}

// TestMarvoTriggersOnAClashAnotherCardStarted — "whenever you win a
// clash" is any clash: Hoarder's Greed's, each time it is won.
func TestMarvoTriggersOnAClashAnotherCardStarted(t *testing.T) {
	g, me, opp, _ := marvoTable(t)
	clashTop(me, "Marvo Draw", "Sorcery", "{1}")
	clashTop(me, "My Zero", "Land", "")
	clashTop(me, "Draw Four", "Sorcery", "{1}")
	clashTop(me, "Draw Three", "Sorcery", "{1}")
	clashTop(me, "My Five", "Sorcery", "{3}{B}{B}")
	clashTop(me, "Draw Two", "Sorcery", "{1}")
	clashTop(me, "Draw One", "Sorcery", "{1}")
	clashTop(opp, "Their Two", "Sorcery", "{1}{U}")

	castHoardersGreed(t, g)
	clashChooseOpponent(t, g, me.ID, opp.ID)
	clashDecide(t, g, me.ID, true)
	clashDecide(t, g, opp.ID, false)
	// Won: the process repeats inside the same resolution, and loses.
	clashChooseOpponent(t, g, me.ID, opp.ID)
	clashDecide(t, g, me.ID, false)
	clashDecide(t, g, opp.ID, false)
	if got := clashWins(g); got[me.ID] != 1 || got[opp.ID] != 1 {
		t.Fatalf("clash wins = %v, want one each", got)
	}
	hand := me.Hand.Size()

	passPriorityAroundTable(t, g)
	// One win, one trigger, one card: the opponent's win is not mine.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (one Marvo trigger for my one win)", me.Hand.Size(), hand+1)
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast)
}

// TestMarvoAttackingAPlaneswalkerClashesWithItsController — the defending
// player of an attack on a planeswalker is its controller.
func TestMarvoAttackingAPlaneswalkerClashesWithItsController(t *testing.T) {
	g, me, opp, marvo := marvoTable(t)
	walker := pushCatalogWalkerWithLine(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "", 4)
	clashTop(me, "My Two", "Sorcery", "{1}{B}")
	theirs := clashTop(opp, "Their Four", "Sorcery", "{2}{U}{U}")

	marvoAttacks(t, g, marvo, walker)
	clashDecide(t, g, me.ID, false)
	clashDecide(t, g, opp.ID, true)
	if clashLibraryBottom(opp) != theirs {
		t.Error("the planeswalker's controller did not clash")
	}
	if got := clashWins(g); got[opp.ID] != 1 || got[me.ID] != 0 {
		t.Errorf("clash wins = %v, want one for the defending player", got)
	}
}
