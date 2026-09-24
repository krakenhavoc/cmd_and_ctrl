package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// theme_deck_aristocrats_test.go — S21's theme-deck smoke test: a
// Korvold-style aristocrats deck plays three of its own turns, end
// to end, through the real turn engine.
//
// It is a Go test rather than a gamecli action script on purpose.
// gamecli drives a LIVE server over the websocket, which means the
// run needs a seeded game, a deck upload and a Scryfall dump on
// disk, and the assertions end up being eyeballed from printed
// snapshots. The interesting parts of "a deck plays" — that the
// triggers fire in the right order, that the sacrifice prompts are
// answerable, that the counters land — are all game-state
// assertions, so they belong somewhere CI runs them on every push.
// What gamecli still gives that this does not is the wire format and
// the client; those have their own e2e coverage.
//
// The deck is real (a 40-card library of catalog cards, shuffled and
// drawn from), but the opening hand is stacked: each turn pulls the
// card the script wants out of the library by name. A smoke test
// that depended on a shuffle would fail for reasons that have
// nothing to do with the catalog.

// aristocratsThemeDeck is the library seat 0 plays out of: the
// sacrifice ecosystem S21 shipped, plus enough lands to cast it.
func aristocratsThemeDeck() []game.Card {
	type entry struct {
		name     string
		typeLine string
		oracle   string
		n        int
	}
	list := []entry{
		{"Swamp", "Basic Land — Swamp", "", 14},
		{"Mountain", "Basic Land — Mountain", "", 6},
		{"Karn's Bastion", "Land", karnsBastionOracle, 1},
		{"High Market", "Land", b03HighMarketOracle, 1},
		{"Korvold, Fae-Cursed King", "Legendary Creature — Dragon Noble", korvoldOracle, 1},
		{"Mazirek, Kraul Death Priest", "Legendary Creature — Insect Shaman", mazirekOracle, 1},
		{"Blood Artist", "Creature — Vampire", bloodArtistOracle, 1},
		{"Bloodflow Connoisseur", "Creature — Vampire", bloodflowOracle, 1},
		{"Pawn of Ulamog", "Creature — Vampire Shaman", pawnOfUlamogOracle, 1},
		{"Grave Titan", "Creature — Giant", graveTitanOracle, 1},
		{"Goblin Bombardment", "Enchantment", goblinBombardmentOracle, 1},
		{"Dragon Fodder", "Sorcery", dragonFodderOracle, 2},
		{"Hordeling Outburst", "Sorcery", hordelingOutburstOracle, 2},
		{"Steady Progress", "Instant", steadyProgressOracle, 2},
		{"Village Rites", "Instant", "365548fb-5acc-4a8a-b20b-26d28b7d029f", 5},
	}
	var deck []game.Card
	for _, e := range list {
		for i := 0; i < e.n; i++ {
			deck = append(deck, game.Card{
				InstanceID: uuid.New(),
				Name:       e.name,
				TypeLine:   e.typeLine,
				OracleID:   e.oracle,
			})
		}
	}
	return deck
}

// newAristocratsThemeGame seats the theme deck at seat 0 and three
// filler opponents around it.
func newAristocratsThemeGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	if _, err := g.AddPlayer("Korvold", aristocratsThemeDeck()); err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	for i := 1; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(7, 21))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// tutorToHand pulls the named card out of a player's library and
// puts it in hand — the "stacked opening hand" the file comment
// describes. Fatals when the deck has run out of that card, which is
// a decklist bug rather than an engine one.
func tutorToHand(t *testing.T, g *game.Game, p *game.Player, name string) uuid.UUID {
	t.Helper()
	// A copy the seeded shuffle already dealt or drew into hand will
	// do: the helper's job is "have this card in hand", and pinning a
	// library order to the RNG derivation broke once already (#744).
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	for _, c := range p.Library.Cards {
		if c.Name != name {
			continue
		}
		if _, err := game.MoveCard(p.Library, p.Hand, c.InstanceID); err != nil {
			t.Fatalf("MoveCard %s: %v", name, err)
		}
		return c.InstanceID
	}
	t.Fatalf("no %s left in the library", name)
	return uuid.Nil
}

// playFromHand advances to a main phase and casts (or plays, for a
// land) a card already in the active player's hand.
func playFromHand(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
}

