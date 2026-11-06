package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_dg_cards_test.go — The Lost Caverns of Ixalan, cards D–G (#4470).
// Each test plays the real catalog entry and asserts what its text says.

const (
	lciDGDeadWeightOracle    = "b1804304-fac1-4b19-a48d-6ade9407972a"
	lciDGSkulltakerOracle    = "b7f56bbb-961a-4641-9446-7292ef188b1d"
	lciDGScallywagOracle     = "97607bf5-ca7b-4b1b-963e-390faf6e413e"
	lciDGMarionetteOracle    = "6c4247d0-6dc8-4139-bbf5-c83dfdc7be1f"
	lciDGDidactOracle        = "8151668c-6459-41b5-9e57-bb47c3876b81"
	lciDGDinotomatonOracle   = "443f2272-6475-4966-8ca2-717d4cf208ee"
	lciDGWanderglyphOracle   = "81dda1a8-72ac-4e69-8367-4f72c445197c"
	lciDGEchoOfDuskOracle    = "fd0980ae-eabf-45da-98e7-23d81bfa185a"
	lciDGEnvoyOracle         = "ebc70734-bf1a-4754-8094-bcbb4db91f30"
	lciDGFamilyReunionOracle = "8972de18-2c49-4162-a285-264a44c0e675"
	lciDGOfferingOracle      = "3e5402bf-ce7b-4c23-9333-0acdcf349caf"
	lciDGCaveWurmOracle      = "ac5294c6-bc79-4a0f-b8b3-52b554561e38"
	lciDGGlimpseOracle       = "fa44fa48-d270-4f0f-b672-d63e1c1942ce"
	lciDGTombRaiderOracle    = "37ce95bb-8ff5-47b9-85e4-21be7a794940"
)

func TestLciDGEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciDGDeadWeightOracle:    "Dead Weight",
		lciDGSkulltakerOracle:    "Deep Goblin Skulltaker",
		lciDGScallywagOracle:     "Enterprising Scallywag",
		lciDGMarionetteOracle:    "Deathcap Marionette",
		lciDGDidactOracle:        "Didact Echo",
		lciDGDinotomatonOracle:   "Dinotomaton",
		lciDGWanderglyphOracle:   "Disruptor Wanderglyph",
		lciDGEchoOfDuskOracle:    "Echo of Dusk",
		lciDGEnvoyOracle:         "Envoy of Okinec Ahau",
		lciDGFamilyReunionOracle: "Family Reunion",
		lciDGOfferingOracle:      "Fanatical Offering",
		lciDGCaveWurmOracle:      "Frilled Cave-Wurm",
		lciDGGlimpseOracle:       "Glimpse the Core",
		lciDGTombRaiderOracle:    "Goblin Tomb Raider",
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

// lciDGSeedDescend puts three permanent cards and a sorcery into p's
// graveyard and refreshes the layers: one short of descend 4, with the
// sorcery not counted (CR 110.4).
func lciDGSeedDescend(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() {
		lciSTPermanentCardsInGraveyard(p, 3)
		pushGraveyardCardTyped(p, "Old Sorcery", "Sorcery")
		g.BumpLayerVersionForTest()
	})
}

// lciDGMillOne mills one card of the given type line from p's library, a
// descent when it is a permanent card (CR 700.11).
func lciDGMillOne(t *testing.T, g *game.Game, p *game.Player, typeLine string) {
	t.Helper()
	ssCard(p.Library, p.ID, "Milled Card", typeLine, 2, 2)
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(p.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
}

// --- Dead Weight -----------------------------------------------------------------

func TestLciDGDeadWeightShrinksAndKills(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	castCatalogSpell(t, g, "Dead Weight", "Enchantment — Aura", lciDGDeadWeightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("the enchanted giant is %d/%d, want 2/2", p, tough)
	}

	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Dead Weight", "Enchantment — Aura", lciDGDeadWeightOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: small}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !opp.Graveyard.Contains(small) {
		t.Error("a 2/2 given -2/-2 did not die")
	}
	if got := countBattlefieldNamed(g, me.ID, "Dead Weight"); got != 1 {
		t.Errorf("%d Dead Weights on the battlefield, want 1 (the one whose host died goes to the graveyard)", got)
	}
}

// --- Deathcap Marionette ---------------------------------------------------------

func TestLciDGDeathcapMarionetteMayMillTwo(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		library := me.Library.Size()
		marionette := castHoldCreature(t, g, "Deathcap Marionette", "Creature — Fungus", lciDGMarionetteOracle, 1, 1)
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		want := 0
		if accept {
			want = 2
		}
		if got := library - me.Library.Size(); got != want {
			t.Errorf("accept=%v: milled %d cards, want %d", accept, got, want)
		}
		if !hasEffectiveKeyword(t, g, marionette, "deathtouch") {
			t.Error("Deathcap Marionette has no deathtouch")
		}
	}
}

