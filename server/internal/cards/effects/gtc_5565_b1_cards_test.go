package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gtc_5565_b1_cards_test.go — Gatecrash, batch 1 (#5565). Every test
// plays the real catalog entry through the engine.

const (
	gtcGutterSkulkOracle         = "1e775053-ab50-4cf7-8eef-e19a6eb79e1b"
	gtcRuinationWurmOracle       = "7d084fe0-4896-4a1d-839c-81f55591c18a"
	gtcDrakewingKrasisOracle     = "782e8499-0714-4cee-869e-8ebd7dad0242"
	gtcMillennialGargoyleOracle  = "fe07a23a-e526-4277-8c39-6bae5355f0d4"
	gtcMerfolkOfTheDepthsOracle  = "263df914-aea1-48dc-b517-7d515307789e"
	gtcAdaptiveSnapjawOracle     = "0a56e2ad-684d-45a7-9145-479d1505bba2"
	gtcClingingAnemonesOracle    = "7058e7c3-b704-47f6-85f8-ee584f76a1fe"
	gtcCloudfinRaptorOracle      = "cb0e1afa-fd9e-4bf9-bcb1-6e8110d6e2ff"
	gtcCrocanuraOracle           = "2071f278-77ba-46f8-85a0-ccd761199f29"
	gtcShambleshark              = "0e235b96-9507-4f4b-9578-4fb8ed854cc2"
	gtcBasilicaGuardsOracle      = "24d2e266-2dcb-40f1-a465-cfc63e9aa6ae"
	gtcKingpinsPetOracle         = "1c4561d1-377c-4a6a-b7a2-8ce81efa39c6"
	gtcKnightOfObligationOracle  = "9051f778-7718-4e8d-88ff-ca2aee1b14e5"
	gtcBorosEliteOracle          = "7c83fb45-03af-4f1a-b7e3-2c4f47511658"
	gtcWarmindInfantryOracle     = "9e32750a-8e5f-4d81-92a8-c3d4ea1477fd"
	gtcDaringSkyjekOracle        = "f15a202d-c1c4-49bd-924a-7a0a5667ff70"
	gtcOrdruunVeteranOracle      = "119af12a-b767-4408-9bb5-72a6d0ec63d1"
	gtcWojekHalberdiersOracle    = "96d6a400-f1bc-40f8-91ef-1d3cb96926f1"
	gtcBomberCorpsOracle         = "a429c351-d281-4e49-8564-bd96bdf5c41b"
	gtcNavSquadCommandosOracle   = "affdf41e-c851-4f5d-bd3a-7ca7ada4416a"
	gtcGhorClanRampagerOracle    = "ec7a9e7f-742b-4344-983e-73372c6a1760"
	gtcScabClanChargerOracle     = "2f947056-fa10-4491-95e5-532ea3b76d10"
	gtcScorchwalkerOracle        = "566d6fcd-dd50-4135-a912-ec923fc433df"
	gtcSkarrgGoliathOracle       = "72350461-8231-4737-a57f-f20fb5d179a3"
	gtcSkinbrandGoblinOracle     = "2e1d5873-50e6-432a-b7a9-9f8d2632a0b9"
	gtcSlaughterhornOracle       = "f1fca987-3484-4a76-86b6-d61ef0a2cf0f"
	gtcViashinoShanktailOracle   = "1e49f33b-84c0-4b68-abd4-e8c00b3ce98f"
	gtcWastelandViperOracle      = "e30b4eb8-6393-486a-9f8e-970ad2351d8a"
	gtcWreckingOgreOracle        = "85d03947-a39a-4f4f-97c3-6d1094c9248e"
	gtcZhurTaaSwineOracle        = "dcc31cbf-f75d-4584-a97b-aa9d19d7b142"
	gtcBorosKeyruneOracle        = "991ae2f7-3f8e-4364-8698-5277786d7dcd"
	gtcGruulKeyruneOracle        = "98bdb1e1-7550-47e8-aa06-f81802b257a0"
	gtcOrzhovKeyruneOracle       = "9808c424-0d74-45b8-a092-fd296c6833d2"
	gtcSimicKeyruneOracle        = "9e737a22-a352-411d-8412-96770c7f6cce"
	gtcDiscipleOldWaysOracle     = "d19d69c5-9f11-4ac1-9fc7-1636689f23a0"
	gtcDutifulThrullOracle       = "744fcf77-b597-48a0-92ec-f3a617c03df1"
	gtcHorrorOfTheDimOracle      = "83e91e0a-2680-4bd3-98eb-7567791e77d5"
	gtcMetropolisSpriteOracle    = "924c7c4e-9c4b-496a-ba49-7fb3918a8d07"
	gtcZarichiTigerOracle        = "0749e79f-e61d-4037-bb39-75f1ebf46c5d"
	gtcFrilledOculusOracle       = "07ba4fb4-f95b-49ca-be6c-6cdda0f1f030"
	gtcTruefirePaladinOracle     = "2a8a239c-7cc7-48d5-b34b-60b0da4e4b20"
	gtcBurningTreeEmissaryOracle = "327d9679-0049-4401-8dab-e0fb362306bd"
	gtcMortusStriderOracle       = "15af3bd0-022b-472e-b21e-29b925b7692b"
	gtcSparkTrooperOracle        = "3c82389c-7442-4599-89a0-fb2482e727a3"
	gtcFortressCyclopsOracle     = "7790bd6c-75e2-4abd-87ab-e1c15f1d53ab"
	gtcArrowsOfJusticeOracle     = "46e311ef-c86e-4aad-8b42-ef0d68a90c2e"
	gtcShatteringBlowOracle      = "dc9374df-134d-4d32-842e-991e2bfb2dc1"
	gtcWildwoodRebirthOracle     = "7a4ddb8f-08ad-4bf9-82f6-72b073dd859a"

	gtcBearType = "Creature — Bear"
)

// gtcGraveyardCard puts a card of the given type line into p's graveyard.
func gtcGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// gtcAttackWith pushes `others` vanilla attackers beside `id` and
// declares them all against the opponent, from a main phase.
func gtcAttackWith(t *testing.T, g *game.Game, id uuid.UUID, others int) {
	t.Helper()
	me, opp := bnsActive(g)
	attackers := []uuid.UUID{id}
	for i := 0; i < others; i++ {
		attackers = append(attackers, b12Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 1, 1))
	}
	declareAttack(t, g, opp.ID, attackers...)
}

