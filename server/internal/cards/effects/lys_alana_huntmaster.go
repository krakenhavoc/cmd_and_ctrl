package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lys Alana Huntmaster — Creature — Elf Warrior {2}{G}{G}, 3/3
// (EDHREC rank 2301):
//
//	"Whenever you cast an Elf spell, you may create a 1/1 green Elf
//	 Warrior creature token."
//
// The Elf deck's token engine. Leaf-Crowned Visionary's cast
// condition (b17ElfSpellCastByYou — the spell is read off the stack,
// so a changeling counts) with "you may" as the prompt and the
// batch 13 Elf Warrior as the payoff. The token arrives before the
// Elf spell resolves (LIFO), as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f3439e1-75ce-482d-881f-836492dca6e9",
		Name:         "Lys Alana Huntmaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17ElfSpellCastByYou(ev, source, g)
			}, "Lys Alana Huntmaster — create a 1/1 green Elf Warrior", Do(CreateToken{Template: b13GreenElfWarriorToken(), N: 1})), "Lys Alana Huntmaster — create a 1/1 Elf Warrior?"),
		},
	})
}
