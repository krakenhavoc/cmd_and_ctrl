package testss82m

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b516_cards_test.go — Shards of Alara, batch 516 (S82 older sets,
// #5588). Every test plays the real catalog entry through the engine.

const (
	oldB516ReflectorOracle  = "0124977f-06e6-430a-a874-f7d05883a97c"
	oldB516SkeletonizeOracl = "2a4cde8f-38e0-439b-9928-95ab23317f70"
	oldB516ArchdemonOracle  = "41e250e6-8e38-4cbc-ae66-98800a5feba8"
	oldB516AssaultOracle    = "d1255776-b88b-4f05-9b71-ff89a2eba453"
	oldB516PowderOracle     = "22d57bc8-e473-47c5-8e48-c0fb98e66948"
	oldB516ShadowfeedOracle = "7681ea90-be9d-4a43-845e-ba051b438429"
	oldB516TitanOracle      = "823cd586-92db-4e44-ba1f-394b2404cafc"
	oldB516ThrashOracle     = "89599a64-19fc-47ad-a116-3ce0a10fa546"
	oldB516OgreOracle       = "343a258e-ea46-4acb-8543-9f869e43bc59"
	oldB516DraggerOracle    = "77630700-3828-4063-95a7-f6283f0f9da5"
	oldB516LeotauOracle     = "dd7df06c-9434-46ce-893c-80081fbc2df7"
	oldB516ScreecherOracle  = "fb51bafe-5a70-4910-a42e-ef75d9dd2f0a"
)

func TestOldB516EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB516ReflectorOracle:  "Minion Reflector",
		oldB516SkeletonizeOracl: "Skeletonize",
		oldB516ArchdemonOracle:  "Archdemon of Unx",
		oldB516AssaultOracle:    "Goblin Assault",
		oldB516PowderOracle:     "Protomatter Powder",
		oldB516ShadowfeedOracle: "Shadowfeed",
		oldB516TitanOracle:      "Salvage Titan",
		oldB516ThrashOracle:     "Carrion Thrash",
		oldB516OgreOracle:       "Fire-Field Ogre",
		oldB516DraggerOracle:    "Viscera Dragger",
		oldB516LeotauOracle:     "Undead Leotau",
		oldB516ScreecherOracle:  "Kathari Screecher",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB516UnearthCreaturesReturnWithHasteAndAreExiledAtTheEndStep(t *testing.T) {
	for _, tc := range []effectstest.BroB04Card{
		{Name: "Fire-Field Ogre", TypeLine: "Creature — Ogre Mutant", Oracle: oldB516OgreOracle, Power: 4, Toughness: 2, Unearth: "{U}{B}{R}"},
		{Name: "Kathari Screecher", TypeLine: "Creature — Bird Soldier", Oracle: oldB516ScreecherOracle, Power: 2, Toughness: 2, Unearth: "{2}{U}"},
		{Name: "Viscera Dragger", TypeLine: "Creature — Zombie Ogre Warrior", Oracle: oldB516DraggerOracle, Power: 3, Toughness: 3, Unearth: "{1}{B}"},
		{Name: "Undead Leotau", TypeLine: "Creature — Zombie Cat", Oracle: oldB516LeotauOracle, Power: 3, Toughness: 4, Unearth: "{2}{B}"},
	} {
		g := effectstest.NewCatalogGame(t)
		id, _ := effectstest.BroB04Unearth(t, g, tc)
		if z := effectstest.B12ZoneOf(g, id); z != game.ZoneBattlefield {
			t.Fatalf("%s is in %v after unearth", tc.Name, z)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", tc.Name)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if z := effectstest.B12ZoneOf(g, id); z != game.ZoneExile {
			t.Errorf("%s is in %v after the end step, want exile", tc.Name, z)
		}
	}
}

func TestOldB516KeywordsAreDeclared(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, tl, oracle, kw string
	}{
		{"Fire-Field Ogre", "Creature — Ogre Mutant", oldB516OgreOracle, "first strike"},
		{"Kathari Screecher", "Creature — Bird Soldier", oldB516ScreecherOracle, "flying"},
		{"Archdemon of Unx", "Creature — Demon", oldB516ArchdemonOracle, "flying"},
		{"Archdemon of Unx", "Creature — Demon", oldB516ArchdemonOracle, "trample"},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.tl, tc.oracle, 2, 2)
		if !effectstest.HasEffectiveKeyword(t, g, id, tc.kw) {
			t.Errorf("%s lacks %s", tc.name, tc.kw)
		}
	}
}

