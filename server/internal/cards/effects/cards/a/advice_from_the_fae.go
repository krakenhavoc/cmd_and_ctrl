package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Advice from the Fae — Sorcery {2/U}{2/U}{2/U} (Shadowmoor, S82 older sets):
//
//	"({2/U} can be paid with any two mana or with {U}. This card's mana value is
//	 6.)
//	 Look at the top five cards of your library. If you control more creatures
//	 than each other player, put two of those cards into your hand. Otherwise,
//	 put one of them into your hand. Then put the rest on the bottom of your
//	 library in any order."
//
// Dig Through Time's take (Max chosen as the spell resolves, the rest on the
// bottom in any order). "More creatures than each other player" is read as the
// spell resolves (CR 608.2h) against every other player still in the game, and
// a tie is not more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5063f015-bd1c-4e14-9888-63e202be2ac8",
		Name:         "Advice from the Fae",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2, Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			take := 1
			if adviceFromTheFaeControlsMoreCreatures(ctx.Game, player) {
				take = 2
			}
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Max:    take,
				Label:  "Advice from the Fae — put the chosen cards into your hand",
				Key:    effects.TakeRestBottomAny,
			}.Apply(ctx)
		},
	})
}

// adviceFromTheFaeControlsMoreCreatures reports whether player controls strictly
// more creatures than every other player still in the game.
func adviceFromTheFaeControlsMoreCreatures(g *game.Game, player uuid.UUID) bool {
	counts := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			counts[p.ID] = 0
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if _, in := counts[c.Controller]; in && c.IsCreature() {
			counts[c.Controller]++
		}
	}
	for id, n := range counts {
		if id != player && n >= counts[player] {
			return false
		}
	}
	return true
}
