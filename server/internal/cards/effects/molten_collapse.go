package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Molten Collapse — Sorcery {B}{R} (The Lost Caverns of Ixalan, #4470):
//
//	"Choose one. If you descended this turn, you may choose both instead.
//	 (You descended if a permanent card was put into your graveyard from
//	 anywhere.)
//	 • Destroy target creature or planeswalker.
//	 • Destroy target noncreature, nonland permanent with mana value 1 or
//	   less."
//
// The mode count is the Commander Legends Wills' conditional one
// (OrUpToIf(2, …)), with "you descended this turn" (CR 700.11) as its
// condition, read for the caster as the modes are chosen (CR 601.2b) and
// fixed from then on. Each bullet has its own target and is Archive
// Arbiter's destroy, skipped when its target is no longer legal
// (CR 608.2b); with both chosen they resolve in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0d1636e-6f80-44fe-869a-1749ab249815",
		Name:         "Molten Collapse",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				DestroyTheModesTarget),
			ModeDoing("Destroy target noncreature, nonland permanent with mana value 1 or less.",
				TargetPermanent("target noncreature, nonland permanent with mana value 1 or less", Noncreature(), Nonland(), ManaValueLE(1)),
				DestroyTheModesTarget),
		).OrUpToIf(2, youDescendedThisTurnForModes),
	})
}

// youDescendedThisTurnForModes is Molten Collapse's mode count — "If you
// descended this turn, you may choose both instead" — read for the
// chooser as the modes are chosen.
var youDescendedThisTurnForModes = game.ModeCondition("descended-this-turn", func(g *game.Game, chooser uuid.UUID) bool {
	return descendedThisTurn(g, chooser)
})
