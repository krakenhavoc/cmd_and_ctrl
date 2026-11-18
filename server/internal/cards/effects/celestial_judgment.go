package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Celestial Judgment — Sorcery {4}{W}{W} (Midnight Hunt Commander, S82 MID
// batch 13):
//
//	"For each different power among creatures on the battlefield, choose a
//	 creature with that power. Destroy each creature not chosen this way."
//
// The caster chooses as it resolves (CR 608.2c), over every creature on the
// battlefield, through the layers. One creature per power, so the pick is
// exactly as many creatures as there are different powers, with the
// "different powers" set rule (midB13DifferentPowers): that set holds one
// creature of every power, and no power twice. When no two creatures share a
// power every creature is chosen and nothing is asked or destroyed. Every
// creature that was on the battlefield and was not chosen is then destroyed
// together; a creature that has entered since is not touched. Indestructible
// and regeneration apply as for any destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d18c40cd-0f38-498b-8882-10381a831751",
		Name:         "Celestial Judgment",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			creatures := midB13CreaturesOnBattlefield(ctx.Game)
			powers := midB13DistinctPowers(creatures)
			if powers == len(creatures) {
				return nil
			}
			midB13ChooseDifferentPowers(ctx, creatures, powers, powers,
				"Celestial Judgment — for each different power among creatures on the battlefield, choose a creature with that power; destroy each creature not chosen",
				celestialJudgmentPick)
			return nil
		},
	})
}

// celestialJudgmentPick destroys every creature that was on the battlefield
// when the spell began to resolve (the prompt's Carry) and was not chosen.
// An on-disk key.
var celestialJudgmentPick = askedPickRule("choose-cards/celestial-judgment-pick", midB13DifferentPowers,
	func(ctx *Context, r game.CardsChosen) error {
		doomed := make(map[uuid.UUID]bool, len(r.Carry))
		for _, id := range r.Carry {
			doomed[id] = true
		}
		for _, id := range r.Picked {
			delete(doomed, id)
		}
		return DestroyAllMatching{
			Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return doomed[c.InstanceID] && c.IsCreature() },
		}.Apply(ctx)
	})
