package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eladamri, Lord of Leaves — Legendary Creature — Elf Warrior {G}{G}, 2/2
// (Tempest, S82 OLD, #5628):
//
//	"Other Elf creatures have forestwalk. (They can't be blocked as long as
//	 defending player controls a Forest.)
//	 Other Elves have shroud. (They can't be the targets of spells or
//	 abilities.)"
//
// Two layer-6 keyword grants over every OTHER Elf at the table, an opponent's
// included (Eladamri says "you control" of neither), read through the post-layer
// subtypes so a changeling counts. "Other" excludes Eladamri himself, who has
// neither keyword.
//
// No simplification.
func init() {
	elves := effects.TribeFilter{Tribes: []string{"Elf"}, Others: true}
	effects.Register(effects.Spec{
		OracleID:     "dcbb6b0e-2219-4ba4-9720-3e58b59b4328",
		Name:         "Eladamri, Lord of Leaves",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(elves, "forestwalk"),
			effects.TribalKeywordGrant(elves, "shroud"),
		},
	})
}
