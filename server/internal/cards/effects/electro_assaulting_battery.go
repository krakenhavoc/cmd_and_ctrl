package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electro, Assaulting Battery — {1}{R}{R} Legendary Creature — Human
// Villain 2/3 (Marvel's Spider-Man promos, #4460):
//
//	"Flying
//	 You don't lose unspent red mana as steps and phases end.
//	 Whenever you cast an instant or sorcery spell, add {R}.
//	 When Electro leaves the battlefield, you may pay {X}. When you do,
//	 he deals X damage to target player."
//
// The mana clause is Leyline Tyrant's Spec.ManaPool (red only, gone the
// moment Electro leaves). The cast trigger is Archmage of Runes's
// trigger with Birgi's AddMana body; it uses the stack, since it does
// not trigger from a mana ability (CR 605.1b).
//
// The leaves trigger is Leyline Tyrant's "you may pay X, when you do"
// with a generic {X}: X is chosen as the trigger resolves, paid through
// the ordinary "you may pay" prompt, and paying 0 is declining. The
// CR 603.12 reflexive trigger picks its target player as it goes on the
// stack and carries X; Electro, as he last existed, deals the damage
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3db16fb8-5650-4568-b975-42ad3d44ebcd",
		Name:            "Electro, Assaulting Battery",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaPool: []game.ManaPoolStatic{
			{Kind: game.ManaPoolKeep, Colors: []string{"R"}},
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Electro, Assaulting Battery — add {R}", Do(AddMana{Produced: "{R}"})),
			WhenThisLeaves("Electro, Assaulting Battery — you may pay {X}; he deals X damage to target player",
				electroPay),
		},
	})
}

// electroPaid is the paid half, keyed: the X paid rides the prompt.
var electroPaid = PayUnlessBranches("pay-unless/electro-assaulting-battery-damage",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Electro, Assaulting Battery — X damage to target player",
			Body:   electroDamageBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

// electroDamageBody is the reflexive trigger: Params.Amount damage to the
// chosen player.
var electroDamageBody = game.ReflexiveBody("electro-assaulting-battery/damage", breechesBlastEffect,
	constTargets(func() *game.TargetSpec { return TargetPlayer("target player") }))

func electroPay(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}",
		Label: "Electro, Assaulting Battery",
		Buys:  "deal X damage to target player",
		Goal:  XAsHighAsYouCan,
		Unit:  game.PayAmountDamage,
		Then:  electroPaid,
	}.Apply(NewContext(g, item))
}
