package testsp

import (
	"testing"

	"github.com/google/uuid"

	cardsri "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ri"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pilot_makers_test.go — ADR 0186 slice 10 (#3552): the cards that make
// the saddle-and-crew Pilot token. Each proof asserts the token by what
// it counts for crew and saddle, not by its name alone, so a card that
// minted a vanilla 1/1 named Pilot would fail.

const (
	defendTheRiderOracle      = "561af57d-d43d-4d98-9ef6-18a0843f22ae"
	valorsFlagshipOracle      = "a3023de2-61e7-4f46-aa4a-09d08ee7d939"
	backOnTrackOracle         = "aaf46eaf-2e01-46ac-b79d-6b70739d11c2"
	roadsideAssistanceOracle  = "814f30a1-c0d6-4243-839d-0645d7bcbe4c"
	cloudspireCoordinatorOrcl = "65d8b481-ebe9-4029-98a8-dd4693340bff"
)

// pilotTokens returns every Pilot token on the battlefield.
func pilotTokens(g *game.Game) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Pilot" && c.TokenKey == game.TokenKey("pilot-saddles-and-crews") {
				out = append(out, c)
			}
		}
	})
	return out
}

// requirePilots asserts n saddle-and-crew Pilots, each counting 3 for
// crew and for saddle.
func requirePilots(t *testing.T, g *game.Game, n int) {
	t.Helper()
	toks := pilotTokens(g)
	if len(toks) != n {
		t.Fatalf("%d Pilot tokens, want %d", len(toks), n)
	}
	for _, tok := range toks {
		for _, use := range []game.CostPowerUse{game.CostPowerCrew, game.CostPowerSaddle} {
			if got := effectstest.CountedPowerOf(t, g, tok.InstanceID, use); got != 3 {
				t.Errorf("Pilot counts %d for use %d, want 3", got, use)
			}
		}
	}
}

func TestDefendTheRiderMakesAPilot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastModal(t, g, "Defend the Rider", "Instant", defendTheRiderOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	requirePilots(t, g, 1)
}

func TestDefendTheRiderProtectsTheChosenPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.CastModal(t, g, "Defend the Rider", "Instant", defendTheRiderOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	ab := effectstest.EffectiveAbilities(t, g, bear)
	if !effectstest.ContainsString(ab, "hexproof") || !effectstest.ContainsString(ab, "indestructible") {
		t.Errorf("abilities %v, want hexproof and indestructible", ab)
	}
	if n := len(pilotTokens(g)); n != 0 {
		t.Errorf("the protect mode made %d Pilots", n)
	}
}

func TestValorsFlagshipCycledForTwoMakesTwoPilots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogHandCard(me, "Valor's Flagship", "Legendary Artifact — Vehicle", valorsFlagshipOracle)
	// {X}{2}{W} with X=2.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{W}"); err != nil {
		t.Fatal(err)
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn"})
	var idx = -1
	for i, a := range game.ActivatedAbilitiesForCard(game.Card{OracleID: valorsFlagshipOracle}) {
		if a.Cycling {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no cycling ability on Valor's Flagship")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	requirePilots(t, g, 2)
}

func TestBackOnTrackReturnsACreatureOrVehicleAndMakesAPilot(t *testing.T) {
	for _, tc := range []struct {
		Name, TypeLine string
	}{
		{"Grave Bear", "Creature — Bear"},
		{"Grave Vehicle", "Artifact — Vehicle"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		card := uuid.New()
		me.Graveyard.PushTop(game.Card{
			InstanceID: card, Name: tc.Name, TypeLine: tc.TypeLine,
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
		effectstest.CastCatalogSpell(t, g, "Back on Track", "Sorcery", backOnTrackOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: card}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldByName(g, tc.Name) == uuid.Nil {
			t.Errorf("%s did not return to the battlefield", tc.Name)
		}
		requirePilots(t, g, 1)
	}
}

func TestBackOnTrackRefusesAnotherKindOfCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: card, Name: "Grave Relic", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	err := effectstest.CastCatalogSpellErr(t, g, "Back on Track", "Sorcery", backOnTrackOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: card}})
	if err == nil {
		t.Fatal("a plain artifact card was accepted as a target")
	}
}

func TestRoadsideAssistanceEnchantsAVehicleAndMakesAPilot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.PushTestVehicle(g, me.ID, 4, 4)
	aura := effectstest.CastCatalogSpell(t, g, "Roadside Assistance", effectstest.AuraTypeLine, roadsideAssistanceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: vehicle}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("Roadside Assistance did not resolve to the battlefield")
	}
	if got := effectstest.EffectivePower(t, g, vehicle); got != 5 {
		t.Errorf("enchanted Vehicle power %d, want 5", got)
	}
	if got := effectstest.EffectiveToughness(t, g, vehicle); got != 5 {
		t.Errorf("enchanted Vehicle toughness %d, want 5", got)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, vehicle), "lifelink") {
		t.Error("enchanted Vehicle lacks lifelink")
	}
	requirePilots(t, g, 1)
}

