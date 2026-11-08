package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kutzil's Flanker — Creature — Cat Warrior {2}{W}, 3/1 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Flash
//	 When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature for each creature that left
//	   the battlefield under your control this turn.
//	 • You gain 2 life and scry 2.
//	 • Exile target player's graveyard."
//
// Charming Prince's modal enters trigger. The first bullet counts the
// per-player CreaturesLeft tally (Tale of Momo's: every creature you
// controlled that left the battlefield this turn, by any route and to any
// zone, tokens included) as the trigger resolves, and puts the counters
// only on the Flanker that entered (CounterOnThis). The third is Bojuka
// Bog's graveyard exile for the bullet's own target.
//
// No simplification.
func init() {
	flanker := WhenThisEnters("Kutzil's Flanker — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	flanker.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on this creature for each creature that left the battlefield under your control this turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				n := ctx.Game.TurnTallyFor(item.Controller).CreaturesLeft
				if n <= 0 {
					return nil
				}
				return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
			}),
		ModeWithPurpose(ModeDoing("You gain 2 life and scry 2.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			if err := (GainLife{Player: item.Controller, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return Scry{Player: item.Controller, N: 2}.Apply(ctx)
		}), game.Purpose{Scry: 2}),
		ModeDoing("Exile target player's graveyard.", TargetPlayer("target player"),
			func(item *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return exileGraveyardForEffect(ctx.Game, item, t.ID)
			}),
	)
	Register(Spec{
		OracleID:        "ba557e19-a409-4c02-a9ac-a8fb77762e43",
		Name:            "Kutzil's Flanker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{flanker},
	})
}
