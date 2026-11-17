package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// leg_b2_cards_test.go — Legends, batch 2 (#5639). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	legB2BarbaryApesOracle       = "39bfe194-1662-4c0e-a7c6-ea7764d29b65"
	legB2BarktoothOracle         = "4a221518-c850-4e8a-a697-765c0a0e76e7"
	legB2CrimsonKoboldsOracle    = "93fdcc49-e3c8-4f8c-a3bc-064c6392b4ca"
	legB2CrookshankKoboldsOracle = "61023b48-0433-4993-9534-308c08cef07e"
	legB2DurkwoodBoarsOracle     = "8bb2ac6c-20aa-46dd-883f-b629855cabb0"
	legB2HeadlessHorsemanOracle  = "5bb4961f-0a24-40f8-bfba-8618a9375c93"
	legB2JasmineBorealOracle     = "942407bf-862a-4a0b-b19d-b9f8ee833026"
	legB2JeditOjanenOracle       = "523cca7d-7d21-45de-b82d-58e777bf7f72"
	legB2JerrardOracle           = "fd20576d-4abe-4af8-ac53-ab9356593fe7"
	legB2KasimirOracle           = "64b1a2a2-d3ff-41c6-8d92-987ef29bb776"
	legB2KeepersOfTheFaithOracle = "6ddc1256-a166-464f-8136-bd111cba6327"
	legB2KoboldsOfKherKeepOracle = "badf0669-9f65-44e3-bf91-6c57ac2fecff"
	legB2LadyOrcaOracle          = "fcae9e87-50af-46b4-ab48-42ed92985e10"
	legB2MossMonsterOracle       = "60efa8a2-6903-4756-92fd-29f6ba6ef67a"
	legB2RagingBullOracle        = "341680e2-12a8-484f-be09-24dae9dbfaea"
	legB2SirShandlarOracle       = "7a9942cd-32ac-4cc9-8840-3fe6b092d0b8"
	legB2SivitriScarzamOracle    = "333cc91b-01ab-4491-a993-d8304fc382c6"
	legB2LadyOfTheMountainOracle = "3ef03975-0f46-43aa-b461-bd3b6f5d4613"
	legB2TobiasAndrionOracle     = "55dfe722-7487-4d25-b90c-e6ad2c82a671"
	legB2TorstenVonUrsusOracle   = "633b8f22-0384-4f60-ba78-b52c2d82dfde"

	legB2AzureDrakeOracle        = "15cc068c-f424-433e-b165-8457c62a4b35"
	legB2ThunderSpiritOracle     = "6efad3d6-4886-4b62-a4ff-bc0622f9672a"
	legB2ZephyrFalconOracle      = "b5e44bd6-0b67-4770-868d-c46af5b6a42f"
	legB2TundraWolvesOracle      = "2155864e-5788-4e72-a800-e2cf25cf59a7"
	legB2HornetCobraOracle       = "4958baf3-2f5a-4087-bc14-ffee9887db1d"
	legB2RamirezOracle           = "48b9c510-03ed-41cb-8c17-16110d447490"
	legB2LostSoulOracle          = "1c68e877-ee47-42bf-a867-3f7f479eae1c"
	legB2CatWarriorsOracle       = "cee3d8e5-f26f-4f47-8a7b-9ac07ebf9025"
	legB2DevouringDeepOracle     = "e0f18302-5ccf-47d2-a8a1-bb7aeaca9a53"
	legB2SegovianLeviathanOracle = "853b02e1-fc0a-413b-8381-88596624440f"
	legB2RighteousAvengersOracle = "e7c07e41-3654-47da-afe3-88aae4c034b6"
	legB2MountainYetiOracle      = "375ea10b-da68-4c66-b3eb-f5113c09ea4b"
	legB2WallOfEarthOracle       = "1e3e1916-bb82-481e-9b10-225504ef06b3"
	legB2WallOfHeatOracle        = "c7bbfd32-2a6f-44e9-b51e-9b272f48d458"
	legB2WallOfLightOracle       = "d0c05cad-6e14-4726-b7fc-8f2fc7df7267"

	legB2BoomerangOracle    = "dc4a4996-108a-4aac-850f-2d9f76403446"
	legB2RemoveSoulOracle   = "b13c0f76-fbda-4911-9442-c3d7e97f1aac"
	legB2FlashCounterOracle = "75b2dd64-a2f5-4032-a19f-5b485dc44b16"
	legB2ForceSpikeOracle   = "896ea218-3019-47e4-b0a0-2487335872bb"

	legB2CarrionAntsOracle      = "87bc6be7-4734-4a11-806e-e9e4f5616fe0"
	legB2KillerBeesOracle       = "5da1f1af-d2e5-4e14-914b-d93f15626636"
	legB2PavelMalikiOracle      = "41026872-3c7b-40c7-8d0b-9d2f9b6c3e91"
	legB2WallOfOppositionOracle = "30254aa2-2307-4f42-926a-54a416f55f87"
)

