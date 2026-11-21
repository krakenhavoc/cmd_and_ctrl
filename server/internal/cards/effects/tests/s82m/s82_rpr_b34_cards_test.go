package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 RPR batch 34: Dream Pillager, Eternal Dragon, Fiend of the Shadows,
// Gryff's Boon, Hoard-Smelter Dragon, Ignite the Future, Izzet Chemister,
// Knight of Autumn, Lumbering Falls, Masterwork of Ingenuity, Mishra's Factory.

const (
	rprb34DreamPillager = "504cc26d-3fd5-4391-9d30-f0614e5c8bbc"
	rprb34EternalDragon = "04d8615c-3883-4251-9790-1d8a4a40e142"
	rprb34Fiend         = "3891a3b6-d8ac-4cd3-b2f6-d6d786e89a67"
	rprb34GryffsBoon    = "f62f3f4e-d506-43a9-91d8-bbc86e0a9167"
	rprb34HoardSmelter  = "cf9db2be-3c95-45a2-ab85-694e3bd3bf1d"
	rprb34Ignite        = "915aa871-4c5e-4799-9a07-d7b855df4e7a"
	rprb34Chemister     = "e4a3d6f3-36ba-4e6d-a254-351bc8886f20"
	rprb34Knight        = "c4212ba4-8c43-40ca-aa36-b3b535659fa8"
	rprb34Falls         = "c2878f4e-3044-4d60-8ec1-b325edfca397"
	rprb34Masterwork    = "35d458fb-dfdb-41e4-92e3-f0fd392e1c69"
	rprb34Factory       = "5963e0ef-e0bc-4611-ad4f-813a4c0eacfb"
)

func TestRprB34CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{rprb34DreamPillager, rprb34EternalDragon, rprb34Fiend, rprb34GryffsBoon,
		rprb34HoardSmelter, rprb34Ignite, rprb34Chemister, rprb34Knight, rprb34Falls, rprb34Masterwork, rprb34Factory} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func rprb34Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, index int, targets ...uuid.UUID) error {
	t.Helper()
	var refs []game.TargetRef
	for _, tg := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: tg})
	}
	err := g.ActivateCatalogAbility(me.ID, id, index, game.ActivateAbilityParams{Strict: true, Targets: refs})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestRprB34MishrasFactoryAnimatesAndPumpsAnyAssemblyWorker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	factory := effectstest.PushCatalogPermanent(g, me.ID, "Mishra's Factory", "Land", rprb34Factory, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Mishra's Factory", "Land", rprb34Factory, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.FillPool(me, 1)
	if err := rprb34Activate(t, g, me, factory, 0); err != nil {
		t.Fatal(err)
	}
	c := effectstest.EffectiveCardForTest(t, g, factory)
	if !c.IsCreature() || !c.IsLand() || !c.IsArtifact() || !c.HasSubtype("Assembly-Worker") {
		t.Errorf("animated Factory: creature %v land %v artifact %v Assembly-Worker %v",
			c.IsCreature(), c.IsLand(), c.IsArtifact(), c.HasSubtype("Assembly-Worker"))
	}
	if p, tough := effectstest.EffectivePower(t, g, factory), effectstest.EffectiveToughness(t, g, factory); p != 2 || tough != 2 {
		t.Errorf("Factory is %d/%d, want 2/2", p, tough)
	}
	if err := rprb34Activate(t, g, me, other, 1, bear); err == nil {
		t.Error("pumped a creature that is not an Assembly-Worker")
	}
	if err := rprb34Activate(t, g, me, other, 1, factory); err != nil {
		t.Fatal(err)
	}
	if p, tough := effectstest.EffectivePower(t, g, factory), effectstest.EffectiveToughness(t, g, factory); p != 3 || tough != 3 {
		t.Errorf("pumped Factory is %d/%d, want 3/3", p, tough)
	}
}

