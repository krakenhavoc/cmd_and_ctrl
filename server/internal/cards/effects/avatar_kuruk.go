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
// "EXHAUST — WATERBEND {20}: TAKE AN EXTRA TURN AFTER THIS ONE." The
// cost is `WaterbendCost("{20}")` with `Exhaust: true` (#1310): the same
// component Aang's "Waterbend {8}" and Katara's "Waterbend {X}" use, so
// the artifacts and creatures tapped to help each pay for {1}, and the
// ability can be activated once per object for the rest of the game.
// The effect is TakeExtraTurn (CR 500.7, ADR 0059 Decision 5, #753).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     theLegendOfKurukOracleID + "#1",
		Name:         "Avatar Kuruk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(nil, "Avatar Kuruk — create a 1/1 colorless Spirit",
				Do(CreateToken{Template: kurukSpiritToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — Waterbend {20}: Take an extra turn after this one.",
			Cost:    WaterbendCost("{20}"),
			Exhaust: true,
			Effect:  youTakeAnExtraTurnEffect,
		}},
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
