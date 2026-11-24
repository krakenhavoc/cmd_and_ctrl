package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Break Ties — Instant {2}{W} (Modern Horizons 2, #5501):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Exile target card from a graveyard.
//	 Reinforce 1—{W} ({W}, Discard this card: Put a +1/+1 counter on target
//	 creature.)"
//
// Reinforce is Mosquito Guard's shared hand ability (MorReinforce).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e44bb4e-aeea-4f43-9cbd-4c24a9b6b022",
		Name:         "Break Ties",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Exile target card from a graveyard.",
				effects.TargetCardInGraveyard("target card from a graveyard"),
				effects.ExileTheModesTarget),
		),
		Activated: []effects.ActivatedAbility{
			effects.MorReinforce(1, "{W}", "Reinforce 1—{W} ({W}, Discard this card: Put a +1/+1 counter on target creature.)"),
		},
	})
}
