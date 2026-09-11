package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Austere Command — Sorcery {4}{W}{W}:
//
//	"Choose two —
//	 • Destroy all artifacts.
//	 • Destroy all enchantments.
//	 • Destroy all creatures with mana value 3 or less.
//	 • Destroy all creatures with mana value 4 or greater."
//
// S20 sub-PR 4's "choose N" example: four untargeted options, Min =
// Max = 2. The modes resolve in printed order (CR 700.2c), and each
// sweep re-snapshots the battlefield, so a card destroyed by an
// earlier option is no longer there for a later one and an artifact
// creature is destroyed once rather than twice.
//
// The mana-value split is why this gets played over a flat wrath:
// "3 or less" plus "4 or greater" together is every creature, but
// either alone is a wipe that spares half a board — usually yours.
// Both halves come from the S20 predicate library, so the sweep's
// filter is the same object a targeting clause would use.
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
			sweeps := []CardPredicate{
				Artifact(),
				Enchantment(),
				And(Creature(), ManaValueLE(3)),
				And(Creature(), ManaValueGE(4)),
			}
			for i, match := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				if err := (DestroyAllMatching{Match: match}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