// --- every card is in the catalog ----------------------------------------

func TestGtc5565B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gtcGutterSkulkOracle: "Gutter Skulk", gtcRuinationWurmOracle: "Ruination Wurm",
		gtcDrakewingKrasisOracle: "Drakewing Krasis", gtcMillennialGargoyleOracle: "Millennial Gargoyle",
		gtcMerfolkOfTheDepthsOracle: "Merfolk of the Depths", gtcAdaptiveSnapjawOracle: "Adaptive Snapjaw",
		gtcClingingAnemonesOracle: "Clinging Anemones", gtcCloudfinRaptorOracle: "Cloudfin Raptor",
		gtcCrocanuraOracle: "Crocanura", gtcShambleshark: "Shambleshark",
		gtcBasilicaGuardsOracle: "Basilica Guards", gtcKingpinsPetOracle: "Kingpin's Pet",
		gtcKnightOfObligationOracle: "Knight of Obligation", gtcBorosEliteOracle: "Boros Elite",
		gtcWarmindInfantryOracle: "Warmind Infantry", gtcDaringSkyjekOracle: "Daring Skyjek",
		gtcOrdruunVeteranOracle: "Ordruun Veteran", gtcWojekHalberdiersOracle: "Wojek Halberdiers",
		gtcBomberCorpsOracle: "Bomber Corps", gtcNavSquadCommandosOracle: "Nav Squad Commandos",
		gtcGhorClanRampagerOracle: "Ghor-Clan Rampager", gtcScabClanChargerOracle: "Scab-Clan Charger",
		gtcScorchwalkerOracle: "Scorchwalker", gtcSkarrgGoliathOracle: "Skarrg Goliath",
		gtcSkinbrandGoblinOracle: "Skinbrand Goblin", gtcSlaughterhornOracle: "Slaughterhorn",
		gtcViashinoShanktailOracle: "Viashino Shanktail", gtcWastelandViperOracle: "Wasteland Viper",
		gtcWreckingOgreOracle: "Wrecking Ogre", gtcZhurTaaSwineOracle: "Zhur-Taa Swine",
		gtcBorosKeyruneOracle: "Boros Keyrune", gtcGruulKeyruneOracle: "Gruul Keyrune",
		gtcOrzhovKeyruneOracle: "Orzhov Keyrune", gtcSimicKeyruneOracle: "Simic Keyrune",
		gtcDiscipleOldWaysOracle: "Disciple of the Old Ways", gtcDutifulThrullOracle: "Dutiful Thrull",
		gtcHorrorOfTheDimOracle: "Horror of the Dim", gtcMetropolisSpriteOracle: "Metropolis Sprite",
		gtcZarichiTigerOracle: "Zarichi Tiger", gtcFrilledOculusOracle: "Frilled Oculus",
		gtcTruefirePaladinOracle: "Truefire Paladin", gtcBurningTreeEmissaryOracle: "Burning-Tree Emissary",
		gtcMortusStriderOracle: "Mortus Strider", gtcSparkTrooperOracle: "Spark Trooper",
		gtcFortressCyclopsOracle: "Fortress Cyclops", gtcArrowsOfJusticeOracle: "Arrows of Justice",
		gtcShatteringBlowOracle: "Shattering Blow", gtcWildwoodRebirthOracle: "Wildwood Rebirth",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v with %d caveats, want Full with none", name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords -----------------------------------------------------------

