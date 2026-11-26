package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grief Tyrant — Creature — Horror {5}{B/R}, 8/8 (Shadowmoor, S82 OLD b532,
// #5591):
//
//	"This creature enters with four -1/-1 counters on it.
//	 When this creature dies, put a -1/-1 counter on target creature for each
//	 -1/-1 counter on this creature."
//
// Soulstinger's dies trigger without the "you may": the target is chosen as it
// goes on the stack, and the count is the -1/-1 counters the Tyrant had as it
// died (last-known information, CR 603.10a, 608.2h), read as the ability
// resolves. It may target itself's former place only by another creature: the
// Tyrant is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce855b35-4a31-4e42-9dbc-fea8d244fbe0",
		Name:         "Grief Tyrant",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB532MinusOneCountersAtEntry("Grief Tyrant", 4)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Grief Tyrant — put a -1/-1 counter on target creature for each -1/-1 counter on this creature",
				effects.OldB532MinusOneCountersOnTargetForEachOnThis),
				effects.TargetCreature("target creature")),
		},
	})
}
