package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b752_cards_test.go — Tempest, activated abilities, batch 752 (#5628,
// S82 OLD). Every test plays the real catalog entry through the engine.

const (
	oldB752Matriarch  = "1d260526-db57-486d-b0bc-d523139a0148"
	oldB752Grimoire   = "090052a8-9e5b-4024-881f-5f8b65db7aa7"
	oldB752Ranger     = "04bb6731-5662-48d5-9e48-99d378c4adaf"
	oldB752Excavator  = "2d5ebcd7-07c2-412e-b0d9-54e3a88895fd"
	oldB752Carrionet  = "24315f60-731a-40bb-85ba-eedba1b011ff"
	oldB752Troll      = "2e6d3120-d41d-4e48-b33f-0f73c876fe69"
	oldB752StormFront = "3c8044b7-3999-4825-8448-0a1a325f1b48"
	oldB752Pegasus    = "07462467-e4a3-409e-bcef-9cc92ca4c299"
	oldB752Vhati      = "1759547a-52b3-4bae-817a-658a6fd047e3"
	oldB752Licid      = "2bdd9ca3-542a-43c0-9831-c6add23890d6"
)

func TestOldB752CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB752Matriarch: "Rootwater Matriarch", oldB752Grimoire: "Phyrexian Grimoire", oldB752Ranger: "Ranger en-Vec",
		oldB752Excavator: "Excavator", oldB752Carrionet: "Carrionette", oldB752Troll: "Skyshroud Troll",
		oldB752StormFront: "Storm Front", oldB752Pegasus: "Pegasus Refuge", oldB752Vhati: "Vhati il-Dal",
		oldB752Licid: "Quickening Licid",
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

func TestOldB752RangerEnVecHasFirstStrikeAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ranger := effectstest.B12Push(g, me.ID, "Ranger en-Vec", "Creature — Human Soldier Archer Ranger", oldB752Ranger, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, ranger, "first strike") {
		t.Error("no first strike")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, ranger, 0, game.ActivateAbilityParams{}, "G")
	effectstest.P1Destroy(t, g, ranger)
	if !g.Battlefield.Contains(ranger) {
		t.Error("the regenerating creature should survive being destroyed")
	}
}

func TestOldB752SkyshroudTrollRegeneratesForOneGreenAndOneOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	troll := effectstest.B12Push(g, me.ID, "Skyshroud Troll", "Creature — Troll Giant", oldB752Troll, 3, 3)
	if err := effectstest.RavB2Activate(t, g, me.ID, troll, 0, game.ActivateAbilityParams{}, "G"); err == nil {
		t.Error("activated for {G}: the cost is {1}{G}")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	troll2 := effectstest.B12Push(g2, me2.ID, "Skyshroud Troll", "Creature — Troll Giant", oldB752Troll, 3, 3)
	effectstest.RavB2MustActivate(t, g2, me2.ID, troll2, 0, game.ActivateAbilityParams{}, "C", "G")
	effectstest.P1Destroy(t, g2, troll2)
	if !g2.Battlefield.Contains(troll2) {
		t.Error("the regenerating Troll should survive being destroyed")
	}
}

func TestOldB752StormFrontTapsOnlyACreatureWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	front := effectstest.PushCatalogPermanent(g, me.ID, "Storm Front", "Enchantment", oldB752StormFront, false)
	flier := effectstest.B12Push(g, opp.ID, "Abbey Griffin", "Creature — Griffin", "6eb74f6e-1c9d-4caf-a0d8-fa472d435346", 2, 2)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	if err := effectstest.RavB2Activate(t, g, me.ID, front, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(walker)}, "G", "G"); err == nil {
		t.Error("a creature without flying was accepted as the target")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, front, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(flier)}, "G", "G")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, flier) {
		t.Error("the flier was not tapped")
	}
}

func TestOldB752PegasusRefugeMakesAFlierForTwoAndADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	refuge := effectstest.PushCatalogPermanent(g, me.ID, "Pegasus Refuge", "Enchantment", oldB752Pegasus, false)
	card := effectstest.PushCatalogHandCard(me, "Spare Card", "Instant", "")
	effectstest.RavB2MustActivate(t, g, me.ID, refuge, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}}, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card) {
		t.Error("the card was not discarded as the cost")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Pegasus" && c.Controller == me.ID {
			found = true
			if c.Power != 1 || c.Toughness != 1 || !c.HasColor("W") || !game.HasKeyword(&c, "flying") {
				t.Errorf("token is %d/%d %v, want a 1/1 white Pegasus with flying", c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if !found {
		t.Error("no Pegasus token was created")
	}
	if err := effectstest.RavB2Activate(t, g, me.ID, refuge, 0, game.ActivateAbilityParams{}, "C", "C"); err == nil {
		t.Error("activated without discarding a card")
	}
}

func TestOldB752VhatiIlDalSetsBasePowerOrBaseToughnessToOne(t *testing.T) {
	for _, power := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		vhati := effectstest.B12Push(g, me.ID, "Vhati il-Dal", "Legendary Creature — Human Warrior", oldB752Vhati, 3, 3)
		target := effectstest.B12Creature(g, opp.ID, "Big Ogre", "Creature — Ogre", 4, 5)
		effectstest.RavB2MustActivate(t, g, me.ID, vhati, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerConfirmFor(t, g, me, power)
		effectstest.PassPriorityAroundTable(t, g)
		p, tough := effectstest.EffectivePTOf(t, g, target)
		wantP, wantT := 4, 1
		if power {
			wantP, wantT = 1, 5
		}
		if p != wantP || tough != wantT {
			t.Errorf("power choice %v: the creature is %d/%d, want %d/%d", power, p, tough, wantP, wantT)
		}
		if !effectstest.TappedForTest(t, g, vhati) {
			t.Error("Vhati was not tapped")
		}
	}
}

func TestOldB752QuickeningLicidBecomesAnAuraGivingFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	licid := effectstest.B12Push(g, me.ID, "Quickening Licid", "Creature — Licid", oldB752Licid, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, licid, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "C", "W")
	effectstest.PassPriorityAroundTable(t, g)
	l := effectstest.B12Card(t, g, licid)
	if !l.IsAura() || l.IsCreature() || l.AttachedTo.ID != bear {
		t.Fatalf("the Licid is not an Aura on the bear (aura %v, creature %v, attached %+v)", l.IsAura(), l.IsCreature(), l.AttachedTo)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the enchanted creature has no first strike")
	}
	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{}); err != nil {
		t.Fatalf("end_effect: %v", err)
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the creature kept first strike after the effect ended")
	}
	if l = effectstest.B12Card(t, g, licid); l.IsAura() || !l.IsCreature() {
		t.Error("the Licid is not a creature again")
	}
}

