package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "393ef069-ea6b-4969-b878-d942281cc918",
		Name:         "Exquisite Huntmaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Exquisite Huntmaster — create a 1/1 green Elf Warrior",
				Do(CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
		Activated: []ActivatedAbility{Encore("{4}{B}")},
	})
}
