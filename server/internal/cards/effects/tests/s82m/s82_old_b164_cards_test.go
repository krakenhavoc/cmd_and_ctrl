package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB164ClawsOracle      = "0a064c10-5bd1-41ec-b2b5-d0eb33c73c9b"
	oldB164VirtueOracle     = "4262893c-8cbe-4854-9762-b3b6b75c0e35"
	oldB164AmbusherOracle   = "72949d87-76cd-472c-86b3-2b7d5fce3159"
	oldB164RimestagOracle   = "731898d3-465b-4ecb-a346-b91881714ebc"
	oldB164FrostwallaOracle = "90772e8f-9b39-4cdd-864f-5e2c917f645e"
	oldB164ObsessionOracle  = "1c55af05-7fbe-4078-9463-689b21c0d3e4"
	oldB164SoulStrikeOracle = "2a6b32cd-a0f5-4e3e-9069-b525d485f84f"
	oldB164ReprobationOracl = "445dfef5-358e-49b1-880c-b30fc45f0a9a"
	oldB164ThinIceOracle    = "6736bb96-6718-48bb-9385-789d3349dc20"
	oldB164TreefolkOracle   = "fa321366-9bc1-48c7-873a-0b12a026ce49"
)

// --- Treefolk Umbra -----------------------------------------------------------

func TestOldB164TreefolkUmbraPumpsToughnessAssignsByToughnessAndHasUmbraArmor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	umbra := effectstest.AuraCast(t, g, "Treefolk Umbra", oldB164TreefolkOracle, bear)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 2 || tough != 4 {
		t.Fatalf("enchanted Bear is %d/%d, want 2/4", p, tough)
	}
	if !effectstest.EffectiveOf(t, g, bear).DamageByToughnessAlways {
		t.Error("the enchanted creature does not assign combat damage by toughness")
	}
	effectstest.AssertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}

func TestOldB164TreefolkUmbraAssignsToughnessDamageInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.UmbraCreature(g, me.ID, 1, 1)
	effectstest.AuraCast(t, g, "Treefolk Umbra", oldB164TreefolkOracle, bear)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the 1/3 enchanted Bear dealt %d, want 3 (its toughness)", got)
	}
}

// --- Soul-Strike Technique ----------------------------------------------------

func TestOldB164SoulStrikeTechniquePumpsGrantsVigilanceAndManifestsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Card", TypeLine: "Creature — Elf", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.AuraCast(t, g, "Soul-Strike Technique", oldB164SoulStrikeOracle, bear)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("enchanted Bear is %d/%d, want 3/3", p, tough)
	}
	effectstest.AssertEffectiveKeyword(t, g, bear, "vigilance")
	effectstest.B18Kill(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Fatalf("the top card is in %q, not manifested onto the battlefield", effectstest.ZoneOf(g, top))
	}
	c, _ := effectstest.BattlefieldCard(g, top)
	if !c.FaceDown {
		t.Error("the manifested card is not face down")
	}
	if p, tough := effectstest.EffectivePower(t, g, top), effectstest.EffectiveToughness(t, g, top); p != 2 || tough != 2 {
		t.Errorf("the manifested card is %d/%d, want a face-down 2/2", p, tough)
	}
}

// --- Volatile Claws -----------------------------------------------------------

func TestOldB164VolatileClawsPumpsYourCreaturesAndMakesThemEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.UmbraCreature(g, me.ID, 2, 2)
	b := effectstest.UmbraCreature(g, me.ID, 3, 3)
	theirs := effectstest.UmbraCreature(g, opp.ID, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Volatile Claws", "Instant", oldB164ClawsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 4, b: 5, theirs: 2} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
	if !effectstest.EffectiveOf(t, g, a).AllCreatureTypes {
		t.Error("your creature is not every creature type")
	}
	if effectstest.EffectiveOf(t, g, theirs).AllCreatureTypes {
		t.Error("their creature became every creature type")
	}
	late := effectstest.UmbraCreature(g, me.ID, 1, 1)
	if got := effectstest.EffectivePower(t, g, late); got != 1 {
		t.Errorf("a creature that entered afterwards has power %d, want 1 (the set is fixed on resolution)", got)
	}
}

