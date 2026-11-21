package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const elvishMarinerTapLabel = "Elvish Mariner — tap up to X target nonland permanents, where X is the number of cards looked at while scrying"

// Elvish Mariner — Creature — Elf Pilot {2}{U}, 3/2 (The Lord of the Rings:
// Tales of Middle-earth Commander):
//
//	"Whenever this creature attacks, scry 1.
//	 Whenever you scry, tap up to X target nonland permanents, where X is
//	 the number of cards looked at while scrying this way."
//
// A self-attack scry (Purpose declares the one card) and a scry watch whose
// target clause is sized from the event (Elrond's TargetsFrom): a scry 1
// offers one target, a scry 2 two. Attacking scries, so it triggers the
// second ability, which can tap a blocker. Each target still legal as the
// ability resolves is tapped (CR 608.2b).
//
// No simplification.
func init() {
	tap := game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventScry},
		AppliesTo: effects.ByYou,
		Key:       elvishMarinerTapLabel,
		TargetsFrom: effects.GapB127ScryTargetsFromLookedAt(func(x int) *game.TargetSpec {
			return effects.TargetPermanent("up to X target nonland permanents", effects.Nonland()).WithCount(0, x)
		}),
		Effect: effects.Msc13TapTheLegalTargets,
	}
	effects.Register(effects.Spec{
		OracleID:     "2c595ccb-6703-4670-a1e5-0e74d4b8d3c8",
		Name:         "Elvish Mariner",
		Purpose:      game.Purpose{Scry: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Elvish Mariner — scry 1", effects.Do(effects.Scry{N: 1})),
			tap,
		},
	})
}
