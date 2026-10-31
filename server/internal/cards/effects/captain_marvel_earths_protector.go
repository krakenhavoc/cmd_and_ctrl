package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Marvel, Earth's Protector — {3}{W}{W} Legendary Creature — Human
// Kree Hero 5/4:
//
//	"Flash
//	 Flying, lifelink
//	 Power-up — {5}{W}{W}: Put a +1/+1 counter and an indestructible
//	 counter on Captain Marvel. (Activate each power-up ability only once.
//	 Reduce the cost by her mana cost if she entered this turn.)"
//
// Power-up is the engine's (CR 702.193, game/power_up.go). The ability is
// one effect that puts both counters on her; the indestructible counter is
// a real counter (CR 122.1b), so the row also declares that it survives.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	Register(Spec{
		OracleID:        "2e665a03-bece-43df-bd15-b4496e233f2e",
		Name:            "Captain Marvel, Earth's Protector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying", "lifelink"},
		Activated:       []ActivatedAbility{captainMarvelPowerUpAbility()},
	})
}

// captainMarvelPowerUpAbility builds the row with PowerUp and declares its
// answers by hand (powerUpAnswering is being retired, #4571).
func captainMarvelPowerUpAbility() ActivatedAbility {
	ab := PowerUp("{5}{W}{W}: Put a +1/+1 counter and an indestructible counter on Captain Marvel.",
		ManaCost("{5}{W}{W}"), captainMarvelPowerUp)
	ab.Purpose.Answers = game.AnswerPump | game.AnswerSurvive
	return ab
}

func captainMarvelPowerUp(g *game.Game, item *game.StackItem) error {
	if err := plusOneCountersOnThis(1)(g, item); err != nil {
		return err
	}
	return putIndestructibleCounterOnSource()(g, item)
}
