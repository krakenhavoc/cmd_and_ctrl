package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crushing Canopy — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "618cd1bc-4422-441b-906f-1a209277be93", Name: "Crushing Canopy", Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(effects.Mode("Destroy target creature with flying.", effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))), effects.Mode("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment()))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.DestroyTarget{Target: id}.Apply(ctx)
		},
	})
}
