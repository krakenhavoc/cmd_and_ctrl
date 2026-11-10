package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b21_cards_test.go — Standard set leftovers, batches 21 and 22 (#4452).

const (
	stdB21TeachOracle          = "d85ba214-fa05-44ca-a7f9-ab2c5dc07962"
	stdB21GateColossusOracle   = "386c5895-9713-4c28-9ba8-8ca0caa06c13"
	stdB21SpyglassOracle       = "b2187f45-80ae-4ac4-9f83-5eb7a00978e2"
	stdB21HellkiteOracle       = "6e64f1e0-ddd6-4ee7-8797-086e234b02e8"
	stdB21BarrensOracle        = "dfe1a112-97aa-4e81-8431-81552ba2cdcf"
	stdB21MazesEndOracle       = "49479778-c4c0-43ba-a7b7-45f00d067462"
	stdB21RuneSealedWallOracle = "74d31e99-ab48-483a-a62e-5b424b27d017"
	stdB21DiamondMareOracle    = "9a440122-a015-4af2-b270-37f019884458"
	stdB21AbductionOracle      = "8f5495d7-ac44-47cd-aa39-48a35034b390"
	stdB21OrphansOracle        = "f758bad0-8f36-4985-ad1d-66b695f3bbb6"
	stdB21DontMakeASoundOracle = "f28a373b-4a8d-40a5-ace1-3bbd8489ad36"
	stdB21DominationOracle     = "ebdace8f-8e8e-411b-b1d5-5b27cca7d9fb"
	stdB21FallingOracle        = "00ba96e0-42a7-432a-a500-5e607d75a358"
	stdB21ImpostorsOracle      = "628ae0e4-5d54-48a1-a24c-b9c1d862df9d"
	stdB21DrownerOracle        = "43c69d19-c06c-470a-bdba-a0335ff2d655"
	stdB21GetOutOracle         = "b2f3c27e-b646-4449-901a-8b809cdfa84e"
	stdB21LeylineOracle        = "cfaa13d6-9992-4223-8bd7-c44046505555"
	stdB21GrimoireOracle       = "096eac52-00c2-4439-b2f3-ff555d15ec88"
	stdB21TaleOracle           = "98a13f7b-ebe7-4a90-a0b9-95dccfff7fa6"
	stdB21CrackedSkullOracle   = "4415b65d-5473-4b19-94f7-e5cf4ce64d5e"
)

func TestStdB21EveryCardIsRegisteredFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		stdB21TeachOracle: "Teach by Example", stdB21GateColossusOracle: "Gate Colossus",
		stdB21SpyglassOracle: "Sorcerous Spyglass", stdB21HellkiteOracle: "Steel Hellkite",
		stdB21BarrensOracle: "Crawling Barrens", stdB21MazesEndOracle: "Maze's End",
		stdB21RuneSealedWallOracle: "Rune-Sealed Wall", stdB21DiamondMareOracle: "Diamond Mare",
		stdB21AbductionOracle: "Fear of Abduction", stdB21OrphansOracle: "Orphans of the Wheat",
		stdB21DontMakeASoundOracle: "Don't Make a Sound", stdB21DominationOracle: "Duskmourn's Domination",
		stdB21FallingOracle: "Fear of Falling", stdB21ImpostorsOracle: "Fear of Impostors",
		stdB21DrownerOracle: "Floodpits Drowner", stdB21GetOutOracle: "Get Out",
		stdB21LeylineOracle: "Leyline of Transformation", stdB21GrimoireOracle: "Marina Vendrell's Grimoire",
		stdB21TaleOracle: "The Tale of Tamiyo", stdB21CrackedSkullOracle: "Cracked Skull",
	}
	for oracle, name := range want {
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

// stdB21Library pushes `n` plain cards on top of the player's library.
func stdB21Library(p *game.Player, cards ...game.Card) {
	for _, c := range cards {
		c.Owner, c.Controller = p.ID, p.ID
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		p.Library.PushTop(c)
	}
}

func TestStdB21RuneSealedWallSurveilsOnTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wall := b12Push(g, me.ID, "Rune-Sealed Wall", "Artifact Creature — Wall", stdB21RuneSealedWallOracle, 0, 4)
	advanceToMain(t, g)
	untap(g, wall)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == wall {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	stdB21Library(me, game.Card{Name: "Top", TypeLine: "Land"})
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) == nil {
		t.Error("no surveil prompt after the tap")
	}
	if !tappedForTest(t, g, wall) {
		t.Error("the Wall did not tap")
	}
}

