package testscm

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b9_cards_test.go — Commander deck cards, batches 9 to 12 (#5368).

const (
	cmdrB9SecludedSteppeOracle   = "8dec6fcf-1254-4b1b-ba23-7a3e492a7241"
	cmdrB9TranquilThicketOracle  = "9f8fe514-77ed-41b4-a6f3-c6f095bb97be"
	cmdrB9LonelySandbarOracle    = "765863c8-1be0-4bb1-9e9c-db7701cffde3"
	cmdrB9DakmorSalvageOracle    = "cdc4048a-73ec-4ec1-a179-2b36c397bf1a"
	cmdrB9RazortideBridgeOracle  = "6cb37ac1-dd11-4a8c-bca5-ef44d828059f"
	cmdrB9RustvaleBridgeOracle   = "a3faf70d-c034-4692-9e92-1922029e3852"
	cmdrB9SilverbluffBridgeOracl = "081bfd50-a436-463b-9d2c-5bc8a32b387c"
)

// cmdrB9Registered pins that every oracle id is a Full card with no caveats.
func cmdrB9Registered(t *testing.T, want map[string]string) {
	t.Helper()
	for oracle, name := range want {
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

func TestCmdrB9LandsAreRegisteredFull(t *testing.T) {
	cmdrB9Registered(t, map[string]string{
		cmdrB9SecludedSteppeOracle: "Secluded Steppe", cmdrB9TranquilThicketOracle: "Tranquil Thicket",
		cmdrB9LonelySandbarOracle: "Lonely Sandbar", cmdrB9DakmorSalvageOracle: "Dakmor Salvage",
		cmdrB9RazortideBridgeOracle: "Razortide Bridge", cmdrB9RustvaleBridgeOracle: "Rustvale Bridge",
		cmdrB9SilverbluffBridgeOracl: "Silverbluff Bridge",
	})
}

func TestCmdrB9CyclingLandsEnterTappedAndCycleForTheirColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Secluded Steppe", cmdrB9SecludedSteppeOracle, "{W}"},
		{"Tranquil Thicket", cmdrB9TranquilThicketOracle, "{G}"},
		{"Lonely Sandbar", cmdrB9LonelySandbarOracle, "{U}"},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
		if !effectstest.TappedForTest(t, g, id) {
			t.Errorf("%s did not enter tapped", tc.name)
		}
		g2 := effectstest.NewCatalogGame(t)
		me := g2.Seats[g2.Turn.ActiveSeat]
		before := me.Hand.Size()
		cid, _ := effectstest.CycleFromHand(t, g2, tc.name, "Land", tc.oracle, tc.mana)
		if me.Hand.Contains(cid) {
			t.Errorf("%s is still in hand after cycling", tc.name)
		}
		if me.Hand.Size() != before {
			t.Errorf("%s: hand %d after cycling, want %d (discard one, draw one)", tc.name, me.Hand.Size(), before)
		}
	}
}

func TestCmdrB9CyclingNeedsItsColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Secluded Steppe", "Land", cmdrB9SecludedSteppeOracle)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("cycling {W} was paid with colourless mana")
	}
}

func TestCmdrB9BridgesEnterTappedIndestructibleAndTapForEitherColour(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		colours      []string
	}{
		{"Razortide Bridge", cmdrB9RazortideBridgeOracle, []string{"{W}", "{U}"}},
		{"Rustvale Bridge", cmdrB9RustvaleBridgeOracle, []string{"{R}", "{W}"}},
		{"Silverbluff Bridge", cmdrB9SilverbluffBridgeOracl, []string{"{U}", "{R}"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12PlayFromHand(t, g, tc.name, "Artifact Land", tc.oracle, game.CastSpellParams{})
		if !effectstest.TappedForTest(t, g, id) {
			t.Errorf("%s did not enter tapped", tc.name)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("%s is not indestructible", tc.name)
		}
		spec, _ := effects.Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 {
			t.Fatalf("%s: %d mana abilities", tc.name, len(spec.ManaAbilities))
		}
		produced := spec.ManaAbilities[0].Produced
		for _, col := range tc.colours {
			if !strings.Contains(produced, col[1:2]) {
				t.Errorf("%s produces %q, missing %s", tc.name, produced, col)
			}
		}
	}
}

func TestCmdrB9DakmorSalvageDredgesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Dakmor Salvage", TypeLine: "Land",
		OracleID: cmdrB9DakmorSalvageOracle, Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	_ = g.DrawCard(me.ID)
	offer := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
	if offer == nil || offer.DredgeOffer() != 2 {
		t.Fatalf("no dredge 2 offer: %+v", offer)
	}
	if err := g.ResolveOptionalReplacement(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Error("Dakmor Salvage did not return to hand")
	}
	if got := me.Library.Size(); got != lib-2 {
		t.Errorf("library %d, want %d (milled two, drew none)", got, lib-2)
	}
}

// cmdrB9Oracle is the oracle id the catalog registered `name` under.
func cmdrB9Oracle(t *testing.T, name string) string {
	t.Helper()
	for _, s := range effects.All() {
		if s.Name == name {
			return s.OracleID
		}
	}
	t.Fatalf("%s is not registered", name)
	return ""
}

func TestCmdrB9Group1CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{}
	for _, n := range []string{"Baloth Prime", "Patrolling Peacemaker", "Kilo, Apogee Mind", "Surge Conductor",
		"Uurg, Spawn of Turg", "Enthusiastic Mechanaut", "Wake the Past", "Coretapper", "Golem Foundry",
		"Titan Forge", "Elsha, Threefold Master", "Crystalline Crawler", "Mindless Automaton", "Groundskeeper",
		"Experimental Augury"} {
		want[cmdrB9Oracle(t, n)] = n
	}
	cmdrB9Registered(t, want)
}

