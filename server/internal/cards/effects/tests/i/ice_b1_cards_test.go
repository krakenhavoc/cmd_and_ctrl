package testsi

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ice_b1_cards_test.go — Ice Age, batch 1 (#5636). Each test plays the
// real catalog entry and asserts what its text says.

const (
	iceB1AbyssalSpecterOracle    = "51833cff-6519-4806-8d91-0040e9a02189"
	iceB1AdarkarSentinelOracle   = "bd372a53-48f9-4e9a-ab00-0f9c4606d452"
	iceB1HoarShadeOracle         = "23c74c42-11bd-4fba-b1cb-93b96f5a11fb"
	iceB1WallOfLavaOracle        = "bf7734fb-011d-4f12-a174-1ae879dab95b"
	iceB1ThunderWallOracle       = "904b9056-ad76-4805-b14f-c426e4a57bae"
	iceB1ZuranSpellcasterOracle  = "bc7b90b1-3517-4e5d-9bd8-68b4d8a259fd"
	iceB1CentaurArcherOracle     = "d48eb6ee-d4ca-4bc7-8d7b-0e6261b88ae5"
	iceB1StormSpiritOracle       = "6a413d09-dd67-4c3e-b81b-7bfe8b864dc9"
	iceB1HydroblastOracle        = "b94c0336-2ca8-4bb1-bb1f-e9babd1826ad"
	iceB1IllusionaryForcesOracle = "e17e3bda-75cb-47a1-9f60-f50705a46a1e"
	iceB1IllusionaryWallOracle   = "77a8945f-4d52-4445-8913-425007e7948c"
	iceB1SoldeviSimulacrumOracle = "227ed0b7-7a9a-41f1-a190-1d1b85a44720"
	iceB1WoollyMammothsOracle    = "6e5f1088-a5ab-48ec-be27-c88eb164cf63"
	iceB1DriftOfTheDeadOracle    = "bcea5e7e-655d-48ea-871e-28ca49e5c574"
)

func TestIceB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		iceB1AbyssalSpecterOracle:    "Abyssal Specter",
		iceB1AdarkarSentinelOracle:   "Adarkar Sentinel",
		iceB1HoarShadeOracle:         "Hoar Shade",
		iceB1WallOfLavaOracle:        "Wall of Lava",
		iceB1ThunderWallOracle:       "Thunder Wall",
		iceB1ZuranSpellcasterOracle:  "Zuran Spellcaster",
		iceB1CentaurArcherOracle:     "Centaur Archer",
		iceB1StormSpiritOracle:       "Storm Spirit",
		iceB1HydroblastOracle:        "Hydroblast",
		iceB1IllusionaryForcesOracle: "Illusionary Forces",
		iceB1IllusionaryWallOracle:   "Illusionary Wall",
		iceB1SoldeviSimulacrumOracle: "Soldevi Simulacrum",
		iceB1WoollyMammothsOracle:    "Woolly Mammoths",
		iceB1DriftOfTheDeadOracle:    "Drift of the Dead",
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

// --- printed keywords -----------------------------------------------------------

func TestIceB1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Abyssal Specter", iceB1AbyssalSpecterOracle, []string{"flying"}},
		{"Wall of Lava", iceB1WallOfLavaOracle, []string{"defender"}},
		{"Thunder Wall", iceB1ThunderWallOracle, []string{"defender", "flying"}},
		{"Storm Spirit", iceB1StormSpiritOracle, []string{"flying"}},
		{"Illusionary Forces", iceB1IllusionaryForcesOracle, []string{"flying"}},
		{"Illusionary Wall", iceB1IllusionaryWallOracle, []string{"defender", "flying", "first strike"}},
		{"Drift of the Dead", iceB1DriftOfTheDeadOracle, []string{"defender"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.IceB1Push(g, g.Seats[0].ID, tc.name, "Creature", tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

// --- the pumps ------------------------------------------------------------------------

func TestIceB1SelfPumpsGrowTheirSourceEachActivation(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		index, p, tough              int
		dp, dt                       int
	}{
		{"Adarkar Sentinel", "Artifact Creature — Soldier", iceB1AdarkarSentinelOracle, "{C}", 0, 3, 3, 0, 1},
		{"Hoar Shade", "Creature — Shade", iceB1HoarShadeOracle, "{B}", 0, 1, 2, 1, 1},
		{"Wall of Lava", "Creature — Wall", iceB1WallOfLavaOracle, "{R}", 0, 1, 3, 1, 1},
		{"Thunder Wall", "Creature — Wall", iceB1ThunderWallOracle, "{U}", 0, 0, 2, 1, 1},
		{"Soldevi Simulacrum", "Artifact Creature — Soldier", iceB1SoldeviSimulacrumOracle, "{C}", 0, 2, 4, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.IceB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			for i := 0; i < 2; i++ {
				effectstest.DcMana(t, g, me.ID, tc.mana)
				if err := effectstest.IceB1Activate(g, me.ID, id, tc.index); err != nil {
					t.Fatalf("activation %d: %v", i+1, err)
				}
				effectstest.PassPriorityAroundTable(t, g)
			}
			if p, tough := effectstest.PtOf(t, g, id); p != tc.p+2*tc.dp || tough != tc.tough+2*tc.dt {
				t.Errorf("%s is %d/%d after two pumps, want %d/%d", tc.name, p, tough, tc.p+2*tc.dp, tc.tough+2*tc.dt)
			}
			if err := effectstest.IceB1Activate(g, me.ID, id, tc.index); err == nil {
				t.Error("the pump was activated with an empty pool")
			}
		})
	}
}

// --- Abyssal Specter -----------------------------------------------------------------

func TestIceB1AbyssalSpecterMakesTheDamagedPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	specter := effectstest.IceB1Push(g, me.ID, "Abyssal Specter", "Creature — Specter", iceB1AbyssalSpecterOracle, 2, 3)
	bear := effectstest.IceB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.D2dSetHand(opp, 3)
	effectstest.D2dSetHand(other, 3)

	effectstest.DealNoncombatDamage(g, me.ID, specter, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 0 {
		t.Fatalf("damage to a creature asked its controller to discard %d", n)
	}

	effectstest.DealCombatDamage(g, me.ID, specter, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1: %+v", n, g.PendingChoices)
	}
	if n := effectstest.DiscardOwed(g, other.ID); n != 0 {
		t.Errorf("a player not dealt damage owes %d discards", n)
	}
}

// --- Zuran Spellcaster, Centaur Archer, Storm Spirit ----------------------------------

func TestIceB1ZuranSpellcasterPingsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	caster := effectstest.IceB1Push(g, me.ID, "Zuran Spellcaster", "Creature — Human Wizard", iceB1ZuranSpellcasterOracle, 1, 1)
	before := opp.Life
	if err := effectstest.IceB1Activate(g, me.ID, caster, 0, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("the target is at %d life, want %d", opp.Life, before-1)
	}
	if !effectstest.FindCardAnywhere(t, g, caster).Tapped {
		t.Error("Zuran Spellcaster did not tap")
	}
}

func TestIceB1CentaurArcherShootsOnlyAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	archer := effectstest.IceB1Push(g, me.ID, "Centaur Archer", "Creature — Centaur Archer", iceB1CentaurArcherOracle, 3, 2)
	flier := effectstest.IceB1Push(g, opp.ID, "Bird", "Creature — Bird", "", 2, 2, "flying")
	bear := effectstest.IceB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.IceB1Activate(g, me.ID, archer, 0, effectstest.CardTarget(bear)...); err == nil {
		t.Fatal("a creature without flying was accepted as the target")
	}
	if err := effectstest.IceB1Activate(g, me.ID, archer, 0, effectstest.CardTarget(flier)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, flier).DamageMarked; got != 1 {
		t.Errorf("the flier has %d damage, want 1", got)
	}
}

func TestIceB1StormSpiritDealsTwoToTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	spirit := effectstest.IceB1Push(g, me.ID, "Storm Spirit", "Creature — Elemental Spirit", iceB1StormSpiritOracle, 3, 3)
	ogre := effectstest.IceB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	if err := effectstest.IceB1Activate(g, me.ID, spirit, 0, effectstest.PlayerTarget(opp.ID)); err == nil {
		t.Fatal("a player was accepted as the target")
	}
	if err := effectstest.IceB1Activate(g, me.ID, spirit, 0, effectstest.CardTarget(ogre)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the target has %d damage, want 2", got)
	}
}

// --- Hydroblast -------------------------------------------------------------------------

func TestIceB1HydroblastTargetsAnythingButOnlyActsOnRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := me.Life
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "{R}", []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.CastModal(t, g, "Hydroblast", "Instant", iceB1HydroblastOracle, []int{0}, effectstest.CardTarget(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("the red Bolt resolved: life %d -> %d", before, me.Life)
	}

	merfolk := effectstest.PushCostedPermanentForTest(g, opp.ID, "Merfolk", "Creature — Merfolk", "{U}")
	effectstest.CastModal(t, g, "Hydroblast", "Instant", iceB1HydroblastOracle, []int{1}, effectstest.CardTarget(merfolk))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(merfolk) {
		t.Error("a blue permanent was destroyed")
	}

	goblin := effectstest.PushCostedPermanentForTest(g, opp.ID, "Goblin", "Creature — Goblin", "{R}")
	effectstest.CastModal(t, g, "Hydroblast", "Instant", iceB1HydroblastOracle, []int{1}, effectstest.CardTarget(goblin))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(goblin) {
		t.Error("the red permanent survived")
	}
}

