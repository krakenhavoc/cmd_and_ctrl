package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b05_test.go — Special Guests, S82 BNS batch 5 (#5929). Each test
// plays the real catalog entry through the engine.

const (
	bnsMephidrossOID  = "a6763205-47c7-4d7f-be8c-2f69e8321048"
	bnsMistbindOID    = "d1946f91-0653-4647-b3de-95d78b9597c4"
	bnsIbexOID        = "7e52e91b-70e6-4f26-a5d4-b8e33d1debbf"
	bnsPolyraptorOID  = "e47e3a40-f51f-41ed-8e7c-06200a2abc22"
	bnsRatColonyOID   = "ec77d23b-0165-450d-9aae-73b755163753"
	bnsRelentlessOID  = "104ea189-14cd-420f-afdc-57b0f827ab8e"
	bnsSubtletyOID    = "377179d5-ac83-4d34-b5b1-f3d8caa60f79"
	bnsThoughtKnotOID = "cef7396d-2168-4282-bf38-670ca9eebd91"
	bnsWanderwineOID  = "9ab36915-b600-4970-961a-403f5da0dba1"
	bnsBridgeOID      = "0f071f64-b69b-4fa0-999c-5028429e3cfb"
	bnsDeafeningOID   = "6284ef01-d17d-4584-aecb-1b841ae638e5"
)

func TestBnsB05EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsMephidrossOID:  "Mephidross Vampire",
		bnsMistbindOID:    "Mistbind Clique",
		bnsIbexOID:        "Pathbreaker Ibex",
		bnsPolyraptorOID:  "Polyraptor",
		bnsRatColonyOID:   "Rat Colony",
		bnsRelentlessOID:  "Relentless Rats",
		bnsSubtletyOID:    "Subtlety",
		bnsThoughtKnotOID: "Thought-Knot Seer",
		bnsWanderwineOID:  "Wanderwine Prophets",
		bnsBridgeOID:      "Bridge from Below",
		bnsDeafeningOID:   "Deafening Silence",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Deafening Silence -------------------------------------------------------

func TestBnsB05DeafeningSilenceRefusesASecondNoncreatureSpellButNotCreatures(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	seedEnchantment(g, bnsDeafeningOID, "Deafening Silence", "Enchantment", active.ID)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}

	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant", burstLightningOracle, target, nil, nil); err != nil {
		t.Fatalf("the first noncreature spell was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	id, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant", burstLightningOracle, target, nil, nil)
	assertCantCast(t, err, "noncreature")
	if !active.Hand.Contains(id) {
		t.Error("a refused cast did not leave the card in hand")
	}
	if _, err := castWithOptionalCosts(t, g, "Grizzly Bears", "Creature — Bear", "", nil, nil, nil); err != nil {
		t.Errorf("a creature spell was refused: %v", err)
	}
}

func TestBnsB05DeafeningSilenceBindsEveryPlayerFromAnOpponentsBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	seedEnchantment(g, bnsDeafeningOID, "Deafening Silence", "Enchantment", g.Seats[1].ID)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}
	if _, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant", burstLightningOracle, target, nil, nil); err != nil {
		t.Fatalf("first spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	_, err := castWithOptionalCosts(t, g, "Burst Lightning", "Instant", burstLightningOracle, target, nil, nil)
	assertCantCast(t, err, "noncreature")
	_ = active
}

// --- Polyraptor --------------------------------------------------------------

