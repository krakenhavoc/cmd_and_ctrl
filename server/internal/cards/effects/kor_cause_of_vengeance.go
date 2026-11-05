package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kor, Cause of Vengeance — Legendary Creature — Klingon Warrior Officer
// {1}{R}{W}{B}, 4/4 (Star Trek Commander, #4974):
//
//	"Whenever you attack, you may sacrifice another creature. If you do,
//	 draw a card, then Klingons and Warriors you control get +2/+2 until
//	 end of turn."
//
// One trigger per attack declaration (trekB40YouAttacked). The sacrifice
// is asked as the trigger resolves, over the other creatures you control
// then (maySacrificeOne, CR 608.2c); declining, or having nothing to
// sacrifice, does nothing else. "If you do" follows only a creature that
// really left. The draw comes first and the pump waits for any question
// the draw asks (a draw replacement's prompt); the pump's set is fixed as
// it applies (CR 611.2c), every creature you control that is a Klingon or
// a Warrior then, Kor included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc8d8c51-428c-4b2f-965d-c43ec94997c3",
		Name:         "Kor, Cause of Vengeance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			trekB40YouAttacked("Kor, Cause of Vengeance — you may sacrifice another creature; if you do, draw a card, then Klingons and Warriors you control get +2/+2",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Kor, Cause of Vengeance — sacrifice another creature to draw a card and give Klingons and Warriors you control +2/+2?",
						Creature(), true, korCauseOfVengeanceSacrificed)
				}),
		},
	})
}

// korCauseOfVengeanceSacrificed is "if you do": the draw, then the pump.
// Keyed: an on-disk identity, never renamed or reused.
var korCauseOfVengeanceSacrificed = RegisterSacrificedOne("prompt-run/kor-cause-of-vengeance-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return korCauseOfVengeanceDrawThenPump.Run(ctx)
})

// korCauseOfVengeanceDrawThenPump is "draw a card, then Klingons and
// Warriors you control get +2/+2 until end of turn".
var korCauseOfVengeanceDrawThenPump = RegisterThenSteps("then-steps/kor-cause-of-vengeance-draw-then-pump",
	DrawCards{N: 1},
	Instruction(func(ctx *Context) error {
		return BoostUntilEOT{
			Match:     And(Creature(), YouControl(), AnySubtype(ctx.TextSubtypes("Klingon", "Warrior")...)),
			Power:     2,
			Toughness: 2,
			Label:     "Kor, Cause of Vengeance — Klingons and Warriors you control get +2/+2",
		}.Apply(ctx)
	}),
)
