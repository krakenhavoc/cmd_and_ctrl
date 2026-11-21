package cardsg

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glamdring — Legendary Artifact — Equipment {2} (The Hobbit Commander):
//
//	"Equipped creature has first strike and gets +1/+0 for each instant and
//	 sorcery card in your graveyard.
//	 Whenever equipped creature deals combat damage to a player, you may cast
//	 an instant or sorcery spell from your hand with mana value less than or
//	 equal to that damage without paying its mana cost.
//	 Equip {3}"
//
// #4454, S76. The pump is PumpAttachedPer over Fandaniel's graveyard count,
// read on every recompute from the Glamdring controller's graveyard (the
// "your" of the text). The free cast is Buster Sword's offer: the damage
// dealt rides the trigger event, the candidates are the cards in hand with
// an instant or sorcery spell of at most that mana value (judged on the
// face being cast, #4509: an Adventure counts), and the cast is the engine's own,
// so its targets, modes and X are chosen as it goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52826984-44cb-48dc-a737-bb02983ffea8",
		Name:         "Glamdring",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("first strike"),
			effects.PumpAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
				return effects.FandanielInstantsAndSorceries(g, source.Controller)
			}),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Glamdring — you may cast an instant or sorcery spell from your hand with mana value at most that damage without paying its mana cost",
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					damage := effects.ThatDamage(ctx.Item)
					effects.OfferFreeHandCast(ctx, ctx.Controller(), glamdringFreeCast,
						"an instant or sorcery spell with mana value "+strconv.Itoa(damage)+" or less",
						effects.FreeInstantOrSorcery().WithManaValueAtMost(damage))
					return nil
				}),
			),
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

var glamdringFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/glamdring-cast", "Glamdring")
