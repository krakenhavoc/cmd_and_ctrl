package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b46_cards_test.go — Star Trek Commander, batch 46 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB46KirkOracle    = "254177f2-5a03-4a29-8b98-248f6a41a351"
	trekB46SpockOracle   = "4d338db9-e968-49ec-8959-141480b17b8f"
	trekB46UhuraOracle   = "5d5c4a17-fa04-4889-819e-408fc2ff112e"
	trekB46ChekovOracle  = "058a5b8b-08bc-4021-8a8b-8e149941e775"
	trekB46ScottyOracle  = "3bd553db-2c7d-42cd-8ded-ab2c955b8f4e"
	trekB46SuluOracle    = "9565feaf-252d-43c2-a28e-4291c9ff5872"
	trekB46WorfOracle    = "8153cd58-cb8b-461c-a2c9-aaa9e2ea1b31"
	trekB46QOracle       = "e4e96682-1130-49cc-b1e5-04a774024f7a"
	trekB46LaForgeOracle = "d5e2ca35-053e-4155-97e7-73f75ae87d52"
	trekB46DataOracle    = "25ecd09e-d8bf-4929-96e2-e5132ea4f718"
)

func TestTrekB46EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB46KirkOracle:    "Kirk, Man of Action",
		trekB46SpockOracle:   "Spock, Voice of Reason",
		trekB46UhuraOracle:   "Uhura, Xenolinguist",
		trekB46ChekovOracle:  "Chekov and McCoy",
		trekB46ScottyOracle:  "Scotty, Chief Engineer",
		trekB46SuluOracle:    "Sulu, Heroic Helmsman",
		trekB46WorfOracle:    "Worf, Still Not Merry",
		trekB46QOracle:       "Q, Your Illustrious Host",
		trekB46LaForgeOracle: "La Forge, Thoughtful as Ever",
		trekB46DataOracle:    "Data, Professor of Humanities",
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

// --- fixtures -------------------------------------------------------------------

// trekB46Creature puts a creature of the given type line on p's side,
// past its summoning sickness.
func trekB46Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return effectstest.B12Push(g, owner, name, typeLine, "", power, toughness)
}

// trekB46Faeries creates n of the batch's Faerie tokens for p and returns
// their IDs.
func trekB46Faeries(t *testing.T, g *game.Game, p uuid.UUID, n int) []uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(p, effects.TokenCard(effects.TrekB46FaerieToken), n); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] {
			out = append(out, c.InstanceID)
		}
	}
	if len(out) != n {
		t.Fatalf("created %d Faeries, want %d", len(out), n)
	}
	return out
}

// --- Kirk, Man of Action ----------------------------------------------------------

func TestTrekB46KirkPumpsTheNamedTypesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kirk := effectstest.B12Push(g, me.ID, "Kirk, Man of Action", "Legendary Creature — Human Officer", trekB46KirkOracle, 3, 3)
	pumped := map[string]uuid.UUID{
		"Officer":    trekB46Creature(g, me.ID, "Ensign", "Creature — Human Officer", 1, 1),
		"Artificer":  trekB46Creature(g, me.ID, "Tinker", "Creature — Human Artificer", 1, 1),
		"Doctor":     trekB46Creature(g, me.ID, "Medic", "Creature — Human Doctor", 1, 1),
		"Pilot":      trekB46Creature(g, me.ID, "Flyer", "Creature — Human Pilot", 1, 1),
		"Scientist":  trekB46Creature(g, me.ID, "Boffin", "Creature — Vulcan Scientist", 1, 1),
		"Spacecraft": trekB46Creature(g, me.ID, "Shuttle", "Artifact Creature — Spacecraft", 1, 1),
		"two types":  trekB46Creature(g, me.ID, "Pilot Doctor", "Creature — Human Pilot Doctor", 1, 1),
	}
	bear := trekB46Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := trekB46Creature(g, opp.ID, "Their Officer", "Creature — Human Officer", 1, 1)

	for what, id := range pumped {
		if c := effectstest.LayeredCard(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
			t.Errorf("%s: %d/%d, want 3/3 (+2/+2 once)", what, c.CurrentPower(), c.CurrentToughness())
		}
	}
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 2 {
		t.Errorf("a Bear got the bonus: power %d", c.CurrentPower())
	}
	if c := effectstest.LayeredCard(t, g, theirs); c.CurrentPower() != 1 {
		t.Errorf("an opponent's Officer got the bonus: power %d", c.CurrentPower())
	}
	if c := effectstest.LayeredCard(t, g, kirk); c.CurrentPower() != 3 {
		t.Errorf("Kirk, an Officer, pumped himself: power %d (the card says other)", c.CurrentPower())
	}
}

