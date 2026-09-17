package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hall of Oracles — Land (EDHREC rank 6679):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {T}: Put a +1/+1 counter on target creature. Activate only as a
//	 sorcery and only if you've cast an instant or sorcery spell this
//	 turn."
//
// A Strixhaven filter land with a Spellslinger rider. The filter is a
// mana ability with a mana cost, paid from mana already floating (the
// Signet posture). The counter ability prints both instructions:
// "only as a sorcery" is SorcerySpeed, and "only if you've cast an
// instant or sorcery spell this turn" is the condition (CR 602.1b,
// #743), read off this turn's cast events for you. A spell counts
// once cast, whether it resolved or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bb896ba-e15e-43a9-9120-e674d7ba003c",
		Name:         "Hall of Oracles",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:                    ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced:                "{W|U|B|R|G}",
				Label:                   "{1}, {T}: Add one mana of any color",
				IgnoreCommanderIdentity: true,
			},
		},
		Activated: []ActivatedAbility{{
			Label:        "{T}: Put a +1/+1 counter on target creature. Activate only as a sorcery and only if you've cast an instant or sorcery spell this turn.",
			Cost:         TapCost(),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Condition:    hallOfOraclesCastInstantOrSorceryThisTurn,
			Effect:       b36CounterOnChosenAnimal,
		}},
	})
}

// hallOfOraclesCastInstantOrSorceryThisTurn is the Hall's counter
// condition: an EventCast this turn, by you, of an instant or sorcery.
func hallOfOraclesCastInstantOrSorceryThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCast || ev.Actor != controller {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && (c.IsInstant() || c.IsSorcery()) {
			return true
		}
	}
	return false
}
