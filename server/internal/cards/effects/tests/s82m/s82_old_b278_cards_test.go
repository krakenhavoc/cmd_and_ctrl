package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b278_cards_test.go — Eldritch Moon, batch 278 (#5541). Every test
// plays the real catalog entry through the engine.

const (
	b278CaptiveOracle    = "0272ca81-e727-4f4b-b06e-072d70bb5558"
	b278HomunculusOracle = "c38db10f-c0b7-4391-b49a-77079a4d087c"
	b278BoonOracle       = "20d30d4d-cf38-41fe-a2ce-8327f7c87c01"
	b278RavingsOracle    = "69af7bff-fb8c-4f59-88a1-6bfa87ba400a"
	b278ShredsOracle     = "6eaa343d-0097-4f05-80ab-deb016d4e15e"
	b278SanctifierOracle = "806fc324-a17a-4e65-a9b6-d739c471c7f8"
	b278AxeOracle        = "e024870c-af13-4ef3-be27-cdd84dab7d7d"
	b278CleaverOracle    = "e51cbdce-017e-4b14-b3a7-2db50ddd8037"
	b278LunarOracle      = "31b18e10-c74d-4384-a0e2-def0b3ec4799"
	b278VildinOracle     = "6e6f9269-c72b-4449-9c2d-a720b9bd6f06"
	b278ScarecrowOracle  = "b925c4ed-8767-46b8-a423-a5f094557d94"
	b278MoonOracle       = "bbb4f08f-3016-40c2-ae7c-afc720eeb27a"
	b278ScourgeOracle    = "c8b0436e-16d2-4d48-a894-454d69a3743b"
)

var (
	b278Captive = effectstest.IwCard{Oracle: b278CaptiveOracle, Front: "Ulvenwald Captive", FrontType: "Creature — Werewolf Horror", Cost: "{1}{G}",
		FrontText: "Defender\n{T}: Add {G}.\n{5}{G}{G}: Transform this creature.", Fp: "1", Ft: "2",
		Back: "Ulvenwald Abomination", BackText: "{T}: Add {C}{C}.", Bp: "4", Bt: "6", Colors: []string{"G"}, Keywords: []string{"Defender"}}
	b278Homunculus = effectstest.IwCard{Oracle: b278HomunculusOracle, Front: "Curious Homunculus", FrontType: "Creature — Homunculus", Cost: "{1}{U}",
		FrontText: "{T}: Add {C}. Spend this mana only to cast an instant or sorcery spell.\nAt the beginning of your upkeep, if there are three or more instant and/or sorcery cards in your graveyard, transform this creature.",
		Fp:        "1", Ft: "1", Back: "Voracious Reader", BackText: "Prowess\nInstant and sorcery spells you cast cost {1} less to cast.", Bp: "3", Bt: "4", Colors: []string{"U"}, Keywords: []string{"Prowess"}}
	b278Vildin = effectstest.IwCard{Oracle: b278VildinOracle, Front: "Vildin-Pack Outcast", FrontType: "Creature — Werewolf Horror", Cost: "{4}{R}",
		FrontText: "Trample\n{R}: This creature gets +1/-1 until end of turn.\n{5}{R}{R}: Transform this creature.", Fp: "4", Ft: "4",
		Back: "Dronepack Kindred", BackText: "Trample\n{1}: This creature gets +1/+0 until end of turn.", Bp: "5", Bt: "7", Colors: []string{"R"}, Keywords: []string{"Trample"}}
)

func TestB278EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		b278CaptiveOracle: "Ulvenwald Captive", b278CaptiveOracle + "#1": "Ulvenwald Abomination",
		b278HomunculusOracle: "Curious Homunculus", b278HomunculusOracle + "#1": "Voracious Reader",
		b278BoonOracle: "Boon of Emrakul", b278RavingsOracle: "Prophetic Ravings", b278ShredsOracle: "Shreds of Sanity",
		b278SanctifierOracle: "Sanctifier of Souls", b278AxeOracle: "Thirsting Axe", b278CleaverOracle: "Slayer's Cleaver",
		b278LunarOracle: "Lunar Force", b278VildinOracle: "Vildin-Pack Outcast", b278VildinOracle + "#1": "Dronepack Kindred",
		b278ScarecrowOracle: "Geist-Fueled Scarecrow", b278MoonOracle: "Waxing Moon", b278ScourgeOracle: "Eternal Scourge",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func b278Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func b278Grave(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Ulvenwald Captive ---

func TestB278CaptiveTapsForGreenThenTransformsToTapForTwoColorless(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, b278Captive, me)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("G"); got != 1 {
		t.Fatalf("the Captive made %d G, want 1", got)
	}
	effectstest.AssertKeywords(t, g, id, "defender")
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, effectstest.BroB04Pool("{5}{G}{G}"), nil); err != nil {
		t.Fatalf("transform: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("the Captive is on face %d, want the Abomination", effectstest.IwFace(g, id))
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool(); effectstest.FindBattlefieldCardForTest(g, id).Tapped = false })
	effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("C"); got != 2 {
		t.Errorf("the Abomination made %d C, want 2", got)
	}
}

