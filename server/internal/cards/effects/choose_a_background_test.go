package effects

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_a_background_test.go — #2874, ADR 0144: a commander with
// Choose a Background and a Background as the deck's two commanders,
// and the Background cards' grants to "commander creatures you own".
// The deck half is internal/deck's partner_test.go.

const (
	cabKarlachOracle       = "037355be-71e7-4866-80a6-80352c304970"
	cabAgentIronOracle     = "325032d5-c452-4454-8976-82f86fee5ab8"
	cabFlamingFistOracle   = "cdaefc96-4560-43bb-9514-bf87657bd481"
	cabSwordCoastOracle    = "71248bf6-a6a2-440e-9313-a7abce05a84b"
	cabShadowThievesOracle = "b2142d8b-ea53-443e-bd60-9e70d9da9bd7"
	cabCandlekeepOracle    = "f9516b60-36dd-4f8a-bebb-ac1bc0961f07"
	cabClanCrafterOracle   = "6ebd4b3f-246e-475f-8cec-838eaaf597e2"
	cabRaisedByGiantsOrcl  = "1682cf24-17a3-49ad-8b6f-9b7f13ebf53c"
	cabGuildArtisanOracle  = "aceac269-5100-428a-a4a3-bac57031e30b"
)

func TestBackgroundCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		cabKarlachOracle:       "Karlach, Fury of Avernus",
		cabFlamingFistOracle:   "Flaming Fist",
		cabSwordCoastOracle:    "Sword Coast Sailor",
		cabShadowThievesOracle: "Agent of the Shadow Thieves",
		cabCandlekeepOracle:    "Candlekeep Sage",
		cabClanCrafterOracle:   "Clan Crafter",
		cabRaisedByGiantsOrcl:  "Raised by Giants",
		cabGuildArtisanOracle:  "Guild Artisan",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v name=%q", name, ok, spec.Name)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v %q, want Full with no caveat", name, spec.Completeness, spec.Caveats)
		}
	}
}

// cabRow is a Scryfall record for the import road (deck.Resolve →
// deck.Validate → List.ToGameCards → Game.AddPlayer).
func cabRow(name, oracle, typeLine, manaCost, text string, identity ...string) cards.Card {
	c := cards.Card{
		ID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost, OracleText: text,
		ColorIdentity: identity, Legalities: map[string]string{"commander": "legal"},
	}
	if oracle != "" {
		c.OracleID = uuid.MustParse(oracle)
	}
	return c
}

