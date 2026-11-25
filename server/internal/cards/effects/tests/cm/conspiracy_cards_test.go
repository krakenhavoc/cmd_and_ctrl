package testscm

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Conspiracies in the command zone and hidden agenda (#6102, ADR 0227,
// CR 315, CR 702.106), proved with four cards: Weight Advantage (face up),
// Immediate Action, Brago's Favor and Iterative Analysis (hidden agenda).

const (
	consWeightAdvantageOracle   = "42bede0a-de6d-4de3-90ea-9d92b0fe8c64"
	consImmediateActionOracle   = "5a9c1c2d-d4d7-4df3-a853-f1b81c67b3e3"
	consBragosFavorOracle       = "395b312b-8fe6-4c28-b91d-620b5fcb9e6d"
	consIterativeAnalysisOracle = "79172853-240c-4924-8b20-edf469946df4"
)

func consCard(name, oracle string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Conspiracy", OracleID: oracle, IsSideboard: true}
}

// consTable starts a four-seat game whose seat 0 has `side` in its
// sideboard. The deal may be waiting on a hidden agenda.
func consTable(t *testing.T, side ...game.Card) (*game.Game, *game.Player) {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if i == 0 {
			deck = append(deck, side...)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g, g.Seats[0]
}

// consAgendaPrompts is the hidden agenda prompts open for chooser.
func consAgendaPrompts(g *game.Game, chooser uuid.UUID) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName && c.Chooser == chooser {
			out = append(out, c)
		}
	}
	return out
}

// consName answers the first open agenda prompt of chooser with name.
func consName(t *testing.T, g *game.Game, chooser uuid.UUID, name string) {
	t.Helper()
	open := consAgendaPrompts(g, chooser)
	if len(open) == 0 {
		t.Fatal("no hidden agenda prompt is open")
	}
	if err := g.ResolveCardNameChoice(open[0].ID, chooser, name); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
}

// consBegin keeps every hand and walks to seat 0's precombat main phase.
func consBegin(t *testing.T, g *game.Game) {
	t.Helper()
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
}

func consConspiracy(t *testing.T, g *game.Game, p *game.Player, name string) game.Card {
	t.Helper()
	for _, c := range g.ConspiraciesOf(p.ID) {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("%s is not in the command zone", name)
	return game.Card{}
}

func consReveal(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionRevealAgenda, game.SpecialActionParams{}); err != nil {
		t.Fatalf("reveal_agenda: %v", err)
	}
}

func TestConspiracyHiddenAgendaIsNamedBeforeTheDeal(t *testing.T) {
	g, me := consTable(t,
		consCard("Weight Advantage", consWeightAdvantageOracle),
		consCard("Immediate Action", consImmediateActionOracle))
	if len(me.Hand.Cards) != 0 || g.PendingDealSeat == nil {
		t.Fatalf("hands were dealt before the hidden agenda was named (hand %d)", len(me.Hand.Cards))
	}
	if err := g.KeepHand(me.ID); !errors.Is(err, game.ErrDealPending) {
		t.Fatalf("KeepHand before the deal: %v, want ErrDealPending", err)
	}
	if n := len(consAgendaPrompts(g, me.ID)); n != 1 {
		t.Fatalf("%d agenda prompts, want 1 (Weight Advantage has no hidden agenda)", n)
	}
	for _, c := range me.Sideboard.Cards {
		if game.IsConspiracy(c) {
			t.Fatalf("%s is still in the sideboard", c.Name)
		}
	}
	// CR 315.5b / 315.7: face down, its owner's alone; the face-up one is
	// everyone's.
	ia := consConspiracy(t, g, me, "Immediate Action")
	wa := consConspiracy(t, g, me, "Weight Advantage")
	other := g.Seats[1].ID
	if !ia.FaceDown || ia.FaceDownKind != game.FaceDownConspiracy || ia.IsKnownTo(other) || !ia.IsKnownTo(me.ID) {
		t.Fatalf("Immediate Action: face down %v (%q), known to another seat %v", ia.FaceDown, ia.FaceDownKind, ia.IsKnownTo(other))
	}
	if wa.FaceDown || !wa.IsKnownTo(other) {
		t.Fatal("Weight Advantage is not face up and public")
	}
	if game.CatalogKey(ia) != "" {
		t.Fatal("a face-down conspiracy has a catalog key: it has characteristics (CR 315.5b)")
	}
	consName(t, g, me.ID, "Goblin Guide")
	if g.PendingDealSeat != nil || len(me.Hand.Cards) != game.OpeningHandSize {
		t.Fatalf("the deal did not go on after the name (hand %d)", len(me.Hand.Cards))
	}
	if ia = consConspiracy(t, g, me, "Immediate Action"); ia.ChosenName != "Goblin Guide" {
		t.Fatalf("chosen name %q", ia.ChosenName)
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventCardNameChosen {
			t.Fatal("the secret name reached the public event log")
		}
	}
	// Not an emblem.
	if len(g.EmblemsForPlayer(me.ID)) != 0 {
		t.Fatal("a conspiracy is listed as an emblem")
	}
}

