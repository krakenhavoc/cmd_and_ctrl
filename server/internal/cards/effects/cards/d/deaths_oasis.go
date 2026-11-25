package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death's Oasis — Enchantment {W}{B}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever a nontoken creature you control dies, mill two cards. Then
//	 return a creature card with lesser mana value than the creature that
//	 died from your graveyard to your hand.
//	 {1}, Sacrifice this enchantment: You gain life equal to the greatest
//	 mana value among creatures you control."
//
// Midnight Reaper's trigger (the nontoken creature's last-known controller,
// CR 603.10a). The creature's mana value is read off the trigger as it last
// existed (TriggerContext.Object, Orah's "lesser" ceiling) and rides the mill
// as the tail's number; the return is not targeted, so it is chosen after the
// mill, from the graveyard as it then is (the dead creature and the milled
// cards among it, none of them lesser than itself). The life gain reads the
// greatest mana value as the ability resolves, the Enchantment already
// sacrificed as the cost.
//
// No simplification.
func init() {
	const label = "Death's Oasis — mill two cards, then return a creature card with lesser mana value than the creature that died from your graveyard to your hand"
	trigger := effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		dead, ok := effects.DiedCreature(ev, g)
		return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && !effects.IsToken(dead)
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		mv := 0
		if obj := ctx.Trigger().Object; obj != nil {
			mv = obj.ManaValue
		}
		return effects.S82OldB117MillThenReturnLesserCreature(g, item.Controller, ctx.Source(), 2, mv)
	})
	effects.Register(effects.Spec{
		OracleID:     "b10f6915-96b6-4e87-8171-4d3e96bfd48c",
		Name:         "Death's Oasis",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.TriggerWithPurpose(trigger, game.Purpose{DeathPayoff: true})},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this enchantment: You gain life equal to the greatest mana value among creatures you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.GapB153GreatestManaValueAmongYourCreatures(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return effects.GainLife{Player: item.Controller, Amount: n}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