// --- Spock, Voice of Reason ---------------------------------------------------------

func TestTrekB46SpockDrawsForEachNontokenPermanentThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spock := effectstest.B12Push(g, me.ID, "Spock, Voice of Reason", "Legendary Creature — Vulcan Human Scientist", trekB46SpockOracle, 3, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")

	// Two nontoken permanents enter this turn, and a token.
	effectstest.CastFromHand(t, g, game.Card{Name: "First Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastFromHand(t, g, game.Card{Name: "Some Relic", TypeLine: "Artifact"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	trekB46Faeries(t, g, me.ID, 1)

	if err := g.ActivateCatalogAbility(me.ID, spock, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana to pay {2}")
	}
	effectstest.FloatForTest(g, me, "CC")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, spock, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (two nontoken permanents entered; the token and Spock, here from before, don't count)", got)
	}
	if !effectstest.LayeredCard(t, g, spock).Tapped {
		t.Error("the {T} did not tap Spock")
	}
}

// --- Uhura, Xenolinguist -------------------------------------------------------------

func TestTrekB46UhuraOffersADrawForANonHumanCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Uhura, Xenolinguist", "Legendary Creature — Human Officer", trekB46UhuraOracle, false)
	effectstest.SeedLibrary(me, "A", "B")
	before := me.Hand.Size()

	// An opponent's non-Human creature counts: "a creature", anyone's.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard(effects.TrekB46FaerieToken), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a Faerie entering should ask Uhura's controller to pay {1}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("paying drew %d, want 1", got)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Human"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a Human entering triggered Uhura")
	}
}

func TestTrekB46UhuraDecliningDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Uhura, Xenolinguist", "Legendary Creature — Human Officer", trekB46UhuraOracle, false)
	effectstest.SeedLibrary(me, "A")
	before := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("3/3 green Beast"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("declining drew %d", got)
	}
}

// --- Chekov and McCoy -----------------------------------------------------------------

func TestTrekB46ChekovAndMcCoyDrawsThatManyOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	duo := effectstest.PushCatalogPermanent(g, me.ID, "Chekov and McCoy", "Legendary Creature — Human Pilot Doctor", trekB46ChekovOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")

	effectstest.TsPut(t, g, me.ID, bear, game.CounterPlusOne, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("counters on another creature triggered Chekov and McCoy")
	}

	hand := me.Hand.Size()
	effectstest.TsPut(t, g, me.ID, duo, game.CounterPlusOne, 3)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("drew %d, want 3 (that many)", got)
	}

	hand = me.Hand.Size()
	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	effectstest.TsPut(t, g, me.ID, duo, game.CounterPlusOne, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("asked again after the turn's draw was used")
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d more this turn, want 0", got)
	}
}

func TestTrekB46ChekovAndMcCoyIgnoresAnOpponentsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	duo := effectstest.PushCatalogPermanent(g, me.ID, "Chekov and McCoy", "Legendary Creature — Human Pilot Doctor", trekB46ChekovOracle, false)
	effectstest.TsPut(t, g, opp.ID, duo, game.CounterPlusOne, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Error("an opponent putting the counters is not you putting them")
	}
}

// --- Scotty, Chief Engineer -------------------------------------------------------------

func TestTrekB46ScottyTapsAnArtifactToMakeACreatureUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	scotty := effectstest.PushCatalogPermanent(g, me.ID, "Scotty, Chief Engineer", "Legendary Creature — Human Artificer", trekB46ScottyOracle, false)
	bear := trekB46Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := trekB46Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	if err := g.ActivateCatalogAbility(me.ID, scotty, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("activated with no artifact to tap")
	}
	tapped := effectstest.B12Permanent(g, me.ID, "Tapped Relic", "Artifact")
	effectstest.TmtB4Tap(g, tapped)
	if err := g.ActivateCatalogAbility(me.ID, scotty, 0, game.ActivateAbilityParams{Strict: true, Targets: target, TapIDs: []uuid.UUID{tapped}}); err == nil {
		t.Fatal("paid with a tapped artifact")
	}
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, scotty, 0, game.ActivateAbilityParams{Strict: true, Targets: target, TapIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.LayeredCard(t, g, relic).Tapped {
		t.Error("the cost did not tap the artifact")
	}
	if effectstest.LayeredCard(t, g, scotty).Tapped {
		t.Error("Scotty tapped himself; the cost is an artifact")
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("a creature that can't be blocked was blocked")
	}
}