// legB2Vanilla is every vanilla creature in the batch, with its printed
// type line and body.
var legB2Vanilla = []struct {
	name, typeLine, oracle string
	power, toughness       int
}{
	{"Barbary Apes", "Creature — Ape", legB2BarbaryApesOracle, 2, 2},
	{"Barktooth Warbeard", "Legendary Creature — Human Warrior", legB2BarktoothOracle, 6, 5},
	{"Crimson Kobolds", "Creature — Kobold", legB2CrimsonKoboldsOracle, 0, 1},
	{"Crookshank Kobolds", "Creature — Kobold", legB2CrookshankKoboldsOracle, 0, 1},
	{"Durkwood Boars", "Creature — Boar", legB2DurkwoodBoarsOracle, 4, 4},
	{"Headless Horseman", "Creature — Zombie Knight", legB2HeadlessHorsemanOracle, 2, 2},
	{"Jasmine Boreal", "Legendary Creature — Human", legB2JasmineBorealOracle, 4, 5},
	{"Jedit Ojanen", "Legendary Creature — Cat Warrior", legB2JeditOjanenOracle, 5, 5},
	{"Jerrard of the Closed Fist", "Legendary Creature — Human Knight", legB2JerrardOracle, 6, 5},
	{"Kasimir the Lone Wolf", "Legendary Creature — Human Warrior", legB2KasimirOracle, 5, 3},
	{"Keepers of the Faith", "Creature — Human Cleric", legB2KeepersOfTheFaithOracle, 2, 3},
	{"Kobolds of Kher Keep", "Creature — Kobold", legB2KoboldsOfKherKeepOracle, 0, 1},
	{"Lady Orca", "Legendary Creature — Demon", legB2LadyOrcaOracle, 7, 4},
	{"Moss Monster", "Creature — Elemental", legB2MossMonsterOracle, 3, 6},
	{"Raging Bull", "Creature — Ox", legB2RagingBullOracle, 2, 2},
	{"Sir Shandlar of Eberyn", "Legendary Creature — Human Knight", legB2SirShandlarOracle, 4, 7},
	{"Sivitri Scarzam", "Legendary Creature — Human", legB2SivitriScarzamOracle, 6, 4},
	{"The Lady of the Mountain", "Legendary Creature — Giant", legB2LadyOfTheMountainOracle, 5, 5},
	{"Tobias Andrion", "Legendary Creature — Human Advisor", legB2TobiasAndrionOracle, 4, 4},
	{"Torsten Von Ursus", "Legendary Creature — Human Soldier", legB2TorstenVonUrsusOracle, 5, 5},
}

// legB2Keywords is every creature whose whole text is engine keywords.
var legB2Keywords = []struct {
	name, typeLine, oracle string
	keywords               []string
}{
	{"Azure Drake", "Creature — Drake", legB2AzureDrakeOracle, []string{"flying"}},
	{"Thunder Spirit", "Creature — Elemental Spirit", legB2ThunderSpiritOracle, []string{"flying", "first strike"}},
	{"Zephyr Falcon", "Creature — Bird", legB2ZephyrFalconOracle, []string{"flying", "vigilance"}},
	{"Tundra Wolves", "Creature — Wolf", legB2TundraWolvesOracle, []string{"first strike"}},
	{"Hornet Cobra", "Creature — Snake", legB2HornetCobraOracle, []string{"first strike"}},
	{"Ramirez DePietro", "Legendary Creature — Human Pirate", legB2RamirezOracle, []string{"first strike"}},
	{"Lost Soul", "Creature — Spirit Minion", legB2LostSoulOracle, []string{"swampwalk"}},
	{"Cat Warriors", "Creature — Cat Warrior", legB2CatWarriorsOracle, []string{"forestwalk"}},
	{"Devouring Deep", "Creature — Fish", legB2DevouringDeepOracle, []string{"islandwalk"}},
	{"Segovian Leviathan", "Creature — Leviathan", legB2SegovianLeviathanOracle, []string{"islandwalk"}},
	{"Righteous Avengers", "Creature — Human Soldier", legB2RighteousAvengersOracle, []string{"plainswalk"}},
	{"Mountain Yeti", "Creature — Yeti", legB2MountainYetiOracle, []string{"mountainwalk", "protection from white"}},
	{"Wall of Earth", "Creature — Wall", legB2WallOfEarthOracle, []string{"defender"}},
	{"Wall of Heat", "Creature — Wall", legB2WallOfHeatOracle, []string{"defender"}},
	{"Wall of Light", "Creature — Wall", legB2WallOfLightOracle, []string{"defender", "protection from black"}},
	{"Killer Bees", "Creature — Insect", legB2KillerBeesOracle, []string{"flying"}},
	{"Wall of Opposition", "Creature — Wall", legB2WallOfOppositionOracle, []string{"defender"}},
}

