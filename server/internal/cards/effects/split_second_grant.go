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
