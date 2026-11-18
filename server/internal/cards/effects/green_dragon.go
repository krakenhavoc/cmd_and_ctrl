package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Green Dragon — Creature — Dragon {4}{G}{G}, 4/4 (Adventures in the
// Forgotten Realms):
//
//	"Flying
//	 Poison Breath — When this creature enters, until end of turn, whenever
//	 a creature an opponent controls is dealt damage, destroy it."
//
// The enters trigger creates a repeating delayed triggered ability (CR
// 603.7a-b) controlled by its controller, Shriveling Rot's first bullet
// narrowed to creatures an opponent controls (afrB18OpponentCreatureDealtDamage).
// It triggers once for each creature in each damage instance and destroys
// the object that was dealt the damage if it is still on the battlefield
// (CR 603.7c); the Dragon itself leaving does not end it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c2a4c6c7-e5b4-4cc8-91d8-5de8cc95ed00",
		Name:            "Green Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Green Dragon — until end of turn, whenever a creature an opponent controls is dealt damage, destroy it",
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:        "Green Dragon — destroy the creature that was dealt damage",
						On:           []game.EventKind{game.EventDealDamage},
						Condition:    afrB18OpponentCreatureDealtDamageCondition,
						Body:         destroyThePayloadPermanentBody,
						Repeats:      true,
						OncePerBatch: true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
