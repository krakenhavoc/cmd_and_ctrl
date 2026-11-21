package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladewing, Deathless Tyrant — Legendary Creature — Dragon Skeleton
// {5}{B}{R}, 6/6 (Dominaria United Commander, S82):
//
//	"Flying, haste
//	 Whenever Bladewing deals combat damage to a player or planeswalker,
//	 for each creature card in your graveyard, create a 2/2 black Zombie
//	 Knight creature token with menace."
//
// The creature cards are counted as the trigger resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cc4bc5d7-ed37-4a80-a5e7-eb68c255cbef",
		Name:            "Bladewing, Deathless Tyrant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Bladewing, Deathless Tyrant — for each creature card in your graveyard, create a 2/2 Zombie Knight with menace",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("2/2 black Zombie Knight with menace"),
					Name:     "the creature cards in your graveyard",
					Count: func(ctx *effects.Context) int {
						return effects.B11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
					},
				})),
		},
	})
}
