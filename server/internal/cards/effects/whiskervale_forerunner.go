package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whiskervale Forerunner — Creature — Mouse Bard {3}{W}, 3/4 (Bloomburrow,
// #5230):
//
//	"Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, look at the top five
//	 cards of your library. You may reveal a creature card with mana value 3
//	 or less from among them. You may put it onto the battlefield if it's
//	 your turn. If you don't put it onto the battlefield, put it into your
//	 hand. Put the rest on the bottom of your library in a random order."
//
// The shared valiant trigger (valiant.go). The look is private to you. On
// someone else's turn it is Gavony Dawnguard's take: you may reveal one
// such card and put it into your hand, and the rest go to the bottom in a
// random order. On your turn (read as the ability resolves) the same
// choice has a second destination, asked in the order that lets one card
// reach either place: first whether to put one onto the battlefield, and
// only if you put none there, whether to reveal one and put it into your
// hand. Either way you end with at most one of them, where you chose, and
// the other four (or five) on the bottom in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "861e427f-cbb4-4ed5-8356-747ac638ac34",
		Name:         "Whiskervale Forerunner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Valiant("Whiskervale Forerunner — Valiant: look at the top five cards of your library", whiskervaleForerunnerDig),
		},
	})
}

// whiskervaleForerunnerFind is the clause's filter: "a creature card with
// mana value 3 or less".
func whiskervaleForerunnerFind() CardPredicate { return And(Creature(), ManaValueLE(3)) }

func whiskervaleForerunnerDig(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 5)
	if !IsYourTurn(g, player) {
		return whiskervaleForerunnerToHand(ctx, player, looked)
	}
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    looked,
		Match:    whiskervaleForerunnerFind(),
		Max:      1,
		Optional: true,
		Label:    "Whiskervale Forerunner — you may put a creature card with mana value 3 or less onto the battlefield",
		ThenTail: whiskervaleForerunnerAfterPut,
	}.Apply(ctx)
}

// whiskervaleForerunnerToHand is the hand half: reveal one, take it, and
// the rest to the bottom in a random order.
func whiskervaleForerunnerToHand(ctx *Context, player uuid.UUID, cards []uuid.UUID) error {
	return TakeFromLibraryToHand{
		Player:   player,
		Cards:    cards,
		Match:    whiskervaleForerunnerFind(),
		Max:      1,
		Optional: true,
		Label:    "Whiskervale Forerunner — you may reveal a creature card with mana value 3 or less and put it into your hand",
		Key:      takeRevealRestBottomRandom,
	}.Apply(ctx)
}

// whiskervaleForerunnerAfterPut runs once the battlefield question is
// answered. A card that entered leaves only the rest to put away; none
// entering is "if you don't put it onto the battlefield", and the hand
// offer follows over the same cards. An on-disk key.
var whiskervaleForerunnerAfterPut = PutFromLibraryTail("whiskervale-forerunner/after-put",
	func(ctx *Context, res PutFromLibraryResult) error {
		if len(res.Entered) > 0 {
			return PutRestOnBottomInRandomOrder(ctx.Game, res)
		}
		return whiskervaleForerunnerToHand(ctx, res.Player, res.Rest)
	})
