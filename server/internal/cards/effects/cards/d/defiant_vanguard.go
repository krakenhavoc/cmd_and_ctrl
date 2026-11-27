package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiant Vanguard — Creature — Human Rebel {2}{W}, 2/2 (Nemesis, S82 OLD,
// #5618):
//
//	"When this creature blocks, at end of combat, destroy it and all creatures
//	 it blocked this turn.
//	 {5}, {T}: Search your library for a Rebel permanent card with mana value 4
//	 or less, put it onto the battlefield, then shuffle."
//
// The block trigger is CR 509.3a's (once however many attackers it blocks); it
// names the Vanguard and the creatures it blocks as the declaration is made,
// each as the object it is now (CR 400.7), and a delayed trigger at the
// beginning of this combat's end of combat step destroys those still on the
// battlefield (CR 511.2). The recruiter is the Rebel search at 4.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "680766e2-637b-40bc-a15f-0b4dc55bd619",
		Name:         "Defiant Vanguard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB677DefiantVanguardBlocks("Defiant Vanguard — destroy it and the creatures it blocked at end of combat"),
		},
		Activated: []effects.ActivatedAbility{
			effects.NemRecruiterRow("Defiant Vanguard", "{5}", "a", "Rebel", 4),
		},
	})
}
