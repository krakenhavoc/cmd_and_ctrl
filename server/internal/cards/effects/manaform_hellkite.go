package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manaform Hellkite — Creature — Dragon {2}{R}{R}, 4/4 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Flying
//	 Whenever you cast a noncreature spell, create an X/X red Dragon
//	 Illusion creature token with flying and haste, where X is the amount
//	 of mana spent to cast that spell. Exile that token at the beginning
//	 of the next end step."
//
// X is the mana actually spent, read off the spell's payment
// record as it is cast (Aetherflux Conduit's Build) and carried on the
// item, so a spell countered in response still makes its Dragon. A spell
// cast without paying its mana cost spent nothing: the token is a 0/0 and
// dies to CR 704.5f (the ruling). The exile is a delayed trigger over the
// token that landed (CR 603.7), scheduled by the creation's keyed tail.
//
// No simplification.
func init() {
	const label = "Manaform Hellkite — create an X/X red Dragon Illusion with flying and haste"
	t := WheneverYouCast(Noncreature(), label, manaformHellkiteDragon)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = g.StackItemPaidForEffect(ev.CardID).Spent().Total()
		return item
	}
	Register(Spec{
		OracleID:        "f426aa1a-e1dc-4563-b450-2e23d6bd980b",
		Name:            "Manaform Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ReadsManaSpent:  game.ReadsYourCasts,
		Triggered:       []game.TriggeredAbility{t},
	})
}

// manaformHellkiteDragon creates the X/X Dragon Illusion and hands the
// tokens that landed to the end step exile.
func manaformHellkiteDragon(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := max(item.Params.Amount, 0)
	tmpl := game.Card{
		Name:           "Dragon Illusion",
		TypeLine:       "Token Creature — Dragon Illusion",
		Power:          x,
		Toughness:      x,
		Colors:         []string{"R"},
		Keywords:       []string{"flying", "haste"},
		PrintedPTKnown: true,
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(tmpl), Count: 1}},
	}, ctx.ItemTail(manaformHellkiteExile))
}

// manaformHellkiteExile is "exile that token at the beginning of the next
// end step". An on-disk key.
var manaformHellkiteExile = socB8ExileTokensAtNextEndStep("manaform-hellkite/exile-the-token",
	"Manaform Hellkite — exile the Dragon Illusion")
