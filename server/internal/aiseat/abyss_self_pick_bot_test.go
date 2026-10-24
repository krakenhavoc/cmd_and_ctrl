package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// abyss_self_pick_bot_test.go — #4249, whole-engine: the real catalog
// The Abyss and Magus of the Abyss, the real enumerator and the real
// heuristic.
//
// "Destroy target nonartifact creature that player controls of their
// choice" is picked by the player whose upkeep it is. On its controller's
// own upkeep the controller picks among its own creatures, and the prompt
// carries no `chosen_for` (the chooser is the controller). The heuristic
// scored every permanent it controls the same, so it destroyed whichever
// creature was listed first, its best one included. The clause is now
// declared Harmful (effects.Harmful), which the wire carries as
// pick_target.harmful, and the bot gives up the least valuable creature.
//
// Not behind AISEAT_GAME_TESTS: one upkeep, milliseconds.

const (
	botAbyssOracle = "a02a7816-c967-4503-bb08-f8db44915250"
	botMagusOracle = "66958c48-3ac7-4f41-aa3e-adacaa24ea4f"
)

// abyssTable parks a two-seat game just before the bot's upkeep. The bot
// is the seat that is not active when the table settles; it controls a
// 5/5 and a 1/1, pushed in the given order so a pick that falls back on
// "first listed" is caught both ways round. Returns the bot, its
// opponent and the two creatures.
func abyssTable(t *testing.T, seed uint64, bestFirst bool) (g *game.Game, bot, opp *game.Player, best, least uuid.UUID) {
	t.Helper()
	g = newSettledTable(t, seed)
	advanceToStep(t, g, game.StepPrecombatMain)
	botIdx := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	bot, opp = g.Seats[botIdx], g.Seats[g.Turn.ActiveSeat]
	everyone := seatIDs(g)
	mk := func(name string, p, tough int) game.Card {
		c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
			ManaCost: "{2}{G}", Power: p, Toughness: tough, Owner: bot.ID, Controller: bot.ID}
		c.AddKnowersAll(everyone)
		return c
	}
	big, small := mk("Test Titan", 5, 5), mk("Test Mouse", 1, 1)
	best, least = big.InstanceID, small.InstanceID
	g.WithWriteLock(func() {
		if bestFirst {
			g.Battlefield.PushTop(big)
			g.Battlefield.PushTop(small)
		} else {
			g.Battlefield.PushTop(small)
			g.Battlefield.PushTop(big)
		}
	})
	return g, bot, opp, best, least
}

// pushAbyssCard puts a catalog permanent under controller on the
// battlefield. The trigger is found by oracle id when the upkeep begins.
func pushAbyssCard(g *game.Game, controller uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: name, OracleID: oracle, TypeLine: typeLine,
		ManaCost: "{3}{B}", Power: p, Toughness: tough, Owner: controller, Controller: controller}
	c.AddKnowersAll(seatIDs(g))
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

// toUpkeepOf advances to the upkeep of the given seat.
func toUpkeepOf(t *testing.T, g *game.Game, seat uuid.UUID) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Seats[g.Turn.ActiveSeat].ID == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s's upkeep: %v", seat, err)
		}
	}
	t.Fatalf("never reached %s's upkeep", seat)
}

func onBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

// pickPromptFor returns the open pick_target prompt the seat owes.
func pickPromptFor(t *testing.T, g *game.Game, seat uuid.UUID) (chooser, chosenFor uuid.UUID) {
	t.Helper()
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == seat {
				chooser, chosenFor = c.Chooser, c.PickTargetFor
			}
		}
	})
	if chooser == uuid.Nil {
		t.Fatalf("the seat owes no pick_target prompt: %s", openChoices(g, seat))
	}
	return chooser, chosenFor
}

// resolveUpkeep answers the bot's pick with the heuristic, then plays the
// upkeep out so the trigger resolves.
func resolveUpkeep(t *testing.T, g *game.Game, bot *game.Player) []string {
	t.Helper()
	pol := heuristic.New()
	took := driveChoices(t, g, pol, bot.ID, 10)
	playHeliodUntil(t, g, bot, pol, func() bool { return g.Turn.Step != game.StepUpkeep })
	return took
}

