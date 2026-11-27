package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crush Underfoot — Kindred Instant — Giant {1}{R} (Lorwyn):
//
//	"Choose a Giant creature you control. It deals damage equal to its power to
//	 target creature."
//
// The target is chosen as the spell is cast; the Giant is chosen as it
// resolves, by a pick and not a target (CR 608.2c), so hexproof and protection
// don't stop it. The chosen Giant, not the spell, is the source of the damage,
// its current power read when the damage is dealt (CR 608.2h: a Giant that has
// left deals it as it last existed). No Giant to choose, or no legal target,
// does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81686da0-85fc-4938-962b-24443de3ada1",
		Name:         "Crush Underfoot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			giant := ctx.TextSubtype("Giant") // a word of the spell's text (CR 612.2)
			return effects.ChoosePermanents{
				Question: "Crush Underfoot — choose a Giant creature you control",
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var out []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && c.IsCreature() && c.HasSubtype(giant) {
							out = append(out, c.InstanceID)
						}
					}
					if len(out) == 0 {
						return nil, 0, 0
					}
					return out, 1, 1
				},
				Keyed: crushUnderfootChosen,
			}.Apply(ctx)
		},
	})
}

// crushUnderfootChosen is the chosen Giant dealing damage equal to its power to
// the target, keyed (#4290).
var crushUnderfootChosen = effects.RegisterPermanentsPick("prompt-run/s82-old-b552-crush-underfoot", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		giants := picked.Cards()
		if len(giants) == 0 {
			return nil
		}
		target, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.DealDamage{Source: giants[0], Target: target, Amount: effects.B31CurrentPowerOf(ctx.Game, giants[0])}.Apply(ctx)
	},
})
