package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cryptoplasm — Creature — Shapeshifter {1}{U}{U}, 2/2 (Mirrodin Besieged,
// S82 OLD b467, #5576):
//
//	"At the beginning of your upkeep, you may have this creature become a copy
//	 of another target creature, except it has this ability."
//
// Artisan of Forms' copy on an upkeep trigger. The trigger targets another
// creature as it goes on the stack (never itself, CR 115.4) and asks "you may"
// (CR 603.5). The Cryptoplasm becomes a copy of the target with no duration
// (CR 611.2a; a copy effect in layer 1), and the exception (CR 707.9) gives
// the copy this same upkeep ability as a granted bundle, so it can do it
// again whatever it copied. A target that has gone, or a Cryptoplasm that left
// and came back as a new object (CR 400.7), copies nothing.
//
// No simplification.
func init() {
	const grant = "cryptoplasm/upkeep-copy"
	const label = "Cryptoplasm — you may have this creature become a copy of another target creature, except it has this ability"
	build := func() game.TriggeredAbility {
		return effects.Targeting(
			effects.Optional(effects.AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
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
					Label:    "Cryptoplasm — becomes a copy",
				}.Apply(ctx)
			}), "Cryptoplasm — have it become a copy of another target creature?"),
			effects.Another(effects.TargetCreature("another target creature")))
	}
	effects.Register(effects.Spec{
		OracleID:     "e7bef318-2720-45bc-b663-aafd724bc511",
		Name:         "Cryptoplasm",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{build()},
		Grants: []effects.AbilityGrant{{
			Key:       grant,
			Triggered: []game.TriggeredAbility{build()},
			Text:      "At the beginning of your upkeep, you may have this creature become a copy of another target creature, except it has this ability.",
		}},
	})
}
