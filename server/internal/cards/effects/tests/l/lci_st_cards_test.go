package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_st_cards_test.go — The Lost Caverns of Ixalan, cards S–T (#4470).
// Each test plays the real catalog entry and asserts what its text says.

const (
	lciSTPhantomOracle     = "d32f570a-3738-4a9d-845f-6c7892b69991"
	lciSTSentinelOracle    = "a5bbfe32-2fb9-48a0-bc8a-e9899aac1fda"
	lciSTSirenOracle       = "4ec9d9eb-14e2-47ea-88a2-d2394e363784"
	lciSTSizeOracle        = "57bda813-ad4b-40ab-aa35-3f8343ff8668"
	lciSTSandwingOracle    = "5d840e49-3ca5-4e4f-9152-a365ad2b1773"
	lciSTMonstrosaurOracle = "3b66d2c2-be7a-4296-9888-f0cfb2975e89"
	lciSTEvangelistOracle  = "a70d9645-7c25-4fcb-9d16-fb73ce8b55e2"
	lciSTNecromageOracle   = "f5eace41-0f53-4084-8bec-1e079c04287b"
	lciSTInfantryOracle    = "f064a39e-ebb3-47b5-bd06-f39b02b95eaf"
	lciSTToteOracle        = "cd1e6225-31b7-4a70-81ab-681abe860328"
	lciSTSeekerOracle      = "cba9d6a2-1772-42dc-9bfb-c933a092be19"
	lciSTCrawlerOracle     = "a9a35c77-637f-4d56-afa2-6c8a4ded4838"
)

func TestLciSTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciSTPhantomOracle:     "Screaming Phantom",
		lciSTSentinelOracle:    "Sentinel of the Nameless City",
		lciSTSirenOracle:       "Spyglass Siren",
		lciSTSizeOracle:        "Staggering Size",
		lciSTSandwingOracle:    "Soaring Sandwing",
		lciSTMonstrosaurOracle: "Seismic Monstrosaur",
		lciSTEvangelistOracle:  "Sanguine Evangelist",
		lciSTNecromageOracle:   "Synapse Necromage",
		lciSTInfantryOracle:    "Thousand Moons Infantry",
		lciSTToteOracle:        "Tinker's Tote",
		lciSTSeekerOracle:      "Seeker of Sunlight",
		lciSTCrawlerOracle:     "Stinging Cave Crawler",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// lciSTActivateFromHand seeds a catalog card into the active seat's hand
// in its main phase, floats `mana`, and activates the hand ability at
// `index`, returning the card's id.
func lciSTActivateFromHand(t *testing.T, g *game.Game, name, typeLine, oracle, mana string, index int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, name, typeLine, oracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, index, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: activate hand ability %d: %v", name, index, err)
	}
	return id
}

// lciSTPermanentCardsInGraveyard puts n permanent cards into p's graveyard.
func lciSTPermanentCardsInGraveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushGraveyardCardTyped(p, "Dead Bear", "Creature — Bear")
	}
}

// --- Screaming Phantom -----------------------------------------------------------

func TestLciSTScreamingPhantomMillsACardWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phantom := effectstest.B12Push(g, me.ID, "Screaming Phantom", "Creature — Spirit", lciSTPhantomOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, phantom, "flying") {
		t.Error("Screaming Phantom has no flying")
	}
	top := effectstest.PutOnTopForTest(me, "Top Card", "Sorcery")
	library := me.Library.Size()
	effectstest.LciB1Attack(t, g, phantom, opp.ID)
	if !me.Graveyard.Contains(top) {
		t.Error("the top card of the library was not milled")
	}
	if got := library - me.Library.Size(); got != 1 {
		t.Errorf("the library lost %d cards, want 1", got)
	}
}

// --- Sentinel of the Nameless City -----------------------------------------------

