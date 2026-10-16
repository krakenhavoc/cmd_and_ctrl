package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Maralen of the Mornsong — Legendary Creature — Elf Wizard {1}{B}{B},
// 2/3:
//
//	"Players can't draw cards.
//	 At the beginning of each player's draw step, that player loses 3
//	 life, searches their library for a card, puts it into their hand,
//	 then shuffles."
//
// #3927. "Players can't draw cards" is the draw limit at zero
// (Spec.DrawLimits, CR 121.3): no draw is performed, the draw step's
// included, and none can be replaced. The draw-step trigger is a search,
// not a draw (CR 121.5), so the limit doesn't touch it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b40032b-711d-499f-8bb9-87855f1fce9e",
		Name:         "Maralen of the Mornsong",
		Completeness: CompletenessFull,
		DrawLimits:   PlayersCantDrawCards(),
		Triggered:    []game.TriggeredAbility{mornsongDrawStep("Maralen of the Mornsong")},
	})
}
