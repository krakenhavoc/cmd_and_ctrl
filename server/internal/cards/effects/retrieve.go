package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Retrieve — Sorcery {2}{G} (Innistrad: Crimson Vow):
//
//	"Return up to one target creature card and up to one target noncreature
//	 permanent card from your graveyard to your hand. Exile Retrieve."
//
// Rite of Renewal's shape: two optional target clauses in printed order (slot
// 0 a creature card, slot 1 a noncreature permanent card, both from your
// graveyard), each checked again on resolution (CR 608.2b), so a card that
// has left the graveyard stays where it went. The last sentence exiles the
// spell itself instead of letting it go to the graveyard; a spell countered
// or fizzled for illegal targets is not exiled, because it never resolved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "157a4cba-b2e6-4c42-b7dc-62db19625c9c",
		Name:         "Retrieve",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: Clauses(
			TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1),
			TargetCardInGraveyard("up to one target noncreature permanent card from your graveyard", YouOwn(), Permanent(), Not(Creature())).WithCount(0, 1),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, slot := range []int{0, 1} {
					for _, t := range ctx.ClauseTargets(slot) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
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
				// "Exile Retrieve." The spell moves itself, so the resolution
				// frame does not put it into the graveyard (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
