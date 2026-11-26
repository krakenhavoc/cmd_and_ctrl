package cardss

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Scribe of the Mindful — {2}{U} Creature — Human Cleric 2/2 (Amonkhet, S82
// old batch 248):
//
//	"{1}, {T}, Sacrifice this creature: Return target instant or sorcery card
//	 from your graveyard to your hand."
//
// The target is chosen as the ability is put on the stack and re-checked as it
// resolves (CR 608.2b): a card that has left the graveyard is not returned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "019a3229-dd1c-4f70-a8fb-e33dbc54df39",
		Name:         "Scribe of the Mindful",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this creature: Return target instant or sorcery card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			Effect:  effects.ReturnTargetedCardToHand,
		}},
	})
}
