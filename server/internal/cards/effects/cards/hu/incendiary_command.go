package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incendiary Command — Sorcery {3}{R}{R} (Commander 2013, S82 OLD b380, #5560):
//
//	"Choose two —
//	 • Incendiary Command deals 4 damage to target player or planeswalker.
//	 • Incendiary Command deals 2 damage to each creature.
//	 • Destroy target nonbasic land.
//	 • Each player discards all the cards in their hand, then draws that many
//	   cards."
//
// Brigid's Command's "choose two" walk, the bullets resolving in printed
// order: the burn at a player or planeswalker, the sweep through
// DamageEachMatching, Magmatic Hellkite's nonbasic land, and the wheel
// (S82OldB380DiscardHandsThenDraw).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d45a4924-daa0-4ac3-afd7-b66f636ce870",
		Name:         "Incendiary Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Incendiary Command deals 4 damage to target player or planeswalker.",
				effects.TargetPlayerOrPlaneswalker(),
				effects.DealFixedDamageToModesTarget(4)), effects.ForTargets(effects.DamageToTarget(0, 4))),
			effects.ModeWithPurpose(effects.ModeDoing("Incendiary Command deals 2 damage to each creature.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.Creature(), 2)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}}),
			effects.ModeDoing("Destroy target nonbasic land.",
				effects.TargetPermanent("target nonbasic land", effects.Land(), effects.B03Nonbasic()),
				effects.DestroyTheModesTarget),
			effects.ModeSteps("Each player discards all the cards in their hand, then draws that many cards.", nil,
				effects.Instruction(effects.S82OldB380DiscardHandsThenDraw)),
		),
	})
}
