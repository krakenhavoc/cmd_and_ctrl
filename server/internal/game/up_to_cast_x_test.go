package game

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// up_to_cast_x_test.go — #4617 at the engine level: "choose up to X"
// where X is the X paid for the spell that became the source (CR
// 107.3m), and the empty answer to any "choose up to" trigger (CR
// 603.3c, CR 700.2b). The card is
// internal/cards/effects/the_ruinous_wrecking_crew_test.go.

// upToXTrigger is threeBulletTrigger counted "choose up to X".
func upToXTrigger(ran *[]string) TriggeredAbility {
	t := threeBulletTrigger(ran)
	t.Modes.Prompt = "Choose up to X"
	t.Modes.Min, t.Modes.Max, t.Modes.UpToCastX = 0, len(t.Modes.Options), true
	return t
}

// upToXSource puts a creature that entered from a spell cast for `x`
// onto the battlefield and fires its enters trigger.
func upToXSource(t *testing.T, g *Game, owner *Player, x int, ability TriggeredAbility) uuid.UUID {
	t.Helper()
	oracle := "test-up-to-x-" + uuid.NewString()
	id := modalTriggerSource(t, g, owner, oracle, ability)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Provenance = CastProvenance{X: x, Caster: owner.ID}
		}
	}
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventETB, CardID: id, Actor: owner.ID}) })
	return id
}

func TestUpToCastXBoundIsTheXCappedAtTheBullets(t *testing.T) {
	g := newActiveGame(t)
	ms := &ModeSpec{Options: make([]ModeOption, 4), Min: 0, Max: 4, UpToCastX: true}
	for x, want := range map[int]int{0: 0, 1: 1, 2: 2, 4: 4, 6: 4, -1: 0} {
		lo, hi := g.modeBoundsLocked(ms, ModeCountQuery{X: x})
		if lo != 0 || hi != want {
			t.Errorf("X = %d: bounds %d..%d, want 0..%d", x, lo, hi, want)
		}
	}
}

func TestUpToCastXAtZeroAsksNothingAndPutsNothingOnTheStack(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	src := upToXSource(t, g, me, 0, upToXTrigger(&ran))
	if c := openModePick(t, g, me.ID); c != nil {
		t.Fatalf("X = 0: no prompt, got %+v", c.ModeOptionIndex)
	}
	if it := triggerItemFrom(g, src); it != nil {
		t.Fatalf("X = 0: the ability is removed (CR 700.2b), got an item with modes %v", it.Modes)
	}
}

func TestUpToCastXPromptCarriesTheX(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	upToXSource(t, g, me, 2, upToXTrigger(&ran))
	c := openModePick(t, g, me.ID)
	if c == nil {
		t.Fatal("X = 2: no prompt")
	}
	// No creature on the battlefield but the source itself, which the
	// third bullet may target, so all three are offered.
	if c.ModeMin != 0 || c.ModeMax != 2 {
		t.Errorf("bounds %d..%d, want 0..2", c.ModeMin, c.ModeMax)
	}
	if !strings.Contains(c.Reason, "Choose up to X (X is 2)") {
		t.Errorf("reason %q names the X", c.Reason)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1, 2}); err != ErrInvalidParam {
		t.Errorf("three at X = 2: %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 0}); err != ErrInvalidParam {
		t.Errorf("one bullet twice: %v, want ErrInvalidParam (CR 700.2d)", err)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("one at X = 2: %v", err)
	}
}

// The prompt never says more than the bullets on offer: X = 9 with a
// bullet whose target can't be chosen is up to two.
func TestUpToCastXPromptCapsAtTheBulletsOnOffer(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var ran []string
	ability := upToXTrigger(&ran)
	// Only noncreature targets for the third bullet: none on the board.
	ability.Modes.Options[2].Targets.CardOK = func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return !c.IsCreature() && !c.IsLand() }
	upToXSource(t, g, me, 9, ability)
	c := openModePick(t, g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if len(c.ModeOptionIndex) != 2 || c.ModeMax != 2 {
		t.Errorf("offered %v with max %d, want two and 2", c.ModeOptionIndex, c.ModeMax)
	}
}

// An empty answer to a "choose up to" trigger is final: the ability is
// removed (CR 603.3c, CR 700.2b) and the question is not asked again.
// Before #4617 the empty answer read as "not asked yet" and re-queued
// the prompt, for Ertai Resurrected's "choose up to one" too.
func TestChoosingNoModeRemovesTheTrigger(t *testing.T) {
	for name, ability := range map[string]func(*[]string) TriggeredAbility{
		"up to X": upToXTrigger,
		"up to one": func(ran *[]string) TriggeredAbility {
			t := threeBulletTrigger(ran)
			t.Modes.Min = 0
			return t
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			var ran []string
			src := upToXSource(t, g, me, 2, ability(&ran))
			c := openModePick(t, g, me.ID)
			if c == nil || c.ModeMin != 0 {
				t.Fatalf("a prompt that allows none: %+v", c)
			}
			if err := g.ResolveModePick(c.ID, me.ID, []int{}); err != nil {
				t.Fatalf("no mode: %v", err)
			}
			if again := openModePick(t, g, me.ID); again != nil {
				t.Fatal("the empty answer asked the question again")
			}
			if it := triggerItemFrom(g, src); it != nil {
				t.Fatalf("no mode chosen: the ability is removed, got modes %v", it.Modes)
			}
			if len(ran) != 0 {
				t.Errorf("nothing runs: %v", ran)
			}
		})
	}
}
