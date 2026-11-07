package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_sz_cards_test.go — Tarkir: Dragonstorm, the first batch of cards
// from S to Z (#4463). Each test plays the real catalog entry and asserts
// what its text says.

const (
	tdmSultaiDevoteeOracle  = "ee7b3fc0-6128-4bf6-b3e7-7371e6f93c4b"
	tdmSultaiMonumentOracle = "9db71b76-eda7-49dd-af5a-3de18acb8ccb"
	tdmTeemingOracle        = "2154cc68-41a4-4e88-9fe0-e0150dc0be45"
	tdmTemurDevoteeOracle   = "e444f072-ce24-4fd1-99eb-d20d12bc119a"
	tdmTemurMonumentOracle  = "5464d5ae-5be5-4e52-b9d0-ddce48fd93b3"
	tdmLeopardOracle        = "511fa8a3-0a65-47fc-9f3b-57eaef2698cd"
	tdmStormsingerOracle    = "4aff2f31-619f-406e-b2c9-a2c733c04bc4"
	tdmWatcherOracle        = "b30eafed-2755-4423-9592-c8b83c68804b"
	tdmWingbladeOracle      = "c566317e-cd82-46c1-b506-f41256d815f6"
	tdmStrikemasterOracle   = "d81ade13-e7c7-4fe3-bb0c-5007ed9f1198"
	tdmEarthcarverOracle    = "75da657e-0e36-4f2a-9b81-b7eba7d8e050"
	tdmZurgosVanguardOracle = "d6fdfd94-cef5-482c-83ed-11c5ba40abe9"
)

func TestTdmSZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmSultaiDevoteeOracle:  "Sultai Devotee",
		tdmSultaiMonumentOracle: "Sultai Monument",
		tdmTeemingOracle:        "Teeming Dragonstorm",
		tdmTemurDevoteeOracle:   "Temur Devotee",
		tdmTemurMonumentOracle:  "Temur Monument",
		tdmLeopardOracle:        "Undergrowth Leopard",
		tdmStormsingerOracle:    "Venerated Stormsinger",
		tdmWatcherOracle:        "Watcher of the Wayside",
		tdmWingbladeOracle:      "Wingblade Disciple",
		tdmStrikemasterOracle:   "Sunset Strikemaster",
		tdmEarthcarverOracle:    "Unburied Earthcarver",
		tdmZurgosVanguardOracle: "Zurgo's Vanguard",
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

// --- fixtures ------------------------------------------------------------------

// tdmSZPlay casts a catalog creature with a body, so it survives to see
// its own enters trigger, and resolves what it puts on the stack.
func tdmSZPlay(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	return msc17Play(t, g, p, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: power, Toughness: toughness})
}

// tdmSZLives is every seat's life total, by seat index.
func tdmSZLives(g *game.Game) []int {
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	return lives
}

// tdmSZDrained asserts seat `me` gained `gain` and every other seat lost
// `loss` since `before`.
func tdmSZDrained(t *testing.T, g *game.Game, before []int, me, loss, gain int) {
	t.Helper()
	for i, p := range g.Seats {
		want := before[i] - loss
		if i == me {
			want = before[i] + gain
		}
		if p.Life != want {
			t.Errorf("seat %d has %d life, want %d", i, p.Life, want)
		}
	}
}

// --- Sultai Devotee and Temur Devotee ------------------------------------------