func TestLegB2EveryCardIsRegisteredFull(t *testing.T) {
	names := map[string]string{
		legB2BoomerangOracle:    "Boomerang",
		legB2RemoveSoulOracle:   "Remove Soul",
		legB2FlashCounterOracle: "Flash Counter",
		legB2ForceSpikeOracle:   "Force Spike",
		legB2CarrionAntsOracle:  "Carrion Ants",
		legB2PavelMalikiOracle:  "Pavel Maliki",
	}
	for _, c := range legB2Vanilla {
		names[c.oracle] = c.name
	}
	for _, c := range legB2Keywords {
		names[c.oracle] = c.name
	}
	if len(names) != 43 {
		t.Fatalf("%d cards in the batch, want 43", len(names))
	}
	for oracle, name := range names {
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

// --- the vanilla creatures ------------------------------------------------------

// Each is cast from hand, resolves onto the battlefield with its printed
// body, and declares no rules of its own.
func TestLegB2VanillaCreaturesResolveAsPrinted(t *testing.T) {
	for _, tc := range legB2Vanilla {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := Lookup(tc.oracle)
			if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 ||
				len(spec.Triggered) != 0 || len(spec.ManaAbilities) != 0 || spec.OnResolve != nil {
				t.Fatalf("%s declares rules: %+v", tc.name, spec)
			}
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			id := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
				Power: tc.power, Toughness: tc.toughness, Owner: me.ID, Controller: me.ID,
			})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast %s: %v", tc.name, err)
			}
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve onto the battlefield", tc.name)
			}
			if p, tough := ptOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.toughness)
			}
		})
	}
}

// --- the keyword creatures ------------------------------------------------------

func TestLegB2KeywordCreatures(t *testing.T) {
	for _, tc := range legB2Keywords {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := blbB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !hasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Each landwalker can't be blocked while the defending player controls a
// land of its type, and can be when they don't (CR 702.14c).
func TestLegB2LandwalkersAreUnblockableAgainstTheirLand(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, land string
	}{
		{"Lost Soul", legB2LostSoulOracle, "Swamp"},
		{"Cat Warriors", legB2CatWarriorsOracle, "Forest"},
		{"Devouring Deep", legB2DevouringDeepOracle, "Island"},
		{"Segovian Leviathan", legB2SegovianLeviathanOracle, "Island"},
		{"Righteous Avengers", legB2RighteousAvengersOracle, "Plains"},
		{"Mountain Yeti", legB2MountainYetiOracle, "Mountain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, withLand := range []bool{true, false} {
				g := newCatalogGame(t)
				me, them := seatsForCombat(g)
				walker := blbB1Push(g, me.ID, tc.name, "Creature — Test", tc.oracle, 2, 2)
				bear := blbB1Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
				if withLand {
					seedLand(g, them.ID, tc.land, "Basic Land — "+tc.land, "")
				}
				attackIntoBlocks(t, g, them.ID, walker)
				got := blockRefusal(t, g, bear, walker)
				if withLand && got != game.BlockReasonLandwalk {
					t.Errorf("block against a %s: reason %q, want landwalk", tc.land, got)
				}
				if !withLand && got != "" {
					t.Errorf("block with no %s: refused (%q)", tc.land, got)
				}
			}
		})
	}
}

