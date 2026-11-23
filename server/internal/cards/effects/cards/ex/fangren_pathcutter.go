package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "36d1a3f4-7903-41aa-9fb7-401d8789b00c",
		Name:         "Fangren Pathcutter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Fangren Pathcutter — attacking creatures gain trample until end of turn",
				effects.Do(effects.GrantKeywordUntilEOT{
					Match:    effects.AttackingCreature(),
					Keywords: []string{"trample"},
					Label:    "Fangren Pathcutter — trample until end of turn",
				})),
		},
	})
}
