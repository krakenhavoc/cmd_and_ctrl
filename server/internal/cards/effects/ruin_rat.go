package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f77012cd-170f-40df-816b-330bf85e809d",
		Name:            "Ruin Rat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Ruin Rat — exile target card from an opponent's graveyard", exileTheFirstLegalTarget),
				TargetCardInGraveyard("target card from an opponent's graveyard", Not(YouOwn())),
			),
		},
	})
}
