package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Itzquinth, Firstborn of Gishath — Legendary Creature — Dinosaur {R}{G},
// 2/3 (The Lost Caverns of Ixalan, #4470):
//
//	"Haste
//	 When Itzquinth enters, you may pay {2}. When you do, target Dinosaur
//	 you control deals damage equal to its power to another target
//	 creature."
//
// Snaremaster Sprite's "you may pay. When you do" shape: MayPay asks as
// the enters trigger resolves, and only a payment queues the reflexive
// trigger (CR 603.12), which then picks its two targets. "Another" is
// the second clause's Distinct: the creature dealt damage differs from
// the Dinosaur dealing it (Itzquinth itself may be either). The body is
// Bite Down's one-sided bite (oneSidedBite), reading the Dinosaur's power
// as the reflexive trigger resolves. The Dinosaur type is matched through
// OfSubtype's text lens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f976a244-1240-4430-9182-bb8d6d932ea0",
		Name:            "Itzquinth, Firstborn of Gishath",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Itzquinth, Firstborn of Gishath — you may pay {2}", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Itzquinth — pay {2} to have target Dinosaur you control deal damage equal to its power to another target creature?",
					Then:     itzquinthPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

var itzquinthPaid = PayUnlessBranches("pay-unless/itzquinth-dinosaur-bites",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Itzquinth, Firstborn of Gishath — target Dinosaur you control deals damage equal to its power to another target creature",
			itzquinthBiteBody).Apply(ctx)
	}, nil)

var itzquinthBiteBody = game.ReflexiveBody("itzquinth/dinosaur-bites",
	simpleBody(func(g *game.Game, item *game.StackItem) error { return oneSidedBite(item, NewContext(g, item)) }),
	constTargets(func() *game.TargetSpec {
		return Clauses(
			TargetCreature("target Dinosaur you control", YouControl(), OfSubtype("Dinosaur")),
			Distinct(TargetCreature("another target creature")),
		)
	}))
