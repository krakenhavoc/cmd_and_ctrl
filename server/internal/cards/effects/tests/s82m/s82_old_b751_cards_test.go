package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b751_cards_test.go — Stronghold, enters and dies triggers, card draw
// and library, graveyard recursion, lands, Auras, artifacts and token makers,
// batch 751 (#5627, S82 OLD). Every test plays the real catalog entry through
// the engine.

const (
	oldB751Mask       = "f1a16f6d-ea11-485e-8c5b-7f919a731fe0"
	oldB751Heartstone = "8f8ebf49-37c5-4da0-9cee-84ad22ab5d37"
	oldB751Mortuary   = "8bd2af55-5a48-4300-a893-ae6cd548db24"
	oldB751Lab        = "f2a6e26c-6bd6-4fca-ba9f-d52d01fc03ef"
	oldB751Phoenix    = "3b629793-0efb-4d60-9d82-38b7b400abfd"
	oldB751Stronghold = "73b8cf90-3c71-4f8b-a29f-61894b7f27c9"
	oldB751Mogg       = "cfd61b53-83c6-4886-8503-ddcb92dc7f85"
	oldB751Portcullis = "6230c179-0263-4fa4-b0ff-f2cb621de147"
	oldB751Contempt   = "811c8364-fafb-49c8-a5bf-7b6cdfceb475"
)

func TestOldB751CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB751Mask: "Mask of the Mimic", oldB751Heartstone: "Heartstone", oldB751Mortuary: "Mortuary",
		oldB751Lab: "Volrath's Laboratory", oldB751Phoenix: "Shard Phoenix", oldB751Stronghold: "Volrath's Stronghold",
		oldB751Mogg: "Mogg Infestation", oldB751Portcullis: "Portcullis", oldB751Contempt: "Contempt",
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

func TestOldB751HeartstoneTakesOneOffAnyCreaturesAbilityButLeavesOneMana(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.PushCatalogPermanent(g, me.ID, "Heartstone", "Artifact", oldB751Heartstone, false)
	for _, tc := range []struct {
		who, source uuid.UUID
		name        string
		cost, want  string
	}{
		{me.ID, mine, "my creature", "{3}", "{2}"},
		{me.ID, mine, "my creature", "{2}{G}", "{1}{G}"},
		{opp.ID, theirs, "an opponent's creature", "{3}", "{2}"},
		{me.ID, mine, "my creature", "{1}", "{1}"},
		{me.ID, mine, "my creature", "{R}{R}", "{R}{R}"},
		{me.ID, rock, "an artifact", "{3}", "{3}"},
	} {
		got := effectstest.LcPrice(t, g, tc.who, tc.source, tc.cost, 0)
		if got != tc.want {
			t.Errorf("%s: %s costs %s under Heartstone, want %s", tc.name, tc.cost, got, tc.want)
		}
	}
}

func TestOldB751MortuaryPutsYourDeadCreatureOnTopOfYourLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mortuary", "Enchantment", oldB751Mortuary, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.SeedLibrary(opp, "X", "Y", "Z")
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{mine, theirs}) })
	effectstest.PassPriorityAroundTable(t, g)
	top := func(p *game.Player) uuid.UUID { return p.Library.Cards[len(p.Library.Cards)-1].InstanceID }
	if top(me) != mine || me.Graveyard.Contains(mine) {
		t.Error("my creature is not on top of my library")
	}
	if !opp.Graveyard.Contains(theirs) || top(opp) == theirs {
		t.Error("an opponent's creature, in their graveyard, was put on top of a library")
	}
}

func TestOldB751MortuaryPutsAStolenCreatureOnTopOfItsOwnersLibraryOnlyIfYouOwnIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mortuary", "Enchantment", oldB751Mortuary, false)
	// A creature I own, under their control, goes to MY graveyard.
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, mine).Controller = opp.ID })
	effectstest.SeedLibrary(me, "A", "B")
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{mine}) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; got != mine {
		t.Error("a creature put into my graveyard under another player's control was not put on top of my library")
	}
}

