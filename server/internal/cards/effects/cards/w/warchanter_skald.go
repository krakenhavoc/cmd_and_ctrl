package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warchanter Skald — Creature — Dwarf Cleric {2}{W}, 2/3:
//
//	"Whenever this creature becomes tapped, if it's enchanted or equipped,
//	 create a 2/1 red Dwarf Berserker creature token."
//
// Moonfolk Puzzlemaker's "becomes tapped" (the tap event, and an attack that
// taps it; a vigilance attack is not one) with an intervening "if" (CR 603.4):
// an Aura or Equipment is attached to the Skald as it becomes tapped and again
// as the trigger resolves, so removing it in response fizzles the token.
//
// No simplification.
func init() {
	cond := effects.CondBoard{
		Name: "it's enchanted or equipped",
		Test: func(ctx *effects.Context) bool {
			return effects.RprB03IsEnchantedOrEquipped(ctx.Game, ctx.Item.SourceCardID)
		},
	}
	trigger := effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			return effects.ThisBecameTapped(ev, source, lki, g) && effects.RprB03IsEnchantedOrEquipped(g, source.InstanceID)
		},
		"Warchanter Skald — create a 2/1 red Dwarf Berserker creature token",
		effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 red Dwarf Berserker"), N: 1}))
	effects.Register(effects.Spec{
		OracleID:     "dc7e259a-d965-4e25-8145-81ab3087f718",
		Name:         "Warchanter Skald",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered:    []game.TriggeredAbility{effects.InterveningIf(cond, trigger)},
	})
}
