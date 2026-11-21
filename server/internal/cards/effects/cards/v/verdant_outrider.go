package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verdant Outrider — Creature — Human Knight {2}{G}, 4/2 (Wilds of
// Eldraine, #4471):
//
//	"{1}{G}: This creature can't be blocked by creatures with power 2 or
//	 less this turn."
//
// Cavern Stomper's activated ability at a lower price: the block rule
// (CantBeBlockedThisTurnByPower) is pinned to this object as the ability
// resolves and swept at the turn's end, with each blocker's power read as
// blockers are declared (CR 509.1b). An Outrider that left and came back
// before the ability resolved is a new object and gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fc049e7-a16d-4b48-b930-5c346f6718b9",
		Name:         "Verdant Outrider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: This creature can't be blocked by creatures with power 2 or less this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.CantBeBlockedThisTurnByPower{
					Target: item.SourceCardID,
					AtMost: 2,
					Label:  "Verdant Outrider",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
