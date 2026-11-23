package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b201_test.go — cards missing from sets since AFR, batch 201
// (#5948). Each test plays the real catalog entry and asserts what its text
// says.

const (
	gapB201Soothsayer = "5fe127e0-ee38-4f80-969b-13e84712a26f"
	gapB201Ghave      = "e579a72f-4933-40fe-9e57-96f8d65370bc"
	gapB201Teneb      = "d41f0357-f1d6-45cc-b8ea-4efa41da1448"
	gapB201Ashenmoor  = "f779754a-3de9-4621-b190-cf32949da0e8"
	gapB201Fireblade  = "52630eaf-995e-4d08-be99-888a38b598b2"
	gapB201Marchesa   = "17a59d3d-9e01-48cd-bb4a-3eaaa077751c"
	gapB201Sedraxis   = "eccc1230-df5a-44a0-859b-9e1e0d176dff"
	gapB201Thrax      = "9e0e4217-fefe-48dd-9153-032460192b19"
	gapB201Tariel     = "9c00b503-ae5e-4ce3-ad6a-92b541b2e3cd"
)

func TestGapB201EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB201Soothsayer: "Sultai Soothsayer",
		gapB201Ghave:      "Ghave, Guru of Spores",
		gapB201Teneb:      "Teneb, the Harvester",
		gapB201Ashenmoor:  "Ashenmoor Liege",
		gapB201Fireblade:  "Fireblade Artist",
		gapB201Marchesa:   "Marchesa, the Black Rose",
		gapB201Sedraxis:   "Sedraxis Specter",
		gapB201Thrax:      "Thraximundar",
		gapB201Tariel:     "Tariel, Reckoner of Souls",
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

// --- Sultai Soothsayer -------------------------------------------------------

func TestGapB201SultaiSoothsayerTakesOneOfFourAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var top []uuid.UUID
	for _, name := range []string{"D", "C", "B", "A"} {
		top = append(top, effectstest.TopLibraryCard(me, name, "Sorcery"))
	}
	lib := me.Library.Size()
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Sultai Soothsayer", TypeLine: "Creature — Snake Shaman", OracleID: gapB201Soothsayer, Power: 2, Toughness: 5})
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top four cards")
	}
	if len(pick.ChooseCards) != 4 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top four", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Error("the chosen card is not in hand")
	}
	for _, id := range []uuid.UUID{top[0], top[2], top[3]} {
		if !me.Graveyard.Contains(id) {
			t.Error("one of the other three is not in the graveyard")
		}
	}
	if me.Library.Size() != lib-4 {
		t.Errorf("library %d -> %d, want exactly the top four gone", lib, me.Library.Size())
	}
}

// --- Sedraxis Specter --------------------------------------------------------

func TestGapB201SedraxisSpecterFliesDiscardsAndUnearths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	specter := effectstest.B12Push(g, me.ID, "Sedraxis Specter", "Creature — Specter", gapB201Sedraxis, 3, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, specter, "flying") {
		t.Error("Sedraxis Specter does not have flying")
	}
	keep := effectstest.PushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := effectstest.PushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	effectstest.DealNoncombatDamage(g, me.ID, specter, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		t.Fatal("noncombat damage asked for a discard")
	}
	effectstest.DealCombatDamage(g, me.ID, specter, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}

	// Unearth {1}{B}: from the graveyard, sorcery speed, with haste.
	effectstest.AdvanceToMain(t, g)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Sedraxis Specter", OracleID: gapB201Sedraxis,
		TypeLine: "Creature — Specter", ManaCost: "{U}{B}{R}", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("unearth with one mana")
	}
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, dead, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth {1}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, dead) {
		t.Fatal("the unearthed Specter is not on the battlefield")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, dead, "haste") {
		t.Error("the unearthed Specter lacks haste")
	}
}

// --- Ashenmoor Liege ---------------------------------------------------------

