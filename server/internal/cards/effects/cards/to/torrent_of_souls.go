package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torrent of Souls — Sorcery {4}{B/R} (Eternal Masters, S82 OLD, #5542):
//
//	"Return up to one target creature card from your graveyard to the
//	 battlefield if {B} was spent to cast this spell. Creatures target player
//	 controls get +2/+0 and gain haste until end of turn if {R} was spent to
//	 cast this spell. (Do both if {B}{R} was spent.)"
//
// River's Grasp's two halves: the mana spent is read as the spell resolves
// (ReadsOwnPayment: a copy has none, so neither half happens). The black half
// reanimates the up-to-one target creature card; the red half then pumps the
// creatures the target player controls then (CR 611.2c), the returned one
// included, +2/+0 and haste until end of turn. The two targets are separate
// clauses, so one that is illegal does not stop the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "3ec1b309-8c26-4dc4-b4c3-f97c2c3f8a68",
		Name:           "Torrent of Souls",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   effects.CompletenessFull,
		WantsColors:    []string{"B", "R"},
		Targets: effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1).
			Then(effects.TargetPlayer("target player")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.ManaSpentOfColor("B") > 0 {
				if err := effects.ReturnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item); err != nil {
					return err
				}
			}
			if ctx.ManaSpentOfColor("R") == 0 {
				return nil
			}
			p, ok := ctx.ClauseTarget(1)
			if !ok || p.Kind != game.TargetPlayer {
				return nil
			}
			match := effects.And(effects.Creature(), effects.ControlledByPlayer(p.ID))
			if err := (effects.BoostUntilEOT{Match: match, Power: 2, Label: "Torrent of Souls — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"haste"}, Label: "Torrent of Souls — haste"}.Apply(ctx)
		},
	})
}
