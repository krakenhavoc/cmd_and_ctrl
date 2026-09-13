package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "777b8ec4-a783-4297-96b7-4f200d0eb734",
		Name:         "Return to Nature",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment())),
			Mode("Exile target card from a graveyard.", TargetCardInGraveyard("target card in a graveyard")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0), ctx.HasMode(1):
				return DestroyTarget{Target: target}.Apply(ctx)
			case ctx.HasMode(2):
				return ExileTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