func TestOldB751MaskOfTheMimicSacrificesAndFetchesTheTargetsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	target := effectstest.B12Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "U")
	id := effectstest.HandCardOf(g, me, "Mask of the Mimic", "Instant", "{U}", oldB751Mask)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(token), SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("Mask of the Mimic targeted a token")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(target)}); err == nil {
		t.Error("cast without sacrificing a creature")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(target), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if effectstest.FindBattlefieldCardForTest(g, ids[0]) == nil {
		t.Error("the Grizzly Bears did not enter the battlefield")
	}
	if effectstest.FindBattlefieldCardForTest(g, ids[1]) != nil {
		t.Error("a differently named card entered")
	}
}

func TestOldB751PortcullisExilesAThirdCreatureAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	portcullis := effectstest.PushCatalogPermanent(g, me.ID, "Portcullis", "Artifact", oldB751Portcullis, false)
	effectstest.B12Creature(g, me.ID, "Bear One", "Creature — Bear", 2, 2)
	// With only one other creature on the battlefield the second is not exiled.
	second := effectstest.EnterCreature(t, g, opp.ID, "Bear Two", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(second) {
		t.Fatal("a creature was exiled with only one other creature on the battlefield")
	}
	// Two others: the third is exiled.
	third := effectstest.EnterCreature(t, g, opp.ID, "Bear Three", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(third) {
		t.Fatal("the third creature was not exiled")
	}
	if effectstest.B12ZoneOf(g, third) != game.ZoneExile {
		t.Errorf("the third creature is in %s, want exile", effectstest.B12ZoneOf(g, third))
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{portcullis}) })
	effectstest.PassPriorityAroundTable(t, g)
	// A card returned from exile is a new object (CR 400.7), so it is found by name.
	returned := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear Three" {
			returned++
			if c.Controller != opp.ID {
				t.Error("the creature returned under the wrong player's control")
			}
		}
	}
	if returned != 1 {
		t.Errorf("%d copies of the exiled creature on the battlefield, want 1 back under its owner's control", returned)
	}
}

func TestOldB751VolrathsStrongholdMakesColorlessAndRebuysACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Volrath's Stronghold", "Legendary Land", oldB751Stronghold, false)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want one {C}", got)
	}
	me.ManaPool = nil
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, land).Tapped = false })
	effectstest.SeedLibrary(me, "A", "B")
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Spent Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := effectstest.RavB2Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(spell)}, "1", "B"); err == nil {
		t.Error("a non-creature card was accepted as the target")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)}, "C", "B")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(dead) || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != dead {
		t.Error("the creature card is not on top of the library")
	}
}

func TestOldB751ShardPhoenixSacrificesToHitCreaturesWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phoenix := effectstest.B12Push(g, me.ID, "Shard Phoenix", "Creature — Phoenix", oldB751Phoenix, 2, 2)
	ground := effectstest.B12Creature(g, opp.ID, "Ground Bear", "Creature — Bear", 2, 2)
	bigGround := effectstest.B12Creature(g, opp.ID, "Big Ground", "Creature — Bear", 3, 3)
	// Abbey Griffin is a catalog 2/2 with flying and vigilance.
	flier := effectstest.B12Push(g, opp.ID, "Abbey Griffin", "Creature — Griffin", "6eb74f6e-1c9d-4caf-a0d8-fa472d435346", 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, phoenix, "flying") {
		t.Fatal("the Phoenix has no flying")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, phoenix, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(phoenix) {
		t.Error("the Phoenix was not sacrificed")
	}
	if g.Battlefield.Contains(ground) {
		t.Error("a 2/2 creature without flying survived 2 damage")
	}
	if !g.Battlefield.Contains(bigGround) || !g.Battlefield.Contains(flier) {
		t.Error("the 3/3 or the flier died")
	}
	if got := effectstest.LayeredCard(t, g, bigGround).DamageMarked; got != 2 {
		t.Errorf("the 3/3 has %d damage, want 2", got)
	}
	if got := effectstest.LayeredCard(t, g, flier).DamageMarked; got != 0 {
		t.Errorf("the flier has %d damage, want 0", got)
	}
}

