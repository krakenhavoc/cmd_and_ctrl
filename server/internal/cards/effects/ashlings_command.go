package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashling's Command — Kindred Instant — Elemental {3}{U}{R} (Lorwyn
// Eclipsed, #4458):
//
//	"Choose two —
//	 • Create a token that's a copy of target Elemental you control.
//	 • Target player draws two cards.
//	 • Ashling's Command deals 2 damage to each creature target player
//	   controls.
//	 • Target player creates two Treasure tokens."
//
// Prismari Command's "choose two" template (ChooseN 2..2): each chosen
// bullet has its own target group, checked again against its own clause
// as the spell resolves (CR 608.2b), and the bullets run in printed order
// (CR 608.2c), so a spell with one target left still does that bullet (the
// ruling). The copy bullet is KindredCommandCopyMode: the token copies
// the Elemental's copiable values (CR 707.2), and its enters abilities
// fire. The damage bullet is Balefire Dragon's damage to each creature
// one player controls, the set fixed before any is dealt; the Treasures
// bullet is Casey & Raph's, under the target player's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9651d197-a408-4665-a292-58d5471f4ffc",
		Name:         "Ashling's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			KindredCommandCopyMode("Elemental"),
			TargetPlayerDrawsTwoMode(),
			ModeWithPurpose(ModeDoing("Ashling's Command deals 2 damage to each creature target player controls.",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return b11DamageEachCreatureControlledBy(ctx.Game, item, t.ID, 2)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}}),
			ModeWithPurpose(ModeDoing("Target player creates two Treasure tokens.",
				TargetPlayer("target player"), caseyAndRaphTreasures),
				ForTargets(game.TargetPurpose{Slot: 0, Tokens: 2})),
		),
	})
}