func TestB278CaptiveTransformNeedsSevenMana(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, b278Captive, me)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, effectstest.BroB04Pool("{4}{G}{G}"), nil); err == nil {
		t.Fatal("six mana paid for {5}{G}{G}")
	}
}

// --- Curious Homunculus ---

func TestB278HomunculusManaPaysOnlyForInstantsAndSorceries(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, b278Homunculus, me)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
	if _, err := effectstest.CastFromHandErr(g, me, "Bear Cub", "Creature — Bear", "{1}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the Homunculus's mana paid for a creature spell")
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Cheap Sorcery", "Sorcery", "{1}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Homunculus's mana could not pay for a sorcery: %v", err)
	}
}

func TestB278HomunculusTransformsAtUpkeepWithThreeInstantsOrSorceries(t *testing.T) {
	for _, n := range []int{2, 3} {
		g, me := effectstest.IwTable(t)
		id := effectstest.IwPut(t, g, b278Homunculus, me)
		for i := 0; i < n; i++ {
			typ := "Instant"
			if i%2 == 1 {
				typ = "Sorcery"
			}
			b278Grave(me, "Spell", typ)
		}
		b278Grave(me, "Bear", "Creature — Bear")
		// Walk to the controller's next upkeep.
		for i := 0; i < 80 && !(g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == 0 && i > 5); i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		effectstest.MonSettle(t, g)
		want := 0
		if n == 3 {
			want = 1
		}
		if got := effectstest.IwFace(g, id); got != want {
			t.Errorf("%d instants/sorceries: face %d, want %d", n, got, want)
		}
	}
}

func TestB278VoraciousReaderDiscountsInstantsAndSorceriesOnly(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, b278Homunculus, me)
	effectstest.IwTransform(t, g, id)
	effectstest.AssertKeywords(t, g, id, "prowess")
	effectstest.B06AddMana(me, "R")
	if _, err := effectstest.CastFromHandErr(g, me, "Shock", "Instant", "{1}{R}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a {1}{R} instant should cost {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Bear Cub", "Creature — Bear", "{2}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("one mana paid for a {2} creature spell")
	}
	effectstest.B06AddMana(me, "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Bear Cub", "Creature — Bear", "{2}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("two mana pay a {2} creature: %v", err)
	}
}

// --- Boon of Emrakul ---

func TestB278BoonOfEmrakulGivesPlusThreeMinusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 2, 4)
	effectstest.B06AddMana(me, "B", "C", "C")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Boon of Emrakul", "Enchantment — Aura", b278BoonOracle, b278Target(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, big); p != 5 || tough != 1 {
		t.Errorf("enchanted creature is %d/%d, want 5/1", p, tough)
	}
}