// --- Sulu, Heroic Helmsman ----------------------------------------------------------------

func TestTrekB46SuluCrewsAsThoughHisPowerWereTripled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sulu := effectstest.B12Push(g, me.ID, "Sulu, Heroic Helmsman", "Legendary Creature — Human Pilot", trekB46SuluOracle, 2, 2)
	if got := effectstest.CountedPowerOf(t, g, sulu, game.CostPowerCrew); got != 6 {
		t.Fatalf("CountedPower(crew) = %d, want 6", got)
	}
	if got := effectstest.CountedPowerOf(t, g, sulu, game.CostPowerSaddle); got != 2 {
		t.Errorf("CountedPower(saddle) = %d, want 2: Sulu doesn't saddle as though tripled", got)
	}
	if c := effectstest.LayeredCard(t, g, sulu); c.CurrentPower() != 2 {
		t.Errorf("Sulu's power is %d, want 2: only the payment counts it tripled", c.CurrentPower())
	}
	vehicle := effectstest.PushVehicleForTest(g, me.ID, "Cost Power Vehicle", effectstest.CostPowerVehicleOracle, 4, 4)
	if err := effectstest.CrewCostPowerVehicle(t, g, me.ID, vehicle, sulu); err != nil {
		t.Fatalf("a 2-power Sulu counting as 6 could not crew a crew-3 Vehicle alone: %v", err)
	}
}

func TestTrekB46SuluStationsSpacecraftAsThoughTripled(t *testing.T) {
	g, me, ship, _ := effectstest.StationBoard(t, "The Seriema", "Legendary Artifact — Spacecraft", effectstest.TheSeriemaOracle, 1)
	sulu := effectstest.B12Push(g, me.ID, "Sulu, Heroic Helmsman", "Legendary Creature — Human Pilot", trekB46SuluOracle, 2, 2)
	effectstest.Station(t, g, me, ship, sulu)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ship, game.CounterCharge); got != 6 {
		t.Fatalf("charge counters = %d, want 6 (Sulu's 2 power, tripled)", got)
	}
}

func TestTrekB46SuluStationsAPlanetAtHisPlainPower(t *testing.T) {
	g, me, land, _ := effectstest.StationBoard(t, "Adagia, Windswept Bastion", "Land — Planet", effectstest.AdagiaOracle, 1)
	sulu := effectstest.B12Push(g, me.ID, "Sulu, Heroic Helmsman", "Legendary Creature — Human Pilot", trekB46SuluOracle, 2, 2)
	effectstest.Station(t, g, me, land, sulu)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, land, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters = %d, want 2: Sulu stations Spacecraft tripled, not Planets", got)
	}
}

// CR 608.2h: Sulu killed in response stations for his last-known power,
// tripled.
func TestTrekB46SuluStationsTripledFromLastKnownInformation(t *testing.T) {
	g, me, ship, _ := effectstest.StationBoard(t, "The Seriema", "Legendary Artifact — Spacecraft", effectstest.TheSeriemaOracle, 1)
	sulu := effectstest.B12Push(g, me.ID, "Sulu, Heroic Helmsman", "Legendary Creature — Human Pilot", trekB46SuluOracle, 2, 2)
	effectstest.Station(t, g, me, ship, sulu)
	effectstest.PutCounters(t, g, sulu, game.CounterPlusOne, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(sulu); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ship, game.CounterCharge); got != 9 {
		t.Fatalf("charge counters = %d, want 9 (his last-known 3 power, tripled)", got)
	}
}

// --- Worf, Still Not Merry ------------------------------------------------------------------

func TestTrekB46WorfMakesAFaerieForEachAttackingNonFaerie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worf := effectstest.B12Push(g, me.ID, "Worf, Still Not Merry", "Legendary Creature — Klingon Warrior Officer", trekB46WorfOracle, 5, 5)
	bear := trekB46Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sprite := trekB46Creature(g, me.ID, "Sprite", "Creature — Faerie", 1, 1)
	idle := trekB46Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	_ = idle

	effectstest.DeclareAttack(t, g, opp.ID, worf, bear, sprite)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)

	var faeries []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Faerie" && effects.IsToken(c) {
			faeries = append(faeries, c)
		}
	}
	if len(faeries) != 2 {
		t.Fatalf("%d Faeries, want 2 (Worf and the Bear; the Faerie and the idle Bear don't count)", len(faeries))
	}
	for _, c := range faeries {
		lc := effectstest.LayeredCard(t, g, c.InstanceID)
		if !c.Tapped || c.AttackingTarget != opp.ID || !game.HasKeyword(&lc, "flying") || lc.CurrentPower() != 1 {
			t.Errorf("Faerie: tapped %v, attacking %s (want %s), flying %v, power %d",
				c.Tapped, c.AttackingTarget, opp.ID, game.HasKeyword(&lc, "flying"), lc.CurrentPower())
		}
	}
}

