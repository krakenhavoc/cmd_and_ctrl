package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunderous Debut — Sorcery {6}{G}{G} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Look at the top twenty cards of your library. You may reveal up to two
//	 creature cards from among them. If this spell was bargained, put the
//	 revealed cards onto the battlefield. Otherwise, put the revealed cards
//	 into your hand. Then shuffle."
//
// Bargain is the shared optional cost (g2Bargain, read back at
// resolution). The look is private; the "up to two creature cards" pick is
// a floor-of-none choice over the twenty (a shorter library offers what it
// has). The unbargained put is a take to hand with the cards revealed
// (takeRevealThenShuffle); the bargained put is Animist's Awakening's
// battlefield put, the cards entering together. Either way the library is
// shuffled afterwards, whatever was taken (the shuffle is the registered
// tail / take rest), and the reveal of the bargained pair is the entry
// itself, which every player sees.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "07995092-af6c-43b8-b30b-17e020ead37c",
		Name:          "Thunderous Debut",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tutors: 2},
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			look := ctx.Game.LookAtTopOfLibraryForEffect(player, 20)
			if g2WasBargained(ctx) {
				return PutFromLibraryOntoBattlefield{
					Player:   player,
					Cards:    look,
					Match:    Creature(),
					Max:      2,
					Optional: true,
					ThenTail: thunderousDebutShuffle,
				}.Apply(ctx)
			}
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    look,
				Match:    Creature(),
				Max:      2,
				Optional: true,
				Label:    "Thunderous Debut — you may reveal up to two creature cards from among them and put them into your hand",
				Key:      takeRevealThenShuffle,
			}.Apply(ctx)
		},
	})
}

// takeRevealThenShuffle is "…reveal … and put it into your hand. Then
// shuffle." (Thunderous Debut's unbargained half). An on-disk key.
var takeRevealThenShuffle = RegisterTake("choose-cards/take-reveal-then-shuffle",
	TakeDef{Reveal: true, Then: func(g *game.Game, res TakeFromLibraryResult) error {
		return g.ShuffleLibraryForEffect(res.Player)
	}})

// thunderousDebutShuffle is the shuffle after the bargained put. An on-disk
// key.
var thunderousDebutShuffle = CardTail("thunderous-debut/shuffle-after-put", func(ctx *Context, _ game.TailOutcome) error {
	return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
})
