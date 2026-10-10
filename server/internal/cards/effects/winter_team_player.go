package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winter, Team Player — Legendary Creature — Human Warrior {4}{R}, 3/3:
//
//	"Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Whenever you cast a noncreature spell, creatures you control get
//	 +1/+0 until end of turn."
//
// The creature set is locked as the trigger resolves (CR 611.2c), so a
// creature that arrives later in the turn is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "254f133a-eee5-48b9-b19f-110555e49300",
		Name:         "Winter, Team Player",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(),
				"Winter, Team Player — creatures you control get +1/+0 until end of turn",
				Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1,
					Label: "Winter, Team Player — +1/+0"})),
		},
	})
}