func TestImmediateActionGrantsHasteOnlyFaceUp(t *testing.T) {
	g, me := consTable(t, consCard("Immediate Action", consImmediateActionOracle))
	consName(t, g, me.ID, "Goblin Guide")
	consBegin(t, g)
	guide := effectstest.B12Creature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, guide, "haste") {
		t.Fatal("a face-down hidden agenda granted haste (CR 315.5)")
	}
	consReveal(t, g, me, consConspiracy(t, g, me, "Immediate Action").InstanceID)
	ia := consConspiracy(t, g, me, "Immediate Action")
	if ia.FaceDown || !ia.IsKnownTo(g.Seats[1].ID) {
		t.Fatal("the revealed conspiracy is not face up and public")
	}
	if !effectstest.HasEffectiveKeyword(t, g, guide, "haste") {
		t.Error("a creature you control with the chosen name lacks haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, other, "haste") {
		t.Error("a creature with another name has haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature with the chosen name has haste")
	}
	// Turning it up is once: there is nothing left to reveal.
	if err := g.PerformSpecialAction(me.ID, ia.InstanceID, game.SpecialActionRevealAgenda, game.SpecialActionParams{}); err == nil {
		t.Error("a face-up conspiracy was revealed again")
	}
}

func TestBragosFavorReducesOnlyFaceUpAndOnlyTheChosenName(t *testing.T) {
	g, me := consTable(t, consCard("Brago's Favor", consBragosFavorOracle))
	consName(t, g, me.ID, "Grizzly Bears")
	consBegin(t, g)
	early := effectstest.HandSpell(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, early, game.CastSpellParams{Strict: true}), "a face-down Brago's Favor")
	consReveal(t, g, me, consConspiracy(t, g, me, "Brago's Favor").InstanceID)
	other := effectstest.HandSpell(me, "Runeclaw Bear", "Creature — Bear", "{1}{G}")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true}), "a spell with another name")
	if err := g.CastSpell(me.ID, early, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Grizzly Bears for {G} under a face-up Brago's Favor: %v", err)
	}
}

func TestIterativeAnalysisDrawsOnTheChosenInstantFaceUp(t *testing.T) {
	g, me := consTable(t, consCard("Iterative Analysis", consIterativeAnalysisOracle))
	consName(t, g, me.ID, "Opt")
	consBegin(t, g)
	cast := func() int {
		id := effectstest.HandSpell(me, "Opt", "Instant", "{U}")
		effectstest.FillPoolColored(me, "U", 1)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast Opt: %v", err)
		}
		return len(me.Hand.Cards)
	}
	before := cast()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || len(me.Hand.Cards) != before {
		t.Fatal("a face-down Iterative Analysis triggered")
	}
	consReveal(t, g, me, consConspiracy(t, g, me, "Iterative Analysis").InstanceID)
	before = cast()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Fatalf("hand %d after the trigger, want %d", got, before+1)
	}
}