// A flyer can't be blocked by a creature without flying or reach
// (CR 702.9b).
func TestLegB2FlyersRefuseAGroundBlocker(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Azure Drake", legB2AzureDrakeOracle},
		{"Thunder Spirit", legB2ThunderSpiritOracle},
		{"Zephyr Falcon", legB2ZephyrFalconOracle},
		{"Killer Bees", legB2KillerBeesOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			flyer := blbB1Push(g, me.ID, tc.name, "Creature — Test", tc.oracle, 2, 2)
			bear := blbB1Push(g, them.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
			attackIntoBlocks(t, g, them.ID, flyer)
			if got := blockRefusal(t, g, bear, flyer); got != game.BlockReasonFlying {
				t.Errorf("a ground blocker: reason %q, want flying", got)
			}
		})
	}
}

// Zephyr Falcon attacks without tapping (CR 702.20b).
func TestLegB2ZephyrFalconAttacksUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	falcon := blbB1Push(g, me.ID, "Zephyr Falcon", "Creature — Bird", legB2ZephyrFalconOracle, 1, 1)
	attackIntoBlocks(t, g, them.ID, falcon)
	if alliTapped(t, g, falcon) {
		t.Error("the vigilant Falcon tapped to attack")
	}
}

// A Wall with defender can't be declared as an attacker (CR 702.3b).
func TestLegB2WallsCantAttack(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Wall of Earth", legB2WallOfEarthOracle},
		{"Wall of Heat", legB2WallOfHeatOracle},
		{"Wall of Light", legB2WallOfLightOracle},
		{"Wall of Opposition", legB2WallOfOppositionOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			wall := blbB1Push(g, me.ID, tc.name, "Creature — Wall", tc.oracle, 0, 6)
			advanceToStepInTurn(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(wall, them.ID); err == nil {
				t.Errorf("%s was declared as an attacker", tc.name)
			}
		})
	}
}

// legB2TargetRefused reports whether a removal spell of `color`, cast by
// the active seat, is refused when it targets `id`.
func legB2TargetRefused(t *testing.T, g *game.Game, id uuid.UUID, name, oracle, cost, color string) bool {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: name, TypeLine: "Instant", OracleID: oracle,
		ManaCost: cost, Colors: []string{color}, Owner: me.ID, Controller: me.ID,
	})
	for range 3 {
		me.ManaPool.AddMana(game.ManaToken{Color: color})
	}
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: cardTarget(id)})
	me.ManaPool.EmptyPool()
	return err != nil
}

// Protection from black refuses a black spell's target and protection
// from white refuses a white creature's block (CR 702.16b, 702.16f);
// neither stops the other colour.
func TestLegB2ProtectionFromAColor(t *testing.T) {
	t.Run("Wall of Light refuses Murder", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		wall := blbB1Push(g, opp.ID, "Wall of Light", "Creature — Wall", legB2WallOfLightOracle, 1, 5)
		if !legB2TargetRefused(t, g, wall, "Murder", b07MurderOracle, "{1}{B}{B}", "B") {
			t.Error("a black Murder targeted Wall of Light")
		}
		if legB2TargetRefused(t, g, wall, "Lightning Bolt", lightningBoltOracle, "{R}", "R") {
			t.Error("a red Lightning Bolt could not target Wall of Light")
		}
	})
	t.Run("Mountain Yeti can't be blocked by a white creature", func(t *testing.T) {
		for _, color := range []string{"W", "G"} {
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			yeti := blbB1Push(g, me.ID, "Mountain Yeti", "Creature — Yeti", legB2MountainYetiOracle, 3, 3)
			blocker := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Knight", TypeLine: "Creature — Knight",
				Colors: []string{color}, Power: 2, Toughness: 2, Owner: them.ID, Controller: them.ID,
			})
			attackIntoBlocks(t, g, them.ID, yeti)
			got := blockRefusal(t, g, blocker, yeti)
			if color == "W" && got != game.BlockReasonProtection {
				t.Errorf("a white blocker: reason %q, want protection", got)
			}
			if color == "G" && got != "" {
				t.Errorf("a green blocker: refused (%q)", got)
			}
		}
	})
}

// --- the spells -------------------------------------------------------------------

