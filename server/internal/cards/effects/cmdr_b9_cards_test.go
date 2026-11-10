package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

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
		g := newCatalogGame(t)
		id := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
		if !tappedForTest(t, g, id) {
			t.Errorf("%s did not enter tapped", tc.name)
		}
		g2 := newCatalogGame(t)
		me := g2.Seats[g2.Turn.ActiveSeat]
		before := me.Hand.Size()
		cid, _ := cycleFromHand(t, g2, tc.name, "Land", tc.oracle, tc.mana)
		if me.Hand.Contains(cid) {
			t.Errorf("%s is still in hand after cycling", tc.name)
		}
		if me.Hand.Size() != before {
			t.Errorf("%s: hand %d after cycling, want %d (discard one, draw one)", tc.name, me.Hand.Size(), before)
		}
	}
}

func TestCmdrB9CyclingNeedsItsColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Secluded Steppe", "Land", cmdrB9SecludedSteppeOracle)
	fillPool(me, 1)
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
		g := newCatalogGame(t)
		id := b12PlayFromHand(t, g, tc.name, "Artifact Land", tc.oracle, game.CastSpellParams{})
		if !tappedForTest(t, g, id) {
			t.Errorf("%s did not enter tapped", tc.name)
		}
		if !hasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("%s is not indestructible", tc.name)
		}
		spec, _ := Lookup(tc.oracle)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Dakmor Salvage", TypeLine: "Land",
		OracleID: cmdrB9DakmorSalvageOracle, Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	_ = g.DrawCard(me.ID)
	offer := latestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID)
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
	for _, s := range All() {
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
	advanceToMain(t, g)
	floatForTest(g, me, mana)
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
	passPriorityAroundTable(t, g)
	return id
}

func TestCmdrB9BalothPrimeEntersStunnedAndFeedsOnSacrificedLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	prime := cmdrB9Cast(t, g, me, "Baloth Prime", "Creature — Beast", "{3}{G}", "GGGG", 10, 10)
	c, ok := battlefieldCard(g, prime)
	if !ok {
		t.Fatalf("Baloth Prime did not resolve: stack %d, pending %d, hand has it %v", g.Stack.Size(), len(g.PendingChoices), me.Hand.Contains(prime))
	}
	if !c.Tapped || c.Counters[game.CounterStun] != 6 {
		t.Fatalf("tapped %v, stun %d; want tapped with six stun counters", c.Tapped, c.Counters[game.CounterStun])
	}
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	floatForTest(g, me, "CCCC")
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, prime, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{forest},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	if p, tough := effectivePTOf(t, g, beast.InstanceID); p != 4 || tough != 4 {
		t.Errorf("Beast is %d/%d, want 4/4", p, tough)
	}
	c, _ = battlefieldCard(g, prime)
	if c.Counters[game.CounterStun] != 5 || !c.Tapped {
		t.Errorf("untap with a stun counter: stun %d tapped %v, want 5 and still tapped", c.Counters[game.CounterStun], c.Tapped)
	}
}

func TestCmdrB9PatrollingPeacemakerEntersWithTwoCountersAndProliferatesOnAnOpponentsCrime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	peace := cmdrB9Cast(t, g, me, "Patrolling Peacemaker", "Artifact Creature — Robot Soldier", "{2}{W}", "WCC", 0, 0)
	if got := counterCount(g, peace, game.CounterPlusOne); got != 2 {
		t.Fatalf("entered with %d counters, want 2", got)
	}
	emit := func(actor uuid.UUID, crime bool) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: actor, CrimeCommitted: crime})
		})
		settleAnsweringProliferate(t, g)
	}
	emit(me.ID, true)
	if got := counterCount(g, peace, game.CounterPlusOne); got != 2 {
		t.Errorf("its controller's own crime proliferated: %d counters", got)
	}
	emit(opp.ID, false)
	if got := counterCount(g, peace, game.CounterPlusOne); got != 2 {
		t.Errorf("a non-crime targeting proliferated: %d counters", got)
	}
	emit(opp.ID, true)
	if got := counterCount(g, peace, game.CounterPlusOne); got != 3 {
		t.Errorf("an opponent's crime: %d counters, want 3", got)
	}
}

func TestCmdrB9KiloProliferatesWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	kilo := b12Push(g, me.ID, "Kilo, Apogee Mind", "Legendary Artifact Creature — Robot Artificer", cmdrB9Oracle(t, "Kilo, Apogee Mind"), 3, 3)
	ally := pushCounterCreature(g, me.ID, "Ally", game.CounterPlusOne, 1)
	untap(g, kilo)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == kilo {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
		g.EmitEvent(game.Event{Kind: game.EventTapCard, CardID: kilo, Actor: me.ID})
	})
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("Kilo's tap proliferated %d counters on the ally, want 2", got)
	}
}

