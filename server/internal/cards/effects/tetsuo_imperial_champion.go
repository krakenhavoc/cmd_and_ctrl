package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tetsuo, Imperial Champion — Legendary Creature — Human Samurai {U}{B}{R},
// 3/3:
//
//	"Whenever Tetsuo attacks, if it's equipped, choose one —
//	 • Tetsuo deals damage equal to the greatest mana value among Equipment
//	   attached to it to any target.
//	 • You may cast an instant or sorcery spell from your hand with mana
//	   value less than or equal to the greatest mana value among Equipment
//	   attached to Tetsuo without paying its mana cost."
//
// The intervening "if it's equipped" (CR 603.4) is read as the ability
// triggers (an Equipment attached to Tetsuo on the battlefield) and again
// as it resolves, in each bullet's first step: an Equipment that was
// unattached or has left in the meantime no longer counts, and with none
// the ability does nothing. The greatest mana value is read as it resolves,
// from the Equipment attached to Tetsuo as he last existed (CR 608.2h,
// PermanentInfo.Attachments), so a Tetsuo that has left still deals the
// damage his Equipment's value says. The first bullet's target is any
// target, chosen as the ability goes on the stack with its mode. The second
// is Glamdring's offer with that value as the ceiling: the cast is the
// engine's own, so its targets, modes and X are chosen as it goes on the
// stack, and declining leaves the card in hand.
//
// No simplification.
func init() {
	t := On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !ThisAttacked(ev, source, game.Characteristic{}, g) {
			return false
		}
		_, equipped := dmuB19GreatestEquipmentValueNow(g, source)
		return equipped
	}, "Tetsuo, Imperial Champion — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeSteps("Tetsuo deals damage equal to the greatest mana value among Equipment attached to it to any target.", TargetAny(),
			Instruction(func(ctx *Context) error {
				value, equipped := dmuB19GreatestEquipmentValue(ctx.Game, ctx.Source())
				if !equipped {
					return nil
				}
				target, ok := ModeTarget(ctx, ctx.Occurrence())
				if !ok {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: target.ID, Amount: value}.Apply(ctx)
			})),
		ModeSteps("You may cast an instant or sorcery spell from your hand with mana value less than or equal to the greatest mana value among Equipment attached to Tetsuo without paying its mana cost.", nil,
			Instruction(func(ctx *Context) error {
				value, equipped := dmuB19GreatestEquipmentValue(ctx.Game, ctx.Source())
				if !equipped {
					return nil
				}
				offerFreeHandCast(ctx, ctx.Controller(), tetsuoFreeCast,
					"an instant or sorcery spell with mana value "+strconv.Itoa(value)+" or less",
					freeInstantOrSorcery().withManaValueAtMost(value))
				return nil
			})),
	)
	Register(Spec{
		OracleID:     "1111e042-0bce-4819-ab42-9912aed4d930",
		Name:         "Tetsuo, Imperial Champion",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

var tetsuoFreeCast = registerFreeHandCastOffer("choose-cards/tetsuo-imperial-champion-cast", "Tetsuo, Imperial Champion")