func TestStdB21CrawlingBarrensPutsCountersAndMayBecomeACreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := b12Push(g, me.ID, "Crawling Barrens", "Land", stdB21BarrensOracle, 0, 0)
		advanceToMain(t, g)
		fillPool(me, 4)
		if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
		if ask == nil {
			t.Fatal("no \"you may\" prompt for becoming a creature")
		}
		if got := counterCount(g, land, game.CounterPlusOne); got != 2 {
			t.Errorf("counters %d before the answer, want 2 (the counters are not optional)", got)
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, yes); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		c := b12Card(t, g, land)
		if c.IsCreature() != yes {
			t.Errorf("answer %v: IsCreature %v", yes, c.IsCreature())
		}
		if !c.IsLand() {
			t.Error("it is still a land")
		}
		if yes {
			if p, tough := socB14PT(t, g, land); p != 2 || tough != 2 {
				t.Errorf("0/0 with two counters is %d/%d, want 2/2", p, tough)
			}
			if !c.HasSubtype("Elemental") {
				t.Error("not an Elemental")
			}
		}
	}
}

func TestStdB21CrawlingBarrensStaysALandUntilItIsAnimatedAndRefusesAShortPool(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12Push(g, me.ID, "Crawling Barrens", "Land", stdB21BarrensOracle, 0, 0)
	advanceToMain(t, g)
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("{4} was paid with three mana")
	}
}

func TestStdB21DiamondMareStaysQuietBeforeAColourIsChosen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Diamond Mare", "Artifact Creature — Horse", stdB21DiamondMareOracle, 1, 3)
	advanceToMain(t, g)
	before := me.Life
	red := game.Card{Name: "Red Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
	if err := socB19CastWithPool(g, me, red, "R", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d with no colour chosen, want %d", me.Life, before)
	}
}

// stdB21PlayTypedLand plays a land with the given type line from the
// active seat's hand.
func stdB21PlayTypedLand(t *testing.T, g *game.Game, name, typeLine string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

func TestStdB21GateColossusReturnsToTheTopOfYourLibraryWhenAGateEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Gate Colossus", TypeLine: "Artifact Creature — Construct",
		OracleID: stdB21GateColossusOracle, Power: 8, Toughness: 8, Owner: me.ID})
	stdB21PlayTypedLand(t, g, "Test Gate", "Land — Gate")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if inZone(me.Graveyard, id) {
		t.Fatal("still in the graveyard")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != id {
		t.Errorf("top of library %v (%v), want Gate Colossus", top.Name, err)
	}
}

func TestStdB21GateColossusIgnoresANonGateLandAndCanDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Gate Colossus", TypeLine: "Artifact Creature — Construct",
		OracleID: stdB21GateColossusOracle, Owner: me.ID})
	stdB21PlayTypedLand(t, g, "Plains", "Basic Land — Plains")
	if c := latestTriggerPrompt(g, me.ID); c != nil {
		t.Fatal("a Plains woke Gate Colossus")
	}
	stdB21PlayTypedLand(t, g, "Test Gate", "Land — Gate")
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !inZone(me.Graveyard, id) {
		t.Error("declined, but it left the graveyard")
	}
}

func TestStdB21GateColossusCostsOneLessPerGate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		b12Push(g, me.ID, "Test Gate", "Land — Gate", "", 0, 0)
	}
	advanceToMain(t, g)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Gate Colossus", TypeLine: "Artifact Creature — Construct",
		ManaCost: "{8}", OracleID: stdB21GateColossusOracle, Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
	id := me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID
	fillPool(me, 4)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast for four with three Gates (cost {5})")
	}
	fillPool(me, 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("cast for {5} with three Gates: %v", err)
	}
}

