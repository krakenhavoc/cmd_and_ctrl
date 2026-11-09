package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yathan Roadwatcher — Creature — Human Scout {1}{W}{B}{G}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, if you cast it, mill four cards. When you
//	 do, return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield."
//
// Primal Witchstalker's shape with the intervening "if you cast it" (CR
// 603.4, Crystalline Entity's condition, read as the trigger would
// fire and again as it resolves). The reflexive trigger (CR 603.12) is
// made only if cards were milled, so an empty library does nothing, and
// its target is chosen as it goes on the stack, from the graveyard the mill
// just filled; a target that is gone or no longer a creature card with
// mana value 3 or less by then is illegal and the trigger does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c3d86b6-5dcd-4016-9b4a-d4f66b1e963f",
		Name:         "Yathan Roadwatcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, crestingMosasaurusYouCastIt),
				"Yathan Roadwatcher — mill four cards",
				Do(
					Instruction(func(ctx *Context) error {
						p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
						if p == nil || p.Library == nil || p.Library.Size() == 0 {
							return nil
						}
						ctx.RecordStep(yathanRoadwatcherRecord, nil, 1)
						return MillCards{N: 4}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						if ctx.StepRecord(yathanRoadwatcherRecord).N() == 0 {
							return nil
						}
						return WhenYouDo("Yathan Roadwatcher — return target creature card with mana value 3 or less from your graveyard to the battlefield",
							yathanRoadwatcherReturnBody).Apply(ctx)
					}),
				)),
		},
	})
}

const yathanRoadwatcherRecord = "yathan-roadwatcher/milled"

var yathanRoadwatcherReturnBody = game.ReflexiveBody("yathan-roadwatcher/return",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
		}
		return nil
	},
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
			Creature(), ManaValueLE(3), YouOwn())
	}))
