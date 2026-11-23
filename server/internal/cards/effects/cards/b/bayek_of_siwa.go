package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bayek of Siwa — Legendary Creature — Human Assassin {3}{R}{W}, 3/4 (Assassin's
// Creed, #5948):
//
//	"Double strike
//	 During your turn, other historic creatures you control have double
//	 strike.
//	 Disguise {1}{R}{W} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise cost.)"
//
// Anara, Wolvid Familiar's "during your turn" grant: a layer 6 keyword on the
// other historic creatures (Jhoira's b09IsHistoric: artifact, legendary or Saga)
// its controller controls, on only while that player is the active player, read
// on every recompute. Face down, Bayek has no abilities (CR 708.2a), so the
// grant is off until he is turned up. Disguise is the shared alternative cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b0322e2e-e9f9-4fe2-a7a5-3e49665ad177",
		Name:             "Bayek of Siwa",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"double strike"},
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{1}{R}{W}")},
		Static: []game.StaticAbility{
			func() game.StaticAbility {
				s := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID != source.InstanceID && target.IsCreature() &&
						target.Controller == source.Controller && effects.IsActivePlayer(g, source.Controller) && effects.B09IsHistoric(*target)
				}, "double strike")
				s.Label = "During your turn, other historic creatures you control have double strike."
				return s
			}(),
		},
	})
}
