package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slimefoot, the Stowaway — Legendary Creature — Fungus {1}{B}{G}, 2/3
// (Commander Masters):
//
//	"Whenever a Saproling you control dies, Slimefoot deals 1 damage to
//	 each opponent and you gain 1 life.
//	 {4}: Create a 1/1 green Saproling creature token."
//
// The drain reads the Saproling as it died (CR 603.10a), so a token
// counts and an opponent's Saproling does not. Slimefoot may itself be the
// one that has gone: the damage is dealt by it as it last existed (CR 608.2h).
// "Saproling" is a word of this card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd541139-de13-4b32-93ba-070ecf711d33",
		Name:         "Slimefoot, the Stowaway",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, gapB101SaprolingYouControlDied,
				"Slimefoot, the Stowaway — 1 damage to each opponent; gain 1 life",
				effects.Do(effects.DamageEachOpponent{N: 1}, effects.GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}: Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{4}"),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}),
		}},
	})
}

// gapB101SaprolingYouControlDied is "whenever a Saproling you control dies".
func gapB101SaprolingYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := effects.DiedCreature(ev, g)
	return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Saproling"))
}
