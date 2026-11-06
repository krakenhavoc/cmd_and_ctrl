package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seeker of Sunlight — Creature — Merfolk Scout {G}, 1/1 (The Lost
// Caverns of Ixalan, #4470):
//
//	"{2}{G}: This creature explores. Activate only as a sorcery."
//
// The explorer is the object the ability came from (the item's source
// object, #1418), so a Seeker that left and came back as a new object
// before the ability resolved does not get the counter: the explore uses
// the last known information of the one that left (CR 701.44c), and the
// card is still revealed. A land goes to the hand; anything else gives a
// +1/+1 counter and asks whether the card goes to the graveyard (CR
// 701.44a). Sorcery timing is CR 602.5d.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cba9d6a2-1772-42dc-9bfb-c933a092be19",
		Name:         "Seeker of Sunlight",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{2}{G}: This creature explores. Activate only as a sorcery.",
			Cost:         ManaCost("{2}{G}"),
			SorcerySpeed: true,
			Effect:       thisCreatureExplores,
		}},
	})
}

// thisCreatureExplores is "This creature explores" from the creature's own
// activated ability: the object the ability came from explores under the
// ability's controller (CR 701.44a, CR 701.44c).
func thisCreatureExplores(g *game.Game, item *game.StackItem) error {
	ref, ok := NewContext(g, item).SourceRef()
	if !ok {
		ref = game.ObjectRef{ID: item.SourceCardID, Epoch: item.SourceEpoch}
	}
	return g.ExploreForEffect(item.SourceCardID, ref, item.Controller)
}
