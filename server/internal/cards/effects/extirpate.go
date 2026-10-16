package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Extirpate — Instant {B}:
//
//	"Split second (As long as this spell is on the stack, players can't
//	 cast spells or activate abilities that aren't mana abilities.)
//	 Choose target card in a graveyard other than a basic land card.
//	 Search its owner's graveyard, hand, and library for all cards with
//	 the same name as that card and exile them. Then that player
//	 shuffles."
//
// Surgical Extraction's body with two differences. Split second is the
// engine's keyword (CR 702.61, #1519): while Extirpate is on the stack
// nobody casts a spell or activates an ability that isn't a mana ability,
// so the target can't be cast or shuffled away in response. And "all":
// every card with the target's name in its owner's graveyard is exiled,
// the target among them, because a graveyard is a public zone and has no
// fail to find (CR 701.23b); the caster may still leave matches in the
// hand and the library, which are hidden (extractionSearch, #4033).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa857e9d-dfe4-4348-9c1f-16aa240d69cd",
		Name:            "Extirpate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets: TargetCardInGraveyard("target card in a graveyard other than a basic land card",
			notABasicLandCard()),
		OnResolve: extractFromGraveyardTarget("Extirpate", true),
	})
}
