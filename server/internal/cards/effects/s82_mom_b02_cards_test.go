package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	momB02IntercessorOracle = "d1d719e9-9c02-40cf-8849-91a5c120a516"
	momB02ChargerOracle     = "fea60c45-cc36-4dd1-8475-23a9361e370d"
	momB02AncientOracle     = "e71eca6f-388c-4750-b70e-3ee1832931f0"
	momB02TerrorOracle      = "e7866ce6-d507-4221-bb19-d3e6fd7b57d8"
	momB02CarapaceOracle    = "06a6130f-f27b-4db0-915d-decd9a7d3c15"
	momB02HalberdOracle     = "450e84c7-afbe-4b30-a3e9-1dff968a5200"
	momB02FanOracle         = "c4614b6c-5e77-465f-996f-504121ea8eff"
	momB02BanesplitterOracl = "aeaf3ce8-6bdf-4a90-8b81-cdbce0f3bb12"
	momB02UrnOracle         = "420290d9-7bb9-4e5a-82ec-c777152ccc72"
	momB02RacerOracle       = "0399b932-2a59-471c-828e-2b16614e0dae"
	momB02StasisOracle      = "83633ea1-a8d5-4351-9047-625d6905386d"
	momB02GraspOracle       = "f2baa0b9-ab0b-4ae9-b5f0-d5b3fcd9cf37"
)

func TestMomB02CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		momB02IntercessorOracle: "Alabaster Host Intercessor",
		momB02ChargerOracle:     "Furnace Host Charger",
		momB02AncientOracle:     "Timberland Ancient",
		momB02TerrorOracle:      "Tidal Terror",
		momB02CarapaceOracle:    "Crystal Carapace",
		momB02HalberdOracle:     "Kor Halberd",
		momB02FanOracle:         "Bladed Battle-Fan",
		momB02BanesplitterOracl: "Mirran Banesplitter",
		momB02UrnOracle:         "Urn of Godfire",
		momB02RacerOracle:       "Flywheel Racer",
		momB02StasisOracle:      "Stasis Field",
		momB02GraspOracle:       "Realmbreaker's Grasp",
	}
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

// momB02Typecycle cycles `name` from hand and checks the search offers only the
// land of the named type.
func momB02Typecycle(t *testing.T, name, typeLine, oracle, landName, landType, otherName, otherType string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := seedSearchLibrary(me,
		searchTestLand(landName, landType),
		searchTestLand(otherName, otherType),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, name, typeLine, oracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search = %+v, want only the %s", c, landName)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Errorf("the %s did not reach the hand", landName)
	}
}

func TestMomB02TypecyclersFindOnlyTheirLandType(t *testing.T) {
	momB02Typecycle(t, "Alabaster Host Intercessor", "Creature — Phyrexian Samurai", momB02IntercessorOracle,
		"Plains", "Basic Land — Plains", "Island", "Basic Land — Island")
	momB02Typecycle(t, "Furnace Host Charger", "Creature — Phyrexian Giant", momB02ChargerOracle,
		"Mountain", "Basic Land — Mountain", "Forest", "Basic Land — Forest")
	momB02Typecycle(t, "Timberland Ancient", "Creature — Treefolk", momB02AncientOracle,
		"Forest", "Basic Land — Forest", "Mountain", "Basic Land — Mountain")
	momB02Typecycle(t, "Tidal Terror", "Creature — Octopus", momB02TerrorOracle,
		"Island", "Basic Land — Island", "Plains", "Basic Land — Plains")
}

func TestMomB02ChargerHasHasteAndAncientReachAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	charger := b12Push(g, me.ID, "Furnace Host Charger", "Creature — Phyrexian Giant", momB02ChargerOracle, 5, 5)
	ancient := b12Push(g, me.ID, "Timberland Ancient", "Creature — Treefolk", momB02AncientOracle, 6, 5)
	if !hasKeywordOf(g, charger, "haste") {
		t.Error("Furnace Host Charger has no haste")
	}
	for _, kw := range []string{"reach", "trample"} {
		if !hasKeywordOf(g, ancient, kw) {
			t.Errorf("Timberland Ancient has no %s", kw)
		}
	}
}