func TestCmdrB9SurgeConductorProliferatesForEachOtherNontokenArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Surge Conductor", "Artifact Creature — Robot", cmdrB9Oracle(t, "Surge Conductor"), 3, 2)
	ally := pushCounterCreature(g, me.ID, "Ally", game.CounterPlusOne, 1)
	floatForTest(g, me, "CC")
	castFromHandForTest(t, g, me, "Trinket", "Artifact", "{2}", "", game.CastSpellParams{Strict: true})
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("a nontoken artifact entering: %d counters, want 2", got)
	}
	// A token does not trigger it.
	g.WithWriteLock(func() {
		_, _ = g.CreateTokensForEffect(me.ID, TreasureToken(), 1, game.TokenEntryOptions{})
	})
	settleAnsweringProliferate(t, g)
	if got := counterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("a token entering proliferated: %d counters, want 2", got)
	}
}

func TestCmdrB9ElshaMakesAMonkForEachPointOfCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if opp.ID == me.ID {
		opp = g.Seats[2]
	}
	advanceToMain(t, g)
	elsha := b12Push(g, me.ID, "Elsha, Threefold Master", "Legendary Creature — Djinn Monk", cmdrB9Oracle(t, "Elsha, Threefold Master"), 1, 1)
	dealCombatDamageToPlayer(g, elsha, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 3 {
		t.Fatalf("%d Monk tokens, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Monk" && !hasKeywordOnBattlefield(t, g, c.InstanceID, "prowess") {
			t.Error("a Monk token lacks prowess")
		}
	}
}

func TestCmdrB9UurgCountsLandsInYourGraveyardAndSurveilsAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	uurg := b12Push(g, me.ID, "Uurg, Spawn of Turg", "Legendary Creature — Frog Beast", cmdrB9Oracle(t, "Uurg, Spawn of Turg"), 0, 5)
	pushToGraveyard(g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushToGraveyard(g, me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	pushToGraveyard(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if p, tough := ptOf(t, g, uurg); p != 2 || tough != 5 {
		t.Errorf("%d/%d with two land cards in the graveyard, want 2/5", p, tough)
	}
	// Sacrificing a land for {B}{G} gains 2 life.
	land := seedLandOnBattlefield(g, me.ID, "Swamp", "Basic Land — Swamp")
	floatForTest(g, me, "BG")
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, uurg, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2", before, me.Life)
	}
	if p, _ := ptOf(t, g, uurg); p != 3 {
		t.Errorf("power %d after the sacrificed land joined the graveyard, want 3", p)
	}
}

func TestCmdrB9WakeThePastReturnsEveryArtifactCardWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushGraveyardCardTyped(me, "Relic", "Artifact")
	b := pushGraveyardCardTyped(me, "Golem", "Artifact Creature — Golem")
	c := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	cmdrB9Cast(t, g, me, "Wake the Past", "Sorcery", "{5}{R}{W}", "RWCCCCC", 0, 0)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("not every artifact card came back")
	}
	if g.Battlefield.Contains(c) || !me.Graveyard.Contains(c) {
		t.Error("a nonartifact card came back")
	}
	if !hasKeywordOnBattlefield(t, g, a, "haste") || !hasKeywordOnBattlefield(t, g, b, "haste") {
		t.Error("the returned artifacts did not gain haste")
	}
}

func TestCmdrB9MechanautMakesArtifactSpellsCostOneLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Enthusiastic Mechanaut", "Artifact Creature — Goblin Artificer", cmdrB9Oracle(t, "Enthusiastic Mechanaut"), 2, 2)
	floatForTest(g, me, "C")
	castFromHandForTest(t, g, me, "Trinket", "Artifact", "{2}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Trinket", me.ID); got != 1 {
		t.Errorf("a {2} artifact for one mana: %d on the battlefield, want 1", got)
	}
	floatForTest(g, me, "C")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a {2} creature was castable for one mana")
	}
}

