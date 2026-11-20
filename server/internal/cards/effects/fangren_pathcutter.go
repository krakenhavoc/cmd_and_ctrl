package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fangren Pathcutter — Creature — Beast {4}{G}{G}, 4/6 (Fifth Dawn, #5604):
//
//	"Whenever this creature attacks, attacking creatures gain trample
//	 until end of turn."
//
// Chieftain en-Dal's attack trigger, granting trample: every creature
// attacking as it resolves gains it (CR 611.2c), the Pathcutter included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36d1a3f4-7903-41aa-9fb7-401d8789b00c",
		Name:         "Fangren Pathcutter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Fangren Pathcutter — attacking creatures gain trample until end of turn",
				Do(GrantKeywordUntilEOT{
					Match:    AttackingCreature(),
					Keywords: []string{"trample"},
					Label:    "Fangren Pathcutter — trample until end of turn",
				})),
		},
	})
}