func TestGapB201AshenmoorLiegeGivesBlackAndRedCreaturesPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := effectstest.B12Push(g, me.ID, "Ashenmoor Liege", "Creature — Elemental Knight", gapB201Ashenmoor, 4, 1)
	mk := func(owner uuid.UUID, colors ...string) uuid.UUID {
		id := effectstest.B12Creature(g, owner, "Bear", "Creature — Bear", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Colors = colors
				}
			}
			g.BumpLayerVersionForTest()
		})
		return id
	}
	b, r, both, green, theirs := mk(me.ID, "B"), mk(me.ID, "R"), mk(me.ID, "B", "R"), mk(me.ID, "G"), mk(opp.ID, "B")
	for _, want := range []struct {
		id  uuid.UUID
		who string
		p   int
		t   int
	}{
		{liege, "the Liege", 4, 1}, {b, "black", 3, 3}, {r, "red", 3, 3}, {both, "black and red", 4, 4},
		{green, "green", 2, 2}, {theirs, "an opponent's black creature", 2, 2},
	} {
		if p, tough := effectstest.PtOf(t, g, want.id); p != want.p || tough != want.t {
			t.Errorf("%s is %d/%d, want %d/%d", want.who, p, tough, want.p, want.t)
		}
	}
}

func TestGapB201AshenmoorLiegeMakesAnOpponentLoseFourWhenTheyTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := effectstest.B12Push(g, me.ID, "Ashenmoor Liege", "Creature — Elemental Knight", gapB201Ashenmoor, 4, 1)
	life := opp.Life
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: liege})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("the opponent's life %d -> %d, want -4 for targeting the Liege", life, opp.Life)
	}
	// Its own controller targeting it loses nothing.
	mine := me.Life
	liege = effectstest.B12Push(g, me.ID, "Ashenmoor Liege", "Creature — Elemental Knight", gapB201Ashenmoor, 4, 1)
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: liege})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("the controller's life %d -> %d, want no change", mine, me.Life)
	}
}

// --- Ghave, Guru of Spores ---------------------------------------------------

func TestGapB201GhaveEntersWithFiveCountersAndMakesSaprolingsAndCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ghave := effectstest.Msc17Play(t, g, me, game.Card{Name: "Ghave, Guru of Spores", TypeLine: "Legendary Creature — Fungus Shaman", OracleID: gapB201Ghave})
	if got := effectstest.CounterCount(g, ghave, game.CounterPlusOne); got != 5 {
		t.Fatalf("Ghave entered with %d +1/+1 counters, want 5", got)
	}
	if p, tough := effectstest.PtOf(t, g, ghave); p != 5 || tough != 5 {
		t.Errorf("Ghave is %d/%d, want 5/5", p, tough)
	}

	// {1}, remove a counter from a creature you control: a Saproling.
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, ghave, 0, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{ghave}}); err != nil {
		t.Fatalf("remove-a-counter ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ghave, game.CounterPlusOne); got != 4 {
		t.Errorf("Ghave has %d counters after paying one, want 4", got)
	}
	saprolings := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Saproling" && c.Controller == me.ID {
			saprolings++
		}
	}
	if saprolings != 1 {
		t.Fatalf("%d Saprolings, want 1", saprolings)
	}

	// {1}, sacrifice a creature: a +1/+1 counter on target creature.
	var sap uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Saproling" {
			sap = c.InstanceID
		}
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, ghave, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sap},
		Targets: effectstest.CardRefs(ghave)}); err != nil {
		t.Fatalf("sacrifice ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, sap) {
		t.Error("the Saproling was not sacrificed")
	}
	if got := effectstest.CounterCount(g, ghave, game.CounterPlusOne); got != 5 {
		t.Errorf("Ghave has %d counters, want 5 after the counter was put on it", got)
	}
	// Strict: an empty pool refuses.
	if err := g.ActivateCatalogAbility(me.ID, ghave, 0, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{ghave}}); err == nil {
		t.Error("the first ability was activated with no mana")
	}
}

// --- Teneb, the Harvester ----------------------------------------------------

