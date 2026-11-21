package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Romana II — Legendary Creature — Time Lord Scientist {3}{W}, 3/3 (Doctor
// Who, #5948):
//
//	"Vigilance
//	 {1}, {T}: Create a tapped token that's a copy of target token that
//	 entered this turn.
//	 Doctor's companion"
//
// Vigilance is a printed keyword. The target is a token that entered the
// battlefield this turn (the turn's entry tally), judged on announcement and
// again as the ability resolves; the copy is a token created tapped (CR
// 707.2), made from the target's copiable values (or its last-known ones if
// it has left, CR 608.2h). Doctor's companion is a deck-construction rule.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3346a585-d35a-4f8b-b9ab-7e453ebd15f4",
		Name:            "Romana II",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Tokens: 1},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Create a tapped token that's a copy of target token that entered this turn.",
			Purpose: game.Purpose{Tokens: 1},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPermanent("target token that entered this turn", func(g *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsToken() && g.EnteredThisTurn(c.InstanceID)
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return CreateTokenCopy{
					Controller: item.Controller,
					Copy:       id,
					N:          1,
					Except:     func(t *game.Card) { t.Tapped = true },
				}.Apply(ctx)
			},
		}},
	})
}
