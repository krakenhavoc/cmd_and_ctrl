package game

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// color_choice_test.go pins #742: the "choose a color" prompt in both
// of its forms, the stored colour's lifecycle, and "N mana of any one
// color" as one pick minting N tokens.

func choiceByKind(g *Game, kind PendingChoiceKind) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	return nil
}

func TestColorChoiceStoredFormStampsChosenColor(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	heart := pushTypedTestCard(g, Card{Name: "Coldsteel Heart", TypeLine: "Snow Artifact", Owner: seat, Controller: seat})
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.QueueColorChoiceForEffect(seat, heart, "Coldsteel Heart", nil) })

	c := choiceByKind(g, PendingChoiceColor)
	if c == nil || !reflect.DeepEqual(c.ColorOptions, AllColors) {
		t.Fatalf("queued choice = %+v, want all five colours on offer", c)
	}
	// Lower case is accepted and normalised.
	if err := g.ResolveColorChoice(id, seat, "g"); err != nil {
		t.Fatalf("ResolveColorChoice: %v", err)
	}
	if got := layeredBattlefieldCard(t, g, heart).ChosenColor; got != "G" {
		t.Errorf("ChosenColor = %q, want G", got)
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("pending choices after the answer = %d, want 0", n)
	}
	var sawEvent bool
	for _, ev := range g.Events {
		if ev.Kind == EventColorChosen && ev.CardID == heart && ev.Label == "G" {
			sawEvent = true
		}
	}
	if !sawEvent {
		t.Error("no color_chosen event naming the card and the colour")
	}
}

func TestColorChoiceRefusesAnswersOffTheList(t *testing.T) {
	g := newActiveGame(t)
	seat, other := g.Seats[0].ID, g.Seats[1].ID
	isle := pushTypedTestCard(g, Card{Name: "Thriving Isle", TypeLine: "Land", Owner: seat, Controller: seat})
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.QueueColorChoiceForEffect(seat, isle, "Thriving Isle", ColorsOtherThan("U")) })

	for _, bad := range []string{"U", "C", "", "blue", "WU"} {
		if err := g.ResolveColorChoice(id, seat, bad); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("answer %q: err = %v, want ErrInvalidParam", bad, err)
		}
	}
	if err := g.ResolveColorChoice(id, other, "R"); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another seat answering: err = %v, want ErrNotTheChooser", err)
	}
	if choiceByKind(g, PendingChoiceColor) == nil {
		t.Fatal("a refused answer dequeued the prompt")
	}
	if err := g.ResolveColorChoice(id, seat, "R"); err != nil {
		t.Fatalf("a legal answer after refusals: %v", err)
	}
	if got := layeredBattlefieldCard(t, g, isle).ChosenColor; got != "R" {
		t.Errorf("ChosenColor = %q, want R", got)
	}
}

// Colorless is not a colour (CR 105.4), so a caller cannot put it on
// offer by accident.
func TestColorChoiceOptionsAreRealColorsOnly(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	g.WithWriteLock(func() { g.QueueColorChoiceForEffect(seat, uuid.New(), "x", []string{"g", "C", "Z", "W", "G"}) })
	if got := choiceByKind(g, PendingChoiceColor).ColorOptions; !reflect.DeepEqual(got, []string{"W", "G"}) {
		t.Errorf("options = %v, want [W G]", got)
	}
}

func TestColorChoiceResolutionFormRunsThenAndStoresNothing(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	src := pushTypedTestCard(g, Card{Name: "Oona", TypeLine: "Legendary Creature — Faerie Wizard", Owner: seat, Controller: seat})
	got := ""
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueColorChoiceThenForEffect(ColorPrompt{
			Chooser: seat, Source: src, Question: "Oona — choose a color",
			Then: func(_ *Game, color string) error { got = color; return nil },
		})
	})
	// While open it holds a continuation, so the game is not a
	// restore point.
	if snap := g.CaptureSnapshot(); snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 1 {
		t.Errorf("census = %+v, want the chooseColorResume frame counted", snap.Continuations)
	}
	if err := g.ResolveColorChoice(id, seat, "B"); err != nil {
		t.Fatalf("ResolveColorChoice: %v", err)
	}
	if got != "B" {
		t.Errorf("continuation received %q, want B", got)
	}
	if c := layeredBattlefieldCard(t, g, src).ChosenColor; c != "" {
		t.Errorf("the resolution form stored %q on the source; it stores nothing", c)
	}
}