func TestTdmSZDevoteesFilterOnceEachTurn(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		power, toughness                int
		colours                         []string
		refused                         string
	}{
		{"Sultai Devotee", "Creature — Zombie Snake Druid", tdmSultaiDevoteeOracle, "deathtouch", 2, 1, []string{"B", "G", "U"}, "W"},
		{"Temur Devotee", "Creature — Human Druid", tdmTemurDevoteeOracle, "defender", 3, 3, []string{"G", "U", "R"}, "B"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, colour := range tc.colours {
				g := newCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				devotee := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
				if !hasEffectiveKeyword(t, g, devotee, tc.keyword) {
					t.Errorf("%s has no %s", tc.name, tc.keyword)
				}
				advanceToMain(t, g)
				g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
				fillPoolColored(me, tc.refused, 3)
				if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{tc.refused}}); err == nil {
					t.Errorf("the Devotee made {%s}", tc.refused)
				}
				if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
					t.Fatalf("{1}: add {%s}: %v", colour, err)
				}
				if got := mcPoolCount(me, colour); got != 1 {
					t.Errorf("{%s} in the pool %d, want 1", colour, got)
				}
				if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{colour}}); err == nil {
					t.Error("a second activation this turn was allowed")
				}
			}
		})
	}
}

// --- Sultai Monument and Temur Monument ----------------------------------------

func TestTdmSZMonumentsFindABasicOfTheirColours(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, find, refuse string
	}{
		{"Sultai Monument", tdmSultaiMonumentOracle, "Island", "Plains"},
		{"Temur Monument", tdmTemurMonumentOracle, "Mountain", "Swamp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			seedSearchLibrary(me,
				searchTestLand(tc.refuse, "Basic Land — "+tc.refuse),
				searchTestLand("Snowy "+tc.find, "Land — "+tc.find),
				searchTestLand(tc.find, "Basic Land — "+tc.find),
			)
			castCatalogSpell(t, g, tc.name, "Artifact", tc.oracle, nil)
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("the Monument's entry asked for no search")
			}
			if searchOptionNamed(g, c, tc.refuse) != uuid.Nil || searchOptionNamed(g, c, "Snowy "+tc.find) != uuid.Nil {
				t.Error("a basic of another type, or a nonbasic, was offered")
			}
			answerSearchNamed(t, g, me.ID, tc.find)
			found := false
			for _, card := range me.Hand.Cards {
				if card.Name == tc.find {
					found = true
				}
			}
			if !found {
				t.Errorf("the %s is not in hand", tc.find)
			}
		})
	}
}

func TestTdmSZMonumentsMakeTheirTokens(t *testing.T) {
	for _, tc := range []struct {
		name, oracle    string
		colours         map[string]int
		token, colour   string
		n, power, tough int
	}{
		{"Sultai Monument", tdmSultaiMonumentOracle, map[string]int{"B": 2, "G": 2, "U": 2}, "Zombie Druid", "B", 2, 2, 2},
		{"Temur Monument", tdmTemurMonumentOracle, map[string]int{"G": 3, "U": 2, "R": 2}, "Elephant", "G", 1, 5, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			monument := pushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)

			// Not in combat.
			advanceTo(t, g, game.StepBeginCombat)
			for c, n := range tc.colours {
				fillPoolColored(me, c, n)
			}
			if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err == nil {
				t.Fatal("the Monument's ability was activated in combat")
			}

			advanceTo(t, g, game.StepPostcombatMain)
			g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
			for c, n := range tc.colours {
				fillPoolColored(me, c, n)
			}
			if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(monument) {
				t.Fatal("the Monument was not sacrificed as the cost")
			}
			passPriorityAroundTable(t, g)
			tokens := socB8Named(g, me.ID, tc.token)
			if len(tokens) != tc.n {
				t.Fatalf("%d %s tokens, want %d", len(tokens), tc.token, tc.n)
			}
			for _, tok := range tokens {
				if tok.Power != tc.power || tok.Toughness != tc.tough || len(tok.Colors) != 1 || tok.Colors[0] != tc.colour {
					t.Errorf("%s %d/%d %v, want %d/%d %s", tc.token, tok.Power, tok.Toughness, tok.Colors, tc.power, tc.tough, tc.colour)
				}
			}
		})
	}
}

// --- Teeming Dragonstorm -------------------------------------------------------

