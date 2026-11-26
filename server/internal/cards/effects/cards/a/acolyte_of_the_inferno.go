package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acolyte of the Inferno — Creature — Human Monk {2}{R}, 3/1 (Magic Origins, S82
// older sets):
//
//	"Renown 1 (When this creature deals combat damage to a player, if it isn't
//	 renowned, put a +1/+1 counter on it and it becomes renowned.)
//	 Whenever this creature becomes blocked by a creature, it deals 2 damage to
//	 that creature."
//
// Renown is the engine's keyword trigger (Honored Hierarch's). The second
// ability is Pygmy Troll's per-blocker trigger (CR 509.3d): one for each creature
// that blocks it, naming that blocker, with Ornery Goblin's damage from the
// Acolyte as an object, so one that has left the battlefield still deals it (CR
// 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "35bf67e0-8141-4664-b762-2c613353e1eb",
		Name:            "Acolyte of the Inferno",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"renown 1"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.ViciousBattleragerBlocked,
				"Acolyte of the Inferno — 2 damage to the creature blocking it", effects.OldB323DamageToTheBlocker(2)),
		},
	})
}
