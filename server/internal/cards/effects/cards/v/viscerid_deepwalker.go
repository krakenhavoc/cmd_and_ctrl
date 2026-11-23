package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viscerid Deepwalker — Creature — Homarid Warrior {4}{U}, 2/3 (Time
// Spiral, #5596):
//
//	"{U}: This creature gets +1/+0 until end of turn.
//	 Suspend 4—{U} (Rather than cast this card from your hand, you may pay
//	 {U} and exile it with four time counters on it. At the beginning of
//	 your upkeep, remove a time counter. When the last is removed, you may
//	 cast it without paying its mana cost. It has haste.)"
//
// The shared firebreathing row (selfPumpRow) and the engine's suspend
// (CR 702.62a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9dea848b-d758-46d0-829d-b096061fd460",
		Name:         "Viscerid Deepwalker",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Viscerid Deepwalker", "{U}", 1, 0)},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(4, "{U}"),
		},
	})
}
