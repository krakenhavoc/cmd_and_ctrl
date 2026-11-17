package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "cc4bc5d7-ed37-4a80-a5e7-eb68c255cbef",
		Name:            "Bladewing, Deathless Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				"Bladewing, Deathless Tyrant — for each creature card in your graveyard, create a 2/2 Zombie Knight with menace",
				Do(CreateTokensForEach{
					Template: TokenCard("2/2 black Zombie Knight with menace"),
					Name:     "the creature cards in your graveyard",
					Count: func(ctx *Context) int {
						return b11CreatureCardsInGraveyard(ctx.Game, ctx.Controller())
					},
				})),
		},
	})
}
