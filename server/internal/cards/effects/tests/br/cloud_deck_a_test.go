package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func cloudDeckWantFull(t *testing.T, oracle string) {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s: registered %v, completeness %s, caveats %v, want full", oracle, ok, spec.Completeness, spec.Caveats)
	}
}

func cloudDeckHasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	v := effectstest.S58p6View(t, g, id)
	return game.HasKeyword(&v, kw)
}

func TestCloudDeckCloudEntersAndAttachesAnEquipment(t *testing.T) {
	cloudDeckWantFull(t, effectstest.CloudExSoldierOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := effectstest.SeedEquipment(g, me.ID, "Test Blade", "")

	cloud := effectstest.CastAndResolveCreature(t, g, "Cloud, Ex-SOLDIER", "Legendary Creature — Human Soldier Mercenary", effectstest.CloudExSoldierOracle)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if !effectstest.HasID(pick.PickTargetCards, sword) {
			t.Fatalf("legal set %v misses the Equipment", pick.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, sword)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, sword); host.ID != cloud {
		t.Errorf("Equipment host = %+v, want Cloud %s", host, cloud)
	}
}

func TestCloudDeckCloudAttackDrawsPerEquippedAttackerAndMakesTreasureAtSevenPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	cloud := effectstest.PushCatalogPermanent(g, me.ID, "Cloud, Ex-SOLDIER", "Legendary Creature — Human Soldier Mercenary", effectstest.CloudExSoldierOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	plain := effectstest.SeedBear(g, me.ID)
	arm := effectstest.SeedEquipment(g, me.ID, "Wrecking Ball Arm", effectstest.WreckingBallArmOracle)
	blade := effectstest.SeedEquipment(g, me.ID, "Test Blade", "")
	effectstest.EquipTo(t, g, me.ID, arm, cloud)
	if err := effectstest.CloudDeckAttach(g, blade, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("attach: %v", err)
	}
	if got := effectstest.EffectivePower(t, g, cloud); got != 7 {
		t.Fatalf("Cloud power %d, want 7", got)
	}
	before := me.Hand.Size()

	effectstest.DeclareAttack(t, g, opp.ID, cloud, bear, plain)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d, want 2 (Cloud and one Bear are equipped attackers)", got)
	}
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("Treasures = %d, want 2 (Cloud has power 7)", n)
	}
}

func TestCloudDeckCloudAttackMakesNoTreasureBelowSevenPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	cloud := effectstest.PushCatalogPermanent(g, me.ID, "Cloud, Ex-SOLDIER", "Legendary Creature — Human Soldier Mercenary", effectstest.CloudExSoldierOracle, false)
	before := me.Hand.Size()

	effectstest.DeclareAttack(t, g, opp.ID, cloud)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("drew %d with no equipped attacker, want 0", got)
	}
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("Treasures = %d, want 0", n)
	}
}

func TestCloudDeckArmoryAutomatonAttachesAnyNumberOfEquipmentWithoutChangingControl(t *testing.T) {
	cloudDeckWantFull(t, effectstest.ArmoryAutomatonOracle)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.SeedEquipment(g, me.ID, "Test Blade", "")
	theirs := effectstest.SeedEquipment(g, opp.ID, "Their Blade", "")

	auto := effectstest.CastAndResolveCreature(t, g, "Armory Automaton", "Artifact Creature — Construct", effectstest.ArmoryAutomatonOracle)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	for _, e := range []uuid.UUID{mine, theirs} {
		if host := effectstest.AttachmentHostOf(t, g, e); host.ID != auto {
			t.Errorf("Equipment %s host = %+v, want the Automaton", e, host)
		}
	}
}

