package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wren's Run Packmaster — Creature — Elf Warrior {3}{G} 5/5 (Commander 2014):
//
//	"Champion an Elf (When this creature enters, sacrifice it unless you
//	 exile another Elf you control. When this creature leaves the battlefield,
//	 that card returns to the battlefield.)
//	 {2}{G}: Create a 2/2 green Wolf creature token.
//	 Wolves you control have deathtouch."
//
// The shared champion helper (CR 702.72) handles the enters choice and the
// return; the lord line grants deathtouch to every Wolf its controller has,
// the Packmaster's own tokens included.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4d27fef-e86b-4dd3-a15e-a0db8d0725ee",
		Name:         "Wren's Run Packmaster",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Wolf"}, YoursOnly: true}, "deathtouch"),
		},
		Triggered: []game.TriggeredAbility{wrensRunPackmasterChampion.Trigger("Wren's Run Packmaster")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Create a 2/2 green Wolf creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{2}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// wrensRunPackmasterChampion is "Champion an Elf". Its pick key is an
// on-disk identity: never renamed, never reused.
var wrensRunPackmasterChampion = effects.NewChampion("Wren's Run Packmaster", "Elf", "choose-cards/s82-old-b343-wrens-run-packmaster-champion", game.TailRef{})
