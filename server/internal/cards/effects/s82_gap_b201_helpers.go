package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// GapB201Dethrone is the dethrone triggered ability (CR 702.105a):
// "Whenever this creature attacks the player with the most life or tied for
// most life, put a +1/+1 counter on it." The defender is checked once, as
// the attack is declared; an attack at a planeswalker or battle never
// triggers it. label names the card for the stack overlay.
func GapB201Dethrone(label string) game.TriggeredAbility {
	return On(game.EventAttack, PreacherAttackedTheLifeLeader, label,
		Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))
}