func TestOldB751ShardPhoenixReturnsToHandOnlyDuringYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Shard Phoenix", TypeLine: "Creature — Phoenix", OracleID: oldB751Phoenix, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated in the main phase")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	id2 := uuid.New()
	me2.Graveyard.PushTop(game.Card{InstanceID: id2, Name: "Shard Phoenix", TypeLine: "Creature — Phoenix", OracleID: oldB751Phoenix, Owner: me2.ID, Controller: me2.ID})
	effectstest.AdvanceToUpkeepOfSeat(t, g2, 0)
	effectstest.FloatMana(t, g2, me2, "{R}{R}{R}")
	if err := g2.ActivateCatalogAbility(me2.ID, id2, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate during upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !me2.Hand.Contains(id2) || me2.Graveyard.Contains(id2) {
		t.Errorf("the Phoenix did not return to hand (zone %s)", effectstest.B12ZoneOf(g2, id2))
	}
}

func TestOldB751MoggInfestationDestroysTheirCreaturesAndGivesThemTwoGoblinsEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Their Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Bear B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCRR")
	id := effectstest.HandCardOf(g, me, "Mogg Infestation", "Sorcery", "{3}{R}{R}", oldB751Mogg)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was destroyed")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("one of their creatures survived")
	}
	goblins := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.IsToken() {
			goblins++
			if c.Controller != opp.ID || c.Power != 1 || c.Toughness != 1 || !c.HasColor("R") {
				t.Errorf("a Goblin is %d/%d under %v, want a 1/1 red one under the target player", c.Power, c.Toughness, c.Controller)
			}
		}
	}
	if goblins != 4 {
		t.Errorf("%d Goblin tokens, want 4 (two for each of two creatures)", goblins)
	}
}

func TestOldB751ContemptReturnsTheCreatureAndTheAuraAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, opp.ID, "Contempt", "Enchantment — Aura", oldB751Contempt, false)
	effectstest.NcdAttach(t, g, aura, theirs)
	effectstest.DeclareAttack(t, g, opp.ID, theirs)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(aura) {
		t.Fatal("the creature or the Aura left before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !effectstest.HandHasID(me, theirs) {
		t.Error("the enchanted creature did not return to its owner's hand")
	}
	if g.Battlefield.Contains(aura) || !effectstest.HandHasID(opp, aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
}

func TestOldB751VolrathsLaboratoryMakesATwoTwoOfTheChosenColorAndType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lab := effectstest.PushCatalogPermanent(g, me.ID, "Volrath's Laboratory", "Artifact", oldB751Lab, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, lab)
		c.ChosenColor, c.NamedTribe = "G", "Elf"
	})
	effectstest.RavB2MustActivate(t, g, me.ID, lab, 0, game.ActivateAbilityParams{}, "C", "C", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == me.ID {
			found = true
			if c.Name != "Elf" || !c.HasSubtype("Elf") || !c.IsCreature() || c.Power != 2 || c.Toughness != 2 || !c.HasColor("G") || c.HasColor("W") {
				t.Errorf("token is %q %d/%d %v, want a 2/2 green Elf creature", c.Name, c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if !found {
		t.Error("no token was created")
	}
}

func TestOldB751VolrathsLaboratoryAsksForAColorAndACreatureTypeAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lab := effectstest.HandCardOf(g, me, "Volrath's Laboratory", "Artifact", "{5}", oldB751Lab)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCCC")
	if err := g.CastSpell(me.ID, lab, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	kinds := map[game.PendingChoiceKind]bool{}
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			kinds[c.Kind] = true
		}
	}
	if len(kinds) < 2 {
		t.Errorf("pending prompts %v, want both a colour and a creature-type choice", kinds)
	}
}