func TestGtc5565B1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Drakewing Krasis", gtcDrakewingKrasisOracle, []string{"flying", "trample"}},
		{"Millennial Gargoyle", gtcMillennialGargoyleOracle, []string{"flying"}},
		{"Merfolk of the Depths", gtcMerfolkOfTheDepthsOracle, []string{"flash"}},
		{"Clinging Anemones", gtcClingingAnemonesOracle, []string{"defender"}},
		{"Cloudfin Raptor", gtcCloudfinRaptorOracle, []string{"flying"}},
		{"Crocanura", gtcCrocanuraOracle, []string{"reach"}},
		{"Shambleshark", gtcShambleshark, []string{"flash"}},
		{"Basilica Guards", gtcBasilicaGuardsOracle, []string{"defender"}},
		{"Kingpin's Pet", gtcKingpinsPetOracle, []string{"flying"}},
		{"Knight of Obligation", gtcKnightOfObligationOracle, []string{"vigilance"}},
		{"Ghor-Clan Rampager", gtcGhorClanRampagerOracle, []string{"trample"}},
		{"Skarrg Goliath", gtcSkarrgGoliathOracle, []string{"trample"}},
		{"Viashino Shanktail", gtcViashinoShanktailOracle, []string{"first strike"}},
		{"Wasteland Viper", gtcWastelandViperOracle, []string{"deathtouch"}},
		{"Wrecking Ogre", gtcWreckingOgreOracle, []string{"double strike"}},
		{"Metropolis Sprite", gtcMetropolisSpriteOracle, []string{"flying"}},
		{"Truefire Paladin", gtcTruefirePaladinOracle, []string{"vigilance"}},
		{"Spark Trooper", gtcSparkTrooperOracle, []string{"trample", "lifelink", "haste"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, "Creature", tc.oracle, 2, 2)
			assertKeywords(t, g, id, tc.keywords...)
		})
	}
}

func TestGtc5565B1VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Gutter Skulk", gtcGutterSkulkOracle},
		{"Ruination Wurm", gtcRuinationWurmOracle},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", tc.name)
		}
		if len(spec.PrintedKeywords)+len(spec.Activated)+len(spec.Triggered)+len(spec.Static) != 0 {
			t.Errorf("%s declares abilities it does not print", tc.name)
		}
	}
}

// Merfolk of the Depths is cast outside a main phase.
func TestGtc5565B1MerfolkOfTheDepthsHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(me, "Merfolk of the Depths", "Creature — Merfolk Soldier", gtcMerfolkOfTheDepthsOracle)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Merfolk of the Depths at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Merfolk of the Depths did not resolve")
	}
}

// --- evolve ---------------------------------------------------------------------

