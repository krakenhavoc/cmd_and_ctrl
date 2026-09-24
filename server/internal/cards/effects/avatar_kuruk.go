package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar Kuruk — Legendary Creature — Avatar, 4/3 (colorless):
//
//	"Whenever you cast a spell, create a 1/1 colorless Spirit creature
//	 token with 'This token can't block or be blocked by non-Spirit
//	 creatures.'
//	 Exhaust — Waterbend {20}: Take an extra turn after this one.
//	 (While paying a waterbend cost, you can tap your artifacts and
//	 creatures to help. Each one pays for {1}. Activate each exhaust
//	 ability only once.)"
//
// The BACK face of The Legend of Kuruk ("<oracle_id>#1", ADR 0034),
// reached only through chapter III's exile-and-return
// (the_legend_of_kuruk.go). Because that verb makes a NEW object
// (CR 400.7), Avatar Kuruk is always summoning sick the turn it
// arrives — printed and correct, not a gap.
//
// # The Spirit token
//
// "This token can't block or be blocked by non-Spirit creatures" is a
// pair rule on the TOKEN (#750, ADR 0045 addendum Decision 11 and its
// 2026-09-24 amendment), declared on its catalog template below and
// read off the battlefield through the token's key like any card's
// Spec.BlockRules. It is two rules, one per side of the pair: a
// non-Spirit can't block the token, and the token can't block a
// non-Spirit. "Spirit" is read as an EFFECTIVE creature type, so a
// changeling is a Spirit for both.
//
// SANDBOX SIMPLIFICATION, and this is why the card is not Full:
//
//   - "Exhaust — Waterbend {20}: Take an extra turn after this one."
//     is not registered, and the reason is now ONE, not two: extra
//     turns have no queue/insertion primitive anywhere in this engine
//     (the identity foundation is present, but the "Extra turns
//     primitive" seam remains in docs/engine-seams.md, #753). The
//     COST is expressible since #1310: `WaterbendCost("{20}")`
//     with `Exhaust: true` is the whole declaration, the same
//     component Aang's "Waterbend {8}" and Katara's "Waterbend {X}"
//     use, and Boom Scholar's exhaust discount already reaches it
//     through the ability's mana component. An ability that costs
//     twenty and does nothing when activated would not be a weaker
//     card, it would be a broken one, so it stays off entirely until
//     #753 lands the effect.
func init() {
	Register(Spec{
		OracleID:     theLegendOfKurukOracleID + "#1",
		Name:         "Avatar Kuruk",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"\"Exhaust — Waterbend {20}: Take an extra turn after this one\" isn't implemented — extra turns don't exist in the engine yet.",
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(nil, "Avatar Kuruk — create a 1/1 colorless Spirit",
				Do(CreateToken{Template: kurukSpiritToken(), N: 1})),
		},
	})
}

// kurukSpiritToken is Avatar Kuruk's 1/1 colorless Spirit with "This
// token can't block or be blocked by non-Spirit creatures."
func kurukSpiritToken() game.Card { return tokenFromCatalog(printedKurukSpiritToken) }

// printedKurukSpiritToken is the Spirit as PRINTED — its block rule
// included. Not the plain "1/1 colorless Spirit" row Forbidden
// Orchard makes: that one prints no text, and a row shared between
// the two would hand Orchard's Spirits a restriction they don't have,
// or strip Kuruk's of the one they do.
func printedKurukSpiritToken() tokenTemplate {
	return tokenTemplate{
		Slug: "spirit-only-spirits",
		Card: game.Card{
			Name:      "Spirit",
			TypeLine:  "Token Creature — Spirit",
			Power:     1,
			Toughness: 1,
		},
		BlockRules: CantBlockOrBeBlockedBy(OnSelf(), Not(OfCreatureType("Spirit")), "non-Spirit creatures"),
		Text:       "This token can't block or be blocked by non-Spirit creatures.",
	}
}
