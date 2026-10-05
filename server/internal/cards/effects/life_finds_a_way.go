package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Life Finds a Way — Enchantment {2}{G}:
//
//	"Whenever a nontoken creature you control with power 4 or greater
//	 enters, populate. (Create a token that's a copy of a creature
//	 token you control.)"
//
// An ETB watcher: the entering permanent must be a creature you
// control, not a token, with current power 4 or more when it enters
// (so a +1/+1 counter or a lord counts, as printed). Power is not
// re-checked at resolution; the printed trigger has no "if". Populate
// is the shared primitive in populate.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa4f22b2-7fe4-4f7f-8afd-32add76d740d",
		Name:         "Life Finds a Way",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && !IsToken(c) && c.CurrentPower() >= 4
			}, "Life Finds a Way — populate", Do(Populate{})),
		},
	})
}
