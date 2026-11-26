package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 432 (Planechase 2012): Preyseizer Dragon, Thorn-Thrash Viashino,
// Thromok the Insatiable, Nullmage Advocate, Krond the Dawn-Clad, Bramble
// Elemental, Dreampod Druid, Sunken Hope, Sai of the Shinobi, Predatory Urge.

const (
	oldB432Dragon    = "1b36eb85-0b1e-41fb-8ecd-dda13c4feffc"
	oldB432Viashino  = "83d9a5a2-09b0-44bf-b0b8-82a5fc53b00a"
	oldB432Thromok   = "fbb6ec41-3017-4d8d-bca0-c93040f1b291"
	oldB432Nullmage  = "558568a6-5e22-482f-9096-e76682965c21"
	oldB432Krond     = "62ca69ed-ebd3-4e84-b17a-4bcc9488f8b3"
	oldB432Bramble   = "ac5c0078-e2f1-4ac1-8748-575d69e09cf8"
	oldB432Dreampod  = "db2af2b0-a5aa-4e4a-b5bc-52a5ac3968fc"
	oldB432Hope      = "1820b6e9-83fd-40b9-953e-ba1a8ae69614"
	oldB432Sai       = "ec57dcb8-127f-4922-a9f9-5227db1751a9"
	oldB432Urge      = "f6175cca-0e19-4db8-a175-3369fb650f24"
	oldB432GiftAura  = "5c77bf76-bb0f-4e44-babe-0e7b7b555e1f" // Quicksilver Dagger, a plain Aura for the board
	oldB432AuraTypes = "Enchantment — Aura"
)

func TestOldB432EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB432Dragon: "Preyseizer Dragon", oldB432Viashino: "Thorn-Thrash Viashino", oldB432Thromok: "Thromok the Insatiable",
		oldB432Nullmage: "Nullmage Advocate", oldB432Krond: "Krond the Dawn-Clad", oldB432Bramble: "Bramble Elemental",
		oldB432Dreampod: "Dreampod Druid", oldB432Hope: "Sunken Hope", oldB432Sai: "Sai of the Shinobi", oldB432Urge: "Predatory Urge",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

// oldB432Enchant puts a plain Aura (Quicksilver Dagger) on host, which is
// "enchanted" for these cards' purposes.
func oldB432Enchant(t *testing.T, g *game.Game, host uuid.UUID) uuid.UUID {
	t.Helper()
	return effectstest.AuraCast(t, g, "Quicksilver Dagger", oldB432GiftAura, host)
}

// --- Devour -----------------------------------------------------------------

func TestOldB432DragonDevoursForTwoCountersEachAndHitsForItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.PushDevourFodder(g, me.ID, "Goblin")
	id := effectstest.CastDevourer(t, g, "Preyseizer Dragon", oldB432Dragon)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DevourPlusOnes(g, id); got != 2 {
		t.Fatalf("+1/+1 counters %d, want 2 (devour 2, one creature)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("no flying")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	before := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, id)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if lost := before - opp.Life; lost < 2 {
		t.Errorf("the opponent lost %d, want at least the 2 damage of the trigger", lost)
	}
}

func TestOldB432ViashinoDevoursAndGainsTrampleForG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushDevourFodder(g, me.ID, "Goblin A")
	b := effectstest.PushDevourFodder(g, me.ID, "Goblin B")
	id := effectstest.CastDevourer(t, g, "Thorn-Thrash Viashino", oldB432Viashino)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DevourPlusOnes(g, id); got != 4 {
		t.Fatalf("+1/+1 counters %d, want 4", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Fatal("trample before the ability")
	}
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("no trample after {G}")
	}
}

func TestOldB432ThromokDevoursXForXCountersEach(t *testing.T) {
	for _, k := range []int{1, 3} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		var fodder []uuid.UUID
		for i := 0; i < k; i++ {
			fodder = append(fodder, effectstest.PushDevourFodder(g, me.ID, "Goblin"))
		}
		id := effectstest.CastDevourer(t, g, "Thromok the Insatiable", oldB432Thromok)
		effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, fodder...)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.DevourPlusOnes(g, id); got != k*k {
			t.Errorf("devoured %d: counters %d, want %d", k, got, k*k)
		}
	}
}

func TestOldB432ThromokThatDevoursNothingGetsNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kept := effectstest.PushDevourFodder(g, me.ID, "Goblin")
	id := effectstest.CastDevourer(t, g, "Thromok the Insatiable", oldB432Thromok)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DevourPlusOnes(g, id); got != 0 {
		t.Errorf("+1/+1 counters %d, want 0", got)
	}
	if !g.Battlefield.Contains(kept) {
		t.Error("the creature that was not devoured left the battlefield")
	}
}

// --- Nullmage Advocate ------------------------------------------------------

func TestOldB432NullmageReturnsTwoCardsAndDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adv := effectstest.PushCatalogPermanent(g, me.ID, "Nullmage Advocate", "Creature — Insect Druid", oldB432Nullmage, false)
	a := effectstest.PushGraveyardCardWithTypeLineID(opp, "Their A", "Creature — Elf")
	b := effectstest.PushGraveyardCardWithTypeLineID(opp, "Their B", "Instant")
	c := effectstest.PushGraveyardCardWithTypeLineID(opp, "Their C", "Sorcery")
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	other := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Land")
	targets := append(effectstest.CardRefs(a, b), game.TargetRef{Kind: game.TargetCard, ID: relic})
	if err := g.ActivateCatalogAbility(me.ID, adv, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("the two cards did not return to their owner's hand")
	}
	if !opp.Graveyard.Contains(c) {
		t.Error("an untargeted card left the graveyard")
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("an untargeted permanent was destroyed")
	}
}

