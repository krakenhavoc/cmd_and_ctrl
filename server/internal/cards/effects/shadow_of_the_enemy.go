package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadow of the Enemy — Sorcery {3}{B}{B}{B} (The Hobbit Commander):
//
//	"Exile all creature cards from target player's graveyard. You may cast
//	 spells from among those cards for as long as they remain exiled, and
//	 mana of any type can be spent to cast them."
//
// Crypt Incursion's exile of the creature cards in one graveyard, with
// Hostage Taker's permission over each card that lands: the caster may CAST
// it (a creature card is never a land, so "cast spells" loses nothing),
// for as long as it stays in exile (CR 611.2b), paying with mana of any
// type (CR 106.8; AnyType implies any colour, so a {C} in the cost is
// payable too). The permission is granted to the spell's controller over
// every card that actually reached exile, all as one event, so a commander
// that took the command zone instead is not given one. A card cast this way
// has left exile, and the rest keep their permission. The target player may
// be the caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f0c72ff-cd93-46a6-8649-2cd9ce8b8625",
		Name:         "Shadow of the Enemy",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := TargetedPlayer(ctx)
			if p := ctx.PlayerByID(player); p == nil || p.Eliminated {
				return nil
			}
			ids := b42CardsInGraveyardOf(ctx.Game, player, func(c game.Card) bool { return c.IsCreature() })
			if len(ids) == 0 {
				return nil
			}
			grant := ExileWithPermission{
				GrantTo:     item.Controller,
				CastOnly:    true,
				AnyType:     true,
				WhileExiled: true,
			}.permission()
			return exileAllWithPermission(ctx.Game, ids, grant)
		},
	})
}
