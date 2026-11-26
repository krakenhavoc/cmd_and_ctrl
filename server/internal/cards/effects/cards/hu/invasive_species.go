package cardshu

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasive Species — Creature — Insect {2}{G}, 3/3 (Magic 2015, #5556):
//
//	"When this creature enters, return another permanent you control to its
//	 owner's hand."
//
// Aviary Mechanic's resolution-time pick (nothing targets) with the "may"
// taken out: with another permanent under your control you must choose one,
// and with none nothing is returned. The Species itself is never offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c4c478c-6f0c-430a-91eb-4125a73d03a7",
		Name:         "Invasive Species",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasive Species — return another permanent you control to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					self := item.SourceCardID
					return effects.ChoosePermanents{
						Question: "Invasive Species — return another permanent you control to its owner's hand",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							var out []uuid.UUID
							for _, id := range effects.PermanentsControlledBy(g, of) {
								if id != self {
									out = append(out, id)
								}
							}
							if len(out) == 0 {
								return nil, 0, 0
							}
							return out, 1, 1
						},
						Keyed: invasiveSpeciesReturn,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// invasiveSpeciesReturn returns the permanent the controller picked. An
// on-disk key: never renamed, never reused.
var invasiveSpeciesReturn = effects.RegisterPermanentsPick("prompt-run/invasive-species-return", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			return effects.BounceToHand{Target: id}.Apply(ctx)
		}
		return nil
	},
})
