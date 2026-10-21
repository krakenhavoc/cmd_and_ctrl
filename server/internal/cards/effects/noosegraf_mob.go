package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// noosegrafMobRemoved is 1 once the counter came off.
const noosegrafMobRemoved = "noosegraf-mob/removed"

// Noosegraf Mob — Creature — Zombie {4}{B}{B}, 0/0:
//
//	"This creature enters with five +1/+1 counters on it.
//	 Whenever a player casts a spell, remove a +1/+1 counter from this
//	 creature. If you do, create a 2/2 black Zombie creature token."
//
// The counters are an entry replacement (b10EntersWithCounters, CR
// 614.1c). The cast trigger is anyone's spell; "if you do" means the
// Zombie comes only when a counter really came off, so a Mob that has
// none left (or has left the battlefield) makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f27fb53f-a983-410a-821b-e48cd8c01f2e",
		Name:         "Noosegraf Mob",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 5, "Noosegraf Mob: enters with five +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AnyPlayer, "Noosegraf Mob — remove a +1/+1 counter; if you do, create a 2/2 Zombie",
				// The Zombie waits for a prompt the counter removal asks (a CR 616
				// order), and only comes when a counter really came off ("if you do";
				// CR 608.2c, #4155).
				Do(
					Instruction(func(ctx *Context) error {
						item := ctx.Item
						ctx.RecordStep(noosegrafMobRemoved, nil, 0)
						if ctx.isNewSourceObject(item.SourceCardID) || !onBattlefield(ctx.Game, item.SourceCardID) {
							return nil
						}
						c, ok := ctx.Game.LookupCardForEffect(item.SourceCardID)
						if !ok || c.Counters[game.CounterPlusOne] <= 0 {
							return nil
						}
						if err := ctx.Game.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, -1); err != nil {
							return err
						}
						ctx.RecordStep(noosegrafMobRemoved, nil, 1)
						return nil
					}),
					Instruction(func(ctx *Context) error {
						if ctx.StepRecord(noosegrafMobRemoved).N() != 1 {
							return nil
						}
						return CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
					}),
				)),
		},
	})
}
