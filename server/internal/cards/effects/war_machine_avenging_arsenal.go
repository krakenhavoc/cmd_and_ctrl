package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "b4f35cdd-3857-4a7d-8ef6-df5424dd40e7",
		Name:            "War Machine, Avenging Arsenal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("War Machine — attacking modified creatures you control gain double strike until end of turn",
				Do(GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl(), AttackingCreature(), modifiedCreature),
					Keywords: []string{"double strike"},
					Label:    "War Machine — double strike",
				})),
		},
	})
}
