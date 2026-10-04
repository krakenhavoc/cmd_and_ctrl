package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiery Inscription — Enchantment {2}{R}:
//
//	"When this enchantment enters, the Ring tempts you.
//	 Whenever you cast an instant or sorcery spell, this enchantment
//	 deals 2 damage to each opponent."
//
// The second ability is Guttersnipe's: it triggers on the cast (CR
// 601.2i), so it resolves before the spell does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "111889d4-bcca-4b1f-ad48-3077e2f5136f",
		Name:         "Fiery Inscription",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fiery Inscription — the Ring tempts you", Do(TheRingTemptsYou{})),
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastByYou(ev, source, g)
			}, "Fiery Inscription — 2 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 2)
			}),
		},
	})
}
