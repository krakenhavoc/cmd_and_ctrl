package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Labyrinth Adversary — Creature — Minotaur {3}{R}, 4/3 (Assassin's Creed,
// #5948):
//
//	"Trample
//	 Whenever you attack, you may pay {1}{R}. When you do, target creature
//	 can't block this turn."
//
// "Whenever you attack" is one trigger per attack declaration (Adeline's
// OncePerBatch over attackDeclaredByYou). The payment is Rodolf
// Duskbringer's MayPay; paying it raises the CR 603.12 reflexive trigger,
// whose target (any creature) is chosen when it goes on the stack, after the
// mana is spent, and which gives that creature "can't block" until end of
// turn (Ahn-Crop Crasher's body).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "00a9bdd0-8a39-4489-bd42-2ac92e550781",
		Name:            "Labyrinth Adversary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Labyrinth Adversary — you may pay {1}{R}; when you do, target creature can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{R}",
						Question: "Labyrinth Adversary — pay {1}{R}? If you do, target creature can't block this turn.",
						Then:     labyrinthAdversaryPaid,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// labyrinthAdversaryPaid is the paid branch, keyed: the reflexive trigger.
var labyrinthAdversaryPaid = PayUnlessBranches("pay-unless/labyrinth-adversary-cant-block",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Labyrinth Adversary — target creature can't block this turn",
			Body:  labyrinthAdversaryCantBlockBody,
		}.Apply(ctx)
	}, nil)

// labyrinthAdversaryCantBlockBody is "target creature can't block this turn".
// An on-disk key.
var labyrinthAdversaryCantBlockBody = game.ReflexiveBody("labyrinth-adversary/cant-block",
	simpleBody(targetCreatureCantBlockThisTurn("Labyrinth Adversary — target creature can't block this turn")),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