// --- Deep Goblin Skulltaker ------------------------------------------------------

func TestLciDGDeepGoblinSkulltakerGrowsAtYourEndStepIfYouDescended(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		want     int
	}{
		{"a permanent card milled", "Creature — Bear", 1},
		{"only an instant card milled", "Instant", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			skull := b12Push(g, me.ID, "Deep Goblin Skulltaker", "Creature — Goblin Warrior", lciDGSkulltakerOracle, 2, 2)
			if !hasEffectiveKeyword(t, g, skull, "menace") {
				t.Error("Deep Goblin Skulltaker has no menace")
			}
			lciDGMillOne(t, g, me, tc.typeLine)
			advanceToEndStepOf(t, g, 0)
			if tc.want == 0 && triggersOnStackFrom(g, skull) != 0 {
				t.Fatal("the end-step trigger went on the stack without a descent")
			}
			passPriorityAroundTable(t, g)
			if got := countersOn(g, skull, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

// --- Didact Echo -----------------------------------------------------------------

func TestLciDGDidactEchoDrawsAndFliesWithDescendFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	echo := msc17Play(t, g, me, game.Card{Name: "Didact Echo", TypeLine: "Creature — Spirit Cleric", OracleID: lciDGDidactOracle, Power: 3, Toughness: 2})
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards on entry, want 1", got)
	}
	lciDGSeedDescend(g, me)
	if hasEffectiveKeyword(t, g, echo, "flying") {
		t.Fatal("Didact Echo flies with three permanent cards in the graveyard")
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	if !hasEffectiveKeyword(t, g, echo, "flying") {
		t.Error("Didact Echo does not fly with four permanent cards in the graveyard")
	}
}

// --- Dinotomaton -----------------------------------------------------------------

func TestLciDGDinotomatonGivesACreatureYouControlMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	dino := castHoldCreature(t, g, "Dinotomaton", "Artifact Creature — Dinosaur Gnome", lciDGDinotomatonOracle, 4, 3)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no enters-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, theirs) || !hasID(pick.PickTargetCards, bear) || !hasID(pick.PickTargetCards, dino) {
		t.Errorf("offered %v; want the creatures you control only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("the targeted creature did not gain menace")
	}
	if !hasEffectiveKeyword(t, g, dino, "menace") {
		t.Error("Dinotomaton has no menace")
	}
}

// --- Disruptor Wanderglyph -------------------------------------------------------

func TestLciDGDisruptorWanderglyphExilesACardFromAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glyph := b12Push(g, me.ID, "Disruptor Wanderglyph", "Artifact Creature — Golem", lciDGWanderglyphOracle, 3, 4)
	mine := pushGraveyardCardTyped(me, "My Dead Bear", "Creature — Bear")
	theirs := pushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
	declareAttack(t, g, opp.ID, glyph)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, mine) || !hasID(pick.PickTargetCards, theirs) {
		t.Errorf("offered %v; want the opponent's graveyard card only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("the opponent's card was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("your own graveyard card moved")
	}
}

// --- Echo of Dusk ----------------------------------------------------------------

func TestLciDGEchoOfDuskGrowsAndGainsLifelinkWithDescendFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	echo := b12Push(g, me.ID, "Echo of Dusk", "Creature — Vampire Spirit", lciDGEchoOfDuskOracle, 2, 2)
	lciDGSeedDescend(g, me)
	if p, tough := effectivePower(t, g, echo), effectiveToughness(t, g, echo); p != 2 || tough != 2 {
		t.Errorf("with three permanent cards the Echo is %d/%d, want 2/2", p, tough)
	}
	if hasEffectiveKeyword(t, g, echo, "lifelink") {
		t.Error("lifelink with three permanent cards in the graveyard")
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	if p, tough := effectivePower(t, g, echo), effectiveToughness(t, g, echo); p != 3 || tough != 3 {
		t.Errorf("with four permanent cards the Echo is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, echo, "lifelink") {
		t.Error("no lifelink with four permanent cards in the graveyard")
	}
}

// --- Enterprising Scallywag ------------------------------------------------------

func TestLciDGEnterprisingScallywagMakesATreasureIfYouDescended(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		want     int
	}{
		{"a land card milled", "Basic Land — Mountain", 1},
		{"only a sorcery card milled", "Sorcery", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			b12Push(g, me.ID, "Enterprising Scallywag", "Creature — Goblin Pirate", lciDGScallywagOracle, 2, 2)
			lciDGMillOne(t, g, me, tc.typeLine)
			advanceToEndStepOf(t, g, 0)
			passPriorityAroundTable(t, g)
			if got := len(trekB33Tokens(g, me.ID, "Treasure")); got != tc.want {
				t.Errorf("%d Treasure tokens, want %d", got, tc.want)
			}
		})
	}
}

