package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Proctor of Potential — Creature — Human Cleric {W}{U}, 3/1:
//
//	"Whenever this creature or another creature you control enters,
//	 surveil 1.
//	 {W}{U}: Return this card from your graveyard to the battlefield
//	 with a finality counter on it. Activate only if you've scried or
//	 surveilled this turn. (If a creature with a finality counter on it
//	 would die, exile it instead.)"
//
// Grim Repriser's graveyard activation with a different gate: this
// turn's scry or surveil (rfCreatureBScriedOrSurveilledThisTurn). The
// trigger is "a creature you control enters", which includes the
// Proctor itself (Impact Tremors' reading), so it surveils as it comes
// back. The finality counter is half the card's and half the engine's:
// the Proctor enters with one only when it comes from a graveyard (a
// self-replacement), and a permanent carrying one is exiled instead of
// dying by the engine's CR 122.1h replacement (game/finality_counters.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c612ba0a-5134-45fa-a762-f4621bda4f83",
		Name:         "Proctor of Potential",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.RfCreatureBEntersWithCounterFromGraveyard(effects.RfCreatureBFinalityCounter, 1,
				"Proctor of Potential: enters with a finality counter"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Proctor of Potential — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}: Return this card from your graveyard to the battlefield with a finality counter on it. Activate only if you've scried or surveilled this turn.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{W}{U}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.RfCreatureBScriedOrSurveilledThisTurn(g, controller)
			},
			Effect: effects.ReturnThisFromGraveyardToBattlefield,
		}},
	})
}