func cmdrB9Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, mana string, power, toughness int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: power, Toughness: toughness,
			OracleID: cmdrB9Oracle(t, name), Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestCmdrB9BalothPrimeEntersStunnedAndFeedsOnSacrificedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	prime := cmdrB9Cast(t, g, me, "Baloth Prime", "Creature — Beast", "{3}{G}", "GGGG", 10, 10)
	c, ok := effectstest.BattlefieldCard(g, prime)
	if !ok {
		t.Fatalf("Baloth Prime did not resolve: stack %d, pending %d, hand has it %v", g.Stack.Size(), len(g.PendingChoices), me.Hand.Contains(prime))
	}
	if !c.Tapped || c.Counters[game.CounterStun] != 6 {
		t.Fatalf("tapped %v, stun %d; want tapped with six stun counters", c.Tapped, c.Counters[game.CounterStun])
	}
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.FloatForTest(g, me, "CCCC")
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, prime, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{forest},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2", before, me.Life)
	}
	var beast *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Beast" && g.Battlefield.Cards[i].Controller == me.ID {
			beast = &g.Battlefield.Cards[i]
		}
	}
	if beast == nil {
		t.Fatal("no Beast token for the sacrificed land")
	}
	if !beast.Tapped {
		t.Error("the Beast token did not enter tapped")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, beast.InstanceID); p != 4 || tough != 4 {
		t.Errorf("Beast is %d/%d, want 4/4", p, tough)
	}
	c, _ = effectstest.BattlefieldCard(g, prime)
	if c.Counters[game.CounterStun] != 5 || !c.Tapped {
		t.Errorf("untap with a stun counter: stun %d tapped %v, want 5 and still tapped", c.Counters[game.CounterStun], c.Tapped)
	}
}

func TestCmdrB9PatrollingPeacemakerEntersWithTwoCountersAndProliferatesOnAnOpponentsCrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	peace := cmdrB9Cast(t, g, me, "Patrolling Peacemaker", "Artifact Creature — Robot Soldier", "{2}{W}", "WCC", 0, 0)
	if got := effectstest.CounterCount(g, peace, game.CounterPlusOne); got != 2 {
		t.Fatalf("entered with %d counters, want 2", got)
	}
	emit := func(actor uuid.UUID, crime bool) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: actor, CrimeCommitted: crime})
		})
		effectstest.SettleAnsweringProliferate(t, g)
	}
	emit(me.ID, true)
	if got := effectstest.CounterCount(g, peace, game.CounterPlusOne); got != 2 {
		t.Errorf("its controller's own crime proliferated: %d counters", got)
	}
	emit(opp.ID, false)
	if got := effectstest.CounterCount(g, peace, game.CounterPlusOne); got != 2 {
		t.Errorf("a non-crime targeting proliferated: %d counters", got)
	}
	emit(opp.ID, true)
	if got := effectstest.CounterCount(g, peace, game.CounterPlusOne); got != 3 {
		t.Errorf("an opponent's crime: %d counters, want 3", got)
	}
}

func TestCmdrB9KiloProliferatesWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	kilo := effectstest.B12Push(g, me.ID, "Kilo, Apogee Mind", "Legendary Artifact Creature — Robot Artificer", cmdrB9Oracle(t, "Kilo, Apogee Mind"), 3, 3)
	ally := effectstest.PushCounterCreature(g, me.ID, "Ally", game.CounterPlusOne, 1)
	effectstest.Untap(g, kilo)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == kilo {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		g.EmitEvent(game.Event{Kind: game.EventTapCard, CardID: kilo, Actor: me.ID})
	})
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("Kilo's tap proliferated %d counters on the ally, want 2", got)
	}
}

func TestCmdrB9SurgeConductorProliferatesForEachOtherNontokenArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Surge Conductor", "Artifact Creature — Robot", cmdrB9Oracle(t, "Surge Conductor"), 3, 2)
	ally := effectstest.PushCounterCreature(g, me.ID, "Ally", game.CounterPlusOne, 1)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.CastFromHandForTest(t, g, me, "Trinket", "Artifact", "{2}", "", game.CastSpellParams{Strict: true})
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("a nontoken artifact entering: %d counters, want 2", got)
	}
	// A token does not trigger it.
	g.WithWriteLock(func() {
		_, _ = g.CreateTokensForEffect(me.ID, effects.TreasureToken(), 1, game.TokenEntryOptions{})
	})
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("a token entering proliferated: %d counters, want 2", got)
	}
}

func TestCmdrB9ElshaMakesAMonkForEachPointOfCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	effectstest.AdvanceToMain(t, g)
	elsha := effectstest.B12Push(g, me.ID, "Elsha, Threefold Master", "Legendary Creature — Djinn Monk", cmdrB9Oracle(t, "Elsha, Threefold Master"), 1, 1)
	effectstest.DealCombatDamageToPlayer(g, elsha, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Monk", me.ID); got != 3 {
		t.Fatalf("%d Monk tokens, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Monk" && !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "prowess") {
			t.Error("a Monk token lacks prowess")
		}
	}
}

func TestCmdrB9UurgCountsLandsInYourGraveyardAndSurveilsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	uurg := effectstest.B12Push(g, me.ID, "Uurg, Spawn of Turg", "Legendary Creature — Frog Beast", cmdrB9Oracle(t, "Uurg, Spawn of Turg"), 0, 5)
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if p, tough := effectstest.PtOf(t, g, uurg); p != 2 || tough != 5 {
		t.Errorf("%d/%d with two land cards in the graveyard, want 2/5", p, tough)
	}
	// Sacrificing a land for {B}{G} gains 2 life.
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.FloatForTest(g, me, "BG")
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, uurg, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2", before, me.Life)
	}
	if p, _ := effectstest.PtOf(t, g, uurg); p != 3 {
		t.Errorf("power %d after the sacrificed land joined the graveyard, want 3", p)
	}
}

func TestCmdrB9WakeThePastReturnsEveryArtifactCardWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushGraveyardCardTyped(me, "Relic", "Artifact")
	b := effectstest.PushGraveyardCardTyped(me, "Golem", "Artifact Creature — Golem")
	c := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	cmdrB9Cast(t, g, me, "Wake the Past", "Sorcery", "{5}{R}{W}", "RWCCCCC", 0, 0)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("not every artifact card came back")
	}
	if g.Battlefield.Contains(c) || !me.Graveyard.Contains(c) {
		t.Error("a nonartifact card came back")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, a, "haste") || !effectstest.HasKeywordOnBattlefield(t, g, b, "haste") {
		t.Error("the returned artifacts did not gain haste")
	}
}

