package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revival // Revenge — split card {W/B}{W/B} // {4}{W}{B} (Ravnica
// Remastered):
//
//	Revival (Sorcery): "Return target creature card with mana value 3 or less
//	 from your graveyard to the battlefield."
//	Revenge (Sorcery): "Double your life total. Target opponent loses half
//	 their life, rounded up."
//
// Revival is Ral Zarek, Guest Lecturer's reanimation. Revenge doubles the
// caster's life (Live Long and Prosper's doubleLifeTotal), then the opponent
// loses (life + 1) / 2 (Doomsday's half-life), each read as the clause
// happens (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "836ecd67-6487-4f3d-9e08-013746be829a",
		Name:         "Revival",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		Targets:      effects.TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", effects.YouOwn(), effects.And(effects.Creature(), effects.ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.RalZarekReanimate(ctx.Game, item)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "836ecd67-6487-4f3d-9e08-013746be829a#1",
		Name:         "Revenge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DoubleLifeTotal(ctx.Game, ctx.Source(), ctx.Controller())
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				who, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return effects.R35Msc18DoomsdayLoseHalfOfYourLife(ctx, who)
			}),
		),
	})
}