// cabKarlachGame imports a Karlach deck with Agent of the Iron Throne as
// its second commander, through the importer, and seats it at a
// four-player table. The other three seats play filler.
func cabKarlachGame(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	karlach := cabRow("Karlach, Fury of Avernus", cabKarlachOracle, "Legendary Creature — Tiefling Barbarian", "{4}{R}",
		"Whenever you attack, if it's the first combat phase of the turn, untap all attacking creatures. They gain first strike until end of turn. After this phase, there is an additional combat phase.\nChoose a Background (You can have a Background as a second commander.)", "R")
	karlach.Power, karlach.Toughness = "5", "4"
	agent := cabRow("Agent of the Iron Throne", cabAgentIronOracle, "Legendary Enchantment — Background", "{2}{B}",
		"Commander creatures you own have \"Whenever an artifact or creature you control is put into a graveyard from the battlefield, each opponent loses 1 life.\"", "B")
	idx := cards.NewIndex()
	idx.Put(karlach)
	idx.Put(agent)
	entries := []deck.Entry{
		// Listed Background first, as Moxfield's alphabetical
		// commanders board does: Resolve puts Karlach first.
		{Name: agent.Name, Count: 1, IsCommander: true},
		{Name: karlach.Name, Count: 1, IsCommander: true},
	}
	for i := 0; i < 49; i++ {
		row := cabRow(fmt.Sprintf("Rakdos Spell %d", i), "", "Instant", "{B}{R}", "", "B", "R")
		idx.Put(row)
		entries = append(entries, deck.Entry{Name: row.Name, Count: 1})
	}
	swamp := cabRow("Swamp", "", "Basic Land — Swamp", "", "", "B")
	idx.Put(swamp)
	entries = append(entries, deck.Entry{Name: swamp.Name, Count: 49})

	list, err := deck.Resolve(idx, "Karlach", entries)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if err := deck.Validate(list); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	g := game.NewGame()
	me, err := g.AddPlayer("Karlach", list.ToGameCards())
	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	for i := 0; i < 3; i++ {
		filler := make([]game.Card, 20)
		for j := range filler {
			filler[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('B'+i)), filler); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	advanceToMain(t, g)
	return g, me
}

func cabCommandZoneCard(t *testing.T, p *game.Player, name string) uuid.UUID {
	t.Helper()
	for _, c := range p.Command.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("%s is not in the command zone", name)
	return uuid.Nil
}

// cabCastCommander casts a commander out of its owner's command zone
// and lets it resolve.
func cabCastCommander(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("CastSpell from the command zone: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("the commander did not resolve onto the battlefield")
	}
}

// The whole road: Karlach and a Background import as two commanders,
// both start in the command zone, each is cast from there with its own
// tax, and the Background's grant reaches Karlach only once the
// Background is on the battlefield — a Background in the command zone
// does nothing.
func TestKarlachAndABackgroundPlayAsTwoCommanders(t *testing.T) {
	g, me := cabKarlachGame(t)
	if names := []string{me.Command.Cards[0].Name, me.Command.Cards[1].Name}; len(me.Command.Cards) != 2 ||
		!slices.Contains(names, "Karlach, Fury of Avernus") || !slices.Contains(names, "Agent of the Iron Throne") {
		t.Fatalf("command zone %+v, want Karlach and the Background", me.Command.Cards)
	}
	for _, c := range me.Command.Cards {
		if !c.IsCommander {
			t.Errorf("%s is in the command zone but is not a commander", c.Name)
		}
	}
	karlach := cabCommandZoneCard(t, me, "Karlach, Fury of Avernus")
	agent := cabCommandZoneCard(t, me, "Agent of the Iron Throne")

	cabCastCommander(t, g, me, karlach)
	if n := grantedTriggerCount(g, karlach); n != 0 {
		t.Errorf("a Background in the command zone grants nothing: Karlach has %d granted triggers", n)
	}
	cabCastCommander(t, g, me, agent)
	if n := grantedTriggerCount(g, karlach); n != 1 {
		t.Fatalf("the Background on the battlefield grants its trigger to Karlach: %d", n)
	}
	if me.CommanderCasts[karlach] != 1 || me.CommanderCasts[agent] != 1 {
		t.Errorf("each commander keeps its own tax count (CR 702.124d): %v", me.CommanderCasts)
	}
	if c, ok := battlefieldCard(g, agent); !ok || !c.IsCommander {
		t.Error("the Background is a commander on the battlefield")
	}

	opp := g.Seats[1]
	before := opp.Life
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	b18Kill(t, g, bear)
	if opp.Life != before-1 {
		t.Errorf("Karlach's granted drain: opponent %d → %d, want one less", before, opp.Life)
	}
}

// Flaming Fist: the commander that attacks gains double strike; a
// creature that is not a commander does not.
func TestFlamingFistGivesAnAttackingCommanderDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Flaming Fist", "Legendary Enchantment — Background", cabFlamingFistOracle, 0, 0, "W")
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, commander, bear)
	passPriorityAroundTable(t, g)
	c := ringBFCard(t, g, commander)
	if !effectiveHasKeyword(c, "double strike") {
		t.Error("the attacking commander has double strike")
	}
	b := ringBFCard(t, g, bear)
	if effectiveHasKeyword(b, "double strike") {
		t.Error("a Bear is not a commander")
	}
}

// Sword Coast Sailor: attacking the (joint) healthiest opponent makes
// the commander unblockable this turn; attacking a player another
// opponent is ahead of does not.
func TestSwordCoastSailorMakesTheCommanderUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b21Push(g, me.ID, "Sword Coast Sailor", "Legendary Enchantment — Background", cabSwordCoastOracle, 0, 0, "U")
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other.Life = opp.Life + 5
	declareAttack(t, g, opp.ID, commander)
	if n := len(g.PendingTriggers) + len(g.StackMeta); n != 0 {
		t.Errorf("another opponent has more life: no trigger, got %d", n)
	}
	if reason := blockRefusal(t, g, wall, commander); reason != "" {
		t.Errorf("another opponent has more life: the commander can still be blocked, got %q", reason)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Sword Coast Sailor", "Legendary Enchantment — Background", cabSwordCoastOracle, 0, 0, "U")
	commander = b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	wall = b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, commander)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if reason := blockRefusal(t, g, wall, commander); reason == "" {
		t.Error("everyone on 40: the commander can't be blocked this turn")
	}
}

// Agent of the Shadow Thieves: a +1/+1 counter, deathtouch and
// indestructible on the commander that attacked.
func TestAgentOfTheShadowThievesPumpsTheAttackingCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Agent of the Shadow Thieves", "Legendary Enchantment — Background", cabShadowThievesOracle, 0, 0, "B")
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	declareAttack(t, g, opp.ID, commander)
	passPriorityAroundTable(t, g)
	c := ringBFCard(t, g, commander)
	if c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters %d, want 1", c.Counters[game.CounterPlusOne])
	}
	for _, kw := range []string{"deathtouch", "indestructible"} {
		if !effectiveHasKeyword(c, kw) {
			t.Errorf("the commander has %s until end of turn", kw)
		}
	}
}

