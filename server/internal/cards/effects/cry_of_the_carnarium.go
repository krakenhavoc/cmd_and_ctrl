package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cry of the Carnarium — Sorcery {1}{B}{B}:
//
//	"All creatures get -2/-2 until end of turn. Exile all creature cards
//	 in all graveyards that were put there from the battlefield this
//	 turn. If a creature would die this turn, exile it instead."
//
// Printed order is resolution order (CR 608.2c). The -2/-2 locks its set
// as it begins (CR 611.2c), so a creature cast later this turn is not
// shrunk. The middle sentence reads this turn's events: a creature card
// counts if its most recent arrival in a graveyard was from the
// battlefield this turn. A card that left the graveyard and came back is
// a new object, so only that newest arrival counts (CR 400.7). The
// creatures the -2/-2 is about to kill are still on the battlefield
// then, so they are not exiled by it (the 2019-01-25 ruling). They die
// after the spell finishes resolving, and the replacement exiles them.
// "If a creature would die this turn" is read live (CR 611.2c), so it
// reaches a creature that was not on the battlefield as the spell
// resolved.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1b9a5170-39c0-4cbf-a041-f3c15f1359ae",
		Name:         "Cry of the Carnarium",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{Match: Creature(), Power: -2, Toughness: -2}).Apply(ctx); err != nil {
				return err
			}
			if err := exileCreatureCardsPutIntoGraveyardsFromBattlefieldThisTurn(ctx); err != nil {
				return err
			}
			return ExileIfCreaturesWouldDieThisTurn{}.Apply(ctx)
		},
	})
}
