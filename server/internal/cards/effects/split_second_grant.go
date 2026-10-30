package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// split_second_grant.go — #3502: the static that gives split second
// (CR 702.61) to spells, declared in Spec.SplitSecondGrants. The engine
// half is game/split_second.go.

// InstantAndSorcerySpellsYouControlHaveSplitSecond is Samut, Tyrant of
// Naktamun's static: "Instant and sorcery spells you control have split
// second." Read live off the battlefield (CR 611.3a), not stamped when a
// spell is cast.
func InstantAndSorcerySpellsYouControlHaveSplitSecond() game.SplitSecondGrant {
	return game.SplitSecondGrant{
		Spells: game.PermissionFilter{InstantOrSorceryOnly: true},
		Label:  "Instant and sorcery spells you control have split second.",
	}
}

// EachSpellYouCastHasSplitSecondIfManaFrom is "each spell you cast has
// split second if mana from a <kind> was spent to cast it" (Shadow the
// Hedgehog's Chaos Control, #4724): every spell its controller casts,
// read off the spell's payment record (game.SplitSecondGrant.SpentFrom,
// #1212's source snapshot), and still read live off the battlefield. A
// card using it declares ReadsManaSpent (ReadsYourCasts, plus
// ReadsSourceKinds for a kind a paper payment can satisfy), so a waived
// payment asks what paid it.
func EachSpellYouCastHasSplitSecondIfManaFrom(kinds game.ManaSourceKinds, label string) game.SplitSecondGrant {
	return game.SplitSecondGrant{
		YouCast:   true,
		SpentFrom: kinds,
		Label:     label,
	}
}
