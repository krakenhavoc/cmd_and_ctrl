package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetherworks Marvel — Legendary Artifact {4}:
//
//	"Whenever a permanent you control is put into a graveyard, you get
//	 {E} (an energy counter).
//	 {T}, Pay six {E}: Look at the top six cards of your library. You may
//	 cast a spell from among them without paying its mana cost. Put the
//	 rest on the bottom of your library in a random order."
//
// The trigger is an EventLTB bound for a graveyard, asked of the
// permanent as it last stood under its controller
// (APermanentYouControlWasPutIntoAGraveyard, CR 603.10a): a creature you stole and sacrificed died under YOUR
// control, a bounce or an exile is not "put into a graveyard", and the
// Marvel's own death counts, because a leaves-the-battlefield ability
// looks back. A token counts too: it is put into a graveyard on its way
// out (CR 111.7). A board wipe that takes five of your permanents
// triggers it five times, as printed.
//
// The activated ability pays six energy as a cost (ADR 0129 §2) and runs
// lookAtTopAndOfferFreeCast: a PRIVATE look (only the looker knows the
// cards), a choose-up-to-one over the nonland cards, the rest to the
// bottom in a random order, and a pass-closed free cast of the pick from
// the library (free_library_cast.go). A land is not a spell, so it is
// never offered. An uncast pick goes to the bottom with the rest.
func init() {
	Register(Spec{
		OracleID:     "470ec98e-120c-446d-ade6-7d80aa836b55",
		Name:         "Aetherworks Marvel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, APermanentYouControlWasPutIntoAGraveyard,
				"Aetherworks Marvel — you get {E}", Do(GetEnergy{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay six {E}: Look at the top six cards of your library. You may cast a spell from among them without paying its mana cost. Put the rest on the bottom of your library in a random order.",
			Cost:    Plus(TapCost(), PayEnergy(6)),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return lookAtTopAndOfferFreeCast(g, ctx.Controller(), ctx.Source(), "Aetherworks Marvel", 6)
			},
		}},
	})
}
