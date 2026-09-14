package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// Sandbox simplification, declared (the Azorius Chancery / Mount
// Doom posture): "return a creature you control" is a resolution-
// time choice, not a target, and the pick_target prompt is the one
// picker the engine has for choosing among permanents — so the
// creature is chosen at announce as a target. Three consequences,
// all weaker than printed: opponents see the choice before the spell
// resolves; a creature you control with hexproof or shroud cannot
// be the one saved; and with no creature of your own the spell has
// no legal target and cannot be cast at all, where the printed card
// is simply a five-mana Wrath. If the chosen creature is gone by
// resolution the spell is countered by game rules (CR 608.2b) and
// the sweep does not happen — also weaker.
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
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The creature you return to hand is picked when you cast the spell rather than as it resolves, so opponents can respond to the choice, a creature with hexproof or shroud can't be picked, and you need a creature of your own to cast it at all.",
		},
		Targets: TargetCreature("a creature you control to return to hand", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			targets := ctx.Targets()
			if len(targets) > 0 && targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(targets[0]) {
				if err := (BounceToHand{Target: targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return DestroyAllMatching{Match: Creature()}.Apply(ctx)
		},
	})
}