// --- Force of Virtue ----------------------------------------------------------

func TestOldB164ForceOfVirtuePitchesAWhiteCardOnlyOffYourTurnAndPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	mine := effectstest.UmbraCreature(g, me.ID, 2, 2)
	theirs := effectstest.UmbraCreature(g, opp.ID, 2, 2)
	white := effectstest.HandCardFull(opp, "Their White Card", "Instant", "{W}", "", []string{"W"})
	blue := effectstest.HandCardFull(opp, "Their Blue Card", "Instant", "{U}", "", []string{"U"})
	virtue := effectstest.HandCardFull(opp, "Force of Virtue", "Enchantment", "{2}{W}{W}", oldB164VirtueOracle, []string{"W"})
	if err := g.CastSpell(opp.ID, virtue, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{blue}}); err == nil {
		t.Fatal("a blue card paid for the white pitch")
	}
	if err := g.CastSpell(opp.ID, virtue, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{white}}); err != nil {
		t.Fatalf("Force of Virtue off the pitch on another player's turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(virtue) || !g.Exile.Contains(white) {
		t.Fatal("Force of Virtue did not resolve or the white card was not exiled")
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 3 {
		t.Errorf("their creature has power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("my creature has power %d, want 2 (it is not their anthem)", got)
	}
}

func TestOldB164ForceOfVirtueCannotPitchOnYourOwnTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	white := effectstest.HandCardFull(me, "My White Card", "Instant", "{W}", "", []string{"W"})
	virtue := effectstest.HandCardFull(me, "Force of Virtue", "Enchantment", "{2}{W}{W}", oldB164VirtueOracle, []string{"W"})
	if err := g.CastSpell(me.ID, virtue, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{white}}); err == nil {
		t.Fatal("pitched Force of Virtue on your own turn")
	}
}

// --- Treetop Ambusher ---------------------------------------------------------

func TestOldB164TreetopAmbusherPumpsATargetCreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ambusher := effectstest.PushCatalogPermanent(g, me.ID, "Treetop Ambusher", "Creature — Elf Berserker", oldB164AmbusherOracle, false)
	effectstest.Unsick(g, ambusher)
	friend := effectstest.UmbraCreature(g, me.ID, 2, 2)
	effectstest.AttackWithTrigger(t, g, ambusher, opp.ID, friend)
	if p := effectstest.EffectivePower(t, g, friend); p != 3 {
		t.Errorf("the target has power %d, want 3", p)
	}
}

func TestOldB164TreetopAmbusherDashesWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastWithAltCost(t, g, "Treetop Ambusher", "Creature — Elf Berserker", oldB164AmbusherOracle, "dash")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatalf("the dashed Ambusher is in %q", effectstest.ZoneOf(g, id))
	}
	c := effectstest.EffectiveCardForTest(t, g, id)
	if !effectstest.EffectiveHasKeyword(&c, "haste") {
		t.Error("the dashed Ambusher has no haste")
	}
}

// --- Saddled Rimestag ---------------------------------------------------------

func TestOldB164SaddledRimestagGrowsWhenAnotherCreatureEntersThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	stag := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: stag, Name: "Saddled Rimestag", TypeLine: "Snow Creature — Elk", ManaCost: "{1}{G}", OracleID: oldB164RimestagOracle,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, stag, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, stag); p != 2 {
		t.Fatalf("the Rimestag is %d-power alone, want 2 (its own entry does not count)", p)
	}
	effectstest.CastCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, stag), effectstest.EffectiveToughness(t, g, stag); p != 4 || tough != 4 {
		t.Errorf("after another creature entered under your control it is %d/%d, want 4/4", p, tough)
	}
}

// --- Frostwalla ---------------------------------------------------------------

