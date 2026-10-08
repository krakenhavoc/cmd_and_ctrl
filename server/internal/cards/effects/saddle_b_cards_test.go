package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// saddle_b_cards_test.go — slice saddle-b (#2695): one test per Mount
// (and Kolodin and the Armory) built on the saddle seam of #2714.

const (
	gilaCourserOracle    = "bc983c67-eb5d-4a85-a12e-a410f85e949e"
	dracosaurAuxOracle   = "457e3218-b2ef-4280-931a-5320987b9657"
	lagorinOracle        = "f7b6f5d4-b122-4ab4-a2c2-449ac781359a"
	congregationGryffOr  = "376cbc85-2941-4d97-b031-bffff86550e7"
	brightfieldMustangOr = "1b05e355-38a5-4307-9d8b-e99ec30f8ed3"
	unswervingSlothOr    = "5e512ce9-5e2e-4ec3-9486-a10c63cff299"
	bridledBighornOracle = "6a1ad848-6fda-46dd-aad8-2f6a5c71a696"
	autarchMammothOracle = "eb50b5f2-4840-4f90-9c85-40fc581be9c7"
	trainedArynxOracle   = "bc875500-f6df-4365-b2d6-7cf425daea97"
	archmagesNewtOracle  = "75ffebc4-8db9-4de6-a330-e3f41cdccecc"
	alacrianArmoryOracle = "3fa2d51c-d549-4287-ac3b-76bc72bcd8bc"
	kolodinOracle        = "45682055-d02a-450c-a41d-aa67162ff015"
	quilledChargerOracle = "6498e3e9-1a7b-4b47-9e7c-cdbd8928c9c8"
	brightfieldGliderOr  = "acff6e5c-7f25-4ae1-8672-8672450ad844"
	alacrianJaguarOracle = "caeb4ec1-a8ba-4ac4-ad32-edbc671e5a39"
	venomsacLagacOracle  = "2fac62be-dcbe-409f-a362-f0e519ec14a9"
)

// The four pure pump Mounts: saddled attack gives the printed bonus,
// an unsaddled one gives nothing.
func TestSaddleBPumpMounts(t *testing.T) {
	cases := []struct {
		name    string
		oracle  string
		saddle  int
		dp, dt  int
		keyword string
	}{
		{"Quilled Charger", quilledChargerOracle, 2, 1, 2, "menace"},
		{"Brightfield Glider", brightfieldGliderOr, 3, 1, 2, "flying"},
		{"Alacrian Jaguar", alacrianJaguarOracle, 1, 2, 2, ""},
		{"Venomsac Lagac", venomsacLagacOracle, 2, 0, 3, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name+" saddled", func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mount := pushMountForTest(g, me.ID, tc.name, tc.oracle, 3)
			saddler := pushCrewerForTest(g, me.ID, "Saddler", tc.saddle)
			saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
			passPriorityAroundTable(t, g)
			if p, th := effectivePower(t, g, mount), effectiveToughness(t, g, mount); p != 3+tc.dp || th != 3+tc.dt {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, th, 3+tc.dp, 3+tc.dt)
			}
			if tc.keyword != "" && !effectiveAbilitiesContain(t, g, mount, tc.keyword) {
				t.Errorf("%s did not gain %s", tc.name, tc.keyword)
			}
		})
		t.Run(tc.name+" unsaddled", func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mount := pushMountForTest(g, me.ID, tc.name, tc.oracle, 3)
			declareAttack(t, g, opp.ID, mount)
			passPriorityAroundTable(t, g)
			if p, th := effectivePower(t, g, mount), effectiveToughness(t, g, mount); p != 3 || th != 3 {
				t.Errorf("an unsaddled %s is %d/%d, want 3/3", tc.name, p, th)
			}
		})
	}
}

func TestCongregationGryffCountsEveryMountYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gryff := pushMountForTest(g, me.ID, "Congregation Gryff", congregationGryffOr, 1)
	pushMountForTest(g, me.ID, "Other Mount", gildedGhodaOracle, 2)
	pushMountForTest(g, opp.ID, "Their Mount", gildedGhodaOracle, 2)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 3)

	saddleThenAttack(t, g, me.ID, opp.ID, gryff, saddler)
	passPriorityAroundTable(t, g)
	// Two Mounts are mine, the Gryff and the other; X = 2.
	if p, th := effectivePower(t, g, gryff), effectiveToughness(t, g, gryff); p != 3 || th != 3 {
		t.Errorf("Gryff is %d/%d, want 3/3 (1/1 plus +2/+2)", p, th)
	}
}

func TestBrightfieldMustangUntapsItAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Brightfield Mustang", brightfieldMustangOr, 3)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	if !saddleCard(t, g, mount).Tapped {
		t.Fatal("setup: the attacking Mustang should be tapped")
	}
	passPriorityAroundTable(t, g)
	if saddleCard(t, g, mount).Tapped {
		t.Error("the Mustang was not untapped")
	}
	if got := countersOf(t, g, mount); got != 1 {
		t.Errorf("the Mustang has %d counters, want 1", got)
	}
}

func TestUnswervingSlothIsIndestructibleAndUntapsEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Unswerving Sloth", unswervingSlothOr, 5)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 4)
	theirs := pushCrewerForTest(g, opp.ID, "Theirs", 1)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == theirs {
			g.Battlefield.Cards[i].Tapped = true
		}
	}

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	if !saddleCard(t, g, saddler).Tapped {
		t.Fatal("setup: the saddler should be tapped")
	}
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, mount, "indestructible") {
		t.Error("the Sloth did not gain indestructible")
	}
	if saddleCard(t, g, saddler).Tapped {
		t.Error("the creature that saddled it was not untapped")
	}
	if !saddleCard(t, g, theirs).Tapped {
		t.Error("an opponent's creature was untapped")
	}
}

func TestBridledBighornMakesASheepWhenItAttacksSaddled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Bridled Bighorn", bridledBighornOracle, 3)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 2)

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)
	sheep := saddleBNamed(g, "Sheep", me.ID)
	if len(sheep) != 1 {
		t.Fatalf("%d Sheep, want 1", len(sheep))
	}
	if sheep[0].CurrentPower() != 1 || sheep[0].CurrentToughness() != 1 {
		t.Errorf("Sheep is %d/%d, want 1/1", sheep[0].CurrentPower(), sheep[0].CurrentToughness())
	}
}

func saddleBNamed(g *game.Game, name string, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

func TestAutarchMammothMakesAnElephantWhenItEntersAndWhenItAttacksSaddled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Autarch Mammoth", "Creature — Elephant Mount", autarchMammothOracle, nil)
	passPriorityAroundTable(t, g)
	if n := len(saddleBNamed(g, "Elephant", me.ID)); n != 1 {
		t.Fatalf("%d Elephants after entering, want 1", n)
	}

	// The attack half: a Mammoth that is not summoning sick.
	attacker := pushMountForTest(g, me.ID, "Autarch Mammoth", autarchMammothOracle, 5)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 5)
	saddleThenAttack(t, g, me.ID, opp.ID, attacker, saddler)
	passPriorityAroundTable(t, g)
	elephants := saddleBNamed(g, "Elephant", me.ID)
	if len(elephants) != 2 {
		t.Fatalf("%d Elephants after the saddled attack, want 2", len(elephants))
	}
	if elephants[0].CurrentPower() != 3 || elephants[0].CurrentToughness() != 3 {
		t.Errorf("Elephant is %d/%d, want 3/3", elephants[0].CurrentPower(), elephants[0].CurrentToughness())
	}
}

func TestTrainedArynxGainsFirstStrikeAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Trained Arynx", trainedArynxOracle, 3)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 2)
	pushLibraryCard(me, game.Card{Name: "Top", TypeLine: "Sorcery"})

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)
	if !effectiveAbilitiesContain(t, g, mount, "first strike") {
		t.Error("the Arynx did not gain first strike")
	}
	scrying := false
	for _, c := range g.PendingChoices {
		if c.Chooser == me.ID {
			scrying = true
		}
	}
	if !scrying {
		t.Error("no scry prompt was queued")
	}
}

func TestGilaCourserExilesTheTopCardAndLetsYouPlayIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Gila Courser", gilaCourserOracle, 4)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)
	top := seedTopOfLibrary(me, game.Card{Name: "Impulse Target", TypeLine: "Basic Land — Forest"})

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)
	if me.Library.Contains(top) {
		t.Fatal("the top card is still in the library")
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if grantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("the exiled card carries no permission to play it")
	}
}

func TestDracosaurAuxiliaryDealsTwoToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Dracosaur Auxiliary", dracosaurAuxOracle, 4)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 3)
	victim := pushCrewerForTest(g, opp.ID, "Victim", 5)

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := saddleCard(t, g, victim).DamageMarked; got != 2 {
		t.Errorf("the target has %d damage, want 2", got)
	}
}