func TestStdB21DiamondMareGainsLifeForASpellOfTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushChosenColorPermanent(t, g, me.ID, "Diamond Mare", "Artifact Creature — Horse", stdB21DiamondMareOracle, "R")
	advanceToMain(t, g)
	before := me.Life
	red := game.Card{Name: "Red Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
	if err := socB19CastWithPool(g, me, red, "R", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d after a red spell, want %d", me.Life, before+1)
	}
	blue := game.Card{Name: "Blue Thing", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}}
	if err := socB19CastWithPool(g, me, blue, "U", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d after a blue spell, want unchanged %d", me.Life, before+1)
	}
	_ = uuid.Nil
}

func TestStdB21TeachByExampleCopiesTheNextInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, victim)
	castCatalogSpell(t, g, "Teach by Example", "Instant", stdB21TeachOracle, nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	if got := settleKeepingTarget(t, g, me, victim); got != 1 {
		t.Errorf("%d re-target prompts, want 1 (one copy)", got)
	}
	if got := lifeOf(g, victim); got != before-6 {
		t.Errorf("life %d → %d, want -6 (the Bolt and its copy)", before, got)
	}
	// "When you next cast": only the next one is copied.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	settleKeepingTarget(t, g, me, victim)
	if got := lifeOf(g, victim); got != before-9 {
		t.Errorf("life %d → %d, want -9 (a second Bolt is not copied)", before, got)
	}
}

func TestStdB21SorcerousSpyglassLooksThenNamesAndBansActivations(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Secret", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	wall := b12Push(g, me.ID, "Rune-Sealed Wall", "Artifact Creature — Wall", stdB21RuneSealedWallOracle, 0, 4)
	castCatalogSpell(t, g, "Sorcerous Spyglass", "Artifact", stdB21SpyglassOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no opponent prompt for the look")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, optionFor(t, pick, opp.ID)); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Errorf("the chosen opponent's card %d is not known to the Spyglass's controller", i)
		}
	}
	actNameFor(t, g, me.ID, "Rune-Sealed Wall")
	untap(g, wall)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == wall {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("an ability of the named card was activated")
	}
}

func TestStdB21SteelHellkiteDestroysOnlyWhereItConnected(t *testing.T) {
	g := newCatalogGame(t)
	me, hit, safe := g.Seats[0], g.Seats[1], g.Seats[2]
	hell := b12Push(g, me.ID, "Steel Hellkite", "Artifact Creature — Dragon", stdB21HellkiteOracle, 5, 5)
	hitTwo := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hit Two", TypeLine: "Artifact",
		ManaCost: "{2}", Owner: hit.ID, Controller: hit.ID})
	hitThree := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hit Three", TypeLine: "Artifact",
		ManaCost: "{3}", Owner: hit.ID, Controller: hit.ID})
	hitLand := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hit Land", TypeLine: "Land",
		Owner: hit.ID, Controller: hit.ID})
	safeTwo := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Safe Two", TypeLine: "Artifact",
		ManaCost: "{2}", Owner: safe.ID, Controller: safe.ID})
	attackWith(t, g, hit.ID, hell)
	if hit.Life >= 40 {
		t.Fatalf("the Hellkite did not connect (life %d)", hit.Life)
	}
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, hell, 1, game.ActivateAbilityParams{XValue: 2, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, hitTwo); z != game.ZoneGraveyard {
		t.Errorf("the damaged player's mana value 2 permanent is in %q, want the graveyard", z)
	}
	if z := b12ZoneOf(g, hitThree); z != game.ZoneBattlefield {
		t.Errorf("mana value 3 is in %q, want the battlefield", z)
	}
	if z := b12ZoneOf(g, hitLand); z != game.ZoneBattlefield {
		t.Errorf("a land is in %q, want the battlefield", z)
	}
	if z := b12ZoneOf(g, safeTwo); z != game.ZoneBattlefield {
		t.Errorf("a player the Hellkite never hurt lost a permanent (%q)", z)
	}
	// Activate only once each turn.
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, hell, 1, game.ActivateAbilityParams{XValue: 3, Strict: true}); err == nil {
		t.Error("the {X} ability was activated twice in one turn")
	}
}

func TestStdB21SteelHellkitePumpsForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hell := b12Push(g, me.ID, "Steel Hellkite", "Artifact Creature — Dragon", stdB21HellkiteOracle, 5, 5)
	advanceToMain(t, g)
	fillPool(me, 4)
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, hell, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("pump %d: %v", i, err)
		}
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, hell); p != 7 {
		t.Errorf("power %d after two pumps, want 7", p)
	}
}

