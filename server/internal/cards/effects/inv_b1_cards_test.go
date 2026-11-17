package effects

import (
	"testing"

	"github.com/google/uuid"

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

// invB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func invB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// invB1Activate activates the source's ability `index` for `who` with
// the given targets, without resolving it.
func invB1Activate(g *game.Game, who, source uuid.UUID, index int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, index, game.ActivateAbilityParams{Strict: true, Targets: targets})
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
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", faces[face].Name, err)
	}
	return card.InstanceID
}

// --- Armadillo Cloak -----------------------------------------------------------------

func TestInvB1ArmadilloCloakPumpsGrantsTrampleAndGainsLifeOnAnyDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	castCatalogSpell(t, g, "Armadillo Cloak", auraTypeLine, invB1ArmadilloOracle, cardTarget(mine))
	passPriorityAroundTable(t, g)

	if p, tough := ptOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, mine, "trample") {
		t.Error("the enchanted creature does not have trample")
	}

	before := me.Life
	emitNoncombatDamage(g, mine, theirs, 3)
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 3 {
		t.Errorf("damage to a creature gained %d life, want 3", got)
	}
	dealCombatDamageToPlayer(g, mine, opp.ID, 4)
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 7 {
		t.Errorf("combat damage to a player brought the gain to %d, want 7", got)
	}
	emitNoncombatDamage(g, theirs, mine, 2)
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 7 {
		t.Errorf("damage dealt TO the enchanted creature gained life: total %d, want 7", got)
	}
}

// --- Fires of Yavimaya ---------------------------------------------------------------

func TestInvB1FiresOfYavimayaHastesYourCreaturesAndSacrificesForAPump(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	fires := invB1Push(g, me.ID, "Fires of Yavimaya", "Enchantment", invB1FiresOracle, 0, 0)
	mine := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	if !hasKeywordOnBattlefield(t, g, mine, "haste") {
		t.Error("your creature does not have haste")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
	if err := invB1Activate(g, me.ID, fires, 0, cardTarget(theirs)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fires) {
		t.Fatal("Fires of Yavimaya was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if hasKeywordOnBattlefield(t, g, mine, "haste") {
		t.Error("your creature kept haste after Fires of Yavimaya left")
	}
}

// --- Blazing Specter -----------------------------------------------------------------

func TestInvB1BlazingSpecterMakesTheDamagedPlayerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	specter := invB1Push(g, me.ID, "Blazing Specter", "Creature — Specter", invB1SpecterOracle, 2, 2)
	for _, kw := range []string{"flying", "haste"} {
		if !hasKeywordOnBattlefield(t, g, specter, kw) {
			t.Errorf("Blazing Specter does not have %s", kw)
		}
	}
	keep := pushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := pushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	pushCatalogHandCard(other, "Other", "Instant", "")

	dealNoncombatDamage(g, me.ID, specter, opp.ID)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, opp.ID) != nil {
		t.Fatal("noncombat damage asked for a discard")
	}

	dealCombatDamage(g, me.ID, specter, opp.ID)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d discards, want 1", n)
	}
	if discardChoiceFor(g, other.ID) != nil {
		t.Error("a player not dealt damage was asked to discard")
	}
	answerDiscard(t, g, opp.ID, pitch)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
}

// --- Charging Troll ------------------------------------------------------------------

func TestInvB1ChargingTrollHasVigilanceAndRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	troll := invB1Push(g, me.ID, "Charging Troll", "Creature — Troll", invB1TrollOracle, 3, 3)
	if !hasKeywordOnBattlefield(t, g, troll, "vigilance") {
		t.Error("Charging Troll does not have vigilance")
	}
	dcMana(t, g, me.ID, "{G}")
	if err := invB1Activate(g, me.ID, troll, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := regenShieldsOn(t, g, troll); n != 1 {
		t.Fatalf("%d regeneration shields, want 1", n)
	}
	b15Destroy(t, g, troll)
	if !g.Battlefield.Contains(troll) {
		t.Fatal("the regeneration shield did not save Charging Troll")
	}
	if !findCardAnywhere(t, g, troll).Tapped {
		t.Error("the regenerated Troll is not tapped")
	}
}

// --- Riptide Crab --------------------------------------------------------------------

func TestInvB1RiptideCrabDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top")
	crab := invB1Push(g, me.ID, "Riptide Crab", "Creature — Crab", invB1CrabOracle, 1, 3)
	if !hasKeywordOnBattlefield(t, g, crab, "vigilance") {
		t.Error("Riptide Crab does not have vigilance")
	}
	before := me.Hand.Size()
	b15Destroy(t, g, crab)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand is %d after the Crab died, want %d", got, before+1)
	}
}

// --- Hanna, Ship's Navigator ---------------------------------------------------------

