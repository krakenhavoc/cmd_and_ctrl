package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "836ecd67-6487-4f3d-9e08-013746be829a",
		Name:         "Revival",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(Reanimates(0)),
		Targets:      TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", YouOwn(), And(Creature(), ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ralZarekReanimate(ctx.Game, item)
		},
	})
	Register(Spec{
		OracleID:     "836ecd67-6487-4f3d-9e08-013746be829a#1",
		Name:         "Revenge",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return doubleLifeTotal(ctx.Game, ctx.Source(), ctx.Controller())
			}),
			Instruction(func(ctx *Context) error {
				who, ok := firstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return msc18DoomsdayLoseHalfOfYourLife(ctx, who)
			}),
		),
	})
}