func TestCmdrB9CoretapperPutsChargeCountersOnAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tapper := b12Push(g, me.ID, "Coretapper", "Artifact Creature — Myr", cmdrB9Oracle(t, "Coretapper"), 1, 1)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	bear := seedCreature(g, "Bear", me.ID)
	if err := g.ActivateCatalogAbility(me.ID, tapper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a nonartifact target was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, tapper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err != nil {
		t.Fatalf("tap ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, rock, game.CounterCharge); got != 1 {
		t.Errorf("%d charge counters after the tap, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, tapper, 1, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err != nil {
		t.Fatalf("sacrifice ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, rock, game.CounterCharge); got != 3 {
		t.Errorf("%d charge counters after the sacrifice, want 3", got)
	}
	if g.Battlefield.Contains(tapper) {
		t.Error("Coretapper was not sacrificed")
	}
}

func TestCmdrB9GolemFoundryChargesOnArtifactCastsAndMakesAGolem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	foundry := b12Push(g, me.ID, "Golem Foundry", "Artifact", cmdrB9Oracle(t, "Golem Foundry"), 0, 0)
	for i := 0; i < 3; i++ {
		floatForTest(g, me, "C")
		castFromHandForTest(t, g, me, "Trinket", "Artifact", "{1}", "", game.CastSpellParams{Strict: true})
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	if got := counterCount(g, foundry, game.CounterCharge); got != 3 {
		t.Fatalf("%d charge counters after three artifact spells, want 3", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, foundry, game.CounterCharge); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
	if got := countOnBattlefield(g, "Golem", me.ID); got != 1 {
		t.Errorf("%d Golem tokens, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without three charge counters")
	}
}

func TestCmdrB9TitanForgeChargesThenMakesANineNineGolem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	forge := b12Push(g, me.ID, "Titan Forge", "Artifact", cmdrB9Oracle(t, "Titan Forge"), 0, 0)
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("charge: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, forge, game.CounterCharge); got != 1 {
		t.Fatalf("%d counters, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(forge, game.CounterCharge, 2) })
	untap(g, forge)
	if err := g.ActivateCatalogAbility(me.ID, forge, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("make a Golem: %v", err)
	}
	passPriorityAroundTable(t, g)
	var golem uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Golem" && c.Controller == me.ID {
			golem = c.InstanceID
		}
	}
	if golem == uuid.Nil {
		t.Fatal("no Golem token")
	}
	if p, tough := effectivePTOf(t, g, golem); p != 9 || tough != 9 {
		t.Errorf("Golem is %d/%d, want 9/9", p, tough)
	}
	if got := counterCount(g, forge, game.CounterCharge); got != 0 {
		t.Errorf("%d counters left, want 0", got)
	}
}

func TestCmdrB9CrystallineCrawlerConvergesAndUsesItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crawler := cmdrB9Cast(t, g, me, "Crystalline Crawler", "Artifact Creature — Construct", "{4}", "WUBR", 1, 1)
	if got := counterCount(g, crawler, game.CounterPlusOne); got != 4 {
		t.Fatalf("%d counters for four colours of mana, want 4", got)
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	crawler2 := cmdrB9Cast(t, g2, me2, "Crystalline Crawler", "Artifact Creature — Construct", "{4}", "CCCC", 1, 1)
	if got := counterCount(g2, crawler2, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters for colourless mana, want 0", got)
	}
	// Tap for a counter.
	untap(g2, crawler2)
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
	passPriorityAroundTable(t, g2)
	if got := counterCount(g2, crawler2, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after the tap, want 1", got)
	}
}

func TestCmdrB9MindlessAutomatonGrowsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	auto := cmdrB9Cast(t, g, me, "Mindless Automaton", "Artifact Creature — Construct", "{4}", "CCCC", 0, 0)
	if got := counterCount(g, auto, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters on entering, want 2", got)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, auto, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
	if _, ok := battlefieldCard(g, auto); ok {
		t.Error("a 0/0 with no counters stayed on the battlefield")
	}
}

func TestCmdrB9GroundskeeperReturnsOnlyBasicLandCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	keeper := b12Push(g, me.ID, "Groundskeeper", "Creature — Human Druid", cmdrB9Oracle(t, "Groundskeeper"), 1, 1)
	basic := pushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	nonbasic := pushGraveyardCardTyped(me, "Tarn", "Land")
	floatForTest(g, me, "CG")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: nonbasic}}}); err == nil {
		t.Fatal("a nonbasic land card was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: basic}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(basic) {
		t.Error("the basic land card did not return to hand")
	}
}

func TestCmdrB9ExperimentalAuguryTakesOneBottomsTwoAndProliferates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ally := pushCounterCreature(g, me.ID, "Ally", game.CounterPlusOne, 1)
	ids := seedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	advanceToMain(t, g)
	floatForTest(g, me, "UC")
	castFromHandForTest(t, g, me, "Experimental Augury", "Instant", "{1}{U}", cmdrB9Oracle(t, "Experimental Augury"), game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 3 {
		t.Fatalf("want a pick of exactly one of three, got %+v", pick)
	}
	answerChooseCards(t, g, me.ID, ids[1])
	if order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[0], ids[2]}, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	if proliferatePrompt(g) == nil {
		t.Fatal("no proliferate prompt after the take and the ordering")
	}
	answerProliferate(t, g, ally)
	passPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 1 || !me.Hand.Contains(ids[1]) {
		t.Errorf("hand grew by %d, want B alone", me.Hand.Size()-hand)
	}
	if got := counterCount(g, ally, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters after proliferate, want 2", got)
	}
}