// A token dying is not a descent: a token is not a card (CR 700.11).
func TestLciDGEnterprisingScallywagIgnoresATokenDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Enterprising Scallywag", "Creature — Goblin Pirate", lciDGScallywagOracle, 2, 2)
	advanceToMain(t, g)
	var token uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(me.ID, TokenCard("1/1 colorless Gnome artifact"), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("CreateTokensForEffect: %v %v", ids, err)
		}
		token = ids[0]
	})
	killCreature(t, g, me.ID, token)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := len(trekB33Tokens(g, me.ID, "Treasure")); got != 0 {
		t.Errorf("%d Treasure tokens after only a token died, want 0", got)
	}
}

// --- Envoy of Okinec Ahau --------------------------------------------------------

func TestLciDGEnvoyOfOkinecAhauMakesAGnome(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	envoy := pushCatalogPermanent(g, me.ID, "Envoy of Okinec Ahau", "Creature — Cat Advisor", lciDGEnvoyOracle, false)
	advanceToMain(t, g)
	fillPoolColored(me, "W", 4)
	if err := g.ActivateCatalogAbility(me.ID, envoy, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with four mana for {4}{W}")
	}
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, envoy, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if tappedNow(g, envoy) {
		t.Error("the Envoy tapped; the cost has no {T}")
	}
	passPriorityAroundTable(t, g)
	gnomes := trekB33Tokens(g, me.ID, "Gnome")
	if len(gnomes) != 1 {
		t.Fatalf("%d Gnome tokens, want 1", len(gnomes))
	}
	if c := layeredCard(t, g, gnomes[0].InstanceID); !c.IsCreature() || !c.IsArtifact() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 {
		t.Errorf("token = %s %d/%d %v, want a 1/1 colorless Gnome artifact creature", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
}

// --- Family Reunion --------------------------------------------------------------

func TestLciDGFamilyReunionPumpsOrShieldsYourCreatures(t *testing.T) {
	t.Run("+1/+1", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		castModal(t, g, "Family Reunion", "Instant", lciDGFamilyReunionOracle, []int{0}, nil)
		passPriorityAroundTable(t, g)
		if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || tough != 3 {
			t.Errorf("your bear is %d/%d, want 3/3", p, tough)
		}
		if p, tough := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 2 || tough != 2 {
			t.Errorf("an opponent's bear is %d/%d, want 2/2", p, tough)
		}
		if hasEffectiveKeyword(t, g, mine, "hexproof") {
			t.Error("the +1/+1 mode also gave hexproof")
		}
	})
	t.Run("hexproof", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		castModal(t, g, "Family Reunion", "Instant", lciDGFamilyReunionOracle, []int{1}, nil)
		passPriorityAroundTable(t, g)
		if !hasEffectiveKeyword(t, g, mine, "hexproof") {
			t.Error("your bear did not gain hexproof")
		}
		if hasEffectiveKeyword(t, g, theirs, "hexproof") {
			t.Error("an opponent's bear gained hexproof")
		}
		if p := effectivePower(t, g, mine); p != 2 {
			t.Errorf("the hexproof mode also pumped: power %d, want 2", p)
		}
		// A creature that arrives after it resolves is not one it affected
		// (CR 611.2c).
		late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
		if hasEffectiveKeyword(t, g, late, "hexproof") {
			t.Error("a creature that entered afterwards gained hexproof")
		}
	})
}

// --- Fanatical Offering ----------------------------------------------------------

func TestLciDGFanaticalOfferingSacrificesThenDrawsTwoAndMakesAMap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := pushArtifactForTest(g, me.ID, "Mind Stone")
	hand := me.Hand.Size()
	castWithSacrifice(t, g, "Fanatical Offering", "Instant", lciDGOfferingOracle, rock)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed as the spell was cast")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if got := lciB1MapsControlledBy(g, me.ID); got != 1 {
		t.Errorf("%d Map tokens, want 1", got)
	}
}

