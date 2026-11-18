package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mordenkainen — Legendary Planeswalker — Mordenkainen {4}{U}{U}, loyalty 5
// (Adventures in the Forgotten Realms, S82 batch 15, #5499):
//
//	"+2: Draw two cards, then put a card from your hand on the bottom of
//	     your library.
//	 −2: Create a blue Dog Illusion creature token with 'This token's power
//	     and toughness are each equal to twice the number of cards in your
//	     hand.'
//	 −10: Exchange your hand and library, then shuffle. You get an emblem
//	      with 'You have no maximum hand size.'"
//
// The +2 is Errand-Rider of Gondor's draw then put-back, with the draw taken
// first (a one-card hand is the pick, an empty one asks nothing). The Dog
// Illusion is the printed token template, its size a characteristic-defining
// ability read live (CR 604.3). The −10 exchange (CR 701.12d, 701.12f: a swap of two
// zones' contents) is two moves of cards decided up front — the library's
// cards go to the hand, the hand's original cards go into the library — then
// the shuffle, chained as keyed tails so a commander's CR 903.9 choice on
// either leg is a restore point. The emblem's one ability is a maximum-hand-
// size static (EmblemSpec.HandSize), as Wrenn and Seven's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6be0bf15-fdcb-4425-9bd8-3c39039cab7d",
		Name:            "Mordenkainen",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &EmblemSpec{
			Label:    "Mordenkainen emblem",
			Text:     "You have no maximum hand size.",
			HandSize: []game.HandSizeStatic{{Kind: game.HandSizeNoMaximum}},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Draw two cards, then put a card from your hand on the bottom of your library.",
				Cost:    LoyaltyCost(2),
				Purpose: game.Purpose{Draws: 2},
				Effect: Do(DrawCards{N: 2}, Instruction(func(ctx *Context) error {
					return rfPutCardFromHandOnBottom(ctx, "Mordenkainen — put a card from your hand on the bottom of your library")
				})),
			},
			{
				Label:  "−2: Create a blue Dog Illusion creature token with \"This token's power and toughness are each equal to twice the number of cards in your hand.\"",
				Cost:   LoyaltyCost(-2),
				Effect: Do(CreateToken{Template: AfrB15DogIllusionToken(), N: 1}),
			},
			{
				Label: "−10: Exchange your hand and library, then shuffle. You get an emblem with \"You have no maximum hand size.\"",
				Cost:  LoyaltyCost(-10),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := mordenkainenExchange(g, item.Controller, item.SourceCardID); err != nil {
						return err
					}
					return CreateEmblem{}.Apply(ctx)
				},
			},
		},
	})
}

// mordenkainenExchange swaps the contents of the player's hand and library:
// the library's cards go to the hand, then the hand's original cards go into
// the library, then the library is shuffled.
func mordenkainenExchange(g *game.Game, player, source uuid.UUID) error {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Library == nil {
		return nil
	}
	hand := allHandCardIDs(g, player)
	library := make([]uuid.UUID, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		library = append(library, c.InstanceID)
	}
	shuffle := game.KeyedTail{Ref: mordenkainenShuffleTail, Player: player, Source: source}
	toLibrary := game.KeyedTail{Ref: mordenkainenHandToLibraryTail, Player: player, Source: source, Carry: hand, Next: &shuffle}
	if len(library) == 0 {
		return toLibrary.Run(g, nil)
	}
	return g.TakeFromLibraryToHandKeyedForEffect(player, library, toLibrary)
}

// mordenkainenHandToLibraryTail is the exchange's second leg: the cards that
// were in the hand go into the library, then the shuffle. On-disk key: never
// renamed, never reused.
var mordenkainenHandToLibraryTail = game.RegisterTail("tail/mordenkainen-hand-to-library", func(g *game.Game, o game.TailOutcome) error {
	next := game.KeyedTail{}
	if o.Tail.Next != nil {
		next = *o.Tail.Next
	}
	if len(o.Tail.Carry) == 0 {
		return next.Run(g, nil)
	}
	return g.TuckCardsToLibraryKeyedForEffect(o.Tail.Carry, game.TuckOptions{}, next)
})

// mordenkainenShuffleTail is the exchange's last step: shuffle the library.
var mordenkainenShuffleTail = game.RegisterTail("tail/mordenkainen-shuffle", func(g *game.Game, o game.TailOutcome) error {
	return g.ShuffleLibraryForEffect(o.Tail.Player)
})
