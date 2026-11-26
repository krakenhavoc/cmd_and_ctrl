package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b357_cards_test.go — S82 OLD batch 357 (#5556, Magic 2015, part 1):
// Avarice Amulet, Haunted Plate Mail, Shield of the Avatar, Hot Soup, Warden
// of the Beyond, Zof Shade, Selfless Cathar, Wall of Fire, Statute of Denial,
// Jalira, Master Polymorphist and Boonweaver Giant. Each test plays the real
// catalog entry through the engine.

const (
	oldB357Amulet    = "4c71074b-b8bf-47f4-87e4-67c275abdbc5"
	oldB357PlateMail = "84594f86-724b-4f6b-b374-2265a00db566"
	oldB357Shield    = "a4433da6-78b7-4669-a0dd-e7ea9672b0bc"
	oldB357Soup      = "a853e48b-1e15-4492-a5b4-9924b1a4cf9a"
	oldB357Warden    = "6e9d5e5e-d571-4a1a-be8c-6672fd6b582f"
	oldB357Zof       = "8e4b8fe8-882c-4d0f-8fec-0256c62a0df0"
	oldB357Cathar    = "a6e4c8de-4c35-4ea4-b865-596ae8cef075"
	oldB357Wall      = "f38c8b47-e8e0-4d2f-b1da-d8d986805a48"
	oldB357Statute   = "03e04cee-4727-44fe-96f3-42cbd08beb76"
	oldB357Jalira    = "62ad15f1-ee51-4991-8e52-06a3cbd8f395"
	oldB357Giant     = "79eed0c6-1853-4032-9d12-aebf5e15e78c"
)

func TestOldB357EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB357Amulet: "Avarice Amulet", oldB357PlateMail: "Haunted Plate Mail",
		oldB357Shield: "Shield of the Avatar", oldB357Soup: "Hot Soup",
		oldB357Warden: "Warden of the Beyond", oldB357Zof: "Zof Shade",
		oldB357Cathar: "Selfless Cathar", oldB357Wall: "Wall of Fire",
		oldB357Statute: "Statute of Denial", oldB357Jalira: "Jalira, Master Polymorphist",
		oldB357Giant: "Boonweaver Giant",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB357Attach(g *game.Game, equipment, host uuid.UUID) {
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: host})
	})
}

// --- Avarice Amulet ------------------------------------------------------------

func TestOldB357AmuletPumpsGrantsVigilanceAndDrawsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	amulet := effectstest.PushCatalogPermanent(g, me.ID, "Avarice Amulet", "Artifact — Equipment", oldB357Amulet, false)
	oldB357Attach(g, amulet, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power = %d, want 4", got)
	}
	if !effectstest.HasKeywordOf(g, bear, "vigilance") {
		t.Error("the equipped creature has no vigilance")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand = %d, want %d (the granted upkeep draw)", got, before+1)
	}
}

func TestOldB357AmuletGoesToAnOpponentWhenTheEquippedCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	amulet := effectstest.PushCatalogPermanent(g, me.ID, "Avarice Amulet", "Artifact — Equipment", oldB357Amulet, false)
	oldB357Attach(g, amulet, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.ControllerOf(t, g, amulet); got != opp.ID {
		t.Errorf("the Amulet's controller is %s, want the opponent %s", got, opp.ID)
	}
}

// --- Haunted Plate Mail --------------------------------------------------------

func TestOldB357PlateMailBecomesACreatureOnlyWhileYouControlNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mail := effectstest.PushCatalogPermanent(g, me.ID, "Haunted Plate Mail", "Artifact — Equipment", oldB357PlateMail, false)
	oldB357Attach(g, mail, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 6 {
		t.Errorf("equipped power = %d, want 6", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, mail, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated while controlling a creature")
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	if err := g.ActivateCatalogAbility(me.ID, mail, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with no creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mail); got != 4 {
		t.Errorf("animated power = %d, want 4", got)
	}
	if got := effectstest.EffectiveToughness(t, g, mail); got != 4 {
		t.Errorf("animated toughness = %d, want 4", got)
	}
	c, _ := effectstest.BattlefieldCard(g, mail)
	if !c.IsCreature() || !c.HasSubtype("Spirit") || c.HasSubtype("Equipment") || !c.HasCardType("artifact") {
		t.Errorf("the animated Mail is %q, want a Spirit artifact creature that is no longer an Equipment", c.TypeLine)
	}
}

// --- Shield of the Avatar ------------------------------------------------------

func TestOldB357ShieldPreventsOneDamagePerCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 9)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 9)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Bear", 2, 2)
	shield := effectstest.PushCatalogPermanent(g, me.ID, "Shield of the Avatar", "Artifact — Equipment", oldB357Shield, false)
	oldB357Attach(g, shield, holder)
	effectstest.Damage(t, g, src, holder, 5) // X = 2 (Holder and Bare)
	if got := effectstest.DamageOn(t, g, holder); got != 3 {
		t.Errorf("5 damage left %d marked, want 3", got)
	}
	effectstest.B12Creature(g, me.ID, "Third", "Creature — Bear", 1, 1) // X = 3
	effectstest.Damage(t, g, src, holder, 3)
	if got := effectstest.DamageOn(t, g, holder); got != 3 {
		t.Errorf("3 damage with X=3 left %d marked in total, want 3 (fully prevented)", got)
	}
	effectstest.Damage(t, g, src, bare, 4)
	if got := effectstest.DamageOn(t, g, bare); got != 4 {
		t.Errorf("an unequipped creature took %d, want 4", got)
	}
}