// Each evolve creature gets one +1/+1 counter when a creature bigger in
// both power and toughness enters under its controller's control.
func TestGtc5565B1EvolveCreaturesGrow(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
	}{
		{"Adaptive Snapjaw", gtcAdaptiveSnapjawOracle, 6, 2},
		{"Clinging Anemones", gtcClingingAnemonesOracle, 1, 4},
		{"Cloudfin Raptor", gtcCloudfinRaptorOracle, 0, 1},
		{"Crocanura", gtcCrocanuraOracle, 1, 3},
		{"Shambleshark", gtcShambleshark, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat].ID
			id := b12Push(g, me, tc.name, "Creature", tc.oracle, tc.power, tc.tough)
			enterCreature(t, g, me, "Big", tc.power+1, tc.tough+1)
			passPriorityAroundTable(t, g)
			if got := plusOnes(g, id); got != 1 {
				t.Errorf("%d +1/+1 counters, want 1", got)
			}
		})
	}
}

// --- extort ---------------------------------------------------------------------

func TestGtc5565B1ExtortDrainsOnACast(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Basilica Guards", gtcBasilicaGuardsOracle},
		{"Kingpin's Pet", gtcKingpinsPetOracle},
		{"Knight of Obligation", gtcKnightOfObligationOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			b12Push(g, me.ID, tc.name, "Creature", tc.oracle, 2, 2)
			e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
			for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatal(err)
				}
			}
			if !hasPayUnlessFor(g, me.ID) {
				t.Fatal("casting a spell did not offer the extort payment")
			}
			l0, l1 := me.Life, opp.Life
			floatForTest(g, me, "W")
			answerPayUnless(t, g, me.ID, true)
			if me.Life != l0+3 || opp.Life != l1-1 {
				t.Errorf("life me %d->%d, opp %d->%d; want a 1-point drain on each of three opponents", l0, me.Life, l1, opp.Life)
			}
		})
	}
}

// --- battalion ------------------------------------------------------------------

func TestGtc5565B1BattalionBonuses(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		wantP, wantT int
		keyword      string
	}{
		{"Boros Elite", gtcBorosEliteOracle, 1, 1, 3, 3, ""},
		{"Warmind Infantry", gtcWarmindInfantryOracle, 2, 3, 4, 3, ""},
		{"Daring Skyjek", gtcDaringSkyjekOracle, 3, 1, 3, 1, "flying"},
		{"Ordruun Veteran", gtcOrdruunVeteranOracle, 3, 1, 3, 1, "double strike"},
		{"Wojek Halberdiers", gtcWojekHalberdiersOracle, 3, 2, 3, 2, "first strike"},
		{"Nav Squad Commandos", gtcNavSquadCommandosOracle, 3, 5, 4, 6, ""},
	} {
		t.Run(tc.name+" with two others", func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := bnsActive(g)
			advanceToMain(t, g)
			id := b12Push(g, me.ID, tc.name, "Creature — Human Soldier", tc.oracle, tc.power, tc.tough)
			gtcAttackWith(t, g, id, 2)
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("no %s", tc.keyword)
			}
		})
		t.Run(tc.name+" with one other", func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := bnsActive(g)
			advanceToMain(t, g)
			id := b12Push(g, me.ID, tc.name, "Creature — Human Soldier", tc.oracle, tc.power, tc.tough)
			gtcAttackWith(t, g, id, 1)
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, id); p != tc.power || tough != tc.tough {
				t.Errorf("%d/%d without battalion, want %d/%d", p, tough, tc.power, tc.tough)
			}
			if tc.keyword != "" && hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s without battalion", tc.keyword)
			}
		})
	}
}

func TestGtc5565B1NavSquadCommandosUntapsItself(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := bnsActive(g)
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Nav Squad Commandos", "Creature — Human Soldier", gtcNavSquadCommandosOracle, 3, 5)
	gtcAttackWith(t, g, id, 2)
	if !hmlTapped(t, g, id) {
		t.Fatal("attacking did not tap the Commandos")
	}
	passPriorityAroundTable(t, g)
	if hmlTapped(t, g, id) {
		t.Error("battalion did not untap the Commandos")
	}
}

func TestGtc5565B1BomberCorpsPingsAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Bomber Corps", "Creature — Human Soldier", gtcBomberCorpsOracle, 1, 2)
	life := opp.Life
	gtcAttackWith(t, g, id, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-1)
	}
}

func TestGtc5565B1BomberCorpsNeedsTwoOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := bnsActive(g)
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Bomber Corps", "Creature — Human Soldier", gtcBomberCorpsOracle, 1, 2)
	gtcAttackWith(t, g, id, 1)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("battalion triggered with one other attacker")
	}
}

