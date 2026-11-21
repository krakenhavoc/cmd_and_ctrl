package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestGravePactAsksEveryOpponentButNotYou is the core shape: one prompt
// each for the other players, none for the controller, and each prompt
// offering only that player's own creatures.
func TestGravePactAsksEveryOpponentButNotYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", effectstest.GravePactOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Keeper", "Creature — Bear", "", false)

	type opp struct {
		p           *game.Player
		bear, spare uuid.UUID
	}
	var opps []opp
	for _, s := range g.Seats[1:] {
		opps = append(opps, opp{p: s,
			bear:  effectstest.PushCatalogPermanent(g, s.ID, "Their Bear", "Creature — Bear", "", false),
			spare: effectstest.PushCatalogPermanent(g, s.ID, "Their Spare", "Creature — Bear", "", false)})
	}
	if len(opps) == 0 {
		t.Fatal("test needs at least one opponent")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)

	// The controller is never asked.
	if c := effectstest.SacrificeChoiceFor(g, me.ID); c != nil {
		t.Error("Grave Pact asked its own controller to sacrifice")
	}
	// Every opponent is, one at a time in turn order (CR 101.4), and
	// only with their own creatures on offer.
	for _, o := range opps {
		c := effectstest.SacrificeChoiceFor(g, o.p.ID)
		if c == nil {
			t.Fatalf("no prompt for opponent %s", o.p.ID)
		}
		opts := effectstest.SacrificeOptions(c)
		if len(opts) != 2 || !effectstest.EdictOffers(opts, o.bear) || !effectstest.EdictOffers(opts, o.spare) {
			t.Errorf("opponent %s offered %v, want just their own two", o.p.ID, opts)
		}
		effectstest.AnswerSacrifice(t, g, o.p.ID, o.bear)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, o := range opps {
		if _, ok := effectstest.BattlefieldCard(g, o.bear); ok {
			t.Errorf("opponent %s's creature survived", o.p.ID)
		}
		if _, ok := effectstest.BattlefieldCard(g, o.spare); !ok {
			t.Errorf("opponent %s's unchosen creature died", o.p.ID)
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, mine); !ok {
		t.Error("my own creature died to my own Grave Pact")
	}
}

// TestGravePactCannotSacrificeSomeoneElsesCreature — the prompt is
// answered by its own chooser, with their own card. Both halves are
// enforced, because a client that got either wrong would otherwise let
// one player dismantle another's board.
func TestGravePactCannotSacrificeSomeoneElsesCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", effectstest.GravePactOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Keeper", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Spare", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no prompt for the opponent")
	}
	// Their prompt, my creature.
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{mine}); err == nil {
		t.Error("opponent sacrificed a creature they don't control")
	}
	// My hand on their prompt.
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{theirs}); err == nil {
		t.Error("a player answered someone else's sacrifice prompt")
	}
	if _, ok := effectstest.BattlefieldCard(g, mine); !ok {
		t.Error("my creature died to a rejected answer")
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); !ok {
		t.Error("their creature died to a rejected answer")
	}
}

// TestGravePactSkipsPlayersWithNoCreature — "sacrifices a creature" is
// "if you can". A player with an empty board must not be left holding
// a prompt with nothing in it, which would wedge the queue forever.
func TestGravePactSkipsPlayersWithNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, empty := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", effectstest.GravePactOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	// `empty` gets a noncreature permanent, to prove the skip is about
	// the clause and not merely about an empty battlefield.
	rock := effectstest.PushCatalogPermanent(g, empty.ID, "Sol Ring", "Artifact", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.SacrificeChoiceFor(g, empty.ID); c != nil {
		t.Errorf("player with no creature was prompted (offered %v)", effectstest.SacrificeOptions(c))
	}
	if _, ok := effectstest.BattlefieldCard(g, rock); !ok {
		t.Error("a noncreature permanent was eaten by a creature clause")
	}
}

