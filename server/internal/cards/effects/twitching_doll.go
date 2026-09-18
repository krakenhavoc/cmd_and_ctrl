package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twitching Doll — Artifact Creature — Spider Toy {1}{G}, 2/2 (EDHREC
// rank 1501):
//
//	"{T}: Add one mana of any color. Put a nest counter on this
//	 creature.
//	 {T}, Sacrifice this creature: Create a 2/2 green Spider creature
//	 token with reach for each counter on this creature. Activate
//	 only as a sorcery."
//
// A mana dork that banks every tap as a Spider. The mana ability is
// the any-colour pick with the nest counter as its rider — the
// counter goes on through AddCounter, so a counter doubler sees it —
// and both abilities are a creature's {T}, so summoning sickness
// applies to each (CR 302.6).
//
// "For each counter on this creature" is read for a creature that
// was sacrificed to pay the cost: it is in the graveyard with its
// counters cleared by the time the ability resolves (CR 400.7), so
// the count is read back off the event log — the most recent total
// of every counter kind it carried, b13LastKnownCounterTotal. Every
// kind counts, as printed: a +1/+1 counter from elsewhere makes a
// Spider too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd6e1967-237a-41f6-bbf4-2c869f9447c8",
		Name:         "Twitching Doll",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color. Put a nest counter on this creature.",
			Rider: func(g *game.Game, _ uuid.UUID, source uuid.UUID) error {
				return g.AddCounterForEffect(source, "nest", 1)
			},
		}},
		Activated: []ActivatedAbility{{
			Label:        "{T}, Sacrifice this creature: Create a 2/2 green Spider creature token with reach for each counter on this creature.",
			Cost:         Plus(TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := b13LastKnownCounterTotal(g, item.SourceCardID)
				return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 green Spider with reach"), N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