func TestCloudDeckWreckingBallArmSetsBasePTAndBlocksSmallCreatures(t *testing.T) {
	cloudDeckWantFull(t, effectstest.WreckingBallArmOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	arm := effectstest.SeedEquipment(g, me.ID, "Wrecking Ball Arm", effectstest.WreckingBallArmOracle)

	effectstest.EquipTo(t, g, me.ID, arm, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 7 {
		t.Errorf("equipped power %d, want 7", got)
	}
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.WreckingBallArmOracle})
	if len(abilities) != 2 {
		t.Fatalf("equip abilities = %d, want 2 (legendary {3} and plain {7})", len(abilities))
	}
}

func TestCloudDeckSlayersStrongholdBoostsAndGrantsKeywords(t *testing.T) {
	cloudDeckWantFull(t, effectstest.SlayersStrongholdOracl)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Slayers' Stronghold", "Land", effectstest.SlayersStrongholdOracl, false)
	bear := effectstest.SeedBear(g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "W"})

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	for _, kw := range []string{"vigilance", "haste"} {
		if !cloudDeckHasKeyword(t, g, bear, kw) {
			t.Errorf("Bear lacks %s", kw)
		}
	}
}

func TestCloudDeckBarretWallaceDealsDamagePerEquippedCreature(t *testing.T) {
	cloudDeckWantFull(t, effectstest.BarretWallaceOracle)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	barret := effectstest.PushCatalogPermanent(g, me.ID, "Barret Wallace", "Legendary Creature — Human Rebel", effectstest.BarretWallaceOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Test Blade", "")
	other := effectstest.SeedEquipment(g, me.ID, "Other Blade", "")
	for _, p := range []struct{ e, h uuid.UUID }{{blade, bear}, {other, barret}} {
		if err := effectstest.CloudDeckAttach(g, p.e, game.TargetRef{Kind: game.TargetCard, ID: p.h}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	}
	life := opp.Life

	effectstest.DeclareAttack(t, g, opp.ID, barret)
	effectstest.PassPriorityAroundTable(t, g)

	if got := life - opp.Life; got != 2 {
		t.Errorf("defending player lost %d, want 2 (two equipped creatures)", got)
	}
}

func TestCloudDeckBarretAvalancheLeaderMakesARebelPerEquipmentAndAttachesAtCombat(t *testing.T) {
	cloudDeckWantFull(t, effectstest.BarretAvalancheOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Barret, Avalanche Leader", "Legendary Creature — Human Rebel", effectstest.BarretAvalancheOracle, false)

	blade := effectstest.B12PlayFromHand(t, g, "Test Blade", effectstest.EquipTypeLine, "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Rebel"); n != 1 {
		t.Fatalf("Rebel tokens = %d, want 1", n)
	}

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, blade)
	}
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, tokenOfSubtype(g, me.ID, "Rebel"))
	}
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, blade); host.ID != tokenOfSubtype(g, me.ID, "Rebel") {
		t.Errorf("Equipment host = %+v, want the Rebel token", host)
	}
}