func TestCmdrB9MechanautMakesArtifactSpellsCostOneLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Enthusiastic Mechanaut", "Artifact Creature — Goblin Artificer", cmdrB9Oracle(t, "Enthusiastic Mechanaut"), 2, 2)
	effectstest.FloatForTest(g, me, "C")
	effectstest.CastFromHandForTest(t, g, me, "Trinket", "Artifact", "{2}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Trinket", me.ID); got != 1 {
		t.Errorf("a {2} artifact for one mana: %d on the battlefield, want 1", got)
	}
	effectstest.FloatForTest(g, me, "C")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a {2} creature was castable for one mana")
	}
}

func TestCmdrB9CoretapperPutsChargeCountersOnAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tapper := effectstest.B12Push(g, me.ID, "Coretapper", "Artifact Creature — Myr", cmdrB9Oracle(t, "Coretapper"), 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	if err := g.ActivateCatalogAbility(me.ID, tapper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a nonartifact target was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, tapper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err != nil {
		t.Fatalf("tap ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, rock, game.CounterCharge); got != 1 {
		t.Errorf("%d charge counters after the tap, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, tapper, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err != nil {
		t.Fatalf("sacrifice ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, rock, game.CounterCharge); got != 3 {
		t.Errorf("%d charge counters after the sacrifice, want 3", got)
	}
	if g.Battlefield.Contains(tapper) {
		t.Error("Coretapper was not sacrificed")
	}
}

func TestCmdrB9GolemFoundryChargesOnArtifactCastsAndMakesAGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	foundry := effectstest.B12Push(g, me.ID, "Golem Foundry", "Artifact", cmdrB9Oracle(t, "Golem Foundry"), 0, 0)
	for i := 0; i < 3; i++ {
		effectstest.FloatForTest(g, me, "C")
		effectstest.CastFromHandForTest(t, g, me, "Trinket", "Artifact", "{1}", "", game.CastSpellParams{Strict: true})
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, foundry, game.CounterCharge); got != 3 {
		t.Fatalf("%d charge counters after three artifact spells, want 3", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, foundry, game.CounterCharge); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
	if got := effectstest.CountOnBattlefield(g, "Golem", me.ID); got != 1 {
		t.Errorf("%d Golem tokens, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without three charge counters")
	}
}

func TestCmdrB9TitanForgeChargesThenMakesANineNineGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	forge := effectstest.B12Push(g, me.ID, "Titan Forge", "Artifact", cmdrB9Oracle(t, "Titan Forge"), 0, 0)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("charge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, forge, game.CounterCharge); got != 1 {
		t.Fatalf("%d counters, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(forge, game.CounterCharge, 2) })
	effectstest.Untap(g, forge)
	if err := g.ActivateCatalogAbility(me.ID, forge, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("make a Golem: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var golem uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Golem" && c.Controller == me.ID {
			golem = c.InstanceID
		}
	}
	if golem == uuid.Nil {
		t.Fatal("no Golem token")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, golem); p != 9 || tough != 9 {
		t.Errorf("Golem is %d/%d, want 9/9", p, tough)
	}
	if got := effectstest.CounterCount(g, forge, game.CounterCharge); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
}

func TestCmdrB9CrystallineCrawlerConvergesAndUsesItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crawler := cmdrB9Cast(t, g, me, "Crystalline Crawler", "Artifact Creature — Construct", "{4}", "WUBR", 1, 1)
	if got := effectstest.CounterCount(g, crawler, game.CounterPlusOne); got != 4 {
		t.Fatalf("%d counters for four colours of mana, want 4", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	crawler2 := cmdrB9Cast(t, g2, me2, "Crystalline Crawler", "Artifact Creature — Construct", "{4}", "CCCC", 1, 1)
	if got := effectstest.CounterCount(g2, crawler2, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters for colourless mana, want 0", got)
	}
	// Tap for a counter.
	effectstest.Untap(g2, crawler2)
	g2.WithWriteLock(func() {
		for i := range g2.Battlefield.Cards {
			if g2.Battlefield.Cards[i].InstanceID == crawler2 {
				g2.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g2.ActivateCatalogAbility(me2.ID, crawler2, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CounterCount(g2, crawler2, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after the tap, want 1", got)
	}
}

func TestCmdrB9MindlessAutomatonGrowsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	auto := cmdrB9Cast(t, g, me, "Mindless Automaton", "Artifact Creature — Construct", "{4}", "CCCC", 0, 0)
	if got := effectstest.CounterCount(g, auto, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters on entering, want 2", got)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, auto, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if _, ok := effectstest.BattlefieldCard(g, auto); ok {
		t.Error("a 0/0 with no counters stayed on the battlefield")
	}
}

func TestCmdrB9GroundskeeperReturnsOnlyBasicLandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	keeper := effectstest.B12Push(g, me.ID, "Groundskeeper", "Creature — Human Druid", cmdrB9Oracle(t, "Groundskeeper"), 1, 1)
	basic := effectstest.PushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	nonbasic := effectstest.PushGraveyardCardTyped(me, "Tarn", "Land")
	effectstest.FloatForTest(g, me, "CG")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: nonbasic}}}); err == nil {
		t.Fatal("a nonbasic land card was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: basic}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(basic) {
		t.Error("the basic land card did not return to hand")
	}
}

func TestCmdrB9ExperimentalAuguryTakesOneBottomsTwoAndProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ally := effectstest.PushCounterCreature(g, me.ID, "Ally", game.CounterPlusOne, 1)
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UC")
	effectstest.CastFromHandForTest(t, g, me, "Experimental Augury", "Instant", "{1}{U}", cmdrB9Oracle(t, "Experimental Augury"), game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 3 {
		t.Fatalf("want a pick of exactly one of three, got %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	if order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[0], ids[2]}, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	if effectstest.ProliferatePrompt(g) == nil {
		t.Fatal("no proliferate prompt after the take and the ordering")
	}
	effectstest.AnswerProliferate(t, g, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 1 || !me.Hand.Contains(ids[1]) {
		t.Errorf("hand grew by %d, want B alone", me.Hand.Size()-hand)
	}
	if got := effectstest.CounterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters after proliferate, want 2", got)
	}
}

func TestCmdrB9Group2CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{}
	for _, n := range []string{"Sprouting Goblin", "Eumidian Hatchery", "Scouring Swarm", "Gavel of the Righteous"} {
		want[cmdrB9Oracle(t, n)] = n
	}
	cmdrB9Registered(t, want)
}

func TestCmdrB9SproutingGoblinSearchesOnlyWhenKickedAndOnlyForABasicLandType(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		wastes := effectstest.FrPushTop(me, "Wastes", "Basic Land", "")
		dual := effectstest.FrPushTop(me, "Tarn", "Land — Island Mountain", "")
		effectstest.AdvanceToMain(t, g)
		mana, params := "RC", game.CastSpellParams{Strict: true}
		if kicked {
			mana, params.OptionalCosts = "RCG", []int{0}
		}
		effectstest.FloatForTest(g, me, mana)
		effectstest.CastFromHandForTest(t, g, me, "Sprouting Goblin", "Creature — Goblin Druid", "{1}{R}", cmdrB9Oracle(t, "Sprouting Goblin"), params)
		effectstest.PassPriorityAroundTable(t, g)
		search := effectstest.SearchChoiceFor(g, me.ID)
		if !kicked {
			if search != nil {
				t.Error("an unkicked Goblin searched")
			}
			continue
		}
		if search == nil {
			t.Fatal("a kicked Goblin did not search")
		}
		if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{wastes}); err == nil {
			t.Error("a Wastes (no basic land type) was accepted")
		}
		if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{dual}); err != nil {
			t.Fatalf("search: %v", err)
		}
		g.SettleResolution()
		if !me.Hand.Contains(dual) {
			t.Error("the dual land did not reach the hand")
		}
	}
}

func TestCmdrB9SproutingGoblinSacrificesALandToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	goblin := effectstest.B12Push(g, me.ID, "Sprouting Goblin", "Creature — Goblin Druid", cmdrB9Oracle(t, "Sprouting Goblin"), 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.FloatForTest(g, me, "R")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || g.Battlefield.Contains(land) {
		t.Errorf("hand %d → %d, land still on the battlefield %v", hand, me.Hand.Size(), g.Battlefield.Contains(land))
	}
}

func TestCmdrB9EumidianHatcheryCountsHatchlingsAndHatchesThemWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Push(g, me.ID, "Eumidian Hatchery", "Land", cmdrB9Oracle(t, "Eumidian Hatchery"), 0, 0)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	if got := effectstest.CounterCount(g, land, "hatchling"); got != 1 {
		t.Fatalf("%d hatchling counters, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(land, "hatchling", 2) })
	effectstest.KillCreature(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Insect", me.ID); got != 3 {
		t.Fatalf("%d Insect tokens, want 3 (one per hatchling counter)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Insect" && !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "flying") {
			t.Error("an Insect lacks flying")
		}
	}
}

func TestCmdrB9ScouringSwarmMakesAnInsectOrACopyDependingOnTheGraveyard(t *testing.T) {
	for _, lands := range []int{2, 7} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		swarm := effectstest.B12Push(g, me.ID, "Scouring Swarm", "Creature — Insect", cmdrB9Oracle(t, "Scouring Swarm"), 1, 1)
		for i := 0; i < lands; i++ {
			effectstest.PushToGraveyard(g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		}
		land := effectstest.SeedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventSacrifice, Actor: me.ID, CardID: land})
		})
		effectstest.PassPriorityAroundTable(t, g)
		var made *game.Card
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != swarm && c.Controller == me.ID && c.IsCreature() {
				made = c
			}
		}
		if made == nil {
			t.Fatalf("%d lands: no token", lands)
		}
		if !made.Tapped || !made.IsToken() {
			t.Errorf("%d lands: token tapped %v, token %v", lands, made.Tapped, made.IsToken())
		}
		isCopy := made.Name == "Scouring Swarm"
		if isCopy != (lands >= 7) {
			t.Errorf("%d lands: token is %q", lands, made.Name)
		}
		if !isCopy && made.Name != "Insect" {
			t.Errorf("%d lands: token is %q, want an Insect", lands, made.Name)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, made.InstanceID, "flying") {
			t.Errorf("%d lands: the token lacks flying", lands)
		}
	}
}

