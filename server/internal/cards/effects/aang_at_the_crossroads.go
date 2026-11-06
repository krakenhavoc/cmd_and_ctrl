package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang, at the Crossroads — Legendary Creature — Human Avatar Ally
// {2}{G}{W}{U}, 3/3, the front face of a transforming double-faced card
// (back: Aang, Destined Savior, aang_destined_savior.go):
//
//	"Flying
//	 When Aang enters, look at the top five cards of your library. You
//	 may put a creature card with mana value 4 or less from among them
//	 onto the battlefield. Put the rest on the bottom of your library in
//	 a random order.
//	 When another creature you control leaves the battlefield, transform
//	 Aang at the beginning of the next upkeep."
//
// The enters trigger is Turntimber Symbiosis's look-and-put over five
// cards, narrowed to creature cards with mana value 4 or less, the rest
// going to the bottom in a random order.
//
// The last line is a trigger that creates a delayed trigger (CR 603.7a)
// for the next upkeep, whoever's turn it is (the ruling). Its transform
// is TransformThis, which obeys CR 701.27f: a delayed trigger created
// before Aang transformed does nothing, so several creatures leaving in
// one turn transform him once and never back (the ruling).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        aangAtTheCrossroadsOracle,
		Name:            "Aang, at the Crossroads",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aang, at the Crossroads — look at the top five cards; you may put a creature card with mana value 4 or less onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 5),
						Match:    And(Creature(), ManaValueLE(4)),
						Max:      1,
						Optional: true,
						Label:    "Aang, at the Crossroads — you may put a creature card with mana value 4 or less from among them onto the battlefield",
						ThenTail: PutRestOnBottomInRandomOrderTail,
					}.Apply(ctx)
				}),
			On(game.EventLTB, anotherCreatureYouControlLeft,
				"Aang, at the Crossroads — transform Aang at the beginning of the next upkeep",
				func(g *game.Game, item *game.StackItem) error {
					return ScheduleDelayedTrigger{
						At:    game.StepUpkeep,
						Label: "Aang, at the Crossroads — transform Aang",
						Body:  aangAtTheCrossroadsTransformBody,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

const aangAtTheCrossroadsOracle = "96e5d4a1-e59f-4140-823f-a17e15ee5d8d"

// aangAtTheCrossroadsTransformBody is the delayed trigger's "transform
// Aang". An on-disk key.
var aangAtTheCrossroadsTransformBody = game.SimpleDelayedBody("aang-at-the-crossroads/transform", func(g *game.Game, item *game.StackItem) error {
	return TransformThis{}.Apply(NewContext(g, item))
})
