package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lethal Throwdown — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 sacrifice a modified creature. (Equipment, Auras you control, and
//	 counters are modifications.) Destroy target creature or
//	 planeswalker. If the modified creature was sacrificed, draw a card."
//
// The either/or cost (ADR 0100 §2) whose two branches are both
// sacrifices, and the one card in the either/or set whose resolution
// reads WHICH branch was paid: ctx.PaidCostBranch("modified"). The
// draw follows the announcement, not the board — a modified creature
// sacrificed under the "a creature" branch was not "the modified
// creature", because that branch names no such creature.
//
// "Modified" is CR 700.9's word, read by the catalog's one predicate for
// it (b07IsModified: a counter on it, an Equipment attached, or an Aura
// its controller controls).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ab565bb-3188-4d23-93d8-5de67fc0d056",
		Name:         "Lethal Throwdown",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a creature", Creature()).Keyed("creature"),
			SacrificeCost("a modified creature", Creature(), modifiedCreature).Keyed("modified"),
		),
		Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			if !ctx.PaidCostBranch("modified") {
				return nil
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}

// modifiedCreature is b07IsModified as a card predicate.
func modifiedCreature(g *game.Game, _ uuid.UUID, c game.Card) bool {
	return b07IsModified(g, c)
}
