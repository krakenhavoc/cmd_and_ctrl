package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turf War — Enchantment {4}{R}:
//
//	"When this enchantment enters, for each player, put a contested
//	 counter on target land that player controls.
//	 Whenever a creature deals combat damage to a player, if that player
//	 controls one or more lands with contested counters on them, that
//	 creature's controller gains control of one of those lands of their
//	 choice and untaps it."
//
// The first ability has one target clause per seat, counted from Turf
// War's controller (the table has at most four). The second is the
// intervening-"if" trigger of CR 603.4, checked as the damage is dealt and
// again as the ability resolves. The creature's controller, not Turf War's,
// chooses the land and gains control of it indefinitely (CR 611.2a), so the
// land stays stolen if Turf War leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e1b84d7-b0ae-4e9d-bc1a-c84832fea9f9",
		Name:         "Turf War",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Turf War — for each player, put a contested counter on target land that player controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: effects.S82SNC22ContestedKind, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}), effects.S82SNC22SeatLands()),
			effects.On(game.EventDealDamage, effects.S82SNC22TurfWarHit,
				"Turf War — the creature's controller gains control of one of those lands and untaps it",
				effects.S82SNC22TurfWarTake),
		},
	})
}
