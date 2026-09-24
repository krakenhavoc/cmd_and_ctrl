package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Runaway Steam-Kin — Creature — Elemental {1}{R}, 1/1 (EDHREC rank
// 1663):
//
//	"Whenever you cast a red spell, if this creature has fewer than
//	 three +1/+1 counters on it, put a +1/+1 counter on this creature.
//	 Remove three +1/+1 counters from this creature: Add {R}{R}{R}."
//
// The red storm deck's ritual on legs: three red spells charge it,
// and the charge is three red mana back. The trigger is an
// intervening-if (CR 603.4) — fewer than three counters gates it
// going on the stack and is checked again at resolution, so a
// counter placed in response stops it. Red is read off the spell on
// the stack (b15RedSpellCastByYou).
//
// The mana ability's cost has no component of its own in the
// engine's mana-cost vocabulary (tap, sacrifice, life, mana — the
// counter case never had a card asking for it). It is modelled
// with two components that DO exist and together enforce the cost:
// a Condition gating activation on three or more +1/+1 counters
// (checked before anything is paid, so a two-counter Steam-Kin
// cannot activate — CR 602.5), and a Rider that removes three
// counters as the mana lands. A mana ability resolves as one atomic
// step without the stack (CR 605.3b), so nothing can observe that
// the counters left after the mana arrived rather than before. The
// auto-tapper never reaches for an ability with a rider, so the
// counters are only ever spent on purpose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb414a8f-566f-4635-80a1-77b2e18ac8dc",
		Name:         "Runaway Steam-Kin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{},
			Produced: "{R}{R}{R}",
			Label:    "Remove three +1/+1 counters from this creature: Add {R}{R}{R}",
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				c, ok := g.LookupCardForEffect(source)
				return ok && c.Counters["+1/+1"] >= 3
			},
			Rider: func(g *game.Game, _, source uuid.UUID) error {
				return g.AddCounterForEffect(source, "+1/+1", -3)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15RedSpellCastByYou(ev, source, g) && source.Counters["+1/+1"] < 3
			}, "Runaway Steam-Kin — +1/+1 counter", func(g *game.Game, item *game.StackItem) error {
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !b15OnBattlefield(g, item.SourceCardID) || c.Counters["+1/+1"] >= 3 {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