// castEntry brings a permanent in through the real cast path, which is
// what writes the per-turn entry tally the Coordinator reads.
func castEntry(t *testing.T, g *game.Game, name, typeLine string) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, name, typeLine, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestCloudspireCoordinatorCountsMountsAndVehiclesThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	coord := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: coord, Name: "Cloudspire Coordinator", OracleID: cloudspireCoordinatorOrcl,
		TypeLine: "Creature — Human Pilot", Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	castEntry(t, g, "Fresh Mount", "Creature — Horse Mount")
	castEntry(t, g, "Fresh Vehicle", "Artifact — Vehicle")
	if err := g.ActivateCatalogAbility(me.ID, coord, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	requirePilots(t, g, 2)
}

func TestCloudspireCoordinatorIgnoresLastTurnsEntries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	coord := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: coord, Name: "Cloudspire Coordinator", OracleID: cloudspireCoordinatorOrcl,
		TypeLine: "Creature — Human Pilot", Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	castEntry(t, g, "Old Mount", "Creature — Horse Mount")
	castEntry(t, g, "Old Vehicle", "Artifact — Vehicle")
	// Walk around the table back to this seat's next turn.
	turn := g.Turn.Seq
	for g.Turn.Seq < turn+len(g.Seats) || g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatal("setup: not back on the Coordinator's controller's turn")
	}
	if err := g.ActivateCatalogAbility(me.ID, coord, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(pilotTokens(g)); n != 0 {
		t.Errorf("%d Pilots from permanents that entered on an earlier turn, want 0", n)
	}
}

func TestRoadsEnterTappedUnlessYouControlAMountOrVehicle(t *testing.T) {
	for _, r := range cardsri.RoadsCycle {
		t.Run(r.Name+" bare", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.PlayLandFromHand(t, g, r.Name, r.OracleID)
			if c, ok := g.LookupCardForEffect(id); !ok || !c.Tapped {
				t.Errorf("%s entered untapped with no Mount or Vehicle", r.Name)
			}
		})
		t.Run(r.Name+" vehicle", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.PushTestVehicle(g, g.Seats[g.Turn.ActiveSeat].ID, 4, 4)
			id := effectstest.PlayLandFromHand(t, g, r.Name, r.OracleID)
			if c, ok := g.LookupCardForEffect(id); !ok || c.Tapped {
				t.Errorf("%s entered tapped while its controller had a Vehicle", r.Name)
			}
		})
		t.Run(r.Name+" mount", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.PushMountForTest(g, g.Seats[g.Turn.ActiveSeat].ID, "Road Mount", "", 2)
			id := effectstest.PlayLandFromHand(t, g, r.Name, r.OracleID)
			if c, ok := g.LookupCardForEffect(id); !ok || c.Tapped {
				t.Errorf("%s entered tapped while its controller had a Mount", r.Name)
			}
		})
	}
	t.Run("an opponent's Vehicle does not count", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		r := cardsri.RoadsCycle[0]
		effectstest.PushTestVehicle(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, 4, 4)
		id := effectstest.PlayLandFromHand(t, g, r.Name, r.OracleID)
		if c, ok := g.LookupCardForEffect(id); !ok || !c.Tapped {
			t.Errorf("%s entered untapped on an opponent's Vehicle", r.Name)
		}
	})
}

func TestRoadsSacrificeForAPilotAtSorcerySpeed(t *testing.T) {
	for _, r := range cardsri.RoadsCycle {
		t.Run(r.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.PushTestVehicle(g, me.ID, 4, 4)
			id := effectstest.PlayLandFromHand(t, g, r.Name, r.OracleID)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{"+r.Colour+"}"); err != nil {
				t.Fatal(err)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(id) {
				t.Errorf("%s is still on the battlefield after paying its sacrifice cost", r.Name)
			}
			effectstest.PassPriorityAroundTable(t, g)
			requirePilots(t, g, 1)
		})
	}
}

func TestRoadsRefuseToSacrificeOutsideAMainPhase(t *testing.T) {
	r := cardsri.RoadsCycle[0]
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushTestVehicle(g, me.ID, 4, 4)
	id := effectstest.PlayLandFromHand(t, g, r.Name, r.OracleID)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{"+r.Colour+"}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a Road sacrificed for a Pilot in combat")
	}
}
