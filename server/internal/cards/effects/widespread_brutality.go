package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Widespread Brutality — Sorcery {1}{B}{R}{R} (EDHREC rank 7598):
//
//	"Amass Zombies 2, then the Army you amassed deals damage equal to
//	 its power to each non-Army creature."
//
// The proof card for the two halves of amass that no other card in the
// catalog exercises, which is why it is here rather than on a roadmap
// batch's list:
//
//   - CR 701.47c's "the Army you amassed" — the creature the amass
//     CHOSE, which is what `Amass.Then` is handed. Re-deriving it
//     afterwards is not possible: with two Armies out, the one that
//     grew is the one the player picked at the prompt, and the board
//     alone does not say which.
//   - "each NON-Army creature" — the read side, as `Not(Army())`. The
//     Army that just amassed is excluded, and so is every other Army
//     on the table, including an opponent's.
//
// The power is read AFTER the counters land, so an Army amassed from
// nothing deals 2 and an Army that was already a 4/4 deals 6. That
// ordering is `finishAmassLocked`'s contract and not this file's
// arrangement.
//
// Sandbox simplification, declared: the damage is dealt creature by
// creature in battlefield order rather than simultaneously, so a
// creature that dies to it is gone before the next one is dealt to —
// visible only through a death trigger that changes the Army's power
// mid-sweep (a +1/+1 counter from a Vampiric Rites-style death
// payoff). Every other catalog card that deals one source's damage to
// many creatures takes the same posture.
func init() {
	Register(Spec{
		OracleID:     "3a06f4c1-2af1-495f-914f-1ac4e26f87d4",
		Name:         "Widespread Brutality",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Army's damage hits each creature one after another rather than all at once."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Amass{Subtype: "Zombie", N: 2, Then: widespreadBrutalitySweep}.Apply(ctx)
		},
	})
}

// widespreadBrutalitySweep is "then the Army you amassed deals damage
// equal to its power to each non-Army creature".
//
// A `uuid.Nil` army is the CR 701.47b case — the amass happened, but
// there was no Army to choose and none could be made (a token creation
// replaced away). Nothing deals damage then, and the spell has still
// resolved.
func widespreadBrutalitySweep(ctx *Context, army uuid.UUID) error {
	if army == uuid.Nil {
		return nil
	}
	// CurrentPower, not Effective().Power: the +1/+1 counters the amass
	// just placed are added ON TOP of the layered value
	// (Card.PowerForComparison), so the effective read alone would see
	// the 0/0 body the token entered with and the sweep would deal
	// nothing. The recompute is for the other half — an anthem or a
	// lord registered this turn — the way Fell the Mighty and Tree of
	// Perdition take it before reading a P/T at resolution.
	ctx.Game.RecomputeLayersIfStaleLocked()
	src, ok := ctx.Game.LookupCardForEffect(army)
	if !ok {
		return nil
	}
	n := src.CurrentPower()
	if n <= 0 {
		return nil
	}
	// The victims are snapshotted before the first point of damage —
	// CR 608.2's "objects as they existed when the spell began
	// resolving", and the reason a creature killed by the sweep is
	// not skipped when the sweep reaches it.
	//
	// "each NON-Army creature" is the read side amass ships, composed
	// rather than open-coded: the Army the spell just amassed is spared,
	// and so is every other Army on the table, an opponent's included.
	nonArmyCreature := And(Creature(), Not(Army()))
	var victims []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if nonArmyCreature(ctx.Game, ctx.Controller(), c) {
			victims = append(victims, c.InstanceID)
		}
	}
	for _, v := range victims {
		if err := (DealDamage{Source: army, Target: v, Amount: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
