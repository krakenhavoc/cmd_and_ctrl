package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conquest Mission — Enchantment {2}{W} (Star Trek, #4973):
//
//	"Whenever a nontoken artifact you control enters, put a quest
//	 counter on this enchantment. Create a 1/1 colorless Borg artifact
//	 creature token. If this enchantment has six or more quest counters
//	 on it, create two of those tokens instead."
//
// One trigger, three sentences in order (CR 608.2c): the counter first,
// so the sixth artifact's own counter already counts toward "six or
// more". The Borg tokens it makes are tokens and so do not trigger it.
// An enchantment that left before the trigger resolved gets no counter
// and is read as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "35906166-71d1-4089-bbe4-4055edb81fb0",
		Name:         "Conquest Mission",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b31NontokenArtifactYouControlEntered(ev, source, g)
			}, "Conquest Mission — put a quest counter on it and create a Borg token", Do(
				Instruction(func(ctx *Context) error {
					if !onBattlefield(ctx.Game, ctx.Item.SourceCardID) {
						return nil
					}
					return AddCounter{Target: ctx.Item.SourceCardID, Kind: "quest", N: 1}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					n := 1
					if trekB1CountersNowOrLastKnown(ctx.Game, ctx.Item.SourceCardID, "quest") >= 6 {
						n = 2
					}
					return CreateToken{Template: TokenCard("1/1 colorless Borg artifact"), N: n}.Apply(ctx)
				}),
			)),
		},
	})
}