func TestLciSTSentinelMakesAMapOnEntryAndOnEachAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentinel := effectstest.Msc17Play(t, g, me, game.Card{Name: "Sentinel of the Nameless City", TypeLine: "Creature — Merfolk Warrior Scout", OracleID: lciSTSentinelOracle, Power: 3, Toughness: 4})
	if got := lciB1MapsControlledBy(g, me.ID); got != 1 {
		t.Fatalf("%d Map tokens after it entered, want 1", got)
	}
	// It has been under its controller's control since the turn began
	// only on a later turn; the attack is declared on the next one.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.LciB1Attack(t, g, sentinel, opp.ID)
	if got := lciB1MapsControlledBy(g, me.ID); got != 2 {
		t.Errorf("%d Map tokens after it attacked, want 2", got)
	}
	if effectstest.TappedNow(g, sentinel) {
		t.Error("vigilance: attacking tapped the Sentinel")
	}
}

// --- Spyglass Siren --------------------------------------------------------------

func TestLciSTSpyglassSirenMakesAMapAndFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	siren := effectstest.Msc17Play(t, g, me, game.Card{Name: "Spyglass Siren", TypeLine: "Creature — Siren Pirate", OracleID: lciSTSirenOracle, Power: 1, Toughness: 1})
	if got := lciB1MapsControlledBy(g, me.ID); got != 1 {
		t.Errorf("%d Map tokens, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, siren, "flying") {
		t.Error("Spyglass Siren has no flying")
	}
}

// --- Staggering Size -------------------------------------------------------------

func TestLciSTStaggeringSizePumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	c := effectstest.SeedCreature(g, "Bear", g.Seats[0].ID)
	effectstest.CastCatalogSpell(t, g, "Staggering Size", "Instant", lciSTSizeOracle, effectstest.RfCardTarget(c))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, c), effectstest.EffectiveToughness(t, g, c); p != 5 || tough != 5 {
		t.Errorf("P/T = %d/%d, want 5/5 (a 2/2 gets +3/+3)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, c, "trample") {
		t.Error("the creature did not gain trample")
	}
}

// --- Soaring Sandwing ------------------------------------------------------------

func TestLciSTSoaringSandwingGainsThreeLifeOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	wing := effectstest.Msc17Play(t, g, me, game.Card{Name: "Soaring Sandwing", TypeLine: "Creature — Dinosaur", OracleID: lciSTSandwingOracle, Power: 3, Toughness: 5})
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wing, "flying") {
		t.Error("Soaring Sandwing has no flying")
	}
}

func TestLciSTSoaringSandwingPlainscyclesForAPlainsCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Sacred Foundry", "Land — Mountain Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id := lciSTActivateFromHand(t, g, "Soaring Sandwing", "Creature — Dinosaur", lciSTSandwingOracle, "{C}{C}", 0)
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Sandwing is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after plainscycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two Plains-typed lands", len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the fetched land did not reach the hand")
	}
}

// --- Seismic Monstrosaur ---------------------------------------------------------

func TestLciSTSeismicMonstrosaurSacrificesALandToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	m := effectstest.B12Push(g, me.ID, "Seismic Monstrosaur", "Creature — Dinosaur", lciSTMonstrosaurOracle, 6, 5)
	land := effectstest.B12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, m, "trample") {
		t.Error("Seismic Monstrosaur has no trample")
	}
	effectstest.FillPoolColored(me, "R", 3)
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("an artifact paid \"sacrifice a land\"")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate with a land: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestLciSTSeismicMonstrosaurMountaincyclesForAMountainCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Mountain", "Basic Land — Mountain"),
		effectstest.SearchTestLand("Stomping Ground", "Land — Mountain Forest"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	// The draw is the first printed ability; mountaincycling the second.
	lciSTActivateFromHand(t, g, "Seismic Monstrosaur", "Creature — Dinosaur", lciSTMonstrosaurOracle, "{C}{C}", 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after mountaincycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two Mountain-typed lands", len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the fetched Mountain did not reach the hand")
	}
}

