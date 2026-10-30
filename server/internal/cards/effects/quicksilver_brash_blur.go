package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quicksilver, Brash Blur — Legendary Creature — Mutant Hero {R}, 1/1
// (Marvel Super Heroes, #4455):
//
//	"If Quicksilver, Brash Blur is in your opening hand, you may begin the
//	 game with him on the battlefield.
//	 Haste
//	 Power-up — {4}{R}: Put a +1/+1 counter and a double strike counter on
//	 Quicksilver."
//
// The opening-hand clause is Spec.OpeningHand (Leyline of Sanctity, CR
// 103.6a). The power-up is the shared PowerUp seam (once per permanent, its
// mana cost off the turn he entered); a double strike counter gives the
// keyword (CR 122.1b).
//
// No simplification.
func init() {
	ab := PowerUp("{4}{R}: Put a +1/+1 counter and a double strike counter on Quicksilver.",
		ManaCost("{4}{R}"), mshB7QuicksilverPowerUp)
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "b76ce131-e51a-445e-8216-00df5904ab93",
		Name:            "Quicksilver, Brash Blur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		OpeningHand:     BeginTheGameOnTheBattlefield(),
		Activated:       []ActivatedAbility{ab},
	})
}

func mshB7QuicksilverPowerUp(g *game.Game, item *game.StackItem) error {
	if err := plusOneCountersOnThis(1)(g, item); err != nil {
		return err
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterDoubleStrike, N: 1}.Apply(NewContext(g, item))
}