func TestWeightAdvantageStartsFaceUpAndWorks(t *testing.T) {
	g, me := consTable(t, consCard("Weight Advantage", consWeightAdvantageOracle))
	if g.PendingDealSeat != nil || len(me.Hand.Cards) != game.OpeningHandSize {
		t.Fatal("a conspiracy without hidden agenda held up the deal")
	}
	consBegin(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 6)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Wall", "Creature — Wall", 0, 6)
	if !effectstest.LayeredCard(t, g, mine).Effective().DamageByToughness {
		t.Error("a creature you control does not assign combat damage by toughness")
	}
	if effectstest.LayeredCard(t, g, theirs).Effective().DamageByToughness {
		t.Error("an opponent's creature assigns combat damage by toughness")
	}
}

func TestHiddenAgendaRevealedAtGameEnd(t *testing.T) {
	g, me := consTable(t, consCard("Immediate Action", consImmediateActionOracle))
	consName(t, g, me.ID, "Goblin Guide")
	consBegin(t, g)
	g.End()
	ia := consConspiracy(t, g, me, "Immediate Action")
	for _, p := range g.Seats {
		if !ia.IsKnownTo(p.ID) {
			t.Fatal("a face-down conspiracy was not revealed as the game ended (CR 702.106e)")
		}
	}
}

// consSeatView is seat 0's projection as viewer sees it.
func consSeatView(t *testing.T, g *game.Game, viewer uuid.UUID) protocol.PlayerView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, s := range v.Seats {
		if s.ID == g.Seats[0].ID.String() {
			return s
		}
	}
	t.Fatal("seat 0 is not in the view")
	return protocol.PlayerView{}
}

func TestHiddenAgendaIsRedactedOnTheWireUntilRevealed(t *testing.T) {
	g, me := consTable(t, consCard("Immediate Action", consImmediateActionOracle))
	other := g.Seats[1].ID
	// The prompt says nothing about the card to another seat.
	for _, c := range protocol.ViewOfGameFor(g, other.String()).PendingChoices {
		if c.Kind == string(game.PendingChoiceCardName) && (c.Source != "" || c.Reason != "Secretly choosing a hidden agenda name") {
			t.Fatalf("another seat sees the agenda prompt's source %q and reason %q", c.Source, c.Reason)
		}
	}
	consName(t, g, me.ID, "Goblin Guide")
	consBegin(t, g)
	mine := consSeatView(t, g, me.ID).Conspiracies
	if len(mine) != 1 || mine[0].Name != "Immediate Action" || mine[0].ChosenName != "Goblin Guide" {
		t.Fatalf("the owner's view: %+v", mine)
	}
	if len(mine[0].SpecialActions) == 0 || mine[0].SpecialActions[0].Kind != string(game.SpecialActionRevealAgenda) {
		t.Fatalf("the owner has no reveal row: %+v", mine[0].SpecialActions)
	}
	theirs := consSeatView(t, g, other).Conspiracies
	if len(theirs) != 1 || theirs[0].Name != "" || theirs[0].ChosenName != "" || !theirs[0].FaceDown || len(theirs[0].SpecialActions) != 0 {
		t.Fatalf("another seat sees a face-down conspiracy: %+v", theirs)
	}
	consReveal(t, g, me, consConspiracy(t, g, me, "Immediate Action").InstanceID)
	theirs = consSeatView(t, g, other).Conspiracies
	if len(theirs) != 1 || theirs[0].Name != "Immediate Action" || theirs[0].ChosenName != "Goblin Guide" || theirs[0].FaceDown {
		t.Fatalf("another seat after the reveal: %+v", theirs)
	}
}

func TestHiddenAgendaIsRevealedOnlyWithPriority(t *testing.T) {
	g, me := consTable(t, consCard("Immediate Action", consImmediateActionOracle))
	consName(t, g, me.ID, "Goblin Guide")
	id := consConspiracy(t, g, me, "Immediate Action").InstanceID
	// The mulligan window: nobody has priority.
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionRevealAgenda, game.SpecialActionParams{}); !errors.Is(err, game.ErrSpecialActionTiming) {
		t.Fatalf("reveal before the game began: %v, want ErrSpecialActionTiming", err)
	}
	consBegin(t, g)
	// Another seat cannot turn it up (CR 315.6: it is its owner's).
	if err := g.PerformSpecialAction(g.Seats[1].ID, id, game.SpecialActionRevealAgenda, game.SpecialActionParams{}); err == nil {
		t.Fatal("another seat revealed my hidden agenda")
	}
	consReveal(t, g, me, id)
}
