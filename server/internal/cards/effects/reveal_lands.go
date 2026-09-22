package effects

import (
	"fmt"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_lands.go — the ten Shadows over Innistrad / Strixhaven
// "reveal-lands", all of them one sentence:
//
//	"As this land enters, you may reveal an <A> or <B> card from your
//	 hand. If you don't, this land enters tapped."
//	"{T}: Add {X} or {Y}."
//
// Six of them (Choked Estuary, Foreboding Ruins, Port Town, Fortified
// Village, Frostboil Snarl, Furycalm Snarl) were the first skips on
// the "Reveal-from-hand entry choice" seam row, filed in batch 02
// (#295) under the census's own name for the gap — *a card choice
// inside a replacement effect* — and the diagnosis there was right:
// the replacement pipeline was synchronous with no per-card prompt.
// Four more (Game Trail, Shineshadow Snarl, Necroblossom Snarl,
// Vineglimmer Snarl) joined from batches 03 and 04. #1198 built the
// prompt; this file is the cycle spending it.
//
// # Why this is not a checkland with a different question
//
// A checkland's condition is read off the BOARD
// (SelfEntersTappedUnless, tapland_helpers.go), so it contributes no
// applicable replacement at all when the board already satisfies it
// and nobody is asked anything. A reveal-land's condition is a
// DECISION, and one the player can decline even holding a card that
// would satisfy it — bluffing an empty hand is a real play, and it is
// the reason the prompt has a decline rather than an automatic
// reveal. So the replacement is always applicable and the question
// lives inside it, which is EntersTappedUnlessYouPayLife's shape with
// a card where the shockland has a number.
//
// # The reveal is free, and it is not a cost
//
// Nothing moves (CR 701.20b): the named card is still in hand when
// the land has finished entering. It is shown to the whole table and
// every seat is entitled to remember it (CR 701.20), which the engine
// does through the one reveal primitive — so the OTHER seats read
// what was shown in the log rather than in the picker, which is the
// owner's alone.
//
// # The type test is the PRINTED one
//
// "An Island or Swamp card" is a test on a card in a HAND, where no
// layer has been applied. A Dryad of the Ilysian Grove or an Urborg
// on the battlefield changes what the lands in PLAY are and says
// nothing about the card being revealed, so IsLandWithSubtype reading
// the type line is both simpler and correct.
//
// No simplifications.

// EntersTappedUnlessYouRevealFromHand is "as this permanent enters,
// you may reveal a <matching> card from your hand. If you don't, it
// enters tapped."
//
// `clause` is the card's own words for what may be revealed ("an
// Island or Swamp card"), used to build the prompt header, and
// `matches` is the same clause as a predicate over a card in hand.
//
// The Controller hook names the revealer the way the shockland's
// names the payer, through the same EnteringPermanentChooser
// (helpers.go): the player performing the play (ev.Actor, stamped by
// the land path) first, then the card's own controller / owner for an
// entry driven by something else — a fetchland cracking for it, a
// Farseek, a reanimation.
func EntersTappedUnlessYouRevealFromHand(
	name, clause string,
	matches func(game.Card) bool,
) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		Label:           name,
		PromptQuestion:  fmt.Sprintf("%s — reveal %s from your hand so it enters untapped?", name, clause),
		EntryHandReveal: &game.EntryHandReveal{
			Matches: matches,
			Min:     0,
			Max:     1,
		},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Controller: EnteringPermanentChooser,
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
	}
}

// revealLandClause turns the two land types a reveal-land names into
// the printed phrase and the predicate, so the prompt copy and the
// candidate filter can never drift apart.
//
// Needles MUST be lowercase — containsFoldASCII folds the haystack
// and not the needle, the trap tapland_helpers.go already documents
// on youControlLandTyped.
func revealLandClause(a, b string) (string, func(game.Card) bool) {
	first, second := IsLandWithSubtype(a), IsLandWithSubtype(b)
	article := "a"
	if strings.ContainsRune("aeiou", rune(a[0])) {
		article = "an"
	}
	clause := fmt.Sprintf("%s %s or %s card", article, revealLandTypeName(a), revealLandTypeName(b))
	return clause, func(c game.Card) bool { return first(c) || second(c) }
}

// revealLandTypeName upper-cases a land type for the prompt
// ("island" → "Island"). The predicates want lowercase and the player
// wants the printed capital, so the two spellings are derived from
// one input here rather than written twice per card.
func revealLandTypeName(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func init() {
	for _, land := range []struct{ oracleID, name, subA, subB, a, b string }{
		{"d473b507-8c33-4118-bc10-b0a268776074", "Choked Estuary", "island", "swamp", "U", "B"},
		{"5c87e2fa-77f1-4978-b25f-f14d227301d1", "Foreboding Ruins", "swamp", "mountain", "B", "R"},
		{"458d2b12-f578-4392-98d3-c3bc83f316c4", "Port Town", "plains", "island", "W", "U"},
		{"56f1a16a-9f41-41fb-b580-c200bca27cd6", "Fortified Village", "forest", "plains", "G", "W"},
		{"7137aae6-260d-41de-8b4e-42a8cf752697", "Frostboil Snarl", "island", "mountain", "U", "R"},
		{"651dea9c-2375-4e44-8e65-ba8e40f0c0ef", "Furycalm Snarl", "mountain", "plains", "R", "W"},
		{"00de57d2-7cb6-4337-9bc6-f6711e4dfabf", "Game Trail", "mountain", "forest", "R", "G"},
		{"c9fc13d6-bd10-47bc-b2b6-7f67a1f3371e", "Shineshadow Snarl", "plains", "swamp", "W", "B"},
		{"761ee6f9-b0fa-43c9-8d1f-9591ea18e52d", "Necroblossom Snarl", "swamp", "forest", "B", "G"},
		{"33f52df8-4b44-4422-8b0a-37fead9c894b", "Vineglimmer Snarl", "forest", "island", "G", "U"},
	} {
		clause, matches := revealLandClause(land.subA, land.subB)
		Register(Spec{
			OracleID:     land.oracleID,
			Name:         land.name,
			Completeness: CompletenessCaveats,
			Caveats: []string{
				"A reveal-land a spell PUTS onto the battlefield out of a hand or library — Genesis Wave, Coiling Oracle, Arboreal Grazer — always enters tapped. Playing it as a land, or fetching it with a search, does offer the reveal.",
			},
			Replacements: []game.ReplacementEffect{
				EntersTappedUnlessYouRevealFromHand(land.name, clause, matches),
			},
			ManaAbilities: []ManaAbility{dualManaAbility(land.a, land.b)},
		})
	}
}
