package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Co-Option Adjutant — Artifact Creature — Borg {3}{U}, 3/3 (Star Trek
// Commander, #4974):
//
//	"Artifact creatures you control have "Whenever this creature deals
//	 combat damage to a player, draw a card.""
//
// Hive Mind Coprocessor's layer-6 grant of a trigger bundle (ADR 0093)
// without the payment: the trigger is each recipient's own, so "this
// creature" is the artifact creature that connected and the card goes to
// its controller. The Adjutant is an artifact creature itself and has it
// too; two Adjutants grant two instances, which trigger separately
// (CR 603.2).
//
// No simplification.
const coOptionAdjutantGrant = "co-option-adjutant/combat-damage-draw"

func init() {
	const text = "Whenever this creature deals combat damage to a player, draw a card."
	effects.Register(effects.Spec{
		OracleID:     "caafecc0-adfa-478c-8a9d-234d4069300b",
		Name:         "Co-Option Adjutant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Grants: []effects.AbilityGrant{{
			Key: coOptionAdjutantGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Co-Option Adjutant — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.Labelled(
			effects.GrantAbilities(effects.TrekB34ArtifactCreaturesYouControl, coOptionAdjutantGrant),
			"Artifact creatures you control have \""+text+"\"")},
	})
}