func TestBnsB05PolyraptorCopiesItselfEachTimeItIsDealtDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raptor := b27Push(g, me.ID, "Polyraptor", "Creature — Dinosaur", bnsPolyraptorOID, "{6}{G}{G}", 5, 5, "G")
	hitter := b16Creature(g, opp.ID, "Their Imp", "Creature — Imp", 1, 1, "B")
	cmdr2B08Damage(g, hitter, raptor, 1)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Polyraptor"); n != 2 {
		t.Fatalf("%d Polyraptors after one damage event, want 2", n)
	}
	tokens := trekB37Tokens(g, me.ID, "Polyraptor")
	if len(tokens) != 1 {
		t.Fatalf("%d Polyraptor tokens, want 1", len(tokens))
	}
	if tokens[0].Power != 5 || tokens[0].Toughness != 5 {
		t.Errorf("the token is %d/%d, want 5/5", tokens[0].Power, tokens[0].Toughness)
	}
	// The token has the trigger too.
	cmdr2B08Damage(g, hitter, tokens[0].InstanceID, 1)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Polyraptor"); n != 3 {
		t.Errorf("%d Polyraptors after the token was damaged, want 3", n)
	}
}

// --- Rat Colony and Relentless Rats -------------------------------------------

func TestBnsB05RatColonyCountsOtherRatsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colony := b27Push(g, me.ID, "Rat Colony", "Creature — Rat", bnsRatColonyOID, "{1}{B}", 2, 1, "B")
	if p, _ := effectivePTOf(t, g, colony); p != 2 {
		t.Fatalf("power %d alone, want 2", p)
	}
	b16Creature(g, me.ID, "Other Rat", "Creature — Rat", 1, 1, "B")
	b16Creature(g, me.ID, "Not a Rat", "Creature — Bear", 1, 1, "G")
	b16Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1, "B")
	p, tough := effectivePTOf(t, g, colony)
	if p != 3 || tough != 1 {
		t.Errorf("Rat Colony is %d/%d with one other Rat of mine, want 3/1", p, tough)
	}
	b27Push(g, me.ID, "Rat Colony", "Creature — Rat", bnsRatColonyOID, "{1}{B}", 2, 1, "B")
	if p, _ := effectivePTOf(t, g, colony); p != 4 {
		t.Errorf("power %d with a second Rat Colony, want 4", p)
	}
}

func TestBnsB05RelentlessRatsCountsOnlyCreaturesNamedRelentlessRats(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := b27Push(g, me.ID, "Relentless Rats", "Creature — Rat", bnsRelentlessOID, "{1}{B}{B}", 2, 2, "B")
	b16Creature(g, me.ID, "Other Rat", "Creature — Rat", 1, 1, "B")
	if p, tough := effectivePTOf(t, g, rats); p != 2 || tough != 2 {
		t.Fatalf("Relentless Rats is %d/%d next to a differently named Rat, want 2/2", p, tough)
	}
	b27Push(g, opp.ID, "Relentless Rats", "Creature — Rat", bnsRelentlessOID, "{1}{B}{B}", 2, 2, "B")
	b27Push(g, me.ID, "Relentless Rats", "Creature — Rat", bnsRelentlessOID, "{1}{B}{B}", 2, 2, "B")
	if p, tough := effectivePTOf(t, g, rats); p != 4 || tough != 4 {
		t.Errorf("Relentless Rats is %d/%d with two other Relentless Rats on the battlefield, want 4/4", p, tough)
	}
}

// --- Pathbreaker Ibex --------------------------------------------------------

func TestBnsB05PathbreakerIbexPumpsEveryCreatureYouControlByTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ibex := b27Push(g, me.ID, "Pathbreaker Ibex", "Creature — Goat", bnsIbexOID, "{4}{G}{G}", 3, 3, "G")
	big := b16Creature(g, me.ID, "Big Bear", "Creature — Bear", 6, 6, "G")
	small := b16Creature(g, me.ID, "Small Bear", "Creature — Bear", 1, 1, "G")
	theirs := b16Creature(g, opp.ID, "Their Giant", "Creature — Giant", 9, 9, "R")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ibex, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{ibex: 9, big: 12, small: 7, theirs: 9} {
		if p, _ := effectivePTOf(t, g, id); p != want {
			t.Errorf("%v has power %d, want %d", id, p, want)
		}
	}
	for _, id := range []uuid.UUID{ibex, big, small} {
		if c := effectiveCardForTest(t, g, id); !game.HasKeyword(&c, "trample") {
			t.Errorf("%v did not gain trample", id)
		}
	}
	if c := effectiveCardForTest(t, g, theirs); game.HasKeyword(&c, "trample") {
		t.Error("an opponent's creature gained trample")
	}
}

