package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reconstruct History — Sorcery {2}{R}{W}:
//
//	"Return up to one target artifact card, up to one target enchantment
//	 card, up to one target instant card, up to one target sorcery card, and
//	 up to one target planeswalker card from your graveyard to your hand.
//	 Exile Reconstruct History."
//
// Retrieve's shape with five clauses: each "up to one" is announced
// separately and may be left empty, and each target is re-checked as the
// spell resolves (CR 608.2b). The same card may be chosen for two clauses
// when it fits both (CR 115.3), and is returned once. The spell then
// exiles itself instead of going to the graveyard.
func init() {
	Register(Spec{
		OracleID:     "97fad58a-8ee4-4f01-a47a-5ddc9ed845af",
		Name:         "Reconstruct History",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: Clauses(
			TargetCardInGraveyard("up to one target artifact card from your graveyard", YouOwn(), Artifact()).WithCount(0, 1),
			TargetCardInGraveyard("up to one target enchantment card from your graveyard", YouOwn(), Enchantment()).WithCount(0, 1),
			TargetCardInGraveyard("up to one target instant card from your graveyard", YouOwn(), Instant()).WithCount(0, 1),
			TargetCardInGraveyard("up to one target sorcery card from your graveyard", YouOwn(), Sorcery()).WithCount(0, 1),
			TargetCardInGraveyard("up to one target planeswalker card from your graveyard", YouOwn(), Planeswalker()).WithCount(0, 1),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for slot := 0; slot < 5; slot++ {
					for _, t := range ctx.ClauseTargets(slot) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if z := ctx.Game.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
							continue
						}
						if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				// "Exile Reconstruct History." The spell moves itself, so the
				// resolution frame does not put it into the graveyard (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
