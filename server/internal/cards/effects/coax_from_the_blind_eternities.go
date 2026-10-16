package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coax from the Blind Eternities — Sorcery {2}{U}:
//
//	"You may reveal an Eldrazi card you own from outside the game or
//	 choose a face-up Eldrazi card you own in exile. Put that card into
//	 your hand."
//
// TakeFromExileOrOutside (ADR 0198): the face-up Eldrazi cards the caster
// owns in exile are offered first, and declining them leads to the pick
// from outside the game, which is the deck's sideboard when it has one
// (ADR 0191) and the card database when it does not. An "Eldrazi card" is
// one with the Eldrazi subtype, which a kindred Eldrazi spell and a
// changeling also have (CR 702.73a).
func init() {
	Register(Spec{
		OracleID:     "63488c95-c249-40d8-a579-3ca311ed57ca",
		Name:         "Coax from the Blind Eternities",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			// "Eldrazi" is a word of the spell's text (CR 612.2, #3984).
			eldrazi := ctx.TextSubtype("Eldrazi")
			return TakeFromExileOrOutside{
				Match:           func(c game.Card) bool { return c.HasSubtype(eldrazi) },
				ExileQuestion:   "Coax from the Blind Eternities — choose a face-up Eldrazi card you own in exile, or none to reveal one from outside the game",
				OutsideQuestion: "Coax from the Blind Eternities — reveal an Eldrazi card you own from outside the game, or none",
			}.Apply(ctx)
		},
	})
}