func TestInvB1HannaReturnsAnArtifactOrEnchantmentCardFromYourGraveyard(t *testing.T) {
	for _, typeLine := range []string{"Artifact", "Enchantment — Aura"} {
		t.Run(typeLine, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, _ := blbB1Seats(g)
			hanna := invB1Push(g, me.ID, "Hanna, Ship's Navigator", "Legendary Creature — Human Artificer", invB1HannaOracle, 1, 2)
			want := pushGraveyardCardTyped(me, "Wanted", typeLine)
			bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
			theirs := pushGraveyardCardTyped(opp, "Their Relic", "Artifact")

			dcMana(t, g, me.ID, "{C}{W}{U}")
			for _, bad := range []uuid.UUID{bear, theirs} {
				if err := invB1Activate(g, me.ID, hanna, 0, cardTarget(bad)...); err == nil {
					t.Fatal("a creature card, or a card in another graveyard, was accepted as the target")
				}
			}
			if err := invB1Activate(g, me.ID, hanna, 0, cardTarget(want)...); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(want) {
				t.Errorf("the %s card did not come back to hand", typeLine)
			}
			if !findCardAnywhere(t, g, hanna).Tapped {
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			seedLibrary(me, "A", "B", "C")
			relic := invB1Push(g, opp.ID, "Relic", "Artifact", "", 0, 0)
			bear := seedBear(g, opp.ID)
			if _, err := castWithOptionalCosts(t, g, "Dismantling Blow", "Instant", invB1BlowOracle, cardTarget(bear), tc.optional, nil); err == nil {
				t.Fatal("a creature was accepted as the target")
			}
			hand := me.Hand.Size()
			if _, err := castWithOptionalCosts(t, g, "Dismantling Blow", "Instant", invB1BlowOracle, cardTarget(relic), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	bear := seedBear(g, g.Seats[0].ID)
	invB1CastHalf(t, g, invB1WaxWaneOracle, invB1WaxWaneFaces, 0, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
}

func TestInvB1WaneDestroysTargetEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ench := invB1Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	invB1CastHalf(t, g, invB1WaxWaneOracle, invB1WaxWaneFaces, 1, cardTarget(ench))
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	invB1CastHalf(t, g, invB1AssaultOracle, invB1AssaultBatteryFaces, 0, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("the target lost %d life, want 2", got)
	}
}

func TestInvB1BatteryCreatesAThreeThreeElephant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	invB1CastHalf(t, g, invB1AssaultOracle, invB1AssaultBatteryFaces, 1, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant", OracleID: "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	before := me.Life
	invB1CastHalf(t, g, invB1SpiteOracle, invB1SpiteMaliceFaces, 0, cardTarget(shock))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(shock) {
		t.Error("the countered Shock is not in its owner's graveyard")
	}
	if me.Life != before {
		t.Errorf("the countered Shock still dealt damage: life %d, want %d", me.Life, before)
	}
}

func TestInvB1MaliceDestroysANonblackCreatureThroughRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	black := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Black", TypeLine: "Creature — Zombie", ManaCost: "{1}{B}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	green := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green", TypeLine: "Creature — Troll", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, green); c != nil {
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
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardTarget(black)}); err == nil {
		t.Fatal("a black creature was accepted as Malice's target")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardTarget(green)}); err != nil {
		t.Fatalf("cast Malice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("a regeneration shield saved the creature")
	}
}

// --- Ordered Migration -------------------------------------------------------------------

func TestInvB1OrderedMigrationMakesABirdForEachBasicLandType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	pushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Ordered Migration", "Sorcery", invB1MigrationOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ogre := invB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	tapped := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped", TypeLine: "Creature — Ogre",
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	if err := castCatalogSpellErr(t, g, "Backlash", "Instant", invB1BacklashOracle, cardTarget(tapped)); err == nil {
		t.Fatal("a tapped creature was accepted as the target")
	}
	before := opp.Life
	castCatalogSpell(t, g, "Backlash", "Instant", invB1BacklashOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if !findCardAnywhere(t, g, ogre).Tapped {
		t.Error("the target was not tapped")
	}
	if got := before - opp.Life; got != 3 {
		t.Errorf("its controller lost %d life, want 3", got)
	}
}

// --- Stalking Assassin -------------------------------------------------------------------

func TestInvB1StalkingAssassinTapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	assassin := invB1Push(g, me.ID, "Stalking Assassin", "Creature — Human Assassin", invB1AssassinOracle, 1, 1)
	bear := seedBear(g, opp.ID)
	dcMana(t, g, me.ID, "{C}{C}{C}{U}")
	if err := invB1Activate(g, me.ID, assassin, 0, cardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !findCardAnywhere(t, g, bear).Tapped {
		t.Error("the target was not tapped")
	}
	if !findCardAnywhere(t, g, assassin).Tapped {
		t.Error("Stalking Assassin did not tap")
	}
}

func TestInvB1StalkingAssassinDestroysOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	assassin := invB1Push(g, me.ID, "Stalking Assassin", "Creature — Human Assassin", invB1AssassinOracle, 1, 1)
	untapped := seedBear(g, opp.ID)
	tapped := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	dcMana(t, g, me.ID, "{C}{C}{C}{B}")
	if err := invB1Activate(g, me.ID, assassin, 1, cardTarget(untapped)...); err == nil {
		t.Fatal("an untapped creature was accepted as the target")
	}
	if err := invB1Activate(g, me.ID, assassin, 1, cardTarget(tapped)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
}

// --- Heroes' Reunion ---------------------------------------------------------------------

func TestInvB1HeroesReunionGivesTargetPlayerSevenLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Heroes' Reunion", "Instant", invB1ReunionOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := opp.Life - theirs; got != 7 {
		t.Errorf("the target gained %d life, want 7", got)
	}
	if me.Life != mine {
		t.Errorf("the caster's life changed: %d, want %d", me.Life, mine)
	}
}
