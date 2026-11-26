package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB317PreventOneOfThatDamage is the Replace body of "prevent 1 of that
// damage" (CR 615.1a; Djeru, With Eyes Open and Orbs of Warding): it takes 1
// off the event and cancels it when nothing is left.
func S82OldB317PreventOneOfThatDamage(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
	ev.DamageAmount--
	if ev.DamageAmount <= 0 {
		ev.Cancel()
	}
	return nil
}
