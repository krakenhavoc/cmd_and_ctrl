package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fates' Reversal — Sorcery {1}{B}:
//
//	"Return up to one target creature card from your graveyard to your
//	 hand. Venture into the dungeon."
//
// With no target chosen the spell still resolves and ventures; with a
// target that has left the graveyard it does not resolve (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8cd88e87-6585-437a-bcad-b494fffc0544",
		Name:         "Fates' Reversal",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return VentureIntoTheDungeon{}.Apply(ctx)
		},
	})
}
