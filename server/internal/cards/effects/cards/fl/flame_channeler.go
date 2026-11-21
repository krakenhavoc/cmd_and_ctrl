package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// flameChannelerOracle is shared by both faces: the front registers under
// it, the back under flameChannelerOracle + "#1".
const FlameChannelerOracle = "ca8950ce-43d9-47bd-85f3-58c1fad42420"

// Flame Channeler // Embodiment of Flame — a transforming creature
// (Midnight Hunt):
//
//	Flame Channeler {1}{R}, Creature — Human Wizard, 2/2
//	  "When a spell you control deals damage, transform this creature."
//
//	Embodiment of Flame, Creature — Elemental Wizard, 3/3
//	  "Whenever a spell you control deals damage, put a flame counter on
//	   this creature.
//	   {1}, Remove a flame counter from this creature: Exile the top card
//	   of your library. You may play that card this turn."
//
// Both triggers read the damage event's source: it must be a SPELL, not a
// permanent's ability, and one its controller controls (the helper file's
// s82MidB17SpellYouControlDealtDamage). A spell that deals damage to
// several things at once is one occurrence, so one trigger. The turn over
// is in place (TransformThis, CR 701.27f). The activated ability pays the
// counter as a cost, then exiles the top card and lets its controller play
// it this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     FlameChannelerOracle,
		Name:         "Flame Channeler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82MidB17WheneverASpellYouControlDealsDamage(
				"Flame Channeler — transform this creature", effects.Do(effects.TransformThis{})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     FlameChannelerOracle + "#1",
		Name:         "Embodiment of Flame",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82MidB17WheneverASpellYouControlDealsDamage(
				"Embodiment of Flame — put a flame counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: "flame", N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Remove a flame counter from this creature: Exile the top card of your library. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis("flame", 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