func TestCmdrB9GavelGrowsEachCombatAndEquipsForManaOrACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	gavel := effectstest.B12Push(g, me.ID, "Gavel of the Righteous", "Artifact — Equipment", cmdrB9Oracle(t, "Gavel of the Righteous"), 0, 0)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, gavel, game.CounterCharge); got != 1 {
		t.Fatalf("%d charge counters after the beginning of combat, want 1", got)
	}
	// Equip by removing the counter.
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	branch := 1
	if err := g.ActivateCatalogAbility(me.ID, gavel, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, CostBranch: &branch, CounterKind: game.CounterCharge}); err != nil {
		t.Fatalf("equip by removing a counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, gavel, game.CounterCharge); got != 0 {
		t.Errorf("%d counters after paying the counter branch, want 0", got)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a 2/2 equipped with no counters is %d/%d", p, tough)
	}
	// With no counter left that branch is unpayable; {3} is.
	if err := g.ActivateCatalogAbility(me.ID, gavel, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, CostBranch: &branch, CounterKind: game.CounterCharge}); err == nil {
		t.Error("removed a counter from an Equipment with none")
	}
	mana := 0
	if err := g.ActivateCatalogAbility(me.ID, gavel, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, CostBranch: &mana}); err == nil {
		t.Error("equipped for {3} with no mana")
	}
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, gavel, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}, CostBranch: &mana}); err != nil {
		t.Fatalf("equip for {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(gavel, game.CounterCharge, 4) })
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("a 2/2 with four counters on its Gavel is %d/%d, want 6/6", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Error("four counters did not grant double strike")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(gavel, game.CounterCharge, -1) })
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Error("three counters still grant double strike")
	}
}

func TestCmdrB9Group3CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{}
	for _, n := range []string{"Hearthhull, the Worldseed", "Inspirit, Flagship Vessel", "Exploration Broodship"} {
		want[cmdrB9Oracle(t, n)] = n
	}
	cmdrB9Registered(t, want)
}

func cmdrB9Charge(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterCharge, n) })
}