func TestLagorinPutsACounterOnEachChosenMountOrVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Lagorin, Soul of Alacria", lagorinOracle, 1)
	other := pushMountForTest(g, me.ID, "Other Mount", gildedGhodaOracle, 2)
	bear := pushCrewerForTest(g, me.ID, "Not A Mount", 1)
	saddler := pushCrewerForTest(g, me.ID, "Saddler", 1)

	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mount}, {Kind: game.TargetCard, ID: other},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOf(t, g, mount) != 1 || countersOf(t, g, other) != 1 {
		t.Errorf("counters = %d and %d, want 1 each", countersOf(t, g, mount), countersOf(t, g, other))
	}
	if countersOf(t, g, bear) != 0 {
		t.Error("a creature that is neither Mount nor Vehicle got a counter")
	}
}

func TestArchmagesNewtGrantsFlashbackAtManaCostOrFreeWhenSaddled(t *testing.T) {
	for _, saddled := range []bool{false, true} {
		name := "unsaddled"
		if saddled {
			name = "saddled"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			newt := pushMountForTest(g, me.ID, "Archmage's Newt", archmagesNewtOracle, 2)
			spell := pushGraveyardCardTyped(me, "Brainstorm", "Instant")
			if saddled {
				g.WithWriteLock(func() { g.SaddleForEffect(newt, nil) })
			}
			dealCombatDamageToPlayer(g, newt, opp.ID, 2)
			pickTriggerTarget(t, g, me.ID, spell)
			passPriorityAroundTable(t, g)
			grant := grantedPermissionOn(g, me.ID, spell, game.ZoneGraveyard)
			if grant == nil {
				t.Fatal("the graveyard card did not gain flashback")
			}
			if grant.AltCostKey != "flashback" {
				t.Errorf("alt cost key %q, want flashback", grant.AltCostKey)
			}
			want := ""
			if saddled {
				want = "{0}"
			}
			if grant.Cost != want {
				t.Errorf("flashback cost override %q, want %q", grant.Cost, want)
			}
		})
	}
}

func TestAlacrianArmoryBuffsCreaturesAndSaddlesAMountAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Alacrian Armory", TypeLine: "Artifact", OracleID: alacrianArmoryOracle,
		Owner: me.ID, Controller: me.ID,
	})
	mount := pushMountForTest(g, me.ID, "Gilded Ghoda", gildedGhodaOracle, 2)

	if th := effectiveToughness(t, g, mount); th != 3 {
		t.Errorf("toughness %d, want 3 (2 plus the Armory's +0/+1)", th)
	}
	if !effectiveAbilitiesContain(t, g, mount, "vigilance") {
		t.Error("a creature you control lacks vigilance")
	}
	advanceTo(t, g, game.StepBeginCombat)
	pickTriggerTarget(t, g, me.ID, mount)
	passPriorityAroundTable(t, g)
	if !saddleCard(t, g, mount).Saddled {
		t.Error("the Mount did not become saddled")
	}
}

func TestKolodinGivesHasteAndSaddlesAMountThatEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kolodin := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kolodin, Triumph Caster", TypeLine: "Legendary Creature — Human Pilot",
		OracleID: kolodinOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if effectiveAbilitiesContain(t, g, kolodin, "haste") {
		t.Error("Kolodin gave himself haste")
	}
	castCatalogSpell(t, g, "Gilded Ghoda", "Creature — Horse Mount", gildedGhodaOracle, nil)
	passPriorityAroundTable(t, g)
	var mount uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Gilded Ghoda" {
			mount = c.InstanceID
		}
	}
	if mount == uuid.Nil {
		t.Fatal("the Mount did not enter")
	}
	if !saddleCard(t, g, mount).Saddled {
		t.Error("the Mount that entered did not become saddled")
	}
	if !effectiveAbilitiesContain(t, g, mount, "haste") {
		t.Error("the Mount did not gain haste")
	}
}

func TestKolodinMakesAnEnteringVehicleAnArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kolodin, Triumph Caster", TypeLine: "Legendary Creature — Human Pilot",
		OracleID: kolodinOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	cart := castCatalogSpell(t, g, "Test Cart", "Artifact — Vehicle", "", nil)
	passPriorityAroundTable(t, g)
	found := false
	for _, ty := range effectiveTypes(t, g, cart) {
		if ty == "Creature" {
			found = true
		}
	}
	if !found {
		t.Errorf("the entering Vehicle's types are %v, want Creature among them", effectiveTypes(t, g, cart))
	}
}
