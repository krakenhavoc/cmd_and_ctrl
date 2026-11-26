package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sage of the Inward Eye — Khans of Tarkir (S82 OLD b348, #5554):
//
//	Creature — Djinn Wizard {2}{U}{R}{W}, 3/4: Flying. Whenever you cast a
//	noncreature spell, creatures you control gain lifelink until end of turn.
//
// The set of creatures is fixed as the trigger resolves (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "02c741d0-99d3-48e5-9846-0fea5a0a29fe",
		Name:            "Sage of the Inward Eye",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Sage of the Inward Eye — creatures you control gain lifelink until end of turn",
				effects.Do(effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"lifelink"},
					Label: "Sage of the Inward Eye — lifelink until end of turn"})),
		},
	})
}
