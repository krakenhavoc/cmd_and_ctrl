package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pheres-Band Warchief — Creature — Centaur Warrior {3}{G}, 3/3 (Journey into Nyx, #5558):
//
//	"Vigilance, trample
//	 Other Centaur creatures you control get +1/+1 and have vigilance
//	 and trample."
//
// A Centaur lord: the anthem and the two keyword grants share one filter,
// other Centaurs you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "79bc9b71-5b04-46b8-a9de-9e9f17936f29",
		Name:            "Pheres-Band Warchief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Static: []game.StaticAbility{
			TribalAnthem(pheresBandWarchiefCentaurs, 1, 1),
			TribalKeywordGrant(pheresBandWarchiefCentaurs, "vigilance"),
			TribalKeywordGrant(pheresBandWarchiefCentaurs, "trample"),
		},
	})
}

// pheresBandWarchiefCentaurs is "other Centaur creatures you control".
var pheresBandWarchiefCentaurs = TribeFilter{Tribes: []string{"Centaur"}, Others: true, YoursOnly: true}