func TestCmdrB9HearthhullDrawsAndSacrificesLandsOnlyFromTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hull := effectstest.B12Push(g, me.ID, "Hearthhull, the Worldseed", "Legendary Artifact — Spacecraft", cmdrB9Oracle(t, "Hearthhull, the Worldseed"), 6, 7)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.FloatForTest(g, me, "C")
	// Ability 0 is station, 1 the draw ability.
	if err := g.ActivateCatalogAbility(me.ID, hull, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("the 2+ ability worked with no charge counters")
	}
	cmdrB9Charge(g, hull, 2)
	hand, drops := me.Hand.Size(), g.LandDropsRemainingFor(me.ID)
	if err := g.ActivateCatalogAbility(me.ID, hull, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2", hand, me.Hand.Size())
	}
	if got := g.LandDropsRemainingFor(me.ID); got != drops+1 {
		t.Errorf("land drops %d → %d, want +1", drops, got)
	}
	if c, _ := effectstest.BattlefieldCard(g, hull); c.IsCreature() {
		t.Error("a Spacecraft with two counters is a creature")
	}
}

func TestCmdrB9HearthhullDrainsOnEachSacrificedLandFromEightCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hull := effectstest.B12Push(g, me.ID, "Hearthhull, the Worldseed", "Legendary Artifact — Spacecraft", cmdrB9Oracle(t, "Hearthhull, the Worldseed"), 6, 7)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	emit := func() {
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventSacrifice, Actor: me.ID, CardID: land}) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	cmdrB9Charge(g, hull, 7)
	emit()
	for _, s := range g.Seats {
		if s.ID != me.ID && s.Life != 40 {
			t.Fatalf("seven counters drained: %s at %d", s.Name, s.Life)
		}
	}
	cmdrB9Charge(g, hull, 1)
	emit()
	for _, s := range g.Seats {
		want := 40
		if s.ID != me.ID {
			want = 38
		}
		if s.Life != want {
			t.Errorf("%s at %d, want %d", s.Name, s.Life, want)
		}
	}
	if p, tough := effectstest.PtOf(t, g, hull); p != 6 || tough != 7 {
		t.Errorf("Hearthhull is %d/%d, want 6/7", p, tough)
	}
	for _, kw := range []string{"flying", "vigilance", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, hull, kw) {
			t.Errorf("no %s at eight counters", kw)
		}
	}
}

func TestCmdrB9InspiritPutsYourChoiceOfCountersOnAnotherArtifact(t *testing.T) {
	for _, charge := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		inspirit := effectstest.B12Push(g, me.ID, "Inspirit, Flagship Vessel", "Legendary Artifact — Spacecraft", cmdrB9Oracle(t, "Inspirit, Flagship Vessel"), 5, 5)
		cmdrB9Charge(g, inspirit, 1)
		rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.AnswerPickTargetIfOpen(t, g, rock)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestOptionPickFor(g, me.ID)
		if pick == nil {
			t.Fatal("no counter choice")
		}
		idx := 0
		if charge {
			idx = 1
		}
		if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		plus, charges := effectstest.CounterCount(g, rock, game.CounterPlusOne), effectstest.CounterCount(g, rock, game.CounterCharge)
		if charge && (plus != 0 || charges != 2) || !charge && (plus != 1 || charges != 0) {
			t.Errorf("charge %v: +1/+1 %d, charge %d", charge, plus, charges)
		}
	}
}

func TestCmdrB9InspiritDoesNothingBelowOneCounterAndGrantsHexproofAtEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	inspirit := effectstest.B12Push(g, me.ID, "Inspirit, Flagship Vessel", "Legendary Artifact — Spacecraft", cmdrB9Oracle(t, "Inspirit, Flagship Vessel"), 5, 5)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("the 1+ trigger ran with no charge counters")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, rock, "hexproof") {
		t.Error("hexproof below eight counters")
	}
	cmdrB9Charge(g, inspirit, 8)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, rock, kw) {
			t.Errorf("the other artifact lacks %s at eight counters", kw)
		}
		if effectstest.HasKeywordOnBattlefield(t, g, inspirit, kw) {
			t.Errorf("Inspirit itself has %s (other artifacts only)", kw)
		}
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, inspirit, "flying") {
		t.Error("no flying at eight counters")
	}
}

func TestCmdrB9BroodshipGivesAnExtraLandAtThreeAndRecastsPermanentsFromTheGraveyardAtEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	ship := effectstest.B12Push(g, me.ID, "Exploration Broodship", "Artifact — Spacecraft", cmdrB9Oracle(t, "Exploration Broodship"), 4, 4)
	base := g.LandDropsRemainingFor(me.ID)
	cmdrB9Charge(g, ship, 2)
	if got := g.LandDropsRemainingFor(me.ID); got != base {
		t.Errorf("two counters: land drops %d, want %d", got, base)
	}
	cmdrB9Charge(g, ship, 1)
	if got := g.LandDropsRemainingFor(me.ID); got != base+1 {
		t.Errorf("three counters: land drops %d, want %d", got, base+1)
	}
	// 3 counters: no graveyard cast yet.
	cardID := uuid.New()
	effectstest.PushToGraveyard(g, me, game.Card{InstanceID: cardID, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2})
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	var bear uuid.UUID
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Bear" {
			bear = c.InstanceID
		}
	}
	effectstest.FloatForTest(g, me, "C")
	params := game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: cardsex.ExplorationBroodshipAltCost, AltCostIDs: []uuid.UUID{land}}
	if err := g.CastSpell(me.ID, bear, params); err == nil {
		t.Fatal("cast from the graveyard with only three counters")
	}
	cmdrB9Charge(g, ship, 5)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: cardsex.ExplorationBroodshipAltCost}); err == nil {
		t.Fatal("cast without naming a land to sacrifice")
	}
	if err := g.CastSpell(me.ID, bear, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear did not return")
	}
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	// Once during each of your turns.
	other := effectstest.PushToGraveyard(g, me, game.Card{Name: "Ox", TypeLine: "Creature — Ox", ManaCost: "{1}", Power: 2, Toughness: 2})
	land2 := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: cardsex.ExplorationBroodshipAltCost, AltCostIDs: []uuid.UUID{land2}}); err == nil {
		t.Error("a second graveyard cast in one turn")
	}
}

func TestCmdrB9Group4CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{}
	for _, n := range []string{"Szarel, Genesis Shepherd", "Evendo Brushrazer"} {
		want[cmdrB9Oracle(t, n)] = n
	}
	cmdrB9Registered(t, want)
}

