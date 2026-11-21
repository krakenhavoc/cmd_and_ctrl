package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryotheory Adept — Creature — Human Wizard {1}{U}, 2/1:
//
//	"Prowess
//	 {3}{U}, Exile this card from your graveyard: Tap target creature
//	 and put a stun counter on it. Activate only as a sorcery."
//
// The exile is a cost, so the ability is on the stack with the card
// already gone from the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a9b3ee21-1af1-4e7e-8e48-d7c90d056f5c",
		Name:            "Cryotheory Adept",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{U}, Exile this card from your graveyard: Tap target creature and put a stun counter on it. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}{U}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect:       effects.TapTargetAndStunIt,
		}},
	})
}
