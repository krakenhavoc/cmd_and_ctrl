package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devastate — Sorcery {3}{R}{R} (Prophecy, S82 OLD b672, #5617):
//
//	"Destroy target land. Devastate deals 1 damage to each creature and each
//	 player."
//
// Two instructions in order (CR 608.2c): the destruction of the land, then
// Pyrohemia's one damage to each creature and each player, the spell as the
// source. The one target is the land's: if it is illegal on resolution the spell
// doesn't resolve and deals no damage (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "90b3e2df-182d-4dad-9934-597a048fefc6",
		Name:         "Devastate",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.B23DamageEachCreatureAndEachPlayer(ctx, 1)
		},
	})
}
