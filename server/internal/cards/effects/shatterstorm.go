package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shatterstorm — Sorcery {2}{R}{R}:
//
//	"Destroy all artifacts. They can't be regenerated."
//
// The unconditional artifact wrath, and the answer to the mana-rock
// pile every Commander deck is built on. Symmetric, unlike
// Vandalblast's overload, which is why a red deck plays whichever
// one matches its own board: Vandalblast if you have rocks of your
// own, Shatterstorm if you do not and want them all gone.
//
// "They can't be regenerated" is cosmetic — regeneration is not
// modelled. See mass.go.
func init() {
	Register(Spec{
		OracleID: "96ce2403-4607-440a-92ae-80aceb458c5d",
		Name:     "Shatterstorm",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Artifact()}.Apply(ctx)
		},
	})
}
