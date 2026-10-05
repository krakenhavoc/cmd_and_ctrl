package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tocasia's Welcome — Enchantment {2}{W}:
//
//	"Whenever one or more creatures you control with mana value 3 or
//	 less enter, draw a card. This ability triggers only once each
//	 turn."
//
// "One or more" is the once-per-batch collapse (the engine emits one
// entry event per creature), and "only once each turn" is the
// per-object check Curator of Sun's Creation reads: the harvester
// stamps every trigger it queues, so a second batch later in the turn
// is declined. Mana value is read off the entering creature, so a
// token (value 0) counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25c983e0-a8c9-4784-91a4-8fe04c6df882",
		Name:         "Tocasia's Welcome",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, tocasiasWelcomeSmallCreatureEntered, tocasiasWelcomeLabel, Do(DrawCards{N: 1}))),
		},
	})
}

const tocasiasWelcomeLabel = "Tocasia's Welcome — draw a card"

func tocasiasWelcomeSmallCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() || c.ManaValue() > 3 {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, tocasiasWelcomeLabel)
}
