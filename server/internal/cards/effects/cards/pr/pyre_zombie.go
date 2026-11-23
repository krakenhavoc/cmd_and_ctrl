package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyre Zombie — Creature — Zombie {1}{B}{R}, 2/1 (Dominaria Remastered):
//
//	"At the beginning of your upkeep, if this card is in your graveyard,
//	 you may pay {1}{B}{B}. If you do, return it to your hand.
//	 {1}{R}{R}, Sacrifice this creature: It deals 2 damage to any target."
//
// The upkeep trigger is Master of Death's with mana instead of life
// (InGraveyard, MayPay): "you" is the card's owner while it is in the
// graveyard (CR 108.4), the intervening "if" is the zone (CR 603.4) and is
// checked again as it resolves, and the card returns only once {1}{B}{B}
// is paid. The sacrifice ability is Ember Hauler's, over {1}{R}{R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "accb82be-4f90-4d9a-bafd-f8de3622b3a3",
		Name:         "Pyre Zombie",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.AtYourUpkeep("Pyre Zombie — you may pay {1}{B}{B} to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{B}{B}",
						Question: "Pyre Zombie — pay {1}{B}{B} to return it to your hand?",
						Then:     pyreZombiePaid,
					}.Apply(effects.NewContext(g, item))
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}, Sacrifice this creature: It deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.B33DamageChosenTargetFromSource(2),
		}},
	})
}

// pyreZombiePaid is "return it to your hand", once the mana is paid. An
// on-disk key: never renamed, never reused.
var pyreZombiePaid = effects.PayUnlessBranches("pay-unless/pyre-zombie-return",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReturnThisCardFromYourGraveyardToHand(ctx.Game, ctx.Item)
	}, nil)
