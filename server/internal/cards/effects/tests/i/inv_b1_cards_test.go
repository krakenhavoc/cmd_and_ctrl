package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// inv_b1_cards_test.go — Invasion, batch 1 (#5616). Each test plays the
// real catalog entry and asserts what its text says.

const (
	invB1ArmadilloOracle = "15397a59-70a8-441f-abf7-230405a1602b"
	invB1FiresOracle     = "e23d6f3b-0e18-423b-943b-15db7837255b"
	invB1SpecterOracle   = "4c046368-5899-42ef-9972-3a3a6c8fe3fd"
	invB1TrollOracle     = "bbc1b96d-5a63-4091-a2c7-08bd289e0bf2"
	invB1CrabOracle      = "8ae8a583-9843-474a-bd46-8090df883621"
	invB1HannaOracle     = "3b008c71-e2a0-4f56-83d3-655c5dd58e0e"
	invB1BlowOracle      = "300cba5a-adbb-4852-be06-ac00d9d6fd37"
	invB1WaxWaneOracle   = "4744b086-fea2-4992-bfef-68c58cdf9cf0"
	invB1AssaultOracle   = "fcdeb83f-531c-4087-9254-7f19eb065f00"
	invB1SpiteOracle     = "b829de81-04cd-498e-9f31-074f23e2a488"
	invB1MigrationOracle = "ee0a59ba-8d34-43d5-b9a5-42015588e3d5"
	invB1BacklashOracle  = "3db9bf7b-da4e-428f-aa5d-a0cd1afeb16c"
	invB1AssassinOracle  = "12117089-67f7-4d0e-88db-82b27a4c7cdc"
	invB1ReunionOracle   = "beab66e0-2b6c-478f-ba42-8c0c216a527e"
)

func TestInvB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		invB1ArmadilloOracle:      "Armadillo Cloak",
		invB1FiresOracle:          "Fires of Yavimaya",
		invB1SpecterOracle:        "Blazing Specter",
		invB1TrollOracle:          "Charging Troll",
		invB1CrabOracle:           "Riptide Crab",
		invB1HannaOracle:          "Hanna, Ship's Navigator",
		invB1BlowOracle:           "Dismantling Blow",
		invB1WaxWaneOracle:        "Wax",
		invB1WaxWaneOracle + "#1": "Wane",
		invB1AssaultOracle:        "Assault",
		invB1AssaultOracle + "#1": "Battery",
		invB1SpiteOracle:          "Spite",
		invB1SpiteOracle + "#1":   "Malice",
		invB1MigrationOracle:      "Ordered Migration",
		invB1BacklashOracle:       "Backlash",
		invB1AssassinOracle:       "Stalking Assassin",
		invB1ReunionOracle:        "Heroes' Reunion",
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

// invB1CastHalf casts one half of a split card from the active
// player's hand, without resolving it.
func invB1CastHalf(t *testing.T, g *game.Game, oracle string, faces [2]game.Face, face int, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: faces[:],
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", faces[face].Name, err)
	}
	return card.InstanceID
}

// --- Armadillo Cloak -----------------------------------------------------------------

func TestInvB1ArmadilloCloakPumpsGrantsTrampleAndGainsLifeOnAnyDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Armadillo Cloak", effectstest.AuraTypeLine, invB1ArmadilloOracle, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "trample") {
		t.Error("the enchanted creature does not have trample")
	}

	before := me.Life
	effectstest.EmitNoncombatDamage(g, mine, theirs, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - before; got != 3 {
		t.Errorf("damage to a creature gained %d life, want 3", got)
	}
	effectstest.DealCombatDamageToPlayer(g, mine, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - before; got != 7 {
		t.Errorf("combat damage to a player brought the gain to %d, want 7", got)
	}
	effectstest.EmitNoncombatDamage(g, theirs, mine, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - before; got != 7 {
		t.Errorf("damage dealt TO the enchanted creature gained life: total %d, want 7", got)
	}
}

// --- Fires of Yavimaya ---------------------------------------------------------------

func TestInvB1FiresOfYavimayaHastesYourCreaturesAndSacrificesForAPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	fires := effectstest.InvB1Push(g, me.ID, "Fires of Yavimaya", "Enchantment", invB1FiresOracle, 0, 0)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "haste") {
		t.Error("your creature does not have haste")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
	if err := effectstest.InvB1Activate(g, me.ID, fires, 0, effectstest.CardTarget(theirs)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fires) {
		t.Fatal("Fires of Yavimaya was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, theirs); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, mine, "haste") {
		t.Error("your creature kept haste after Fires of Yavimaya left")
	}
}

// --- Blazing Specter -----------------------------------------------------------------

func TestInvB1BlazingSpecterMakesTheDamagedPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	specter := effectstest.InvB1Push(g, me.ID, "Blazing Specter", "Creature — Specter", invB1SpecterOracle, 2, 2)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, specter, kw) {
			t.Errorf("Blazing Specter does not have %s", kw)
		}
	}
	keep := effectstest.PushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := effectstest.PushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	effectstest.PushCatalogHandCard(other, "Other", "Instant", "")

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
	if effectstest.DiscardChoiceFor(g, other.ID) != nil {
		t.Error("a player not dealt damage was asked to discard")
	}
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
}

