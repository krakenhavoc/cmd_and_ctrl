package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aragorn and Arwen, Wed — Legendary Creature — Human Elf Noble {4}{G}{W},
// 3/6 (The Lord of the Rings: Tales of Middle-earth, in The Hobbit's deck):
//
//	"Vigilance
//	 Whenever Aragorn and Arwen enters or attacks, put a +1/+1 counter on
//	 each other creature you control. You gain 1 life for each other
//	 creature you control."
//
// The counters are Mikaeus, the Lunarch's (the set is snapshotted before the
// first lands); the life is counted after them, once any prompt a counter
// asked has been answered (CR 608.2c), and counts the creatures you control
// then other than this one, a token a counter payoff made included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b8a3440a-03c5-444c-a1de-252524c3db82",
		Name:            "Aragorn and Arwen, Wed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Aragorn and Arwen — put a +1/+1 counter on each other creature you control; you gain 1 life for each other creature you control",
				Do(
					Instruction(func(ctx *Context) error {
						return putCounterOnEachOtherCreatureYouControl(ctx.Game, ctx.Item)
					}),
					Instruction(func(ctx *Context) error {
						n := 0
						for _, id := range permanentsControlledByMatching(ctx.Game, ctx.Controller(), Creature()) {
							if id != ctx.Item.SourceCardID {
								n++
							}
						}
						return GainLife{Amount: n}.Apply(ctx)
					}))),
		},
	})
}
