package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lotus Cobra — Creature — Snake {1}{G}, 2/1 (EDHREC rank 328):
//
//	"Landfall — Whenever a land you control enters, add one mana of
//	 any color."
//
// Every land drop is a Lotus Petal. Landfall is Tireless
// Provisioner's ETB-filtered trigger; the mana is the AddMana
// primitive batch 01 added for Dark Ritual, here with a five-colour
// pipe, so the trigger's resolution queues the same colour pick a
// Birds of Paradise activation does. The mana lands in the pool when
// the trigger RESOLVES — it uses the stack, unlike a mana ability —
// and empties with the pool at the end of the step, so a land played
// in the main phase pays for a spell in that main phase, as printed.
// The printed text says "any color", so the pick offers all five
// colours, commander identity first (AddMana.NarrowToCommanderIdentity
// stays off).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ad91f64-ccab-4edc-bd54-b2ee9267d614",
		Name:         "Lotus Cobra",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Lotus Cobra — add one mana of any color (landfall)", Do(AddMana{Produced: "{W|U|B|R|G}"})),
		},
	})
}
