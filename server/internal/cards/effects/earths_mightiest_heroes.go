package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earth's Mightiest Heroes — {4}{G}{G} Sorcery:
//
//	"Teamwork 5 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 5 or more.)
//	 Reveal the top eight cards of your library. You may put a creature
//	 card from among them onto the battlefield. If this spell was cast
//	 using teamwork, put any number of creature cards from among them
//	 onto the battlefield instead. Put the rest into your graveyard."
//
// #1703: Teamwork(5). This is Genesis Wave's shape exactly —
// PutFromLibraryOntoBattlefield's three-shape doc names both of them:
// Max 1 (Optional) for "a creature card", Max 0 (Optional) for "any
// number" — with the teamwork announcement picking which ceiling
// applies instead of a printed X. "Any number" already includes zero,
// so it reads as optional with no explicit "you may" needed, same as
// Genesis Wave. No simplification.
func init() {
	Register(Spec{
		OracleID:      "2b87bcac-ec64-4012-9e75-f459d3640eb3",
		Name:          "Earth's Mightiest Heroes",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(5)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 8,
				"Earth's Mightiest Heroes — revealed from the top of the library")
			max := 1
			if ctx.UsedTeamwork() {
				max = 0
			}
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    revealed,
				Match:    Creature(),
				Max:      max,
				Optional: true,
				Label:    "Earth's Mightiest Heroes — put a creature card onto the battlefield",
				Then:     PutRestIntoGraveyard,
			}.Apply(ctx)
		},
	})
}
