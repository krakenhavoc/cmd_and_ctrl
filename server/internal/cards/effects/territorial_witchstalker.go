package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Territorial Witchstalker — Creature — Wolf {1}{G}, 2/3 (Wilds of
// Eldraine, #4471):
//
//	"Defender
//	 At the beginning of combat on your turn, if you control a creature
//	 with power 4 or greater, this creature gets +1/+0 until end of turn
//	 and can attack this turn as though it didn't have defender."
//
// Nasty Little Rabbit's ferocious condition (any creature you control,
// the Witchstalker included) as an intervening if, checked as combat
// begins and again as the trigger resolves (CR 603.4), the way Nessian
// Hornbeetle declares it. The body is a self pump and Hightide Hermit's
// waiver of defender for this turn (CR 702.3b): the Witchstalker keeps
// the keyword, it just may attack. A Witchstalker that has left the
// battlefield gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "97b9e3f1-7569-45e0-aef9-0d8d2845dbf1",
		Name:            "Territorial Witchstalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control a creature with power 4 or greater",
				Test: func(ctx *Context) bool {
					return b24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 4)
				},
			}, On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			}, "Territorial Witchstalker — +1/+0 until end of turn, and it can attack this turn as though it didn't have defender",
				territorialWitchstalkerUnleash)),
		},
	})
}

// territorialWitchstalkerUnleash is the pump, then the waiver.
func territorialWitchstalkerUnleash(g *game.Game, item *game.StackItem) error {
	if err := thisGetsUntilEndOfTurn(1, 0, "Territorial Witchstalker — +1/+0 until end of turn")(g, item); err != nil {
		return err
	}
	return CanAttackThisTurnAsThoughNoDefender("Territorial Witchstalker — can attack this turn as though it didn't have defender")(g, item)
}