// --- Mephidross Vampire ------------------------------------------------------

func TestBnsB05MephidrossVampireMakesYourCreaturesVampiresThatGrow(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vamp := b27Push(g, me.ID, "Mephidross Vampire", "Creature — Vampire", bnsMephidrossOID, "{4}{B}{B}", 3, 4, "B")
	bear := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4, "R")
	if !effectiveCardForTest(t, g, bear).HasSubtype("Vampire") {
		t.Error("my Bear is not a Vampire")
	}
	if effectiveCardForTest(t, g, theirs).HasSubtype("Vampire") {
		t.Error("their Ogre is a Vampire")
	}
	if !effectiveCardForTest(t, g, bear).HasSubtype("Bear") {
		t.Error("my Bear lost its own creature type")
	}
	// My Bear deals damage to a creature: it gets a counter.
	cmdr2B08Damage(g, bear, theirs, 2)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Errorf("my Bear has %d counters after damaging a creature, want 1", n)
	}
	// The Vampire itself too.
	cmdr2B08Damage(g, vamp, theirs, 1)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, vamp, "+1/+1"); n != 1 {
		t.Errorf("the Vampire has %d counters, want 1", n)
	}
	// An opponent's creature gets nothing, and damage to a player is not damage to a creature.
	cmdr2B08Damage(g, theirs, bear, 1)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, theirs, "+1/+1"); n != 0 {
		t.Errorf("their Ogre has %d counters, want 0", n)
	}
	// When the Vampire leaves, the grant and the type go with it.
	b27Kill(g, vamp)
	passPriorityAroundTable(t, g)
	if effectiveCardForTest(t, g, bear).HasSubtype("Vampire") {
		t.Error("my Bear is still a Vampire without the Mephidross Vampire")
	}
	cmdr2B08Damage(g, bear, theirs, 1)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Errorf("my Bear has %d counters after the Vampire left, want still 1", n)
	}
}

// --- Thought-Knot Seer -------------------------------------------------------

func TestBnsB05ThoughtKnotSeerExilesANonlandCardAndGivesACardWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := revealHand(opp,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	seer, err := trekB37CastFromHand(t, g, game.Card{Name: "Thought-Knot Seer", TypeLine: "Creature — Eldrazi", OracleID: bnsThoughtKnotOID, Power: 4, Toughness: 4}, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	pick := openVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[0]}); err == nil {
		t.Fatal("the land was accepted as a nonland card")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[1]) {
		t.Error("the chosen card was not exiled")
	}
	if !opp.Hand.Contains(ids[0]) {
		t.Error("the land left the hand")
	}
	seedLibrary(opp, "Their Top")
	before := opp.Hand.Size()
	b27Kill(g, seer)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before+1 {
		t.Errorf("the opponent's hand went %d -> %d, want one card drawn", before, got)
	}
}

// --- Bridge from Below -------------------------------------------------------

func bnsBridgeInGraveyard(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Bridge from Below", TypeLine: "Enchantment", OracleID: bnsBridgeOID,
		ManaCost: "{B}{B}{B}", Owner: p.ID, Controller: p.ID})
	return id
}

func TestBnsB05BridgeFromBelowMakesAZombieForEachOfYourNontokenCreaturesThatDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bridge := bnsBridgeInGraveyard(me)
	a := b16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b := b16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2, "G")
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rat", TypeLine: "Token Creature — Rat",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	stolen := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	_ = stolen
	b27Kill(g, a)
	passPriorityAroundTable(t, g)
	if got := len(trekB37Tokens(g, me.ID, "Zombie")); got != 1 {
		t.Fatalf("%d Zombies after one of my creatures died, want 1", got)
	}
	z := trekB37Tokens(g, me.ID, "Zombie")[0]
	if z.Power != 2 || z.Toughness != 2 || len(z.Colors) != 1 || z.Colors[0] != "B" {
		t.Errorf("the token is %d/%d %v, want a 2/2 black Zombie", z.Power, z.Toughness, z.Colors)
	}
	b27Kill(g, token)
	passPriorityAroundTable(t, g)
	if got := len(trekB37Tokens(g, me.ID, "Zombie")); got != 1 {
		t.Errorf("a token's death made a Zombie: %d Zombies", got)
	}
	b27Kill(g, b)
	passPriorityAroundTable(t, g)
	if got := len(trekB37Tokens(g, me.ID, "Zombie")); got != 2 {
		t.Errorf("%d Zombies after a second death, want 2", got)
	}
	if !me.Graveyard.Contains(bridge) {
		t.Error("the Bridge left the graveyard though no opponent's creature died")
	}
}

