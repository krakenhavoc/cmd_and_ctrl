package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b500_cards_test.go — Magic 2010, batch 500 (#5585). Every test plays
// the real catalog entry through the engine.

const (
	o500Vines     = "51e7405e-a714-4c76-8913-10d626d2361e"
	o500SoulBleed = "e4f1acd6-b883-470b-88c2-3989011869d7"
	o500Weakness  = "f07a24c0-bf3c-4733-9473-c6be3b16950e"
	o500Polymorph = "039115f2-6322-4674-bc64-f21883ed375a"
	o500Paladin   = "ad62e9d6-61bc-485f-a9b8-2ef9ba98a4b0"
	o500Mirror    = "38ffe614-e8f7-4625-b921-9409e7e7ee5b"
	o500Shade     = "3ce8560d-dc68-4551-81c9-3a7b8d8da9a1"
	o500Armor     = "fc24af7b-f3d0-41d0-aaac-d83424455717"
	o500Bear      = "Creature — Bear"
	o500Aura      = "Enchantment — Aura"
)

func TestS82OldB500EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o500Vines, o500SoulBleed, o500Weakness, o500Polymorph, o500Paladin, o500Mirror, o500Shade, o500Armor} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// o500Cast casts a catalog spell from the active player's hand in a main
// phase with {10} floating, under strict rules.
func o500Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "{2}{G}{G}{B}{U}{R}{W}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func o500Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func o500Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int, tapped bool) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner, Tapped: tapped,
	})
}

// --- Entangling Vines --------------------------------------------------------------------

func TestS82OldB500EntanglingVinesEnchantsOnlyATappedCreatureAndHoldsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	untapped := o500Perm(g, opp.ID, "Untapped Bear", o500Bear, "", 2, 2, false)
	if err := o500Cast(t, g, "Entangling Vines", o500Aura, "{3}{G}", o500Vines, o500Target(untapped)); err == nil {
		t.Fatal("Entangling Vines was cast on an untapped creature")
	}
	g = effectstest.NewCatalogGame(t)
	opp = g.Seats[1]
	host := o500Perm(g, opp.ID, "Tapped Bear", o500Bear, "", 2, 2, true)
	if err := o500Cast(t, g, "Entangling Vines", o500Aura, "{3}{G}", o500Vines, o500Target(host)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	attached := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Entangling Vines" && c.AttachedTo.ID == host {
			attached = true
		}
	}
	if !attached {
		t.Fatal("Entangling Vines did not attach")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Soul Bleed --------------------------------------------------------------------------

func TestS82OldB500SoulBleedDrainsTheCreaturesControllerEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	host := o500Perm(g, opp.ID, "Bear", o500Bear, "", 2, 2, false)
	if err := o500Cast(t, g, "Soul Bleed", o500Aura, "{2}{B}", o500SoulBleed, o500Target(host)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	me := g.Seats[0]
	meLife, oppLife := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, oppLife-1)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != meLife {
		t.Errorf("a life total moved in the aura controller's upkeep: %d, %d", me.Life, opp.Life)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("after the second upkeep the creature's controller is at %d, want %d", opp.Life, oppLife-2)
	}
}

// --- Weakness ----------------------------------------------------------------------------

func TestS82OldB500WeaknessShrinksTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	host := o500Perm(g, opp.ID, "Bear", o500Bear, "", 2, 2, false)
	if err := o500Cast(t, g, "Weakness", o500Aura, "{B}", o500Weakness, o500Target(host)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, host); p != 0 || tough != 1 {
		t.Errorf("the enchanted Bear is %d/%d, want 0/1", p, tough)
	}
}

// --- Looming Shade -----------------------------------------------------------------------

func TestS82OldB500LoomingShadePumpsForBlackMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	shade := o500Perm(g, me.ID, "Looming Shade", "Creature — Shade", o500Shade, 1, 1, false)
	if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the pump was activated with no mana")
	}
	effectstest.FillPoolColored(me, "B", 2)
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.PtOf(t, g, shade); p != 3 || tough != 3 {
		t.Errorf("Looming Shade is %d/%d after two pumps, want 3/3", p, tough)
	}
}

// --- Polymorph ---------------------------------------------------------------------------

