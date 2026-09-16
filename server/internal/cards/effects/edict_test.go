package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// edict_test.go — "each player sacrifices a creature of their choice"
// (CR 701.21a): Grave Pact, Dictate of Erebos, Fleshbag Marauder,
// Butcher of Malakir.
//
// The assertions that matter are about WHO chooses and WHO is asked,
// because those are the two things a loop-driven implementation gets
// wrong while still reducing everyone's board:
//
//   - each affected player gets their own prompt, so nobody picks for
//     anyone else;
//   - the controller is asked for "each player" and skipped for "each
//     other player";
//   - a player with no creature is skipped rather than left holding a
//     prompt nobody can answer;
//   - and the effect isn't targeted, so hexproof is irrelevant.

const (
	gravePactOracle        = "6f4ac4a4-53ec-4bc9-8f5c-d4b801d867b2"
	dictateOfErebosOracle  = "7c777a41-e40a-4b40-96bf-8ddd5c12924c"
	fleshbagMarauderOracle = "4b1bf05e-753e-4350-a913-894cf3cecc0c"
	butcherOfMalakirOracle = "a85197ab-dc94-4b72-9716-8dbdbbe90ff8"
)

// sacrificeChoiceFor returns the queued sacrifice prompt addressed to a
// player, or nil.
func sacrificeChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceSacrifice && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// answerSacrifice answers a player's sacrifice prompt with one of their
// permanents.
func answerSacrifice(t *testing.T, g *game.Game, chooser, cardID uuid.UUID) {
	t.Helper()
	c := sacrificeChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no sacrifice prompt for %s", chooser)
	}
	if err := g.ResolveSacrificeChoice(c.ID, chooser, cardID); err != nil {
		t.Fatalf("ResolveSacrificeChoice: %v", err)
	}
}

// TestGravePactAsksEveryOpponentButNotYou is the core shape: one prompt
// each for the other players, none for the controller, and each prompt
// offering only that player's own creatures.
func TestGravePactAsksEveryOpponentButNotYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", gravePactOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	mine := pushCatalogPermanent(g, me.ID, "My Keeper", "Creature — Bear", "", false)

	type opp struct {
		p    *game.Player
		bear uuid.UUID
	}
	var opps []opp
	for _, s := range g.Seats[1:] {
		opps = append(opps, opp{p: s, bear: pushCatalogPermanent(g, s.ID, "Their Bear", "Creature — Bear", "", false)})
	}
	if len(opps) == 0 {
		t.Fatal("test needs at least one opponent")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)

	// The controller is never asked.
	if c := sacrificeChoiceFor(g, me.ID); c != nil {
		t.Error("Grave Pact asked its own controller to sacrifice")
	}
	// Every opponent is, and only with their own creature on offer.
	for _, o := range opps {
		c := sacrificeChoiceFor(g, o.p.ID)
		if c == nil {
			t.Fatalf("no prompt for opponent %s", o.p.ID)
		}
		if len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != o.bear {
			t.Errorf("opponent %s offered %v, want just their own %v",
				o.p.ID, c.SacrificeOptions, o.bear)
		}
	}

	for _, o := range opps {
		answerSacrifice(t, g, o.p.ID, o.bear)
	}
	passPriorityAroundTable(t, g)

	for _, o := range opps {
		if _, ok := battlefieldCard(g, o.bear); ok {
			t.Errorf("opponent %s's creature survived", o.p.ID)
		}
	}
	if _, ok := battlefieldCard(g, mine); !ok {
		t.Error("my own creature died to my own Grave Pact")
	}
}

// TestGravePactCannotSacrificeSomeoneElsesCreature — the prompt is
// answered by its own chooser, with their own card. Both halves are
// enforced, because a client that got either wrong would otherwise let
// one player dismantle another's board.
func TestGravePactCannotSacrificeSomeoneElsesCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", gravePactOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	mine := pushCatalogPermanent(g, me.ID, "My Keeper", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)

	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no prompt for the opponent")
	}
	// Their prompt, my creature.
	if err := g.ResolveSacrificeChoice(c.ID, opp.ID, mine); err == nil {
		t.Error("opponent sacrificed a creature they don't control")
	}
	// My hand on their prompt.
	if err := g.ResolveSacrificeChoice(c.ID, me.ID, theirs); err == nil {
		t.Error("a player answered someone else's sacrifice prompt")
	}
	if _, ok := battlefieldCard(g, mine); !ok {
		t.Error("my creature died to a rejected answer")
	}
	if _, ok := battlefieldCard(g, theirs); !ok {
		t.Error("their creature died to a rejected answer")
	}
}

// TestGravePactSkipsPlayersWithNoCreature — "sacrifices a creature" is
// "if you can". A player with an empty board must not be left holding
// a prompt with nothing in it, which would wedge the queue forever.
func TestGravePactSkipsPlayersWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, empty := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", gravePactOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	// `empty` gets a noncreature permanent, to prove the skip is about
	// the clause and not merely about an empty battlefield.
	rock := pushCatalogPermanent(g, empty.ID, "Sol Ring", "Artifact", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)

	if c := sacrificeChoiceFor(g, empty.ID); c != nil {
		t.Errorf("player with no creature was prompted (offered %v)", c.SacrificeOptions)
	}
	if _, ok := battlefieldCard(g, rock); !ok {
		t.Error("a noncreature permanent was eaten by a creature clause")
	}
}

