package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sephara, Sky's Blade — Legendary Creature — Angel {4}{W}{W}{W}, 7/7:
//
//	"You may pay {W} and tap four untapped creatures you control with
//	 flying rather than pay this spell's mana cost.
//	 Flying, lifelink
//	 Other creatures you control with flying have indestructible."
//
// The alternative cost is ADR 0135 §1's tap cost with a mana half
// (TapInsteadPaying): the four fliers are untapped, yours and not
// targeted, may have arrived this turn (CR 302.6, the 2019-07-12 ruling),
// and are tapped with Sephara on the stack. Commander tax is added to the
// {W} (CR 118.9d).
//
// The grant reads flying, which layer 6 adds and removes, so it is applied
// after every effect that would give one of your creatures flying or take
// it away (CR 613.8a, 613.8b; WithKeywordHave, #2663): a creature that
// gains flying from a later Equipment or spell has indestructible, and one
// that loses flying to a later Gravity Sphere does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ac8952c9-bd30-48cd-b964-7e6271f9cbc8",
		Name:            "Sephara, Sky's Blade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		AlternativeCosts: []game.AlternativeCost{
			TapInsteadPaying("{W}", 4, "four untapped creatures you control with flying", Creature(), HasKeyword("flying")),
		},
		Static: []game.StaticAbility{
			WithKeywordHave(OtherCreaturesYouControl, "flying",
				"Other creatures you control with flying have indestructible.", "indestructible"),
		},
	})
}
