package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Speed, Young Avenger — {1}{R} Legendary Creature — Mutant Hero 2/2
// (Marvel Super Heroes, #4455):
//
//	"Haste
//	 Whenever you cast a noncreature spell, you may pay {1}. When you do,
//	 target creature with haste can't be blocked this turn except by
//	 creatures with haste."
//
// Ant-Man, Colony Commander's may-pay and reflexive trigger (CR 603.12):
// the target is chosen only if the {1} was paid. The block rule is a
// cantBeBlockedExceptBy record on the target (CantBeBlockedThisTurnExceptBy),
// pinned to the object for the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2fb0f2d2-ed1f-4181-91f2-684a71c34998",
		Name:            "Speed, Young Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), speedPayLabel, speedPayOffer),
		},
	})
}

const speedPayLabel = "Speed, Young Avenger — pay {1} so target creature with haste can't be blocked except by creatures with haste"

func speedPayOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Speed, Young Avenger — pay {1} so target creature with haste can't be blocked this turn except by creatures with haste?",
		Then:     speedPaid,
	}.Apply(NewContext(g, item))
}

var speedPaid = PayUnlessBranches("pay-unless/speed-young-avenger-unblockable",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Speed, Young Avenger — target creature with haste can't be blocked except by creatures with haste", speedBody).Apply(ctx)
	}, nil)

var speedBody = game.ReflexiveBody("speed-young-avenger/unblockable", simpleBody(speedEvade),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target creature with haste", HasKeyword("haste"))
	}))

func speedEvade(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return CantBeBlockedThisTurnExceptBy{
		Target:   id,
		Keywords: []string{"haste"},
		Text:     "creatures with haste",
		Label:    "Speed, Young Avenger — can't be blocked except by creatures with haste",
	}.Apply(ctx)
}
