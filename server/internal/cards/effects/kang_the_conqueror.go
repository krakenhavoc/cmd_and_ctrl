package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kang the Conqueror — {2}{U}{U} Legendary Creature — Human Villain 4/5:
//
//	"Flying
//	 Power-up — {5}{U}{U}{U}: Put a +1/+1 counter on Kang. Take an extra
//	 turn after this one. During that turn, power-up abilities can't be
//	 activated. (Activate each power-up ability only once. Reduce the
//	 cost by his mana cost if he entered this turn.)"
//
// The counter goes on first, then the extra turn is queued (CR 500.7)
// with a game-wide rules effect attached to it (TakeExtraTurn.DuringThatTurn,
// ADR 0188 Decision 3b, CR 611.2a): ModCantActivatePowerUp, registered as
// that turn begins and swept at its cleanup step. The activation gate
// (game.Game.ActivationGateLocked) refuses every power-up ability while
// it stands, whoever would activate it, so the turn is a plain turn for
// power-ups and the turn Kang's ability resolved in is untouched. If the
// extra turn is skipped the effect goes with it (CR 614.10a). If Kang has
// left the battlefield by resolution the counter has nowhere to go and
// the turn is still taken: nothing in the ability depends on him.
//
// No simplification. Marvel Super Heroes (#4455, seam #4571).
func init() {
	ab := PowerUp(
		"{5}{U}{U}{U}: Put a +1/+1 counter on Kang. Take an extra turn after this one. During that turn, power-up abilities can't be activated.",
		ManaCost("{5}{U}{U}{U}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(1)(g, item); err != nil {
				return err
			}
			return TakeExtraTurn{
				DuringThatTurn: PowerUpAbilitiesCantBeActivatedDuringThatTurn(),
				Label:          "Kang the Conqueror — power-up abilities can't be activated during that turn",
			}.Apply(NewContext(g, item))
		})
	// ADR 0142 / ADR 0188 Decision 10: a counter on itself, so a pump, and
	// the controller's extra turn, which the bot prices.
	ab.Purpose = game.Purpose{Answers: game.AnswerPump, ExtraTurns: 1}
	Register(Spec{
		OracleID:        "da6e24a4-8adc-41a1-b5eb-3e31dc1250c9",
		Name:            "Kang the Conqueror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{ab},
	})
}
