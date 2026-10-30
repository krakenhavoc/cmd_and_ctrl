package effects

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mister Hyde, Monster Within — {2}{G} Legendary Creature — Human Villain
// 2/2 (Marvel Super Heroes, #4455):
//
//	"At the beginning of your upkeep, choose one —
//	 • Put a +1/+1 counter on Mister Hyde.
//	 • Remove a counter from a creature you control. If you do, draw a
//	   card."
//
// The mode is chosen as the trigger goes on the stack. The second mode
// targets nothing: as it resolves, the controller names the creature and the
// kind of counter in one question (every pair on a creature they control is
// an option, in battlefield and then name order), a lone pair is taken
// without asking, and with no counter anywhere nothing is removed and
// nothing is drawn ("if you do"). The draw follows the removal.
//
// No simplification.
func init() {
	hyde := AtYourUpkeep("Mister Hyde, Monster Within — choose one", func(_ *game.Game, _ *game.StackItem) error { return nil })
	hyde.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on Mister Hyde.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return plusOneCounterOnSourceIfSameObject(ctx)
			}),
		ModeDoing("Remove a counter from a creature you control. If you do, draw a card.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return mshB6RemoveCounterFromYourCreatureThenDraw(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "4a7c7ecc-df72-4da0-af35-dd0af58ec02c",
		Name:         "Mister Hyde, Monster Within",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{hyde},
	})
}

// mshB6HydeRemoveInfix joins the kind and the creature's name in an option
// label, which is what the answer reads the kind back from.
const (
	mshB6HydeRemovePrefix = "Remove a "
	mshB6HydeRemoveInfix  = " counter from "
)

// mshB6RemoveCounterFromYourCreatureThenDraw is the second mode: one option
// per (creature, counter kind) pair among the controller's creatures.
func mshB6RemoveCounterFromYourCreatureThenDraw(ctx *Context) error {
	var options []game.ChoiceOption
	var carry []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != ctx.Controller() || !c.IsCreature() {
			continue
		}
		var kinds []string
		for kind, n := range c.Counters {
			if n > 0 {
				kinds = append(kinds, kind)
			}
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			options = append(options, game.ChoiceOption{Label: mshB6HydeRemovePrefix + kind + mshB6HydeRemoveInfix + c.Name})
			carry = append(carry, c.InstanceID)
		}
	}
	switch len(options) {
	case 0:
		return nil
	case 1:
		return mshB6RemoveAndDraw(ctx, carry[0], options[0].Label)
	}
	return PickOption{
		Question: "Mister Hyde — remove which counter?",
		Options:  options,
		ThenKey:  mshB6HydeRemoveAnswer,
		Carry:    carry,
	}.Apply(ctx)
}

// mshB6HydeRemoveAnswer reads the chosen pair back: the creature is the
// carry at the chosen index, the kind sits in the label.
var mshB6HydeRemoveAnswer = OptionPickThen("option-pick/mister-hyde-remove-a-counter", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Index < 0 || r.Index >= len(r.Carry) {
		return nil
	}
	return mshB6RemoveAndDraw(ctx, r.Carry[r.Index], r.Option.Label)
})

// mshB6RemoveAndDraw removes one counter of the kind the label names from
// the creature, then draws a card if it was removed.
func mshB6RemoveAndDraw(ctx *Context, creature uuid.UUID, label string) error {
	rest, ok := strings.CutPrefix(label, mshB6HydeRemovePrefix)
	if !ok {
		return nil
	}
	kind, _, ok := strings.Cut(rest, mshB6HydeRemoveInfix)
	if !ok {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(creature)
	if !ok || !onBattlefield(ctx.Game, creature) || c.Counters[kind] <= 0 {
		return nil
	}
	if err := (AddCounter{Target: creature, Kind: kind, N: -1}).Apply(ctx); err != nil {
		return fmt.Errorf("mister hyde: %w", err)
	}
	return DrawCards{N: 1}.Apply(ctx)
}
