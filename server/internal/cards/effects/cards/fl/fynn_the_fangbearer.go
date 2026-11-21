package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fynn, the Fangbearer — Legendary Creature — Human Warrior {1}{G}, 1/3
// (Foundations, #4465):
//
//	"Deathtouch
//	 Whenever a creature you control with deathtouch deals combat damage to
//	 a player, that player gets two poison counters. (A player with ten or
//	 more poison counters loses the game.)"
//
// Black Widow, Deadly Hunter's watch (deathtouchCreatureYouControlHitAPlayer:
// one event per creature, deathtouch read as the dealer has it now, so a
// granted deathtouch counts and Fynn himself counts) with Virulent
// Silencer's payoff (virulentSilencerPoison): "that player" is the damaged
// player, read off the triggering event. Ten poison counters lose the game
// as a state-based action (CR 704.5c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c0b1fba1-4338-4671-934b-098689ad2085",
		Name:            "Fynn, the Fangbearer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.DeathtouchCreatureYouControlHitAPlayer,
				"Fynn, the Fangbearer — that player gets two poison counters", effects.VirulentSilencerPoison),
		},
	})
}
