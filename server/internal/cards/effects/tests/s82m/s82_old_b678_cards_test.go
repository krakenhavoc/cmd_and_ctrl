package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b678_cards_test.go — S82 OLD batch 678 (#5618, Nemesis: damage,
// fading, enchantments). Each test plays the real catalog entry.

const (
	oldB678ArcMage       = "011bb6cc-f1a1-4790-befc-373acabf63b8"
	oldB678Bully         = "410acdbc-3a00-49db-a712-ad970d24b363"
	oldB678Rackling      = "41d116d7-3b1e-4777-b3c0-54f8ca6b2785"
	oldB678RootwaterThf  = "68abaca6-a86a-46e3-b5b8-9860bb6425d2"
	oldB678Rupture       = "d2179a54-4b45-4b67-a31c-aa1a7bcfda2c"
	oldB678ParallaxWave  = "0accfb04-bc8b-4cab-997a-2337ac42db8f"
	oldB678ParallaxNexus = "40fbcf4b-d79e-4e6b-bcec-1d1544b3c4a2"
	oldB678ParallaxDem   = "6ec7a7fb-0bf9-47fa-af2b-5b13b3af3947"
	oldB678AetherBarrier = "0fe602b7-9f88-4d3d-af24-7790df867ed5"
	oldB678Asylum        = "91924536-7a2b-44ec-9835-2efa402c83f9"
	oldB678RisingWaters  = "d4cbf380-1a05-48ea-9187-de931b00866d"
)

func TestOldB678EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB678ArcMage: "Arc Mage", oldB678Bully: "Spiteful Bully", oldB678Rackling: "Rackling",
		oldB678RootwaterThf: "Rootwater Thief", oldB678Rupture: "Rupture", oldB678ParallaxWave: "Parallax Wave",
		oldB678ParallaxNexus: "Parallax Nexus", oldB678ParallaxDem: "Parallax Dementia",
		oldB678AetherBarrier: "Aether Barrier", oldB678Asylum: "Spiritual Asylum", oldB678RisingWaters: "Rising Waters",
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

func oldB678Hand(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Arc Mage ----------------------------------------------------------------

func TestOldB678ArcMageDividesTwoDamageAndNeedsADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	mage := effectstest.B12Push(g, me.ID, "Arc Mage", "Creature — Human Spellshaper", oldB678ArcMage, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: victim}, {Kind: game.TargetPlayer, ID: opp.ID}}
	dist := map[uuid.UUID]int{victim: 1, opp.ID: 1}
	for me.Hand.Size() > 0 {
		me.Hand.PopTop()
	}
	effectstest.FloatForTest(g, me, "CCR")
	params := func(discard ...uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Strict: true, Targets: targets, Distribution: dist, DiscardIDs: discard}
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, params()); err == nil {
		t.Fatal("activated with no card to discard")
	}
	fodder := oldB678Hand(me, "Fodder", "Sorcery")
	effectstest.FloatForTest(g, me, "CCR")
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, params(fodder)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-1)
	}
	if got := effectstest.B12ZoneOf(g, victim); got != game.ZoneBattlefield {
		t.Errorf("a 2/2 that took 1 is in %s", got)
	}
	if me.Hand.Contains(fodder) || effectstest.B12ZoneOf(g, fodder) != game.ZoneGraveyard {
		t.Error("the discard was not paid")
	}
}

// --- Spiteful Bully ----------------------------------------------------------

func TestOldB678SpitefulBullyHitsACreatureYouControlAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	bully := effectstest.B12Push(g, me.ID, "Spiteful Bully", "Creature — Phyrexian Zombie Mercenary", oldB678Bully, 3, 3)
	pal := effectstest.B12Creature(g, me.ID, "Pal", "Creature — Bear", 2, 2)
	effectstest.B12ToMyNextUpkeep(t, g)
	// The trigger asks for its target; pick the other creature.
	effectstest.PickTriggerTarget(t, g, me.ID, pal)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pal) {
		t.Error("the targeted creature survived 3 damage")
	}
	if !g.Battlefield.Contains(bully) {
		t.Error("the Bully was not the target")
	}
}

// --- Rackling ----------------------------------------------------------------

