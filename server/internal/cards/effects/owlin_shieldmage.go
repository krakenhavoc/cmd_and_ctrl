package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Owlin Shieldmage — Creature — Bird Warlock {3}{W}{B}, 3/3 (Strixhaven:
// School of Mages, #5502):
//
//	"Flying
//	 Ward—Pay 3 life. (Whenever this creature becomes the target of a
//	 spell or ability an opponent controls, counter it unless that player
//	 pays 3 life.)"
//
// Flying is the printed keyword; the ward is Sedgemoor Witch's life ward,
// asking the caster for 3 life (a caster below 3 cannot pay, CR 119.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "50c477c1-8457-40f6-bba9-1a4ba0862178",
		Name:            "Owlin Shieldmage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(3), "Owlin Shieldmage — ward, pay 3 life"),
		},
	})
}
