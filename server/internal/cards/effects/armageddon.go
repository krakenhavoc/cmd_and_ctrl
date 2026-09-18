package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armageddon — Sorcery {3}{W} (EDHREC rank 4051):
//
//	"Destroy all lands."
//
// The oldest mass land destruction spell in the game, and the reason
// it is in the batch is that its scope is the one nobody expects: ALL
// lands, the caster's included, with no "you control" and no
// exceptions. A four-mana sweep with no rider is also the cleanest
// possible check that the mass-destruction primitive really does run
// over every seat rather than the controller's half of the board.
//
// Destroy, not exile: an indestructible land (Darksteel Citadel,
// Urborg under a Blessing) survives, a land with regeneration set up
// regenerates, and every land that does die goes to its OWNER's
// graveyard, where Crucible of Worlds and Ramunap Excavator can find
// it. All of that falls out of routing the sweep through
// DestroyAllMatching, which narrows to the destructible permanents the
// engine's own DestructibleForEffect identifies.
//
// A land that is also a creature (an animated Mutavault, a Dryad
// Arbor) is a land and is destroyed, because the predicate reads the
// post-layer type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9ed8b01-959a-47d6-891e-0abbdccf6e4f",
		Name:         "Armageddon",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Land()}.Apply(ctx)
		},
	})
}
