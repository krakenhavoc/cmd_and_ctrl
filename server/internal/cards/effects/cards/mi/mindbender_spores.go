package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindbender Spores — Creature — Fungus Wall {2}{G}, 0/1 (Mirage):
//
//	"Defender
//	 Flying
//	 Whenever this creature blocks a creature, put four fungus counters on that
//	 creature. The creature gains "This creature doesn't untap during your untap
//	 step if it has a fungus counter on it" and "At the beginning of your upkeep,
//	 remove a fungus counter from this creature.""
//
// The four counters go on the blocked creature as the trigger resolves, if it
// is still on the battlefield, and it gains the two abilities as one bundle with
// no duration, so for as long as it stays that object (CR 611.2a). The untap
// restriction reads the creature's own fungus counters as its controller's untap
// step runs; the upkeep trigger is the creature's controller's. A creature the
// Spores blocks twice, or that two Spores block, gets the bundle twice and so
// loses two counters each upkeep, as two such triggers would (CR 603.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0e3b7ec1-3760-4ac4-81ef-1ab389392f1b",
		Name:            "Mindbender Spores",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Grants: []effects.AbilityGrant{{
			Key: "mindbender-spores/fungus",
			UntapStepRestrictions: []game.UntapStepRestriction{{
				Label: "This creature doesn't untap during your untap step if it has a fungus counter on it.",
				Restricts: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID && target.Counters[effects.OldB804FungusCounter] > 0
				},
			}},
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Mindbender Spores — remove a fungus counter from this creature", effects.RemoveACounterFromThis(effects.OldB804FungusCounter)),
			},
			Text: "This creature doesn't untap during your untap step if it has a fungus counter on it. At the beginning of your upkeep, remove a fungus counter from this creature.",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBlock},
			Key:     "Mindbender Spores — put four fungus counters on the blocked creature, which gains the fungus abilities",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				blocked := item.Trigger.Event.Target
				if !effects.OnBattlefield(g, blocked) {
					return nil
				}
				if err := (effects.AddCounter{Target: blocked, Kind: effects.OldB804FungusCounter, N: 4}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantAbilitiesFor{
					Target:   blocked,
					Keys:     []string{"mindbender-spores/fungus"},
					Duration: game.IndefiniteDuration(),
					Label:    "Mindbender Spores — the creature gains the fungus abilities",
				}.Apply(ctx)
			},
		}},
	})
}