// The continuation may queue the next prompt — Selective
// Obliteration's "each player chooses a color" is a chain.
func TestColorChoiceResolutionFormChains(t *testing.T) {
	g := newActiveGame(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	var picks []string
	var first uuid.UUID
	g.WithWriteLock(func() {
		first = g.QueueColorChoiceThenForEffect(ColorPrompt{
			Chooser: a,
			Then: func(g *Game, color string) error {
				picks = append(picks, color)
				g.QueueColorChoiceThenForEffect(ColorPrompt{
					Chooser: b,
					Then:    func(_ *Game, color string) error { picks = append(picks, color); return nil },
				})
				return nil
			},
		})
	})
	if err := g.ResolveColorChoice(first, a, "W"); err != nil {
		t.Fatalf("first answer: %v", err)
	}
	next := choiceByKind(g, PendingChoiceColor)
	if next == nil || next.Chooser != b {
		t.Fatalf("the chain did not ask the second player: %+v", next)
	}
	if err := g.ResolveColorChoice(next.ID, b, "U"); err != nil {
		t.Fatalf("second answer: %v", err)
	}
	if !reflect.DeepEqual(picks, []string{"W", "U"}) {
		t.Errorf("picks = %v, want [W U]", picks)
	}
}

func TestChosenColorClearsOnBattlefieldLeave(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	id := pushTypedTestCard(g, Card{Name: "Coldsteel Heart", TypeLine: "Artifact", Owner: seat, Controller: seat})
	g.WithWriteLock(func() { g.Battlefield.Cards[cardIndex(t, g, id)].ChosenColor = "U" })
	var moved Card
	g.WithWriteLock(func() {
		var err error
		moved, err = MoveCard(g.Battlefield, g.Seats[0].Graveyard, id)
		if err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	if moved.ChosenColor != "" {
		t.Errorf("ChosenColor after leaving the battlefield = %q, want empty", moved.ChosenColor)
	}
}

func TestChosenColorAndManaAmountsSurviveSnapshotAndClone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushTypedTestCard(g, Card{Name: "Coldsteel Heart", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		g.Battlefield.Cards[cardIndex(t, g, id)].ChosenColor = "R"
		g.QueueChoiceForEffect(PendingChoice{
			Kind: PendingChoiceMana, Chooser: me.ID, FromPlayer: me.ID, Count: 1,
			ColorOptions: []string{"W", "U"}, ManaAmounts: map[string]int{"W": 3, "U": 2},
		})
	})

	_, restored := roundTrip(t, g)
	var card Card
	for _, c := range restored.Battlefield.Cards {
		if c.InstanceID == id {
			card = c
		}
	}
	if card.ChosenColor != "R" {
		t.Errorf("restored ChosenColor = %q, want R", card.ChosenColor)
	}
	if c := choiceByKind(restored, PendingChoiceMana); c == nil || !reflect.DeepEqual(c.ManaAmounts, map[string]int{"W": 3, "U": 2}) {
		t.Errorf("restored mana pick = %+v, want its amounts", c)
	}

	clone := g.Clone()
	choiceByKind(clone, PendingChoiceMana).ManaAmounts["W"] = 99
	if choiceByKind(g, PendingChoiceMana).ManaAmounts["W"] != 3 {
		t.Error("the clone shares the live game's ManaAmounts map")
	}
}

func TestParseProducedManaAmounts(t *testing.T) {
	cases := []struct {
		in   string
		want []ProducedManaEntry
	}{
		{"{W3|U3}", []ProducedManaEntry{{Options: []string{"W", "U"}, Amounts: map[string]int{"W": 3, "U": 3}}}},
		{"{g2|u5}", []ProducedManaEntry{{Options: []string{"G", "U"}, Amounts: map[string]int{"G": 2, "U": 5}}}},
		// Single option: expanded into ordinary slots.
		{"{G3}", []ProducedManaEntry{{Options: []string{"G"}}, {Options: []string{"G"}}, {Options: []string{"G"}}}},
		// Zero drops the option.
		{"{G0|U2}", []ProducedManaEntry{{Options: []string{"U"}}, {Options: []string{"U"}}}},
		{"{G0|U0}", nil},
		// Counted but all ones is an ordinary pick.
		{"{W1|U1}", []ProducedManaEntry{{Options: []string{"W", "U"}}}},
		// Unchanged forms.
		{"{W|U}", []ProducedManaEntry{{Options: []string{"W", "U"}}}},
		{"{C}{C}", []ProducedManaEntry{{Options: []string{"C"}}, {Options: []string{"C"}}}},
	}
	for _, tc := range cases {
		got, err := ParseProducedMana(tc.in)
		if err != nil {
			t.Errorf("%s: %v", tc.in, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s = %+v, want %+v", tc.in, got, tc.want)
		}
	}
	for _, bad := range []string{"{3}", "{W3x}", "{|W}", "{W|}"} {
		if _, err := ParseProducedMana(bad); err == nil {
			t.Errorf("%s parsed, want an error", bad)
		}
	}
}

func oneColorShape(produced string) []ManaAbilityShape {
	return []ManaAbilityShape{{TapCost: true, Produced: produced, Label: "Add three mana of any one color", IgnoreCommanderIdentity: true}}
}

// Gilded Lotus: one pick, three tokens of the picked colour.
func TestOneColorAmountIsOnePickMintingN(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lotus := pushIntrinsicPermanent(g, me, "Gilded Lotus", "Artifact", oneColorShape("{W3|U3|B3|R3|G3}"), nil)

	if err := g.ActivateManaAbility(me.ID, lotus, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMana {
			picks++
		}
	}
	if picks != 1 {
		t.Fatalf("queued %d mana picks, want exactly one", picks)
	}
	c := choiceByKind(g, PendingChoiceMana)
	if err := g.ResolveManaChoice(c.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(me.ManaPool) != 3 {
		t.Fatalf("pool = %v, want three tokens", me.ManaPool)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "U" {
			t.Errorf("token %v, want {U}: every token is the one picked colour", tok)
		}
	}
}

// The planner's model is one slot, one mana, one colour per slot; a
// one-colour-N-mana source would be booked for three different
// colours at once. It plans around the source instead.
func TestAutoTapperPlansAroundOneColorAmountSources(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushIntrinsicPermanent(g, me, "Gilded Lotus", "Artifact", oneColorShape("{W3|U3|B3|R3|G3}"), nil)
	cost, _ := ParseCost("{1}")
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Errorf("the auto-tapper planned %v off a one-colour-amount source", plan)
	}
}

func TestAddManaForEffectOneColorAmount(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(me.ID, uuid.New(), "{W2|U2|B2|R2|G2}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	})
	c := choiceByKind(g, PendingChoiceMana)
	if c == nil || c.ManaAmounts["G"] != 2 {
		t.Fatalf("queued pick = %+v, want amounts carried", c)
	}
	if err := g.ResolveManaChoice(c.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want two {G}", me.ManaPool)
	}
}