func TestOldB752RootwaterMatriarchStealsAnEnchantedCreatureOnlyWhileItIsEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	matriarch := effectstest.B12Push(g, me.ID, "Rootwater Matriarch", "Creature — Merfolk", oldB752Matriarch, 2, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Unenchanted: the ability resolves and nothing begins (CR 611.2b).
	effectstest.RavB2MustActivate(t, g, me.ID, matriarch, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, bear).Controller != opp.ID {
		t.Fatal("control of an unenchanted creature changed")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, matriarch).Tapped = false })
	aura := effectstest.Enchant(t, g, "Some Aura", "", bear)
	effectstest.RavB2MustActivate(t, g, me.ID, matriarch, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, bear).Controller != me.ID {
		t.Fatal("control of an enchanted creature was not taken")
	}
	// The Matriarch leaving does not end it.
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{matriarch}) })
	if effectstest.LayeredCard(t, g, bear).Controller != me.ID {
		t.Error("control ended when the Matriarch left")
	}
	// The last Aura leaving ends it for good.
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{aura}) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, bear).Controller != opp.ID {
		t.Error("control did not return when the creature stopped being enchanted")
	}
	effectstest.Enchant(t, g, "Another Aura", "", bear)
	if effectstest.LayeredCard(t, g, bear).Controller != opp.ID {
		t.Error("control came back when the creature was enchanted again")
	}
}

func TestOldB752ExcavatorGrantsLandwalkOfTheSacrificedLandsTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Excavator", "Artifact", oldB752Excavator, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	forest := effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	nonbasic := effectstest.GaPut(g, me.ID, "Odd Land", "Land — Forest", "")
	if err := effectstest.RavB2Activate(t, g, me.ID, rock, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{nonbasic}}); err == nil {
		t.Fatal("a nonbasic land was accepted as the sacrifice")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, rock, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{forest}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the land was not sacrificed")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "forestwalk") {
		t.Error("the creature did not gain forestwalk")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "islandwalk") {
		t.Error("the creature gained a landwalk the land did not have")
	}
	_ = opp
}

func TestOldB752CarrionetteExilesItselfAndTheCreatureUnlessTheirControllerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Carrionette", TypeLine: "Creature — Skeleton", OracleID: oldB752Carrionet, Owner: me.ID, Controller: me.ID})
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatMana(t, g, me, "{C}{C}{B}{B}")
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err != nil {
			t.Fatalf("activate from the graveyard: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, opp.ID) {
			t.Fatalf("pay=%v: the creature's controller was not asked to pay", pay)
		}
		if pay {
			effectstest.B06AddMana(opp, "C", "C")
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			if !g.Battlefield.Contains(victim) || !me.Graveyard.Contains(id) {
				t.Error("paid: the creature or the card was exiled anyway")
			}
		} else if g.Battlefield.Contains(victim) || me.Graveyard.Contains(id) {
			t.Error("declined: the creature or Carrionette was not exiled")
		} else if effectstest.B12ZoneOf(g, id) != game.ZoneExile || effectstest.B12ZoneOf(g, victim) != game.ZoneExile {
			t.Error("declined: they are not both in exile")
		}
	}
	// Only from the graveyard: on the battlefield the ability does not exist.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	live := effectstest.B12Push(g, me.ID, "Carrionette", "Creature — Skeleton", oldB752Carrionet, 1, 1)
	victim := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, live, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Error("Carrionette activated from the battlefield")
	}
}

func TestOldB752PhyrexianGrimoireLetsTheOpponentPickWhichTopCardIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	book := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexian Grimoire", "Artifact — Book", oldB752Grimoire, false)
	bottom := uuid.New()
	second := uuid.New()
	top := uuid.New()
	for _, c := range []struct {
		id   uuid.UUID
		name string
	}{{bottom, "Bottom Card"}, {second, "Second Card"}, {top, "Top Card"}} {
		me.Graveyard.PushTop(game.Card{InstanceID: c.id, Name: c.name, TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	effectstest.RavB2MustActivate(t, g, me.ID, book, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}, "C", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	// The opponent, not the controller, chooses between the top two.
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the Grimoire's controller was asked to choose")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, second) != game.ZoneExile {
		t.Errorf("the chosen card is in %s, want exile", effectstest.B12ZoneOf(g, second))
	}
	if !me.Hand.Contains(top) {
		t.Error("the other of the top two did not go to its owner's hand")
	}
	if !me.Graveyard.Contains(bottom) {
		t.Error("the third card was touched")
	}
}