// --- cumulative upkeep ------------------------------------------------------------------

func TestIceB1CumulativeUpkeepCardsAreKeptOnlyWhilePaid(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
	}{
		{"Illusionary Forces", "Creature — Illusion", iceB1IllusionaryForcesOracle, "{U}"},
		{"Illusionary Wall", "Creature — Illusion Wall", iceB1IllusionaryWallOracle, "{U}"},
		{"Soldevi Simulacrum", "Artifact Creature — Soldier", iceB1SoldeviSimulacrumOracle, "{C}"},
	} {
		t.Run(tc.name+" paid", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.IceB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.WalkToUpkeepOfSeat0(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.AgeCountersOn(t, g, id); got != 1 {
				t.Fatalf("%d age counters, want 1", got)
			}
			effectstest.DcMana(t, g, me.ID, tc.mana)
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("paying %s did not keep %s", tc.mana, tc.name)
			}
			if n := me.ManaPool.Total(); n != 0 {
				t.Errorf("%d mana left in the pool after paying the upkeep", n)
			}
		})
		t.Run(tc.name+" declined", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.IceB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.WalkToUpkeepOfSeat0(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerPayUnless(t, g, me.ID, false)
			if g.Battlefield.Contains(id) {
				t.Fatalf("declining the upkeep did not sacrifice %s", tc.name)
			}
			if !me.Graveyard.Contains(id) {
				t.Errorf("%s is not in its owner's graveyard", tc.name)
			}
		})
	}
}

// --- snow ---------------------------------------------------------------------------------

func TestIceB1WoollyMammothsHasTrampleOnlyWithASnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mammoths := effectstest.IceB1Push(g, me.ID, "Woolly Mammoths", "Creature — Elephant", iceB1WoollyMammothsOracle, 3, 2)
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, opp.ID, "Snow-Covered Forest", "Basic Snow Land — Forest")
	if effectstest.HasKeywordOnBattlefield(t, g, mammoths, "trample") {
		t.Fatal("trample with no snow land of its controller's")
	}
	effectstest.IceB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
	if !effectstest.HasKeywordOnBattlefield(t, g, mammoths, "trample") {
		t.Error("no trample with a snow land under its controller's control")
	}
}

func TestIceB1DriftOfTheDeadCountsYourSnowLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.PushLandFor(g, me.ID, "Snow-Covered Swamp", "Basic Snow Land — Swamp")
	effectstest.PushLandFor(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island")
	effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.PushLandFor(g, opp.ID, "Snow-Covered Swamp", "Basic Snow Land — Swamp")
	effectstest.PushPermanentForTest(g, me.ID, "Snow Totem", "", "Snow Artifact")
	drift := effectstest.IceB1Push(g, me.ID, "Drift of the Dead", "Creature — Wall", iceB1DriftOfTheDeadOracle, 0, 0)
	if p, tough := effectstest.PtOf(t, g, drift); p != 2 || tough != 2 {
		t.Fatalf("Drift of the Dead is %d/%d, want 2/2 (two snow lands of yours)", p, tough)
	}
	// Entering through the zone-move event, so the layers are recomputed.
	effectstest.IceB1Push(g, me.ID, "Snow-Covered Mountain", "Basic Snow Land — Mountain", "", 0, 0)
	if p, tough := effectstest.PtOf(t, g, drift); p != 3 || tough != 3 {
		t.Errorf("Drift of the Dead is %d/%d after a third snow land, want 3/3", p, tough)
	}
}
