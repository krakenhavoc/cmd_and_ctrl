package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b95_helpers.go — card-side helpers of set GAP, batch 95
// (Doctor Who instants, lands and sorceries, S82, #5948).

// gapB95IsTimeLord is "a Time Lord": the engine reads a type line word by
// word, so the creature type Time Lord is the two words Time and Lord.
func gapB95IsTimeLord(c game.Card) bool {
	return c.HasSubtypeWord("Time") && c.HasSubtypeWord("Lord")
}

// gapB95YouControlATimeLord is the activation condition "Activate only if you
// control a Time Lord".
func gapB95YouControlATimeLord(g *game.Game, controller, _ uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && gapB95IsTimeLord(c) {
			return true
		}
	}
	return false
}

// gapB95DestroyItsControllerLosesLife is "Destroy target creature. Its
// controller loses N life.": the creature's controller is read before the
// destroy, so the life loss still lands when the creature is indestructible,
// and the spell does nothing when its target is gone (CR 608.2b).
func gapB95DestroyItsControllerLosesLife(n int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		for _, t := range ctx.LegalTargets() {
			c, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				continue
			}
			controller := c.Controller
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, -n)
		}
		return nil
	}
}