// --- Q, Your Illustrious Host ----------------------------------------------------------------

func TestTrekB46QTapsThreeFaeriesToCastTheNextNonlandFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	q := effectstest.PushCatalogPermanent(g, me.ID, "Q, Your Illustrious Host", "Legendary Creature — Q", trekB46QOracle, true)
	hit := uuid.New()
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: land, Name: "Exiled Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	})
	faeries := trekB46Faeries(t, g, me.ID, 2)
	bear := trekB46Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, q, 0, game.ActivateAbilityParams{Strict: true, TapIDs: append(append([]uuid.UUID{}, faeries...), bear)}); err == nil {
		t.Fatal("paid with a creature that is not a Faerie")
	}
	if err := g.ActivateCatalogAbility(me.ID, q, 0, game.ActivateAbilityParams{Strict: true, TapIDs: faeries}); err == nil {
		t.Fatal("paid with two Faeries")
	}
	faeries = append(faeries, trekB46Faeries(t, g, me.ID, 1)...)
	// Summoning sickness is no bar: tapping a Faerie to pay a cost is not
	// the {T} symbol (CR 302.6), and Q (sick here) has no {T} either.
	if err := g.ActivateCatalogAbility(me.ID, q, 0, game.ActivateAbilityParams{Strict: true, TapIDs: faeries}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range faeries {
		if !effectstest.LayeredCard(t, g, id).Tapped {
			t.Error("a Faerie paid without tapping")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, hit)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, hit, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the exiled card for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hit) {
		t.Error("the exiled card was not cast")
	}
	if !g.Exile.Contains(land) {
		t.Error("the land exiled on the way left exile")
	}
}

// --- La Forge, Thoughtful as Ever --------------------------------------------------------------

func TestTrekB46LaForgeCountsPermanentsYouPutACounterOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	laForge := effectstest.PushCatalogPermanent(g, me.ID, "La Forge, Thoughtful as Ever", "Legendary Creature — Human Artificer", trekB46LaForgeOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	byOpponent := effectstest.PushVanillaCreature(g, me.ID, "Their Gift", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.TsPut(t, g, me.ID, bear, game.CounterPlusOne, 2)
	effectstest.TsPut(t, g, me.ID, bear, "charge", 1)
	effectstest.TsPut(t, g, me.ID, relic, "charge", 1)
	effectstest.TsPut(t, g, opp.ID, byOpponent, game.CounterPlusOne, 1)
	effectstest.TsPut(t, g, me.ID, theirs, game.CounterPlusOne, 1)
	effectstest.PassPriorityAroundTable(t, g)

	if err := g.ActivateCatalogAbility(me.ID, laForge, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Faerie"); got != 2 {
		t.Errorf("%d Faeries, want 2 (the Bear once and the Relic; not the one an opponent put a counter on, nor theirs)", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, laForge, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated La Forge while tapped")
	}
}

func TestTrekB46LaForgeIsBarredBySummoningSickness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	laForge := effectstest.PushCatalogPermanent(g, me.ID, "La Forge, Thoughtful as Ever", "Legendary Creature — Human Artificer", trekB46LaForgeOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, laForge, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a {T} ability was activated by a creature that arrived this turn")
	}
}

// --- Data, Professor of Humanities ---------------------------------------------------------------

func TestTrekB46DataMakesTwoFaeriesOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Data, Professor of Humanities", "Legendary Artifact Creature — Android Scientist", trekB46DataOracle, false)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	base := g.TurnTallyFor(me.ID).CardsDrawn

	for i := base; i < 1; i++ {
		effectstest.DrawFor(t, g, me, 1)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.B28TokensNamed(g, me.ID, "Faerie"); got != 0 {
		t.Fatalf("%d Faeries after the first draw", got)
	}
	effectstest.DrawFor(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Faerie"); got != 2 {
		t.Fatalf("%d Faeries after the second draw, want 2", got)
	}
	effectstest.DrawFor(t, g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Faerie"); got != 2 {
		t.Errorf("%d Faeries after the third draw, want still 2", got)
	}
}
