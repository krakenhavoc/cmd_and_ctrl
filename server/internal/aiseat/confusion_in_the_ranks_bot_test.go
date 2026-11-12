package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// confusion_in_the_ranks_bot_test.go — #5739, whole-engine: the real
// catalog Confusion in the Ranks, the real view, the real enumerator and
// the real heuristic.
//
// The bot's creature enters under an opponent's Confusion in the Ranks.
// The trigger is the opponent's, but its target is chosen by the bot, as
// the entering creature's controller: the prompt is the bot's, it names
// the opponent in chosen_for, and the bot takes the opponent's better
// creature in the exchange, whichever is listed first.
//
// Not behind AISEAT_GAME_TESTS: one trigger, milliseconds.

const botConfusionOracle = "e477abb9-9b99-48a8-8c09-07b120b72540"

func TestHeuristicTakesTheBetterCreatureWithAnOpponentsConfusionInTheRanks(t *testing.T) {
	for _, bestFirst := range []bool{true, false} {
		name := "mouse listed first"
		if bestFirst {
			name = "titan listed first"
		}
		t.Run(name, func(t *testing.T) {
			g := newSettledTable(t, 5739)
			advanceToStep(t, g, game.StepPrecombatMain)
			bot := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			opp := g.Seats[g.Turn.ActiveSeat]
			pushAbyssCard(g, opp.ID, "Confusion in the Ranks", "Enchantment", botConfusionOracle, 0, 0)
			everyone := seatIDs(g)
			mk := func(name string, p, tough int) game.Card {
				c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
					ManaCost: "{2}{G}", Power: p, Toughness: tough, Owner: opp.ID, Controller: opp.ID}
				c.AddKnowersAll(everyone)
				return c
			}
			big, small := mk("Test Titan", 5, 5), mk("Test Mouse", 1, 1)
			g.WithWriteLock(func() {
				if bestFirst {
					g.Battlefield.PushTop(big)
					g.Battlefield.PushTop(small)
				} else {
					g.Battlefield.PushTop(small)
					g.Battlefield.PushTop(big)
				}
			})

			entering := game.Card{InstanceID: uuid.New(), Name: "Test Goat", TypeLine: "Creature — Goat",
				ManaCost: "{W}", Power: 1, Toughness: 1, Owner: bot.ID, Controller: bot.ID}
			entering.AddKnowersAll(everyone)
			g.WithWriteLock(func() { bot.Hand.PushTop(entering) })
			if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: bot.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, entering.InstanceID); err != nil {
				t.Fatalf("enter: %v", err)
			}
			if chooser, chosenFor := pickPromptFor(t, g, bot.ID); chooser != bot.ID || chosenFor != opp.ID {
				t.Fatalf("prompt for %s on behalf of %s, want the bot choosing for %s", chooser, chosenFor, opp.ID)
			}
			took := driveChoices(t, g, heuristic.New(), bot.ID, 10)
			var targets []game.TargetRef
			g.ReadSnapshot(func() {
				for _, item := range g.StackMeta {
					if item != nil && item.Kind == game.StackItemTriggered {
						targets = item.Targets
					}
				}
			})
			if len(targets) != 1 || targets[0].ID != big.InstanceID {
				t.Errorf("the trigger targets %v, want the opponent's Titan; took %v", targets, took)
			}
		})
	}
}
