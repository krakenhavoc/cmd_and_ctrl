package effects

import (
	"testing"

	"github.com/google/uuid"

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

// iceB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func iceB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner, Keywords: keywords,
	})
}

// iceB1Activate activates the source's ability `index` for `who` with
// the given targets, without resolving it.
func iceB1Activate(g *game.Game, who, source uuid.UUID, index int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, index, game.ActivateAbilityParams{Strict: true, Targets: targets})
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
		g := newCatalogGame(t)
		id := iceB1Push(g, g.Seats[0].ID, tc.name, "Creature", tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
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
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			id := iceB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			for i := 0; i < 2; i++ {
				dcMana(t, g, me.ID, tc.mana)
				if err := iceB1Activate(g, me.ID, id, tc.index); err != nil {
					t.Fatalf("activation %d: %v", i+1, err)
				}
				passPriorityAroundTable(t, g)
			}
			if p, tough := ptOf(t, g, id); p != tc.p+2*tc.dp || tough != tc.tough+2*tc.dt {
				t.Errorf("%s is %d/%d after two pumps, want %d/%d", tc.name, p, tough, tc.p+2*tc.dp, tc.tough+2*tc.dt)
			}
			if err := iceB1Activate(g, me.ID, id, tc.index); err == nil {
				t.Error("the pump was activated with an empty pool")
			}
		})
	}
}

// --- Abyssal Specter -----------------------------------------------------------------

func TestIceB1AbyssalSpecterMakesTheDamagedPlayerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	specter := iceB1Push(g, me.ID, "Abyssal Specter", "Creature — Specter", iceB1AbyssalSpecterOracle, 2, 3)
	bear := iceB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	d2dSetHand(opp, 3)
	d2dSetHand(other, 3)

	dealNoncombatDamage(g, me.ID, specter, bear)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 0 {
		t.Fatalf("damage to a creature asked its controller to discard %d", n)
	}

	dealCombatDamage(g, me.ID, specter, opp.ID)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1: %+v", n, g.PendingChoices)
	}
	if n := discardOwed(g, other.ID); n != 0 {
		t.Errorf("a player not dealt damage owes %d discards", n)
	}
}

// --- Zuran Spellcaster, Centaur Archer, Storm Spirit ----------------------------------

func TestIceB1ZuranSpellcasterPingsAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	caster := iceB1Push(g, me.ID, "Zuran Spellcaster", "Creature — Human Wizard", iceB1ZuranSpellcasterOracle, 1, 1)
	before := opp.Life
	if err := iceB1Activate(g, me.ID, caster, 0, playerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("the target is at %d life, want %d", opp.Life, before-1)
	}
	if !findCardAnywhere(t, g, caster).Tapped {
		t.Error("Zuran Spellcaster did not tap")
	}
}

func TestIceB1CentaurArcherShootsOnlyAFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	archer := iceB1Push(g, me.ID, "Centaur Archer", "Creature — Centaur Archer", iceB1CentaurArcherOracle, 3, 2)
	flier := iceB1Push(g, opp.ID, "Bird", "Creature — Bird", "", 2, 2, "flying")
	bear := iceB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := iceB1Activate(g, me.ID, archer, 0, cardTarget(bear)...); err == nil {
		t.Fatal("a creature without flying was accepted as the target")
	}
	if err := iceB1Activate(g, me.ID, archer, 0, cardTarget(flier)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, flier).DamageMarked; got != 1 {
		t.Errorf("the flier has %d damage, want 1", got)
	}
}

func TestIceB1StormSpiritDealsTwoToTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	spirit := iceB1Push(g, me.ID, "Storm Spirit", "Creature — Elemental Spirit", iceB1StormSpiritOracle, 3, 3)
	ogre := iceB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	if err := iceB1Activate(g, me.ID, spirit, 0, playerTarget(opp.ID)); err == nil {
		t.Fatal("a player was accepted as the target")
	}
	if err := iceB1Activate(g, me.ID, spirit, 0, cardTarget(ogre)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the target has %d damage, want 2", got)
	}
}

// --- Hydroblast -------------------------------------------------------------------------

func TestIceB1HydroblastTargetsAnythingButOnlyActsOnRed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := me.Life
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "{R}", []game.TargetRef{playerTarget(me.ID)})
	castModal(t, g, "Hydroblast", "Instant", iceB1HydroblastOracle, []int{0}, cardTarget(bolt))
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("the red Bolt resolved: life %d -> %d", before, me.Life)
	}

	merfolk := pushCostedPermanentForTest(g, opp.ID, "Merfolk", "Creature — Merfolk", "{U}")
	castModal(t, g, "Hydroblast", "Instant", iceB1HydroblastOracle, []int{1}, cardTarget(merfolk))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(merfolk) {
		t.Error("a blue permanent was destroyed")
	}

	goblin := pushCostedPermanentForTest(g, opp.ID, "Goblin", "Creature — Goblin", "{R}")
	castModal(t, g, "Hydroblast", "Instant", iceB1HydroblastOracle, []int{1}, cardTarget(goblin))
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := iceB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			walkToUpkeepOfSeat0(t, g)
			passPriorityAroundTable(t, g)
			if got := ageCountersOn(t, g, id); got != 1 {
				t.Fatalf("%d age counters, want 1", got)
			}
			dcMana(t, g, me.ID, tc.mana)
			answerPayUnless(t, g, me.ID, true)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("paying %s did not keep %s", tc.mana, tc.name)
			}
			if n := me.ManaPool.Total(); n != 0 {
				t.Errorf("%d mana left in the pool after paying the upkeep", n)
			}
		})
		t.Run(tc.name+" declined", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := iceB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			walkToUpkeepOfSeat0(t, g)
			passPriorityAroundTable(t, g)
			answerPayUnless(t, g, me.ID, false)
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
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mammoths := iceB1Push(g, me.ID, "Woolly Mammoths", "Creature — Elephant", iceB1WoollyMammothsOracle, 3, 2)
	pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	pushLandFor(g, opp.ID, "Snow-Covered Forest", "Basic Snow Land — Forest")
	if hasKeywordOnBattlefield(t, g, mammoths, "trample") {
		t.Fatal("trample with no snow land of its controller's")
	}
	iceB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
	if !hasKeywordOnBattlefield(t, g, mammoths, "trample") {
		t.Error("no trample with a snow land under its controller's control")
	}
}

func TestIceB1DriftOfTheDeadCountsYourSnowLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	pushLandFor(g, me.ID, "Snow-Covered Swamp", "Basic Snow Land — Swamp")
	pushLandFor(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island")
	pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	pushLandFor(g, opp.ID, "Snow-Covered Swamp", "Basic Snow Land — Swamp")
	pushPermanentForTest(g, me.ID, "Snow Totem", "", "Snow Artifact")
	drift := iceB1Push(g, me.ID, "Drift of the Dead", "Creature — Wall", iceB1DriftOfTheDeadOracle, 0, 0)
	if p, tough := ptOf(t, g, drift); p != 2 || tough != 2 {
		t.Fatalf("Drift of the Dead is %d/%d, want 2/2 (two snow lands of yours)", p, tough)
	}
	// Entering through the zone-move event, so the layers are recomputed.
	iceB1Push(g, me.ID, "Snow-Covered Mountain", "Basic Snow Land — Mountain", "", 0, 0)
	if p, tough := ptOf(t, g, drift); p != 3 || tough != 3 {
		t.Errorf("Drift of the Dead is %d/%d after a third snow land, want 3/3", p, tough)
	}
}
