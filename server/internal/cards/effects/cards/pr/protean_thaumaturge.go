package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protean Thaumaturge — Creature — Human Wizard {1}{U}, 1/1:
//
//	"Constellation — Whenever an enchantment you control enters, you may have
//	 this creature become a copy of another target creature, except it has
//	 this ability."
//
// The copy lasts indefinitely (CR 707.9a); the granted copy of this ability
// is the "except" clause, so the new creature has it as well as whatever it
// copied.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "148331cc-24de-481f-9131-c3089d1ff7ce",
		Name:         "Protean Thaumaturge",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       proteanThaumaturgeGrant,
			Text:      proteanThaumaturgeText,
			Triggered: []game.TriggeredAbility{proteanThaumaturgeTrigger()},
		}},
		Triggered: []game.TriggeredAbility{proteanThaumaturgeTrigger()},
	})
}

const (
	proteanThaumaturgeLabel = "Protean Thaumaturge — you may have this creature become a copy of another target creature (constellation)"
	proteanThaumaturgeText  = "Constellation — Whenever an enchantment you control enters, you may have this creature become a copy of another target creature, except it has this ability."
	// proteanThaumaturgeGrant is the "except it has this ability" bundle, an
	// on-disk identity: never renamed, never reused.
	proteanThaumaturgeGrant = "protean-thaumaturge/constellation-copy"
)

func proteanThaumaturgeTrigger() game.TriggeredAbility {
	return effects.Targeting(
		effects.Optional(
			effects.On(game.EventETB, effects.GapB103EnchantmentEnteredUnderYourControl, proteanThaumaturgeLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BecomeCopy{
						Targets:  []uuid.UUID{ctx.Source()},
						Of:       id,
						Duration: effects.CopyIndefinite,
						Except:   func(v *game.PrintedValues) { v.GrantAbility(proteanThaumaturgeGrant) },
						Label:    "Protean Thaumaturge — becomes a copy",
					}.Apply(ctx)
				}),
			"Protean Thaumaturge — have this creature become a copy of the target creature?"),
		effects.Another(effects.TargetCreature("another target creature")),
	)
}
