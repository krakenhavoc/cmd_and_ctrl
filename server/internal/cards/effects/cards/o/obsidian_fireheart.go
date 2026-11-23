package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obsidian Fireheart — Creature — Elemental (4/4) for {1}{R}{R}{R}:
//
//	"{1}{R}{R}: Put a blaze counter on target land without a blaze counter on it. For as long as that land has a blaze counter on it, it has "At the beginning of your upkeep, this land deals 1 damage to you." (The land continues to burn after this creature has left the battlefield.)"
//
// ADR 0109 §2 (#1604): the blaze counter grants the land an upkeep
// trigger, a granted ability bundle (ADR 0093 PR 4), for as long as it
// has a blaze counter on it. The trigger is the land's: "your upkeep"
// and "you" are its controller's, and the land is the source of the
// damage. The Fireheart leaving changes nothing, as the reminder text
// says.
//
// No simplification.
const obsidianFireheartBlaze = "obsidian-fireheart/blaze"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3b3e070-28f6-4852-816e-caffe358b6fe",
		Name:         "Obsidian Fireheart",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  obsidianFireheartBlaze,
			Text: "At the beginning of your upkeep, this land deals 1 damage to you.",
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Obsidian Fireheart — blaze: this land deals 1 damage to you",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
					}),
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}: Put a blaze counter on target land without a blaze counter on it. For as long as that land has a blaze counter on it, it has \"At the beginning of your upkeep, this land deals 1 damage to you.\" (The land continues to burn after this creature has left the battlefield.)",
			Cost:    effects.ManaCost("{1}{R}{R}"),
			Targets: effects.TargetPermanent("target land without a blaze counter on it", effects.Land(), obsidianFireheartNoBlaze()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GrantWhileItHasCounter(ctx, effects.FirstLegalBattlefieldTarget(ctx), "blaze",
					"Obsidian Fireheart — blaze while it has a blaze counter", obsidianFireheartBlaze)
			},
		}},
	})
}

// obsidianFireheartNoBlaze is "without a blaze counter on it".
func obsidianFireheartNoBlaze() effects.CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters["blaze"] == 0 }
}
