package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Cabbage Merchant — Legendary Creature — Human Citizen {2}{G},
// 2/2 (EDHREC rank 2357):
//
//	"Whenever an opponent casts a noncreature spell, create a Food
//	 token. (It's an artifact with "{2}, {T}, Sacrifice this token:
//	 You gain 3 life.")
//	 Whenever a creature deals combat damage to you, sacrifice a Food
//	 token.
//	 Tap two untapped Foods you control: Add one mana of any color."
//
// The Food maker that punishes the table for playing spells. The
// cast trigger is Sunscorch Regent's condition narrowed to a
// noncreature spell, read off the stack where its type line is
// intact. The combat-damage trigger is No Mercy's condition with
// "combat" added, and "sacrifice a Food token" is the controller's
// own choice among their Food tokens through the sacrifice prompt —
// a Gingerbread Cabin is a Food but not a token, so it is not
// offered; with no Food token the trigger does nothing.
//
// The mana ability is Heritage Druid's shape: a TapOthers cost of two
// untapped Foods, no {T} of its own (the Merchant is not tapped), and
// the any-colour pipe slot. It is a mana ability (no target, produces
// mana), so it skips the stack.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e31808bb-fb6e-487b-a973-db44961c84ee",
		Name:         "The Cabbage Merchant",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent("two untapped Foods you control", effects.Subtype("Food")),
				Label:  "two untapped Foods you control",
			}},
			Produced: "{W|U|B|R|G}",
			Label:    "Tap two untapped Foods you control: Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B15OpponentCastSpell(ev, source) {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.IsCreature()
			}, "The Cabbage Merchant — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !ev.Combat {
					return false
				}
				return effects.B13CreatureDealtDamageToYou(ev, source, g)
			}, "The Cabbage Merchant — sacrifice a Food token", func(g *game.Game, item *game.StackItem) error {
				g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
					effects.SacrificeSpec("a Food token", effects.Subtype("Food"), effects.IsTokenPredicate()),
					"The Cabbage Merchant — sacrifice a Food token")
				return nil
			}),
		},
	})
}