func TestLegB2BoomerangReturnsAPermanentToItsOwner(t *testing.T) {
	for _, typeLine := range []string{"Land", "Enchantment", "Creature — Bear"} {
		g := newCatalogGame(t)
		_, opp, _ := blbB1Seats(g)
		target := b12Permanent(g, opp.ID, "Target", typeLine)
		castCatalogSpell(t, g, "Boomerang", "Instant", legB2BoomerangOracle, cardTarget(target))
		passPriorityAroundTable(t, g)
		if !opp.Hand.Contains(target) || g.Battlefield.Contains(target) {
			t.Errorf("the %s did not return to its owner's hand", typeLine)
		}
	}
}

// legB2CounterSetup casts the active seat's spell of `typeLine`, then
// passes priority to the next seat and returns it with the spell's id.
func legB2CounterSetup(t *testing.T, g *game.Game, typeLine string) (*game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	me, opp, _ := blbB1Seats(g)
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Spell", typeLine, "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	return me, opp, victim
}

func TestLegB2RemoveSoulAndFlashCounterCounterOnlyTheirKind(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, right, wrong string
	}{
		{"Remove Soul", legB2RemoveSoulOracle, "Creature — Bear", "Sorcery"},
		{"Flash Counter", legB2FlashCounterOracle, "Instant", "Creature — Bear"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			_, opp, wrong := legB2CounterSetup(t, g, tc.wrong)
			bad := handCardFull(opp, tc.name, "Instant", "", tc.oracle, []string{"U"})
			if err := g.CastSpell(opp.ID, bad, game.CastSpellParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wrong}},
			}); err == nil {
				t.Fatalf("%s targeted a %s spell", tc.name, tc.wrong)
			}

			g = newCatalogGame(t)
			me, opp, victim := legB2CounterSetup(t, g, tc.right)
			castFromSeat(t, g, opp, tc.name, "Instant", tc.oracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			passPriorityAroundTable(t, g)
			if g.Stack.Contains(victim) || g.Battlefield.Contains(victim) || !me.Graveyard.Contains(victim) {
				t.Errorf("the %s spell was not countered into its owner's graveyard", tc.right)
			}
		})
	}
}

func TestLegB2ForceSpikeCountersUnlessItsControllerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, victim := legB2CounterSetup(t, g, "Sorcery")
		castFromSeat(t, g, opp, "Force Spike", "Instant", legB2ForceSpikeOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		passUntilTaxed(t, g, me.ID)
		if !g.Stack.Contains(victim) {
			t.Fatal("the targeted spell left the stack while the {1} was unanswered")
		}
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		answerPayUnless(t, g, me.ID, pay)
		countered := !g.Stack.Contains(victim) && me.Graveyard.Contains(victim)
		if countered == pay {
			t.Errorf("paid %v: countered %v", pay, countered)
		}
	}
}

// --- the pumps ----------------------------------------------------------------------

func TestLegB2SelfPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		mana                   []string // one activation's cost
		wantP, wantT           int
	}{
		{"Carrion Ants", "Creature — Insect", legB2CarrionAntsOracle, 0, 1, []string{"C"}, 2, 3},
		{"Killer Bees", "Creature — Insect", legB2KillerBeesOracle, 0, 1, []string{"G"}, 2, 3},
		{"Pavel Maliki", "Legendary Creature — Human", legB2PavelMalikiOracle, 5, 3, []string{"B", "R"}, 7, 3},
		{"Wall of Opposition", "Creature — Wall", legB2WallOfOppositionOracle, 0, 6, []string{"C"}, 2, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			src := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, true)
			g.WithWriteLock(func() {
				c := findBattlefieldCardForTest(g, src)
				c.Power, c.Toughness = tc.power, tc.toughness
			})
			// Twice: no tap in the cost, and summoning sickness does not matter.
			for range 2 {
				apaMana(me, tc.mana...)
				apaActivate(t, g, me, src, 0, game.ActivateAbilityParams{})
			}
			if p, tt := ptOf(t, g, src); p != tc.wantP || tt != tc.wantT {
				t.Errorf("after two pumps %d/%d, want %d/%d", p, tt, tc.wantP, tc.wantT)
			}
			if alliTapped(t, g, src) {
				t.Error("the ability tapped its source")
			}
			advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if p, tt := ptOf(t, g, src); p != tc.power || tt != tc.toughness {
				t.Errorf("next turn %d/%d, want %d/%d", p, tt, tc.power, tc.toughness)
			}
		})
	}
}