// --- Hot Soup ------------------------------------------------------------------

func TestOldB357SoupDestroysTheEquippedCreatureThatIsDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 9)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 9)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Bear", 2, 2)
	soup := effectstest.PushCatalogPermanent(g, me.ID, "Hot Soup", "Artifact — Equipment", oldB357Soup, false)
	oldB357Attach(g, soup, holder)
	effectstest.AssertRestrictions(t, g, holder, game.CantBeBlocked)
	effectstest.Damage(t, g, src, bare, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bare) {
		t.Error("an unequipped creature was destroyed")
	}
	effectstest.Damage(t, g, src, holder, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(holder) {
		t.Error("the equipped creature survived being dealt damage")
	}
}

// --- Warden of the Beyond ------------------------------------------------------

func TestOldB357WardenGrowsWhileAnOpponentOwnsAnExiledCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warden := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Warden of the Beyond", TypeLine: "Creature — Human Wizard", OracleID: oldB357Warden,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.EffectivePower(t, g, warden); got != 2 {
		t.Fatalf("power = %d, want 2 with nothing in exile", got)
	}
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	effectstest.ExileCreature(t, g, mine)
	if got := effectstest.EffectivePower(t, g, warden); got != 2 {
		t.Errorf("power = %d, want 2: the Warden's controller owns that card", got)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.ExileCreature(t, g, theirs)
	if got := effectstest.EffectivePower(t, g, warden); got != 4 {
		t.Errorf("power = %d, want 4", got)
	}
	if got := effectstest.EffectiveToughness(t, g, warden); got != 4 {
		t.Errorf("toughness = %d, want 4", got)
	}
}

// --- Zof Shade, Wall of Fire ---------------------------------------------------

func TestOldB357ZofShadeGetsPlusTwoPlusTwoPerActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shade := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zof Shade", TypeLine: "Creature — Shade", OracleID: oldB357Zof,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	for i := 0; i < 2; i++ {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}"); err != nil {
			t.Fatal(err)
		}
		if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tt := effectstest.EffectivePower(t, g, shade), effectstest.EffectiveToughness(t, g, shade); p != 6 || tt != 6 {
		t.Errorf("Zof Shade is %d/%d, want 6/6 after two activations", p, tt)
	}
}

func TestOldB357WallOfFireFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall of Fire", TypeLine: "Creature — Wall", OracleID: oldB357Wall,
		Power: 0, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 3; i++ {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
			t.Fatal(err)
		}
		if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.EffectivePower(t, g, wall); got != 3 {
		t.Errorf("Wall of Fire's power = %d, want 3", got)
	}
}

// --- Selfless Cathar -----------------------------------------------------------

func TestOldB357CatharSacrificesToPumpTheTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cathar := effectstest.PushCatalogPermanent(g, me.ID, "Selfless Cathar", "Creature — Human Cleric", oldB357Cathar, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, cathar, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cathar) {
		t.Error("the Cathar was not sacrificed")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("my Bear's power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their Bear's power = %d, want 2", got)
	}
}

// --- Statute of Denial ---------------------------------------------------------

func TestOldB357StatuteCountersAndLootsOnlyWithABlueCreature(t *testing.T) {
	for _, withBlue := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if withBlue {
			effectstest.Pr7Creature(g, me.ID, "Blue Guy", 1, "U")
		}
		victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Statute of Denial", "Instant", oldB357Statute, effectstest.CardRefs(victim))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Stack.Contains(victim) {
			t.Errorf("blue=%v: the spell was not countered", withBlue)
		}
		drew := me.Hand.Size() > hand
		if drew != withBlue {
			t.Errorf("blue=%v: loot happened = %v", withBlue, drew)
		}
	}
}

// --- Jalira, Master Polymorphist -----------------------------------------------

func TestOldB357JaliraRevealsToANonlegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jalira := effectstest.PushCatalogPermanent(g, me.ID, "Jalira, Master Polymorphist", "Legendary Creature — Human Wizard", oldB357Jalira, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	hit := uuid.New()
	legend := uuid.New()
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: hit, Name: "Target Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: legend, Name: "Legend", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, jalira, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without naming a creature to sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, jalira, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{jalira}}); err == nil {
		t.Fatal("Jalira sacrificed itself: the cost says another creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, jalira, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !g.Battlefield.Contains(hit) {
		t.Error("the nonlegendary creature did not enter the battlefield")
	}
	if g.Battlefield.Contains(legend) || g.Battlefield.Contains(land) {
		t.Error("a legendary creature or a land entered the battlefield")
	}
	if effectstest.B12ZoneOf(g, legend) != game.ZoneLibrary || effectstest.B12ZoneOf(g, land) != game.ZoneLibrary {
		t.Error("the rest of the revealed cards did not go to the bottom of the library")
	}
}

// --- Boonweaver Giant ----------------------------------------------------------

func TestOldB357GiantFetchesAnAuraFromHandOntoItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aura := effectstest.HandSpell(me, "Test Aura", "Enchantment — Aura", "{1}")
	giant := effectstest.CastCatalogSpell(t, g, "Boonweaver Giant", "Creature — Giant Monk", oldB357Giant, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(giant) {
		t.Fatal("the Giant is not on the battlefield")
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatalf("the Aura is in %v, want the battlefield", effectstest.B12ZoneOf(g, aura))
	}
}
