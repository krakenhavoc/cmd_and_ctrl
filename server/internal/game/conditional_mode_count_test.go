package game

import (
	"testing"

	"github.com/google/uuid"
)

// conditional_mode_count_test.go — #1590, ADR 0065's 2026-09-27
// amendment, at the engine level: the bound a conditional mode count
// yields for each chooser, and the two non-spell owners of a ModeSpec
// (a trigger's mode_pick prompt, CR 603.3c) reading the same bound the
// cast gate does. The spell path and the cards are
// internal/cards/effects/conditional_mode_count_test.go.

// controlsAnArtifact is a stand-in condition — the engine never knows
// what the card's clause says, only whether it holds for the chooser.
// Registered once, at package init, as a card file would.
var controlsAnArtifact = ModeCondition("test/controls-an-artifact", func(g *Game, chooser uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == chooser && c.IsArtifact() {
			return true
		}
	}
	return false
})

func TestModeMaxIsRaisedOnlyForAChooserTheConditionHoldsFor(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ms := (&ModeSpec{Options: make([]ModeOption, 3), Min: 1, Max: 1}).OrUpToIf(2, controlsAnArtifact)

	if got := modeHi(g, ms, me.ID); got != 1 {
		t.Errorf("condition false: max = %d, want the printed 1", got)
	}
	g.Battlefield.PushTop(Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if got := modeHi(g, ms, me.ID); got != 2 {
		t.Errorf("condition true: max = %d, want 2", got)
	}
	if got := modeHi(g, ms, opp.ID); got != 1 {
		t.Errorf("the condition is the CHOOSER's: opponent's max = %d, want 1", got)
	}
	if err := validateModes(ms, ms.Min, modeHi(g, ms, me.ID), []int{0, 2}); err != nil {
		t.Errorf("two within the raised bound: %v", err)
	}
	if err := validateModes(ms, ms.Min, modeHi(g, ms, opp.ID), []int{0, 2}); err != ErrInvalidParam {
		t.Errorf("two over the printed bound: %v, want ErrInvalidParam", err)
	}
	if err := validateModes(ms, ms.Min, modeHi(g, ms, me.ID), []int{0, 1, 2}); err != ErrInvalidParam {
		t.Errorf("three over the raised bound: %v, want ErrInvalidParam", err)
	}
	if modeHi(g, nil, me.ID) != 0 {
		t.Error("a nil spec has no bound")
	}
	// A key this binary never registered holds for nobody: the printed
	// bound, the weaker reading.
	unknown := (&ModeSpec{Options: make([]ModeOption, 3), Min: 1, Max: 1}).OrUpToIf(2, ModeCountCondition{key: "test/never-registered"})
	if got := modeHi(g, unknown, me.ID); got != 1 {
		t.Errorf("unknown condition: max = %d, want 1", got)
	}
}

func TestModeConditionRefusesABadRegistration(t *testing.T) {
	for name, reg := range map[string]func(){
		"empty key": func() { ModeCondition("", func(*Game, uuid.UUID) bool { return true }) },
		"nil func":  func() { ModeCondition("test/nil-func", nil) },
		"duplicate": func() { ModeCondition("test/controls-an-artifact", func(*Game, uuid.UUID) bool { return true }) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("ModeCondition must panic")
				}
			}()
			reg()
		})
	}
}

// A modal ACTIVATED ability reads the raised bound at activation (CR
// 602.2b). No printed card has this shape yet; a ModeSpec has three
// owners (ADR 0065 §3) and each must honour it.
func TestModalActivatedAbilityHonoursTheRaisedBound(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ms := (&ModeSpec{
		Prompt: "Choose one", Min: 1, Max: 1,
		Options: []ModeOption{{Label: "One."}, {Label: "Two."}},
	}).OrUpToIf(2, controlsAnArtifact)
	exhaustCatalog(t, "test-conditional-modal-activated", ActivatedAbilityShape{
		Label:  "{1}: Choose one.",
		Cost:   AbilityCost{Mana: "{1}"},
		Modes:  ms,
		Effect: func(*Game, *StackItem) error { return nil },
	})
	// An enchantment source, so the source itself is not the artifact
	// the condition looks for.
	c := NewCard("Modal Probe", me.ID)
	c.TypeLine = "Enchantment"
	c.OracleID = "test-conditional-modal-activated"
	c.Controller = me.ID
	g.Battlefield.PushTop(c)
	activate := func() error {
		me.ManaPool.AddMana(ManaToken{Color: "C"})
		err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{Modes: []int{0, 1}})
		me.ManaPool = nil
		return err
	}
	if err := activate(); err != ErrInvalidParam {
		t.Fatalf("no artifact: both modes must be refused, got %v", err)
	}
	g.Battlefield.PushTop(Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := activate(); err != nil {
		t.Fatalf("with an artifact: both modes are legal, got %v", err)
	}
}

// A modal TRIGGER reads the raised bound as it goes on the stack (CR
// 603.3c): the prompt carries it, and the answer gate enforces it.
func TestModalTriggerPromptCarriesTheRaisedBound(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	ability := threeBulletTrigger(&ran)
	ability.Modes.OrUpToIf(2, controlsAnArtifact)
	src := modalTriggerSource(t, g, me, "test-conditional-modal-trigger", ability)

	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })
	c := openModePick(t, g, me.ID)
	if c == nil || c.ModeMax != 1 {
		t.Fatalf("no artifact: the prompt's max is the printed 1: %+v", c)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1}); err != ErrInvalidParam {
		t.Errorf("two bullets over the printed bound: %v, want ErrInvalidParam", err)
	}

	g.Battlefield.PushTop(Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: src, Actor: me.ID}) })
	c = openModePick(t, g, me.ID)
	if c == nil || c.ModeMax != 2 {
		t.Fatalf("with an artifact: the prompt's max is the raised 2: %+v", c)
	}
}
