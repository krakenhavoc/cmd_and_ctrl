package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exquisite Huntmaster — Creature — Elf Warrior {3}{B}, 4/2 (Commander
// Legends, #5507):
//
//	"When this creature dies, create a 1/1 green Elf Warrior creature
//	 token.
//	 Encore {4}{B}"
//
// The dies trigger makes one token for its controller; encore is the
// shared constructor (one hasty token copy per opponent, each bound to
// attack that opponent, sacrificed at the next end step). An encore copy
// that dies makes an Elf Warrior too, since it copies the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "393ef069-ea6b-4969-b878-d942281cc918",
		Name:         "Exquisite Huntmaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Exquisite Huntmaster — create a 1/1 green Elf Warrior",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.Encore("{4}{B}")},
	})
}
