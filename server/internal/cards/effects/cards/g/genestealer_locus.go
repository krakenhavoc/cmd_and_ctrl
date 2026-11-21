package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Genestealer Locus — Creature — Tyranid Human {3}{U}, 3/3 (Warhammer 40,000
// Commander, S82 GAP batch 195, #5948):
//
//	"Neurotraumal Rod — Whenever a creature attacks you, it gets -1/-0 until end
//	 of turn.
//	 Whenever a creature attacks one of your opponents, it gets +0/+1 until end
//	 of turn."
//
// Two triggers over the attack declaration, one per attacking creature (any
// player's), each a Hissing Miasma reading of who is attacked: the first for
// an attack on the Locus's controller, the second for an attack on any other
// player (an attack on a planeswalker or battle is neither). The pump is the
// triggering creature's, if it is still the same creature on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a64cee5-04e2-4cbf-8fa9-d55c854f9500",
		Name:         "Genestealer Locus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.GapB195AnAttackerAttackedYou,
				"Genestealer Locus — it gets -1/-0 until end of turn",
				effects.GapB195AttackerGetsPumpUntilEOT(-1, 0, "Genestealer Locus — -1/-0")),
			effects.On(game.EventAttack, effects.GapB195AnAttackerAttackedOneOfYourOpponents,
				"Genestealer Locus — it gets +0/+1 until end of turn",
				effects.GapB195AttackerGetsPumpUntilEOT(0, 1, "Genestealer Locus — +0/+1")),
		},
	})
}
