package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b64_cards_test.go — GAP batch 64 (#5948, S82): Fallout
// creatures. Each test plays the real catalog entry.

const (
	gapb64ThreeDogOracle  = "cc504f00-f177-4def-956d-92a2217438bb"
	gapb64ShaunOracle     = "0bd11d6c-b707-4b37-b49c-a3274b2a5a81"
	gapb64RedDeathOracle  = "0d40b2c3-9f39-45c6-8b25-12353f300613"
	gapb64ArmoryOracle    = "360f7bfd-3b4c-4e87-a251-fd55c8a62320"
	gapb64ElizabethOracle = "3cfea32e-6b52-441b-90aa-d8a137997297"
	gapb64VeronicaOracle  = "5bf01e86-d5f2-4639-9a07-f442fd1c3bb4"
	gapb64BooneOracle     = "74f4743b-06a9-48e3-9fdf-27a56660ac8e"
	gapb64SarahOracle     = "b368b1bd-3273-40a9-8102-27501f4d0e15"
	gapb64MoiraOracle     = "db02644b-820f-4102-b1f6-ed8fac40baf8"
	gapb64CassOracle      = "f9b6b1db-37c6-43d4-b762-6c9770da5964"
	gapb64DesdemonaOracle = "fd06d110-417c-42f5-a7b2-fd032f631dd2"
)

func TestGapb64CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb64ThreeDogOracle:  "Three Dog, Galaxy News DJ",
		gapb64ShaunOracle:     "Shaun, Father of Synths",
		gapb64RedDeathOracle:  "Red Death, Shipwrecker",
		gapb64ArmoryOracle:    "Armory Paladin",
		gapb64ElizabethOracle: "Paladin Elizabeth Taggerdy",
		gapb64VeronicaOracle:  "Veronica, Dissident Scribe",
		gapb64BooneOracle:     "Craig Boone, Novac Guard",
		gapb64SarahOracle:     "Sentinel Sarah Lyons",
		gapb64MoiraOracle:     "Moira Brown, Guide Author",
		gapb64CassOracle:      "Cass, Hand of Vengeance",
		gapb64DesdemonaOracle: "Desdemona, Freedom's Edge",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func gapb64Named(g *game.Game, owner uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == owner {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

func gapb64Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func gapb64Ready(g *game.Game) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
}

// --- Red Death, Shipwrecker ---------------------------------------------------------

func TestGapb64RedDeathGoadsDrawsForThemAndAddsRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rd := gapb64Creature(g, me.ID, "Red Death, Shipwrecker", "Legendary Creature — Crab Mutant", gapb64RedDeathOracle, 1, 3)
	victim := gapb64Creature(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	gapb64Ready(g)
	hand := opp.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, rd, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 {
		t.Errorf("the creature's controller drew %d cards, want 1", opp.Hand.Size()-hand)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool %d, want {R}", me.ManaPool.Total())
	}
	if c, ok := g.LookupCardForEffect(victim); !ok || !c.IsGoadedBy(me.ID) {
		t.Errorf("the creature is not goaded by Red Death's controller")
	}
}

// --- Armory Paladin -----------------------------------------------------------------

func TestGapb64ArmoryPaladinExilesTheTopCardOnlyForAuraOrEquipmentSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapb64Creature(g, me.ID, "Armory Paladin", "Creature — Human Knight", gapb64ArmoryOracle, 3, 3)
	lib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib {
		t.Fatal("a creature spell exiled a card")
	}
	effectstest.CastCatalogSpell(t, g, "Test Blade", "Artifact — Equipment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d → %d after an Equipment spell, want one exiled", lib, me.Library.Size())
	}
	effectstest.CastCatalogSpell(t, g, "Test Boon", "Enchantment — Aura", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d after an Aura spell too, want %d", me.Library.Size(), lib-2)
	}
}

// --- Veronica, Dissident Scribe -----------------------------------------------------

func TestGapb64VeronicaMakesJunkForTheFirstNonlandDiscardEachTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gapb64Creature(g, me.ID, "Veronica, Dissident Scribe", "Legendary Creature — Human Artificer Rogue", gapb64VeronicaOracle, 3, 3)
	land := uuid.New()
	spell1, spell2 := uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Test Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: spell1, Name: "Spell One", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: spell2, Name: "Spell Two", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	discard := func(id uuid.UUID) {
		g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
		effectstest.AnswerDiscard(t, g, me.ID, id)
		effectstest.PassPriorityAroundTable(t, g)
	}
	discard(land)
	if n := len(gapb64Named(g, me.ID, "Junk")); n != 0 {
		t.Fatalf("a discarded land made %d Junk", n)
	}
	discard(spell1)
	if n := len(gapb64Named(g, me.ID, "Junk")); n != 1 {
		t.Fatalf("the first nonland discard made %d Junk, want 1", n)
	}
	discard(spell2)
	if n := len(gapb64Named(g, me.ID, "Junk")); n != 1 {
		t.Errorf("a second nonland discard the same turn made Junk: %d total", n)
	}
}

