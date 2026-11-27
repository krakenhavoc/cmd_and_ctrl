package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burning of Xinye — Sorcery {4}{R}{R} (Portal Three Kingdoms, S82 #5622):
//
//	"You destroy four lands you control, then target opponent destroys four lands
//	 they control. Then Burning of Xinye deals 4 damage to each creature."
//
// Three steps in the printed order (CR 608.2c). The caster chooses four lands they
// control and destroys them, the target opponent chooses four they control and
// destroys them, and only then does every creature take 4 damage (the creatures
// are read after the lands are gone, so a creature that was a land is hit as it
// then is). A player with four or fewer lands destroys them all and is asked
// nothing; destruction, not sacrifice, so regeneration or indestructible saves a
// land. With the opponent an illegal target the spell does nothing (CR 608.2b),
// the caster's lands included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a3cec13-4766-40d2-88e0-41bcdd5249e9",
		Name:         "Burning of Xinye",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if _, ok := ctx.ClauseTarget(0); !ok {
				return nil
			}
			return effects.OldB708FourLandsOf(ctx, ctx.Controller(), burningOfXinyeMine,
				"Burning of Xinye — choose four lands you control to destroy", burningOfXinyeOpponent)
		},
	})
}

// burningOfXinyeMine is the caster's four lands, chosen: they are destroyed and the
// opponent's turn to choose comes. An on-disk key: never renamed, never reused.
var burningOfXinyeMine = effects.AskedPick("choose-cards/s82-old-b708-burning-of-xinye-mine", func(ctx *effects.Context, r game.CardsChosen) error {
	if err := effects.OldB708DestroyLands(ctx, r.Picked); err != nil {
		return err
	}
	return burningOfXinyeOpponent(ctx)
})

// burningOfXinyeOpponent is the opponent's half: the target chooses four of theirs.
func burningOfXinyeOpponent(ctx *effects.Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return effects.OldB708FourLandsOf(ctx, t.ID, burningOfXinyeTheirs,
		"Burning of Xinye — choose four lands you control to destroy", burningOfXinyeDamage)
}

// burningOfXinyeTheirs is the opponent's four lands, chosen: destroyed, then the damage.
// An on-disk key: never renamed, never reused.
var burningOfXinyeTheirs = effects.AskedPick("choose-cards/s82-old-b708-burning-of-xinye-theirs", func(ctx *effects.Context, r game.CardsChosen) error {
	if err := effects.OldB708DestroyLands(ctx, r.Picked); err != nil {
		return err
	}
	return burningOfXinyeDamage(ctx)
})

// burningOfXinyeDamage is "Then Burning of Xinye deals 4 damage to each creature."
func burningOfXinyeDamage(ctx *effects.Context) error {
	return effects.DamageEachMatching(ctx, effects.Creature(), 4)
}
