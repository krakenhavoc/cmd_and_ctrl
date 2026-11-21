package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ian the Reckless — Legendary Creature — Human Warrior {1}{R}, 2/1
// (Fallout, #5948):
//
//	"Whenever Ian the Reckless attacks, if it's modified, you may have it
//	 deal damage equal to its power to you and any target. (Equipment, Auras
//	 you control, and counters are modifications.)"
//
// An attack trigger with an intervening "if" (CR 603.4): Ian is modified
// (CR 700.9, b07IsModified: a counter, an Equipment, an Aura his controller
// controls) as he attacks and again as the ability resolves. The "any target"
// is chosen as the ability goes on the stack (CR 603.3d); the "may" is asked
// as it resolves (CR 603.5), after which Ian deals damage equal to his
// power, read then (his last-known power if he has left), to his controller
// and to that target. If the target has become illegal the damage still goes
// to his controller, as the printed text is one sentence about two
// recipients (CR 608.2b).
//
// No simplification.
func init() {
	cond := effects.CondBoard{
		Name: "Ian the Reckless is modified",
		Test: func(ctx *effects.Context) bool {
			c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
			return ok && effects.OnBattlefield(ctx.Game, c.InstanceID) && effects.B07IsModified(ctx.Game, c)
		},
	}
	attack := effects.WheneverThisAttacks("Ian the Reckless — you may have it deal damage equal to its power to you and any target",
		ianTheRecklessDamage)
	when := attack.AppliesTo
	attack.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return (when == nil || when(ev, source, lki, g)) && effects.B07IsModified(g, *source)
	}
	effects.Register(effects.Spec{
		OracleID:     "0727ce6a-aae2-4ddc-9e6c-b18011c5bd37",
		Name:         "Ian the Reckless",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.InterveningIf(cond, attack), effects.TargetAny()),
				"Ian the Reckless — have it deal damage equal to its power to you and the target?"),
		},
	})
}

// ianTheRecklessDamage deals Ian's power to his controller and to the chosen
// target. Ian's power is read live, or from the departure record once he has gone.
func ianTheRecklessDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	power := 0
	if effects.OnBattlefield(g, item.SourceCardID) {
		power = effects.B31CurrentPowerOf(g, item.SourceCardID)
	} else if info, ok := g.LastKnownPermanentForEffect(item.SourceCardID); ok {
		power = info.Power
	}
	if power <= 0 {
		return nil
	}
	if err := (effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: power}).Apply(ctx); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: power}.Apply(ctx)
	}
	return nil
}
