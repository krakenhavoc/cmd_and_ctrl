package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gallifrey Stands — Legendary Enchantment {4}{W}{U} (Doctor Who):
//
//	"When Gallifrey Stands enters, return all Doctor cards from your
//	 graveyard to your hand.
//	 At the beginning of your upkeep, you may put a Doctor creature card
//	 from your hand onto the battlefield. Then if you control thirteen or
//	 more Doctors, you win the game."
//
// The enter trigger is Crystal Chimes' whole-graveyard return over Doctor
// cards. The upkeep trigger is an optional put from hand (Howlpack Piper's
// PutFromHandOntoBattlefield) whose tail — run whether or not a card
// entered, since "then" is unconditional — counts the Doctors its controller
// controls as the tail runs and wins the game at thirteen (CR 104.2b).
//
// No simplification.
func init() {
	const name = "Gallifrey Stands"
	Register(Spec{
		OracleID:     "52680471-a407-418b-9e74-c32308322772",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(name+" — return all Doctor cards from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return returnWholeGraveyardToHand(NewContext(g, item), gapIsDoctorCard)
				}),
			AtYourUpkeep(name+" — you may put a Doctor creature card from your hand onto the battlefield; then if you control thirteen or more Doctors, you win the game",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Match:    And(Creature(), HasSubtype("Doctor")),
						Optional: true,
						Label:    name + " — you may put a Doctor creature card from your hand onto the battlefield",
						ThenTail: gallifreyStandsWin,
						ThenItem: true,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// gallifreyStandsWin is "Then if you control thirteen or more Doctors, you
// win the game." An on-disk key.
var gallifreyStandsWin = PutFromHandTail("gallifrey-stands/win-at-thirteen-doctors", func(ctx *Context, res PutFromHandResult) error {
	g := ctx.Game
	doctors := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == res.Player && gapIsDoctorCard(c) {
			doctors++
		}
	}
	if doctors < 13 {
		return nil
	}
	return WinTheGame{Player: res.Player}.Apply(NewContext(g, &game.StackItem{SourceCardID: res.Source, Controller: res.Player}))
})
