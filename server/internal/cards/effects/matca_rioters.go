package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Matca Rioters — Creature — Human Warrior {2}{G}, */* (Conflux, #5587):
//
//	"Domain — Matca Rioters's power and toughness are each equal to the
//	 number of basic land types among lands you control."
//
// A characteristic-defining ability in every zone (CR 604.3, layer 7a,
// CR 613.4a), through the shared DomainPT (domain.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b46294cb-9971-415e-90d1-a02c44cd1391",
		Name:         "Matca Rioters",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			DomainPT("Domain — Matca Rioters's power and toughness are each equal to the number of basic land types among lands you control.", 1, 1),
		},
	})
}
