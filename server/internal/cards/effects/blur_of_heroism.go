package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blur of Heroism — {2}{W} Enchantment (Marvel Super Heroes Commander,
// #4455):
//
//	"When this enchantment enters, draw a card.
//	 You may cast Hero spells as though they had flash."
//
// The draw is an ordinary enters trigger. The timing grant is Valley
// Floodcaller's CastKindAsThoughFlash narrowed to a creature subtype
// (PermissionFilter.CreatureType), so a non-Hero spell still waits for
// its own timing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4a9b328-f976-4390-be9a-711ef357a50d",
		Name:         "Blur of Heroism",
		Completeness: CompletenessFull,
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{CreatureType: "Hero"},
				"You may cast Hero spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Blur of Heroism — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
