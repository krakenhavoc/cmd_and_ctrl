package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yennett, Cryptic Sovereign — Legendary Creature — Sphinx {2}{W}{U}{B},
// 3/5 (Commander Masters):
//
//	"Flying, vigilance, menace
//	 Whenever Yennett attacks, reveal the top card of your library. You may
//	 cast it without paying its mana cost if its mana value is odd. If you
//	 don't cast it, draw a card."
//
// Descendants' Path's reveal-and-offer. The mana value is read off the
// revealed card (CR 202.3: an {X} in a library is 0, which is even, and a
// land is 0 as well); an even card, or a land, is never offered and the
// draw follows at once. An odd card is offered as an ordinary cast after
// the resolution (ADR 0066, ADR 0091); the draw is the offer's tail and runs
// only when nothing was cast, so declining draws the revealed card, which
// is still on top, and a cast card leaves nothing to draw for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4eec6838-3565-407f-8941-ce9421402f0b",
		Name:            "Yennett, Cryptic Sovereign",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(yennettLabel, yennettReveal),
		},
	})
}

const yennettLabel = "Yennett, Cryptic Sovereign — reveal the top card; cast it free if its mana value is odd, otherwise draw a card"

func yennettReveal(g *game.Game, item *game.StackItem) error {
	you, source := item.Controller, item.SourceCardID
	revealed := g.RevealTopOfLibraryForEffect(you, source, 1, yennettLabel)
	if len(revealed) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(revealed[0]); ok && !c.IsLand() {
		if mv, ok := c.ParsedManaValue(); ok && mv%2 == 1 {
			perm := anySpell.on(g.FreeCastFromLibraryPermission(you, source, "Yennett, Cryptic Sovereign"))
			perm.LapseOnPass = game.LapseStaysInHand
			if len(freeCastCandidates(g, &perm, revealed)) > 0 && g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
				Chooser:      you,
				Source:       source,
				Reason:       "Yennett, Cryptic Sovereign",
				Cards:        revealed,
				Permission:   &perm,
				Keyword:      game.MayCastKeywordFreeCast,
				AcceptLabel:  "Cast it free",
				DeclineLabel: "Draw it",
				Then:         yennettDrawIfNotCast,
				Carry:        revealed,
			}) {
				return nil
			}
		}
	}
	return g.DrawNForEffect(you, 1)
}

// yennettDrawIfNotCast is "if you don't cast it, draw a card": the offer's
// tail, run once the prompt is answered with the cards of the offer that
// were cast as its Picked. An on-disk key, never renamed.
var yennettDrawIfNotCast = game.RegisterChooseCardsThen("resolution-cast/yennett-draw-if-not-cast",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) > 0 {
			return nil
		}
		return g.DrawNForEffect(r.Chooser, 1)
	})
