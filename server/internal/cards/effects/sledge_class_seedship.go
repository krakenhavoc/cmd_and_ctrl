package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sledge-Class Seedship — Artifact — Spacecraft {2}{G}, 4/5 (Edge of
// Eternities, #4461):
//
//	"Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 7+.)
//	 7+ | Flying
//	 Whenever this Spacecraft attacks, you may put a creature card from
//	 your hand onto the battlefield."
//
// Station and the 7+ line are Susurian Dirgecraft's helpers: at seven or
// more charge counters it is a 4/5 artifact creature with flying
// (CR 721.2). The attack trigger is printed outside the threshold
// (CR 721.4), though only a creature can attack. The put is Quicksilver
// Amulet's optional PutFromHandOntoBattlefield: the card is put, not
// cast, and it enters untapped and not attacking (CR 508.4).
//
// No simplification.
func init() {
	const label = "Sledge-Class Seedship — you may put a creature card from your hand onto the battlefield"
	Register(Spec{
		OracleID:     "bcd086ce-de5c-4b34-92a1-ece22055e479",
		Name:         "Sledge-Class Seedship",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{Match: Creature(), Optional: true, Label: label}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(7, 4, 5),
			ThresholdKeywords(7, "flying"),
		),
	})
}
