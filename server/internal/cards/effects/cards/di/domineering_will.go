package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Domineering Will — Instant {3}{U} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP #5948):
//
//	"Target player gains control of up to three target nonattacking creatures
//	 until end of turn. Untap those creatures. They block this turn if able."
//
// Two clauses (Harmless Offering's shape): the player, then up to three
// creatures that are not attacking. Each creature still legal as the spell
// resolves changes hands until end of turn (CR 611.2a, layer-2 GainControl, a
// creature that left and came back is a new object and is not taken), is
// untapped, and must block if able (Culling Mark's requirement). With the
// player gone nothing is taken (CR 608.2b); creatures that became illegal are
// simply not taken.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c438ecb-b127-4450-9395-2d1eea294f30",
		Name:         "Domineering Will",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetCreature("up to three target nonattacking creatures", effects.Not(effects.AttackingCreature())).WithCount(0, 3),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := ctx.ClauseTarget(0)
			if !ok || player.Kind != game.TargetPlayer {
				return nil
			}
			for _, t := range ctx.ClauseTargets(1) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (effects.GainControl{Target: t.ID, Controller: player.ID, Duration: effects.DurationUntilEndOfTurn(ctx), Label: "Domineering Will — gain control until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.BlockRequirementUntilEOT{Target: t.ID, Kind: game.BlockRequirementBlocks, Label: "Domineering Will — blocks this turn if able"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