// --- Sanguine Evangelist ---------------------------------------------------------

func TestLciSTSanguineEvangelistMakesABatOnEntryAndOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ev := effectstest.Msc17Play(t, g, me, game.Card{Name: "Sanguine Evangelist", TypeLine: "Creature — Vampire Cleric", OracleID: lciSTEvangelistOracle, Power: 2, Toughness: 1})
	if got := effectstest.B30TokensNamed(g, me.ID, "Bat"); got != 1 {
		t.Fatalf("%d Bats after it entered, want 1", got)
	}
	effectstest.B27Kill(g, ev)
	effectstest.PassPriorityAroundTable(t, g)
	bats := effectstest.TrekB33Tokens(g, me.ID, "Bat")
	if len(bats) != 2 {
		t.Fatalf("%d Bats after it died, want 2", len(bats))
	}
	c := effectstest.LayeredCard(t, g, bats[1].InstanceID)
	if !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" || !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
		t.Errorf("token = %s %d/%d %v %v, want a 1/1 black Bat with flying", c.TypeLine, c.Power, c.Toughness, c.Colors, c.Keywords)
	}
}

func TestLciSTSanguineEvangelistHasBattleCry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ev := effectstest.B12Push(g, me.ID, "Sanguine Evangelist", "Creature — Vampire Cleric", lciSTEvangelistOracle, 2, 1)
	other := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "My Idle Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{ev, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, other); got != 3 {
		t.Errorf("the other attacker's power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, ev); got != 2 {
		t.Errorf("the Evangelist's own power %d, want 2 (battle cry pumps each OTHER attacker)", got)
	}
	if got := effectstest.EffectivePower(t, g, idle); got != 2 {
		t.Errorf("a creature that did not attack has power %d, want 2", got)
	}
}

// --- Synapse Necromage -----------------------------------------------------------

func TestLciSTSynapseNecromageDiesIntoTwoFungiThatCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Synapse Necromage", "Creature — Fungus Wizard", lciSTNecromageOracle, 3, 1)
	effectstest.B27Kill(g, mage)
	effectstest.PassPriorityAroundTable(t, g)
	fungi := effectstest.TrekB33Tokens(g, me.ID, "Fungus")
	if len(fungi) != 2 {
		t.Fatalf("%d Fungus tokens, want 2", len(fungi))
	}
	for _, f := range fungi {
		c := effectstest.LayeredCard(t, g, f.InstanceID)
		if !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "B" || !c.HasSubtype("Fungus") {
			t.Errorf("token = %s %d/%d %v, want a 1/1 black Fungus", c.TypeLine, c.Power, c.Toughness, c.Colors)
		}
		if !game.Restricted(&c, game.CantBlock) {
			t.Error("the Fungus token can block")
		}
	}
}

// --- Thousand Moons Infantry -----------------------------------------------------

func TestLciSTThousandMoonsInfantryUntapsOnEachOtherPlayersUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	inf := effectstest.B12Push(g, me.ID, "Thousand Moons Infantry", "Creature — Human Soldier", lciSTInfantryOracle, 2, 4)
	other := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, inf, other)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.B16Tapped(t, g, inf) {
		t.Error("the Infantry did not untap on another player's untap step")
	}
	if !effectstest.B16Tapped(t, g, other) {
		t.Error("\"this creature\" is only itself: another creature untapped")
	}
}

// --- Tinker's Tote ---------------------------------------------------------------

