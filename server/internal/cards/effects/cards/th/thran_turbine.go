package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thran Turbine — Artifact {1} (Urza's Saga, S82 OLD batch 732, #5624):
//
//	"At the beginning of your upkeep, you may add {C}{C}. This mana can't be
//	 spent to cast spells."
//
// The "may" is asked as the trigger resolves (CR 603.5). The two {C} carry
// game.ManaRestrictNotCast: they pay for an activated ability (and the other
// payments that are not casts) and for no spell. The pool empties as the
// upkeep ends (CR 106.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c31f06d-bf02-48a7-9c17-be2a226b8eae",
		Name:         "Thran Turbine",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Thran Turbine — you may add {C}{C}",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Thran Turbine — add {C}{C} (can't be spent to cast spells)?",
						Then:     thranTurbineAdd,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// thranTurbineAdd is the "yes": {C}{C} that can't be spent to cast spells. An
// on-disk key.
var thranTurbineAdd = effects.MayYes("confirm/thran-turbine-add-cc", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.AddMana{Produced: "{C}{C}", Restrictions: []string{game.ManaRestrictNotCast}}.Apply(ctx)
})
