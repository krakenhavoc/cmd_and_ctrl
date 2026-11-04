package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warren Torchmaster — Creature — Goblin Warrior {1}{R}, 2/2:
//
//	"At the beginning of combat on your turn, you may blight 1. When you
//	 do, target creature gains haste until end of turn. (To blight 1, put
//	 a -1/-1 counter on a creature you control.)"
//
// The blight is asked as the trigger resolves (Blight, #4683). "When you
// do" is a reflexive trigger (CR 603.12), created once a creature was
// chosen and the counter has landed, so its target is picked as it goes on
// the stack, after the blight: the blighted creature itself may be the
// target if it survived.
func init() {
	Register(Spec{
		OracleID:     "20c3c8ab-e996-462e-a1f6-87421fc3ee55",
		Name:         "Warren Torchmaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Warren Torchmaster — you may blight 1; when you do, target creature gains haste", func(g *game.Game, item *game.StackItem) error {
				return Blight{N: 1, May: true, Then: warrenTorchmasterBlighted}.Apply(NewContext(g, item))
			}),
		},
	})
}

// warrenTorchmasterBlighted is "when you do, …": the reflexive trigger. An
// on-disk key.
var warrenTorchmasterBlighted = RegisterBlighted("blight/warren-torchmaster-when-you-do", func(ctx *Context, d BlightDone) error {
	if !d.Did() {
		return nil
	}
	return WhenYouDo("Warren Torchmaster — target creature gains haste until end of turn", warrenTorchmasterHasteBody).Apply(ctx)
})

// warrenTorchmasterHasteBody is the reflexive trigger's body: its target
// gains haste until end of turn. An on-disk key.
var warrenTorchmasterHasteBody = game.ReflexiveBody("warren-torchmaster/haste", simpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Warren Torchmaster — haste until end of turn"}.Apply(ctx)
	}
	return nil
}), constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