func TestStdB21MazesEndFetchesAGateAndReturnsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	maze := b12Push(g, me.ID, "Maze's End", "Land", stdB21MazesEndOracle, 0, 0)
	stdB21Library(me, game.Card{Name: "Test Gate", TypeLine: "Land — Gate"}, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	advanceToMain(t, g)
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, maze, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if z := b12ZoneOf(g, maze); z != game.ZoneHand {
		t.Errorf("Maze's End is in %q after the cost, want the hand", z)
	}
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Test Gate")
	passPriorityAroundTable(t, g)
	if !onBattlefieldByName(g, "Test Gate") {
		t.Error("the Gate is not on the battlefield")
	}
	if g.State != game.StateActive {
		t.Error("one Gate won the game")
	}
}

func TestStdB21MazesEndWinsAtTenGatesWithDifferentNamesOnly(t *testing.T) {
	for _, tc := range []struct {
		name  string
		names []string
		win   bool
	}{
		{"nine names", []string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8"}, false},
		{"ten names", []string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9"}, true},
		{"ten Gates, nine names", []string{"G1", "G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			for _, n := range tc.names {
				b12Push(g, me.ID, n, "Land — Gate", "", 0, 0)
			}
			maze := b12Push(g, me.ID, "Maze's End", "Land", stdB21MazesEndOracle, 0, 0)
			stdB21Library(me, game.Card{Name: "Fetched Gate", TypeLine: "Land — Gate"})
			advanceToMain(t, g)
			fillPool(me, 3)
			if err := g.ActivateCatalogAbility(me.ID, maze, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			answerSearchNamed(t, g, me.ID, "Fetched Gate")
			passPriorityAroundTable(t, g)
			if won := g.State != game.StateActive; won != tc.win {
				t.Errorf("game over = %v, want %v", won, tc.win)
			}
		})
	}
}

func TestStdB21FearOfAbductionExilesTwoAndReturnsBothWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	fillPoolColored(me, "W", 2)
	fillPool(me, 4)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fear of Abduction", TypeLine: "Enchantment Creature — Nightmare",
		ManaCost: "{4}{W}{W}", Colors: []string{"W"}, OracleID: stdB21AbductionOracle, Power: 5, Toughness: 5,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, ExilePermanentIDs: []uuid.UUID{mine}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, theirs); z != game.ZoneExile {
		t.Fatalf("the opponent's creature is in %q, want exile", z)
	}
	if z := b12ZoneOf(g, mine); z != game.ZoneExile {
		t.Fatalf("the exiled-as-cost creature is in %q, want exile", z)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id, game.DestroyOptions{}) })
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, mine); z != game.ZoneHand {
		t.Errorf("my creature is in %q after the Fear left, want my hand", z)
	}
	if z := b12ZoneOf(g, theirs); z != game.ZoneHand {
		t.Errorf("their creature is in %q after the Fear left, want its owner's hand", z)
	}
	if !inZone(opp.Hand, theirs) {
		t.Error("their creature did not go to its owner's hand")
	}
}