func TestBnsB05BridgeFromBelowExilesItselfWhenACreatureGoesToAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bridge := bnsBridgeInGraveyard(me)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bridge) {
		t.Error("the Bridge was not exiled")
	}
	if got := len(trekB37Tokens(g, me.ID, "Zombie")); got != 0 {
		t.Errorf("a Zombie was made for an opponent's creature: %d", got)
	}
	// An opponent's creature that I controlled goes to its owner's graveyard: it is the
	// owner's graveyard that counts, so a stolen creature of theirs dying under my control
	// exiles the Bridge and makes no Zombie.
}

func TestBnsB05BridgeFromBelowDoesNothingOutsideTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bridge := handCardForTest(me, "Bridge from Below", "Enchantment", bnsBridgeOID)
	a := b16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b27Kill(g, a)
	passPriorityAroundTable(t, g)
	if got := len(trekB37Tokens(g, me.ID, "Zombie")); got != 0 {
		t.Errorf("a Bridge in hand made %d Zombies", got)
	}
	if !me.Hand.Contains(bridge) {
		t.Error("the Bridge left the hand")
	}
}

// --- Subtlety ----------------------------------------------------------------

func bnsOpponentCreatureSpell(t *testing.T, g *game.Game, opp *game.Player) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1,
		Keywords: []string{"flash"}, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast their creature spell: %v", err)
	}
	return id
}

func TestBnsB05SubtletyPutsACreatureSpellOnTheTopOrBottomOfItsOwnersLibrary(t *testing.T) {
	for _, bottom := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		seedLibrary(opp, "Their Top")
		elf := bnsOpponentCreatureSpell(t, g, opp)
		sub := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: sub, Name: "Subtlety", TypeLine: "Creature — Elemental Incarnation", OracleID: bnsSubtletyOID,
			Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, sub, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast Subtlety: %v", err)
		}
		passPriorityAroundTable(t, g)
		pick := latestPickTarget(g, me.ID)
		if pick == nil {
			t.Fatal("Subtlety asked for no target")
		}
		pickCard(t, g, me.ID, elf)
		passPriorityAroundTable(t, g)
		if putInLibraryChoiceFor(g, me.ID) != nil {
			t.Fatal("the caster was asked; the spell's owner chooses")
		}
		c := putInLibraryChoiceFor(g, opp.ID)
		if c == nil {
			t.Fatalf("the spell's owner was not asked: %+v", g.PendingChoices)
		}
		var top, bot []uuid.UUID
		if bottom {
			bot = []uuid.UUID{elf}
		} else {
			top = []uuid.UUID{elf}
		}
		if err := g.ResolvePutInLibrary(c.ID, opp.ID, bot, top); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Stack.Contains(elf) {
			t.Fatal("the spell is still on the stack")
		}
		cards := opp.Library.Cards
		if bottom && cards[0].InstanceID != elf {
			t.Errorf("the bottom card is %q, want the spell", cards[0].Name)
		}
		if !bottom && cards[len(cards)-1].InstanceID != elf {
			t.Errorf("the top card is %q, want the spell", cards[len(cards)-1].Name)
		}
		for _, ev := range g.Events {
			if ev.Kind == game.EventCounterSpell && ev.CardID == elf {
				t.Error("the spell was reported as countered")
			}
		}
		if !g.Battlefield.Contains(sub) {
			t.Error("a hard-cast Subtlety left the battlefield")
		}
	}
}

