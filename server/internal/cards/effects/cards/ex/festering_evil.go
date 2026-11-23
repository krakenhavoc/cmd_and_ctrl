package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Festering Evil — Enchantment {3}{B}{B} (Jumpstart 2022):
//
//	"At the beginning of your upkeep, this enchantment deals 1 damage to each
//	 creature and each player.
//	 {B}{B}, Sacrifice this enchantment: It deals 3 damage to each creature and
//	 each player."
//
// Dry Spell's sweep (every creature and every player as one instance of
// damage) at the upkeep and, as the sacrifice ability, for 3. The Evil is the
// damage source, as it last existed once it has been sacrificed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0581b047-ffa1-42cc-a175-b7bf928e279e",
		Name:         "Festering Evil",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Festering Evil — 1 damage to each creature and each player", func(g *game.Game, item *game.StackItem) error {
				return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 1)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}, Sacrifice this enchantment: It deals 3 damage to each creature and each player.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
			Cost:    effects.Plus(effects.ManaCost("{B}{B}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 3)
			},
		}},
	})
}
