package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// equipment_slice296c_test.go covers roadmap slice 296-c (Equipment):
// Helm of the Host, Commander's Plate, The Reaver Cleaver, Brotherhood
// Regalia, Hammer of Nazahn, Buster Sword, Tarrian's Soulcleaver,
// Conqueror's Flail, Embercleave, Caduceus Staff of Hermes, Bloodforged
// Battle-Axe, Sting the Glinting Dagger, and Illusionist's Bracers.
// Reconfigure (The Reality Chip, Lizard Blades) is not implemented and
// both cards are left out of the catalog — see docs/engine-seams.md.

const (
	helmOfTheHostOracle        = "83b43aba-bf9c-4da2-967d-9daa632e97d2"
	commandersPlateOracle      = "cae166de-e681-40a0-83a8-3c17cf40e2fc"
	theReaverCleaverOracle     = "37a2a31d-51e6-4c07-b4c2-b206ad40eb42"
	brotherhoodRegaliaOracle   = "d169738e-af0b-4488-a399-3aa1d1933aa1"
	hammerOfNazahnOracle       = "e3955573-3db5-490e-903e-65e0172a9202"
	busterSwordOracle          = "5e060d58-4d6e-425c-b7d4-727669fcce5b"
	tarriansSoulcleaverOracle  = "2ffb38ec-5852-4e91-85a5-cfccd1f23556"
	conquerorsFlailOracle      = "9aace0d3-89e6-4254-b96f-ee3a878f2f91"
	embercleaveOracle          = "4d6120d6-fcce-40bc-9fc6-e1f5beb6c728"
	caduceusOracle             = "7cea5b12-9483-4142-9efa-735305581e73"
	bloodforgedBattleAxeOracle = "81a7f559-5edd-47ec-91f7-51e5ee998ed0"
	stingOracle                = "973c49a1-8425-40b9-8bdc-2c2434222314"
	illusionistsBracersOracle  = "1d1d78af-7982-419d-b9be-2bf4c149d97d"
)

// --- Helm of the Host ----------------------------------------------

func TestHelmOfTheHostCopiesEquippedCreatureAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	helm := seedEquipment(g, me.ID, "Helm of the Host", helmOfTheHostOracle)
	equipTo(t, g, me.ID, helm, legend)

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plain := seedBear(g, me.ID)
	general := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "General", TypeLine: "Legendary Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	plate := seedEquipment(g, me.ID, "Commander's Plate", commandersPlateOracle)

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
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, general); got != 5 {
		t.Errorf("power %d, want 5 (2 base + 3)", got)
	}

	// The plain {5} ability (index 1) has no such restriction.
	equipTo(t, g, me.ID, plate, plain)
	if host := attachmentHostOf(t, g, plate); host.ID != plain {
		t.Errorf("the unrestricted equip did not move the Plate: %+v", host)
	}
}

// --- The Reaver Cleaver ----------------------------------------------

func TestTheReaverCleaverCreatesTreasuresOnDamageToPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	cleaver := seedEquipment(g, me.ID, "The Reaver Cleaver", theReaverCleaverOracle)
	equipTo(t, g, me.ID, cleaver, bear)

	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("missing trample")
	}

	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID, Counters: map[string]int{"loyalty": 5},
	})

	before := countTreasuresControlledBy(g, me.ID)
	dealCombatDamageToPlayer(g, bear, walker, 3)
	passPriorityAroundTable(t, g)

	if got := countTreasuresControlledBy(g, me.ID) - before; got != 3 {
		t.Errorf("made %d Treasures, want 3 (one per damage)", got)
	}
}

// --- Brotherhood Regalia ---------------------------------------------

func TestBrotherhoodRegaliaGrantsAssassinTypeAndUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plain := seedBear(g, me.ID)
	legend := seedLegendaryCreature(g, me.ID, "General")
	regalia := seedEquipment(g, me.ID, "Brotherhood Regalia", brotherhoodRegaliaOracle)

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
	passPriorityAroundTable(t, g)

	subs := effectiveSubtypes(t, g, legend)
	if !containsString(subs, "Assassin") {
		t.Errorf("subtypes %v missing Assassin", subs)
	}
	assertRestrictions(t, g, legend, game.CantBeBlocked)
}

// --- Hammer of Nazahn --------------------------------------------------

func TestHammerOfNazahnAttachesItselfOnEntryAndGrantsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)

	hammer := castAndResolveCreature(t, g, "Hammer of Nazahn", "Legendary Artifact — Equipment", hammerOfNazahnOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, hammer); host.ID != bear {
		t.Fatalf("the Hammer did not attach itself on entry: %+v", host)
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4 (2 base + 2)", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("missing indestructible")
	}
}

// --- Buster Sword --------------------------------------------------

func TestBusterSwordPumpsAndDrawsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	sword := seedEquipment(g, me.ID, "Buster Sword", busterSwordOracle)
	equipTo(t, g, me.ID, sword, bear)

	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5 (2 base + 3)", got)
	}

	handBefore := me.Hand.Size()
	dealCombatDamageToPlayer(g, bear, opp.ID, 5)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand delta %d, want 1 — the free-cast half is a declared caveat, not implemented", got)
	}
}

// --- Tarrian's Soulcleaver -------------------------------------------

func TestTarriansSoulcleaverGrowsWhenAnotherCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	cleaver := seedEquipment(g, me.ID, "Tarrian's Soulcleaver", tarriansSoulcleaverOracle)
	equipTo(t, g, me.ID, cleaver, bear)

	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("missing vigilance")
	}

	other := seedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	passPriorityAroundTable(t, g)

	if got := currentPowerFor(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1 counter)", got)
	}
}

