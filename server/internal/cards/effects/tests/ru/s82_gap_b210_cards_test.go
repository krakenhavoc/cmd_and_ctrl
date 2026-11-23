package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB210DireMimic  = "b657b3df-9287-44cf-a1cd-a80ead71a553"
	gapB210Fraying    = "b69d2bd6-3e99-4c51-89af-0acedd48739e"
	gapB210Lantern    = "00e25126-b807-4753-ba63-9a533755e5ce"
	gapB210Servant    = "0b59099f-2172-4eab-9c0d-7048e6771018"
	gapB210Mirror     = "3ec7ae18-b203-49bd-95f9-ad5482459a23"
	gapB210Baldric    = "5b173737-d33e-482f-a91e-f37e3e428e86"
	gapB210Orb        = "ba7a46f1-c2a5-42d3-93eb-4dcea7186e5e"
	gapB210Purse      = "f3f10173-fead-48a2-9726-a4e7efec3951"
	gapB210Tome       = "6ec54d84-7026-45a3-a7d7-37cd5c9a8463"
	gapB210PactWeapon = "bc2c02f6-2b6b-4016-b006-12ce0a1a2450"
)

func TestGapB210EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB210DireMimic: "Dire Mimic", gapB210Fraying: "Fraying Line", gapB210Lantern: "Lantern of Revealing",
		gapB210Servant: "Mighty Servant of Leuk-o", gapB210Mirror: "Mirror of Life Trapping",
		gapB210Baldric: "Multiclass Baldric", gapB210Orb: "Navigation Orb", gapB210Purse: "Noble's Purse",
		gapB210Tome: "Sarevok's Tome", gapB210PactWeapon: "Pact Weapon",
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

func TestGapB210NoblesPurseEntersTappedWithThreeCoinsAndMakesTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	purse := effectstest.CastCatalogSpell(t, g, "Noble's Purse", "Artifact", gapB210Purse, nil)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, purse)
	if !ok || !card.Tapped || card.Counters[effects.StdB31CoinCounter] != 3 {
		t.Fatalf("on battlefield %v, tapped %v, coin counters %d; want tapped with 3", ok, card.Tapped, card.Counters[effects.StdB31CoinCounter])
	}
	if err := g.ActivateCatalogAbility(me.ID, purse, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Purse was activated")
	}
	effectstest.UntapForTest(g, purse)
	if err := g.ActivateCatalogAbility(me.ID, purse, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures %d, want 1", n)
	}
	if card, _ := effectstest.BattlefieldCard(g, purse); card.Counters[effects.StdB31CoinCounter] != 2 {
		t.Errorf("coin counters %d, want 2", card.Counters[effects.StdB31CoinCounter])
	}
}

func TestGapB210DireMimicBecomesAFiveFiveShapeshifter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	mimic := effectstest.PushCatalogPermanent(g, me.ID, "Dire Mimic", "Artifact — Treasure", gapB210DireMimic, false)
	if err := g.ActivateCatalogAbility(me.ID, mimic, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("animated without paying {3}")
	}
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, mimic, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mimic); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
	if c, _ := effectstest.BattlefieldCard(g, mimic); !c.IsCreature() || !c.HasSubtype("Shapeshifter") || !c.IsArtifact() {
		t.Errorf("not a Shapeshifter artifact creature: %v", c.Effective().Types)
	}
}

func TestGapB210LanternOfRevealingPutsALandOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lantern := effectstest.PushCatalogPermanent(g, me.ID, "Lantern of Revealing", "Artifact", gapB210Lantern, false)
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, lantern, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if c, ok := effectstest.BattlefieldCard(g, land); !ok || !c.Tapped {
		t.Errorf("the Forest is on the battlefield %v, tapped %v; want tapped", ok, c.Tapped)
	}
}

func TestGapB210LanternOfRevealingMayBottomACardItDidNotPut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lantern := effectstest.PushCatalogPermanent(g, me.ID, "Lantern of Revealing", "Artifact", gapB210Lantern, false)
	bear := effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, lantern, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // put it on the bottom
	if me.Library.Cards[0].InstanceID != bear {
		t.Error("the Bear is not on the bottom of the library")
	}
}

func TestGapB210NavigationOrbFetchesABasicAndAGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	orb := effectstest.PushCatalogPermanent(g, me.ID, "Navigation Orb", "Artifact", gapB210Orb, false)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Basalt Gate", TypeLine: "Land — Gate"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Wastes Mine", TypeLine: "Land"},
	)
	forest, gate := ids[0], ids[1]
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, orb, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, gate)
	if c, ok := effectstest.BattlefieldCard(g, gate); !ok || !c.Tapped {
		t.Fatalf("the Gate is on the battlefield %v, tapped %v; want tapped", ok, c.Tapped)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Error("the Forest did not reach the hand")
	}
	if g.Battlefield.Contains(orb) {
		t.Error("the Orb was not sacrificed")
	}
}

func TestGapB210SarevoksTomeTakesTheInitiativeAndMakesMoreMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tome := effectstest.CastCatalogSpell(t, g, "Sarevok's Tome", "Artifact — Book", gapB210Tome, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if !effects.YouHaveTheInitiative(g, me.ID) {
			t.Error("the Tome's controller does not have the initiative")
		}
	})
	if err := g.ActivateManaAbility(me.ID, tome, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("mana %d with the initiative, want 2", me.ManaPool.Total())
	}
}

