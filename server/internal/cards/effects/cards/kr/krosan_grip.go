package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krosan Grip — Instant {2}{G}:
//
//	"Split second (As long as this spell is on the stack, players
//	 can't cast spells or activate abilities that aren't mana
//	 abilities.)
//	 Destroy target artifact or enchantment."
//
// The point of this card over Naturalize is split second (CR 702.61),
// and since #1519 the engine reads it: the keyword is declared below,
// the cast path stamps it on the stack item, and while the Grip is on
// the stack nobody can cast a spell or activate an ability that isn't
// a mana ability. A Sol Ring still taps for mana, a morph can still be
// turned face up (a special action, CR 702.61b) and a Blood Artist
// still triggers — but nobody can flash in an answer or re-equip a
// Lightning Greaves onto the target. From S14 until #1519 it shipped
// as an ordinary instant with a caveat saying so.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3e39224c-72ce-4ecc-aa17-12c071ea1f3e",
		Name:            "Krosan Grip",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
