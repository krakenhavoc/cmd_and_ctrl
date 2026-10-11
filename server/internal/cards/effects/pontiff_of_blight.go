package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pontiff of Blight — Creature — Zombie Cleric {4}{B}{B}, 2/7:
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do,
//	 each opponent loses 1 life and you gain that much life.)
//	 Other creatures you control have extort. (If a creature has
//	 multiple instances of extort, each triggers separately.)"
//
// The Pontiff's own extort is the shared keyword trigger (extort.go).
// The second line is a layer-6 grant of that same trigger to every
// OTHER creature its controller controls (ADR 0093): each recipient
// carries its own extort, so the trigger is the recipient's, goes on the
// stack under the recipient's controller, and ends with the recipient's
// abilities or the Pontiff's.
//
// Every instance triggers and asks separately (CR 702.101b). Identical
// grants are deliberately not merged in Characteristic.GrantAbility, so
// two other creatures are two more triggers, and a creature that is also
// granted extort by a second Pontiff has two of its own. Each one offers
// its own {W/B} and, when paid, drains one life actually lost and gains
// that much.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5a95092-fb71-4b55-bc54-782bd1c603e1",
		Name:         "Pontiff of Blight",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Extort("Pontiff of Blight")},
		Grants: []AbilityGrant{{
			Key:       pontiffOfBlightGrant,
			Triggered: []game.TriggeredAbility{Extort("Pontiff of Blight")},
			Text:      "Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each opponent loses 1 life and you gain that much life.)",
		}},
		Static: []game.StaticAbility{
			GrantAbilities(otherCreaturesYouControl, pontiffOfBlightGrant),
		},
	})
}

const pontiffOfBlightGrant = "pontiff-of-blight/extort"

// otherCreaturesYouControl is "other creatures you control": a creature
// under the source's controller's control that is not the source.
func otherCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID &&
		target.Controller == source.Controller &&
		target.IsCreature()
}
