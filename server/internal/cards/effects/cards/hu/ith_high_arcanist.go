package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ith, High Arcanist — Legendary Creature — Human Wizard {5}{W}{U}, 3/5:
//
//	"Vigilance
//	 {T}: Untap target attacking creature. Prevent all combat damage that
//	 would be dealt to and dealt by that creature this turn.
//	 Suspend 4—{W}{U}"
//
// ADR 0108 §7, Delivery PR 7 (#1904): Maze of Ith's ability on a
// creature (one to-and-by record), with suspend's special action (#659),
// which gives the creature haste when it is cast from exile.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb933723-b879-44c8-bbc8-d8c84ce5ab12",
		Name:            "Ith, High Arcanist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{effects.UntapAttackerToAndByRow(
			"{T}: Untap target attacking creature. Prevent all combat damage that would be dealt to and dealt by that creature this turn.",
			effects.TapCost(), effects.TargetCreature("target attacking creature", effects.AttackingCreature()))},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(4, "{W}{U}"),
		},
	})
}
