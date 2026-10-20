package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Umaro, Raging Yeti — Legendary Creature — Yeti Berserker {5}{R}, 6/6:
//
//	"Trample
//	 At the beginning of combat on your turn, choose one at random —
//	 • Other creatures you control get +3/+0 and gain trample until end
//	   of turn.
//	 • Discard your hand, then draw four cards.
//	 • Umaro deals 5 damage to any target."
//
// ChooseOneAtRandom (CR 700.2, #3399) draws among the bullets that can
// be taken. "Any target" always has a legal target, so all three are
// always in the pool. The first bullet locks its set as it resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	const name = "Umaro, Raging Yeti"
	t := AtBeginningOfYourCombat(name+" — choose one at random: pump the team, wheel your hand, or 5 damage",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOneAtRandom(
		ModeDoing("Other creatures you control get +3/+0 and gain trample until end of turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				others := And(Creature(), YouControl(), OtherThan(item.SourceCardID))
				if err := (BoostUntilEOT{Match: others, Power: 3, Label: name + " — +3/+0"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: others, Keywords: []string{"trample"}, Label: name + " — trample"}.Apply(ctx)
			}),
		// The draw waits for a prompt a discard asks (a replacement offer such as Library of Leng's; CR 608.2c, #4155).
		ModeSteps("Discard your hand, then draw four cards.", nil,
			Instruction(func(ctx *Context) error {
				_, err := discardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			DrawCards{N: 4}),
		ModeDoing("Umaro deals 5 damage to any target.", TargetAny(),
			func(item *game.StackItem, ctx *Context, occ int) error {
				tgt, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: tgt.ID, Amount: 5}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "0f4350ad-dc44-4942-b097-e4033e29d659",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