func TestB278BoonOfEmrakulKillsAThreeToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	effectstest.B06AddMana(me, "B", "C", "C")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Boon of Emrakul", "Enchantment — Aura", b278BoonOracle, b278Target(small))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("a 2/3 enchanted by +3/-3 survived")
	}
}

// --- Prophetic Ravings ---

func TestB278PropheticRavingsGivesHasteAndALootAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Prophetic Ravings", "Enchantment — Aura", b278RavingsOracle, b278Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, bear, "haste")
	discard := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: discard, Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate the granted loot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d -> %d, want a discard and a draw", hand, len(me.Hand.Cards))
	}
	if me.Hand.Contains(discard) {
		t.Error("the discarded card is still in hand")
	}
}

// --- Shreds of Sanity ---

func TestB278ShredsReturnsBothAndDiscardsThenExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	inst := b278Grave(me, "Bolt", "Instant")
	sor := b278Grave(me, "Divination", "Sorcery")
	effectstest.B06AddMana(me, "R", "C", "C")
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastCatalogSpell(t, g, "Shreds of Sanity", "Sorcery", b278ShredsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: inst}, {Kind: game.TargetCard, ID: sor}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(inst) || !me.Hand.Contains(sor) {
		t.Fatal("the instant and the sorcery did not both return")
	}
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Shreds of Sanity is in %v, want exile", z)
	}
}

func TestB278ShredsWorksWithNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B06AddMana(me, "R", "C", "C")
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastCatalogSpell(t, g, "Shreds of Sanity", "Sorcery", b278ShredsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Shreds of Sanity is in %v, want exile", z)
	}
}

// --- Sanctifier of Souls ---

func TestB278SanctifierPumpsOnAnotherCreatureButNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	san := effectstest.B12Push(g, me.ID, "Sanctifier of Souls", "Creature — Human Cleric", b278SanctifierOracle, 2, 3)
	effectstest.B06AddMana(me, "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Bear Cub", "Creature — Bear", "{1}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, san); p != 3 || tough != 4 {
		t.Errorf("Sanctifier is %d/%d after another creature entered, want 3/4", p, tough)
	}
	// An opponent's creature entering does not trigger it.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, opp.ID, "Rival", "Creature — Bear", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, san); p != 3 {
		t.Errorf("Sanctifier power %d after an opposing creature entered, want 3", p)
	}
}

func TestB278SanctifierExilesAGraveyardCreatureForASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	san := effectstest.B12Push(g, me.ID, "Sanctifier of Souls", "Creature — Human Cleric", b278SanctifierOracle, 2, 3)
	spell := b278Grave(me, "Divination", "Sorcery")
	effectstest.B06AddMana(me, "W", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, san, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery card paid for an exile-a-creature-card cost")
	}
	dead := b278Grave(me, "Dead Bear", "Creature — Bear")
	if err := g.ActivateCatalogAbility(me.ID, san, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{dead}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(dead); z != nil && z.Kind != game.ZoneExile {
		t.Errorf("the creature card is in %v, want exile", z.Kind)
	}
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("the sorcery left the graveyard")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("%d Spirit tokens, want 1", got)
	}
}

// --- Thirsting Axe ---

func b278Equip(t *testing.T, g *game.Game, me *game.Player, equipment, host uuid.UUID, cost string) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, equipment, 0, game.ActivateAbilityParams{Strict: true, Targets: b278Target(host)}); err == nil {
		t.Fatal("equipped with no mana")
	}
	effectstest.FillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, equipment, 0, game.ActivateAbilityParams{Strict: true, Targets: b278Target(host)}); err != nil {
		t.Fatalf("equip %s: %v", cost, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB278AxeGivesPlusFourAndSacrificesAnIdleCreatureAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	axe := effectstest.SeedEquipment(g, me.ID, "Thirsting Axe", b278AxeOracle)
	b278Equip(t, g, me, axe, bear, "{2}")
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 2 {
		t.Fatalf("equipped bear is %d/%d, want 6/2", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the idle equipped creature was not sacrificed")
	}
}

func TestB278AxeLeavesACreatureThatFoughtACreatureAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 5)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	axe := effectstest.SeedEquipment(g, me.ID, "Thirsting Axe", b278AxeOracle)
	b278Equip(t, g, me, axe, bear, "{2}")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature that dealt combat damage to a creature was sacrificed")
	}
}