func TestOldB432NullmageNeedsTwoCardsFromOneGraveyardAndAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adv := effectstest.PushCatalogPermanent(g, me.ID, "Nullmage Advocate", "Creature — Insect Druid", oldB432Nullmage, false)
	a := effectstest.PushGraveyardCardWithTypeLineID(opp, "Their A", "Creature — Elf")
	b := effectstest.PushGraveyardCardWithTypeLineID(g.Seats[2], "Third A", "Instant")
	mine := effectstest.PushGraveyardCardWithTypeLineID(me, "My Card", "Instant")
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Land")
	try := func(refs []game.TargetRef) error {
		return g.ActivateCatalogAbility(me.ID, adv, 0, game.ActivateAbilityParams{Strict: true, Targets: refs})
	}
	tgt := game.TargetRef{Kind: game.TargetCard, ID: relic}
	if err := try(append(effectstest.CardRefs(a, b), tgt)); err == nil {
		t.Error("cards from two different graveyards were accepted")
	}
	if err := try(append(effectstest.CardRefs(a, mine), tgt)); err == nil {
		t.Error("a card from your own graveyard was accepted")
	}
	if err := try(append(effectstest.CardRefs(a), tgt)); err == nil {
		t.Error("a single card was accepted for two")
	}
	opp2 := effectstest.PushGraveyardCardWithTypeLineID(opp, "Their B", "Instant")
	if err := try(append(effectstest.CardRefs(a, opp2), game.TargetRef{Kind: game.TargetCard, ID: land})); err == nil {
		t.Error("a land was accepted as the artifact or enchantment")
	}
}

// --- Krond the Dawn-Clad ----------------------------------------------------

func TestOldB432KrondExilesAPermanentWhenItAttacksEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krond := effectstest.PushCatalogPermanent(g, me.ID, "Krond the Dawn-Clad", "Legendary Creature — Archon", oldB432Krond, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	oldB432Enchant(t, g, krond)
	effectstest.DeclareAttack(t, g, opp.ID, krond)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, victim)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, victim); got != game.ZoneExile {
		t.Errorf("the target is in %q, want exile", got)
	}
}

func TestOldB432KrondDoesNothingUnenchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	krond := effectstest.PushCatalogPermanent(g, me.ID, "Krond the Dawn-Clad", "Legendary Creature — Archon", oldB432Krond, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, krond)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("an unenchanted Krond asked for a target: %+v", p)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the creature was exiled by an unenchanted Krond")
	}
}

// --- Bramble Elemental ------------------------------------------------------

func TestOldB432BrambleMakesTwoSaprolingsWhenAnAuraIsAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bramble := effectstest.PushCatalogPermanent(g, me.ID, "Bramble Elemental", "Creature — Elemental", oldB432Bramble, false)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	oldB432Enchant(t, g, other)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 0 {
		t.Fatalf("%d Saprolings from an Aura on another creature, want 0", n)
	}
	oldB432Enchant(t, g, bramble)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 2 {
		t.Errorf("%d Saprolings, want 2", n)
	}
}

// --- Dreampod Druid ---------------------------------------------------------

func TestOldB432DreampodMakesASaprolingAtEachUpkeepOnlyWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	druid := effectstest.PushCatalogPermanent(g, me.ID, "Dreampod Druid", "Creature — Human Druid", oldB432Dreampod, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 0 {
		t.Fatalf("%d Saprolings while unenchanted, want 0", n)
	}
	oldB432Enchant(t, g, druid)
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 1 {
		t.Errorf("%d Saprolings after an opponent's upkeep, want 1", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	// Seat 3's upkeep came between, so three upkeeps have passed while enchanted.
	if n := effectstest.CountOnBattlefield(g, "Saproling", me.ID); n != 3 {
		t.Errorf("%d Saprolings after three upkeeps while enchanted, want 3", n)
	}
}

// --- Sunken Hope ------------------------------------------------------------

func TestOldB432SunkenHopeReturnsTheUpkeepPlayersChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sunken Hope", "Enchantment", oldB432Hope, false)
	a := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
	if c == nil {
		t.Fatalf("the upkeep player was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(c.ID, opp.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(b) {
		t.Error("the chosen creature did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(mine) {
		t.Error("a creature that was not chosen left the battlefield")
	}
}

// --- Sai of the Shinobi -----------------------------------------------------

func TestOldB432SaiPumpsAndMayAttachToAnEnteringCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sai := effectstest.PushCatalogPermanent(g, me.ID, "Sai of the Shinobi", "Artifact — Equipment", oldB432Sai, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, sai, bear)
	if p, th := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || th != 3 {
		t.Errorf("equipped Bear is %d/%d, want 3/3", p, th)
	}
	effectstest.B12PlayFromHand(t, g, "New Elf", "Creature — Elf", "", game.CastSpellParams{})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	var host uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == sai {
				host = c.AttachedTo.ID
			}
		}
	})
	if host == bear || host == uuid.Nil {
		t.Errorf("the Equipment is attached to %v, want the entering creature", host)
	}
}

// --- Predatory Urge ---------------------------------------------------------

func TestOldB432UrgeLetsTheEnchantedCreatureAndATargetHitEachOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Brute", "Creature — Ogre", 3, 5)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 3)
	effectstest.AuraCast(t, g, "Predatory Urge", oldB432Urge, host)
	if err := g.ActivateCatalogAbility(me.ID, host, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("the 2/3 survived 3 damage")
	}
	if !g.Battlefield.Contains(host) {
		t.Fatal("the 3/5 host died to 2 damage")
	}
	if got := effectstest.DamageOn(t, g, host); got != 2 {
		t.Errorf("the host has %d damage, want 2 (the foe's power)", got)
	}
}