// TestHeuristicGivesUpItsLeastCreatureToItsOwnAbyss: The Abyss on the
// bot's own upkeep destroys the lesser creature, whichever is listed first.
func TestHeuristicGivesUpItsLeastCreatureToItsOwnAbyss(t *testing.T) {
	for _, bestFirst := range []bool{true, false} {
		name := "mouse listed first"
		if bestFirst {
			name = "titan listed first"
		}
		t.Run(name, func(t *testing.T) {
			g, bot, _, best, least := abyssTable(t, 4249, bestFirst)
			pushAbyssCard(g, bot.ID, "The Abyss", "World Enchantment", botAbyssOracle, 0, 0)
			toUpkeepOf(t, g, bot.ID)
			if chooser, chosenFor := pickPromptFor(t, g, bot.ID); chooser != bot.ID || chosenFor != uuid.Nil {
				t.Fatalf("prompt for %s on behalf of %s, want the bot choosing for itself", chooser, chosenFor)
			}
			took := resolveUpkeep(t, g, bot)
			if !onBattlefield(g, best) {
				t.Errorf("the bot destroyed its best creature to its own Abyss; took %v", took)
			}
			if onBattlefield(g, least) {
				t.Errorf("the bot kept its least creature and lost something else; took %v", took)
			}
		})
	}
}

// TestHeuristicGivesUpItsLeastCreatureToItsOwnMagusOfTheAbyss: the Magus
// is itself a legal pick (4/3); the bot still gives up the 1/1.
func TestHeuristicGivesUpItsLeastCreatureToItsOwnMagusOfTheAbyss(t *testing.T) {
	for _, bestFirst := range []bool{true, false} {
		name := "mouse listed first"
		if bestFirst {
			name = "titan listed first"
		}
		t.Run(name, func(t *testing.T) {
			g, bot, _, best, least := abyssTable(t, 4250, bestFirst)
			magus := pushAbyssCard(g, bot.ID, "Magus of the Abyss", "Creature — Human Wizard", botMagusOracle, 4, 3)
			toUpkeepOf(t, g, bot.ID)
			if chooser, chosenFor := pickPromptFor(t, g, bot.ID); chooser != bot.ID || chosenFor != uuid.Nil {
				t.Fatalf("prompt for %s on behalf of %s, want the bot choosing for itself", chooser, chosenFor)
			}
			took := resolveUpkeep(t, g, bot)
			if !onBattlefield(g, best) || !onBattlefield(g, magus) {
				t.Errorf("the bot destroyed its Titan or its Magus rather than the Mouse; took %v", took)
			}
			if onBattlefield(g, least) {
				t.Errorf("the bot kept its least creature and lost something else; took %v", took)
			}
		})
	}
}

// TestHeuristicGivesUpItsLeastCreatureToAnOpponentsAbyss: the opponent's
// Abyss on the bot's upkeep names the opponent in chosen_for, which the
// heuristic already priced; it is unchanged.
func TestHeuristicGivesUpItsLeastCreatureToAnOpponentsAbyss(t *testing.T) {
	for _, bestFirst := range []bool{true, false} {
		name := "mouse listed first"
		if bestFirst {
			name = "titan listed first"
		}
		t.Run(name, func(t *testing.T) {
			g, bot, opp, best, least := abyssTable(t, 4251, bestFirst)
			pushAbyssCard(g, opp.ID, "The Abyss", "World Enchantment", botAbyssOracle, 0, 0)
			toUpkeepOf(t, g, bot.ID)
			if chooser, chosenFor := pickPromptFor(t, g, bot.ID); chooser != bot.ID || chosenFor != opp.ID {
				t.Fatalf("prompt for %s on behalf of %s, want the bot choosing for %s", chooser, chosenFor, opp.ID)
			}
			took := resolveUpkeep(t, g, bot)
			if !onBattlefield(g, best) {
				t.Errorf("the bot destroyed its best creature to the opponent's Abyss; took %v", took)
			}
			if onBattlefield(g, least) {
				t.Errorf("the bot kept its least creature and lost something else; took %v", took)
			}
		})
	}
}