func TestGapB201TenebReanimatesFromAnyGraveyardOnlyIfYouPay(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		teneb := effectstest.B12Push(g, me.ID, "Teneb, the Harvester", "Legendary Creature — Dragon", gapB201Teneb, 6, 6)
		if !effectstest.HasKeywordOnBattlefield(t, g, teneb, "flying") {
			t.Error("Teneb does not have flying")
		}
		dead := uuid.New()
		opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		effectstest.DealCombatDamage(g, me.ID, teneb, opp.ID)
		effectstest.PickIfAsked(t, g, me.ID, dead)
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.FloatForTest(g, me, "BCC")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		onMine := false
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == dead && c.Controller == me.ID {
				onMine = true
			}
		}
		if onMine != pay {
			t.Errorf("pay=%v: the card is on my battlefield=%v", pay, onMine)
		}
		if !pay && !opp.Graveyard.Contains(dead) {
			t.Error("the card left the graveyard without payment")
		}
	}
}

// --- Fireblade Artist --------------------------------------------------------

func TestGapB201FirebladeArtistSacrificesAtUpkeepForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	artist := effectstest.B12Push(g, me.ID, "Fireblade Artist", "Creature — Human Shaman", gapB201Fireblade, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, artist, "haste") {
		t.Error("Fireblade Artist does not have haste")
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	life := opp.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, fodder) {
		t.Error("the Fodder was not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("the opponent's life %d -> %d, want -2", life, opp.Life)
	}
}

// --- Marchesa, the Black Rose ------------------------------------------------

func TestGapB201MarchesaGrantsDethroneAndReturnsCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marchesa := effectstest.B12Push(g, me.ID, "Marchesa, the Black Rose", "Legendary Creature — Human Wizard", gapB201Marchesa, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	// Dethrone on another creature: attack the (tied) life leader.
	effectstest.LciB1Attack(t, g, bear, opp.ID)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("the granted dethrone put %d counters on the attacking Bear, want 1", got)
	}
	if got := effectstest.CounterCount(g, marchesa, game.CounterPlusOne); got != 0 {
		t.Errorf("Marchesa has %d counters before attacking, want 0", got)
	}

	// A creature with a +1/+1 counter dies: it returns at the next end step.
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, bear) {
		t.Fatal("the Bear came back before the end step")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, bear) {
		t.Error("the Bear did not return at the beginning of the next end step")
	}
}

func TestGapB201MarchesaDoesNotReturnACreatureWithoutACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Marchesa, the Black Rose", "Legendary Creature — Human Wizard", gapB201Marchesa, 3, 3)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, plain) {
		t.Error("a creature with no +1/+1 counter returned")
	}
}

// --- Thraximundar ------------------------------------------------------------

func TestGapB201ThraximundarMakesTheDefenderSacrificeAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thrax := effectstest.B12Push(g, me.ID, "Thraximundar", "Legendary Creature — Zombie Assassin", gapB201Thrax, 6, 6)
	if !effectstest.HasKeywordOnBattlefield(t, g, thrax, "haste") {
		t.Error("Thraximundar does not have haste")
	}
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(thrax, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, victim) {
		t.Error("the defending player's only creature was not sacrificed")
	}
	prompt := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("answer Thraximundar: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, thrax, game.CounterPlusOne); got != 1 {
		t.Errorf("Thraximundar has %d counters after the sacrifice, want 1", got)
	}
}

// --- Tariel, Reckoner of Souls -----------------------------------------------

func TestGapB201TarielReanimatesACreatureFromAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tariel := effectstest.B12Push(g, me.ID, "Tariel, Reckoner of Souls", "Legendary Creature — Angel", gapB201Tariel, 4, 7)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, tariel, kw) {
			t.Errorf("Tariel does not have %s", kw)
		}
	}
	effectstest.AdvanceToMain(t, g)
	// An empty graveyard gives nothing (the ability still resolves).
	if err := effectstest.HobB4Activate(t, g, me.ID, tariel, 0, "", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	only := uuid.New()
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: only, Name: "Only Creature", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "A Sorcery", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.UntapForTest(g, tariel)
	if err := effectstest.HobB4Activate(t, g, me.ID, tariel, 0, "", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	onMine := false
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == only && c.Controller == me.ID {
			onMine = true
		}
	}
	if !onMine {
		t.Error("the only creature card did not enter under my control")
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("the sorcery left the graveyard")
	}
	// Tapped now: a second activation is refused.
	if err := effectstest.HobB4Activate(t, g, me.ID, tariel, 0, "", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("activated while tapped")
	}
}
