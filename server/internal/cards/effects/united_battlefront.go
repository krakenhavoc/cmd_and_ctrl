package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// United Battlefront — Sorcery {3}{W} (Tarkir: Dragonstorm, #4463):
//
//	"Look at the top seven cards of your library. Put up to two
//	 noncreature, nonland permanent cards with mana value 3 or less from
//	 among them onto the battlefield. Put the rest on the bottom of your
//	 library in a random order."
//
// Ao, the Dawn Sky's dig: the seven cards are looked at, the pick is up to
// two permanent cards (an instant or sorcery is no permanent card) that
// are neither creatures nor lands with mana value 3 or less (a card with
// {X} in its cost counts X as 0 in a library, CR 202.3e), and the rest go
// to the bottom in a random order. The permanents enter as ordinary
// entries, replacement effects and enters triggers included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "892a83a0-a3cb-4cfb-9106-4008e4e75c0f",
		Name:         "United Battlefront",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := item.Controller
			return PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 7),
				Match:    And(Permanent(), Noncreature(), Nonland(), ManaValueLE(3)),
				Max:      2,
				Optional: true,
				Label:    "United Battlefront — put up to two noncreature, nonland permanent cards with mana value 3 or less onto the battlefield",
				ThenTail: PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		},
	})
}
