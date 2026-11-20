package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grn_5525_b1_cards_test.go — a batch of Guilds of Ravnica cards (#5525).
// Every test plays the real catalog entry through the engine.

const (
	grnB1DouserOracle        = "1e53032e-fe14-4424-86ee-f69336f049fe"
	grnB1HalberdierOracle    = "2ef2f417-2061-4796-974a-92f800ef05e4"
	grnB1CeratokOracle       = "4dc82a49-a063-448d-bd2c-a5b69381e27e"
	grnB1CrabOracle          = "abc519b8-82de-4559-be03-ddb1ddaab993"
	grnB1BatsOracle          = "6987d5d4-0ef6-4326-a590-678ac160fe56"
	grnB1PoisonerOracle      = "a7a6cdfd-10ac-49d0-a0a6-8ad37e326511"
	grnB1RaiderOracle        = "3c741801-bfa5-4ff6-bf06-8d5edc06388d"
	grnB1OkapiOracle         = "46f01d1c-7ab6-4a58-97ce-57a4eb99843c"
	grnB1PredatorOracle      = "b94591b8-2e99-4983-8f36-9b52207bfd2e"
	grnB1LedevOracle         = "edcf36b4-dea4-4314-94da-e0ee5e30cd37"
	grnB1RosemaneOracle      = "54b3ccdc-e5b0-44d9-bc81-a296fc5dc005"
	grnB1ArboretumOracle     = "1708d60d-30af-4184-bc95-2f7a7aed4a6e"
	grnB1BarrierOracle       = "4f25b125-1ded-4a34-9ed1-e6ce087ec48d"
	grnB1InformantOracle     = "fa66d63b-dfb7-42b8-8249-4d60e89058ac"
	grnB1StrayOracle         = "5722a13f-7d70-438c-94ba-97a3f53fbc5c"
	grnB1RestorerOracle      = "77eb2374-44a0-4383-b229-edc116c7d835"
	grnB1PeacemakerOracle    = "495d95cf-dcce-4a15-a544-79aced7ef98d"
	grnB1GuardOracle         = "83ce0859-cb88-45dd-8485-3317792235de"
	grnB1BoarOracle          = "b815cbbd-d79f-4180-8204-504a15df7152"
	grnB1FindbrokerOracle    = "e5456b35-d7a9-4717-90df-c6c46f1ec437"
	grnB1UnicornOracle       = "1f9797e2-7ebb-41f9-9199-e9ee575eefd3"
	grnB1WitnessOracle       = "1fde570c-aa96-45ca-a6be-f59c8e4f4b6a"
	grnB1SphinxOracle        = "a63e0c27-cea8-408a-8e57-2f077826c50f"
	grnB1CentipedeOracle     = "43d44a7f-829a-4f4d-9f71-2f4384ddd803"
	grnB1BanneretOracle      = "a2d9554e-70c0-4213-ba5e-f0954ad32854"
	grnB1ShadeOracle         = "9e9189d9-3b25-476b-b096-bb540398c1d1"
	grnB1DissidentOracle     = "a671fedf-3766-4fda-abeb-6f63e2fdd20b"
	grnB1DartOracle          = "912bc348-9ba3-4001-90a8-104d2bb17527"
	grnB1BorosLocketOracle   = "b044b4fd-d504-4a11-80bf-eb8fb0b6e21e"
	grnB1DimirLocketOracle   = "3ae4b679-2e3c-4f60-bcf4-b279917cdc54"
	grnB1GolgariLocketOracle = "dba6e24e-7875-46b9-b51d-4cd71568ecfa"
	grnB1IzzetLocketOracle   = "7ee07546-2c2e-4679-b671-7ea509b1d3ea"
	grnB1SelesnyaLocket      = "9ecd98d5-8d66-4cb0-bada-23dfa0faacc0"
	grnB1UtopiaOracle        = "77b73f11-a59a-4f2b-83d3-e6fe5064bc87"
	grnB1StormOracle         = "1dc816b5-8918-4a0c-bb10-74abe579a98c"
	grnB1PrecisionOracle     = "5f9ea6e4-f5be-4ea2-95b4-fbdc69685ac5"
	grnB1BlazeOracle         = "5fcc057e-e6c2-4e37-a786-9a5ef94926e1"
	grnB1CollarOracle        = "adc03c37-33c7-4b92-b9fd-4608e6c372c4"
	grnB1BustOracle          = "3ac01298-e45b-4ea0-8379-2074279716a4"
	grnB1VisitOracle         = "04a3bb42-0464-4638-a6ac-d5f412c222cc"
	grnB1LightsOracle        = "394709f0-2d29-4a2d-be05-7b22fec57763"
	grnB1DisappearOracle     = "18d16129-da25-4120-a343-64ce962abba2"
	grnB1SabotageOracle      = "973fd4d4-9255-4825-85b7-503606c4e932"
	grnB1VaporsOracle        = "5731ecb5-6ddc-4452-990c-b69986367f25"
	grnB1CompanionsOracle    = "d1a9fd14-fecb-4d2a-983c-70143238275b"
	grnB1PacksFavorOracle    = "072288da-5ba0-4176-b0e6-b3e43ecab2ec"
	grnB1PauseOracle         = "a7f266b5-7258-4eaa-b447-66460ec38505"
	grnB1QuasiOracle         = "76b709ae-5765-4b7a-99c9-89eade7c8cb8"
	grnB1VelocityOracle      = "96e7e4c8-f66d-4327-aee6-85da648418f5"
	grnB1AltitudeOracle      = "70e8eba9-a106-49c5-932e-b4390edb1c98"

	grnB1BearType = "Creature — Bear"
)

// grnB1Kept is the battlefield card with the given id, or false.
func grnB1Kept(g *game.Game, id uuid.UUID) bool {
	return g.Battlefield.Contains(id)
}

// grnB1SurveilSize is how many cards the open surveil prompt for p looks
// at, or -1 when none is open.
func grnB1SurveilSize(g *game.Game, p uuid.UUID) int {
	c := surveilChoiceFor(g, p)
	if c == nil {
		return -1
	}
	return len(c.ScryCards)
}

// --- every card is in the catalog ----------------------------------------

func TestGrn5525EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		grnB1DouserOracle, grnB1HalberdierOracle, grnB1CeratokOracle, grnB1CrabOracle,
		grnB1BatsOracle, grnB1PoisonerOracle, grnB1RaiderOracle, grnB1OkapiOracle, grnB1PredatorOracle,
		grnB1LedevOracle, grnB1RosemaneOracle, grnB1ArboretumOracle, grnB1BarrierOracle,
		grnB1InformantOracle, grnB1StrayOracle, grnB1RestorerOracle, grnB1PeacemakerOracle,
		grnB1GuardOracle, grnB1BoarOracle, grnB1FindbrokerOracle, grnB1UnicornOracle,
		grnB1WitnessOracle, grnB1SphinxOracle, grnB1CentipedeOracle, grnB1BanneretOracle,
		grnB1ShadeOracle, grnB1DissidentOracle, grnB1DartOracle, grnB1BorosLocketOracle,
		grnB1DimirLocketOracle, grnB1GolgariLocketOracle, grnB1IzzetLocketOracle, grnB1SelesnyaLocket,
		grnB1UtopiaOracle, grnB1StormOracle, grnB1PrecisionOracle, grnB1BlazeOracle,
		grnB1CollarOracle, grnB1BustOracle, grnB1VisitOracle, grnB1LightsOracle,
		grnB1DisappearOracle, grnB1SabotageOracle, grnB1VaporsOracle, grnB1CompanionsOracle,
		grnB1PacksFavorOracle, grnB1PauseOracle, grnB1QuasiOracle, grnB1VelocityOracle,
		grnB1AltitudeOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestGrn5525ConvokeCardsDeclareConvoke(t *testing.T) {
	for _, oracle := range []string{
		grnB1LedevOracle, grnB1RosemaneOracle, grnB1ArboretumOracle, grnB1RestorerOracle,
		grnB1PacksFavorOracle, grnB1PauseOracle,
	} {
		spec, _ := Lookup(oracle)
		if spec.TapCost == nil || spec.TapCost.Key != "convoke" {
			t.Errorf("%s has no convoke", spec.Name)
		}
	}
}

func TestGrn5525InescapableBlazeCantBeCountered(t *testing.T) {
	if spec, _ := Lookup(grnB1BlazeOracle); !spec.CantBeCountered {
		t.Error("Inescapable Blaze can be countered")
	}
}

// --- printed keywords -----------------------------------------------------

func TestGrn5525KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Bartizan Bats", "Creature — Bat", grnB1BatsOracle, []string{"flying"}},
		{"Hired Poisoner", "Creature — Human Assassin", grnB1PoisonerOracle, []string{"deathtouch"}},
		{"Kraul Raider", "Creature — Insect Warrior", grnB1RaiderOracle, []string{"menace"}},
		{"Wary Okapi", "Creature — Antelope", grnB1OkapiOracle, []string{"vigilance"}},
		{"Nightveil Predator", "Creature — Vampire", grnB1PredatorOracle, []string{"flying", "deathtouch", "hexproof"}},
		{"Rosemane Centaur", "Creature — Centaur Soldier", grnB1RosemaneOracle, []string{"vigilance"}},
		{"Arboretum Elemental", "Creature — Elemental", grnB1ArboretumOracle, []string{"hexproof"}},
		{"Barrier of Bones", "Creature — Skeleton Wall", grnB1BarrierOracle, []string{"defender"}},
		{"Citywatch Sphinx", "Creature — Sphinx", grnB1SphinxOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// The vanilla bodies enter and stay: no rules text to run.
func TestGrn5525VanillaCreaturesResolve(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Douser of Lights", "Creature — Horror", grnB1DouserOracle},
		{"Fearless Halberdier", "Creature — Human Warrior", grnB1HalberdierOracle},
		{"Wild Ceratok", "Creature — Rhino", grnB1CeratokOracle},
		{"Wishcoin Crab", "Creature — Crab", grnB1CrabOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 3, 3)
			passPriorityAroundTable(t, g)
			if !grnB1Kept(g, id) {
				t.Errorf("%s did not resolve", tc.name)
			}
		})
	}
}

