package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const DeekahFractalTheoristOracleID = "6b071146-2f83-4fda-a55a-46fcb74e6f28"

// deekahBeforeFractal files the event cursor and X before the creation.
// An on-disk key.
const deekahBeforeFractal = "deekah-fractal-theorist/before-fractal"

func init() {
	effects.Register(effects.Spec{
		OracleID:     DeekahFractalTheoristOracleID,
		Name:         "Deekah, Fractal Theorist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Deekah, Fractal Theorist — create a Fractal with X +1/+1 counters (magecraft)",
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						x := effects.SocB4TriggeringSpellManaValue(ctx.Game, ctx.Item)
						ctx.RecordStep(deekahBeforeFractal, nil, int(effects.B25LastEventSeq(ctx.Game)), x)
						_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(effects.TokenCard("0/0 green and blue Fractal")), 1, game.TokenEntryOptions{})
						return err
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						rec := ctx.StepRecord(deekahBeforeFractal)
						if len(rec.Ns) < 2 || rec.Ns[1] <= 0 {
							return nil
						}
						for _, id := range effects.B27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(rec.Ns[0])) {
							if !effects.OnBattlefield(ctx.Game, id) {
								continue
							}
							if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: rec.Ns[1]}).Apply(ctx.AsGroupMember()); err != nil {
								return err
							}
						}
						return nil
					}),
				)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Target creature token can't be blocked this turn",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{3}{U}"),
			Targets: effects.TargetCreature("target creature token", effects.IsTokenPredicate()),
			Effect:  effects.TargetCantBeBlockedThisTurn("Deekah, Fractal Theorist"),
		}},
	})
}