// battlefieldIDNamed returns the first battlefield card a player
// controls with the given name.
func battlefieldIDNamed(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// TestS21ThemeDeckAristocratsPlaysThreeTurns is the sprint's
// theme-deck smoke test. Three turns of the deck's own play,
// through the real turn engine, with every S21 mechanism in the
// path: tokens, a sacrifice outlet, a mandatory sacrifice with a
// choice, two aristocrats payoffs and a proliferate.
func TestS21ThemeDeckAristocratsPlaysThreeTurns(t *testing.T) {
	g := newAristocratsThemeGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	lifeBefore, myLifeBefore := victim.Life, me.Life

	// --- Turn 1: a land and two Goblins -------------------------
	playFromHand(t, g, me, tutorToHand(t, g, me, "Karn's Bastion"))
	playFromHand(t, g, me, tutorToHand(t, g, me, "Dragon Fodder"))
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Goblin"); n != 2 {
		t.Fatalf("turn 1: %d Goblins, want 2", n)
	}

	// --- Turn 2: the payoff and the outlet ----------------------
	advanceToUpkeepOf(t, g, 0)
	playFromHand(t, g, me, tutorToHand(t, g, me, "Swamp"))
	playFromHand(t, g, me, tutorToHand(t, g, me, "Blood Artist"))
	passPriorityAroundTable(t, g)
	playFromHand(t, g, me, tutorToHand(t, g, me, "Bloodflow Connoisseur"))
	passPriorityAroundTable(t, g)

	vamp := battlefieldIDNamed(g, me.ID, "Bloodflow Connoisseur")
	goblin := battlefieldIDNamed(g, me.ID, "Goblin")
	if vamp == uuid.Nil || goblin == uuid.Nil {
		t.Fatalf("turn 2: board is not set up (vamp=%v goblin=%v)", vamp, goblin)
	}
	if err := g.ActivateCatalogAbility(me.ID, vamp, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{goblin},
	}); err != nil {
		t.Fatalf("turn 2: feed the Connoisseur: %v", err)
	}
	// The Goblin's death triggers Blood Artist, which targets.
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("turn 2: %v", err)
		}
	}
	pickPlayer(t, g, me.ID, victim.ID)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, vamp, game.CounterPlusOne); got != 1 {
		t.Fatalf("turn 2: Connoisseur has %d counters, want 1", got)
	}
	if victim.Life != lifeBefore-1 {
		t.Errorf("turn 2: victim life %d -> %d, want -1", lifeBefore, victim.Life)
	}

	// --- Turn 3: the commander, then proliferate ----------------
	advanceToUpkeepOf(t, g, 0)
	playFromHand(t, g, me, tutorToHand(t, g, me, "Swamp"))
	playFromHand(t, g, me, tutorToHand(t, g, me, "Korvold, Fae-Cursed King"))
	passPriorityAroundTable(t, g)

	korvold := battlefieldIDNamed(g, me.ID, "Korvold, Fae-Cursed King")
	if korvold == uuid.Nil {
		t.Fatalf("turn 3: Korvold never landed")
	}
	// Korvold's entry demands a sacrifice; feed him the last Goblin.
	lastGoblin := battlefieldIDNamed(g, me.ID, "Goblin")
	if lastGoblin == uuid.Nil {
		t.Fatalf("turn 3: no Goblin left to feed Korvold")
	}
	answerSacrifice(t, g, me.ID, lastGoblin)
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("turn 3: %v", err)
		}
	}
	pickPlayer(t, g, me.ID, victim.ID)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, korvold, game.CounterPlusOne); got != 1 {
		t.Fatalf("turn 3: Korvold has %d counters after his own sacrifice, want 1", got)
	}

	// Karn's Bastion proliferates both creatures at once.
	bastion := battlefieldIDNamed(g, me.ID, "Karn's Bastion")
	if err := g.ActivateCatalogAbility(me.ID, bastion, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("turn 3: activate the Bastion: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := counterCount(g, korvold, game.CounterPlusOne); got != 2 {
		t.Errorf("turn 3: Korvold has %d counters after proliferate, want 2", got)
	}
	if got := counterCount(g, vamp, game.CounterPlusOne); got != 2 {
		t.Errorf("turn 3: Connoisseur has %d counters after proliferate, want 2", got)
	}
	if victim.Life != lifeBefore-2 {
		t.Errorf("victim life %d -> %d, want two Blood Artist drains", lifeBefore, victim.Life)
	}
	if me.Life != myLifeBefore+2 {
		t.Errorf("my life %d -> %d, want two Blood Artist gains", myLifeBefore, me.Life)
	}
	if n := countBattlefieldNamed(g, me.ID, "Goblin"); n != 0 {
		t.Errorf("both Goblins should have been eaten, %d left", n)
	}
	// Turn.Round counts rounds, not seats, so seat 0's third turn
	// is round 3 — and the cursor is still on seat 0, which is what
	// "played three of its own turns" means.
	if g.Turn.Round != 3 || g.Turn.ActiveSeat != 0 {
		t.Errorf("ended on round %d seat %d, want round 3 seat 0", g.Turn.Round, g.Turn.ActiveSeat)
	}
}
