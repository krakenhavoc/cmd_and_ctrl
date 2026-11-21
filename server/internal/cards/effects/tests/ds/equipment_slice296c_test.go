package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Helm of the Host ----------------------------------------------

func TestHelmOfTheHostCopiesEquippedCreatureAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	legend := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	helm := effectstest.SeedEquipment(g, me.ID, "Helm of the Host", effectstest.HelmOfTheHostOracle)
	effectstest.EquipTo(t, g, me.ID, helm, legend)

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)

	var token *game.Card
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Name == "Legend" && c.InstanceID != legend {
				token = &game.Card{TypeLine: c.TypeLine, Keywords: append([]string(nil), c.Keywords...)}
			}
		}
	})
	if token == nil {
		t.Fatal("no token copy of the equipped creature was created")
	}
	if hasLegendarySupertype(token.TypeLine) {
		t.Errorf("the token should not be legendary: %q", token.TypeLine)
	}
	found := false
	for _, kw := range token.Keywords {
		if kw == "haste" {
			found = true
		}
	}
	if !found {
		t.Error("the token should have haste")
	}
}

func hasLegendarySupertype(tl string) bool {
	return len(tl) >= len("Legendary") && tl[:len("Legendary")] == "Legendary"
}

// --- Commander's Plate ----------------------------------------------

func TestCommandersPlatePumpsAndCheapEquipIsCommanderOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plain := effectstest.SeedBear(g, me.ID)
	general := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "General", TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	plate := effectstest.SeedEquipment(g, me.ID, "Commander's Plate", effectstest.CommandersPlateOracle)

	err := g.ActivateCatalogAbility(me.ID, plate, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	})
	if err == nil {
		t.Fatal("the cheap equip should refuse a non-commander creature")
	}

	// Index 0 targeting the actual commander succeeds.
	if err := g.ActivateCatalogAbility(me.ID, plate, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: general}},
	}); err != nil {
		t.Fatalf("equip commander {3} on the actual commander: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, general); got != 5 {
		t.Errorf("power %d, want 5 (2 base + 3)", got)
	}

	// The plain {5} ability (index 1) has no such restriction.
	effectstest.EquipTo(t, g, me.ID, plate, plain)
	if host := effectstest.AttachmentHostOf(t, g, plate); host.ID != plain {
		t.Errorf("the unrestricted equip did not move the Plate: %+v", host)
	}
}

// --- The Reaver Cleaver ----------------------------------------------

func TestTheReaverCleaverCreatesTreasuresOnDamageToPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	cleaver := effectstest.SeedEquipment(g, me.ID, "The Reaver Cleaver", effectstest.TheReaverCleaverOracle)
	effectstest.EquipTo(t, g, me.ID, cleaver, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("missing trample")
	}

	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID, Counters: map[string]int{"loyalty": 5},
	})

	before := effectstest.CountTreasuresControlledBy(g, me.ID)
	effectstest.DealCombatDamageToPlayer(g, bear, walker, 3)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountTreasuresControlledBy(g, me.ID) - before; got != 3 {
		t.Errorf("made %d Treasures, want 3 (one per damage)", got)
	}
}

// --- Brotherhood Regalia ---------------------------------------------

func TestBrotherhoodRegaliaGrantsAssassinTypeAndUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	plain := effectstest.SeedBear(g, me.ID)
	legend := effectstest.SeedLegendaryCreature(g, me.ID, "General")
	regalia := effectstest.SeedEquipment(g, me.ID, "Brotherhood Regalia", effectstest.BrotherhoodRegaliaOracle)

	// The cheap "Equip legendary creature {1}" refuses a non-legendary
	// target.
	err := g.ActivateCatalogAbility(me.ID, regalia, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: plain}},
	})
	if err == nil {
		t.Fatal("equip legendary creature {1} accepted a non-legendary creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, regalia, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legend}},
	}); err != nil {
		t.Fatalf("equip legendary creature {1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	subs := effectstest.EffectiveSubtypes(t, g, legend)
	if !effectstest.ContainsString(subs, "Assassin") {
		t.Errorf("subtypes %v missing Assassin", subs)
	}
	effectstest.AssertRestrictions(t, g, legend, game.CantBeBlocked)
}

// --- Hammer of Nazahn --------------------------------------------------

func TestHammerOfNazahnAttachesItselfOnEntryAndGrantsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)

	hammer := effectstest.CastAndResolveCreature(t, g, "Hammer of Nazahn", "Legendary Artifact — Equipment", effectstest.HammerOfNazahnOracle)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, hammer); host.ID != bear {
		t.Fatalf("the Hammer did not attach itself on entry: %+v", host)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 (2 base + 2)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("missing indestructible")
	}
}

// --- Buster Sword --------------------------------------------------

func TestBusterSwordPumpsAndDrawsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Buster Sword", effectstest.BusterSwordOracle)
	effectstest.EquipTo(t, g, me.ID, sword, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5 (2 base + 3)", got)
	}

	handBefore := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand delta %d, want 1 — the free-cast half is a declared caveat, not implemented", got)
	}
}

// --- Tarrian's Soulcleaver -------------------------------------------

func TestTarriansSoulcleaverGrowsWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	cleaver := effectstest.SeedEquipment(g, me.ID, "Tarrian's Soulcleaver", effectstest.TarriansSoulcleaverOracle)
	effectstest.EquipTo(t, g, me.ID, cleaver, bear)

	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("missing vigilance")
	}

	other := effectstest.SeedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CurrentPowerFor(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1 counter)", got)
	}
}