// --- Sentinel Sarah Lyons -----------------------------------------------------------

func TestGapb64SarahPumpsCreaturesOnlyWhileAnArtifactEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sarah := gapb64Creature(g, me.ID, "Sentinel Sarah Lyons", "Legendary Creature — Human Knight", gapb64SarahOracle, 4, 4)
	bear := gapb64Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	_, _ = sarah, bear
	// No artifact has entered yet: the pushed creatures are not artifacts.
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Fatalf("Bear is %d power with no artifact entered, want 2", p)
	}
	rock := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{FromZone: "hand", Strict: true}); err != nil {
		t.Fatalf("cast artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("Bear is %d/%d after an artifact entered, want 4/4", p, tt)
	}
}

func TestGapb64SarahBattalionDealsDamagePerArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sarah := gapb64Creature(g, me.ID, "Sentinel Sarah Lyons", "Legendary Creature — Human Knight", gapb64SarahOracle, 4, 4)
	a := gapb64Creature(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := gapb64Creature(g, me.ID, "Bear B", "Creature — Bear", "", 2, 2)
	gapb64Creature(g, me.ID, "Rock One", "Artifact", "", 0, 0)
	gapb64Creature(g, me.ID, "Rock Two", "Artifact", "", 0, 0)
	gapb64Ready(g)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, sarah, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want 2 damage (two artifacts)", life, opp.Life)
	}
}

// --- Moira Brown, Guide Author ------------------------------------------------------

func TestGapb64MoiraMakesTheGuideAndPutsAQuestCounterOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Moira Brown, Guide Author", TypeLine: "Legendary Creature — Human Citizen",
		OracleID: gapb64MoiraOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{R}{W}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	guides := gapb64Named(g, me.ID, "Wasteland Survival Guide")
	if len(guides) != 1 {
		t.Fatalf("%d Wasteland Survival Guides, want 1", len(guides))
	}
	bear := gapb64Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(guides[0], game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
	})
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Fatalf("equipped Bear is %d power with no quest counters, want 2", p)
	}
	gapb64Ready(g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, "quest"); n != 1 {
		t.Fatalf("%d quest counters on the Bear, want 1", n)
	}
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("equipped Bear with one quest counter is %d/%d, want 3/3", p, tt)
	}
}

// --- Craig Boone, Novac Guard -------------------------------------------------------

func TestGapb64BooneNeedsTwoAttackersAndDamagesTheCreatureUnlessItsControllerTakesIt(t *testing.T) {
	for _, takeIt := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		boone := gapb64Creature(g, me.ID, "Craig Boone, Novac Guard", "Legendary Creature — Human Soldier", gapb64BooneOracle, 3, 3)
		a := gapb64Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		victim := gapb64Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", "", 5, 5)
		gapb64Ready(g)
		life := opp.Life
		effectstest.DeclareAttack(t, g, opp.ID, boone, a)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.CountersOn(g, boone, "quest"); n != 2 {
			t.Fatalf("%d quest counters, want 2", n)
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.PickCard(t, g, me.ID, victim)
		}
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestOptionPickFor(g, opp.ID)
		if pick == nil {
			t.Fatal("the creature's controller was not asked")
		}
		idx := 0
		if takeIt {
			idx = 1
		}
		effectstest.AnswerOptionPick(t, g, opp.ID, idx)
		effectstest.PassPriorityAroundTable(t, g)
		if takeIt {
			if opp.Life != life-2 {
				t.Errorf("the controller took %d, want 2", life-opp.Life)
			}
			if c := effectstest.B12Card(t, g, victim); c.DamageMarked != 0 {
				t.Errorf("the creature was damaged anyway: %d", c.DamageMarked)
			}
		} else if c := effectstest.B12Card(t, g, victim); c.DamageMarked != 2 {
			t.Errorf("the creature has %d damage, want 2", c.DamageMarked)
		}
	}
}

func TestGapb64BooneDoesNothingWhenAttackingAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boone := gapb64Creature(g, me.ID, "Craig Boone, Novac Guard", "Legendary Creature — Human Soldier", gapb64BooneOracle, 3, 3)
	gapb64Ready(g)
	effectstest.Pr7bAttack(t, g, opp.ID, boone)
	if n := effectstest.CountersOn(g, boone, "quest"); n != 0 {
		t.Errorf("%d quest counters from a lone attack", n)
	}
}

// --- Desdemona, Freedom's Edge ------------------------------------------------------