func TestCmdrB9SzarelPutsItsPowerInCountersOnAnotherCreatureForEachOtherNontokenSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	effectstest.AdvanceToMain(t, g)
	szarel := effectstest.B12Push(g, me.ID, "Szarel, Genesis Shepherd", "Legendary Creature — Insect Druid", cmdrB9Oracle(t, "Szarel, Genesis Shepherd"), 2, 5)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	sac := func(actor, card uuid.UUID) {
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventSacrifice, Actor: actor, CardID: card}) })
		effectstest.AnswerPickTargetIfOpen(t, g, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	sac(me.ID, szarel)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("Szarel's own sacrifice put %d counters on the Bear", got)
	}
	sac(me.ID, rock)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters after sacrificing another nontoken permanent, want Szarel's power 2", got)
	}
	var token uuid.UUID
	g.WithWriteLock(func() {
		ids, _ := g.CreateTokensForEffect(me.ID, effects.TreasureToken(), 1, game.TokenEntryOptions{})
		if len(ids) > 0 {
			token = ids[0]
		}
	})
	sac(me.ID, token)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("sacrificing a token changed the counters: %d", got)
	}
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	sac(opp.ID, theirs)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's sacrifice changed the counters: %d", got)
	}
}

func TestCmdrB9SzarelPlaysLandsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushToGraveyard(g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("played a land from the graveyard with no Szarel")
	}
	effectstest.B12Push(g, me.ID, "Szarel, Genesis Shepherd", "Legendary Creature — Insect Druid", cmdrB9Oracle(t, "Szarel, Genesis Shepherd"), 2, 5)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("play: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not reach the battlefield")
	}
}

func TestCmdrB9EvendoBrushrazerExilesOnNontokenSacrificesAndPlaysThemAfterOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	evendo := effectstest.B12Push(g, me.ID, "Evendo Brushrazer", "Creature — Insect Warrior", cmdrB9Oracle(t, "Evendo Brushrazer"), 2, 2)
	forestA := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	top := effectstest.FrPushTop(me, "Mountain", "Basic Land — Mountain", "")
	// {T}, Sacrifice a land: Add {R}{R}.
	if err := g.ActivateManaAbility(me.ID, evendo, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{forestA}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forestA) {
		t.Fatal("the land was not sacrificed")
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the top card of the library was not exiled")
	}
	// A land exiled with Evendo is playable this turn, after the sacrifice.
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	if !g.Battlefield.Contains(top) {
		t.Error("the exiled land did not reach the battlefield")
	}
}

func TestCmdrB9EvendoBrushrazerCardsAreNotPlayableBeforeASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	evendo := effectstest.B12Push(g, me.ID, "Evendo Brushrazer", "Creature — Insect Warrior", cmdrB9Oracle(t, "Evendo Brushrazer"), 2, 2)
	var epoch int
	g.WithWriteLock(func() { epoch = g.ObjectEpochForEffect(evendo) })
	card := effectstest.ElPushExiled(g, me.ID, game.PermissionCardRef{ID: evendo, Epoch: epoch}, "Mountain", "Basic Land — Mountain", "")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("played an exiled card before sacrificing a nontoken permanent")
	}
}

func TestCmdrB9Group5CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{}
	for _, n := range []string{"World Breaker", "Centaur Vinecrasher", "Loamcrafter Faun", "Soul of Windgrace"} {
		want[cmdrB9Oracle(t, n)] = n
	}
	cmdrB9Registered(t, want)
}

func TestCmdrB9WorldBreakerExilesOnCastIsColorlessAndComesBackForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	_ = bear
	effectstest.FloatForTest(g, me, "GCCCCCC")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "World Breaker", TypeLine: "Creature — Eldrazi", ManaCost: "{6}{G}",
			Power: 5, Toughness: 7, OracleID: cmdrB9Oracle(t, "World Breaker"), Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.AnswerPickTargetIfOpen(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !g.Exile.Contains(rock) {
		t.Error("the cast trigger did not exile the artifact")
	}
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("World Breaker did not resolve")
	}
	if cols := c.Effective().Colors; len(cols) != 0 {
		t.Errorf("devoid: colors %v, want none", cols)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "reach") {
		t.Error("no reach")
	}
	// Back from the graveyard for {2}{C} and a land.
	effectstest.KillCreature(t, g, me.ID, id)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) || g.Battlefield.Contains(land) {
		t.Error("World Breaker is not back in hand for the land")
	}
}

func TestCmdrB9WorldBreakerNeedsColorlessManaToReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "World Breaker", TypeLine: "Creature — Eldrazi",
		OracleID: cmdrB9Oracle(t, "World Breaker"), Owner: me.ID, Controller: me.ID})
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.FloatForTest(g, me, "GGG")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("{C} was paid with coloured mana")
	}
}

func TestCmdrB9CentaurVinecrasherEntersWithACounterPerLandCardInAllGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushToGraveyard(g, opp, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	effectstest.PushToGraveyard(g, opp, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	crasher := cmdrB9Cast(t, g, me, "Centaur Vinecrasher", "Creature — Plant Centaur", "{3}{G}", "GCCC", 1, 1)
	if got := effectstest.CounterCount(g, crasher, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2 (a land card in each of two graveyards)", got)
	}
}

func TestCmdrB9CentaurVinecrasherReturnsFromTheGraveyardOncePerLandCardForGG(t *testing.T) {
	routes := map[string]func(g *game.Game, me *game.Player, land uuid.UUID){
		"a land dying": func(g *game.Game, me *game.Player, land uuid.UUID) {
			_ = g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, land)
		},
		"a land milled": func(g *game.Game, me *game.Player, land uuid.UUID) {
			g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
		},
		"a land discarded": func(g *game.Game, me *game.Player, land uuid.UUID) {
			g.WithWriteLock(func() { _ = g.DiscardCardsKeyedForEffect(me.ID, []uuid.UUID{land}, uuid.Nil, game.KeyedTail{}) })
		},
	}
	for name, route := range routes {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		crasher := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: crasher, Name: "Centaur Vinecrasher", TypeLine: "Creature — Plant Centaur",
			OracleID: cmdrB9Oracle(t, "Centaur Vinecrasher"), Owner: me.ID, Controller: me.ID})
		land := uuid.New()
		switch name {
		case "a land dying":
			g.Battlefield.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		case "a land milled":
			me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		case "a land discarded":
			me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		}
		route(g, me, land)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.FloatForTest(g, me, "GG")
		prompts := 0
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePayUnless {
				prompts++
			}
		}
		if prompts != 1 {
			t.Errorf("%s: %d payment prompts, want 1", name, prompts)
			continue
		}
		pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
		if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
			t.Fatalf("%s: ResolvePayUnless: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(crasher) {
			t.Errorf("%s: the Vinecrasher did not return", name)
		}
	}
}