func TestMomB02IntercessorExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	host := castCatalogSpell(t, g, "Alabaster Host Intercessor", "Creature — Phyrexian Samurai", momB02IntercessorOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Fatal("the opponent's creature was not exiled")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(host) {
		t.Fatal("the exile took something other than the target")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(host); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	back := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && c.IsCreature() {
			back++
		}
	}
	if back != 1 {
		t.Errorf("opponent creatures after the Intercessor left = %d, want the exiled one back (as a new object)", back)
	}
}

func TestMomB02TidalTerrorTapsTwoOthersToBecomeUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := pushCatalogPermanent(g, me.ID, "Tidal Terror", "Creature — Octopus", momB02TerrorOracle, false)
	a := b16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, terror)
	passPriorityAroundTable(t, g)
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("no tap-two prompt")
	}
	for _, id := range prompt.ChooseCards {
		if id == terror {
			t.Fatal("the prompt offers Tidal Terror itself")
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !b16Tapped(t, g, a) || !b16Tapped(t, g, b) {
		t.Error("the two creatures were not tapped")
	}
	if restrictionsOf(t, g, terror)&game.CantBeBlocked == 0 {
		t.Error("Tidal Terror can still be blocked")
	}
}

func TestMomB02TidalTerrorDeclinedStaysBlockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := pushCatalogPermanent(g, me.ID, "Tidal Terror", "Creature — Octopus", momB02TerrorOracle, false)
	a := b16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, terror)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			break
		}
	}
	if b16Tapped(t, g, a) {
		t.Error("a creature was tapped though the choice was declined")
	}
	if restrictionsOf(t, g, terror)&game.CantBeBlocked != 0 {
		t.Error("Tidal Terror became unblockable without tapping two creatures")
	}
}

func TestMomB02TidalTerrorWithOneOtherCreatureDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := pushCatalogPermanent(g, me.ID, "Tidal Terror", "Creature — Octopus", momB02TerrorOracle, false)
	b16Creature(g, me.ID, "Lone Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, terror)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			t.Fatal("a tap prompt opened with only one other creature")
		}
	}
	if restrictionsOf(t, g, terror)&game.CantBeBlocked != 0 {
		t.Error("Tidal Terror became unblockable")
	}
}

func TestMomB02CrystalCarapacePumpsAndWardsAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Crystal Carapace", "Enchantment — Aura", momB02CarapaceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the enchanted Bear is %d/%d, want 5/5", p, tough)
	}
	blade := castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward {2} payment for the opponent")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(blade) {
		t.Error("declining ward did not counter the spell")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	g2.WithWriteLock(func() { me2.Hand.Cards = nil })
	drawn := me2.Library.Size()
	id, _ := cycleFromHand(t, g2, "Crystal Carapace", "Enchantment — Aura", momB02CarapaceOracle, "{C}{C}")
	passPriorityAroundTable(t, g2)
	if !me2.Graveyard.Contains(id) || me2.Library.Size() != drawn-1 {
		t.Error("cycling did not discard the Aura and draw a card")
	}
}

func TestMomB02KorHalberdPumpsAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hal := pushCatalogPermanent(g, me.ID, "Kor Halberd", "Artifact — Equipment", momB02HalberdOracle, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	equipTo(t, g, me.ID, hal, bear)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOf(g, bear, "vigilance") {
		t.Error("the equipped Bear has no vigilance")
	}
}

func TestMomB02BanesplitterAttachesOnEntryForTwoPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eq := castCatalogSpell(t, g, "Mirran Banesplitter", "Artifact — Equipment", momB02BanesplitterOracl, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := e2Card(t, g, eq); c.AttachedTo.ID != bear {
		t.Fatalf("attached to %v, want the Bear", c.AttachedTo.ID)
	}
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 4/2", p, tough)
	}
}