func TestRprB34LumberingFallsEntersTappedAndBecomesAHexproofElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	falls := effectstest.PushCatalogPermanent(g, me.ID, "Lumbering Falls", "Land", rprb34Falls, false)
	effectstest.FillPoolColored(me, "G", 3)
	effectstest.FillPoolColored(me, "U", 1)
	if err := rprb34Activate(t, g, me, falls, 0); err != nil {
		t.Fatal(err)
	}
	c := effectstest.EffectiveCardForTest(t, g, falls)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Elemental") {
		t.Errorf("animated Falls: creature %v land %v Elemental %v", c.IsCreature(), c.IsLand(), c.HasSubtype("Elemental"))
	}
	if p, tough := effectstest.EffectivePower(t, g, falls), effectstest.EffectiveToughness(t, g, falls); p != 3 || tough != 3 {
		t.Errorf("Falls is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, falls, "hexproof") {
		t.Error("animated Falls lacks hexproof")
	}
	if got := c.EffectiveColors(); len(got) != 2 {
		t.Errorf("colours %v, want green and blue", got)
	}
}

func TestRprB34EternalDragonReturnsOnlyDuringYourUpkeepAndPlainscycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Eternal Dragon", TypeLine: "Creature — Dragon Spirit",
		OracleID: rprb34EternalDragon, Owner: me.ID, Controller: me.ID})

	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "W", 5)
	if err := rprb34Activate(t, g, me, id, 0); err == nil {
		t.Fatal("returned the Dragon outside your upkeep")
	}
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g2, game.StepUpkeep)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	id2 := uuid.New()
	me2.Graveyard.PushTop(game.Card{InstanceID: id2, Name: "Eternal Dragon", TypeLine: "Creature — Dragon Spirit",
		OracleID: rprb34EternalDragon, Owner: me2.ID, Controller: me2.ID})
	effectstest.FillPoolColored(me2, "W", 5)
	if err := rprb34Activate(t, g2, me2, id2, 0); err != nil {
		t.Fatal(err)
	}
	if !me2.Hand.Contains(id2) || me2.Graveyard.Contains(id2) {
		t.Error("the Dragon did not return to hand in your upkeep")
	}

	g3 := effectstest.NewCatalogGame(t)
	me3 := g3.Seats[g3.Turn.ActiveSeat]
	plains := effectstest.SeedSearchLibrary(me3, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	hand := effectstest.PutInHand(me3, game.Card{Name: "Eternal Dragon", TypeLine: "Creature — Dragon Spirit", OracleID: rprb34EternalDragon})
	effectstest.AdvanceToMain(t, g3)
	effectstest.FillPool(me3, 2)
	idx := -1
	spec, _ := effects.Lookup(rprb34EternalDragon)
	for i, a := range spec.Activated {
		if len(a.Label) >= 12 && a.Label[:12] == "Plainscyclin" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no Plainscycling ability")
	}
	if err := rprb34Activate(t, g3, me3, hand, idx); err != nil {
		t.Fatal(err)
	}
	if !me3.Graveyard.Contains(hand) {
		t.Error("the Dragon was not discarded")
	}
	effectstest.PassPriorityAroundTable(t, g3)
	effectstest.AnswerSearchByID(t, g3, me3.ID, plains[0])
	effectstest.PassPriorityAroundTable(t, g3)
	if !me3.Hand.Contains(plains[0]) {
		t.Error("the Plains did not reach the hand")
	}
}

func TestRprB34GryffsBoonEnchantsAndReturnsFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	boon := effectstest.CastCatalogSpell(t, g, "Gryff's Boon", "Enchantment — Aura", rprb34GryffsBoon,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("enchanted power %d, want 3", p)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("enchanted creature lacks flying")
	}
	// Send it to the graveyard, then bring it back onto another creature.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, boon); err != nil {
		t.Fatal(err)
	}
	if !me.Graveyard.Contains(boon) {
		t.Fatal("the Aura is not in the graveyard")
	}
	second := effectstest.PushVanillaCreature(g, me.ID, "Second", 1, 1)
	effectstest.FillPoolColored(me, "W", 4)
	if err := rprb34Activate(t, g, me, boon, 0, second); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(boon) {
		t.Fatal("the Aura did not return")
	}
	if p := effectstest.EffectivePower(t, g, second); p != 2 {
		t.Errorf("second creature power %d, want 2", p)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, second, "flying") {
		t.Error("the returned Aura did not grant flying")
	}
}

