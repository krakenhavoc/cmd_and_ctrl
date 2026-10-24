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
				Pick:          coaxFromTheBlindEternitiesTake,
				Exile:         coaxFromTheBlindEternitiesExile,
				Word:          eldrazi,
				ExileQuestion: "Coax from the Blind Eternities — choose a face-up Eldrazi card you own in exile, or none to reveal one from outside the game",
			}.Apply(ctx)
		},
	})
}

// coaxFromTheBlindEternitiesTake is the spell's pick from outside the
// game, keyed (#4256). The subtype is the spell's word, which a
// text-changing effect may have changed, so it rides on the prompt.
var coaxFromTheBlindEternitiesTake = registerOutsideTakeOfWord("choose-cards/coax-from-the-blind-eternities-take",
	func(c game.Card, eldrazi string) bool { return c.HasSubtype(eldrazi) }, true)

// coaxFromTheBlindEternitiesExile is the exile pick, keyed (#4264): a
// table waiting on it is a restore point. The spell's word for "Eldrazi"
// rides the prompt, and "none" asks the outside pick with it.
var coaxFromTheBlindEternitiesExile = registerExileOrOutside("choose-cards/coax-from-the-blind-eternities-exile",
	coaxFromTheBlindEternitiesTake, "Coax from the Blind Eternities — reveal an Eldrazi card you own from outside the game, or none")
