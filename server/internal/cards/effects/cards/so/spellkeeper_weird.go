package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spellkeeper Weird — Creature — Weird {2}{U}, 1/4 (War of the Spark, #5520):
//
//	"{2}, {T}, Sacrifice this creature: Return target instant or sorcery card
//	 from your graveyard to your hand."
//
// Mudflat Village's sacrifice-and-return, with the target an instant or
// sorcery card in your graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91e09399-3f09-4323-8533-51536ecb7869",
		Name:         "Spellkeeper Weird",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice this creature: Return target instant or sorcery card from your graveyard to your hand.",
			Cost:  effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
			Effect: effects.ReturnTargetedCardToHand,
		}},
	})
}