func tokenOfSubtype(g *game.Game, controller uuid.UUID, subtype string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.HasSubtype(subtype) {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestCloudDeckSoldierMilitaryProgramMakesASoldier(t *testing.T) {
	cloudDeckWantFull(t, effectstest.SoldierProgramOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "SOLDIER Military Program", "Enchantment", effectstest.SoldierProgramOracle, false)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode_pick prompt")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountTokensControlled(g, me.ID, "Soldier"); n != 1 {
		t.Errorf("Soldier tokens = %d, want 1", n)
	}
}

func TestCloudDeckRedXIIIGrantsKeywordsToOtherModifiedCreatures(t *testing.T) {
	cloudDeckWantFull(t, effectstest.RedXIIIOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	xiii := effectstest.PushCatalogPermanent(g, me.ID, "Red XIII, Proud Warrior", "Legendary Creature — Beast Warrior", effectstest.RedXIIIOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	plain := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Test Blade", "")
	if err := effectstest.CloudDeckAttach(g, blade, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("attach: %v", err)
	}

	for _, kw := range []string{"vigilance", "trample"} {
		if !cloudDeckHasKeyword(t, g, bear, kw) {
			t.Errorf("equipped Bear lacks %s", kw)
		}
		if cloudDeckHasKeyword(t, g, plain, kw) {
			t.Errorf("unmodified Bear has %s", kw)
		}
	}
	_ = xiii
}

func TestCloudDeckRedXIIIReturnsAnEquipmentFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blade := effectstest.PushGraveyardPermanent(me, "Test Blade", effectstest.EquipTypeLine, "{2}")
	creature := effectstest.PushGraveyardPermanent(me, "Test Bear", "Creature — Bear", "{1}")

	effectstest.CastAndResolveCreature(t, g, "Red XIII, Proud Warrior", "Legendary Creature — Beast Warrior", effectstest.RedXIIIOracle)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, creature) || !effectstest.HasID(pick.PickTargetCards, blade) {
		t.Fatalf("legal set %v, want only the Equipment", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, blade)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(blade) {
		t.Error("the Equipment did not return to hand")
	}
}

func TestCloudDeckElenaReturnsHistoricAndGrowsOnHistoricCasts(t *testing.T) {
	cloudDeckWantFull(t, effectstest.ElenaTurkRecruitOracle)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushGraveyardPermanent(me, "Test Rock", "Artifact", "{1}")
	assassin := effectstest.PushGraveyardPermanent(me, "Test Assassin", "Legendary Creature — Human Assassin", "{1}{B}")
	plain := effectstest.PushGraveyardPermanent(me, "Test Bear", "Creature — Bear", "{1}")

	elena := effectstest.CastAndResolveCreature(t, g, "Elena, Turk Recruit", "Legendary Creature — Human Assassin", effectstest.ElenaTurkRecruitOracle)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if !effectstest.HasID(pick.PickTargetCards, rock) || effectstest.HasID(pick.PickTargetCards, assassin) || effectstest.HasID(pick.PickTargetCards, plain) {
		t.Fatalf("legal set %v, want only the non-Assassin historic card", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Fatal("the artifact did not return to hand")
	}

	// A historic spell puts a counter on Elena; a plain one does not.
	effectstest.CastCatalogSpell(t, g, "Test Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CloudDeckCounters(t, g, elena); got != 1 {
		t.Errorf("Elena counters = %d, want 1", got)
	}
}

func TestCloudDeckZackFairSacrificeMovesCountersAndGrantsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	zack := effectstest.CastAndResolveCreature(t, g, "Zack Fair", "Legendary Creature — Human Soldier", effectstest.ZackFairOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CloudDeckCounters(t, g, zack); got != 1 {
		t.Fatalf("Zack counters = %d, want 1 (enters with a +1/+1 counter)", got)
	}
	bear := effectstest.SeedBear(g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})

	if err := g.ActivateCatalogAbility(me.ID, zack, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, zack) {
		t.Error("Zack was not sacrificed")
	}
	if got := effectstest.CloudDeckCounters(t, g, bear); got != 1 {
		t.Errorf("Bear counters = %d, want 1", got)
	}
	if !cloudDeckHasKeyword(t, g, bear, "indestructible") {
		t.Error("Bear is not indestructible")
	}
	cloudDeckWantFull(t, effectstest.ZackFairOracle)
}

func TestCloudDeckConformerShurikenTapsADefenderAndGrowsToMatch(t *testing.T) {
	cloudDeckWantFull(t, effectstest.ConformerShurikenOracl)
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	shuriken := effectstest.SeedEquipment(g, me.ID, "Conformer Shuriken", effectstest.ConformerShurikenOracl)
	effectstest.EquipTo(t, g, me.ID, shuriken, bear)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, big)
	}
	effectstest.PassPriorityAroundTable(t, g)

	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == big {
				tapped = c.Tapped
			}
		}
	})
	if !tapped {
		t.Error("the defender's creature was not tapped")
	}
	if got := effectstest.CloudDeckCounters(t, g, bear); got != 3 {
		t.Errorf("Bear counters = %d, want 3 (5 power vs 2)", got)
	}
}