// --- bloodrush ------------------------------------------------------------------

func TestGtc5565B1BloodrushPumpsAnAttacker(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		generic      int
		colours      []string
		power, tough int
		keyword      string
	}{
		{"Ghor-Clan Rampager", gtcGhorClanRampagerOracle, 0, []string{"R", "G"}, 4, 4, "trample"},
		{"Scab-Clan Charger", gtcScabClanChargerOracle, 1, []string{"G"}, 2, 4, ""},
		{"Scorchwalker", gtcScorchwalkerOracle, 1, []string{"R", "R"}, 5, 1, ""},
		{"Skarrg Goliath", gtcSkarrgGoliathOracle, 5, []string{"G", "G"}, 9, 9, "trample"},
		{"Skinbrand Goblin", gtcSkinbrandGoblinOracle, 0, []string{"R"}, 2, 1, ""},
		{"Slaughterhorn", gtcSlaughterhornOracle, 0, []string{"G"}, 3, 2, ""},
		{"Viashino Shanktail", gtcViashinoShanktailOracle, 2, []string{"R"}, 3, 1, "first strike"},
		{"Wasteland Viper", gtcWastelandViperOracle, 0, []string{"G"}, 1, 2, "deathtouch"},
		{"Wrecking Ogre", gtcWreckingOgreOracle, 3, []string{"R", "R"}, 3, 3, "double strike"},
		{"Zhur-Taa Swine", gtcZhurTaaSwineOracle, 1, []string{"R", "G"}, 5, 4, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
			bear := b12Creature(g, me.ID, "Bear", gtcBearType, 2, 2)
			idle := b12Creature(g, me.ID, "Idle", gtcBearType, 2, 2)
			card := handCardFull(me, tc.name, "Creature", "", tc.oracle, nil)
			advanceToDeclareAttackersOf(t, g, seat)
			if err := g.DeclareAttacker(bear, opp.ID); err != nil {
				t.Fatal(err)
			}
			lockInAttacks(t, g)
			fillPool(me, tc.generic)
			for _, c := range tc.colours {
				fillPoolColored(me, c, 1)
			}
			if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Targets: cardRefs(idle), Strict: true}); err == nil {
				t.Fatal("bloodrush targeted a creature that is not attacking")
			}
			if !me.Hand.Contains(card) {
				t.Fatal("a refused activation discarded the card")
			}
			if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
				t.Fatalf("bloodrush: %v", err)
			}
			if !me.Graveyard.Contains(card) {
				t.Error("the discard was not part of the cost")
			}
			passPriorityAroundTable(t, g)
			c := layeredCard(t, g, bear)
			if c.CurrentPower() != 2+tc.power || c.CurrentToughness() != 2+tc.tough {
				t.Errorf("the attacker is %d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), 2+tc.power, 2+tc.tough)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the attacker has no %s", tc.keyword)
			}
		})
	}
}

// --- Keyrunes -------------------------------------------------------------------

var gtcKeyrunes = []struct {
	name, oracle string
	colours      []string
	subtype      string
	power, tough int
	keyword      string
}{
	{"Boros Keyrune", gtcBorosKeyruneOracle, []string{"R", "W"}, "Soldier", 1, 1, "double strike"},
	{"Gruul Keyrune", gtcGruulKeyruneOracle, []string{"R", "G"}, "Beast", 3, 2, "trample"},
	{"Orzhov Keyrune", gtcOrzhovKeyruneOracle, []string{"W", "B"}, "Thrull", 1, 4, "lifelink"},
	{"Simic Keyrune", gtcSimicKeyruneOracle, []string{"G", "U"}, "Crab", 2, 3, "hexproof"},
}

func TestGtc5565B1KeyrunesTapForEitherColour(t *testing.T) {
	for _, tc := range gtcKeyrunes {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			rune := s58p6Push(g, me.ID, tc.name, "Artifact", tc.oracle, "{3}", 0, 0)
			for _, colour := range tc.colours {
				me.ManaPool.EmptyPool()
				findBattlefieldCardByID(g, rune).Tapped = false
				if err := g.ActivateManaAbility(me.ID, rune, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
					t.Fatalf("tap for %s: %v", colour, err)
				}
				if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != colour {
					t.Errorf("pool = %v, want one %s", me.ManaPool.Units(), colour)
				}
			}
		})
	}
}

