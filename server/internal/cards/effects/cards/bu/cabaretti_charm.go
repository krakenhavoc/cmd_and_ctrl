package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabaretti Charm — Instant {R}{G}{W} (Streets of New Capenna):
//
//	"Choose one —
//	 • Cabaretti Charm deals damage equal to the number of creatures you
//	   control to target creature or planeswalker.
//	 • Creatures you control get +1/+1 and gain trample until end of turn.
//	 • Create two 1/1 green and white Citizen creature tokens."
//
// The damage counts the creatures you control as the spell resolves, so a
// creature that died in response is not counted (CR 608.2b), and the
// spell is the source. The pump locks the set of creatures as it resolves
// (CR 611.2c), as Lorehold Charm's does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36d8b701-f5dc-42c0-9869-268f8149cbd1",
		Name:         "Cabaretti Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Cabaretti Charm deals damage equal to the number of creatures you control to target creature or planeswalker.",
				effects.Harmful(effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					n := 0
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.Controller == item.Controller {
							n++
						}
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}),
			effects.ModeDoing("Creatures you control get +1/+1 and gain trample until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					mine := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Cabaretti Charm — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Cabaretti Charm — trample"}.Apply(ctx)
				}),
			effects.ModeDoing("Create two 1/1 green and white Citizen creature tokens.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 2}.Apply(ctx)
				}),
		),
	})
}