func TestB278AxeSacrificesACreatureThatOnlyHitAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	axe := effectstest.SeedEquipment(g, me.ID, "Thirsting Axe", b278AxeOracle)
	b278Equip(t, g, me, axe, bear, "{2}")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a creature that only damaged a player was not sacrificed")
	}
}

func TestB278AxeUnattachedDoesNothingAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	effectstest.SeedEquipment(g, me.ID, "Thirsting Axe", b278AxeOracle)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("an unequipped creature was sacrificed")
	}
}

// --- Slayer's Cleaver ---

func TestB278CleaverGivesPlusThreePlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	cleaver := effectstest.SeedEquipment(g, me.ID, "Slayer's Cleaver", b278CleaverOracle)
	b278Equip(t, g, me, cleaver, bear, "{4}")
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("equipped bear is %d/%d, want 5/3", p, tough)
	}
}

func TestB278CleaverMustBeBlockedByAnEldraziIfAble(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.SeedBear(g, me.ID)
	cleaver := effectstest.SeedEquipment(g, me.ID, "Slayer's Cleaver", b278CleaverOracle)
	b278Equip(t, g, me, cleaver, host, "{4}")
	eldrazi := effectstest.B12Creature(g, opp.ID, "Scion", "Creature — Eldrazi Scion", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, host); err != nil {
		t.Fatalf("a non-Eldrazi blocking: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); !errors.Is(err, game.ErrIllegalBlock) {
		t.Fatalf("finishing without the Eldrazi: err = %v, want a block requirement refusal", err)
	}
	if err := g.DeclareBlocker(eldrazi, host); err != nil {
		t.Fatalf("the Eldrazi joins: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks with the Eldrazi blocking: %v", err)
	}
}

func TestB278CleaverAsksNothingWithoutAnEldrazi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	host := effectstest.SeedBear(g, me.ID)
	cleaver := effectstest.SeedEquipment(g, me.ID, "Slayer's Cleaver", b278CleaverOracle)
	b278Equip(t, g, me, cleaver, host, "{4}")
	effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("no Eldrazi, no block: %v", err)
	}
}

// --- Lunar Force ---

func TestB278LunarForceSacrificesItselfAndCountersAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	force := effectstest.PushCatalogPermanent(g, opp.ID, "Lunar Force", "Enchantment", b278LunarOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	id, err := effectstest.CastFromHandErr(g, me, "Bear Cub", "Creature — Bear", "{1}", "", game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(force) {
		t.Error("Lunar Force was not sacrificed")
	}
	if g.Battlefield.Contains(id) {
		t.Error("the opponent's creature spell resolved")
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("the countered spell is in %v, want the graveyard", z)
	}
}

func TestB278LunarForceIgnoresItsControllersOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	force := effectstest.PushCatalogPermanent(g, me.ID, "Lunar Force", "Enchantment", b278LunarOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	id, err := effectstest.CastFromHandErr(g, me, "Bear Cub", "Creature — Bear", "{1}", "", game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(force) || !g.Battlefield.Contains(id) {
		t.Error("Lunar Force reacted to its controller's own spell")
	}
}

// --- Vildin-Pack Outcast ---

func TestB278VildinPackOutcastFiresUpAndTransformsIntoADroneKindred(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, b278Vildin, me)
	effectstest.AssertKeywords(t, g, id, "trample")
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{R}", nil); err != nil {
		t.Fatalf("firebreathing: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 5 || tough != 3 {
		t.Errorf("after {R}: %d/%d, want 5/3", p, tough)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 1, effectstest.BroB04Pool("{5}{R}{R}"), nil); err != nil {
		t.Fatalf("transform: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatalf("the Outcast is on face %d, want Dronepack Kindred", effectstest.IwFace(g, id))
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, effectstest.BroB04Pool("{1}"), nil); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 7 || tough != 6 {
		t.Errorf("Kindred after {1}: %d/%d, want 7/6 (5/7, +1/-1 from before, +1/+0 now)", p, tough)
	}
	effectstest.AssertKeywords(t, g, id, "trample")
}

// --- Geist-Fueled Scarecrow ---

func TestB278ScarecrowTaxesOnlyItsControllersCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Geist-Fueled Scarecrow", "Artifact Creature — Scarecrow", b278ScarecrowOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Bear Cub", "Creature — Bear", "{1}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("one mana cast a {1} creature spell through the Scarecrow's tax")
	}
	effectstest.B06AddMana(me, "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Bear Cub", "Creature — Bear", "{1}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("two mana should pay {1}+{1}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Cheap Sorcery", "Sorcery", "{1}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a sorcery is not taxed: %v", err)
	}
	_ = opp
}

