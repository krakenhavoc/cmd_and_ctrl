package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sidequestRaiseAChocoboOracle is shared by both faces: the front
// registers under it, the back under sidequestRaiseAChocoboOracle + "#1".
const sidequestRaiseAChocoboOracle = "aebeaf0b-4df1-4c34-919b-740bfab19e56"

// Sidequest: Raise a Chocobo // Black Chocobo — a transforming
// enchantment (Final Fantasy, #4462):
//
//	Sidequest: Raise a Chocobo, {1}{G} Enchantment
//	  "When this enchantment enters, create a 2/2 green Bird creature
//	   token with "Whenever a land you control enters, this token gets
//	   +1/+0 until end of turn."
//	   At the beginning of your first main phase, if you control four or
//	   more Birds, transform this enchantment."
//
//	Black Chocobo, Creature — Bird, 2/2
//	  "When this permanent transforms into Black Chocobo, search your
//	   library for a land card, put it onto the battlefield tapped, then
//	   shuffle.
//	   Landfall — Whenever a land you control enters, Birds you control
//	   get +1/+0 until end of turn."
//
// The front's enters trigger makes the set's Chocobo Bird
// (chocoboBirdToken). Its second ability is an intervening "if" (CR
// 603.4): checked as the first main phase begins and again as the ability
// resolves (the ruling), over the Birds its controller controls (a word of
// its text, read through the layers); it transforms the enchantment in
// place (CR 701.27f). The back's search is an ordinary land search that
// puts the land on the battlefield tapped, made by a trigger on the
// transform (CR 701.27e), so a permanent that ENTERS on the back face
// never searched (CR 712.18, the ruling). Its landfall pumps every Bird
// you control as of resolution (CR 611.2c), itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     sidequestRaiseAChocoboOracle,
		Name:         "Sidequest: Raise a Chocobo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sidequest: Raise a Chocobo — create a 2/2 green Bird token with landfall +1/+0",
				Do(CreateToken{Template: chocoboBirdToken(), N: 1})),
			InterveningIf(sidequestFourBirds{}, On(game.EventBeginPrecombatMain,
				AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return sidequestBirds(g, source.Controller, source.TextSubtype("Bird")) >= 4
				}),
				"Sidequest: Raise a Chocobo — transform this enchantment", Do(TransformThis{}))),
		},
	})
	Register(Spec{
		OracleID:     sidequestRaiseAChocoboOracle + "#1",
		Name:         "Black Chocobo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventTransform, ThisTransformedInto(1),
				"Black Chocobo — search your library for a land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.IsLand() },
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Black Chocobo — a land card, tapped",
					}.Apply(NewContext(g, item))
				}), game.Purpose{Lands: 1}),
			Landfall("Black Chocobo — Birds you control get +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Creature(), YouControl(), OfSubtype(g.TextReaderForEffect(item).TextSubtype("Bird"))),
						Power: 1, Label: "Black Chocobo — +1/+0",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// sidequestBirds counts the Birds `player` controls.
func sidequestBirds(g *game.Game, player uuid.UUID, bird string) int {
	return countControlled(g, player, func(c game.Card) bool { return c.IsCreature() && c.HasSubtypeWord(bird) })
}

// sidequestFourBirds is the intervening "if you control four or more
// Birds".
type sidequestFourBirds struct{}

func (sidequestFourBirds) Holds(ctx *Context) bool {
	bird := "Bird"
	if src, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok {
		bird = src.TextSubtype("Bird")
	}
	return sidequestBirds(ctx.Game, ctx.Controller(), bird) >= 4
}

func (sidequestFourBirds) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondBoard, 0), true
}
