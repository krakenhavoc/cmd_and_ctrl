package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigilant Sentry — Creature — Human Nomad {1}{W}{W}, 2/2 (Dominaria
// Remastered, #5948):
//
//	"Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +1/+1 and has "{T}: Target attacking or
//	 blocking creature gets +3/+3 until end of turn.""
//
// Two statics read live: a layer 7c +1/+1, and a layer 6 grant of the
// tap ability to itself (a bundle, ADR 0093), both only while the
// controller's graveyard holds seven or more cards. The granted ability is
// Angelic Page's shape with +3/+3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9bb66d53-a83c-4313-91c3-4df09c248558",
		Name:         "Vigilant Sentry",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: vigilantSentryGrantKey,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target attacking or blocking creature gets +3/+3 until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)), game.AnswerPump),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
				Effect:  effects.BoostTargetUntilEOT("Vigilant Sentry — +3/+3", 3, 3),
			}},
			Text: "{T}: Target attacking or blocking creature gets +3/+3 until end of turn.",
		}},
		Static: []game.StaticAbility{
			effects.FraThresholdSelfPT(1, 1),
			effects.Labelled(effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.FraThresholdMet(g, source.Controller)
			}, vigilantSentryGrantKey), `Threshold — This creature has "{T}: Target attacking or blocking creature gets +3/+3 until end of turn." while there are seven or more cards in your graveyard.`),
		},
	})
}

// vigilantSentryGrantKey names the granted ability's bundle: an on-disk
// identity.
const vigilantSentryGrantKey = "vigilant-sentry/tap-pump"