// Candlekeep Sage: a commander creature you own draws as it enters
// and as it leaves, while the Background is on the battlefield.
func TestCandlekeepSageDrawsAsTheCommanderEntersAndLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Candlekeep Sage", "Legendary Enchantment — Background", cabCandlekeepOracle, 0, 0, "U")
	id := uuid.New()
	me.Command.PushTop(game.Card{InstanceID: id, Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, IsCommander: true, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	hand := handSize(me)
	cabCastCommander(t, g, me, id)
	if handSize(me) != hand+1 {
		t.Fatalf("the commander entered: hand %d → %d, want one card drawn", hand, handSize(me))
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	answerCommanderReturn(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+2 {
		t.Errorf("the commander left: hand %d, want %d", handSize(me), hand+2)
	}
}

// Clan Crafter: the commander creature has "{2}, Sacrifice an artifact:
// Put a +1/+1 counter on this creature and draw a card."
func TestClanCrafterGrantsTheCommanderASacrificeAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Clan Crafter", "Legendary Enchantment — Background", cabClanCrafterOracle, 0, 0, "U")
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trinket := b12Permanent(g, me.ID, "Trinket", "Artifact")
	advanceToMain(t, g)
	if i, _ := grantedActivatedIndex(t, g, bear); i >= 0 {
		t.Error("a Bear is not a commander")
	}
	idx, ref := grantedActivatedIndex(t, g, commander)
	if idx < 0 {
		t.Fatal("the commander has the granted ability")
	}
	hand := handSize(me)
	b16Activate(t, g, me.ID, commander, idx, game.ActivateAbilityParams{Ref: ref, SacrificeIDs: []uuid.UUID{trinket}})
	if g.Battlefield.Contains(trinket) {
		t.Error("the artifact is sacrificed")
	}
	c, _ := battlefieldCard(g, commander)
	if c.Counters[game.CounterPlusOne] != 1 || handSize(me) != hand+1 {
		t.Errorf("counter %d and hand %d → %d, want one of each", c.Counters[game.CounterPlusOne], hand, handSize(me))
	}
}

// Raised by Giants: base 10/10 and a Giant, with counters on top; not
// on a creature that is not a commander, and not on a commander you
// don't own.
func TestRaisedByGiantsMakesYourCommanderA10By10Giant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Raised by Giants", "Legendary Enchantment — Background", cabRaisedByGiantsOrcl, 0, 0, "G")
	commander := b21Commander(g, me.ID, "My Commander", "Legendary Creature — Human", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b21Commander(g, opp.ID, "Their Commander", "Legendary Creature — Human", 3, 3)
	if p, tough := effectivePower(t, g, commander), effectiveToughness(t, g, commander); p != 10 || tough != 10 {
		t.Errorf("your commander is %d/%d, want 10/10", p, tough)
	}
	if !slices.Contains(effectiveSubtypes(t, g, commander), "Giant") || !slices.Contains(effectiveSubtypes(t, g, commander), "Human") {
		t.Errorf("a Giant in addition to its other types: %v", effectiveSubtypes(t, g, commander))
	}
	var err error
	g.WithWriteLock(func() { err = g.AddCounterForEffect(commander, game.CounterPlusOne, 1) })
	if err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	if p := ringBFCard(t, g, commander).CurrentPower(); p != 11 {
		t.Errorf("a +1/+1 counter applies on top of base 10: %d", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a Bear is %d/x, not a commander", p)
	}
	if p := effectivePower(t, g, theirs); p != 3 {
		t.Errorf("an opponent's commander is %d/x, not one you own", p)
	}
}

// Guild Artisan is a real grant now: a commander an opponent has
// stolen still has it (you own it), and the Treasures are the thief's.
func TestGuildArtisanStolenCommanderMakesTreasuresForTheThief(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b21Push(g, me.ID, "Guild Artisan", "Legendary Enchantment — Background", cabGuildArtisanOracle, 0, 0, "R")
	// Yours, under the opponent's control since before their turn.
	commander := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human", IsCommander: true,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: opp.ID,
	})
	if grantedTriggerCount(g, commander) != 1 {
		t.Fatal("a stolen commander you own still has the Background's ability")
	}
	advanceToMainOf(t, g, 1)
	declareAttack(t, g, other.ID, commander)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, opp.ID, "Treasure"); n != 2 {
		t.Errorf("the thief's attack makes the thief two Treasures, got %d", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("the owner makes none, got %d", n)
	}
}