func TestGapB210SarevoksTomeNeedsACompletedDungeonToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tome := effectstest.PushCatalogPermanent(g, me.ID, "Sarevok's Tome", "Artifact — Book", gapB210Tome, false)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, tome, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated without having completed a dungeon")
	}
}

func TestGapB210PactWeaponDrawsRevealsPumpsAndCostsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	weapon := effectstest.PushCatalogPermanent(g, me.ID, "Pact Weapon", "Artifact — Equipment", gapB210PactWeapon, false)
	g.WithWriteLock(func() { _ = g.AttachForEffect(weapon, game.TargetRef{Kind: game.TargetCard, ID: bear}) })
	effectstest.SeedLibraryTop(me, "Dragon", "Creature — Dragon", "{4}{R}{R}")
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 8 {
		t.Errorf("power %d, want 2 + 6", p)
	}
	if me.Life != life-6 {
		t.Errorf("life %d, want %d", me.Life, life-6)
	}
}

func TestGapB210PactWeaponKeepsYouAliveAtZeroLifeOnlyWhileAttached(t *testing.T) {
	for _, attached := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
		weapon := effectstest.PushCatalogPermanent(g, me.ID, "Pact Weapon", "Artifact — Equipment", gapB210PactWeapon, false)
		if attached {
			g.WithWriteLock(func() { _ = g.AttachForEffect(weapon, game.TargetRef{Kind: game.TargetCard, ID: bear}) })
		}
		g.WithWriteLock(func() { me.Life = 0 })
		g.RunStateChecksForTest()
		if me.Eliminated == attached {
			t.Errorf("attached=%v: eliminated=%v at 0 life", attached, me.Eliminated)
		}
	}
}

func TestGapB210MultiClassBaldricGrantsKeywordsByPartyType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Human", 1, 1)
	baldric := effectstest.PushCatalogPermanent(g, me.ID, "Multiclass Baldric", "Artifact — Equipment", gapB210Baldric, false)
	g.WithWriteLock(func() { _ = g.AttachForEffect(baldric, game.TargetRef{Kind: game.TargetCard, ID: host}) })
	if effectstest.EffectiveAbilitiesContain(t, g, host, "haste") || effectstest.EffectiveAbilitiesContain(t, g, host, "flying") {
		t.Fatal("keywords granted with no party")
	}
	effectstest.B12Creature(g, me.ID, "Mage", "Creature — Human Wizard", 1, 1)
	effectstest.B12Creature(g, me.ID, "Fighter", "Creature — Human Warrior", 1, 1)
	has := func(kw string) bool { return effectstest.EffectiveAbilitiesContain(t, g, host, kw) }
	if !has("flying") || !has("haste") || has("lifelink") || has("deathtouch") {
		t.Errorf("keywords %v; want flying and haste only", effectstest.EffectiveAbilities(t, g, host))
	}
}

func TestGapB210MirrorOfLifeTrappingExilesACastCreatureAndFreesTheOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Mirror of Life Trapping", "Artifact", gapB210Mirror, false)
	first := effectstest.CastHoldCreature(t, g, "First Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) {
		t.Fatal("the cast creature was not exiled")
	}
	second := effectstest.CastHoldCreature(t, g, "Second Bear", "Creature — Bear", "", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(second) {
		t.Error("the second cast creature was not exiled")
	}
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "First Bear" {
			found = true
		}
	}
	if !found {
		t.Error("the first exiled creature did not return when the second was exiled")
	}
}

func TestGapB210MirrorOfLifeTrappingIgnoresACreatureThatWasNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Mirror of Life Trapping", "Artifact", gapB210Mirror, false)
	bear := effectstest.B12Creature(g, me.ID, "Put Bear", "Creature — Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature that was not cast was exiled")
	}
}

func TestGapB210FrayingLinePutsARopeCounterOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fraying Line", "Artifact", gapB210Fraying, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Counters["rope"] != 1 {
		t.Errorf("rope counters %d, want 1", c.Counters["rope"])
	}
}

func TestGapB210MightyServantCrewedByExactlyTwoDrawsTwoOnCombatDamage(t *testing.T) {
	for _, tc := range []struct {
		name    string
		crewers int
		want    int
	}{{"two crewers draw two", 2, 2}, {"three crewers draw nothing", 3, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			effectstest.AdvanceToMain(t, g)
			servant := effectstest.PushCatalogPermanent(g, me.ID, "Mighty Servant of Leuk-o", "Artifact — Vehicle", gapB210Servant, false)
			var crew []uuid.UUID
			for i := 0; i < tc.crewers; i++ {
				crew = append(crew, effectstest.PushVanillaCreature(g, me.ID, "Crew Member", 4, 4))
			}
			effectstest.CrewBy(t, g, me.ID, servant, crew...)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			hand := me.Hand.Size()
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(servant, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
			effectstest.LockInAttacks(t, g)
			effectstest.AdvanceTo(t, g, game.StepCombatDamage)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.want {
				t.Errorf("drew %d, want %d", got, tc.want)
			}
		})
	}
}