func TestMomB02BattleFanAttachesAndGivesIndestructibleThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eq := castCatalogSpell(t, g, "Bladed Battle-Fan", "Artifact — Equipment", momB02FanOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := e2Card(t, g, eq); c.AttachedTo.ID != bear {
		t.Fatalf("attached to %v, want the Bear", c.AttachedTo.ID)
	}
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 3/2", p, tough)
	}
	if !hasKeywordOf(g, bear, "indestructible") {
		t.Error("the Bear is not indestructible this turn")
	}
}

func TestMomB02UrnOfGodfireMakesManaAndDestroys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	urn := pushCatalogPermanent(g, me.ID, "Urn of Godfire", "Artifact", momB02UrnOracle, false)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)

	if err := g.ActivateManaAbility(me.ID, urn, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the Urn made mana with nothing paid")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateManaAbility(me.ID, urn, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("pay {2} for a mana of any color: %v", err)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID)
	if pick == nil {
		t.Fatal("no color choice for the Urn's mana")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, urn, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foe}},
	}); err != nil {
		t.Fatalf("destroy ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(urn) {
		t.Error("the Urn was not sacrificed")
	}
	if g.Battlefield.Contains(foe) {
		t.Error("the creature was not destroyed")
	}
}

func TestMomB02FlywheelRacerMakesManaOnlyWhileACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	racer := b12Push(g, me.ID, "Flywheel Racer", "Artifact — Vehicle", momB02RacerOracle, 3, 2)
	if !hasKeywordOf(g, racer, "vigilance") {
		t.Error("Flywheel Racer has no vigilance")
	}
	if err := g.ActivateManaAbility(me.ID, racer, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("an uncrewed Racer tapped for mana")
	}
	crewer := b12Creature(g, me.ID, "Crewer", "Creature — Human", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, racer, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, racer, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("a crewed Racer could not tap for mana: %v", err)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceMana, me.ID) == nil {
		t.Error("the Racer's mana of any color asked for no color")
	}
}

func TestMomB02StasisFieldMakesItAZeroTwoDefenderWithNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Flyer", TypeLine: "Creature — Dragon",
		Power: 5, Toughness: 5, Keywords: []string{"flying", "trample"}, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Stasis Field", "Enchantment — Aura", momB02StasisOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, victim); p != 0 || tough != 2 {
		t.Errorf("size %d/%d, want 0/2", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, victim, "defender") {
		t.Error("no defender")
	}
	for _, kw := range []string{"flying", "trample"} {
		if hasKeywordOnBattlefield(t, g, victim, kw) {
			t.Errorf("still has %s", kw)
		}
	}
	if !spm8Live(t, g, victim).HasSubtype("Dragon") {
		t.Error("the creature types were changed; only abilities and size should be")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, victim, game.CounterPlusOne, 2); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if c := spm8Live(t, g, victim); c.CurrentPower() != 2 || c.CurrentToughness() != 4 {
		t.Errorf("with two counters %d/%d, want 2/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestMomB02RealmbreakersGraspStopsCombatAndNonManaAbilitiesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	feeder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Carrion Feeder", TypeLine: "Creature — Zombie",
		Power: 1, Toughness: 1, OracleID: carrionFeederOracle, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Realmbreaker's Grasp", "Enchantment — Aura", momB02GraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: feeder}})
	passPriorityAroundTable(t, g)
	got := restrictionsOf(t, g, feeder)
	if got&game.CantAttackOrBlock != game.CantAttackOrBlock || got&game.CantActivate == 0 {
		t.Errorf("restrictions = %v, want can't attack/block and can't activate", got.Names())
	}
	if got&game.CantActivateMana != 0 {
		t.Error("mana abilities were stopped; the card exempts them")
	}
	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{}); err != game.ErrCantActivate {
		t.Errorf("an enchanted permanent activated an ability: %v", err)
	}
}

func TestMomB02RealmbreakersGraspEnchantsAnArtifactAndRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	land := pushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	if err := castCatalogSpellErr(t, g, "Realmbreaker's Grasp", "Enchantment — Aura", momB02GraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Error("the Aura was cast on a land")
	}
	castCatalogSpell(t, g, "Realmbreaker's Grasp", "Enchantment — Aura", momB02GraspOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, rock)&game.CantActivate == 0 {
		t.Error("the enchanted artifact can still activate abilities")
	}
}
