package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crushing Canopy — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "618cd1bc-4422-441b-906f-1a209277be93", Name: "Crushing Canopy", Completeness: CompletenessFull,
		Modes: ChooseOne(Mode("Destroy target creature with flying.", TargetCreature("target creature with flying", HasKeyword("flying"))), Mode("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment()))),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return DestroyTarget{Target: id}.Apply(ctx)
		},
	})
}
