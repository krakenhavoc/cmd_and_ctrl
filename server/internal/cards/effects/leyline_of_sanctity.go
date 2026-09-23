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
// ONE SIMPLIFICATION, the Gemstone Caverns one. "If this card is in
// your opening hand, you may begin the game with it on the
// battlefield" has no shape: there is no pre-game window in which a
// card in an opening hand can be put onto the battlefield, and the
// mulligan flow (Game.KeepHand) has no hook for one. So the Leyline
// is cast for its four mana like any other enchantment. That is
// strictly WEAKER than printed — you pay for it and you pay a turn
// for it — which is the only direction a simplification may go
// (#259), and it is the same gap gemstone_caverns.go already carries
// a caveat for.
func init() {
	Register(Spec{
		OracleID:     "492e0e6c-8c27-4376-938b-f8a8b6205810",
		Name:         "Leyline of Sanctity",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Starting the game with this on the battlefield from your opening hand isn't implemented — you cast it for {2}{W}{W} like an ordinary enchantment.",
		},
		PlayerKeywords: []string{KeywordHexproof},
	})
}
