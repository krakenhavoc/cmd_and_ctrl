package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon the School — Kindred Sorcery — Merfolk {3}{W} (Lorwyn, S82 OLD,
// #5593):
//
//	"Create two 1/1 blue Merfolk Wizard creature tokens.
//	 Tap four untapped Merfolk you control: Return this card from your graveyard
//	 to your hand."
//
// Two Merfolk Wizards as the spell resolves, and Puppet Crafting's graveyard
// ability under Cloudgoat Ranger's tap-N cost over Merfolk. The cost has no {T}
// symbol, so a Merfolk that arrived this turn can pay it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00909d51-3da8-41eb-a535-0c325fedc59f",
		Name:         "Summon the School",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk Wizard"), N: 2}),
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap four untapped Merfolk you control: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  4,
				Filter: effects.TargetPermanent("four untapped Merfolk you control", effects.Creature(), effects.OfCreatureType("Merfolk")),
				Label:  "four untapped Merfolk you control",
			}},
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