func TestBnsB05SubtletyCantTargetANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Needle", TypeLine: "Instant", OracleID: b5ShockOracle, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatal(err)
	}
	sub := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sub, Name: "Subtlety", TypeLine: "Creature — Elemental Incarnation", OracleID: bnsSubtletyOID,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, sub, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil && hasID(p.PickTargetCards, bolt) {
		t.Error("an instant spell was offered as Subtlety's target")
	}
}

func TestBnsB05SubtletyEvokedExilesABlueCardAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	blue := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: blue, Name: "Blue Card", TypeLine: "Sorcery", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	red := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: red, Name: "Red Card", TypeLine: "Sorcery", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	sub := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sub, Name: "Subtlety", TypeLine: "Creature — Elemental Incarnation", OracleID: bnsSubtletyOID,
		Power: 3, Toughness: 3, Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, sub, game.CastSpellParams{Strict: true, AlternativeCost: "evoke", AltCostIDs: []uuid.UUID{red}}); err == nil {
		t.Fatal("a red card paid for evoke")
	}
	if err := g.CastSpell(me.ID, sub, game.CastSpellParams{Strict: true, AlternativeCost: "evoke", AltCostIDs: []uuid.UUID{blue}}); err != nil {
		t.Fatalf("evoke: %v", err)
	}
	if !g.Exile.Contains(blue) {
		t.Error("the blue card was not exiled")
	}
	passPriorityAroundTable(t, g)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sub) {
		t.Error("the evoked Subtlety is still on the battlefield")
	}
	if !me.Graveyard.Contains(sub) {
		t.Error("the evoked Subtlety did not reach the graveyard")
	}
}

// --- Champion: Mistbind Clique and Wanderwine Prophets ------------------------

func bnsCastChampion(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int) uuid.UUID {
	t.Helper()
	id, err := trekB37CastFromHand(t, g, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: p, Toughness: tough}, 0)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestBnsB05MistbindCliqueChampionsAFaerieTapsLandsAndReturnsItWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sprite := b16Creature(g, me.ID, "Sprite", "Creature — Faerie", 1, 1, "U")
	human := b16Creature(g, me.ID, "Human", "Creature — Human", 1, 1, "W")
	theirFaerie := b16Creature(g, opp.ID, "Their Sprite", "Creature — Faerie", 1, 1, "U")
	land1 := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	land2 := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	clique := bnsCastChampion(t, g, "Mistbind Clique", "Creature — Faerie Wizard", bnsMistbindOID, 4, 4)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatalf("no champion prompt: %+v", g.PendingChoices)
	}
	if hasID(pick.ChooseCards, clique) || hasID(pick.ChooseCards, human) || hasID(pick.ChooseCards, theirFaerie) {
		t.Errorf("the champion prompt offered %v: only my other Faeries are legal", pick.ChooseCards)
	}
	if !hasID(pick.ChooseCards, sprite) {
		t.Fatalf("my Faerie is not offered: %v", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, sprite)
	if _, ok := battlefieldCard(g, sprite); ok {
		t.Error("the championed Faerie is still on the battlefield")
	}
	if !g.Exile.Contains(sprite) {
		t.Error("the championed Faerie was not exiled")
	}
	if _, ok := battlefieldCard(g, clique); !ok {
		t.Fatal("Mistbind Clique did not stay")
	}
	// "When a Faerie is championed with this creature, tap all lands target player controls."
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	for _, l := range []uuid.UUID{land1, land2} {
		if tapped, _ := battlefieldCardTapped(g, l); !tapped {
			t.Errorf("land %v was not tapped", l)
		}
	}
	// When Mistbind Clique leaves, the Faerie returns.
	b27Kill(g, clique)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if n := b27CountNamed(g, me.ID, "Sprite"); n != 1 {
		t.Errorf("%d Sprites of mine on the battlefield after Mistbind Clique left, want the championed one back", n)
	}
	if g.Exile.Contains(sprite) {
		t.Error("the Faerie is still exiled")
	}
}

