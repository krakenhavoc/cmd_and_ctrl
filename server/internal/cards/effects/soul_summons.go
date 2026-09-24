package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Summons — {1}{W} Sorcery: "Manifest the top card of your
// library."
//
// The whole card is the keyword action, which is why it is the one in
// the catalog: everything manifest does — the CR 614 entry window, the
// 2/2 with no name or text, the controller as its only knower
// (CR 708.5), no ETB trigger off the card underneath (CR 708.2a), and
// the CR 701.40b permission to turn it face up for its mana cost IF it
// is a creature card — is the engine's, and the card file is one line.
//
// The card it manifests is not revealed, and that is the point of the
// spell rather than a detail: two mana buys a 2/2 whose identity only
// you know, and whether it is a bear or a Sheoldred is a fact you are
// allowed to sit on.
func init() {
	Register(Spec{
		OracleID:     "0f5b79ca-9f80-420b-a6c5-bb2a9a95c7e7",
		Name:         "Soul Summons",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Manifest{N: 1}.Apply(ctx)
		},
	})
}
