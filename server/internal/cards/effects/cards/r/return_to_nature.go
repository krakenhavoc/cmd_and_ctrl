package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return to Nature — Instant {1}{G} (EDHREC rank 759):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Exile target card from a graveyard."
//
// Naturalize with a graveyard mode. Three targeted options on a
// "choose one" — legal, since the per-mode target limit only bites
// when more than one option may be chosen (Rakdos Charm's shape). The
// graveyard mode reads ANY graveyard, the caster's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "777b8ec4-a783-4297-96b7-4f200d0eb734",
		Name:         "Return to Nature",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact())),
			effects.Mode("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment())),
			effects.Mode("Exile target card from a graveyard.", effects.TargetCardInGraveyard("target card in a graveyard")),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0), ctx.HasMode(1):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			case ctx.HasMode(2):
				return effects.ExileTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
