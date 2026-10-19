package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martial Coup — Sorcery {X}{W}{W} (EDHREC rank 1398):
//
//	"Create X 1/1 white Soldier creature tokens. If X is 5 or more,
//	 destroy all other creatures."
//
// A one-sided Wrath at seven mana and a token maker at less. The
// tokens are created FIRST, then — when X is five or more — every
// creature that is not one of them is destroyed in one simultaneous
// event (DestroyAllMatching, so dies-triggers see the whole batch).
// "Other" is the exclusion by instance ID, so the freshly made
// Soldiers survive and every other creature, the caster's included,
// does not — as printed.
//
// No simplification on the card. The engine gap it shared with every
// wipe in the catalog — the simultaneous destroy path not consulting
// indestructible the way the single-target path had since #380 — was
// reported on #305 and fixed in S30 (#470 / #446).
func init() {
	Register(Spec{
		OracleID: "2b7c4dab-e432-4b34-b058-3cec5c0d72df",
		Name:     "Martial Coup",
		// ADR 0126 §6: the sweep happens only when X is 5 or more, and spares the Soldiers.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		XMatters:     true,
		Completeness: CompletenessFull,
		// The sweep waits for a prompt the tokens ask (a CR 616 order of
		// two token replacements; CR 608.2c, #4163), so the tokens it
		// spares are the ones that were really made: every Soldier the
		// caster created after the step's record of the event log.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				ctx.RecordStep(martialCoupRecord, nil, int(b25LastEventSeq(ctx.Game)))
				_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("1/1 white Soldier")), ctx.X(), game.TokenEntryOptions{})
				return err
			}),
			Instruction(func(ctx *Context) error {
				if ctx.X() < 5 {
					return nil
				}
				made := b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(ctx.StepRecord(martialCoupRecord).N()))
				return DestroyAllMatching{Match: Except(Creature(), b10OneOf(made))}.Apply(ctx)
			}),
		),
	})
}

// martialCoupRecord is the last event logged before the Soldiers.
const martialCoupRecord = "martial-coup/before-tokens"
