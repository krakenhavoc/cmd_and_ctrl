package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turntimber Symbiosis // Turntimber, Serpentine Wood — Modal DFC. Front:
// Sorcery {4}{G}{G}{G}:
//
//	"Look at the top seven cards of your library. You may put a creature
//	 card from among them onto the battlefield. If that card has mana
//	 value 3 or less, it enters with three additional +1/+1 counters on
//	 it. Put the rest on the bottom of your library in a random order."
//
// Back: Land, pay 3 life or enter tapped (mdfc_lands.go).
//
// The put-from-among-them primitive (Armored Skyhunter's shape), with the
// counters decided by the card picked: Counters, gated on its mana value in
// the library (data, so the pick is a restore point, #4253), seeds three
// +1/+1 counters onto the entry (CR 122.6,
// 614.1c), so they are there as it enters and go through any counter
// replacement. A creature card with mana value 4 or more enters plain. The
// rest go to the bottom in a random order whether or not a card was put.
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "403b59f3-7ade-4bc2-a3e6-de0c3c700f18",
		Name:         "Turntimber Symbiosis",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := item.Controller
			return PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 7),
				Match:    Creature(),
				Max:      1,
				Optional: true,
				Label:    "Turntimber Symbiosis — you may put a creature card from among them onto the battlefield",
				Counters: PutCounters{Kind: game.CounterPlusOne, N: 3, IfManaValueAtMost: true, ManaValueAtMost: 3},
				ThenTail: PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		},
	})
}