// --- convoke ----------------------------------------------------------------

func TestGrn5525LedevGuardianIsPaidByConvoke(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tappers := pushTapCostSoldiers(g, me.ID, 4)
	if err := momCastErr(g, me, "Ledev Guardian", "Creature — Human Knight", "{3}{W}", grnB1LedevOracle,
		game.CastSpellParams{TapIDs: tappers[:3]}); err == nil {
		t.Fatal("three creatures paid for {3}{W}")
	}
	guardian := momCast(t, g, me, "Ledev Guardian", "Creature — Human Knight", "{3}{W}", grnB1LedevOracle,
		game.CastSpellParams{TapIDs: tappers})
	for _, s := range tappers {
		if !tapCostTapped(g, s) {
			t.Error("a convoking creature is untapped")
		}
	}
	passPriorityAroundTable(t, g)
	if !grnB1Kept(g, guardian) {
		t.Error("Ledev Guardian did not enter")
	}
}

func TestGrn5525PacksFavorIsConvokedAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	var elves []uuid.UUID
	for i := 0; i < 3; i++ {
		elves = append(elves, pushTapCostPermanent(g, me.ID, "Elf", "Creature — Elf", "G"))
	}
	bear := b12Creature(g, me.ID, "Bear", grnB1BearType, 2, 2)
	momCast(t, g, me, "Pack's Favor", "Instant", "{2}{G}", grnB1PacksFavorOracle,
		game.CastSpellParams{TapIDs: elves, Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
}

func TestGrn5525PauseForReflectionPreventsCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Pause for Reflection", "Instant", grnB1PauseOracle, nil)
	passPriorityAroundTable(t, g)
	attacker := pushVanillaCreature(g, opp.ID, "Attacker", 2, 2)
	defender := pushVanillaCreature(g, me.ID, "Defender", 2, 2)
	if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := findCardAnywhere(t, g, defender).DamageMarked; got != 0 {
		t.Errorf("combat damage marked %d, want 0", got)
	}
}