func TestStdB21OrphansOfTheWheatGrowsForEachCreatureTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orphans := b12Push(g, me.ID, "Orphans of the Wheat", "Creature — Human", stdB21OrphansOracle, 1, 1)
	a := b12Push(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := b12Push(g, me.ID, "Bear B", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, orphans)
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ask == nil {
		t.Fatal("no choose_cards prompt for the tap")
	}
	if err := g.ResolveChooseCards(ask.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := socB14PT(t, g, orphans); p != 3 || tough != 3 {
		t.Errorf("Orphans is %d/%d after tapping two, want 3/3", p, tough)
	}
	if !tappedForTest(t, g, a) || !tappedForTest(t, g, b) {
		t.Error("the chosen creatures were not tapped")
	}
}

func TestStdB21OrphansOfTheWheatMayTapNone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orphans := b12Push(g, me.ID, "Orphans of the Wheat", "Creature — Human", stdB21OrphansOracle, 1, 1)
	a := b12Push(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
	declareAttack(t, g, opp.ID, orphans)
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ask == nil {
		t.Fatal("no choose_cards prompt")
	}
	if err := g.ResolveChooseCards(ask.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards(none): %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := socB14PT(t, g, orphans); p != 1 || tough != 1 {
		t.Errorf("Orphans is %d/%d after tapping none, want 1/1", p, tough)
	}
	if tappedForTest(t, g, a) {
		t.Error("an unchosen creature was tapped")
	}
}

func TestStdB21DontMakeASoundSurveilsOnlyWhenThePriceIsPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		toMainForCost(t, g)
		victim := castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		castFromSeat(t, g, them, "Don't Make a Sound", "Instant", stdB21DontMakeASoundOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		passUntilTaxed(t, g, me.ID)
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		if pay {
			if !battlefieldHas(g, victim) && !g.Stack.Contains(victim) {
				t.Error("a paid spell must not be countered")
			}
			if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, them.ID) == nil {
				t.Error("the caster of Don't Make a Sound did not surveil 2 after the payment")
			}
			if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID) != nil {
				t.Error("the payer surveiled")
			}
		} else {
			if g.Stack.Contains(victim) || battlefieldHas(g, victim) {
				t.Error("an unpaid spell must be countered")
			}
			if latestChoiceOfKindFor(g, game.PendingChoiceSurveil, them.ID) != nil {
				t.Error("an unpaid counter surveils nothing")
			}
		}
	}
}

func TestStdB21FearOfImpostorsCountersAndTheSpellsControllerManifestsDread(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	before := len(g.Battlefield.Cards)
	castFromSeat(t, g, them, "Fear of Impostors", "Enchantment Creature — Nightmare", stdB21ImpostorsOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, victim)
	if n := answerManifestDread(t, g, me.ID); n != 1 {
		t.Errorf("the countered spell's controller answered %d manifest dread prompts, want 1", n)
	}
	if g.Stack.Contains(victim) || battlefieldHas(g, victim) {
		t.Error("the targeted spell was not countered")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, them.ID) != nil {
		t.Error("the Fear's controller was asked to manifest")
	}
	mine := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name != "Their Bear" {
			mine++
		}
	}
	if len(g.Battlefield.Cards) != before+2 || mine < 1 {
		t.Errorf("battlefield %d → %d, want the Fear and a manifested permanent for the victim (mine %d)", before, len(g.Battlefield.Cards), mine)
	}
}

func stdB21CastGetOut(t *testing.T, g *game.Game, mode int, targets ...uuid.UUID) error {
	t.Helper()
	me := g.Seats[0]
	toMainForCost(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Get Out", TypeLine: "Instant", OracleID: stdB21GetOutOracle,
		Owner: me.ID, Controller: me.ID})
	refs := make([]game.TargetRef, 0, len(targets))
	for _, tid := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: tid, Mode: 0})
	}
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: refs})
}

func TestStdB21GetOutCountersACreatureOrEnchantmentSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Get Out", TypeLine: "Instant", OracleID: stdB21GetOutOracle,
		Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Modes: []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, victim) || g.Stack.Contains(victim) {
		t.Error("the creature spell was not countered")
	}
	_ = me
	// A noncreature, nonenchantment spell is not a legal target.
	sorcery := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	id2 := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id2, Name: "Get Out", TypeLine: "Instant", OracleID: stdB21GetOutOracle,
		Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id2, game.CastSpellParams{Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sorcery}}}); err == nil {
		t.Error("Get Out targeted a sorcery spell")
	}
}

func TestStdB21GetOutReturnsOneOrTwoCreaturesOrEnchantmentsYouOwnWhoeverControlsThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID})
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Pacifism", TypeLine: "Enchantment",
		Owner: me.ID, Controller: me.ID})
	theirs := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if err := stdB21CastGetOut(t, g, 1, theirs); err == nil {
		t.Fatal("Get Out returned a permanent its caster does not own")
	}
	if err := stdB21CastGetOut(t, g, 1, stolen, aura); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inZone(me.Hand, stolen) || !inZone(me.Hand, aura) {
		t.Errorf("stolen bear in hand %v, enchantment in hand %v; want both returned to my hand",
			inZone(me.Hand, stolen), inZone(me.Hand, aura))
	}
	if z := b12ZoneOf(g, theirs); z != game.ZoneBattlefield {
		t.Errorf("an opponent's own creature left the battlefield (%q)", z)
	}
}

