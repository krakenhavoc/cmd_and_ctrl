package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "c3b3e070-28f6-4852-816e-caffe358b6fe",
		Name:         "Obsidian Fireheart",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  obsidianFireheartBlaze,
			Text: "At the beginning of your upkeep, this land deals 1 damage to you.",
			Triggered: []game.TriggeredAbility{
				AtYourUpkeep("Obsidian Fireheart — blaze: this land deals 1 damage to you",
					func(g *game.Game, item *game.StackItem) error {
						return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					}),
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}{R}: Put a blaze counter on target land without a blaze counter on it. For as long as that land has a blaze counter on it, it has \"At the beginning of your upkeep, this land deals 1 damage to you.\" (The land continues to burn after this creature has left the battlefield.)",
			Cost:    ManaCost("{1}{R}{R}"),
			Targets: TargetPermanent("target land without a blaze counter on it", Land(), obsidianFireheartNoBlaze()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return GrantWhileItHasCounter(ctx, FirstLegalBattlefieldTarget(ctx), "blaze",
					"Obsidian Fireheart — blaze while it has a blaze counter", obsidianFireheartBlaze)
			},
		}},
	})
}

// obsidianFireheartNoBlaze is "without a blaze counter on it".
func obsidianFireheartNoBlaze() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters["blaze"] == 0 }
}