// --- Conqueror's Flail -------------------------------------------------

func TestConquerorsFlailCountsColorsAndLocksOutOpponentsWhileAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	flail := effectstest.SeedEquipment(g, me.ID, "Conqueror's Flail", effectstest.ConquerorsFlailOracle)
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "White Thing", TypeLine: "Artifact", Colors: []string{"W"}})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Blue Thing", TypeLine: "Artifact", Colors: []string{"U"}})

	effectstest.EquipTo(t, g, me.ID, flail, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 (2 base + 2 colors)", got)
	}

	oppSpell := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: oppSpell, Name: "Their Spell", TypeLine: "Instant",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, oppSpell, game.CastSpellParams{}); err == nil {
		t.Error("opponent should not be able to cast a spell on my turn while the Flail is attached")
	}
}

// --- Embercleave --------------------------------------------------

func TestEmbercleaveCountsOnlyYourAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.EmbercleaveOracle, "Legendary Artifact — Equipment", "{4}{R}{R}"); got != 6 {
		t.Errorf("Embercleave with no attackers: %d, want 6", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Human", AttackingTarget: opp.ID})
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Human", AttackingTarget: me.ID})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.EmbercleaveOracle, "Legendary Artifact — Equipment", "{4}{R}{R}"); got != 5 {
		t.Errorf("Embercleave with one of my attackers (opponent's ignored): %d, want 5", got)
	}
}

func TestEmbercleaveEntersAttachedAndGrantsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)

	embercleave := effectstest.CastCatalogSpell(t, g, "Embercleave", "Legendary Artifact — Equipment", effectstest.EmbercleaveOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, embercleave); host.ID != bear {
		t.Fatalf("Embercleave did not attach to the target creature on entry: %+v", host)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1)", got)
	}
	for _, kw := range []string{"double strike", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("missing %s", kw)
		}
	}
}

// --- Caduceus, Staff of Hermes -----------------------------------------

func TestCaduceusGrantsLifelinkAlwaysAndTheRestOnlyAtThirtyLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	staff := effectstest.SeedEquipment(g, me.ID, "Caduceus, Staff of Hermes", effectstest.CaduceusOracle)
	effectstest.EquipTo(t, g, me.ID, staff, bear)

	me.Life = 29
	g.BumpLayerVersionForTest()
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("lifelink should always be granted")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("below 30 life: power %d, want 2 (no bonus yet)", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("below 30 life: should not be indestructible yet")
	}

	me.Life = 30
	g.BumpLayerVersionForTest()
	if got := effectstest.EffectivePower(t, g, bear); got != 7 {
		t.Errorf("at 30 life: power %d, want 7 (2 base + 5)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("at 30 life: should be indestructible")
	}
}

// --- Bloodforged Battle-Axe ---------------------------------------------

func TestBloodforgedBattleAxeCreatesACopyOfItselfOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	axe := effectstest.SeedEquipment(g, me.ID, "Bloodforged Battle-Axe", effectstest.BloodforgedBattleAxeOracle)
	effectstest.EquipTo(t, g, me.ID, axe, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}

	before := countPermanentsNamed(g, "Bloodforged Battle-Axe")
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)

	if got := countPermanentsNamed(g, "Bloodforged Battle-Axe") - before; got != 1 {
		t.Errorf("made %d copies, want 1", got)
	}
}

func countPermanentsNamed(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return n
}

// --- Sting, the Glinting Dagger -----------------------------------------

func TestStingUntapsEquippedCreatureAtEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	dagger := effectstest.SeedEquipment(g, me.ID, "Sting, the Glinting Dagger", effectstest.StingOracle)
	effectstest.EquipTo(t, g, me.ID, dagger, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("missing haste")
	}
	if !effectstest.IsTapped(g, bear) {
		t.Fatal("setup: bear should start tapped")
	}

	// My own combat.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
		t.Error("Sting should untap the equipped creature at the beginning of my combat")
	}

	// Re-tap, then walk into the NEXT player's combat — "each combat",
	// not "your combat".
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, bear) {
		t.Error("Sting should untap on an opponent's combat too — it says 'each combat'")
	}
}

func TestIllusionistsBracersCopiesActivatedAbilityOfEquippedCreature(t *testing.T) {
	g := effectstest.NewBigLibraryCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gris := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Griselbrand", TypeLine: "Legendary Creature — Demon",
		OracleID: effectstest.GriselbrandOracle, Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	bracers := effectstest.SeedEquipment(g, me.ID, "Illusionist's Bracers", effectstest.IllusionistsBracersOracle)
	effectstest.EquipTo(t, g, me.ID, bracers, gris)

	me.Life = 30
	lifeBefore, handBefore := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Griselbrand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if me.Life != lifeBefore-7 {
		t.Errorf("life %d -> %d, want -7 — the cost is paid once even though the ability resolves twice", lifeBefore, me.Life)
	}
	if got := me.Hand.Size() - handBefore; got != 14 {
		t.Errorf("hand delta %d, want 14 — the Bracers should copy the draw-seven", got)
	}
}

func TestIllusionistsBracersIgnoresAManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact Creature — Golem", effectstest.AcBasaltMonolithOracle, false)
	bracers := effectstest.SeedEquipment(g, me.ID, "Illusionist's Bracers", effectstest.IllusionistsBracersOracle)
	effectstest.EquipTo(t, g, me.ID, bracers, rock)

	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate mana ability: %v", err)
	}
	if effectstest.AcCopiesOnStack(g) != 0 {
		t.Error("a mana ability must not be copied by the Bracers")
	}
}
