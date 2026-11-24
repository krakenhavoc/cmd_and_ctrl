package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breya's Apprentice — Artifact Creature — Human Artificer {2}{R}, 2/3
// (Modern Horizons 2):
//
//	"When this creature enters, create a 1/1 colorless Thopter artifact creature
//	 token with flying.
//	 {T}, Sacrifice an artifact: Choose one —
//	 • Exile the top card of your library. Until the end of your next turn, you
//	 may play that card.
//	 • Target creature gets +2/+0 until end of turn."
//
// A modal activated ability (ADR 0065): the mode and its target are announced
// with the activation; the tap and the sacrifice (the Apprentice may sacrifice
// itself, being an artifact) are paid up front. The exile mode is the shared
// "until the end of your next turn" grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fb9ff2d-1e90-4a26-8397-bb981fda2f7b",
		Name:         "Breya's Apprentice",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Breya's Apprentice — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Choose one — exile the top card of your library, you may play it until the end of your next turn; or target creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.TapCost(), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("an artifact", effects.Artifact())}),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Exile the top card of your library. Until the end of your next turn, you may play that card.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						controller := ctx.Controller()
						_, err := ctx.Game.ExileTopWithPermissionForEffect(controller, controller, 1, game.CastPermission{
							Duration: ctx.Game.UntilEndOfYourNextTurnDuration(controller),
						})
						return err
					}),
				effects.ModeDoing("Target creature gets +2/+0 until end of turn.",
					effects.TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						t, ok := effects.ModeTarget(ctx, occ)
						if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
							return nil
						}
						return effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Breya's Apprentice — +2/+0"}.Apply(ctx)
					}),
			),
		}},
	})
}
