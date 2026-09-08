package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Austere Command — Sorcery for {4}{W}{W}:
//
//	"Choose two —
//	 • Destroy all artifacts.
//	 • Destroy all enchantments.
//	 • Destroy all creatures with mana value 3 or less.
//	 • Destroy all creatures with mana value 4 or greater."
//
// S20 sub-PR 4's "choose N" example: four untargeted options, Min =
// Max = 2. The modes resolve in printed order (CR 700.2c), which
// matters here only for event ordering — a creature that's also an
// artifact is destroyed once, by whichever chosen option reaches it
// first.
func init() {
	Register(Spec{
		OracleID: "09cc8709-fe10-472a-b05c-e89f3523018d",
		Name:     "Austere Command",
		Modes: ChooseN("Choose two", 2, 2,
			Mode("Destroy all artifacts."),
			Mode("Destroy all enchantments."),
			Mode("Destroy all creatures with mana value 3 or less."),
			Mode("Destroy all creatures with mana value 4 or greater."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			preds := []CardPredicate{
				Artifact(),
				Enchantment(),
				And(Creature(), ManaValueLE(3)),
				And(Creature(), ManaValueGE(4)),
			}
			for i, pred := range preds {
				if !ctx.HasMode(i) {
					continue
				}
				// Re-snapshot per mode: a card destroyed by an earlier
				// option is no longer on the battlefield.
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if !pred(ctx.Game, ctx.Controller(), c) {
						continue
					}
					if err := (DestroyTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
