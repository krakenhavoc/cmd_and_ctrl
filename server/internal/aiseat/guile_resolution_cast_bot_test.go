package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// guile_resolution_cast_bot_test.go — #4286, through the real engine, the
// real enumerator and the real heuristic. Guile's free play is a
// resolution_cast prompt asked inside the countering spell's resolution
// (CR 608.2g), for every counter the engine has. The seat that controls
// Guile must answer it (a seat that owes a choice and never takes a move
// wedges the table, #544), and the engine must accept what it chose
// (botAnswersTheDoor, resolution_door_cast_bot_test.go). Not behind
// AISEAT_GAME_TESTS: a few decisions each.

const (
	botGuileOracle        = "591bbf84-259d-4885-8b5f-b29c04222d41"
	botCounterspellOracle = "cc187110-1148-4090-bbb8-e205694a39f5"
	botSpellPierceOracle  = "d64b0848-0193-4025-ba62-63ecd8fb9f50"
	botHinderOracle       = "c9db6b94-a7b1-4b93-b454-4dead8f85e34"
	botRimeshieldOracle   = "ec47cf67-2580-464f-8118-7eabea5be11c"
	botDoomBladeOracle    = "59e7f2ae-4535-4191-98be-3e65b6b2befa"
)

// guileTable is a settled two-seat table in the first main phase with
// empty hands: `caster` (the active seat) has put a spell on the
// stack and `bot` (the other seat) controls Guile and holds priority.
func guileTable(t *testing.T, seed uint64, typeLine string) (g *game.Game, caster, bot *game.Player, bear uuid.UUID) {
	t.Helper()
	g = newSettledTable(t, seed)
	caster = g.Seats[g.Turn.ActiveSeat]
	bot = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToStep(t, g, game.StepPrecombatMain)
	for _, p := range g.Seats {
		p.Hand.Cards = nil
		p.Command.Cards = nil
	}
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Guile", OracleID: botGuileOracle,
		TypeLine: "Creature — Elemental Incarnation", Power: 6, Toughness: 6, Owner: bot.ID, Controller: bot.ID})
	bear = uuid.New()
	b := game.Card{InstanceID: bear, Name: "Their Spell", TypeLine: typeLine, ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: caster.ID, Controller: caster.ID}
	b.AddKnowersAll(seatIDs(g))
	caster.Hand.PushTop(b)
	if err := g.CastSpell(caster.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the spell: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	return g, caster, bot, bear
}

// botCasts puts a catalog instant in bot's hand and casts it at a target.
func botCasts(t *testing.T, g *game.Game, bot *game.Player, name, oracle string, target game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: name, OracleID: oracle, TypeLine: "Instant", Owner: bot.ID, Controller: bot.ID}
	c.AddKnowersAll(seatIDs(g))
	bot.Hand.PushTop(c)
	if err := g.CastSpell(bot.ID, id, game.CastSpellParams{Targets: []game.TargetRef{target}}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// guileBotAnswers drives the table to the bot's offer, lets the heuristic
// answer it, and checks what it chose against where the card is.
func guileBotAnswers(t *testing.T, g *game.Game, bot *game.Player, counterspell, bear uuid.UUID) {
	t.Helper()
	botPassUntilOwed(t, g, bot.ID)
	if !g.Stack.Contains(counterspell) {
		t.Error("the countering spell finished before the bot was asked")
	}
	card := botAnswersTheDoor(t, g, bot.ID, game.MayCastKeywordFreeCast)
	if card != bear {
		t.Fatalf("the offer was over %s, want the spell", card)
	}
	if !g.Stack.Contains(bear) && !g.Exile.Contains(bear) {
		t.Error("the offered card was neither cast nor left in exile")
	}
}

func TestHeuristicSeatAnswersGuilesHardCounterCast(t *testing.T) {
	g, _, bot, bear := guileTable(t, 42861, "Creature — Bear")
	cs := botCasts(t, g, bot, "Counterspell", botCounterspellOracle, game.TargetRef{Kind: game.TargetCard, ID: bear})
	guileBotAnswers(t, g, bot, cs, bear)
	if !g.Stack.Contains(bear) {
		t.Error("the bot did not cast the free spell")
	}
}

func TestHeuristicSeatAnswersGuilesHinderCast(t *testing.T) {
	g, _, bot, bear := guileTable(t, 42862, "Creature — Bear")
	hinder := botCasts(t, g, bot, "Hinder", botHinderOracle, game.TargetRef{Kind: game.TargetCard, ID: bear})
	botPassUntilOwed(t, g, bot.ID)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePutInLibrary && c.Chooser == bot.ID {
			if err := g.ResolvePutInLibrary(c.ID, bot.ID, nil, []uuid.UUID{bear}); err != nil {
				t.Fatalf("Hinder's top-or-bottom: %v", err)
			}
			break
		}
	}
	guileBotAnswers(t, g, bot, hinder, bear)
}

func TestHeuristicSeatAnswersGuilesDeclinedPierceCast(t *testing.T) {
	g, caster, bot, bear := guileTable(t, 42863, "Sorcery")
	pierce := botCasts(t, g, bot, "Spell Pierce", botSpellPierceOracle, game.TargetRef{Kind: game.TargetCard, ID: bear})
	botPassUntilOwed(t, g, caster.ID)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == caster.ID {
			if err := g.ResolvePayUnless(c.ID, caster.ID, false); err != nil {
				t.Fatalf("declining the tax: %v", err)
			}
			break
		}
	}
	guileBotAnswers(t, g, bot, pierce, bear)
}

// A ward is a triggered ability of the warded permanent's controller, so
// it is that seat's Guile that replaces the counter it makes.
func TestHeuristicSeatAnswersGuilesWardCast(t *testing.T) {
	g := newSettledTable(t, 42864)
	caster := g.Seats[g.Turn.ActiveSeat]
	bot := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToStep(t, g, game.StepPrecombatMain)
	for _, p := range g.Seats {
		p.Hand.Cards = nil
		p.Command.Cards = nil
	}
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Guile", OracleID: botGuileOracle,
		TypeLine: "Creature — Elemental Incarnation", Power: 6, Toughness: 6, Owner: bot.ID, Controller: bot.ID})
	giant := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: giant, Name: "Rimeshield Frost Giant", OracleID: botRimeshieldOracle,
		TypeLine: "Creature — Giant Warrior", Power: 3, Toughness: 3, Owner: bot.ID, Controller: bot.ID})
	blade := uuid.New()
	c := game.Card{InstanceID: blade, Name: "Doom Blade", OracleID: botDoomBladeOracle, TypeLine: "Instant",
		ManaCost: "{1}{B}", Owner: caster.ID, Controller: caster.ID}
	c.AddKnowersAll(seatIDs(g))
	caster.Hand.PushTop(c)
	if err := g.CastSpell(caster.ID, blade, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}},
	}); err != nil {
		t.Fatalf("cast Doom Blade at the warded giant: %v", err)
	}
	botPassUntilOwed(t, g, caster.ID)
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoicePayUnless && ch.Chooser == caster.ID {
			if err := g.ResolvePayUnless(ch.ID, caster.ID, false); err != nil {
				t.Fatalf("declining the ward tax: %v", err)
			}
			break
		}
	}
	botPassUntilOwed(t, g, bot.ID)
	card := botAnswersTheDoor(t, g, bot.ID, game.MayCastKeywordFreeCast)
	if card != blade {
		t.Fatalf("the offer was over %s, want the Doom Blade", card)
	}
	if !g.Stack.Contains(blade) && !g.Exile.Contains(blade) {
		t.Error("the offered card was neither cast nor left in exile")
	}
}