// --- Waxing Moon ---

func TestB278WaxingMoonTransformsAWerewolfAndGrantsTrample(t *testing.T) {
	g, me := effectstest.IwTable(t)
	wolf := effectstest.IwPut(t, g, b278Captive, me)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "G", "C")
	effectstest.CastCatalogSpell(t, g, "Waxing Moon", "Instant", b278MoonOracle, b278Target(wolf))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, wolf) != 1 {
		t.Errorf("the Werewolf is on face %d, want 1", effectstest.IwFace(g, wolf))
	}
	effectstest.AssertKeywords(t, g, bear, "trample")
	effectstest.AssertKeywords(t, g, wolf, "trample")
}

func TestB278WaxingMoonWithNoTargetStillGrantsTrampleAndRefusesANonWerewolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "C", "G", "C")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Waxing Moon", "Instant", b278MoonOracle, b278Target(bear)); err == nil {
		t.Fatal("Waxing Moon targeted a Bear")
	}
	effectstest.CastCatalogSpell(t, g, "Waxing Moon", "Instant", b278MoonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, bear, "trample")
}

// --- Eternal Scourge ---

func b278SlayingFire(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	effectstest.B06AddMana(me, "R", "C", "C")
	effectstest.CastCatalogSpell(t, g, "Slaying Fire", "Instant", effectstest.SlayingFireOracle, b278Target(target))
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB278ScourgeExilesItselfWhenAnOpponentTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scourge := effectstest.B12Push(g, opp.ID, "Eternal Scourge", "Creature — Eldrazi Horror", b278ScourgeOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	b278SlayingFire(t, g, me, scourge)
	if z := g.FindCardZoneForEffect(scourge); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the Scourge is in %v, want exile (the trigger beats the damage)", z)
	}
}

func TestB278ScourgeIgnoresItsOwnControllersTargeting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scourge := effectstest.B12Push(g, me.ID, "Eternal Scourge", "Creature — Eldrazi Horror", b278ScourgeOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	b278SlayingFire(t, g, me, scourge)
	if z := g.FindCardZoneForEffect(scourge); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("the Scourge is in %v, want the graveyard (no trigger for its controller's own spell)", z)
	}
}

func TestB278ScourgeCanBeCastFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	scourge := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: scourge, Name: "Eternal Scourge", TypeLine: "Creature — Eldrazi Horror", ManaCost: "{3}",
			OracleID: b278ScourgeOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FillPool(me, 2)
	if err := g.CastSpell(me.ID, scourge, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("a {3} spell was cast with two mana")
	}
	effectstest.FillPool(me, 1)
	if err := g.CastSpell(me.ID, scourge, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the Scourge from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scourge) {
		t.Error("the Scourge cast from exile did not resolve onto the battlefield")
	}
}
