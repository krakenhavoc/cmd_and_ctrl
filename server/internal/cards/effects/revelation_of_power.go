package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revelation of Power — Instant {1}{W} (Streets of New Capenna):
//
//	"Target creature gets +2/+2 until end of turn. If it has a counter on
//	 it, it also gains flying and lifelink until end of turn."
//
// One effect (CR 613.7): the +2/+2 in layer 7c, and, when the target
// carries a counter of any kind as the spell resolves, flying and
// lifelink in layer 6 on the same record. A target that has become
// illegal gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "287f60a6-e83b-487b-add9-2048a06dcca1",
		Name:         "Revelation of Power",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerPump,
			Targets: game.ForTargets(game.TargetPurpose{
				Slot: 0, UntilEndOfTurn: game.TargetBenefit,
				Pump: &game.Pump{Power: 2, Toughness: 2},
			}),
		},
		Targets: TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			id := ts[0].ID
			mods := []game.Mod{game.ModifyPTMod(2, 2)}
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && s82SncHasACounter(c) {
				mods = append(mods, game.AddKeywordsMod("flying", "lifelink"))
			}
			return untilEndOfTurn(ctx, id, nil, "Revelation of Power — +2/+2, and flying and lifelink if it has a counter", mods...)
		},
	})
}
