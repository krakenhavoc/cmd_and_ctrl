package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapB56CatWarriorToken is Ajani Fells the Godsire's token: a 2/1 white Cat
// Warrior creature, as printed (Modern Horizons 3). A vanilla token, built
// here rather than tabled, so the shared token table is not touched.
func GapB56CatWarriorToken() game.Card {
	return game.Card{
		Name:      "Cat Warrior",
		TypeLine:  "Token Creature — Cat Warrior",
		Power:     2,
		Toughness: 1,
		Colors:    []string{"W"},
	}
}

// SpellgorgerWeirdToken is Ral and the Implicit Maze's token: a {2}{R} 2/2
// red Weird creature with "Whenever you cast a noncreature spell, put a
// +1/+1 counter on this creature." (Modern Horizons 3). It has a mana cost
// as printed, so its mana value is 3.
func SpellgorgerWeirdToken() game.Card { return TokenFromCatalog(printedGapB56SpellgorgerWeirdToken) }

// printedGapB56SpellgorgerWeirdToken is that token as PRINTED, with its
// trigger declared as the Weird card's own is.
func printedGapB56SpellgorgerWeirdToken() TokenTemplate {
	return TokenTemplate{
		Slug: "spellgorger-weird",
		Card: game.Card{
			Name:      "Spellgorger Weird",
			TypeLine:  "Token Creature — Weird",
			ManaCost:  "{2}{R}",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Spellgorger Weird — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Text: "Whenever you cast a noncreature spell, put a +1/+1 counter on this creature.\n(This token's mana cost is {2}{R}.)",
	}
}

// gapB56RalExileTwo is the paid half of Ral and the Implicit Maze's chapter
// II: once a card has really been discarded, exile the top two cards of the
// library; they may be played until the end of the controller's next turn.
// An on-disk key.
var GapB56RalExileTwo = game.RegisterDiscardThen("choose-cards/gap-b56-ral-exile-two",
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 || r.Item == nil {
			return nil
		}
		ctx := NewContext(g, r.Item)
		window := g.UntilEndOfYourNextTurnDuration(ctx.Controller())
		return ExileTopWithPermission{
			From:    ctx.Controller(),
			GrantTo: ctx.Controller(),
			N:       2,
			Window:  &window,
		}.Apply(ctx)
	})

// gapB56TamiyoInvestigate is the paid half of Tamiyo Meets the Story Circle's
// chapter II: "investigate twice for each card discarded this way", counted by
// the run (the cards that really left the hand). An on-disk key.
var GapB56TamiyoInvestigate = game.RegisterDiscardThen("choose-cards/gap-b56-tamiyo-investigate-twice-each",
	func(g *game.Game, r game.DiscardsDone) error {
		n := 2 * r.Discarded.Count()
		if n == 0 || r.Item == nil {
			return nil
		}
		return Investigate{Player: r.Player, N: n}.Apply(NewContext(g, r.Item))
	})
