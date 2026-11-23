package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Humbler of Mortals — Enchantment Creature — Elemental {4}{G}{G}, 5/5 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, creatures you control gain trample until end of
//	 turn."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control. The
// creatures are the ones you control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91933f1a-4aad-4f67-8151-a47ab6d0e42c",
		Name:         "Humbler of Mortals",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.ConstellationTriggers, "Humbler of Mortals — creatures you control gain trample until end of turn (constellation)",
				effects.Do(effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"trample"},
					Label: "Humbler of Mortals — trample"})),
		},
	})
}
