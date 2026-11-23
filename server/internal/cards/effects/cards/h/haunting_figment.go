package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haunting Figment — Creature — Illusion {1}{U}, 2/1 (Dominaria United, S82
// DMU batch 4):
//
//	"Vigilance
//	 This creature can't be blocked as long as you've cast an instant or
//	 sorcery spell this turn."
//
// Expedition Lookout's conditional unblockability: a block rule on the
// attacker itself, read live as blockers are declared, over the controller's
// cast tally for the turn (the card types of every spell cast). Vigilance is
// a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed416ae3-9291-4e1e-ad83-2f14138e9e47",
		Name:            "Haunting Figment",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, you uuid.UUID, _ game.Card) bool {
				for _, t := range g.CastTallyFor(you).CardTypesCast() {
					if t == "Instant" || t == "Sorcery" {
						return true
					}
				}
				return false
			}),
		},
	})
}