func TestOldB678RacklingDealsThreeMinusTheHandAtAnOpponentsUpkeep(t *testing.T) {
	for hand, want := range map[int]int{0: 3, 1: 2, 3: 0, 4: 0} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMain(t, g)
		effectstest.B12Push(g, me.ID, "Rackling", "Artifact Creature — Phyrexian Construct", oldB678Rackling, 2, 2)
		for opp.Hand.Size() > 0 {
			opp.Hand.PopTop()
		}
		for i := 0; i < hand; i++ {
			oldB678Hand(opp, "Card", "Sorcery")
		}
		myLife := me.Life
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		before := opp.Life
		effectstest.PassPriorityAroundTable(t, g)
		if lost := before - opp.Life; lost != want {
			t.Errorf("hand %d: opponent lost %d, want %d", hand, lost, want)
		}
		if me.Life != myLife {
			t.Errorf("hand %d: the controller's life changed", hand)
		}
	}
}

// --- Rootwater Thief ---------------------------------------------------------

func TestOldB678RootwaterThiefGainsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	thief := effectstest.B12Push(g, me.ID, "Rootwater Thief", "Creature — Merfolk Rogue", oldB678RootwaterThf, 1, 2)
	effectstest.FloatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, thief, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, thief)
	if !game.HasKeyword(&c, "flying") {
		t.Error("no flying after the activation")
	}
}

func TestOldB678RootwaterThiefPaysTwoToExileACardFromTheirLibrary(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMain(t, g)
		thief := effectstest.B12Push(g, me.ID, "Rootwater Thief", "Creature — Merfolk Rogue", oldB678RootwaterThf, 1, 2)
		effectstest.SeedSearchLibrary(opp, game.Card{Name: "Prize", TypeLine: "Creature — Bear"}, game.Card{Name: "Chaff", TypeLine: "Sorcery"})
		effectstest.AttackWith(t, g, opp.ID, thief)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.B06AddMana(me, "C", "C")
		}
		if !pay {
			effectstest.AnswerPayUnless(t, g, me.ID, false)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.SearchChoiceFor(g, me.ID) != nil {
				t.Error("declining still searched")
			}
			continue
		}
		effectstest.AnswerPayUnless(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSearchNamed(t, g, me.ID, "Prize")
		effectstest.PassPriorityAroundTable(t, g)
		found := false
		for _, c := range g.Exile.Cards {
			if c.Name == "Prize" {
				found = true
			}
		}
		if !found {
			t.Error("the card was not exiled")
		}
	}
}

// --- Rupture -----------------------------------------------------------------

func TestOldB678RuptureSacrificesAndDamagesNonflyersAndPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	flyerCard := game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID}
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, flyerCard)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Rupture", "Sorcery", oldB678Rupture, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still there")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the nonflying bear survived 3 damage")
	}
	if !g.Battlefield.Contains(bird) {
		t.Error("the flyer was damaged")
	}
	if me.Life != mine-3 || opp.Life != theirs-3 {
		t.Errorf("life: me %d (was %d), opponent %d (was %d); want 3 off each", me.Life, mine, opp.Life, theirs)
	}
}

func TestOldB678RuptureWithNoCreatureDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Rupture", "Sorcery", oldB678Rupture, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine || opp.Life != theirs {
		t.Error("damage was dealt with nothing sacrificed")
	}
}

// --- Parallax Wave, Nexus and Dementia ---------------------------------------

func TestOldB678ParallaxWaveExilesCreaturesAndReturnsThemWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	wave := effectstest.B12Push(g, me.ID, "Parallax Wave", "Enchantment", oldB678ParallaxWave, 0, 0)
	_ = g.AddCounter(wave, game.CounterFade, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, wave, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wave}}}); err == nil {
		t.Fatal("a non-creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, wave, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, bear); got != game.ZoneExile {
		t.Fatalf("the creature is in %s, want exile", got)
	}
	if got := effectstest.CounterCount(g, wave, game.CounterFade); got != 4 {
		t.Errorf("fade counters %d, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(wave) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Bear", opp.ID); got != 1 {
		t.Errorf("Bears back under their owner: %d, want 1", got)
	}
}

func TestOldB678ParallaxWaveFadesAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	wave := effectstest.B12Push(g, me.ID, "Parallax Wave", "Enchantment", oldB678ParallaxWave, 0, 0)
	_ = g.AddCounter(wave, game.CounterFade, 1)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, wave, game.CounterFade); got != 0 {
		t.Errorf("fade counters %d, want 0", got)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wave) {
		t.Error("the Wave was not sacrificed with no counter to remove")
	}
}

func TestOldB678ParallaxNexusExilesFromHandAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	nexus := effectstest.B12Push(g, me.ID, "Parallax Nexus", "Enchantment", oldB678ParallaxNexus, 0, 0)
	_ = g.AddCounter(nexus, game.CounterFade, 5)
	for opp.Hand.Size() > 0 {
		opp.Hand.PopTop()
	}
	keep := oldB678Hand(opp, "Keep", "Sorcery")
	pick := oldB678Hand(opp, "Pick", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, nexus, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("the controller was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, nexus, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DcAnswerChooseCards(t, g, pick)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, pick); got != game.ZoneExile {
		t.Fatalf("the chosen card is in %s, want exile", got)
	}
	if !opp.Hand.Contains(keep) {
		t.Error("the card the opponent kept left their hand")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(nexus) })
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(pick) {
		t.Error("the exiled card did not return to its owner's hand")
	}
}

