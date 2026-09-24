package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tree of Perdition — Creature — Plant {3}{B}, 0/13 :
//
//	"Defender.
//	 {T}: Exchange target opponent's life total with this creature's
//	 toughness."
//
// A four-mana wall that takes an opponent from 40 to 13, and then
// keeps going: the Tree's toughness is now 40, so the next activation
// takes the NEXT opponent to 40 and leaves the Tree at whatever they
// had. Point it at the same seat twice and you have put them back
// where they started. The deck it belongs to pairs it with a drain
// that makes 13 a finishing range.
//
// THE TWO HALVES ARE DIFFERENT MACHINERY, and getting the second one
// wrong is the whole risk in this card:
//
//   - The LIFE half is an ordinary set — the player gains or loses
//     the difference (b31LifeBecomes), so the CR 614 replacement
//     window runs and lifegain / life-loss payoffs see it.
//   - The TOUGHNESS half is a CONTINUOUS EFFECT with NO STATED
//     DURATION (CR 611.2a): the Tree's toughness is 40 for the rest
//     of the game, not until end of turn. It is a layer 7b "set"
//     (CR 613.4c), registered at resolution with IndefiniteDuration
//     — the same shape ExchangeControlForEffect registers its two
//     halves with, and structurally BecomeCreatureUntilEOT's 7b set
//     with the duration swapped.
//
// Two consequences worth stating:
//
//   - REPEATED ACTIVATIONS ACCUMULATE. Each one registers a NEW
//     scoped effect rather than replacing the previous one. That is
//     correct — CR 613.6 orders effects in the same sublayer by
//     timestamp, so the newest set wins and the board is right — but
//     the superseded entries stay in Game.ScopedEffects and are
//     walked by every recompute for as long as the Tree stays. A Tree
//     that untaps every turn leaves a slow trail of dead 7b entries.
//     Not worth a sweep for one card; worth knowing before someone
//     writes the second one.
//   - The effect is pinned to (InstanceID, EnteredBattlefieldAt), so
//     a Tree that is flickered comes back as its printed 0/13 rather
//     than keeping the set (CR 400.7), and its entries are swept.
//
// The set is a data record (ADR 0041 phase 3, #1497), so a table
// with an activated Tree is still a restore point.
//
// The toughness is READ before the static is registered and read
// through CurrentToughness, so it is the post-layer value with
// counters — a Tree with a +1/+1 counter hands over 14, and the set
// that lands on it is still modified by that counter afterwards.
// Equal values exchange nothing and register nothing, rather than
// leaving a static that says what was already true.
//
// Defender is a printed keyword and rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0cdece2a-0bdf-4e6d-9ddc-4a8d58b2ec29",
		Name:            "Tree of Perdition",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Exchange target opponent's life total with this creature's toughness.",
			Cost:    TapCost(),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect:  treeOfPerditionExchange,
		}},
	})
}

// treeOfPerditionExchange swaps the chosen opponent's life total with
// the source's toughness.
func treeOfPerditionExchange(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	p := ctx.PlayerByID(victim)
	if p == nil {
		return nil
	}
	// CR 611.2c: the affected set is fixed now. Nil means the Tree
	// has left the battlefield, in which case there is no toughness
	// to exchange with and nothing happens.
	source := ctx.Source()
	affected := eotSnapshot(ctx, source, nil)
	if affected == nil {
		return nil
	}
	// The toughness has to be read BEFORE the set that replaces it,
	// and it has to be the post-layer value: recompute first, because
	// an earlier activation of this same ability is one of the layers
	// being read.
	g.RecomputeLayersIfStaleLocked()
	tree, found := g.LookupCardForEffect(source)
	if !found {
		return nil
	}
	toughness, life := tree.CurrentToughness(), p.Life
	if toughness == life {
		return nil
	}
	// A data record (ADR 0041 phase 3, #1497): the set lasts the rest
	// of the game, and as a closure it kept the table off the restore
	// path for all of it. Pinned to the Tree, so a Tree that leaves
	// takes its entries with it.
	g.RegisterScopedEffectForEffect(source, affected.affectedObjects(),
		[]game.Mod{game.SetBaseToughnessMod(life)},
		g.PinnedTo(game.IndefiniteDuration(), source),
		fmt.Sprintf("Tree of Perdition — toughness becomes %d", life))
	return b31LifeBecomes(ctx, victim, toughness)
}
