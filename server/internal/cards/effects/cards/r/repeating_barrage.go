package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repeating Barrage — Sorcery {1}{R}{R} (Ixalan):
//
//	"Repeating Barrage deals 3 damage to any target.
//	 Raid — {3}{R}{R}: Return this card from your graveyard to your hand.
//	 Activate only if you attacked this turn."
//
// Brackish Trudge's graveyard ability (an activated ability of the card in
// the graveyard, CR 602.2) gated on the turn's attack record as Bloodsoaked
// Champion's raid is; the card is checked to still be in the graveyard as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a08c599e-ee75-4d2b-ae9b-0987c233dd17",
		Name:         "Repeating Barrage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve:    effects.DamageToFirstTargetStep(3),
		Activated: []effects.ActivatedAbility{{
			Label:   "Raid — {3}{R}{R}: Return this card from your graveyard to your hand. Activate only if you attacked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{R}{R}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.B18AttackedThisTurn(g, controller)
			},
			Effect: effects.ReturnThisFromYourGraveyardToHand,
		}},
	})
}