func TestGapb64DesdemonaGivesEscapeToACheapOrArtifactCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	des := gapb64Creature(g, me.ID, "Desdemona, Freedom's Edge", "Legendary Creature — Human Rogue", gapb64DesdemonaOracle, 3, 4)
	cheap := uuid.New()
	pricey := uuid.New()
	golem := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: pricey, Name: "Pricey Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: golem, Name: "Big Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{7}", Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	var fodder []uuid.UUID
	for i := 0; i < 3; i++ {
		fid := uuid.New()
		fodder = append(fodder, fid)
		me.Graveyard.PushTop(game.Card{InstanceID: fid, Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	}
	gapb64Ready(g)
	effectstest.DeclareAttack(t, g, opp.ID, des)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target asked")
	}
	if effectstest.HasID(p.PickTargetCards, pricey) {
		t.Error("a nonartifact creature card with mana value 6 was offered")
	}
	if !effectstest.HasID(p.PickTargetCards, cheap) || !effectstest.HasID(p.PickTargetCards, golem) {
		t.Errorf("offered %v, want the cheap creature and the artifact creature", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: fodder[:2], Strict: true}); err != nil {
		t.Fatalf("escape cast for its mana cost plus two exiled cards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the escaped creature did not resolve")
	}
	if n := g.Exile.Size(); n != 2 {
		t.Errorf("%d cards exiled to pay escape, want 2", n)
	}
}

// --- Cass, Hand of Vengeance --------------------------------------------------------

func TestGapb64CassReturnsTheAuraThatWasAttachedToTheDeadCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	cass := gapb64Creature(g, me.ID, "Cass, Hand of Vengeance", "Legendary Creature — Human Ranger", gapb64CassOracle, 4, 3)
	bear := gapb64Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	aura := effectstest.B15Aura(g, me.ID, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, cass)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, aura)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(aura)
	if !ok || !g.Battlefield.Contains(aura) {
		t.Fatalf("the Aura did not return to the battlefield")
	}
	if c.AttachedTo.ID != cass {
		t.Errorf("the Aura is attached to %v, want Cass", c.AttachedTo.ID)
	}
}

func TestGapb64CassDoesNothingForAPlainCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gapb64Creature(g, me.ID, "Cass, Hand of Vengeance", "Legendary Creature — Human Ranger", gapb64CassOracle, 4, 3)
	bear := gapb64Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Errorf("an unenchanted, unequipped creature dying put Cass's trigger on the stack")
	}
}

// --- Three Dog, Galaxy News DJ ------------------------------------------------------

func TestGapb64ThreeDogCopiesTheSacrificedAuraOntoTheOtherAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dog := gapb64Creature(g, me.ID, "Three Dog, Galaxy News DJ", "Legendary Creature — Human Bard", gapb64ThreeDogOracle, 3, 3)
	pal := gapb64Creature(g, me.ID, "Pal", "Creature — Bear", "", 2, 2)
	aura := effectstest.B15Aura(g, me.ID, dog)
	gapb64Ready(g)
	effectstest.DeclareAttack(t, g, opp.ID, dog, pal)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); sac != nil {
		if err := g.ResolveOwnPermanents(sac.ID, me.ID, []uuid.UUID{aura}); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	var onPal int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Pacifism" && c.AttachedTo.ID == pal {
				onPal++
			}
			if c.Name == "Pacifism" && c.AttachedTo.ID == dog {
				t.Errorf("a copy was attached to Three Dog itself")
			}
		}
	})
	if onPal != 1 {
		t.Errorf("%d Aura copies on the other attacker, want 1", onPal)
	}
}

// --- Shaun, Father of Synths --------------------------------------------------------

func TestGapb64ShaunMakesATappedAttackingSynthCopyAndExilesThemWhenHeLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	shaun := gapb64Creature(g, me.ID, "Shaun, Father of Synths", "Legendary Creature — Human Scientist", gapb64ShaunOracle, 4, 4)
	hero := gapb64Creature(g, me.ID, "Test Hero", "Legendary Creature — Human Soldier", "", 2, 2)
	gapb64Ready(g)
	effectstest.DeclareAttack(t, g, opp.ID, shaun, hero)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, hero)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tp := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID); tp != nil {
		if err := g.ResolveTriggerPrompt(tp.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The token's defender is asked as a pick over the defending players.
	if pa := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayAmount, me.ID); pa != nil {
		if err := g.ResolvePayAmount(pa.ID, me.ID, pa.PayAmount.Max); err != nil {
			t.Fatalf("ResolvePayAmount: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	copies := gapb64Named(g, me.ID, "Test Hero")
	if len(copies) != 2 {
		t.Fatalf("%d Test Heroes, want the original and one copy", len(copies))
	}
	var tok game.Card
	for _, id := range copies {
		if c, _ := g.LookupCardForEffect(id); c.IsToken() {
			tok = c
		}
	}
	if !tok.IsToken() || !tok.HasSubtype("Synth") || !tok.IsArtifact() || !tok.Tapped {
		t.Errorf("the copy is %q token=%v tapped=%v, want a tapped Synth artifact creature", tok.TypeLine, tok.IsToken(), tok.Tapped)
	}
	effectstest.KillCreature(t, g, me.ID, shaun)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapb64Named(g, me.ID, "Test Hero")); n != 1 {
		t.Errorf("%d Test Heroes after Shaun left, want only the original", n)
	}
}
