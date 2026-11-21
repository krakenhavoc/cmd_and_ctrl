package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memorial to Folly — Land (EDHREC rank 2913):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {2}{B}, {T}, Sacrifice this land: Return target creature card
//	 from your graveyard to your hand."
//
// The black Memorial: a tapped Swamp that turns into a Raise Dead
// late. Enters tapped is the self-replacement; the mana is a plain
// {B}; the activation is mana, tap and sacrifice-this with a
// graveyard target clause (targetCreatureInYourGraveyard — the zone
// browser picker), returning the pick to hand if it is still there
// at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2bc38f14-0314-4351-8138-e2b8bf041404",
		Name:         "Memorial to Folly",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}, Sacrifice this land: Return target creature card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreatureInYourGraveyard(),
			Effect:  effects.B27ReturnChosenGraveyardCardToHand,
		}},
	})
}
