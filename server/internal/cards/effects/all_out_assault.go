package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// All-Out Assault — Enchantment {2}{R}{W}{B} (Tarkir: Dragonstorm, #4463):
//
//	"Creatures you control get +1/+1 and have deathtouch.
//	 When this enchantment enters, if it's your main phase, there is an
//	 additional combat phase after this phase followed by an additional
//	 main phase. When you next attack this turn, untap each creature you
//	 control."
//
// The anthem is two continuous effects of one static ability: +1/+1 in
// layer 7c and deathtouch in layer 6, over every creature you control,
// including ones that arrive later (CR 611.3).
//
// The enters trigger has an intervening "if" (CR 603.4): it triggers only
// in its controller's main phase and does nothing if that is no longer
// true when it resolves. Both sentences are its effect. The phases are
// Relentless Assault's ExtraCombatAndMainAfterThisMain (CR 500.8), and the
// last sentence creates a delayed trigger (CR 603.7b) that fires the next
// time you declare attackers this turn, once however many creatures
// attack, and untaps every creature you control. It lasts until the turn
// ends whether or not you attack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de8c0e2d-3207-4bc6-83ef-91e33b26a981",
		Name:         "All-Out Assault",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			allOutAssaultLabel(b16Anthem(b16CreaturesYouControl, 1, 1), "Creatures you control get +1/+1."),
			allOutAssaultLabel(b16GrantKeywords(b16CreaturesYouControl, "deathtouch"), "Creatures you control have deathtouch."),
		},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{Name: "it's your main phase", Test: func(ctx *Context) bool {
				return allOutAssaultYourMainPhase(ctx.Game, ctx.Controller())
			}}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && allOutAssaultYourMainPhase(g, source.Controller)
			}, "All-Out Assault — an additional combat phase and main phase; when you next attack this turn, untap each creature you control",
				Do(ExtraCombatAndMainAfterThisMain(), Instruction(func(ctx *Context) error {
					return DelayedOnEvent{
						Label:        "All-Out Assault — untap each creature you control",
						On:           []game.EventKind{game.EventAttack},
						Condition:    youAttackCondition,
						Body:         allOutAssaultUntapBody,
						OncePerBatch: true,
					}.Apply(ctx)
				})))),
		},
	})
}

// allOutAssaultUntapBody is the delayed trigger's body. An on-disk key.
var allOutAssaultUntapBody = game.SimpleDelayedBody("all-out-assault/untap-each-creature-you-control",
	func(g *game.Game, item *game.StackItem) error {
		return UntapAllCreaturesYouControl{}.Apply(NewContext(g, item))
	})

// allOutAssaultYourMainPhase is "it's your main phase": the controller's
// turn, in a main phase (an added one included).
func allOutAssaultYourMainPhase(g *game.Game, controller uuid.UUID) bool {
	step := g.Turn.Step
	return isActivePlayer(g, controller) && (step == game.StepPrecombatMain || step == game.StepPostcombatMain)
}

// allOutAssaultLabel sets a static's label to its printed half.
func allOutAssaultLabel(s game.StaticAbility, label string) game.StaticAbility {
	s.Label = label
	return s
}
