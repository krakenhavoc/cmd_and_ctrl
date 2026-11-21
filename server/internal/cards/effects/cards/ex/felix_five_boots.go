package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Felix Five-Boots — Legendary Creature — Ooze Rogue {2}{B}{G}{U}, 5/4:
// "Menace, ward {2}. If a creature you control dealing combat damage to a
// player causes a triggered ability of a permanent you control to trigger,
// that ability triggers an additional time."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "95175d35-c815-47bd-8e69-427972b567b6",
		Name:            "Felix Five-Boots",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		TriggerDoublers: []game.TriggerDoubler{effects.Cmdr2S52CombatDamageToPlayerDoubler("Felix Five-Boots")},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Felix Five-Boots — ward {2}"),
		},
	})
}
