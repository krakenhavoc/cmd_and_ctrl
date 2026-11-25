package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scuttlemutt — {3} Artifact Creature — Scarecrow 2/2 (Shadowmoor, S82
// OLD b102, #6146):
//
//	"{T}: Add one mana of any color.
//	 {T}: Target creature becomes the color or colors of your choice until
//	 end of turn."
//
// The second ability asks, as it resolves, for a non-empty set of colours
// (#6146's multi colour pick, effects.ChooseColorsTail) and sets the target
// to exactly those colours in layer 5 until end of turn (CR 105.3, 613.1e):
// its old colours are gone and it returns to them in the cleanup step. A
// target that has left by then gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "564c785b-c2fb-4857-a5bf-a14ae2cf6c93",
		Name:         "Scuttlemutt",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature becomes the color or colors of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target creature", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				effects.ChooseColorsTail(game.ColorForBenefit, g, item.Controller, item.SourceCardID,
					"Scuttlemutt — choose one or more colors", ctx.ItemTail(scuttlemuttColorsChosen))
				return nil
			},
		}},
	})
}

// scuttlemuttColorsChosen sets the target to the chosen colours until end
// of turn. An on-disk key.
var scuttlemuttColorsChosen = effects.CardTail("scuttlemutt/colors-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	colors := game.ColorSetLetters(o.Tail.Word)
	if len(colors) == 0 {
		return nil
	}
	return effects.Gap2B05TargetBecomesColor(ctx, colors, "Scuttlemutt — "+game.ColorSetName(o.Tail.Word)+" until end of turn")
})
