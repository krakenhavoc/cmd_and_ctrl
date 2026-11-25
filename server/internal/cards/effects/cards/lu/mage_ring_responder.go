package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mage-Ring Responder — Artifact Creature — Golem {7}, 7/7 (Magic
// Origins):
//
//	"This creature doesn't untap during your untap step.
//	 {7}: Untap this creature.
//	 Whenever this creature attacks, it deals 7 damage to target creature
//	 defending player controls."
//
// Basalt Monolith's untap restriction and mana untap, with Warkite
// Marauder's defending-player target clause: the defending player is
// read off this attack, so in a multiplayer game only the creatures of
// the player it attacks can be chosen (a planeswalker's controller
// likewise). The Responder deals the damage itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "1d84fb81-2de5-4be1-9638-9ae28e3d867a",
		Name:                  "Mage-Ring Responder",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			UntapSelf: true,
			Label:     "{7}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{7}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.WheneverThisAttacks("Mage-Ring Responder — it deals 7 damage to target creature defending player controls",
					effects.SourceDealsDamageToEachLegalTarget(7)),
				effects.TargetCreatureDefendingPlayerControls),
		},
	})
}