func TestGtc5565B1KeyrunesBecomeArtifactCreaturesUntilEndOfTurn(t *testing.T) {
	for _, tc := range gtcKeyrunes {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			rune := s58p6Push(g, me.ID, tc.name, "Artifact", tc.oracle, "{3}", 0, 0)
			advanceTo(t, g, game.StepPrecombatMain)
			if s58p6View(t, g, rune).IsCreature() {
				t.Fatal("an idle Keyrune is a creature")
			}
			ngheAnimate(t, g, me, rune, 0, tc.colours...)
			c := s58p6View(t, g, rune)
			if !c.IsCreature() || !c.IsArtifact() {
				t.Fatalf("creature=%v artifact=%v, want both", c.IsCreature(), c.IsArtifact())
			}
			if !slices.Contains(c.Effective().Subtypes, tc.subtype) {
				t.Errorf("subtypes %v, want %s", c.Effective().Subtypes, tc.subtype)
			}
			if !slices.Equal(c.EffectiveColors(), tc.colours) {
				t.Errorf("colours %v, want %v", c.EffectiveColors(), tc.colours)
			}
			if c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
				t.Errorf("%d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), tc.power, tc.tough)
			}
			if !hasEffectiveKeyword(t, g, rune, tc.keyword) {
				t.Errorf("no %s", tc.keyword)
			}
			advanceToNextSeatsTurn(t, g)
			c = s58p6View(t, g, rune)
			if c.IsCreature() || !c.IsArtifact() {
				t.Errorf("after the turn: creature=%v artifact=%v, want an artifact only", c.IsCreature(), c.IsArtifact())
			}
		})
	}
}

// --- activated abilities ---------------------------------------------------------

func TestGtc5565B1SelfGrants(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		index        int
		mana         []string
		keyword      string
	}{
		{"Disciple of the Old Ways", gtcDiscipleOldWaysOracle, 0, []string{"R"}, "first strike"},
		{"Horror of the Dim", gtcHorrorOfTheDimOracle, 0, []string{"U"}, "hexproof"},
		{"Truefire Paladin", gtcTruefirePaladinOracle, 1, []string{"R", "W"}, "first strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, "Creature", tc.oracle, 2, 2)
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s before the activation", tc.keyword)
			}
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, tc.index, game.ActivateAbilityParams{})
			if !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("no %s after the activation", tc.keyword)
			}
			advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
			if hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("%s outlasted the turn", tc.keyword)
			}
		})
	}
}

func TestGtc5565B1SelfPumps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		index        int
		mana         []string
		power, tough int
		wantP, wantT int
	}{
		{"Metropolis Sprite", gtcMetropolisSpriteOracle, 0, []string{"U"}, 1, 2, 2, 1},
		{"Frilled Oculus", gtcFrilledOculusOracle, 0, []string{"C", "G"}, 1, 3, 3, 5},
		{"Truefire Paladin", gtcTruefirePaladinOracle, 0, []string{"R", "W"}, 2, 2, 4, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			id := b12Push(g, me.ID, tc.name, "Creature", tc.oracle, tc.power, tc.tough)
			apaMana(me, tc.mana...)
			apaActivate(t, g, me, id, tc.index, game.ActivateAbilityParams{})
			if p, tough := ptOf(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestGtc5565B1FrilledOculusActivatesOnlyOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := b12Push(g, me.ID, "Frilled Oculus", "Creature — Homunculus", gtcFrilledOculusOracle, 1, 3)
	apaMana(me, "C", "G")
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	apaMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Frilled Oculus activated twice in one turn")
	}
}

func TestGtc5565B1DutifulThrullRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	thrull := b12Push(g, me.ID, "Dutiful Thrull", "Creature — Thrull", gtcDutifulThrullOracle, 1, 1)
	apaMana(me, "B")
	apaActivate(t, g, me, thrull, 0, game.ActivateAbilityParams{})
	b27Kill(g, thrull)
	if !g.Battlefield.Contains(thrull) {
		t.Error("the shielded Thrull was destroyed")
	}
	if !hmlTapped(t, g, thrull) {
		t.Error("regenerating did not tap the Thrull")
	}
}

