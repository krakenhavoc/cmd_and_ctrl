package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drown in Filth — Sorcery {B}{G} (Dragon's Maze, S82 OLD batch 404, #5564):
//
//	"Choose target creature. Mill four cards, then that creature gets -1/-1 until
//	 end of turn for each land card in your graveyard."
//
// Drown in Dreams' ordered steps: the mill is a replaceable event that can ask
// before it finishes, so the count waits for it (CR 608.2c). The land cards are
// counted in your graveyard once the mill has settled, the cards the mill just put
// there included, and the target is judged again for the second step: a target
// that became illegal counters the whole spell (CR 608.2b), so the mill never
// happens then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59af03ba-28d0-4f15-b877-2cf3c81d66e1",
		Name:         "Drown in Filth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.MillCards{N: 4}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				n := len(effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsLand() }))
				return effects.BoostUntilEOT{Target: id, Power: -n, Toughness: -n, Label: "Drown in Filth — -1/-1 for each land card in your graveyard"}.Apply(ctx)
			}),
		),
	})
}
