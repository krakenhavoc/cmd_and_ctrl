package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Expel the Interlopers — Sorcery {3}{W}{W} (EDHREC rank 2772):
//
//	"Choose a number between 0 and 10. Destroy all creatures with
//	 power greater than or equal to the chosen number."
//
// The adjustable wrath: 0 is Wrath of God, 4 spares the tokens and
// the mana dorks, 11 does not exist. The number is asked as a mode
// pick — eleven options, one per number — and the sweep is the mass
// destroy over "creature with power ≥ N", power read as the spell
// resolves, so the whole set leaves as one event and a Blood Artist
// caught in it sees every death.
//
// DECLARED SIMPLIFICATIONS, both weaker than printed:
//
//   - The number is chosen as the spell is CAST, not as it resolves.
//     The engine's only "choose a number" prompt is the mode picker,
//     which opens at announce. That hands opponents the number
//     before they respond — a pump in response can lift a creature
//     over the line, and a shrink can drop one under it — which is
//     strictly worse for the caster than choosing with the board
//     in front of them. Never stronger.
//   - Indestructible is not honoured by the mass-destroy path (#446),
//     the same declared hole every "destroy all" in the catalog
//     carries.
func init() {
	options := make([]game.ModeOption, 0, 11)
	for n := 0; n <= 10; n++ {
		options = append(options, Mode(fmt.Sprintf("%d — destroy all creatures with power %d or greater", n, n)))
	}
	Register(Spec{
		OracleID:     "744b7e93-c893-4770-a3bb-6a45b7f8e4f7",
		Name:         "Expel the Interlopers",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The number is chosen when the spell is cast, not when it resolves, so opponents know it before they respond.",
			"Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it.",
		},
		Modes: ChooseN("Choose a number between 0 and 10", 1, 1, options...),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for n := 0; n <= 10; n++ {
				if ctx.HasMode(n) {
					return DestroyAllMatching{Match: And(Creature(), PowerGE(n))}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
