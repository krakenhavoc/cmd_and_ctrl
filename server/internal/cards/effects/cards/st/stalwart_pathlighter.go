package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalwart Pathlighter — Creature — Human Soldier {2}{W}, 3/1 (Midnight
// Hunt, #5497):
//
//	"Vigilance
//	 Coven — At the beginning of combat on your turn, if you control
//	 three or more creatures with different powers, creatures you control
//	 gain indestructible until end of turn."
//
// The set of creatures is fixed as the trigger resolves (CR 611.2c), so a
// creature that arrives later in the turn does not gain it.
//
// No simplification.
func init() {
	const label = "Stalwart Pathlighter — Coven: creatures you control gain indestructible until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "88ba283d-2d03-4455-a4a3-a2a5622209bd",
		Name:            "Stalwart Pathlighter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.MidB05AtBeginningOfCombatIfCoven(label, func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"indestructible"},
					Label:    label,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