// currentPowerFor reads Card.CurrentPower() — the layer-effective
// power PLUS any +1/+1 / -1/-1 counters (CR 613's counters step,
// which effectivePower's plain Effective().Power deliberately
// excludes; counter math is never modelled as a layer, per AGENTS.md
// §7).
func currentPowerFor(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var p int
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				p, found = c.CurrentPower(), true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return p
}

// --- Conqueror's Flail -------------------------------------------------

func TestConquerorsFlailCountsColorsAndLocksOutOpponentsWhileAttached(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	flail := seedEquipment(g, me.ID, "Conqueror's Flail", conquerorsFlailOracle)
	pushPermanent(g, me.ID, game.Card{Name: "White Thing", TypeLine: "Artifact", Colors: []string{"W"}})
	pushPermanent(g, me.ID, game.Card{Name: "Blue Thing", TypeLine: "Artifact", Colors: []string{"U"}})

	equipTo(t, g, me.ID, flail, bear)
	if got := effectivePower(t, g, bear); got != 4 {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := selfPricedMV(t, g, me, embercleaveOracle, "Legendary Artifact — Equipment", "{4}{R}{R}"); got != 6 {
		t.Errorf("Embercleave with no attackers: %d, want 6", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Human", AttackingTarget: opp.ID})
	pushPermanent(g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Human", AttackingTarget: me.ID})
	if got := selfPricedMV(t, g, me, embercleaveOracle, "Legendary Artifact — Equipment", "{4}{R}{R}"); got != 5 {
		t.Errorf("Embercleave with one of my attackers (opponent's ignored): %d, want 5", got)
	}
}

func TestEmbercleaveEntersAttachedAndGrantsKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)

	embercleave := castCatalogSpell(t, g, "Embercleave", "Legendary Artifact — Equipment", embercleaveOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if latestPickTarget(g, me.ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, embercleave); host.ID != bear {
		t.Fatalf("Embercleave did not attach to the target creature on entry: %+v", host)
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (2 base + 1)", got)
	}
	for _, kw := range []string{"double strike", "trample"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("missing %s", kw)
		}
	}
}

// --- Caduceus, Staff of Hermes -----------------------------------------

func TestCaduceusGrantsLifelinkAlwaysAndTheRestOnlyAtThirtyLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	staff := seedEquipment(g, me.ID, "Caduceus, Staff of Hermes", caduceusOracle)
	equipTo(t, g, me.ID, staff, bear)

	me.Life = 29
	g.BumpLayerVersionForTest()
	if !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("lifelink should always be granted")
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("below 30 life: power %d, want 2 (no bonus yet)", got)
	}
	if hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("below 30 life: should not be indestructible yet")
	}

	me.Life = 30
	g.BumpLayerVersionForTest()
	if got := effectivePower(t, g, bear); got != 7 {
		t.Errorf("at 30 life: power %d, want 7 (2 base + 5)", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("at 30 life: should be indestructible")
	}
}

// --- Bloodforged Battle-Axe ---------------------------------------------

func TestBloodforgedBattleAxeCreatesACopyOfItselfOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	axe := seedEquipment(g, me.ID, "Bloodforged Battle-Axe", bloodforgedBattleAxeOracle)
	equipTo(t, g, me.ID, axe, bear)

	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power %d, want 4", got)
	}

	before := countPermanentsNamed(g, "Bloodforged Battle-Axe")
	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	dagger := seedEquipment(g, me.ID, "Sting, the Glinting Dagger", stingOracle)
	equipTo(t, g, me.ID, dagger, bear)

	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("missing haste")
	}
	if !isTapped(g, bear) {
		t.Fatal("setup: bear should start tapped")
	}

	// My own combat.
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Error("Sting should untap the equipped creature at the beginning of my combat")
	}

	// Re-tap, then walk into the NEXT player's combat — "each combat",
	// not "your combat".
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	advanceTo(t, g, game.StepEndCombat)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Error("Sting should untap on an opponent's combat too — it says 'each combat'")
	}
}

// --- Illusionist's Bracers -----------------------------------------------

func newBigLibraryCatalogGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 40)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func TestIllusionistsBracersCopiesActivatedAbilityOfEquippedCreature(t *testing.T) {
	g := newBigLibraryCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gris := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Griselbrand", TypeLine: "Legendary Creature — Demon",
		OracleID: griselbrandOracle, Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID,
	})
	bracers := seedEquipment(g, me.ID, "Illusionist's Bracers", illusionistsBracersOracle)
	equipTo(t, g, me.ID, bracers, gris)

	me.Life = 30
	lifeBefore, handBefore := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gris, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Griselbrand: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Life != lifeBefore-7 {
		t.Errorf("life %d -> %d, want -7 — the cost is paid once even though the ability resolves twice", lifeBefore, me.Life)
	}
	if got := me.Hand.Size() - handBefore; got != 14 {
		t.Errorf("hand delta %d, want 14 — the Bracers should copy the draw-seven", got)
	}
}

func TestIllusionistsBracersIgnoresAManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := pushCatalogPermanent(g, me.ID, "Basalt Monolith", "Artifact Creature — Golem", acBasaltMonolithOracle, false)
	bracers := seedEquipment(g, me.ID, "Illusionist's Bracers", illusionistsBracersOracle)
	equipTo(t, g, me.ID, bracers, rock)

	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate mana ability: %v", err)
	}
	if acCopiesOnStack(g) != 0 {
		t.Error("a mana ability must not be copied by the Bracers")
	}
}