func TestBnsB05MistbindCliqueWithNoOtherFaerieIsSacrificedAndTapsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b16Creature(g, me.ID, "Human", "Creature — Human", 1, 1, "W")
	land := pushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	clique := bnsCastChampion(t, g, "Mistbind Clique", "Creature — Faerie Wizard", bnsMistbindOID, 4, 4)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, clique); ok {
		t.Error("Mistbind Clique stayed with nothing to champion")
	}
	if !me.Graveyard.Contains(clique) {
		t.Error("Mistbind Clique is not in the graveyard")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the tap trigger asked for a target though nothing was championed")
	}
	if tapped, _ := battlefieldCardTapped(g, land); tapped {
		t.Error("a land was tapped")
	}
}

func TestBnsB05MistbindCliqueDecliningTheChampionSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sprite := b16Creature(g, me.ID, "Sprite", "Creature — Faerie", 1, 1, "U")
	clique := bnsCastChampion(t, g, "Mistbind Clique", "Creature — Faerie Wizard", bnsMistbindOID, 4, 4)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, clique); ok {
		t.Error("Mistbind Clique stayed though no Faerie was exiled")
	}
	if _, ok := battlefieldCard(g, sprite); !ok {
		t.Error("the Faerie was exiled anyway")
	}
}

func TestBnsB05WanderwineProphetsChampionsAMerfolkAndReturnsItWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fish := b16Creature(g, me.ID, "Fish", "Creature — Merfolk", 1, 1, "U")
	faerie := b16Creature(g, me.ID, "Sprite", "Creature — Faerie", 1, 1, "U")
	prophets := bnsCastChampion(t, g, "Wanderwine Prophets", "Creature — Merfolk Wizard", bnsWanderwineOID, 4, 4)
	pick := latestChooseCards(g, me.ID)
	if pick == nil || hasID(pick.ChooseCards, faerie) || !hasID(pick.ChooseCards, fish) {
		t.Fatalf("the champion prompt is %+v: only the other Merfolk is legal", pick)
	}
	answerChooseCards(t, g, me.ID, fish)
	if !g.Exile.Contains(fish) {
		t.Error("the Merfolk was not exiled")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("Wanderwine Prophets has no tap trigger")
	}
	b27Kill(g, prophets)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if n := b27CountNamed(g, me.ID, "Fish"); n != 1 {
		t.Errorf("%d Fish of mine on the battlefield after Wanderwine Prophets left, want the championed one back", n)
	}
}

func TestBnsB05WanderwineProphetsSacrificesAMerfolkForAnExtraTurn(t *testing.T) {
	for _, sacrifice := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		prophets := b27Push(g, me.ID, "Wanderwine Prophets", "Creature — Merfolk Wizard", bnsWanderwineOID, "{4}{U}{U}", 4, 4, "U")
		fish := b16Creature(g, me.ID, "Fish", "Creature — Merfolk", 1, 1, "U")
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(prophets, opp.ID); err != nil {
			t.Fatal(err)
		}
		advanceTo(t, g, game.StepCombatDamage)
		passPriorityAroundTable(t, g)
		turns := len(g.ExtraTurns)
		if sacrifice {
			answerChooseCardsOrPick(t, g, me.ID, fish)
		} else {
			pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
			if pick == nil {
				t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
			}
			if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
				t.Fatalf("declining the sacrifice: %v", err)
			}
		}
		passPriorityAroundTable(t, g)
		if got := len(g.ExtraTurns) - turns; sacrifice && got != 1 || !sacrifice && got != 0 {
			t.Errorf("sacrifice=%v: %d extra turns queued", sacrifice, got)
		}
		_, fishStays := battlefieldCard(g, fish)
		if sacrifice == fishStays {
			t.Errorf("sacrifice=%v but the Merfolk on the battlefield is %v", sacrifice, fishStays)
		}
	}
}