func TestOldB678ParallaxNexusIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	nexus := effectstest.B12Push(g, me.ID, "Parallax Nexus", "Enchantment", oldB678ParallaxNexus, 0, 0)
	_ = g.AddCounter(nexus, game.CounterFade, 5)
	oldB678Hand(opp, "Card", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, nexus, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("activated with a spell on the stack")
	}
}

func TestOldB678ParallaxDementiaPumpsThenDestroysWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	aura := effectstest.B12Push(g, me.ID, "Parallax Dementia", "Enchantment — Aura", oldB678ParallaxDem, 0, 0)
	g.WithWriteLock(func() { _ = g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}) })
	if hc := effectstest.LayeredCard(t, g, host); hc.CurrentPower() != 5 || hc.CurrentToughness() != 4 {
		t.Errorf("enchanted creature is %d/%d, want 5/4", hc.CurrentPower(), hc.CurrentToughness())
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(aura) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature was not destroyed when the Aura left")
	}
}

func TestOldB678ParallaxDementiaDoesNothingWhenItsCreatureLeavesFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	aura := effectstest.B12Push(g, me.ID, "Parallax Dementia", "Enchantment — Aura", oldB678ParallaxDem, 0, 0)
	g.WithWriteLock(func() { _ = g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}) })
	effectstest.ExileCreature(t, g, host)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura stayed with no creature")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("an unrelated creature was destroyed")
	}
}

// --- Aether Barrier ----------------------------------------------------------

func TestOldB678AetherBarrierSacrificesAPermanentUnlessTheCasterPays(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, opp.ID, "Aether Barrier", "Enchantment", oldB678AetherBarrier, 0, 0)
		land := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
		effectstest.FloatForTest(g, me, "1")
		effectstest.AlaB1CastCreature(t, g, "Bear", 2, 2)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if !pay {
			effectstest.AnswerSacrifice(t, g, me.ID, land)
			effectstest.PassPriorityAroundTable(t, g)
		}
		gone := !g.Battlefield.Contains(land)
		if gone == pay {
			t.Errorf("pay=%v: the land is gone = %v", pay, gone)
		}
	}
}

func TestOldB678AetherBarrierIgnoresNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Aether Barrier", "Enchantment", oldB678AetherBarrier, 0, 0)
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12PlayFromHand(t, g, "Idea", "Sorcery", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a sorcery asked for payment")
	}
}

// --- Spiritual Asylum --------------------------------------------------------

func TestOldB678SpiritualAsylumGivesShroudAndIsSacrificedWhenYouAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	asylum := effectstest.B12Push(g, me.ID, "Spiritual Asylum", "Enchantment", oldB678Asylum, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{mine, land} {
		c := effectstest.LayeredCard(t, g, id)
		if !game.HasKeyword(&c, "shroud") {
			t.Errorf("%s has no shroud", c.Name)
		}
	}
	if c := effectstest.LayeredCard(t, g, theirs); game.HasKeyword(&c, "shroud") {
		t.Error("an opponent's creature has shroud")
	}
	effectstest.AlaB1Attack(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(asylum) {
		t.Error("the Asylum survived its controller attacking")
	}
}

func TestOldB678SpiritualAsylumIgnoresAnOpponentsAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	asylum := effectstest.B12Push(g, opp.ID, "Spiritual Asylum", "Enchantment", oldB678Asylum, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AlaB1Attack(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(asylum) {
		t.Error("an opponent's attack sacrificed the Asylum")
	}
}

// --- Rising Waters -----------------------------------------------------------

func TestOldB678RisingWatersKeepsLandsTappedAndUntapsOnePerUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Rising Waters", "Enchantment", oldB678RisingWaters, 0, 0)
	a := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == a || g.Battlefield.Cards[i].InstanceID == b {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DcAnswerChooseCards(t, g, a)
	effectstest.PassPriorityAroundTable(t, g)
	ca, _ := g.LookupCardForEffect(a)
	cb, _ := g.LookupCardForEffect(b)
	if ca.Tapped || !cb.Tapped {
		t.Errorf("after the upkeep: first land tapped=%v, second tapped=%v; want only the second tapped", ca.Tapped, cb.Tapped)
	}
}
