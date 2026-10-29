package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wonder Man, Hollywood Hero — {3}{R}{R} Legendary Creature — Human
// Performer Hero 4/4:
//
//	"Flying
//	 Each power-up ability of permanents you control can be activated an
//	 additional time.
//	 Power-up — {5}{R}{R}: Put two +1/+1 counters on Wonder Man.
//	 (Activate each power-up ability only once. Reduce the cost by his
//	 mana cost if he entered this turn.)"
//
// The middle line is a PowerUpExtra (#4571, game/power_up.go): a static
// read from the battlefield at the moment of each activation, so it
// covers every power-up ability of every permanent its controller has,
// Wonder Man's own included, ends the moment he leaves (an ability
// activated twice is then simply spent), and adds to the printed once
// rather than replacing it. The activation record counts each
// activation at the announce, so a countered power-up is still one.
//
// No simplification. Marvel Super Heroes (#4455, seam #4571).
func init() {
	Register(Spec{
		OracleID:        "e2c49754-711e-43b3-82f1-5376a4d99701",
		Name:            "Wonder Man, Hollywood Hero",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		PowerUpExtras: []game.PowerUpExtra{
			YourPowerUpAbilitiesActivateAnAdditionalTime("Each power-up ability of permanents you control can be activated an additional time."),
		},
		Activated: []ActivatedAbility{
			powerUpAnswering("{5}{R}{R}: Put two +1/+1 counters on Wonder Man.",
				ManaCost("{5}{R}{R}"), plusOneCountersOnThis(2)),
		},
	})
}
