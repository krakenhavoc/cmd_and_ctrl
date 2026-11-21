package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Machine, Avenging Arsenal — {4}{R} Legendary Artifact Creature —
// Human Hero, 3/5 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever War Machine attacks, attacking modified creatures you
//	 control gain double strike until end of turn. (Equipment, Auras you
//	 control, and counters are modifications.)"
//
// Flying is the printed keyword. The set is fixed as the trigger resolves
// (CR 611.2c): every creature that is attacking then, controlled by the
// trigger's controller and modified by Kodama of the West Tree's reading
// (b07IsModified: a counter, an Equipment, or an Aura its controller
// controls). War Machine counts herself when she carries a counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b4f35cdd-3857-4a7d-8ef6-df5424dd40e7",
		Name:            "War Machine, Avenging Arsenal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("War Machine — attacking modified creatures you control gain double strike until end of turn",
				effects.Do(effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl(), effects.AttackingCreature(), effects.ModifiedCreature),
					Keywords: []string{"double strike"},
					Label:    "War Machine — double strike",
				})),
		},
	})
}