func TestGtc5565B1ZarichiTigerGainsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	tiger := b12Push(g, me.ID, "Zarichi Tiger", "Creature — Cat", gtcZarichiTigerOracle, 2, 3)
	life := me.Life
	apaMana(me, "C", "W")
	apaActivate(t, g, me, tiger, 0, game.ActivateAbilityParams{})
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if !hmlTapped(t, g, tiger) {
		t.Error("the Tiger did not tap")
	}
}

// --- triggered abilities ----------------------------------------------------------

func TestGtc5565B1BurningTreeEmissaryAddsRedAndGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := castHoldCreature(t, g, "Burning-Tree Emissary", "Creature — Human Shaman", gtcBurningTreeEmissaryOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Burning-Tree Emissary did not resolve")
	}
	var colours []string
	for _, u := range me.ManaPool.Units() {
		colours = append(colours, u.Color)
	}
	slices.Sort(colours)
	if !slices.Equal(colours, []string{"G", "R"}) {
		t.Errorf("pool %v, want {R}{G}", colours)
	}
}

func TestGtc5565B1MortusStriderReturnsToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	strider := b12Push(g, me.ID, "Mortus Strider", "Creature — Skeleton", gtcMortusStriderOracle, 1, 1)
	killCreature(t, g, me.ID, strider)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(strider) || me.Graveyard.Contains(strider) {
		t.Error("Mortus Strider did not return to its owner's hand")
	}
}

func TestGtc5565B1SparkTrooperIsSacrificedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b12Push(g, me.ID, "Spark Trooper", "Creature — Elemental Soldier", gtcSparkTrooperOracle, 6, 1)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Fatal("Spark Trooper was not sacrificed at the end step")
	}
}

func TestGtc5565B1FortressCyclopsGrowsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Fortress Cyclops", "Creature — Cyclops Soldier", gtcFortressCyclopsOracle, 3, 3)
	declareAttack(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 6 || tough != 3 {
		t.Errorf("attacking: %d/%d, want 6/3", p, tough)
	}
}

func TestGtc5565B1FortressCyclopsGrowsWhenItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cyclops := b12Push(g, opp.ID, "Fortress Cyclops", "Creature — Cyclops Soldier", gtcFortressCyclopsOracle, 3, 3)
	attacker := b12Creature(g, me.ID, "Attacker", gtcBearType, 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: cyclops, Target: attacker, Amount: 1})
	})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, cyclops); p != 3 || tough != 6 {
		t.Errorf("blocking: %d/%d, want 3/6", p, tough)
	}
}

// --- the spells --------------------------------------------------------------------

func TestGtc5565B1ArrowsOfJusticeHitsOnlyACombatant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	raider := b12Creature(g, me.ID, "Raider", gtcBearType, 5, 5)
	idle := b12Creature(g, me.ID, "Idle", gtcBearType, 5, 5)
	declareAttack(t, g, opp.ID, raider)
	id := b13HandCard(me, "Arrows of Justice", "Instant", gtcArrowsOfJusticeOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardTarget(idle)}); err == nil {
		t.Fatal("Arrows of Justice targeted a creature that isn't in combat")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardTarget(raider)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, raider).DamageMarked; got != 4 {
		t.Errorf("the attacker has %d damage, want 4", got)
	}
}

func TestGtc5565B1ShatteringBlowExilesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Bear", gtcBearType, 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Shattering Blow", "Instant", gtcShatteringBlowOracle, cardTarget(bear)); err == nil {
		t.Fatal("Shattering Blow targeted a creature")
	}
	castCatalogSpell(t, g, "Shattering Blow", "Instant", gtcShatteringBlowOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not exiled")
	}
}

func TestGtc5565B1WildwoodRebirthReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := gtcGraveyardCard(me, "Bear", gtcBearType)
	rock := gtcGraveyardCard(me, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Wildwood Rebirth", "Instant", gtcWildwoodRebirthOracle, cardTarget(rock)); err == nil {
		t.Fatal("Wildwood Rebirth targeted an artifact card")
	}
	castCatalogSpell(t, g, "Wildwood Rebirth", "Instant", gtcWildwoodRebirthOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || me.Graveyard.Contains(bear) {
		t.Error("the creature card did not return to hand")
	}
}