func TestS82OldB500PolymorphReplacesTheCreatureEvenThroughRegeneration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := o500Perm(g, opp.ID, "Victim", o500Bear, "", 2, 2, false)
	if err := (effects.Regenerate{Target: victim}).Apply(effects.NewContext(g, &game.StackItem{Controller: opp.ID, SourceCardID: uuid.New()})); err != nil {
		t.Fatalf("shield: %v", err)
	}
	var land, hit, below uuid.UUID
	g.WithWriteLock(func() {
		below, hit, land = uuid.New(), uuid.New(), uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: below, Name: "Below", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
		opp.Library.PushTop(game.Card{InstanceID: hit, Name: "Hit Dragon", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
		opp.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	})
	if err := o500Cast(t, g, "Polymorph", "Sorcery", "{3}{U}", o500Polymorph, o500Target(victim)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, victim) != nil || !opp.Graveyard.Contains(victim) {
		t.Error("the regeneration shield saved the creature, but Polymorph says it can't be regenerated")
	}
	c := effectstest.FindBattlefieldCardForTest(g, hit)
	if c == nil || c.Controller != opp.ID {
		t.Fatal("the first creature card revealed is not on the battlefield under its controller's control")
	}
	if effectstest.FindBattlefieldCardForTest(g, below) != nil || !opp.Library.Contains(below) {
		t.Error("the creature below it left the library")
	}
	if !opp.Library.Contains(land) {
		t.Error("the revealed land was not shuffled back into the library")
	}
}

// --- Lightwielder Paladin ----------------------------------------------------------------

func TestS82OldB500LightwielderPaladinExilesABlackOrRedPermanentOfTheDamagedPlayer(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		paladin := o500Perm(g, me.ID, "Lightwielder Paladin", "Creature — Human Knight", o500Paladin, 4, 4, false)
		if !effectstest.HasKeywordOf(g, paladin, "first strike") {
			t.Fatal("the Paladin lacks first strike")
		}
		black := o500Perm(g, opp.ID, "Black Thing", o500Bear, "", 2, 2, false)
		effectstest.HmlSetColors(g, black, "B")
		red := o500Perm(g, opp.ID, "Red Thing", o500Bear, "", 2, 2, false)
		effectstest.HmlSetColors(g, red, "R")
		green := o500Perm(g, opp.ID, "Green Thing", o500Bear, "", 2, 2, false)
		effectstest.HmlSetColors(g, green, "G")
		mine := o500Perm(g, me.ID, "My Black Thing", o500Bear, "", 2, 2, false)
		effectstest.HmlSetColors(g, mine, "B")
		effectstest.DeclareAttack(t, g, opp.ID, paladin)
		effectstest.AdvanceTo(t, g, game.StepFirstStrikeDamage)
		effectstest.B04WaitForPick(t, g, me.ID)
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("no target prompt after combat damage")
		}
		for _, id := range []uuid.UUID{black, red} {
			if !effectstest.HasID(p.PickTargetCards, id) {
				t.Errorf("the damaged player's black or red permanent %s was not offered", id)
			}
		}
		for _, id := range []uuid.UUID{green, mine} {
			if effectstest.HasID(p.PickTargetCards, id) {
				t.Errorf("a permanent that is not a black or red one of the damaged player's was offered: %s", id)
			}
		}
		effectstest.PickCard(t, g, me.ID, red)
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.LatestTriggerPromptFor(g, me.ID); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, accept); err != nil {
				t.Fatalf("trigger prompt: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		} else {
			t.Fatal("the \"you may\" was never asked")
		}
		gone := effectstest.FindBattlefieldCardForTest(g, red) == nil
		if gone != accept {
			t.Errorf("accept=%v but the red permanent's exile is %v", accept, gone)
		}
		if effectstest.FindBattlefieldCardForTest(g, black) == nil {
			t.Error("the black permanent was exiled too")
		}
	}
}

// --- Mirror of Fate ----------------------------------------------------------------------

func TestS82OldB500MirrorOfFateReplacesTheLibraryWithTheChosenExiledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mirror := effectstest.B12Permanent(g, me.ID, "Mirror of Fate", "Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mirror {
				g.Battlefield.Cards[i].OracleID = o500Mirror
			}
		}
	})
	var keep, other, faceDown, theirs, lib1, lib2 uuid.UUID
	g.WithWriteLock(func() {
		keep, other, faceDown, theirs, lib1, lib2 = uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
		g.Exile.PushTop(game.Card{InstanceID: keep, Name: "Keep", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		g.Exile.PushTop(game.Card{InstanceID: other, Name: "Other", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		g.Exile.PushTop(game.Card{InstanceID: faceDown, Name: "Hidden", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID, FaceDown: true})
		g.Exile.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		me.Library.PushTop(game.Card{InstanceID: lib1, Name: "Lib One", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: lib2, Name: "Lib Two", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	g.WithWriteLock(func() {
		for i := 0; i < 7; i++ {
			g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
	})
	effectstest.ApaActivate(t, g, me, mirror, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no choose-cards prompt")
	}
	if p.ChooseMax != 7 || p.ChooseMin != 0 {
		t.Errorf("the choice is %d..%d cards, want up to seven", p.ChooseMin, p.ChooseMax)
	}
	for _, id := range []uuid.UUID{keep, other} {
		if !effectstest.HasID(p.ChooseCards, id) {
			t.Errorf("your face-up exiled card %s was not offered", id)
		}
	}
	for _, id := range []uuid.UUID{faceDown, theirs, lib1, lib2} {
		if effectstest.HasID(p.ChooseCards, id) {
			t.Errorf("%s was offered, but it is not a face-up exiled card you own", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, keep)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, mirror) != nil {
		t.Error("Mirror of Fate was not sacrificed")
	}
	if n := len(me.Library.Cards); n != 1 || me.Library.Cards[0].InstanceID != keep {
		t.Errorf("the library is %d cards, want exactly the chosen one", n)
	}
	for _, id := range []uuid.UUID{lib1, lib2} {
		if !g.Exile.Contains(id) {
			t.Errorf("library card %s was not exiled", id)
		}
	}
}

// --- Magebane Armor ----------------------------------------------------------------------

func TestS82OldB500MagebaneArmorPumpsRemovesFlyingAndPreventsNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	armor := effectstest.SeedEquipment(g, me.ID, "Magebane Armor", o500Armor)
	host := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 3, 3)
	effectstest.FloatForTest(g, me, "{2}")
	effectstest.EquipTo(t, g, me.ID, armor, host)
	if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 6 {
		t.Errorf("the equipped creature is %d/%d, want 4/6", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the equipped creature still has flying")
	}
	effectstest.Cmdr2B08Damage(g, src, host, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, host); n != 0 {
		t.Errorf("noncombat damage was not prevented: %d marked", n)
	}
}