func TestCmdrB9CentaurVinecrasherIgnoresNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Centaur Vinecrasher", TypeLine: "Creature — Plant Centaur",
		OracleID: cmdrB9Oracle(t, "Centaur Vinecrasher"), Owner: me.ID, Controller: me.ID})
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	effectstest.FloatForTest(g, me, "GG")
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Error("a creature card going to the graveyard offered the return")
	}
}

func TestCmdrB9LoamcrafterFaunDiscardsLandsAndReturnsThatManyPermanentCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushToGraveyard(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	ox := effectstest.PushToGraveyard(g, me, game.Card{Name: "Ox", TypeLine: "Creature — Ox", Power: 2, Toughness: 2})
	rock := effectstest.PushToGraveyard(g, me, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bolt := effectstest.PushToGraveyard(g, me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	forestGY := effectstest.PushToGraveyard(g, me, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	cmdrB9Cast(t, g, me, "Loamcrafter Faun", "Creature — Satyr Druid", "{2}{G}", "GCC", 3, 3)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("discard bounds %d..%d, want 0..2 (the land cards in hand)", pick.ChooseMin, pick.ChooseMax)
	}
	var lands []uuid.UUID
	for _, c := range me.Hand.Cards {
		if c.IsLand() {
			lands = append(lands, c.InstanceID)
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{spell}); err == nil {
		t.Fatal("discarding a nonland card was accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lands...)
	effectstest.PassPriorityAroundTable(t, g)
	targets := effectstest.LatestPickTarget(g, me.ID)
	if targets == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if err := g.ResolvePickTargets(targets.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: ox}, {Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("three targets were accepted for two discarded lands")
	}
	if err := g.ResolvePickTargets(targets.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err == nil {
		t.Fatal("an instant card was accepted as a nonland permanent card")
	}
	if err := g.ResolvePickTargets(targets.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: forestGY}}); err == nil {
		t.Fatal("a land card was accepted as a nonland permanent card")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear, ox)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(ox) {
		t.Error("the targeted permanent cards did not return")
	}
	if me.Hand.Contains(rock) {
		t.Error("a third card returned for two discarded lands")
	}
}

func TestCmdrB9SoulOfWindgraceFetchesALandFromAnyGraveyardAndHasLandDiscardAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.PushToGraveyard(g, opp, game.Card{Name: "Their Island", TypeLine: "Basic Land — Island"})
	soul := cmdrB9Cast(t, g, me, "Soul of Windgrace", "Legendary Creature — Cat Avatar", "{1}{B}{R}{G}", "BRGC", 5, 4)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 {
		t.Fatalf("want an optional pick, got %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok || !c.Tapped || c.Controller != me.ID {
		t.Fatalf("the land: on battlefield %v, tapped %v, controller %v", ok, c.Tapped, c.Controller)
	}
	// {G}, Discard a land card: gain 3 life.
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "G")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, soul, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("gain life: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	// {2}{B}, Discard a land card: indestructible, then tap.
	land2 := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land2, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "BCC")
	if err := g.ActivateCatalogAbility(me.ID, soul, 2, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{land2}}); err != nil {
		t.Fatalf("indestructible: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, soul, "indestructible") || !effectstest.TappedForTest(t, g, soul) {
		t.Error("Soul of Windgrace is not tapped and indestructible")
	}
	// A nonland card cannot pay the discard.
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "1R")
	if err := g.ActivateCatalogAbility(me.ID, soul, 1, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Error("a nonland card paid 'discard a land card'")
	}
}

func TestCmdrB9Group6CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{}
	for _, n := range []string{"Eumidian Wastewaker", "Planetary Annihilation"} {
		want[cmdrB9Oracle(t, n)] = n
	}
	cmdrB9Registered(t, want)
}

func TestCmdrB9PlanetaryAnnihilationKeepsSixLandsEachAndDealsSixToEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	effectstest.AdvanceToMain(t, g)
	var mine, theirs []uuid.UUID
	for i := 0; i < 8; i++ {
		mine = append(mine, effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest"))
	}
	for i := 0; i < 6; i++ {
		theirs = append(theirs, effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island"))
	}
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Beast", 7, 7)
	cmdrB9Cast(t, g, me, "Planetary Annihilation", "Sorcery", "{3}{R}{R}", "RRCCC", 0, 0)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 6 || pick.ChooseMax != 6 {
		t.Fatalf("want a pick of exactly six lands, got %+v", pick)
	}
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("a player with exactly six lands was asked")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, mine[:5]); err == nil {
		t.Fatal("five lands were accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, mine[2:]...)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range mine {
		want := i >= 2
		if g.Battlefield.Contains(id) != want {
			t.Errorf("land %d on the battlefield %v, want %v", i, !want, want)
		}
	}
	for _, id := range theirs {
		if !g.Battlefield.Contains(id) {
			t.Error("a land of the player with six was sacrificed")
		}
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived 6 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 7/7 died to 6 damage")
	}
}

func TestCmdrB9PlanetaryAnnihilationStillDamagesWhenNobodyHasMoreThanSixLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	cmdrB9Cast(t, g, me, "Planetary Annihilation", "Sorcery", "{3}{R}{R}", "RRCCC", 0, 0)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived")
	}
}

