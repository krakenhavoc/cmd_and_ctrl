package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// regenerate_this.go — the activated row "<cost>: Regenerate this
// creature." (#5634: Gorilla Chieftain, Lim-Dûl's High Guard, Soldevi
// Steam Beast). Resolving it makes a regeneration shield on the
// ability's source: the next time it would be destroyed this turn, its
// damage is removed and it is tapped and removed from combat instead
// (CR 701.19a). A source that has left the battlefield gets nothing.
//
// Append-only: add a builder, never change what one means.

// regenerateThisRow is "<cost>: Regenerate this creature." for a mana
// cost such as "{1}{G}".
func regenerateThisRow(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:   cost + ": Regenerate this creature.",
		Purpose: game.Purpose{Answers: game.AnswerSurvive},
		Cost:    ManaCost(cost),
		Effect:  regenerateThisSource,
	}
}

// regenerateThisSource is the body: one shield on the item's source.
func regenerateThisSource(g *game.Game, item *game.StackItem) error {
	return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