// --- enter triggers -----------------------------------------------------------

func TestGrn5525GenerousStrayDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castHoldCreature(t, g, "Generous Stray", "Creature — Cat", grnB1StrayOracle, 1, 2)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestGrn5525EntersAndSurveils(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Barrier of Bones", "Creature — Skeleton Wall", grnB1BarrierOracle, 1},
		{"Dimir Informant", "Creature — Human Rogue", grnB1InformantOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibrary(me, "A", "B", "C")
			castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 3)
			passPriorityAroundTable(t, g)
			if got := grnB1SurveilSize(g, me.ID); got != tc.n {
				t.Errorf("surveil looks at %d cards, want %d", got, tc.n)
			}
		})
	}
}

func TestGrn5525LoxodonRestorerGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castHoldCreature(t, g, "Loxodon Restorer", "Creature — Elephant Cleric", grnB1RestorerOracle, 3, 4)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestGrn5525CentaurPeacemakerGainsEachPlayerFour(t *testing.T) {
	g := newCatalogGame(t)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	castHoldCreature(t, g, "Centaur Peacemaker", "Creature — Centaur Cleric", grnB1PeacemakerOracle, 3, 3)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if s.Life != lives[s.ID]+4 {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, lives[s.ID]+4)
		}
	}
}