// TestGravePactIgnoresHexproof — the effect does not target, so
// hexproof, shroud and protection are all irrelevant. This is the test
// that fails loudly if someone "simplifies" the prompt into
// PendingChoicePickTarget, which would filter them out.
func TestGravePactIgnoresHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", gravePactOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)

	hexproof := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: hexproof,
			Name:       "Slippery Bogbonder",
			TypeLine:   "Creature — Human Druid",
			Power:      2,
			Toughness:  2,
			Owner:      opp.ID,
			Controller: opp.ID,
			Keywords:   []string{"hexproof"},
		})
	})

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)

	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("hexproof creature's controller was not prompted")
	}
	if len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != hexproof {
		t.Fatalf("offered %v, want the hexproof creature %v", c.SacrificeOptions, hexproof)
	}
	answerSacrifice(t, g, opp.ID, hexproof)
	if _, ok := battlefieldCard(g, hexproof); ok {
		t.Error("hexproof creature survived a sacrifice it was not targeted by")
	}
}

// TestFleshbagMarauderAsksEveryoneIncludingYou — "each player", and the
// Marauder is on the battlefield when its own ETB trigger resolves, so
// it is a legal answer to its own effect. Both are how the card is
// actually played.
func TestFleshbagMarauderAsksEveryoneIncludingYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	marauder := castCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior",
		fleshbagMarauderOracle, nil)
	passPriorityAroundTable(t, g)

	mine := sacrificeChoiceFor(g, me.ID)
	if mine == nil {
		t.Fatal("controller was not prompted; Fleshbag says EACH player")
	}
	offered := false
	for _, id := range mine.SacrificeOptions {
		if id == marauder {
			offered = true
		}
	}
	if !offered {
		t.Errorf("the Marauder isn't a legal answer to its own trigger (offered %v)",
			mine.SacrificeOptions)
	}
	answerSacrifice(t, g, me.ID, marauder)
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, marauder); ok {
		t.Error("the Marauder survived being fed to its own trigger")
	}
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("opponent's creature survived")
	}
}

// TestButcherOfMalakirTriggersOnItsOwnDeath — "this creature OR another
// creature you control". Grave Pact's wording wouldn't cover it, so the
// distinction is worth a test.
func TestButcherOfMalakirTriggersOnItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butcher := pushCatalogPermanent(g, me.ID, "Butcher of Malakir", "Creature — Vampire Warrior",
		butcherOfMalakirOracle, false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(butcher) })
	passPriorityAroundTable(t, g)

	if sacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the Butcher's own death did not trigger it")
	}
	answerSacrifice(t, g, opp.ID, theirs)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("opponent's creature survived")
	}
}

// TestDictateOfErebosHasFlash — the only thing separating it from Grave
// Pact mechanically, and the reason both are in the catalog.
func TestDictateOfErebosHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Dictate of Erebos", "Enchantment",
		dictateOfErebosOracle, false)
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Dictate not on the battlefield")
	}
	found := false
	for _, kw := range card.Effective().Abilities {
		if kw == "flash" {
			found = true
		}
	}
	if !found {
		t.Errorf("Dictate of Erebos has no flash; abilities are %v", card.Effective().Abilities)
	}
}

// TestSacrificePromptIsDroppedWhenTheBoardEmpties — a prompt can be
// invalidated between being asked and being answered: Blood Artist
// drains in response and kills the only creature the chooser had left.
// The prompt must go away rather than block the queue on a card that
// no longer exists.
func TestSacrificePromptIsDroppedWhenTheBoardEmpties(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", gravePactOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)

	if sacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("no prompt for the opponent")
	}
	// Their creature dies to something else before they answer.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)

	if c := sacrificeChoiceFor(g, opp.ID); c != nil {
		t.Errorf("stale prompt survived with options %v; the queue is wedged", c.SacrificeOptions)
	}
}

// TestSacrificePromptViewCarriesTheOptions is the client's half: the
// picker needs card faces, and only the chooser's own permanents.
func TestSacrificePromptViewCarriesTheOptions(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", gravePactOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	pushCatalogPermanent(g, me.ID, "My Keeper", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)

	view := protocol.ViewOfGameFor(g, opp.ID.String())
	var found *protocol.PendingChoiceView
	for i := range view.PendingChoices {
		if view.PendingChoices[i].Kind == "sacrifice_choice" &&
			view.PendingChoices[i].Chooser == opp.ID.String() {
			found = &view.PendingChoices[i]
		}
	}
	if found == nil {
		t.Fatal("sacrifice_choice missing from the opponent's view")
	}
	if len(found.Options) != 1 {
		t.Fatalf("view offered %d options, want 1", len(found.Options))
	}
	if found.Options[0].InstanceID != theirs.String() {
		t.Errorf("view offered %s, want their own %s", found.Options[0].InstanceID, theirs)
	}
	if found.Reason == "" {
		t.Error("no reason string; the picker has no banner copy")
	}
}
