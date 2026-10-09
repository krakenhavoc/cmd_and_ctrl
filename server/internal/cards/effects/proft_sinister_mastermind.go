package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Proft, Sinister Mastermind — Legendary Creature — Human Rogue {2}{B},
// 5/5:
//
//	"Threshold — You can't cast this spell unless there are seven or
//	 more cards in your graveyard.
//	 Menace
//	 {B}, Discard this card: Target creature gets -3/-1 until end of
//	 turn."
//
// The restriction is a CastCondition on the card itself (the shape
// Rakdos, Lord of Riots uses), checked at announce and never again. The
// discard ability works only from the hand (CR 113.6), the way cycling
// does; the -3/-1 is a modify-P/T effect that locks to the target at
// resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a350856c-87f2-4605-999f-4e26aa362a55",
		Name:            "Proft, Sinister Mastermind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return b31GraveyardSize(g, controller) >= 7
		},
		CastConditionLabel: "You can't cast this spell unless there are seven or more cards in your graveyard.",
		Activated: []ActivatedAbility{{
			Label:   "{B}, Discard this card: Target creature gets -3/-1 until end of turn",
			Cost:    Plus(ManaCost("{B}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return BoostUntilEOT{Target: t.ID, Power: -3, Toughness: -1, Label: "Proft, Sinister Mastermind — -3/-1"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
