package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Laser — Legendary Creature — Elemental Villain {4}{R}, 4/4 (Marvel
// Super Heroes Commander, #4455):
//
//	"Haste
//	 Whenever Living Laser attacks, for each card you've discarded this
//	 turn, create a token that's a copy of Living Laser, except the token
//	 isn't legendary. The tokens enter tapped and attacking. Exile the
//	 tokens at the beginning of the next end step."
//
// Loki, the Deceiver's token copy (msc16AttackingNonlegendaryCopies), made
// once for each card its controller discarded this turn (the discard
// tally), all in one batch, with the exile scheduled over the tokens that
// landed. The tokens were never declared as attackers (CR 508.4). The copy
// is of Living Laser as it last existed if he has left by resolution
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fd15e02a-dd1e-42ee-8c69-a3e5bcb5ede6",
		Name:            "Living Laser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, msc6ThisAttacks,
				"Living Laser — for each card you've discarded this turn, a tapped and attacking token copy of Living Laser; exile the tokens at the next end step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := g.TurnTallyFor(item.Controller).CardsDiscarded
					return msc16AttackingNonlegendaryCopies(ctx, item, item.SourceCardID, n, "",
						"Living Laser", livingLaserExile)
				}),
		},
	})
}

// livingLaserExile is the tokens' tail, keyed: the end step trigger that
// exiles them. An on-disk key.
var livingLaserExile = msc16EndStepForTokens("attacking-tokens/living-laser-exile-the-tokens",
	"Living Laser — exile the tokens", exileListedCardsBody)