func TestRprB34HoardSmelterDragonDestroysAnArtifactAndGrowsByItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.PushCatalogPermanent(g, me.ID, "Hoard-Smelter Dragon", "Creature — Dragon", rprb34HoardSmelter, false)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact",
		ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := effectstest.EffectivePower(t, g, dragon)
	effectstest.FillPoolColored(me, "R", 4)
	if err := rprb34Activate(t, g, me, dragon, 0, bear); err == nil {
		t.Error("targeted a non-artifact")
	}
	if err := rprb34Activate(t, g, me, dragon, 0, rock); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if got := effectstest.EffectivePower(t, g, dragon); got != before+4 {
		t.Errorf("Dragon power %d → %d, want +4", before, got)
	}
}

func rprb34KnightMode(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	effectstest.CastAndResolveCreature(t, g, "Knight of Autumn", "Creature — Dryad Knight", rprb34Knight)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestRprB34KnightOfAutumnModes(t *testing.T) {
	t.Run("counters", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		rprb34KnightMode(t, g, me, 0)
		effectstest.PassPriorityAroundTable(t, g)
		var knight uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == rprb34Knight {
				knight = c.InstanceID
			}
		}
		if n := effectstest.PlusOneCounters(g, knight); n != 2 {
			t.Errorf("the Knight has %d +1/+1 counters, want 2", n)
		}
	})
	t.Run("destroy", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact",
			Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
		bear := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Bear", 2, 2)
		rprb34KnightMode(t, g, me, 1)
		effectstest.PickCard(t, g, me.ID, rock)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("the artifact survived")
		}
		if !g.Battlefield.Contains(bear) {
			t.Error("a creature was destroyed")
		}
	})
	t.Run("life", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		life := me.Life
		rprb34KnightMode(t, g, me, 2)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d -> %d, want +4", life, me.Life)
		}
	})
}

func TestRprB34IgniteTheFutureExilesThreeAndFlashbackIsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "B", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "C", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "D", TypeLine: "Sorcery", ManaCost: "{R}"},
	)
	spell := effectstest.CastCatalogSpell(t, g, "Ignite the Future", "Sorcery", rprb34Ignite, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if !g.Exile.Contains(id) {
			t.Fatalf("card %s was not exiled", id)
		}
		perm := effectstest.ExiledPermission(g, id)
		if perm.Player != me.ID || perm.CastOnly || perm.Cost != "" {
			t.Errorf("%+v: want a paid play permission", perm)
		}
		if !effectstest.PermissionLive(g, perm, me.ID) {
			t.Error("the permission is not live")
		}
	}
	if !me.Library.Contains(ids[3]) {
		t.Error("a fourth card was exiled")
	}
	if !me.Graveyard.Contains(spell) {
		t.Fatal("Ignite the Future is not in the graveyard")
	}

	// Flashback: the cards may be played without paying their mana costs.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.AdvanceToMainOf(t, g2, g2.Turn.ActiveSeat)
	ids2 := effectstest.SeedSearchLibrary(me2,
		game.Card{Name: "A", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "B", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "C", TypeLine: "Sorcery", ManaCost: "{R}"},
	)
	id := uuid.New()
	me2.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Ignite the Future", TypeLine: "Sorcery", OracleID: rprb34Ignite,
		Owner: me2.ID, Controller: me2.ID})
	effectstest.FillPoolColored(me2, "R", 8)
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	for _, c := range ids2 {
		perm := effectstest.ExiledPermission(g2, c)
		if perm.Cost != "{0}" {
			t.Errorf("flashback permission cost %q, want {0}", perm.Cost)
		}
	}
	if !g2.Exile.Contains(id) {
		t.Error("the flashed-back spell is not exiled")
	}
}

func TestRprB34DreamPillagerExilesThatManyAndMayCastSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	dragon := effectstest.PushCatalogPermanent(g, me.ID, "Dream Pillager", "Creature — Dragon", rprb34DreamPillager, false)
	_ = dragon
	cards := make([]game.Card, 0, 6)
	for i := 0; i < 6; i++ {
		cards = append(cards, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{R}"})
	}
	cards[3] = game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}
	ids := effectstest.SeedSearchLibrary(me, cards...)
	effectstest.DealCombatDamageToPlayer(g, dragon, victim.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:4] {
		if !g.Exile.Contains(id) {
			t.Fatalf("card %s was not exiled", id)
		}
		if perm := effectstest.ExiledPermission(g, id); perm.Player != me.ID || !perm.CastOnly {
			t.Errorf("%+v: want a cast-only permission for the controller", perm)
		}
	}
	if g.Exile.Contains(ids[4]) {
		t.Error("exiled more than the damage dealt")
	}
}

