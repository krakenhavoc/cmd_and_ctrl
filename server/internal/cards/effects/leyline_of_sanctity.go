package effects

// Leyline of Sanctity — Enchantment {2}{W}{W}:
//
//	"If this card is in your opening hand, you may begin the game with
//	 it on the battlefield.
//	 You have hexproof."
//
// The card the player-hexproof seam (#1197) exists for, and the
// reason it is one line here: "You have hexproof" is a printed static
// of a permanent, so it is DERIVED. Spec.PlayerKeywords declares it,
// the engine asks the battlefield on every targeting query, and
// nothing is written to the player at all — which is what makes two
// Leylines compose and one of them leaving not revoke the other's
// grant. See ADR 0072's 2026-09-22 amendment and
// game/player_statics.go.
//
// What it stops: every targeted burn spell, every "target player
// discards", every Vampiric Tutor pointed the wrong way — an
// OPPONENT's, always (CR 702.11d). Your own spells still reach you,
// which is not a nicety: a Leyline that stopped you casting your own
// Sylvan Library would be a worse card than the one that is printed,
// and the asymmetry is the whole of the keyword.
//
// What it does NOT stop, and nobody should expect it to: an attack, a
// board wipe, an edict, a "each player sacrifices", or anything else
// that does not use the word "target". Hexproof is a targeting rule.
//
// The opening-hand clause (CR 103.6a) is Spec.OpeningHand (ADR 0133):
// "If this card is in your opening hand, you may begin the game with
// it on the battlefield." The seat holding it is asked as the mulligan
// window closes, and a yes puts the Leyline onto the battlefield
// through the ordinary hand door, so its hexproof is in force before
// anyone has acted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "492e0e6c-8c27-4376-938b-f8a8b6205810",
		Name:           "Leyline of Sanctity",
		Completeness:   CompletenessFull,
		OpeningHand:    BeginTheGameOnTheBattlefield(),
		PlayerKeywords: []string{KeywordHexproof},
	})
}
