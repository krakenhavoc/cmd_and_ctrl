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
// SANDBOX SIMPLIFICATIONS, and this is why the card is not Full:
//
//   - The Spirit token's own clause — "can't block or be blocked by
//     non-Spirit creatures" — is not implemented, and since ADR 0083
//     (#1248) the reason is no longer that a token cannot carry an
//     ability. It can: the Goblin Shaman and the Pest do. This clause
//     is a PAIR RULE, `game.BlockRule.Pair`, and `CardDef` has no
//     `BlockRules` slot for any object to declare one — ADR 0045's
//     addendum, PR 4's card half, tracked on the "Conditional blocking
//     restrictions" row of docs/engine-seams.md (#750). #1248 fixed
//     the half that was this card's: `forEachBlockRuleLocked` used to
//     skip every token by construction, so a token could not have
//     carried the rule even once the Spec field exists. The token
//     itself — a 1/1 colorless Spirit — is created in full; only its
//     restriction text is missing, which is weaker than printed
//     (#259), and the day `Spec.BlockRules` lands this is one
//     `tokenTemplate` away.
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
			"The Spirit token doesn't have its \"can't block or be blocked by non-Spirit creatures\" ability.",
			"\"Exhaust — Waterbend {20}: Take an extra turn after this one\" isn't implemented — extra turns don't exist in the engine yet.",
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(nil, "Avatar Kuruk — create a 1/1 colorless Spirit",
				Do(CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: 1})),
		},
	})
}
