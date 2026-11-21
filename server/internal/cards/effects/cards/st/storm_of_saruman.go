package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm of Saruman — Enchantment {4}{U}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Ward {3}
//	 Whenever you cast your second spell each turn, copy it, except the copy
//	 isn't legendary. You may choose new targets for the copy. (A copy of a
//	 permanent spell becomes a token.)"
//
// The trigger is Shiko and Narset's second-spell-each-turn watcher
// (YouCastYourSecondSpellEachTurn); the copy is The Clone Saga's: a copy of
// the spell as it was cast (CR 707.10), read from last-known information in
// case the spell has already resolved or been countered, with the legendary
// supertype removed from its copiable values (CR 707.10a, 707.9b). The copy is
// made on the stack above the original and the controller may choose new
// targets. A permanent spell's copy becomes a token as it resolves (CR 707.10,
// 608.3f). Ward {3} is the shared ward trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf5f4860-e805-46a3-9352-a2c583e33403",
		Name:         "Storm of Saruman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Storm of Saruman — ward {3}"),
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Storm of Saruman — copy that spell, except the copy isn't legendary; you may choose new targets for the copy",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.CopySpell{
						StackID:          item.Trigger.Event.CardID,
						Controller:       item.Controller,
						ChooseNewTargets: true,
						FromLastKnown:    true,
						Except: func(v *game.PrintedValues) {
							v.RemoveSupertype("Legendary")
						},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