func TestTdmSZTeemingDragonstormMakesTwoSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Teeming Dragonstorm", "Enchantment", tdmTeemingOracle, nil)
	passPriorityAroundTable(t, g)
	soldiers := socB8Named(g, me.ID, "Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("%d Soldiers, want 2", len(soldiers))
	}
	for _, s := range soldiers {
		if s.Power != 2 || s.Toughness != 2 || len(s.Colors) != 1 || s.Colors[0] != "W" || !IsToken(s) {
			t.Errorf("Soldier %d/%d %v, want a 2/2 white Soldier token", s.Power, s.Toughness, s.Colors)
		}
	}
}

func TestTdmSZTeemingDragonstormReturnsWhenADragonEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	storm := pushCatalogPermanent(g, me.ID, "Teeming Dragonstorm", "Enchantment", tdmTeemingOracle, false)
	msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if !g.Battlefield.Contains(storm) {
		t.Fatal("a Bear is not a Dragon")
	}
	msc17Play(t, g, me, game.Card{Name: "Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
	if g.Battlefield.Contains(storm) || !me.Hand.Contains(storm) {
		t.Error("a Dragon you control entering did not return the enchantment to your hand")
	}
}

// --- Undergrowth Leopard -------------------------------------------------------

func TestTdmSZUndergrowthLeopardSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	leopard := b12Push(g, me.ID, "Undergrowth Leopard", "Creature — Cat", tdmLeopardOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, leopard, "vigilance") {
		t.Error("Undergrowth Leopard has no vigilance")
	}
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	advanceToMain(t, g)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, leopard, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, leopard, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(leopard) {
		t.Fatal("the Leopard was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !opp.Graveyard.Contains(aura) {
		t.Error("the enchantment was not destroyed")
	}
}

// --- Venerated Stormsinger -----------------------------------------------------

func TestTdmSZVeneratedStormsingerDrainsWhenItOrAnotherOfYoursDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	singer := b12Push(g, me.ID, "Venerated Stormsinger", "Creature — Orc Cleric", tdmStormsingerOracle, 3, 3)
	advanceToMain(t, g)

	before := tdmSZLives(g)
	b18Kill(t, g, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	tdmSZDrained(t, g, before, 0, 1, 1)

	// An opponent's creature is not yours.
	before = tdmSZLives(g)
	b18Kill(t, g, b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	tdmSZDrained(t, g, before, 0, 0, 0)

	// The Stormsinger itself.
	before = tdmSZLives(g)
	b18Kill(t, g, singer)
	tdmSZDrained(t, g, before, 0, 1, 1)
}

func TestTdmSZVeneratedStormsingerMobilizesAndDrainsForTheWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	singer := b12Push(g, me.ID, "Venerated Stormsinger", "Creature — Orc Cleric", tdmStormsingerOracle, 3, 3)
	declareAttack(t, g, opp.ID, singer)
	passPriorityAroundTable(t, g)
	tdmPlaceWarriors(t, g, me.ID)
	if n := len(socB8Named(g, me.ID, "Warrior")); n != 1 {
		t.Fatalf("mobilize 1 made %d Warriors", n)
	}
	advanceToEndStepOf(t, g, 0)
	before := tdmSZLives(g)
	passPriorityAroundTable(t, g)
	if n := len(socB8Named(g, me.ID, "Warrior")); n != 0 {
		t.Fatalf("%d Warriors survived the end step", n)
	}
	tdmSZDrained(t, g, before, 0, 1, 1)
}

// --- Watcher of the Wayside ----------------------------------------------------

func TestTdmSZWatcherOfTheWaysideMillsTargetPlayerAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(opp, 5)
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	life := me.Life
	tdmSZPlay(t, g, me, "Watcher of the Wayside", "Artifact Creature — Golem", tdmWatcherOracle, 3, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil || !hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the enters trigger asks for any player, you included: %+v", p)
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - graveyard; got != 2 {
		t.Errorf("target player milled %d, want 2", got)
	}
	if got := opp.Library.Size(); got != library-2 {
		t.Errorf("target player's library %d, want %d", got, library-2)
	}
	if me.Life != life+2 {
		t.Errorf("you have %d life, want %d", me.Life, life+2)
	}
}

// --- Wingblade Disciple --------------------------------------------------------

func TestTdmSZWingbladeDiscipleMakesABirdOnTheSecondSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	disciple := b12Push(g, me.ID, "Wingblade Disciple", "Creature — Human Monk", tdmWingbladeOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, disciple, "flying") {
		t.Error("Wingblade Disciple has no flying")
	}
	for i, want := range []int{0, 1, 1} {
		b24CastNoncreature(t, g, "Spell", "Instant", "{U}")
		passPriorityAroundTable(t, g)
		birds := socB8Named(g, me.ID, "Bird")
		if len(birds) != want {
			t.Fatalf("after spell %d, %d Birds, want %d", i+1, len(birds), want)
		}
		for _, b := range birds {
			if b.Power != 1 || b.Toughness != 1 || len(b.Colors) != 1 || b.Colors[0] != "W" {
				t.Errorf("Bird %d/%d %v, want a 1/1 white Bird", b.Power, b.Toughness, b.Colors)
			}
			if !hasEffectiveKeyword(t, g, b.InstanceID, "flying") {
				t.Error("the Bird has no flying")
			}
		}
	}
}

