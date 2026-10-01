package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sulfuric Vortex — Enchantment for {1}{R}:
//
//	"At the beginning of each player's upkeep, this enchantment deals 2
//	 damage to that player.
//	 If a player would gain life, that player gains no life instead."
//
// The upkeep trigger fires on EVERY player's upkeep (it used to fire on
// its controller's alone, which was not the printed card) and the Vortex
// deals the 2 damage to that player — the active player the event
// names — with the Vortex as the source, so damage routing and
// prevention see it as ordinary damage.
//
// "Gains no life instead" is a CR 614 REPLACEMENT on the life window
// (ADR 0107 §5, #1880), not CR 119.7's "can't gain life": it replaces
// every positive life change with nothing (CR 614.10's null
// replacement, so no "whenever you gain life" trigger sees it), and as
// a replacement it is ordered against another "if you would gain life"
// replacement by the affected player (CR 616.1) — a Rhox Faithmender
// doubling first still ends at nothing. A gain of 0 is no life gain
// event at all (CR 119.10), so it never applies to one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7652f328-e142-494b-a869-772ced10c26a",
		Name:         "Sulfuric Vortex",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Sulfuric Vortex — 2 damage to that player", func(g *game.Game, item *game.StackItem) error {
				p := triggeringActor(item)
				if p == uuid.Nil {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: p, Amount: 2}.Apply(NewContext(g, item))
			}),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventLife && ev.LifeDelta > 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			PureCancel: true,
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Sulfuric Vortex — that player gains no life instead",
		}},
	})
}