// --- Charging Troll ------------------------------------------------------------------

func TestInvB1ChargingTrollHasVigilanceAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	troll := effectstest.InvB1Push(g, me.ID, "Charging Troll", "Creature — Troll", invB1TrollOracle, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, troll, "vigilance") {
		t.Error("Charging Troll does not have vigilance")
	}
	effectstest.DcMana(t, g, me.ID, "{G}")
	if err := effectstest.InvB1Activate(g, me.ID, troll, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RegenShieldsOn(t, g, troll); n != 1 {
		t.Fatalf("%d regeneration shields, want 1", n)
	}
	effectstest.B15Destroy(t, g, troll)
	if !g.Battlefield.Contains(troll) {
		t.Fatal("the regeneration shield did not save Charging Troll")
	}
	if !effectstest.FindCardAnywhere(t, g, troll).Tapped {
		t.Error("the regenerated Troll is not tapped")
	}
}

// --- Riptide Crab --------------------------------------------------------------------

func TestInvB1RiptideCrabDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top")
	crab := effectstest.InvB1Push(g, me.ID, "Riptide Crab", "Creature — Crab", invB1CrabOracle, 1, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, crab, "vigilance") {
		t.Error("Riptide Crab does not have vigilance")
	}
	before := me.Hand.Size()
	effectstest.B15Destroy(t, g, crab)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand is %d after the Crab died, want %d", got, before+1)
	}
}

// --- Hanna, Ship's Navigator ---------------------------------------------------------

func TestInvB1HannaReturnsAnArtifactOrEnchantmentCardFromYourGraveyard(t *testing.T) {
	for _, typeLine := range []string{"Artifact", "Enchantment — Aura"} {
		t.Run(typeLine, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp, _ := effectstest.BlbB1Seats(g)
			hanna := effectstest.InvB1Push(g, me.ID, "Hanna, Ship's Navigator", "Legendary Creature — Human Artificer", invB1HannaOracle, 1, 2)
			want := effectstest.PushGraveyardCardTyped(me, "Wanted", typeLine)
			bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
			theirs := effectstest.PushGraveyardCardTyped(opp, "Their Relic", "Artifact")

			effectstest.DcMana(t, g, me.ID, "{C}{W}{U}")
			for _, bad := range []uuid.UUID{bear, theirs} {
				if err := effectstest.InvB1Activate(g, me.ID, hanna, 0, effectstest.CardTarget(bad)...); err == nil {
					t.Fatal("a creature card, or a card in another graveyard, was accepted as the target")
				}
			}
			if err := effectstest.InvB1Activate(g, me.ID, hanna, 0, effectstest.CardTarget(want)...); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(want) {
				t.Errorf("the %s card did not come back to hand", typeLine)
			}
			if !effectstest.FindCardAnywhere(t, g, hanna).Tapped {
				t.Error("Hanna did not tap")
			}
		})
	}
}

// --- Dismantling Blow ----------------------------------------------------------------

func TestInvB1DismantlingBlowDestroysAndDrawsTwoOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		draws    int
	}{
		{"unkicked", nil, 0},
		{"kicked", []int{0}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.SeedLibrary(me, "A", "B", "C")
			relic := effectstest.InvB1Push(g, opp.ID, "Relic", "Artifact", "", 0, 0)
			bear := effectstest.SeedBear(g, opp.ID)
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Dismantling Blow", "Instant", invB1BlowOracle, effectstest.CardTarget(bear), tc.optional, nil); err == nil {
				t.Fatal("a creature was accepted as the target")
			}
			hand := me.Hand.Size()
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Dismantling Blow", "Instant", invB1BlowOracle, effectstest.CardTarget(relic), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(relic) {
				t.Error("the artifact survived")
			}
			// The refused cast left its card in hand; count from after it.
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d cards, want %d", got, tc.draws)
			}
		})
	}
}

// --- Wax // Wane -------------------------------------------------------------------------

var invB1WaxWaneFaces = [2]game.Face{
	{Name: "Wax", TypeLine: "Instant", ManaCost: "{G}", Colors: []string{"G"}},
	{Name: "Wane", TypeLine: "Instant", ManaCost: "{W}", Colors: []string{"W"}},
}

func TestInvB1WaxPumpsTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.SeedBear(g, g.Seats[0].ID)
	invB1CastHalf(t, g, invB1WaxWaneOracle, invB1WaxWaneFaces, 0, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
}

func TestInvB1WaneDestroysTargetEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ench := effectstest.InvB1Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	invB1CastHalf(t, g, invB1WaxWaneOracle, invB1WaxWaneFaces, 1, effectstest.CardTarget(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
}

// --- Assault // Battery ------------------------------------------------------------------

var invB1AssaultBatteryFaces = [2]game.Face{
	{Name: "Assault", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}},
	{Name: "Battery", TypeLine: "Sorcery", ManaCost: "{3}{G}", Colors: []string{"G"}},
}

func TestInvB1AssaultDealsTwoDamageToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	invB1CastHalf(t, g, invB1AssaultOracle, invB1AssaultBatteryFaces, 0, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("the target lost %d life, want 2", got)
	}
}

func TestInvB1BatteryCreatesAThreeThreeElephant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	invB1CastHalf(t, g, invB1AssaultOracle, invB1AssaultBatteryFaces, 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elephant" && c.Controller == me.ID && c.Power == 3 && c.Toughness == 3 && c.HasColor("G") {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d 3/3 green Elephants, want 1", n)
	}
}

// --- Spite // Malice ---------------------------------------------------------------------

var invB1SpiteMaliceFaces = [2]game.Face{
	{Name: "Spite", TypeLine: "Instant", ManaCost: "{3}{U}", Colors: []string{"U"}},
	{Name: "Malice", TypeLine: "Instant", ManaCost: "{3}{B}", Colors: []string{"B"}},
}

func TestInvB1SpiteCountersANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant", OracleID: "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	before := me.Life
	invB1CastHalf(t, g, invB1SpiteOracle, invB1SpiteMaliceFaces, 0, effectstest.CardTarget(shock))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(shock) {
		t.Error("the countered Shock is not in its owner's graveyard")
	}
	if me.Life != before {
		t.Errorf("the countered Shock still dealt damage: life %d, want %d", me.Life, before)
	}
}

func TestInvB1MaliceDestroysANonblackCreatureThroughRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black", TypeLine: "Creature — Zombie", ManaCost: "{1}{B}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green", TypeLine: "Creature — Troll", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() {
		if c := effectstest.BattlefieldCardFor(g, green); c != nil {
			c.RegenerationShields = 1
		}
	})
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: invB1SpiteOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: invB1SpiteMaliceFaces[:],
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardTarget(black)}); err == nil {
		t.Fatal("a black creature was accepted as Malice's target")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardTarget(green)}); err != nil {
		t.Fatalf("cast Malice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("a regeneration shield saved the creature")
	}
}

// --- Ordered Migration -------------------------------------------------------------------

func TestInvB1OrderedMigrationMakesABirdForEachBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Ordered Migration", "Sorcery", invB1MigrationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" && c.Controller == me.ID {
			n++
			if c.Power != 1 || c.Toughness != 1 || !game.HasKeyword(&c, "flying") {
				t.Errorf("a Bird is %d/%d with keywords %v, want a 1/1 flier", c.Power, c.Toughness, c.Keywords)
			}
		}
	}
	if n != 3 {
		t.Errorf("%d Birds, want 3 (Plains, Island, Swamp; the opponent's Forest does not count)", n)
	}
}

// --- Backlash ----------------------------------------------------------------------------

func TestInvB1BacklashTapsTheCreatureAndItHitsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ogre := effectstest.InvB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped", TypeLine: "Creature — Ogre",
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Backlash", "Instant", invB1BacklashOracle, effectstest.CardTarget(tapped)); err == nil {
		t.Fatal("a tapped creature was accepted as the target")
	}
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Backlash", "Instant", invB1BacklashOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FindCardAnywhere(t, g, ogre).Tapped {
		t.Error("the target was not tapped")
	}
	if got := before - opp.Life; got != 3 {
		t.Errorf("its controller lost %d life, want 3", got)
	}
}

// --- Stalking Assassin -------------------------------------------------------------------

func TestInvB1StalkingAssassinTapsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	assassin := effectstest.InvB1Push(g, me.ID, "Stalking Assassin", "Creature — Human Assassin", invB1AssassinOracle, 1, 1)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{U}")
	if err := effectstest.InvB1Activate(g, me.ID, assassin, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FindCardAnywhere(t, g, bear).Tapped {
		t.Error("the target was not tapped")
	}
	if !effectstest.FindCardAnywhere(t, g, assassin).Tapped {
		t.Error("Stalking Assassin did not tap")
	}
}

func TestInvB1StalkingAssassinDestroysOnlyATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	assassin := effectstest.InvB1Push(g, me.ID, "Stalking Assassin", "Creature — Human Assassin", invB1AssassinOracle, 1, 1)
	untapped := effectstest.SeedBear(g, opp.ID)
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{B}")
	if err := effectstest.InvB1Activate(g, me.ID, assassin, 1, effectstest.CardTarget(untapped)...); err == nil {
		t.Fatal("an untapped creature was accepted as the target")
	}
	if err := effectstest.InvB1Activate(g, me.ID, assassin, 1, effectstest.CardTarget(tapped)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
}

// --- Heroes' Reunion ---------------------------------------------------------------------

func TestInvB1HeroesReunionGivesTargetPlayerSevenLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Heroes' Reunion", "Instant", invB1ReunionOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Life - theirs; got != 7 {
		t.Errorf("the target gained %d life, want 7", got)
	}
	if me.Life != mine {
		t.Errorf("the caster's life changed: %d, want %d", me.Life, mine)
	}
}