func TestOldB516VisceraDraggerCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	id, _ := effectstest.CycleFromHand(t, g, "Viscera Dragger", "Creature — Zombie Ogre Warrior", oldB516DraggerOracle, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestOldB516UndeadLeotauGetsPlusOneMinusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Undead Leotau", "Creature — Zombie Cat", oldB516LeotauOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.PowerOf(t, g, id); got != 5 {
		t.Errorf("power = %d, want 5", got)
	}
}

func TestOldB516ShadowfeedExilesAnyGraveyardCardAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	card := effectstest.PushGraveyardCardForTest(opp, "Their Card")
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Shadowfeed", "Instant", oldB516ShadowfeedOracle, effectstest.CardTargets(card))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(card) || effectstest.B12ZoneOf(g, card) != game.ZoneExile {
		t.Error("the card was not exiled from the opponent's graveyard")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d -> %d, want +3", life, got)
	}
}

func TestOldB516ProtomatterPowderReturnsAnArtifactAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	powder := effectstest.PushCatalogPermanent(g, me.ID, "Protomatter Powder", "Artifact", oldB516PowderOracle, false)
	rock := effectstest.PushGraveyardPermanent(me, "Mind Stone", "Artifact", "{2}")
	bear := effectstest.PushGraveyardPermanent(me, "Grizzly", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{W}")
	// A non-artifact card is not a legal target.
	if err := g.ActivateCatalogAbility(me.ID, powder, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err == nil {
		t.Fatal("targeted a creature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, powder, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, powder); ok {
		t.Error("the Powder was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, rock) != game.ZoneBattlefield {
		t.Errorf("the artifact is in %v, want the battlefield", effectstest.B12ZoneOf(g, rock))
	}
}

func TestOldB516SalvageTitanReturnsToHandByExilingThreeArtifactCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	titan := effectstest.PushCatalogGraveyardCard(me, "Salvage Titan", "Artifact Creature — Golem", oldB516TitanOracle, 6, 4)
	a := effectstest.PushGraveyardPermanent(me, "Rock A", "Artifact", "{1}")
	b := effectstest.PushGraveyardPermanent(me, "Rock B", "Artifact", "{1}")
	c := effectstest.PushGraveyardPermanent(me, "Rock C", "Artifact", "{1}")
	nonArtifact := effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: oldB516TitanOracle})
	idx := -1
	for i, ab := range abilities {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no graveyard ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, titan, idx, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b, nonArtifact}}); err == nil {
		t.Fatal("paid with a non-artifact card")
	}
	if err := g.ActivateCatalogAbility(me.ID, titan, idx, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a, b, c}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(titan) {
		t.Error("the Titan did not return to hand")
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("artifact card %s is in %v, want exile", id, effectstest.B12ZoneOf(g, id))
		}
	}
}

func TestOldB516SalvageTitanOffersTheThreeArtifactSacrifice(t *testing.T) {
	spec, ok := effects.Lookup(oldB516TitanOracle)
	if !ok || len(spec.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %d, want 1", len(spec.AlternativeCosts))
	}
	alt := spec.AlternativeCosts[0]
	if alt.ManaCost != "" || alt.Sacrifice == nil || !strings.Contains(alt.Label, "three artifacts") {
		t.Errorf("alternative cost = %+v, want a free cost sacrificing three artifacts", alt)
	}
}

func TestOldB516CarrionThrashReturnsAnotherCreatureCardIfYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		thrash := effectstest.B12Push(g, me.ID, "Carrion Thrash", "Creature — Lizard Warrior", oldB516ThrashOracle, 4, 4)
		bear := effectstest.PushGraveyardPermanent(me, "Grizzly", "Creature — Bear", "{1}{G}")
		effectstest.FloatMana(t, g, me, "{C}{C}")
		effectstest.KillCreature(t, g, me.ID, thrash)
		effectstest.PassPriorityAroundTable(t, g)
		if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
			// The Thrash is in the graveyard too: it must not be offered.
			for _, id := range pick.PickTargetCards {
				if id == thrash {
					t.Error("the Thrash offered itself as 'another' creature card")
				}
			}
			effectstest.PickCard(t, g, me.ID, bear)
			effectstest.PassPriorityAroundTable(t, g)
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Contains(bear); got != pay {
			t.Errorf("pay=%v: the creature card is in hand = %v", pay, got)
		}
		if me.Hand.Contains(thrash) {
			t.Error("the Thrash returned itself")
		}
	}
}

