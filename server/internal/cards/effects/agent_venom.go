package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent Venom — {2}{B} Legendary Creature — Symbiote Soldier Hero 2/3
// (Marvel's Spider-Man promos, #4460):
//
//	"Flash
//	 Menace
//	 Whenever another nontoken creature you control dies, you draw a
//	 card and lose 1 life."
//
// The trigger is Sifter of Skulls's condition (another creature, not a
// token, that died under your control, read from its last-known
// information) with Black Widow's draw-then-lose body. Life loss is not
// damage (CR 119.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "14c356aa-178d-4c14-a88e-fbe92ee6d89f",
		Name:            "Agent Venom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "menace"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, sifterOfSkullsCondition,
				"Agent Venom — you draw a card and lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
				game.Purpose{Draws: 1, DeathPayoff: true}),
		},
	})
}
