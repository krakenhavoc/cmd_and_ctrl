package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dawn to Dusk — Sorcery {2}{W}{W} (Commander 2015, S82 OLD #5546):
//
//	"Choose one or both —
//	 • Return target enchantment card from your graveyard to your hand.
//	 • Destroy target enchantment."
//
// Markov Retribution's "one or both": each chosen bullet has its own target
// and resolves in printed order (CR 608.2c). The return takes the card the
// first bullet's target names if it is still in your graveyard (CR 608.2b);
// the destroy takes the enchantment the second names, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7a98058-b40d-4d74-9b97-6f04adf3c684",
		Name:         "Dawn to Dusk",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Return target enchantment card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.YouOwn(), effects.Enchantment()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
				})),
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
		),
	})
}
