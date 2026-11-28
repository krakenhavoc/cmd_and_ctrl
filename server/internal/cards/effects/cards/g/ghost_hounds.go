package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost Hounds — Creature — Dog Spirit {1}{B}, 1/1 (Homelands, S82 OLD b826,
// #5635):
//
//	"Vigilance
//	 Whenever this creature blocks or becomes blocked by a white creature,
//	 this creature gains first strike until end of turn."
//
// Vigilance is a printed keyword. The trigger is the per-pair block trigger
// (CR 509.3d): one for each blocker-attacker pair that names the Hounds
// and a white creature, so it fires once for each white creature it blocks or
// is blocked by. First strike is granted to the Hounds itself until end of
// turn (CR 514.2), and only while it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "26957a44-6155-4688-a850-a3b49b011a9f",
		Name:            "Ghost Hounds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.OldB826BlocksOrBecomesBlockedByWhite,
				"Ghost Hounds — gains first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike"},
						Label: "Ghost Hounds — first strike"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