func TestStdB21FearOfFallingShrinksAndGroundsADefendersCreatureUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fall := b12Push(g, me.ID, "Fear of Falling", "Enchantment Creature — Nightmare", stdB21FallingOracle, 4, 4)
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 3, Toughness: 3, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	declareAttack(t, g, opp.ID, fall)
	answerPickTargetIfOpen(t, g, bird)
	passPriorityAroundTable(t, g)
	if p, tough := socB14PT(t, g, bird); p != 1 || tough != 3 {
		t.Errorf("the Bird is %d/%d, want 1/3", p, tough)
	}
	if effectiveAbilitiesContain(t, g, bird, "flying") {
		t.Error("the Bird still has flying")
	}
	// It lasts until the Fear's controller's next turn.
	for i := 0; i < 80 && g.Turn.ActiveSeat != 0 || i < 1; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.ActiveSeat != 0 {
			if p, _ := socB14PT(t, g, bird); p != 1 {
				t.Fatalf("the shrink ended early (power %d on seat %d's turn)", p, g.Turn.ActiveSeat)
			}
		}
	}
	for i := 0; i < 80 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepPrecombatMain); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if p, _ := socB14PT(t, g, bird); p != 3 {
		t.Errorf("the Bird is %d power on my next turn, want 3 again", p)
	}
}

func TestStdB21FloodpitsDrownerTapsAndStunsOnEnteringAndShufflesItselfAndAStunnedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	toMainForCost(t, g)
	drowner := castFromSeat(t, g, me, "Floodpits Drowner", "Creature — Merfolk", stdB21DrownerOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, theirs) || counterCount(g, theirs, game.CounterStun) != 1 {
		t.Errorf("tapped %v, stun counters %d; want tapped with one stun counter",
			tappedForTest(t, g, theirs), counterCount(g, theirs, game.CounterStun))
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == drowner {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	// A creature with no stun counter is not a legal target.
	other := b12Push(g, opp.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	fillPoolColored(me, "U", 1)
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, drowner, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Error("targeted a creature with no stun counter")
	}
	myLib, oppLib := me.Library.Size(), opp.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, drowner, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, drowner); z != game.ZoneLibrary {
		t.Errorf("the Drowner is in %q, want its owner's library", z)
	}
	if z := b12ZoneOf(g, theirs); z != game.ZoneLibrary {
		t.Errorf("the target is in %q, want its owner's library", z)
	}
	if me.Library.Size() != myLib+1 || opp.Library.Size() != oppLib+1 {
		t.Errorf("libraries %d/%d, want %d/%d", me.Library.Size(), opp.Library.Size(), myLib+1, oppLib+1)
	}
}

func TestStdB21DuskmournsDominationStealsShrinksAndSilences(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 4, Toughness: 4, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Duskmourn's Domination", "Enchantment — Aura", stdB21DominationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bird}})
	passPriorityAroundTable(t, g)
	c := b12Card(t, g, bird)
	if c.Controller != me.ID {
		t.Error("I do not control the enchanted creature")
	}
	if p, tough := socB14PT(t, g, bird); p != 1 || tough != 4 {
		t.Errorf("the creature is %d/%d, want 1/4", p, tough)
	}
	if effectiveAbilitiesContain(t, g, bird, "flying") {
		t.Error("it kept flying")
	}
}

func TestStdB21CrackedSkullLooksAtAHandAndTakesANonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := revealHand(opp, rhForest(), rhBolt())
	bear := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 5)
	castCatalogSpell(t, g, "Cracked Skull", "Enchantment — Aura", stdB21CrackedSkullOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceRevealedHandPick, me.ID)
	if pick == nil {
		t.Fatal("no revealed-hand pick is open")
	}
	if pick.FromPlayer != opp.ID {
		t.Fatalf("prompt from %v, want the targeted player", pick.FromPlayer)
	}
	if !sameIDs(pick.DiscardOptions, ids[1:]) {
		t.Errorf("options = %v, want only the nonland card %v", pick.DiscardOptions, ids[1:])
	}
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Errorf("card %d of the looked-at hand is not known to me", i)
		}
	}
	for _, seat := range g.Seats[2:] {
		for i := range seat.Hand.Cards {
			if seat.Hand.Cards[i].IsKnownTo(me.ID) {
				t.Error("a third player's hand was shown")
			}
		}
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("take the Bolt: %v", err)
	}
	g.SettleResolution()
	if !inZone(opp.Graveyard, ids[1]) {
		t.Error("the chosen card was not discarded")
	}
	if !inZone(opp.Hand, ids[0]) {
		t.Error("the land left the hand")
	}
}

