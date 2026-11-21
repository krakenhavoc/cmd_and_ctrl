package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	up := effects.WhenThisIsTurnedFaceUp("Den Protector — return target card from your graveyard to your hand",
		effects.ReturnTargetGraveyardCardToHand)
	up.Targets = effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn())
	effects.Register(effects.Spec{
		OracleID:         "59324efc-65b7-4b31-89c5-17aed5c7eeb1",
		Name:             "Den Protector",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{1}{G}")},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.PowerLessThanSource(), effects.OnSelf(),
				"creatures with power less than Den Protector's can't block it"),
		},
		Triggered: []game.TriggeredAbility{up},
	})
}
