package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thousand Moons Crackshot — Creature — Human Soldier {1}{W}, 2/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever this creature attacks, you may pay {2}{W}. When you do, tap
//	 target creature."
//
// Passenger Ferry's MayPay / CR 603.12 pair: the {2}{W} is asked for as
// the attack trigger resolves (CR 118.12), and only a paid offer raises
// the reflexive trigger, whose target is chosen as it goes on the stack,
// after the mana is spent. Declining, or being unable to pay, does
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1ce2971-fffa-40b3-9bfd-29afdb929e74",
		Name:         "Thousand Moons Crackshot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Thousand Moons Crackshot — you may pay {2}{W}; when you do, tap target creature",
				thousandMoonsCrackshotOffer),
		},
	})
}

// thousandMoonsCrackshotOffer asks for the {2}{W}; the reflexive trigger
// lives in the paid branch.
func thousandMoonsCrackshotOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}{W}",
		Question: "Thousand Moons Crackshot — pay {2}{W} to tap target creature?",
		Then:     thousandMoonsCrackshotPaid,
	}.Apply(NewContext(g, item))
}

// thousandMoonsCrackshotPaid is the paid half, keyed: the reflexive trigger
// is queued against the asking item.
var thousandMoonsCrackshotPaid = PayUnlessBranches("pay-unless/thousand-moons-crackshot-tap",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Thousand Moons Crackshot — tap target creature", thousandMoonsCrackshotTapBody).Apply(ctx)
	}, nil)

// thousandMoonsCrackshotTapBody is "when you do, tap target creature".
var thousandMoonsCrackshotTapBody = game.ReflexiveBody("thousand-moons-crackshot/tap", simpleBody(thousandMoonsCrackshotTap),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

func thousandMoonsCrackshotTap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return TapTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}