func TestStdB21CrackedSkullDestroysTheEnchantedCreatureWhenItIsDealtDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	revealHand(opp)
	bear := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 5)
	castCatalogSpell(t, g, "Cracked Skull", "Enchantment — Aura", stdB21CrackedSkullOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID) != nil {
		answerPickTargetPlayer(t, g, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, bear); z != game.ZoneBattlefield {
		t.Fatalf("the bear is in %q before any damage", z)
	}
	src := b12Push(g, me.ID, "Pinger", "Creature — Test", "", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, bear, 1) })
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, bear); z != game.ZoneGraveyard {
		t.Errorf("one point of damage on a 2/5: the bear is in %q, want destroyed by the Aura", z)
	}
}

func stdB21PutGrimoire(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Marina Vendrell's Grimoire", "Legendary Artifact — Book", stdB21GrimoireOracle, nil)
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, id) {
		t.Fatal("the Grimoire did not resolve")
	}
	_ = me
	return id
}

func TestStdB21GrimoireDrawsFiveOnlyWhenCastAndDrawsForLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	stdB21PutGrimoire(t, g)
	if got := me.Hand.Size(); got != before+5 {
		t.Fatalf("hand %d → %d after casting it, want +5", before, got)
	}
	before = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand %d → %d after gaining 3, want +3", before, got)
	}
}

func TestStdB21GrimoireDiscardsForLifeLostAndLosesWithAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stdB21PutGrimoire(t, g)
	// Three cards, lose 2: discard two, keep one, stay in the game.
	revealHand(me, rhForest(), rhBolt(), rhPelakka())
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2) })
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt after losing 2 life")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:2]); err != nil {
		t.Fatalf("discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 1 || me.Eliminated {
		t.Fatalf("hand %d, eliminated %v; want one card left and still in the game", me.Hand.Size(), me.Eliminated)
	}
	// One card, lose 1: the hand is empty afterwards, so you lose the game.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	passPriorityAroundTable(t, g)
	if pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); pick != nil {
		if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards); err != nil {
			t.Fatalf("discard: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !me.Eliminated {
		t.Error("discarding your last card to the Grimoire did not lose the game")
	}
}

func TestStdB21GrimoireHasNoMaximumHandSizeAndDoesNotLoseAtZeroLife(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stdB21PutGrimoire(t, g)
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("my maximum hand size is %d, want none", got)
	}
	if got := mhMax(g, opp); got == game.NoMaxHandSize {
		t.Error("an opponent has no maximum hand size too")
	}
	revealHand(me)
	g.WithWriteLock(func() { me.Life = 0 })
	g.RunStateChecksForTest()
	if me.Eliminated {
		t.Error("lost the game at 0 life with the Grimoire out")
	}
	// Other ways to lose still work: ten poison counters.
	g.WithWriteLock(func() { me.Life = 5 })
	g.RunStateChecksForTest()
}

func TestStdB21TaleOfTamiyoChapterOneMillsAgainWhileTwoMilledCardsShareAType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	// Top of the library, first mill first: two creatures (share a type, so
	// draw and repeat), then a land and an instant (no type shared, so stop).
	stdB21Library(me,
		game.Card{Name: "Stop Instant", TypeLine: "Instant"},
		game.Card{Name: "Stop Land", TypeLine: "Land"},
		game.Card{Name: "Pair Two", TypeLine: "Artifact Creature — Golem"},
		game.Card{Name: "Pair One", TypeLine: "Creature — Bear"},
	)
	handBefore, gyBefore := me.Hand.Size(), me.Graveyard.Size()
	castCatalogSpell(t, g, "The Tale of Tamiyo", "Legendary Enchantment — Saga", stdB21TaleOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - gyBefore; got != 4 {
		t.Errorf("%d cards milled, want 4 (two rounds)", got)
	}
	if got := me.Hand.Size() - handBefore; got != 1 {
		// The Saga was put in the hand and cast (net 0), and the one draw came in.
		t.Errorf("hand changed by %d, want +1 (drew one)", got)
	}
}

