package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hell to Pay — Sorcery {X}{R} (EDHREC rank 2855):
//
//	"Hell to Pay deals X damage to target creature. Create a number
//	 of tapped Treasure tokens equal to the amount of excess damage
//	 dealt to that creature this way."
//
// Overkill that pays you back. X rides the cast; the damage goes
// through b27DealDamageWithExcess, which reads the creature's lethal
// requirement (toughness less damage already marked, CR 120.4a)
// before the damage lands and the amount actually dealt off the
// event afterwards, so a prevention shield lowers the excess along
// with the damage. The Treasures enter tapped
// (b13CreateTappedTreasures, Goldvein Hydra's payout) and are real
// Treasures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7dc7d90d-5979-4c4a-9183-392e0e0b878a",
		Name:         "Hell to Pay",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				excess, err := b27DealDamageWithExcess(ctx, t.ID, ctx.X())
				if err != nil {
					return err
				}
				return b13CreateTappedTreasures(ctx, item.Controller, excess)
			}
			return nil
		},
	})
}