func TestOldB516GoblinAssaultMakesAHastyGoblinAndForcesEveryGoblinToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Goblin Assault", "Enchantment", oldB516AssaultOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := len(oldB516CreaturesOf(g, me.ID))
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	after := oldB516CreaturesOf(g, me.ID)
	if len(after) != before+1 {
		t.Fatalf("creatures %d -> %d, want one new Goblin token", before, len(after))
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("pass with the Goblins home = %v, want an attack-requirement refusal", err)
	}
	for _, c := range after {
		if c.HasSubtype("Goblin") {
			if err := g.DeclareAttacker(c.InstanceID, opp.ID); err != nil {
				t.Fatalf("declare %s: %v", c.Name, err)
			}
		}
	}
	// The Bear is not a Goblin: it may stay home once every Goblin attacks.
	if err := g.PassPriority(); err != nil {
		t.Errorf("pass with every Goblin attacking and the Bear home = %v, want it accepted", err)
	}
	_, _ = theirs, bear
	_ = mine
}

func oldB516CreaturesOf(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.IsCreature() {
				out = append(out, c)
			}
		}
	})
	return out
}

func TestOldB516ArchdemonSacrificesANonZombieThenMakesAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Archdemon of Unx", "Creature — Demon", oldB516ArchdemonOracle, 6, 6)
	zombie := effectstest.B12Creature(g, me.ID, "Plain Zombie", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BeginUpkeep(t, g, me.ID)
	prompt := effectstest.SacrificeChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	for _, id := range effectstest.SacrificeOptions(prompt) {
		if id == zombie {
			t.Error("a Zombie was offered for the sacrifice")
		}
	}
	effectstest.AnswerSacrifice(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the Bear was not sacrificed")
	}
	zombies := 0
	for _, c := range oldB516CreaturesOf(g, me.ID) {
		if c.IsToken() && c.HasSubtype("Zombie") && c.Power == 2 && c.Toughness == 2 {
			zombies++
		}
	}
	if zombies != 1 {
		t.Errorf("2/2 Zombie tokens = %d, want 1", zombies)
	}
}

func TestOldB516SkeletonizeDamagesAndLeavesARegeneratingSkeletonWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runt := effectstest.PushVanillaCreature(g, opp.ID, "Runt", 2, 2)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 3, 5)
	effectstest.CastCatalogSpell(t, g, "Skeletonize", "Instant", oldB516SkeletonizeOracl, effectstest.CardTargets(runt))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, runt); ok {
		t.Fatal("3 damage did not kill the 2/2")
	}
	var skeleton *game.Card
	for _, c := range oldB516CreaturesOf(g, me.ID) {
		if c.Name == "Skeleton" {
			cc := c
			skeleton = &cc
		}
	}
	if skeleton == nil || skeleton.Power != 1 || skeleton.Toughness != 1 || !skeleton.IsToken() {
		t.Fatalf("Skeleton = %+v, want a 1/1 token under the caster", skeleton)
	}
	if abs := game.ActivatedAbilitiesForCard(*skeleton); len(abs) != 1 {
		t.Errorf("the Skeleton has %d activated abilities, want its one regenerate ability", len(abs))
	}
	// A creature that survives the damage makes nothing, even if it dies later.
	n := len(oldB516CreaturesOf(g, me.ID))
	effectstest.CastCatalogSpell(t, g, "Skeletonize", "Instant", oldB516SkeletonizeOracl, effectstest.CardTargets(big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, opp.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(oldB516CreaturesOf(g, me.ID)); got != n+1 {
		// It did die this turn after being dealt damage, so a second Skeleton arrives.
		t.Errorf("creatures %d -> %d, want one more Skeleton (the damaged creature died this turn)", n, got)
	}
}

func TestOldB516MinionReflectorCopiesANontokenCreatureWithHasteIfYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Minion Reflector", "Artifact", oldB516ReflectorOracle, false)
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatMana(t, g, me, "{C}{C}")
		bear := oldB516Cast(t, g, "Grizzly Twin", "", 2, 2)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		copies := 0
		var tok game.Card
		for _, c := range oldB516CreaturesOf(g, me.ID) {
			if c.IsToken() && c.Name == "Grizzly Twin" {
				copies++
				tok = c
			}
		}
		want := 0
		if pay {
			want = 1
		}
		if copies != want {
			t.Fatalf("pay=%v: token copies = %d, want %d", pay, copies, want)
		}
		if !pay {
			continue
		}
		if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
			t.Error("the copy has no haste")
		}
		// Tokens entering do not trigger it again, and the copy is gone at the end step.
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if _, ok := effectstest.BattlefieldCard(g, tok.InstanceID); ok {
			t.Error("the copy survived the end step")
		}
		if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
			t.Error("the original was sacrificed instead of the copy")
		}
	}
}

// oldB516Cast puts a creature card with printed stats in the active player's
// hand and casts it (the test pool pays no costs).
func oldB516Cast(t *testing.T, g *game.Game, name, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}
