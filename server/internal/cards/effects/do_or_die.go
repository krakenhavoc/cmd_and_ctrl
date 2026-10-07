package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Do or Die — Sorcery {1}{B}:
//
//	"Separate all creatures target player controls into two piles.
//	 Destroy all creatures in the pile of that player's choice. They
//	 can't be regenerated."
//
// The first pile split over permanents. Fact or Fiction's two-prompt
// PileSplit is the machinery: the CASTER separates the target player's
// creatures into two piles (a PendingChoiceRevealPick addressed to the
// caster, an empty pile being legal), then the TARGET PLAYER chooses
// which pile is destroyed (a PendingChoiceOptionPick addressed to them,
// each option carrying its pile). That is the printed order of
// decisions: the opponent chooses, the caster divides. Creatures on the
// battlefield are public, so neither prompt needs a reveal.
//
// The destruction is one simultaneous event over the pile
// (CR 701.8a), ignoring regeneration shields (CR 701.19c), and an
// indestructible creature in it survives. The creatures are taken as
// the spell resolves; one that has left by the time the choice is
// answered is not destroyed. A target who controls no creatures
// resolves nothing and asks nothing. A target player who has left the
// game before the pick leaves the first pile as the pile destroyed
// (the engine's answer for a chooser who is gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "f25af42d-f1bf-4bd3-aded-ecadafdbf6e6",
		Name:     "Do or Die",
		// ADR 0126 §6: one pile of the target player's creatures.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true, Partial: true}},
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				return PileSplit{
					Splitter:      ctx.Controller(),
					Chooser:       t.ID,
					Owner:         t.ID,
					SplitQuestion: "Do or Die — separate these creatures into two piles",
					PickQuestion:  "Do or Die — choose a pile; every creature in it is destroyed and can't be regenerated",
					Cards:         creaturesControlledByPlayer(ctx.Game, t.ID),
					Then:          doOrDieDestroyPile,
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// doOrDieDestroyPile destroys the pile the target player chose ("taken"
// is the pile the chooser picked) and leaves the other alone.
func doOrDieDestroyPile(ctx *Context, destroyed, _ []uuid.UUID) error {
	if len(destroyed) == 0 {
		return nil
	}
	ctx.Game.DestroyPermanentsForEffect(destroyed, game.DestroyOptions{CantBeRegenerated: true})
	return nil
}
