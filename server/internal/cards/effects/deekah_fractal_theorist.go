package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deekah, Fractal Theorist — Legendary Creature — Human Wizard {4}{U},
// 3/3 (Secrets of Strixhaven Commander, #4834):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 create a 0/0 green and blue Fractal creature token. Put X +1/+1
//	 counters on it, where X is that spell's mana value.
//	 {3}{U}: Target creature token can't be blocked this turn."
//
// The magecraft trigger fires once per cast and once per copy
// (Magecraft). X is read as the trigger resolves: the spell on the stack
// with its announced X (CR 202.3e), or the spell as it last existed on
// the stack when it was countered in response (CR 608.2h,
// socB4TriggeringSpellManaValue).
//
// The token is created first and the counters are then PUT on it, as
// printed, rather than entering with them: the second step finds the
// tokens the first made in the event log after the cursor it filed, so
// a creation that waits on a question (a CR 616 order) still gets its
// counters, and a doubled creation puts X on each Fractal. With X = 0
// the Fractal is a 0/0 and dies to CR 704.5f, as in paper.
//
// No simplification.
const deekahFractalTheoristOracleID = "6b071146-2f83-4fda-a55a-46fcb74e6f28"

// deekahBeforeFractal files the event cursor and X before the creation.
// An on-disk key.
const deekahBeforeFractal = "deekah-fractal-theorist/before-fractal"

func init() {
	Register(Spec{
		OracleID:     deekahFractalTheoristOracleID,
		Name:         "Deekah, Fractal Theorist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Magecraft("Deekah, Fractal Theorist — create a Fractal with X +1/+1 counters (magecraft)",
				Do(
					Instruction(func(ctx *Context) error {
						x := socB4TriggeringSpellManaValue(ctx.Game, ctx.Item)
						ctx.RecordStep(deekahBeforeFractal, nil, int(b25LastEventSeq(ctx.Game)), x)
						_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 green and blue Fractal")), 1, game.TokenEntryOptions{})
						return err
					}),
					Instruction(func(ctx *Context) error {
						rec := ctx.StepRecord(deekahBeforeFractal)
						if len(rec.Ns) < 2 || rec.Ns[1] <= 0 {
							return nil
						}
						for _, id := range b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(rec.Ns[0])) {
							if !onBattlefield(ctx.Game, id) {
								continue
							}
							if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: rec.Ns[1]}).Apply(ctx.asGroupMember()); err != nil {
								return err
							}
						}
						return nil
					}),
				)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Target creature token can't be blocked this turn",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    ManaCost("{3}{U}"),
			Targets: TargetCreature("target creature token", IsTokenPredicate()),
			Effect:  targetCantBeBlockedThisTurn("Deekah, Fractal Theorist"),
		}},
	})
}
