package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Time Wipe — Sorcery {2}{W}{W}{U} (EDHREC rank 1414):
//
//	"Return a creature you control to its owner's hand, then destroy
//	 all creatures."
//
// A Wrath that saves your best creature — usually the one with the
// enters-the-battlefield trigger you want again. The bounce happens
// first and the sweep second, as one simultaneous destruction event,
// so the saved creature is in hand before anything dies and every
// dies-trigger sees the whole batch.
//
// "Return a creature you control" is a CHOICE, not a target — there
// is no "target" in the printed text — made on resolution (CR 608.2,
// ReturnOneYouControl / ChoosePermanents posture, #1214, #1337). It
// used to be a target clause picked at announce, a declared
// simplification with three consequences, all fixed by the move to a
// resolution-time pick: opponents no longer see the choice before the
// spell resolves; a creature with hexproof or shroud can be the one
// saved; and casting the spell no longer requires a creature of your
// own — with none, the pick is skipped (CR 608.2c, "as much as
// possible") and the sweep still runs, exactly as the printed
// five-mana Wrath would.
//
// The engine gap it shared with every wipe in the catalog — the
// simultaneous destroy path not consulting indestructible the way
// the single-target path had since #380 — was reported on #305 and
// fixed in S30 (#470 / #446). An indestructible creature now
// survives this the way it survives a Wrath of God.
func init() {
	Register(Spec{
		OracleID:     "36c78a5f-0148-4596-a346-f8e35037b694",
		Name:         "Time Wipe",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePermanents{
				Question: "Time Wipe — return a creature you control to its owner's hand",
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var out []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && c.IsCreature() {
							out = append(out, c.InstanceID)
						}
					}
					return out, 1, 1
				},
				Then: func(ctx *Context, picked game.PromptedPicks) error {
					for _, id := range picked.Cards() {
						if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
							return err
						}
					}
					return DestroyAllMatching{Match: Creature()}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
