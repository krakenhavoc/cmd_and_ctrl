package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bandit's Haul — Artifact {3} (Outlaws of Thunder Junction, #4468):
//
//	"Whenever you commit a crime, put a loot counter on this artifact.
//	 This ability triggers only once each turn. (Targeting opponents,
//	 anything they control, and/or cards in their graveyards is a
//	 crime.)
//	 {T}: Add one mana of any color.
//	 {2}, {T}, Remove two loot counters from this artifact: Draw a card."
//
// The crime trigger is Blood Hustler's with a loot counter for the +1/+1
// (CR 700.13, once each turn by the per-turn trigger tally). The mana
// ability is the ordinary any-colour tap. The draw removes the two
// counters as part of its cost, so it can't be activated with fewer.
//
// No simplification.
const otjB1BanditsHaulLabel = "Bandit's Haul — put a loot counter on this artifact"

func init() {
	Register(Spec{
		OracleID:     "84b08b55-0623-4202-98c6-1756d3c1bb8f",
		Name:         "Bandit's Haul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return YouCommittedCrime(ev, source) &&
					!b11TriggeredThisTurn(g, source.InstanceID, otjB1BanditsHaulLabel)
			}, otjB1BanditsHaulLabel, putACounterOnThis(otjB1LootCounter)),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Remove two loot counters from this artifact: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost(), RemoveCountersFromThis(otjB1LootCounter, 2)),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
