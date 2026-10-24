package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geier Reach Bandit // Vildin-Pack Alpha — {2}{R} Creature — Human Rogue
// Werewolf 3/2 // Creature — Werewolf 4/3 (#4071, ADR 0132 amendment
// 2026-10-16):
//
//	Front: "Haste
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Whenever a Werewolf you control enters, you may transform it.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The Alpha's trigger fires for any Werewolf you control entering, and
// "it" is that permanent as it entered: one that left and came back is a
// new object (CR 400.7) and is left alone. The "you may" is asked as the
// trigger resolves (CR 603.5), and only when the transform would do
// something: a daybound or nightbound Werewolf can't be transformed by
// it (CR 702.145b, 702.145e), nor can one Immerwolf covers, nor a
// Werewolf that isn't a double-faced card (CR 701.27c). For those the
// trigger resolves doing nothing without a question whose "yes" would do
// nothing.
//
// No simplification.
func init() {
	const oracle = "64720167-b88b-47f7-bfc1-d5e005c9fd68"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Geier Reach Bandit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			TransformAtUpkeepIfNoSpellsLastTurn("Geier Reach Bandit"),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Vildin-Pack Alpha",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				// "Werewolf" is a word of the Alpha's text (CR 612.2, #3984).
				return ok && c.HasSubtype(source.TextSubtype("Werewolf"))
			}, "Vildin-Pack Alpha — you may transform that Werewolf", vildinPackAlphaTransform),
			TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Vildin-Pack Alpha"),
		},
	})
}

// vildinPackAlphaTransform asks "transform it?" of the Werewolf that
// entered, if it is still that object and can transform.
func vildinPackAlphaTransform(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if id == uuid.Nil || !g.CanTransformForEffect(id) {
		return nil
	}
	if o := item.Trigger.Object; o != nil && o.ID == id && g.ObjectEpochForEffect(id) != o.Epoch {
		return nil // CR 400.7: it left and came back
	}
	name := "that Werewolf"
	if c, ok := g.LookupCardForEffect(id); ok {
		name = c.Name
	}
	return MayChoice{
		Question: "Vildin-Pack Alpha — transform " + name + "?",
		Then:     geierReachBanditTransform,
		Carry:    []uuid.UUID{id},
	}.Apply(NewContext(g, item))
}

// geierReachBanditTransform is the "yes": transform the Werewolf (the
// carried ID). Keyed (#4291).
var geierReachBanditTransform = MayYes("confirm/geier-reach-bandit-transform", func(ctx *Context, r game.ConfirmAnswer) error {
	return Transform{Target: carryAt(r, 0)}.Apply(ctx)
})
