package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nemesis Phoenix — Creature — Phoenix {3}{R}{R}, 4/2 (Commander Legends: Battle for
// Baldur's Gate):
//
//	"Flying
//	 {2}{R}: Return this card from your graveyard to the battlefield tapped and
//	 attacking. Activate only during the declare attackers step and only if you're
//	 attacking two or more opponents."
//
// Flying is printed. The ability works from the graveyard (CR 113.6) and is gated by
// two conditions asked as it is activated: the declare attackers step, and two or more
// distinct opponents being attacked by the activator's creatures (CR 506.3d). It puts
// this card onto the battlefield tapped and attacking, the controller choosing what it
// attacks (CR 508.4); it was never declared as an attacker, so "whenever attacks"
// triggers do not see it (CR 506.3c). The card must still be in the graveyard as the
// ability resolves (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bd0dcb8c-d550-4034-8597-0aba6f6e3b3f",
		Name:            "Nemesis Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: Return this card from your graveyard to the battlefield tapped and attacking. Activate only during the declare attackers step and only if you're attacking two or more opponents.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{R}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.AllConditions(
				effects.DuringStep(game.StepDeclareAttackers),
				effects.GapB220YoureAttackingTwoOrMoreOpponents,
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return effects.PutAttackingYourChoice(effects.NewContext(g, item), item.SourceCardID, game.ZoneGraveyard, "Nemesis Phoenix")
			},
		}},
	})
}