func TestOldB164FrostwallaPumpsOnceEachTurnForSnowMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	walla := effectstest.BlbB1Push(g, me.ID, "Frostwalla", "Snow Creature — Lizard", oldB164FrostwallaOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.BlbB1Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("{S} was paid with a nonsnow land")
	}
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
	effectstest.BlbB1Push(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, walla); p != 4 || tough != 4 {
		t.Errorf("the Frostwalla is %d/%d, want 4/4", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Error("the second activation in one turn was allowed")
	}
}

// --- Sadistic Obsession -------------------------------------------------------

func TestOldB164SadisticObsessionGivesTheCreatureAMinusCounterAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.UmbraCreature(g, me.ID, 2, 2)
	victim := effectstest.UmbraCreature(g, opp.ID, 3, 3)
	effectstest.Unsick(g, bear)
	effectstest.AuraCast(t, g, "Sadistic Obsession", oldB164ObsessionOracle, bear)
	effectstest.AdvanceToMain(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the enchanted creature has no granted activated ability")
	}
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("%d -1/-1 counters on the target, want 1", got)
	}
	if !effectstest.B13Tapped(t, g, bear) {
		t.Error("the enchanted creature was not tapped by its own {T} cost")
	}
}

// --- Reprobation --------------------------------------------------------------

func TestOldB164ReprobationMakesAZeroOneCowardWithNoAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Iron Golem", TypeLine: "Legendary Artifact Creature — Golem",
		Colors: []string{"R"}, Power: 4, Toughness: 4, Keywords: []string{"trample"}, Owner: me.ID, Controller: me.ID})
	effectstest.AuraCast(t, g, "Reprobation", oldB164ReprobationOracl, golem)
	ch := effectstest.EffectiveOf(t, g, golem)
	if len(ch.Types) != 1 || ch.Types[0] != "Creature" || len(ch.Subtypes) != 1 || ch.Subtypes[0] != "Coward" {
		t.Errorf("types %v subtypes %v, want Creature — Coward", ch.Types, ch.Subtypes)
	}
	if len(ch.Supertypes) != 1 || ch.Supertypes[0] != "Legendary" {
		t.Error("the creature lost its supertype")
	}
	if p, tough := effectstest.EffectivePower(t, g, golem), effectstest.EffectiveToughness(t, g, golem); p != 0 || tough != 1 {
		t.Errorf("enchanted creature is %d/%d, want 0/1", p, tough)
	}
	c := effectstest.EffectiveCardForTest(t, g, golem)
	if effectstest.EffectiveHasKeyword(&c, "trample") {
		t.Error("the creature kept trample")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(golem, game.CounterPlusOne, 2); err != nil {
			t.Fatalf("add counters: %v", err)
		}
	})
	if p := effectstest.EffectivePowerWithCounters(t, g, golem); p != 2 {
		t.Errorf("with two +1/+1 counters it has power %d, want 2 (counters apply on top of the set 0/1)", p)
	}
}

// --- On Thin Ice --------------------------------------------------------------

func TestOldB164OnThinIceExilesUntilItLeavesAndNeedsASnowLandYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	theirSnow := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Snow Land", TypeLine: "Basic Snow Land — Island", Owner: opp.ID, Controller: opp.ID})
	snow := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Snow-Covered Plains", TypeLine: "Basic Snow Land — Plains", Owner: me.ID, Controller: me.ID})
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	for _, bad := range []uuid.UUID{forest, theirSnow} {
		if err := effectstest.CastCatalogSpellErr(t, g, "On Thin Ice", "Snow Enchantment — Aura", oldB164ThinIceOracle, effectstest.CardRefs(bad)); err == nil {
			t.Fatal("a land that is not a snow land you control was accepted")
		}
	}
	aura := effectstest.CastCatalogSpell(t, g, "On Thin Ice", "Snow Enchantment — Aura", oldB164ThinIceOracle, effectstest.CardRefs(snow))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil && effectstest.HasID(pick.PickTargetCards, mine) {
		t.Error("a creature you control is offered as a target")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the opposing creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if got := effectstest.EoeCDCount(g, opp.ID, "Their Threat"); got != 1 {
		t.Errorf("after the Aura left, %d copies of the creature are back under its owner, want 1", got)
	}
}
