package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pawpatch Formation — Instant {1}{G} (EDHREC rank 4420):
//
//	"Choose one —
//	 • Destroy target creature with flying.
//	 • Destroy target enchantment.
//	 • Draw a card. Create a Food token."
//
// Bloomburrow's green modal answer, and the reason it plays in
// Commander rather than only in limited is the third mode: a removal
// spell that is never dead. Green has no shortage of enchantment
// removal and no way at all to kill a flier, so the first two modes
// cover green's two real holes and the third covers the games where
// neither hole is open. Roadmap batch 42 (#449), "no new machinery".
//
// Three modes, two of them targeted — legal on a "choose one", where
// the per-mode target limit only bites when the maximum is above one.
// The mode picker runs before targeting, so the legal set the client
// offers is the set the chosen bullet names: fliers for mode 0,
// enchantments for mode 1, nothing at all for mode 2.
//
// "Creature with flying" is the layered read, so a creature that was
// granted flying is a legal target and a flier that lost it is not —
// checked at announce (CR 601.2c) and again at resolution (CR
// 608.2b), which is what lets an opponent's instant-speed answer save
// the bird.
//
// No simplification. The Food is the real token with its real
// "{2}, {T}, Sacrifice this token: You gain 3 life" ability.
func init() {
	Register(Spec{
		OracleID:     "fa99cf82-ebf1-4526-ab9c-b24eb2970f2a",
		Name:         "Pawpatch Formation",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy target creature with flying.",
				TargetCreature("target creature with flying", HasKeyword("flying"))),
			Mode("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment())),
			Mode("Draw a card. Create a Food token."),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0), ctx.HasMode(1):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
			case ctx.HasMode(2):
				if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
					return err
				}
				return CreateToken{Controller: ctx.Controller(), Template: FoodToken(), N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
