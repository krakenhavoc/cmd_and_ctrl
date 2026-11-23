package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cavern Stomper — Creature — Dinosaur {4}{G}{G}, 7/7 (The Lost Caverns
// of Ixalan, #4470):
//
//	"When this creature enters, scry 2.
//	 {3}{G}: This creature can't be blocked by creatures with power 2 or
//	 less this turn."
//
// The enters scry is the footprinted primitive (CR 701.22a). The
// activated ability is Rhonas's Stalwart's block rule
// (CantBeBlockedThisTurnByPower, #2600): a scoped record pinned to this
// object at resolution and swept at the turn's end, with each blocker's
// power read live as blockers are declared (CR 509.1b). A Stomper that
// left and came back before the ability resolved is a new object and
// gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94657522-9977-4d2b-8eb6-f1e64e82cd0f",
		Name:         "Cavern Stomper",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cavern Stomper — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}: This creature can't be blocked by creatures with power 2 or less this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{3}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.CantBeBlockedThisTurnByPower{
					Target: item.SourceCardID,
					AtMost: 2,
					Label:  "Cavern Stomper",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
