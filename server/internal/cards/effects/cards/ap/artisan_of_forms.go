package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Artisan of Forms — Creature — Human Wizard {1}{U}, 1/1 (Theros, S82 OLD b388):
//
//	"Heroic — Whenever you cast a spell that targets this creature, you may have
//	 this creature become a copy of target creature, except it has this ability."
//
// The trigger targets as it goes on the stack and asks "you may" as it resolves.
// The Artisan becomes a copy of the target with no duration (CR 611.2a; a
// BecomeCopy of kind CopyIndefinite, layer 1), and the exception (CR 707.9)
// gives the copy this same heroic ability as a granted bundle, so it can do it
// again whatever it copied. A target that left the battlefield, or an Artisan
// that left and came back (CR 400.7), copies nothing.
//
// No simplification.
func init() {
	const grant = "artisan-of-forms/heroic-copy"
	const label = "Artisan of Forms — you may have this creature become a copy of target creature, except it has this ability (heroic)"
	build := func() game.TriggeredAbility {
		return effects.Targeting(
			effects.Optional(effects.Heroic(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BecomeCopy{
					Targets:  []uuid.UUID{item.SourceCardID},
					Of:       id,
					Duration: effects.CopyIndefinite,
					Except:   func(v *game.PrintedValues) { v.GrantAbility(grant) },
					Label:    "Artisan of Forms — becomes a copy",
				}.Apply(ctx)
			}), "Artisan of Forms — have it become a copy of the target creature?"),
			effects.TargetCreature("target creature"))
	}
	effects.Register(effects.Spec{
		OracleID:     "3e8abec7-1e18-460a-9d88-30969755ebd2",
		Name:         "Artisan of Forms",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{build()},
		Grants: []effects.AbilityGrant{{
			Key:       grant,
			Triggered: []game.TriggeredAbility{build()},
			Text:      "Heroic — Whenever you cast a spell that targets this creature, you may have this creature become a copy of target creature, except it has this ability.",
		}},
	})
}