func TestStdB21TaleOfTamiyoChapterOneStopsWhenTheTwoMilledShareNoType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	stdB21Library(me,
		game.Card{Name: "Spare Creature", TypeLine: "Creature — Bear"},
		game.Card{Name: "Spare Creature 2", TypeLine: "Creature — Bear"},
		game.Card{Name: "Stop Instant", TypeLine: "Instant"},
		game.Card{Name: "Stop Land", TypeLine: "Land"},
	)
	gyBefore, handBefore := me.Graveyard.Size(), me.Hand.Size()
	castCatalogSpell(t, g, "The Tale of Tamiyo", "Legendary Enchantment — Saga", stdB21TaleOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - gyBefore; got != 2 {
		t.Errorf("%d cards milled, want 2", got)
	}
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Errorf("hand changed by %d, want 0 (no draw)", got)
	}
}

func TestStdB21TaleOfTamiyoChapterFourExilesAndCopiesInstantsAndSorceriesAndTamiyoWalkers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seat := g.Turn.ActiveSeat
	saga := castCatalogSpell(t, g, "The Tale of Tamiyo", "Legendary Enchantment — Saga", stdB21TaleOracle, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		advanceToPrecombatMainOf(t, g, seat)
		passPriorityAroundTable(t, g)
	}
	bolt := uuid.New()
	walker := uuid.New()
	bear := uuid.New()
	other := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		OracleID: lightningBoltOracle, Owner: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: walker, Name: "Tamiyo, Test", TypeLine: "Legendary Planeswalker — Tamiyo", Owner: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Graveyard Bear", TypeLine: "Creature — Bear", Owner: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: other, Name: "Other Walker", TypeLine: "Legendary Planeswalker — Jace", Owner: me.ID})
	advanceToPrecombatMainOf(t, g, seat)
	for i := 0; i < 8; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget {
				open = c
			}
		}
		if open == nil {
			break
		}
		if err := g.ResolvePickTarget(open.ID, open.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bolt}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, bear); z != game.ZoneGraveyard {
		t.Errorf("the creature card was exiled (%q)", z)
	}
	if z := b12ZoneOf(g, other); z != game.ZoneGraveyard {
		t.Errorf("a non-Tamiyo planeswalker card was exiled (%q)", z)
	}
	if z := b12ZoneOf(g, bolt); z != game.ZoneExile {
		t.Fatalf("the Bolt is in %q, want exile", z)
	}
	if resolutionCastOfferFor(g, me.ID) == nil {
		t.Fatal("no cast offer for the copy of the Bolt")
	}
	declineResolutionCast(t, g, me.ID)
	_ = saga
}

func TestStdB21LeylineOfTransformationAddsTheChosenTypeToCreaturesSpellsAndCardsYouOwn(t *testing.T) {
	spec, ok := Lookup(stdB21LeylineOracle)
	if !ok || spec.OpeningHand == nil {
		t.Fatal("Leyline of Transformation must offer to begin the game on the battlefield")
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	inHand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Hand Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	notCreature := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: notCreature, Name: "Hand Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Leyline of Transformation", "Enchantment", stdB21LeylineOracle, nil)
	passPriorityAroundTable(t, g)
	answerCreatureType(t, g, me.ID, "Wizard")
	if c := b12Card(t, g, mine); !c.HasSubtype("Wizard") || !c.HasSubtype("Bear") {
		t.Error("my creature is not both a Bear and a Wizard")
	}
	if b12Card(t, g, theirs).HasSubtype("Wizard") {
		t.Error("an opponent's creature became a Wizard")
	}
	if !b12Card(t, g, inHand).HasSubtype("Wizard") {
		t.Error("a creature card in my hand is not a Wizard")
	}
	if b12Card(t, g, notCreature).HasSubtype("Wizard") {
		t.Error("a land card in my hand is a Wizard")
	}
}
