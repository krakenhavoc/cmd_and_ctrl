package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spock, Always Your Friend — Legendary Creature — Vulcan Human Scientist
// {U}{R}, 2/2 (Star Trek, #4973):
//
//	"When Spock enters, look at the top four cards of your library. You
//	 may reveal a noncreature, nonland card from among them and put it
//	 into your hand. Put the rest on the bottom of your library in any
//	 order."
//
// Gavony Dawnguard's look-and-take dig (TakeFromLibraryToHand: a private
// look at four, an optional revealed take of one noncreature, nonland
// card, the rest ordered onto the bottom by the player, ADR 0088).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdfe5809-17d4-4239-a9ee-6accdef99035",
		Name:         "Spock, Always Your Friend",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spock, Always Your Friend — look at the top four cards of your library", spockAlwaysYourFriendDig),
		},
	})
}

func spockAlwaysYourFriendDig(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player:   player,
		Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
		Match:    And(Noncreature(), Nonland()),
		Max:      1,
		Optional: true,
		Label:    "Spock, Always Your Friend — you may reveal a noncreature, nonland card and put it into your hand",
		Key:      takeRevealRestBottomAny,
	}.Apply(ctx)
}