func TestLciDGFanaticalOfferingNeedsSomethingToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Some Land", "Land", "", false)
	if _, err := tryCastWithSacrifice(g, "Fanatical Offering", "Instant", lciDGOfferingOracle, []uuid.UUID{land}); err == nil {
		t.Error("a land was accepted as the sacrifice")
	}
	if _, err := tryCastWithSacrifice(g, "Fanatical Offering", "Instant", lciDGOfferingOracle, nil); err == nil {
		t.Error("cast with nothing sacrificed")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land left the battlefield")
	}
}

// --- Frilled Cave-Wurm -----------------------------------------------------------

func TestLciDGFrilledCaveWurmGetsPlusTwoWithDescendFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wurm := b12Push(g, me.ID, "Frilled Cave-Wurm", "Creature — Salamander Wurm", lciDGCaveWurmOracle, 2, 5)
	lciDGSeedDescend(g, me)
	if p, tough := effectivePower(t, g, wurm), effectiveToughness(t, g, wurm); p != 2 || tough != 5 {
		t.Errorf("with three permanent cards the Wurm is %d/%d, want 2/5", p, tough)
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	if p, tough := effectivePower(t, g, wurm), effectiveToughness(t, g, wurm); p != 4 || tough != 5 {
		t.Errorf("with four permanent cards the Wurm is %d/%d, want 4/5", p, tough)
	}
}

// --- Glimpse the Core ------------------------------------------------------------

func TestLciDGGlimpseTheCoreFetchesABasicForestTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Temple Garden", "Land — Forest Plains"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
	)
	castModal(t, g, "Glimpse the Core", "Sorcery", lciDGGlimpseOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 1 {
		t.Fatalf("search offers %d cards, want only the basic Forest", len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !g.Battlefield.Contains(ids[0]) {
		t.Fatal("the Forest is not on the battlefield")
	}
	if !tappedNow(g, ids[0]) {
		t.Error("the Forest entered untapped")
	}
}

func TestLciDGGlimpseTheCoreReturnsACaveTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cave := pushGraveyardCardTyped(me, "Some Cave", "Land — Cave")
	other := pushGraveyardCardTyped(me, "Some Plains", "Basic Land — Plains")
	if _, err := castModalErr(t, g, "Glimpse the Core", "Sorcery", lciDGGlimpseOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: other}}); err == nil {
		t.Fatal("a land card that is not a Cave was accepted as the target")
	}
	castModal(t, g, "Glimpse the Core", "Sorcery", lciDGGlimpseOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: cave}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cave) {
		t.Fatal("the Cave was not returned to the battlefield")
	}
	if !tappedNow(g, cave) {
		t.Error("the Cave returned untapped")
	}
	if !me.Graveyard.Contains(other) {
		t.Error("the other land left the graveyard")
	}
}

// --- Goblin Tomb Raider ----------------------------------------------------------

func TestLciDGGoblinTombRaiderGetsPowerAndHasteWithAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := pushCatalogPermanent(g, me.ID, "Goblin Tomb Raider", "Creature — Goblin Pirate", lciDGTombRaiderOracle, true)
	if p := effectivePower(t, g, raider); p != 1 {
		t.Errorf("with no artifact the Raider has power %d, want 1", p)
	}
	if hasEffectiveKeyword(t, g, raider, "haste") {
		t.Error("haste with no artifact")
	}
	rock := pushArtifactForTest(g, me.ID, "Mind Stone")
	if p := effectivePower(t, g, raider); p != 2 {
		t.Errorf("with an artifact the Raider has power %d, want 2", p)
	}
	if !hasEffectiveKeyword(t, g, raider, "haste") {
		t.Fatal("no haste with an artifact")
	}
	// An opponent's artifact is not one you control.
	pushArtifactForTest(g, opp.ID, "Their Mind Stone")
	exileCreature(t, g, rock)
	if p := effectivePower(t, g, raider); p != 1 {
		t.Errorf("with only an opponent's artifact the Raider has power %d, want 1", p)
	}
	if hasEffectiveKeyword(t, g, raider, "haste") {
		t.Error("haste with only an opponent's artifact")
	}
}

func TestLciDGGoblinTombRaiderAttacksTheTurnItArrives(t *testing.T) {
	for _, withArtifact := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withArtifact {
			pushArtifactForTest(g, me.ID, "Mind Stone")
		}
		raider := pushCatalogPermanent(g, me.ID, "Goblin Tomb Raider", "Creature — Goblin Pirate", lciDGTombRaiderOracle, true)
		advanceTo(t, g, game.StepDeclareAttackers)
		err := g.DeclareAttacker(raider, opp.ID)
		if withArtifact && err != nil {
			t.Errorf("a summoning-sick Raider with an artifact could not attack: %v", err)
		}
		if !withArtifact && err == nil {
			t.Error("a summoning-sick Raider with no artifact attacked")
		}
	}
}