// TestGravePactIgnoresHexproof — the effect does not target, so
// hexproof, shroud and protection are all irrelevant. This is the test
// that fails loudly if someone "simplifies" the prompt into
// PendingChoicePickTarget, which would filter them out.
func TestGravePactIgnoresHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", effectstest.GravePactOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)

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
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Spare", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("hexproof creature's controller was not prompted")
	}
	if !effectstest.EdictOffers(effectstest.SacrificeOptions(c), hexproof) {
		t.Fatalf("offered %v, want the hexproof creature %v among them", effectstest.SacrificeOptions(c), hexproof)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, hexproof)
	if _, ok := effectstest.BattlefieldCard(g, hexproof); ok {
		t.Error("hexproof creature survived a sacrifice it was not targeted by")
	}
}

// TestFleshbagMarauderAsksEveryoneIncludingYou — "each player", and the
// Marauder is on the battlefield when its own ETB trigger resolves, so
// it is a legal answer to its own effect. Both are how the card is
// actually played.
func TestFleshbagMarauderAsksEveryoneIncludingYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Spare", "Creature — Bear", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "My Keeper", "Creature — Bear", "", false)

	marauder := effectstest.CastCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior",
		effectstest.FleshbagMarauderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	mine := effectstest.SacrificeChoiceFor(g, me.ID)
	if mine == nil {
		t.Fatal("controller was not prompted; Fleshbag says EACH player")
	}
	if !effectstest.EdictOffers(effectstest.SacrificeOptions(mine), marauder) {
		t.Errorf("the Marauder isn't a legal answer to its own trigger (offered %v)",
			effectstest.SacrificeOptions(mine))
	}
	effectstest.AnswerSacrifice(t, g, me.ID, marauder)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, marauder); ok {
		t.Error("the Marauder survived being fed to its own trigger")
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("opponent's creature survived")
	}
}

// TestButcherOfMalakirTriggersOnItsOwnDeath — "this creature OR another
// creature you control". Grave Pact's wording wouldn't cover it, so the
// distinction is worth a test.
func TestButcherOfMalakirTriggersOnItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butcher := effectstest.PushCatalogPermanent(g, me.ID, "Butcher of Malakir", "Creature — Vampire Warrior",
		effectstest.ButcherOfMalakirOracle, false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Spare", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(butcher) })
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the Butcher's own death did not trigger it")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("opponent's creature survived")
	}
}

// TestDictateOfErebosHasFlash — the only thing separating it from Grave
// Pact mechanically, and the reason both are in the catalog.
func TestDictateOfErebosHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Dictate of Erebos", "Enchantment",
		effectstest.DictateOfErebosOracle, false)
	card, ok := effectstest.BattlefieldCard(g, id)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", effectstest.GravePactOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	spare := effectstest.PushCatalogPermanent(g, opp.ID, "Their Spare", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("no prompt for the opponent")
	}
	// Their creatures die to something else before they answer.
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(theirs)
		_ = g.DestroyPermanentForEffect(spare)
	})
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		t.Errorf("stale prompt survived with options %v; the queue is wedged", effectstest.SacrificeOptions(c))
	}
}

// TestSacrificePromptViewCarriesTheOptions is the client's half: the
// picker needs card faces, and only the chooser's own permanents.
func TestSacrificePromptViewCarriesTheOptions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Grave Pact", "Enchantment", effectstest.GravePactOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "My Keeper", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	spare := effectstest.PushCatalogPermanent(g, opp.ID, "Their Spare", "Creature — Bear", "", false)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)

	view := protocol.ViewOfGameFor(g, opp.ID.String())
	var found *protocol.PendingChoiceView
	for i := range view.PendingChoices {
		if view.PendingChoices[i].Kind == "choose_cards" &&
			view.PendingChoices[i].Chooser == opp.ID.String() {
			found = &view.PendingChoices[i]
		}
	}
	if found == nil {
		t.Fatal("the sacrifice pick is missing from the opponent's view")
	}
	if found.ChooseDestination != "sacrifice" {
		t.Errorf("choose_destination = %q, want sacrifice (#3664)", found.ChooseDestination)
	}
	got := map[string]bool{}
	for _, o := range found.Options {
		got[o.InstanceID] = true
	}
	if len(found.Options) != 2 || !got[theirs.String()] || !got[spare.String()] {
		t.Errorf("view offered %v, want their own two creatures", found.Options)
	}
	if found.Reason == "" {
		t.Error("no reason string; the picker has no banner copy")
	}
}
