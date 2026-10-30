package effects

import (
	"strconv"

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
	Register(Spec{
		OracleID:     "52826984-44cb-48dc-a737-bb02983ffea8",
		Name:         "Glamdring",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("first strike"),
			PumpAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
				return fandanielInstantsAndSorceries(g, source.Controller)
			}),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key: "Glamdring — you may cast an instant or sorcery spell from your hand with mana value at most that damage without paying its mana cost",
			Effect: Do(
				Instruction(func(ctx *Context) error {
					damage := thatDamage(ctx.Item)
					offerFreeHandCast(ctx, ctx.Controller(), glamdringFreeCast,
						"an instant or sorcery spell with mana value "+strconv.Itoa(damage)+" or less",
						freeInstantOrSorcery().withManaValueAtMost(damage))
					return nil
				}),
			),
		}},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

var glamdringFreeCast = registerFreeHandCastOffer("choose-cards/glamdring-cast", "Glamdring")
