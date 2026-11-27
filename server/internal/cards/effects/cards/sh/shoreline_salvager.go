package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shoreline Salvager — Creature — Surrakar {3}{B}, 3/3 (Worldwake, S82 OLD b488):
//
//	"Whenever this creature deals combat damage to a player, if you control an
//	 Island, you may draw a card."
//
// An intervening "if" (CR 603.4), Wandering Champion's shape: an Island you control
// is checked as the damage is dealt and again before the "you may" is asked, "Island"
// read through the card's text and the land's types through the layers (so a land turned
// into an Island counts). The draw is the optional one, asked as the ability resolves
// (CR 603.5).
//
// No simplification.
func init() {
	const label = "Shoreline Salvager — you may draw a card"
	islands := effects.CondBoard{
		Name: "you control an Island",
		Test: func(ctx *effects.Context) bool {
			island := ctx.TextSubtype("Island")
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.HasSubtype(island) {
					return true
				}
			}
			return false
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "02c08d0b-0d76-4533-a7e9-d40c7c4b3450",
		Name:         "Shoreline Salvager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.InterveningIf(islands, effects.WheneverThisDealsCombatDamageToAPlayer(label, effects.Do(effects.DrawCards{N: 1}))),
				"Shoreline Salvager — draw a card?", 1),
		},
	})
}