func TestCmdrB9WastewakerEachPlayerDiscardsOrSacrificesAndDrawsForLandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	effectstest.AdvanceToMain(t, g)
	ww := effectstest.B12Push(g, me.ID, "Eumidian Wastewaker", "Creature — Insect Cleric", cmdrB9Oracle(t, "Eumidian Wastewaker"), 4, 4)
	// I will discard a land card; the defender will sacrifice a land.
	myLand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: myLand, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	theirLand := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: ww, Actor: me.ID, Target: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not asked discard or sacrifice")
	}
	if effectstest.LatestOptionPickFor(g, opp.ID) != nil {
		t.Error("the defending player was asked before the controller answered")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if c := effectstest.ChooseCardsChoiceFor(g, me.ID); c == nil {
		t.Fatal("no discard prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, myLand)
	effectstest.PassPriorityAroundTable(t, g)
	pick = effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the defending player was not asked")
	}
	sacIdx := -1
	for i, o := range pick.PickOptions {
		if o.Label == cardse.WastewakerSacrificeOption {
			sacIdx = i
		}
	}
	if sacIdx < 0 {
		t.Fatalf("no sacrifice option: %+v", pick.PickOptions)
	}
	hand := me.Hand.Size()
	if err := g.ResolveOptionPick(pick.ID, opp.ID, sacIdx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	// The defending player names the land to sacrifice.
	effectstest.AnswerSacrifice(t, g, opp.ID, theirLand)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(myLand) || g.Battlefield.Contains(theirLand) {
		t.Fatal("the discard or the sacrifice did not happen")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards for two land cards put into graveyards, want 2", got)
	}
}

func TestCmdrB9WastewakerForcedSacrificeAndLandTokensDoNotDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for _, p := range []*game.Player{me, opp} {
			for p.Hand.Size() > 0 {
				c, _ := p.Hand.PopTop()
				p.Library.PushBottom(c)
			}
		}
	})
	ww := effectstest.B12Push(g, me.ID, "Eumidian Wastewaker", "Creature — Insect Cleric", cmdrB9Oracle(t, "Eumidian Wastewaker"), 4, 4)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Land Token", TypeLine: "Token Land",
		Owner: opp.ID, Controller: opp.ID})
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: ww, Actor: me.ID, Target: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("a player with no cards in hand was offered the discard")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, ww)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ww) || g.Battlefield.Contains(token) {
		t.Error("the forced sacrifices did not happen")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw (a creature card and a land token)", hand, me.Hand.Size())
	}
}

func TestCmdrB9DepthshakerTitanAnimatesArtifactsGrantsKeywordsAndSacrificesThemAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 4, 4)
	other := effectstest.B12Permanent(g, g.Seats[1].ID, "Their Rock", "Artifact")
	effectstest.FloatForTest(g, me, "RRCCCCC")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Depthshaker Titan", TypeLine: "Artifact Creature — Robot", ManaCost: "{5}{R}{R}",
			Power: 5, Toughness: 5, OracleID: cmdrB9Oracle(t, "Depthshaker Titan"), Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: golem}}); err == nil {
		t.Fatal("an artifact creature was accepted as a noncreature artifact")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: other}}); err == nil {
		t.Fatal("an opponent's artifact was accepted")
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, rock); p != 3 || tough != 3 {
		t.Fatalf("the Rock is %d/%d, want 3/3", p, tough)
	}
	if c, _ := effectstest.BattlefieldCard(g, rock); !c.IsCreature() || !c.IsArtifact() {
		t.Error("the Rock is not an artifact creature")
	}
	for _, who := range []uuid.UUID{rock, golem, id} {
		for _, kw := range []string{game.KeywordMelee, "trample", "haste"} {
			if !effectstest.HasKeywordOnBattlefield(t, g, who, kw) {
				t.Errorf("an artifact creature lacks %s", kw)
			}
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, other, "haste") {
		t.Error("an opponent's artifact got haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the animated Rock was not sacrificed at the beginning of the end step")
	}
	if !g.Battlefield.Contains(golem) || !g.Battlefield.Contains(id) {
		t.Error("something besides the animated artifact was sacrificed")
	}
}

func TestCmdrB9MoxiteRefineryMovesXCountersOffOnePermanentOfAnyKinds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	refinery := effectstest.B12Push(g, me.ID, "Moxite Refinery", "Artifact", cmdrB9Oracle(t, "Moxite Refinery"), 0, 0)
	bear := effectstest.PushCounterCreature(g, me.ID, "Bear", game.CounterPlusOne, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterCharge, 1) })
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	other := effectstest.PushCounterCreature(g, me.ID, "Ox", game.CounterPlusOne, 1)
	effectstest.FloatForTest(g, me, "CC")
	// Counters from two permanents are refused.
	if err := g.ActivateCatalogAbility(me.ID, refinery, 0, game.ActivateAbilityParams{Strict: true,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
		CounterSourceIDs: []uuid.UUID{bear, other}, CounterCounts: []int{1, 1}, CounterKinds: []string{game.CounterPlusOne, game.CounterPlusOne}}); err == nil {
		t.Fatal("counters from two permanents were accepted")
	}
	// Three counters of two kinds off the Bear, onto the Rock.
	if err := g.ActivateCatalogAbility(me.ID, refinery, 0, game.ActivateAbilityParams{Strict: true,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
		CounterSourceIDs: []uuid.UUID{bear, bear}, CounterCounts: []int{2, 1}, CounterKinds: []string{game.CounterPlusOne, game.CounterCharge}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, rock, game.CounterCharge); got != 3 {
		t.Errorf("%d charge counters on the Rock, want X = 3", got)
	}
	if effectstest.CounterCount(g, bear, game.CounterPlusOne) != 0 || effectstest.CounterCount(g, bear, game.CounterCharge) != 0 {
		t.Error("the counters were not removed from the Bear")
	}
}

func TestCmdrB9MoxiteRefineryPutsPlusOnePlusOneCountersAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	refinery := effectstest.B12Push(g, me.ID, "Moxite Refinery", "Artifact", cmdrB9Oracle(t, "Moxite Refinery"), 0, 0)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, game.CounterCharge, 2) })
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	effectstest.FloatForTest(g, me, "CC")
	params := game.ActivateAbilityParams{Strict: true,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
		CounterSourceIDs: []uuid.UUID{rock}, CounterCounts: []int{2}, CounterKinds: []string{game.CounterCharge}}
	if err := g.ActivateCatalogAbility(me.ID, refinery, 1, params); err == nil {
		t.Fatal("an artifact was accepted as the target of the creature mode")
	}
	params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := g.ActivateCatalogAbility(me.ID, refinery, 1, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters on the Bear, want 2", got)
	}
	// Not at instant speed.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.Untap(g, refinery)
	effectstest.FloatForTest(g, me, "CC")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, game.CounterCharge, 1) })
	params.CounterCounts = []int{1}
	if err := g.ActivateCatalogAbility(me.ID, refinery, 1, params); err == nil {
		t.Error("activated outside a main phase")
	}
}
