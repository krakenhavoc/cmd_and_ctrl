package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skarrgan Firebird — Creature — Phoenix {4}{R}{R}, 3/3 (Duel Decks: Heroes
// vs. Monsters, S82 old sets):
//
//	"Bloodthirst 3 (If an opponent was dealt damage this turn, this creature
//	 enters with three +1/+1 counters on it.)
//	 Flying
//	 {R}{R}{R}: Return this card from your graveyard to your hand. Activate
//	 only if an opponent was dealt damage this turn."
//
// Bloodthirst is Battering Wurm's. The graveyard ability is Abzan Devotee's
// return, with the activation instruction read as the ability is activated
// (CR 602.5b): any damage dealt to an opponent this turn, which is the same
// reading the bloodthirst uses. It does nothing if the card has left the
// graveyard by resolution (CR 608.2b).
//
// No simplification.
func init() {
	dealt := effects.AnOpponentWasDealtAtLeastThisTurn(1)
	effects.Register(effects.Spec{
		OracleID:        "42a96276-fddc-435e-bd47-04f9c6001ec7",
		Name:            "Skarrgan Firebird",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.Bloodthirst(3, "Skarrgan Firebird")},
		Activated: []effects.ActivatedAbility{{
			Label:     "{R}{R}{R}: Return this card from your graveyard to your hand. Activate only if an opponent was dealt damage this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{R}{R}{R}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool { return dealt(g, controller) },
			Effect:    effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
