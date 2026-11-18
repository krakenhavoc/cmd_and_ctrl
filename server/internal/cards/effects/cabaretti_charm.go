package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "36d8b701-f5dc-42c0-9869-268f8149cbd1",
		Name:         "Cabaretti Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Cabaretti Charm deals damage equal to the number of creatures you control to target creature or planeswalker.",
				Harmful(TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					n := 0
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.Controller == item.Controller {
							n++
						}
					}
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}),
			ModeDoing("Creatures you control get +1/+1 and gain trample until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					mine := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Cabaretti Charm — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Cabaretti Charm — trample"}.Apply(ctx)
				}),
			ModeDoing("Create two 1/1 green and white Citizen creature tokens.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 2}.Apply(ctx)
				}),
		),
	})
}