func TestGrn5525EnterPumpsTargetCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Tenth District Guard", "Creature — Human Soldier", grnB1GuardOracle, 2, 3},
		{"Rubblebelt Boar", "Creature — Boar", grnB1BoarOracle, 4, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			theirs := b12Creature(g, opp.ID, "Their Bear", grnB1BearType, 2, 2)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			b04WaitForPick(t, g, me.ID)
			pickCard(t, g, me.ID, theirs)
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, theirs); p != tc.p || tough != tc.tough {
				t.Errorf("the target is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
		})
	}
}

func TestGrn5525GolgariFindbrokerReturnsAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sorcery := pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	rock := pushGraveyardPermanent(me, "Dead Rock", "Artifact", "{1}")
	castHoldCreature(t, g, "Golgari Findbroker", "Creature — Elf Shaman", grnB1FindbrokerOracle, 3, 4)
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: sorcery}); err == nil {
		t.Fatal("a sorcery card was accepted")
	}
	pickTriggerTarget(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || me.Graveyard.Contains(rock) {
		t.Error("the artifact card did not return to your hand")
	}
}

// --- attack and dies triggers ---------------------------------------------------

func TestGrn5525InspiringUnicornPumpsYourTeamWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unicorn := b12Push(g, me.ID, "Inspiring Unicorn", "Creature — Unicorn", grnB1UnicornOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", grnB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", grnB1BearType, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(unicorn, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{unicorn: {3, 3}, mine: {3, 3}, theirs: {2, 2}} {
		if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%s is %d/%d after the attack trigger, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
}

func TestGrn5525HuntedWitnessLeavesALifelinkSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	witness := b12Push(g, me.ID, "Hunted Witness", "Creature — Human", grnB1WitnessOracle, 1, 1)
	killCreature(t, g, me.ID, witness)
	passPriorityAroundTable(t, g)
	soldiers := tokensNamed(g, "Soldier")
	if len(soldiers) != 1 {
		t.Fatalf("%d Soldier tokens, want 1", len(soldiers))
	}
	if !hasAbility(soldiers[0].Keywords, "lifelink") || soldiers[0].Controller != me.ID {
		t.Errorf("the Soldier is %+v, want a lifelink token under your control", soldiers[0])
	}
}

func TestGrn5525CitywatchSphinxSurveilsTwoWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	sphinx := b12Push(g, me.ID, "Citywatch Sphinx", "Creature — Sphinx", grnB1SphinxOracle, 5, 4)
	killCreature(t, g, me.ID, sphinx)
	passPriorityAroundTable(t, g)
	if got := grnB1SurveilSize(g, me.ID); got != 2 {
		t.Errorf("surveil looks at %d cards, want 2", got)
	}
}

func TestGrn5525SpinalCentipedeCountersYourCreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", grnB1BearType, 2, 2)
	centipede := b12Push(g, me.ID, "Spinal Centipede", "Creature — Insect", grnB1CentipedeOracle, 3, 2)
	killCreature(t, g, me.ID, centipede)
	eoeABPickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
}

// --- activated abilities --------------------------------------------------------

func TestGrn5525SelfPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		mana                   []string
		wantP, wantT           int
	}{
		{"Goblin Banneret", "Creature — Goblin Soldier", grnB1BanneretOracle, 1, 1, []string{"R", "R"}, 3, 1},
		{"Veiled Shade", "Creature — Shade", grnB1ShadeOracle, 2, 2, []string{"B", "B"}, 3, 3},
		{"Devkarin Dissident", "Creature — Elf Warrior", grnB1DissidentOracle, 2, 2, []string{"G", "G", "G", "G", "G"}, 4, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestGrn5525GoblinBanneretHasMentor(t *testing.T) {
	spec, _ := Lookup(grnB1BanneretOracle)
	if len(spec.Triggered) != 1 {
		t.Fatalf("Goblin Banneret has %d triggers, want mentor", len(spec.Triggered))
	}
}

func TestGrn5525SilentDartShootsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	dart := pushCatalogPermanent(g, me.ID, "Silent Dart", "Artifact", grnB1DartOracle, false)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 5)
	apaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, dart, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(ogre)}); err == nil {
		t.Fatal("three mana paid {4}")
	}
	apaMana(me, "C")
	apaActivate(t, g, me, dart, 0, game.ActivateAbilityParams{Targets: cardTarget(ogre)})
	if grnB1Kept(g, dart) {
		t.Error("the Dart was not sacrificed")
	}
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 3 {
		t.Errorf("the Ogre has %d damage, want 3", got)
	}
}

func TestGrn5525LocketsMakeEitherColorAndSacrificeToDraw(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		a, b         string
	}{
		{"Boros Locket", grnB1BorosLocketOracle, "R", "W"},
		{"Dimir Locket", grnB1DimirLocketOracle, "U", "B"},
		{"Golgari Locket", grnB1GolgariLocketOracle, "B", "G"},
		{"Izzet Locket", grnB1IzzetLocketOracle, "U", "R"},
		{"Selesnya Locket", grnB1SelesnyaLocket, "G", "W"},
	} {
		t.Run(tc.name+" mana", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			locket := pushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, locket, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("tap for mana: %v", err)
			}
			if pick := riderLatestManaPick(g, me.ID); pick != nil {
				if err := g.ResolveManaChoice(pick.ID, me.ID, tc.b); err != nil {
					t.Fatalf("ResolveManaChoice: %v", err)
				}
			}
			if got := me.ManaPool.Total(); got != 1 {
				t.Errorf("pool = %d, want 1", got)
			}
		})
		t.Run(tc.name+" draw", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			locket := pushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
			apaMana(me, tc.a, tc.b, tc.a)
			if err := g.ActivateCatalogAbility(me.ID, locket, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("three mana paid for four hybrid symbols")
			}
			apaMana(me, tc.b)
			hand := me.Hand.Size()
			if err := g.ActivateCatalogAbility(me.ID, locket, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if grnB1Kept(g, locket) {
				t.Error("the Locket was not sacrificed as a cost")
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != 2 {
				t.Errorf("drew %d, want 2", got)
			}
		})
	}
}

// --- the Aura ---------------------------------------------------------------------

func TestGrn5525UrbanUtopiaDrawsAndGrantsAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Urban Utopia", "Enchantment — Aura", grnB1UtopiaOracle, cardTarget(land))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size = %d, want %d (aura cast, one card drawn)", got, before+1)
	}
	idx, ref, ok := grantedManaAbilityFor(g, land)
	if !ok {
		t.Fatal("the enchanted land has no granted mana ability")
	}
	if err := g.ActivateManaAbility(me.ID, land, idx, game.ManaAbilityParams{Ref: ref, Colors: []string{"B"}}); err != nil {
		t.Fatalf("activate granted ability: %v", err)
	}
}

