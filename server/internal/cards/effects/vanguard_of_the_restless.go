package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanguard of the Restless — Creature — Spirit Knight {2}{W}, 2/2
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying
//	 Spirits you control get +1/+1 for each time you've cast your
//	 commander from the command zone this game.
//	 Whenever a Spirit you control enters, you may pay {2}{W}. If you do,
//	 return this card from your graveyard to the battlefield."
//
// The lord is a scaling anthem over Player.CommanderCasts (summed over every
// commander, the tally only a cast from the command zone bumps), read on
// every recompute; it counts Vanguard itself, a Spirit. The return trigger
// watches from the GRAVEYARD only (InGraveyard), where "you" is the card's
// owner (CR 108.4); the payment is asked as it resolves and checked against
// the same graveyard object (CR 400.7), like Nether Traitor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5564c782-be40-49fc-abc0-6cdcbaaa575a",
		Name:            "Vanguard of the Restless",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{vanguardOfTheRestlessAnthem()},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventETB, vanguardOfTheRestlessSeesASpirit,
				"Vanguard of the Restless — you may pay {2}{W} to return it to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return mayPayForGraveyardReturn(g, item, "{2}{W}", "Vanguard of the Restless — pay {2}{W} to return it to the battlefield?")
				})),
		},
	})
}

// vanguardOfTheRestlessSeesASpirit is "whenever a Spirit you control enters"
// for the copy in the graveyard.
func vanguardOfTheRestlessSeesASpirit(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Spirit")) // a word of the source's text (CR 612.2, #3984)
}

// vanguardOfTheRestlessAnthem is the scaling lord. DependsOnSpellsCast drops
// the cached layer resolution when a spell is cast: the commander-cast tally
// is bumped just before the cast event, so the Spirits grow the moment the
// commander is on the stack.
func vanguardOfTheRestlessAnthem() game.StaticAbility {
	ab := TribalScalingAnthem(
		TribeFilter{Tribes: []string{"Spirit"}, YoursOnly: true},
		func(source *game.Card, g *game.Game) int {
			return socB3CommanderCastsFromCommandZone(g, source.Controller)
		})
	ab.DependsOnSpellsCast = true
	return ab
}
