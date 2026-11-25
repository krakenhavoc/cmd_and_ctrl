package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nivix Guildmage — Creature — Human Wizard {U}{R}, 2/2 (Commander 2017, S82
// older sets):
//
//	"{1}{U}{R}: Draw a card, then discard a card.
//	 {2}{U}{R}: Copy target instant or sorcery spell you control. You may
//	 choose new targets for the copy."
//
// The loot is Agna Qela's, the copy League Guildmage's second ability without
// the mana-value tie: the copy is put on the stack above the original, and its
// controller may choose new targets (CR 707.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d04356f1-0e1a-4689-8e54-f88c4c6dd936",
		Name:         "Nivix Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}{R}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{1}{U}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.LootOne(g, item, 1)
				},
			},
			{
				Label: "{2}{U}{R}: Copy target instant or sorcery spell you control. You may choose new targets for the copy.",
				Cost:  effects.ManaCost("{2}{U}{R}"),
				Targets: effects.TargetSpell("target instant or sorcery spell you control",
					effects.Or(effects.Instant(), effects.Sorcery()), effects.YouControl()),
				Effect: effects.CopyTheFirstLegalTargetSpell,
			},
		},
	})
}
