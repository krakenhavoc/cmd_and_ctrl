package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Epistolary Librarian — Creature — Astartes Wizard {2}{W}{U}, 3/4 (Warhammer
// 40,000 Commander, S82 GAP batch 195, #5948):
//
//	"Veil of Time — Whenever this creature attacks, you may cast a spell with
//	 mana value X or less from your hand without paying its mana cost, where X is
//	 the number of attacking creatures."
//
// Electrodominance's free hand cast (a "you may" resolution_cast offer of the
// castable spells in the hand), with X counted as the trigger resolves: the
// number of creatures attacking then, the Librarian's own attack included. The
// spell is cast during the trigger's resolution (CR 608.2g) at its normal
// timing exemption, and its other costs and restrictions still apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb4f65b0-2841-42bf-8e18-d04a3b9f1a76",
		Name:         "Epistolary Librarian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Epistolary Librarian — you may cast a spell with mana value X or less from your hand without paying its mana cost, X the number of attacking creatures",
				effects.GapB195VeilOfTime),
		},
	})
}