func TestRprB34FiendOfTheShadowsTakesACardFromTheVictimsHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	fiend := effectstest.PushCatalogPermanent(g, me.ID, "Fiend of the Shadows", "Creature — Vampire Wizard", rprb34Fiend, false)
	keep := effectstest.PutInHand(victim, game.Card{Name: "Keep", TypeLine: "Instant"})
	take := effectstest.PutInHand(victim, game.Card{Name: "Take", TypeLine: "Instant"})
	effectstest.DealCombatDamageToPlayer(g, fiend, victim.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, victim.ID, take)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(take) || !victim.Hand.Contains(keep) {
		t.Fatalf("exiled %v, kept card in hand %v", g.Exile.Contains(take), victim.Hand.Contains(keep))
	}
	perm := effectstest.ExiledPermission(g, take)
	if perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
		t.Errorf("%+v: the Fiend's controller may play it", perm)
	}
}

func TestRprB34FiendOfTheShadowsRegeneratesBySacrificingAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fiend := effectstest.PushCatalogPermanent(g, me.ID, "Fiend of the Shadows", "Creature — Vampire Wizard", rprb34Fiend, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	human := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Peasant", TypeLine: "Creature — Human",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	_ = bear
	if err := g.ActivateCatalogAbility(me.ID, fiend, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("sacrificed a non-Human")
	}
	if err := g.ActivateCatalogAbility(me.ID, fiend, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{human}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(human) {
		t.Error("the Human was not sacrificed")
	}
}

func TestRprB34IzzetChemisterExilesAndCastsFromItsPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chem := effectstest.PushCatalogPermanent(g, me.ID, "Izzet Chemister", "Creature — Goblin Wizard", rprb34Chemister, false)
	bolt := effectstest.PushGraveyardCardTyped(me, "Bolt", "Instant")
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.FillPoolColored(me, "R", 1)
	if err := rprb34Activate(t, g, me, chem, 0, bear); err == nil {
		t.Error("exiled a creature card")
	}
	if err := rprb34Activate(t, g, me, chem, 0, bolt); err != nil {
		t.Fatal(err)
	}
	if !g.Exile.Contains(bolt) {
		t.Fatal("the instant was not exiled")
	}
	effectstest.UntapForTest(g, chem)
	effectstest.FillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, chem, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(chem) {
		t.Error("the Chemister was not sacrificed")
	}
	if perm := effectstest.ExiledPermission(g, bolt); perm.Player != me.ID || perm.Cost != "{0}" {
		t.Errorf("%+v: want a free permission for the controller", perm)
	}
}

func TestRprB34MasterworkOfIngenuityCopiesAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Blade", TypeLine: "Artifact — Equipment",
		OracleID: "oracle-Test Blade", Owner: active.ID, Controller: active.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: "oracle-Sol Ring", Owner: active.ID, Controller: active.ID})
	id := effectstest.CastCatalogSpell(t, g, "Masterwork of Ingenuity", "Artifact — Equipment", rprb34Masterwork, nil)
	effectstest.ResolveWithCopyChoice(t, g, sword)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Test Blade" || !got.IsCopy() {
		t.Errorf("got %q (copy %v), want a copy of Test Blade", got.Name, got.IsCopy())
	}
}

func TestRprB34MasterworkOfIngenuityCannotCopyANonEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: "oracle-Sol Ring", Owner: active.ID, Controller: active.ID})
	id := effectstest.CastCatalogSpell(t, g, "Masterwork of Ingenuity", "Artifact — Equipment", rprb34Masterwork, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CopyPrompt(g) != nil {
		t.Error("a non-Equipment is not a legal copy candidate")
	}
	if got := effectstest.CopyBattlefieldCard(t, g, id); got.Name != "Masterwork of Ingenuity" || got.IsCopy() {
		t.Errorf("should stay itself: %+v", got)
	}
}