// --- Sunset Strikemaster -------------------------------------------------------

func TestTdmSZSunsetStrikemasterTapsForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := b12Push(g, me.ID, "Sunset Strikemaster", "Creature — Human Monk", tdmStrikemasterOracle, 3, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, monk, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: add {R}: %v", err)
	}
	if got := mcPoolCount(me, "R"); got != 1 {
		t.Errorf("red in the pool %d, want 1", got)
	}
	if !isTapped(g, monk) {
		t.Error("the Strikemaster did not tap for mana")
	}
}

func TestTdmSZSunsetStrikemasterShootsDownAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	monk := b12Push(g, me.ID, "Sunset Strikemaster", "Creature — Human Monk", tdmStrikemasterOracle, 3, 1)
	walker := b12Creature(g, opp.ID, "Ground Bear", "Creature — Bear", 2, 2)
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Bird", TypeLine: "Creature — Bird", Power: 4, Toughness: 6,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "R", 3)
	if err := g.ActivateCatalogAbility(me.ID, monk, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker}},
	}); err == nil {
		t.Fatal("a creature without flying was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, monk, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: flyer}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(monk) {
		t.Fatal("the Strikemaster was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) || !opp.Graveyard.Contains(flyer) {
		t.Error("6 damage did not kill the 4/6 flyer")
	}
}

// --- Unburied Earthcarver ------------------------------------------------------

func TestTdmSZUnburiedEarthcarverSacrificesAnotherForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	carver := b12Push(g, me.ID, "Unburied Earthcarver", "Creature — Human Warrior", tdmEarthcarverOracle, 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	fillPoolColored(me, "B", 4)
	if err := g.ActivateCatalogAbility(me.ID, carver, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{carver}}); err == nil {
		t.Fatal("the Earthcarver sacrificed itself to its \"another creature\" cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, carver, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the other creature was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, carver, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Zurgo's Vanguard ----------------------------------------------------------

func TestTdmSZZurgosVanguardCountsCreaturesAndMobilizes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanguard := b12Push(g, me.ID, "Zurgo's Vanguard", "Creature — Dog Soldier", tdmZurgosVanguardOracle, 0, 3)
	if got := effectivePower(t, g, vanguard); got != 1 {
		t.Fatalf("alone, power %d, want 1 (itself)", got)
	}
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, vanguard); got != 2 {
		t.Fatalf("with one other creature of yours, power %d, want 2", got)
	}
	declareAttack(t, g, opp.ID, vanguard)
	passPriorityAroundTable(t, g)
	tdmPlaceWarriors(t, g, me.ID)
	if n := len(socB8Named(g, me.ID, "Warrior")); n != 1 {
		t.Fatalf("mobilize 1 made %d Warriors", n)
	}
	if got := effectivePower(t, g, vanguard); got != 3 {
		t.Errorf("with the Warrior, power %d, want 3", got)
	}
}
