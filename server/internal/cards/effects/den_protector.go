package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Den Protector — Creature — Human Warrior {1}{G}, 2/1 (Murders at
// Karlov Manor Commander):
//
//	"Creatures with power less than this creature's power can't block it.
//	 Megamorph {1}{G}
//	 When this creature is turned face up, return target card from your
//	 graveyard to your hand."
//
// The evasion is Howlgeist's block rule, read live, so the megamorph
// counter (which arrives before the turn-face-up event) widens it. The
// graveyard card is any card, and the target is re-judged as the trigger
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	up := WhenThisIsTurnedFaceUp("Den Protector — return target card from your graveyard to your hand",
		returnTargetGraveyardCardToHand)
	up.Targets = TargetCardInGraveyard("target card from your graveyard", YouOwn())
	Register(Spec{
		OracleID:         "59324efc-65b7-4b31-89c5-17aed5c7eeb1",
		Name:             "Den Protector",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Megamorph("{1}{G}")},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(PowerLessThanSource(), OnSelf(),
				"creatures with power less than Den Protector's can't block it"),
		},
		Triggered: []game.TriggeredAbility{up},
	})
}
