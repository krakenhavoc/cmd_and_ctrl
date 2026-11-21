package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruin Rat — Creature — Rat {1}{B}, 1/1 (Hour of Devastation, #5534):
//
//	"Deathtouch
//	 When this creature dies, exile target card from an opponent's
//	 graveyard."
//
// Deathtouch is a printed keyword. The dies trigger is Leonin of the Lost
// Pride's: the target is a card in a graveyard its controller doesn't own,
// and the trigger is removed when there is none (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f77012cd-170f-40df-816b-330bf85e809d",
		Name:            "Ruin Rat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Ruin Rat — exile target card from an opponent's graveyard", effects.ExileTheFirstLegalTarget),
				effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.Not(effects.YouOwn())),
			),
		},
	})
}
