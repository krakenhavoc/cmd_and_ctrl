package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bothersome Quasit — Creature — Demon {2}{R}, 3/2 (Commander Legends: Battle for
// Baldur's Gate):
//
//	"Menace
//	 Goaded creatures your opponents control can't block.
//	 Whenever you cast a noncreature spell, goad target creature an opponent controls.
//	 (Until your next turn, that creature attacks each combat if able and attacks a
//	 player other than you if able.)"
//
// Menace is printed. The second line is a block rule on the blocker's side (CR
// 509.1b): a creature a player other than the Quasit's controller controls that is
// goaded as the block is declared, whether by a resolved goad or a continuous one,
// can't block any attacker. The goad is the shared GoadTarget, put on the stack
// with its target as the noncreature spell is cast (CR 701.15a, CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "06cf7b77-ffd7-4192-903e-19bd06887ba1",
		Name:            "Bothersome Quasit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.GapB219GoadedCreatureOfAnOpponent, effects.OnMatching(effects.Creature()),
				"it is a goaded creature an opponent of Bothersome Quasit's controller controls"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.Noncreature(), "Bothersome Quasit — goad target creature an opponent controls", effects.GoadTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
