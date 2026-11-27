package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b516_helpers.go — shared bodies for S82 OLD batch 516 (Shards of
// Alara, #5588).

// printedOldB516SkeletonToken is Skeletonize's Skeleton: a 1/1 black Skeleton
// creature token with "{B}: Regenerate this creature." (the printed token
// card's wording). Its slug is an on-disk identity: never renamed, never
// reused.
func printedOldB516SkeletonToken() TokenTemplate {
	return TokenTemplate{
		Slug: "skeleton-regenerate",
		Card: game.Card{
			Name:     "Skeleton",
			TypeLine: "Token Creature — Skeleton",
			Power:    1, Toughness: 1,
			Colors: []string{"B"},
		},
		Text: "{B}: Regenerate this creature.",
		Activated: []game.ActivatedAbilityShape{{
			Label:   "{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	}
}

// OldB516SkeletonToken is the regenerating Skeleton's card value.
func OldB516SkeletonToken() game.Card { return TokenFromCatalog(printedOldB516SkeletonToken) }

// OldB516SacrificeThisPermanentAtEndStepGrant is the quoted ability "At the
// beginning of the end step, sacrifice this permanent." as a named bundle for
// a copy effect to grant (Minion Reflector): SacrificeThisTokenAtEndStepGrant
// with the printed word. It triggers at the beginning of every end step, for
// as long as the token has it.
func OldB516SacrificeThisPermanentAtEndStepGrant(key, label string) AbilityGrant {
	g := SacrificeThisTokenAtEndStepGrant(key, label)
	g.Text = "At the beginning of the end step, sacrifice this permanent."
	return g
}

// OldB516ReturnTargetArtifactToBattlefield is "return target artifact card
// from your graveyard to the battlefield" (Protomatter Powder).
func OldB516ReturnTargetArtifactToBattlefield(g *game.Game, item *game.StackItem) error {
	return ReturnFirstLegalGraveyardTargetToBattlefield(g, item)
}
