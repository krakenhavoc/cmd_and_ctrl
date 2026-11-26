package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drakestown Forgotten — Creature — Zombie {4}{B}, 0/0 (Conspiracy):
//
//	"This creature enters with X +1/+1 counters on it, where X is the number
//	 of creature cards in all graveyards.
//	 {2}{B}, Remove a +1/+1 counter from this creature: Target creature gets
//	 -1/-1 until end of turn."
//
// The counters are the CR 614.1c entry clause, counted as the permanent
// enters (B19EntersWithCountersCounted), off the creature cards in every
// graveyard (Mortivore's count). The cost removes the counter at announce.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3c8d9ef-4ffb-482b-8859-486e7dc4621e",
		Name:         "Drakestown Forgotten",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return effects.B41CreatureCardsInAllGraveyards(g)
			}, "Drakestown Forgotten enters with X +1/+1 counters, X the number of creature cards in all graveyards"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Remove a +1/+1 counter from this creature: Target creature gets -1/-1 until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
			Effect:  effects.PumpAndGrantFirstTarget(-1, -1, "Drakestown Forgotten — -1/-1 until end of turn"),
		}},
	})
}
