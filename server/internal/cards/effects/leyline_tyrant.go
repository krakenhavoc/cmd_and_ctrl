package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline Tyrant — Creature — Dragon {2}{R}{R}, 4/4:
//
//	"Flying
//	 You don't lose unspent red mana as steps and phases end.
//	 When this creature dies, you may pay any amount of {R}. When you
//	 do, it deals that much damage to any target."
//
// The mana clause is Spec.ManaPool (#2166): red only, derived from the
// battlefield, so it ends the moment the Tyrant leaves.
//
// The dies trigger is "pay any amount of {R}" (MayPayX with Symbol, #3172)
// followed by a CR 603.12 reflexive trigger. X is chosen as the trigger
// resolves, from 0 to the red the controller can produce right now, and
// is then paid through the ordinary "you may pay" prompt, so the pool,
// the auto-tapper and the Always / Never settings all apply. Paying
// nothing is the same as declining: no reflexive trigger and no damage.
// The trigger the payment creates goes on the stack above the dies
// trigger and picks its "any target" there (CR 603.3d), carrying X as its
// amount; the damage is dealt by the Tyrant as it last existed.
func init() {
	Register(Spec{
		OracleID:        "f92aaa00-6ece-4033-b3df-b2fc2c4718d9",
		Name:            "Leyline Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaPool: []game.ManaPoolStatic{
			{Kind: game.ManaPoolKeep, Colors: []string{"R"}},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Leyline Tyrant — pay any amount of {R}; it deals that much damage to any target", leylineTyrantPay),
		},
	})
}

// leylineTyrantPay is the dies trigger's body.
func leylineTyrantPay(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Symbol: "{R}",
		Label:  "Leyline Tyrant",
		Buys:   "deal X damage to any target",
		Goal:   XAsHighAsYouCan,
		Unit:   game.PayAmountDamage,
		OnPay: func(ctx *Context, x int) error {
			if x <= 0 {
				return nil
			}
			return ReflexiveTrigger{
				Label:  "Leyline Tyrant — damage equal to the {R} paid to any target",
				Body:   leylineTyrantDamageBody,
				Params: game.EffectParams{Amount: x},
			}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