func TestLciSTTinkersToteMakesTwoGnomesThenSacrificesForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tote := effectstest.Msc17Play(t, g, me, game.Card{Name: "Tinker's Tote", TypeLine: "Artifact", OracleID: lciSTToteOracle})
	gnomes := effectstest.TrekB33Tokens(g, me.ID, "Gnome")
	if len(gnomes) != 2 {
		t.Fatalf("%d Gnome tokens, want 2", len(gnomes))
	}
	if c := effectstest.LayeredCard(t, g, gnomes[0].InstanceID); !c.IsCreature() || !c.IsArtifact() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 {
		t.Errorf("token = %s %d/%d %v, want a 1/1 colorless Gnome artifact creature", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
	life := me.Life
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, tote, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(tote) {
		t.Error("the Tote was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Seeker of Sunlight ----------------------------------------------------------

func TestLciSTSeekerOfSunlightExploresAtSorcerySpeed(t *testing.T) {
	t.Run("a nonland gives a counter and may be binned", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seeker := effectstest.B12Push(g, me.ID, "Seeker of Sunlight", "Creature — Merfolk Scout", lciSTSeekerOracle, 1, 1)
		effectstest.AdvanceToMain(t, g)
		top := effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
		effectstest.FillPoolColored(me, "G", 2)
		if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Fatal("two mana paid {2}{G}")
		}
		effectstest.FillPoolColored(me, "G", 1)
		if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountersOn(g, seeker, game.CounterPlusOne); got != 1 {
			t.Errorf("the Seeker has %d +1/+1 counters, want 1", got)
		}
		effectstest.AnswerExploreQuestion(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(top) {
			t.Error("the revealed card was not put into the graveyard")
		}
	})
	t.Run("a land goes to the hand", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seeker := effectstest.B12Push(g, me.ID, "Seeker of Sunlight", "Creature — Merfolk Scout", lciSTSeekerOracle, 1, 1)
		effectstest.AdvanceToMain(t, g)
		top := effectstest.PutOnTopForTest(me, "Forest", "Basic Land — Forest")
		effectstest.FillPoolColored(me, "G", 3)
		if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(top) {
			t.Error("the revealed land did not go to the hand")
		}
		if got := effectstest.CountersOn(g, seeker, game.CounterPlusOne); got != 0 {
			t.Errorf("the Seeker has %d +1/+1 counters after revealing a land, want 0", got)
		}
	})
	t.Run("not outside a main phase", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seeker := effectstest.B12Push(g, me.ID, "Seeker of Sunlight", "Creature — Merfolk Scout", lciSTSeekerOracle, 1, 1)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.FillPoolColored(me, "G", 3)
		if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Error("activated at the beginning of combat; it is sorcery speed")
		}
	})
}

// --- Stinging Cave Crawler -------------------------------------------------------

func TestLciSTStingingCaveCrawlerDescendFourDrawsAndLosesLife(t *testing.T) {
	setup := func(t *testing.T, permanents int) (g *game.Game, me *game.Player, hand, life int) {
		g = effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		crawler := effectstest.B12Push(g, me.ID, "Stinging Cave Crawler", "Creature — Insect Horror", lciSTCrawlerOracle, 1, 3)
		lciSTPermanentCardsInGraveyard(me, permanents)
		// A nonpermanent card is not counted (CR 110.4).
		effectstest.PushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
		effectstest.AdvanceToMain(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, crawler, "deathtouch") {
			t.Error("Stinging Cave Crawler has no deathtouch")
		}
		hand, life = me.Hand.Size(), me.Life
		effectstest.LciB1Attack(t, g, crawler, opp.ID)
		return g, me, hand, life
	}
	t.Run("four permanent cards", func(t *testing.T) {
		_, me, hand, life := setup(t, 4)
		if got := me.Hand.Size() - hand; got != 1 {
			t.Errorf("drew %d cards, want 1", got)
		}
		if me.Life != life-1 {
			t.Errorf("life %d, want %d", me.Life, life-1)
		}
	})
	t.Run("three permanent cards", func(t *testing.T) {
		_, me, hand, life := setup(t, 3)
		if got := me.Hand.Size() - hand; got != 0 {
			t.Errorf("drew %d cards with three permanent cards in the graveyard, want 0", got)
		}
		if me.Life != life {
			t.Errorf("life %d, want %d", me.Life, life)
		}
	})
}