// --- damage and removal ---------------------------------------------------------

func TestGrn5525CommandTheStormDealsFiveToACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	wurm := b12Creature(g, opp.ID, "Wurm", "Creature — Wurm", 6, 6)
	if err := castCatalogSpellErr(t, g, "Command the Storm", "Instant", grnB1StormOracle, []game.TargetRef{playerTarget(opp.ID)}); err == nil {
		t.Fatal("Command the Storm targeted a player")
	}
	castCatalogSpell(t, g, "Command the Storm", "Instant", grnB1StormOracle, cardTarget(wurm))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, wurm).DamageMarked; got != 5 {
		t.Errorf("the Wurm has %d damage, want 5", got)
	}
}

func TestGrn5525BurnToAPlayer(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		n                      int
	}{
		{"Precision Bolt", "Sorcery", grnB1PrecisionOracle, 3},
		{"Inescapable Blaze", "Instant", grnB1BlazeOracle, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			opp := top100Opponent(g)
			life := opp.Life
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, []game.TargetRef{playerTarget(opp.ID)})
			passPriorityAroundTable(t, g)
			if opp.Life != life-tc.n {
				t.Errorf("the player is at %d, want %d", opp.Life, life-tc.n)
			}
		})
	}
}

func TestGrn5525CollarTheCulpritNeedsToughnessFour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	small := b12Creature(g, opp.ID, "Small", grnB1BearType, 2, 3)
	big := b12Creature(g, opp.ID, "Big", grnB1BearType, 2, 4)
	if err := castCatalogSpellErr(t, g, "Collar the Culprit", "Instant", grnB1CollarOracle, cardTarget(small)); err == nil {
		t.Fatal("Collar the Culprit targeted a creature with toughness 3")
	}
	castCatalogSpell(t, g, "Collar the Culprit", "Instant", grnB1CollarOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if grnB1Kept(g, big) {
		t.Error("the toughness 4 creature survived")
	}
}

func TestGrn5525CitywideBustDestroysOnlyToughnessFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	big := b12Creature(g, opp.ID, "Big", grnB1BearType, 1, 4)
	mine := b12Creature(g, me.ID, "Mine", grnB1BearType, 5, 5)
	small := b12Creature(g, opp.ID, "Small", grnB1BearType, 4, 3)
	castCatalogSpell(t, g, "Citywide Bust", "Sorcery", grnB1BustOracle, nil)
	passPriorityAroundTable(t, g)
	if grnB1Kept(g, big) || grnB1Kept(g, mine) {
		t.Error("a creature with toughness 4 or greater survived")
	}
	if !grnB1Kept(g, small) {
		t.Error("a creature with toughness 3 was destroyed")
	}
}

func TestGrn5525DeadlyVisitDestroysThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	seedLibrary(me, "A", "B", "C")
	bear := b12Creature(g, opp.ID, "Bear", grnB1BearType, 2, 2)
	castCatalogSpell(t, g, "Deadly Visit", "Sorcery", grnB1VisitOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if grnB1Kept(g, bear) {
		t.Error("the creature survived")
	}
	if got := grnB1SurveilSize(g, me.ID); got != 2 {
		t.Errorf("surveil looks at %d cards, want 2", got)
	}
}

func TestGrn5525DazzlingLightsShrinksThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	seedLibrary(me, "A", "B", "C")
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Dazzling Lights", "Instant", grnB1LightsOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, ogre); p != 1 || tough != 4 {
		t.Errorf("the Ogre is %d/%d, want 1/4", p, tough)
	}
	if got := grnB1SurveilSize(g, me.ID); got != 2 {
		t.Errorf("surveil looks at %d cards, want 2", got)
	}
}

func TestGrn5525UnexplainedDisappearanceBouncesThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	seedLibrary(me, "A", "B", "C")
	bear := b12Creature(g, opp.ID, "Bear", grnB1BearType, 2, 2)
	castCatalogSpell(t, g, "Unexplained Disappearance", "Instant", grnB1DisappearOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if grnB1Kept(g, bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
	if got := grnB1SurveilSize(g, me.ID); got != 1 {
		t.Errorf("surveil looks at %d cards, want 1", got)
	}
}

func TestGrn5525SinisterSabotageCountersThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C")
	bolt := b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := b13HandCard(me, "Sinister Sabotage", "Instant", grnB1SabotageOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the spell was not countered")
	}
	if got := grnB1SurveilSize(g, me.ID); got != 1 {
		t.Errorf("surveil looks at %d cards, want 1", got)
	}
}

func TestGrn5525MephiticVaporsShrinksEveryCreatureThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	seedLibrary(me, "A", "B", "C")
	mine := b12Creature(g, me.ID, "Mine", grnB1BearType, 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", grnB1BearType, 2, 2)
	castCatalogSpell(t, g, "Mephitic Vapors", "Sorcery", grnB1VaporsOracle, nil)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{mine: {2, 2}, theirs: {1, 1}} {
		if p, tough := ptOf(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", id, p, tough, want[0], want[1])
		}
	}
	if got := grnB1SurveilSize(g, me.ID); got != 2 {
		t.Errorf("surveil looks at %d cards, want 2", got)
	}
}

func TestGrn5525SwornCompanionsMakesTwoLifelinkSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Sworn Companions", "Sorcery", grnB1CompanionsOracle, nil)
	passPriorityAroundTable(t, g)
	soldiers := tokensNamed(g, "Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("%d Soldier tokens, want 2", len(soldiers))
	}
	for _, s := range soldiers {
		if !hasAbility(s.Keywords, "lifelink") {
			t.Error("a Soldier has no lifelink")
		}
	}
}

// --- jump-start ---------------------------------------------------------------------

func TestGrn5525QuasiduplicateCopiesYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Grizzly", grnB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Grizzly", grnB1BearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "Quasiduplicate", "Sorcery", grnB1QuasiOracle, cardTarget(theirs)); err == nil {
		t.Fatal("Quasiduplicate targeted a creature you don't control")
	}
	castCatalogSpell(t, g, "Quasiduplicate", "Sorcery", grnB1QuasiOracle, cardTarget(mine))
	passPriorityAroundTable(t, g)
	if got := tokensNamed(g, "Grizzly"); len(got) != 1 || got[0].Controller != me.ID {
		t.Errorf("tokens named Grizzly: %d, want one under your control", len(got))
	}
}

func TestGrn5525MaximizeSpellsPumpAndGrant(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Maximize Velocity", grnB1VelocityOracle, "haste"},
		{"Maximize Altitude", grnB1AltitudeOracle, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			bear := b12Creature(g, me.ID, "Bear", grnB1BearType, 2, 2)
			castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
				t.Errorf("the Bear is %d/%d, want 3/3", p, tough)
			}
			if !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}

func TestGrn5525JumpStartCastsFromTheGraveyard(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Maximize Velocity", grnB1VelocityOracle},
		{"Maximize Altitude", grnB1AltitudeOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := seedGraveyardCard(t, g, tc.name, "Sorcery", tc.oracle)
			bear := b12Creature(g, me.ID, "Bear", grnB1BearType, 2, 2)
			fodder := handCard(me, "Fodder", "Sorcery")
			s39Fund(g, me)
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
				AlternativeCost: "jump-start", Targets: cardTarget(bear)}); err == nil {
				t.Fatal("jump-start was castable without discarding a card")
			}
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
				AlternativeCost: "jump-start", AltCostIDs: []uuid.UUID{fodder}, Targets: cardTarget(bear)}); err != nil {
				t.Fatalf("jump-start cast: %v", err)
			}
			if !me.Graveyard.Contains(fodder) {
				t.Error("the discard is a cost, paid as the spell is cast")
			}
			passPriorityAroundTable(t, g)
			if p, _ := ptOf(t, g, bear); p != 3 {
				t.Errorf("the Bear has power %d, want 3", p)
			}
			if !g.Exile.Contains(id) {
				t.Error("the jump-started sorcery is not exiled")
			}
		})
	}
}
