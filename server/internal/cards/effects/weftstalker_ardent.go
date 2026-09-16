package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weftstalker Ardent — 2/3 Creature — Drix Artificer for {2}{R}:
//
//	"Whenever another creature or artifact you control enters, this
//	 creature deals 1 damage to each opponent.
//	 Warp {R}"
//
// Reckless Fireweaver widened to creatures as well as artifacts,
// which in this deck means every Treasure AND every Pirate pings the
// table. The "another" is load-bearing: it doesn't ping off its own
// arrival.
//
// Warp {R} (S29, #324). "You may cast this card from your hand for
// its warp cost. Exile this creature at the beginning of the next
// end step, then you may cast it from exile on a later turn."
//
// A three-mana ping engine for one red now, taken back at end of
// turn, and bought again at full price later. The whole clause is
// composed from machinery that already shipped: the discount is an
// AlternativeCost (#257), the take-back is a CR 603.7 delayed
// trigger (#255), and the later cast is an ordinary cast from exile
// under the ExilePlayPermission impulse exile introduced (#269) with
// S29's NotBeforeTurn floor carrying "on a later turn".
//
// This card was the one exception ADR 0037 §5 named — a Spec that
// shipped knowingly omitting a printed clause. The omission is gone.
func init() {
	Register(Spec{
		OracleID:         "926d52a5-4db1-46ce-9567-17c28bf56ae7",
		Name:             "Weftstalker Ardent",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{R}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && (c.IsCreature() || c.IsArtifact())
			}, "Weftstalker Ardent — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
