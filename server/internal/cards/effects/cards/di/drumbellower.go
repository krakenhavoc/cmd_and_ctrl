package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drumbellower — Creature — Spirit {2}{W}, 2/1 (EDHREC rank 1918):
//
//	"Flying
//	 Untap all creatures you control during each other player's
//	 untap step."
//
// The white, three-mana, creatures-only half of Seedborn Muse — the
// vigilance-for-the-whole-table effect that makes a board of tappers
// and blockers available on every turn of the rotation rather than
// one in four.
//
// Flying rides PrintedKeywords; the untap clause is the shared
// permission narrowed to creatures. The Spirit untaps itself along
// with the rest, which matters for a creature that spent the turn
// cycle tapped blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "03dee43b-6377-4f7b-956b-a384160322e4",
		Name:            "Drumbellower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		UntapStep: []game.UntapStepPermission{
			effects.UntapDuringEachOtherPlayersUntapStep(
				"Drumbellower — untap all creatures you control",
				func(c game.Card) bool { return c.IsCreature() }),
		},
	})
}
