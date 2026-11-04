package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b34_cards_test.go — Secrets of Strixhaven (sos), batch 34 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB34RootOracle        = "d870948e-721a-46fd-86b6-ecd42e3aa9bd"
	socB34RouserOracle      = "f35aa497-bd0f-4642-8713-820b9f014da5"
	socB34RunBehindOracle   = "7578a192-82e5-43ff-8dee-225f313c484a"
	socB34ShadelockOracle   = "2796c3d9-9e56-40f2-8398-7a131c4657ff"
	socB34AdminOracle       = "3b8240ad-71d3-4117-8531-6c8e79bffdb9"
	socB34SendInOracle      = "955bfcb8-6477-4ed0-b6e7-8940874ae5e4"
	socB34AcolyteOracle     = "57f58d52-0014-4cac-94fc-668063cc147a"
	socB34BaneOracle        = "e7572317-6bec-4fb0-81ac-76aec04f14ad"
	socB34CharmOracle       = "d985b12e-2731-4ccf-9f51-90218b20b42a"
	socB34TrudgeOracle      = "d04bce9a-822b-4a52-8826-b503ef5f5b2e"
	socB34InterjectOracle   = "47bbadcf-11c6-4a4d-9575-1eb8284a174c"
	socB34BearTypeLine      = "Creature — Bear"
	socB34InstantTypeLine   = "Instant"
	socB34SorceryTypeLine   = "Sorcery"
	socB34VenomousWordsName = "Venomous Words"
)

func TestSocB34EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB34RootOracle:             "Root Manipulation",
		socB34RouserOracle:           "Rubble Rouser",
		socB34RunBehindOracle:        "Run Behind",
		socB34ShadelockOracle:        "Scathing Shadelock",
		socB34ShadelockOracle + "#1": socB34VenomousWordsName,
		socB34AdminOracle:            "Scolding Administrator",
		socB34SendInOracle:           "Send in the Pest",
		socB34AcolyteOracle:          "Shattered Acolyte",
		socB34BaneOracle:             "Shopkeeper's Bane",
		socB34CharmOracle:            "Silverquill Charm",
		socB34TrudgeOracle:           "Slumbering Trudge",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB34AttackWith declares `attackers` against `defender` from the declare
// attackers step of seat 0 and settles the attack triggers.
func socB34AttackWith(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceToDeclareAttackersOf(t, g, 0)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

// --- Root Manipulation ---------------------------------------------------------

func TestSocB34RootManipulationPumpsGrantsMenaceAndTheAttackTriggerToTheCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", socB34BearTypeLine, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", socB34BearTypeLine, 2, 2)
	castCatalogSpell(t, g, "Root Manipulation", socB34SorceryTypeLine, socB34RootOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 4 || tg != 4 {
		t.Errorf("my creature is %d/%d, want 4/4", p, tg)
	}
	if !hasEffectiveKeyword(t, g, mine, "menace") {
		t.Error("my creature did not gain menace")
	}
	if p := effectivePower(t, g, theirs); p != 2 || hasEffectiveKeyword(t, g, theirs, "menace") {
		t.Errorf("an opposing creature was affected: power %d", p)
	}
	// The set is fixed as the spell resolves (CR 611.2c).
	late := b12Creature(g, me.ID, "Late Bear", socB34BearTypeLine, 2, 2)
	if p := effectivePower(t, g, late); p != 2 || hasEffectiveKeyword(t, g, late, "menace") {
		t.Errorf("a creature that came later was affected: power %d", p)
	}
	life := me.Life
	socB34AttackWith(t, g, opp.ID, mine, late)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1 from the one creature that has the trigger", life, me.Life)
	}
	endTurn(t, g)
	if p := effectivePower(t, g, mine); p != 2 || hasEffectiveKeyword(t, g, mine, "menace") {
		t.Errorf("the effect outlasted the turn: power %d", p)
	}
}

// --- Rubble Rouser ------------------------------------------------------------

func socB34PlayRouser(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	msc17Play(t, g, me, game.Card{Name: "Rubble Rouser", TypeLine: "Creature — Dwarf Sorcerer", ManaCost: "{2}{R}",
		OracleID: socB34RouserOracle, Power: 1, Toughness: 4})
}

func TestSocB34RubbleRouserMayDiscardAndIfSoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn")
	discard := pushHandCard(g, me)
	socB34PlayRouser(t, g, me)
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, discard)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand after discard one, draw one: %d, want %d", got, hand)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the chosen card was not discarded")
	}
}

func TestSocB34RubbleRouserDeclinedDiscardDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn")
	pushHandCard(g, me)
	socB34PlayRouser(t, g, me)
	hand, library := me.Hand.Size(), me.Library.Size()
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Library.Size() != library {
		t.Errorf("declining drew anyway: hand %d -> %d, library %d -> %d", hand, me.Hand.Size(), library, me.Library.Size())
	}
}

func TestSocB34RubbleRouserTapsAndExilesACardForRedAndPingsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rouser := b12Push(g, me.ID, "Rubble Rouser", "Creature — Dwarf Sorcerer", socB34RouserOracle, 1, 4)
	dead := socB11Graveyard(me, "Dead Bear", socB34BearTypeLine, "{1}{G}")
	advanceToMain(t, g)
	// The exile is a cost the activator must pay: nothing named, nothing happens.
	if err := g.ActivateManaAbility(me.ID, rouser, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated without exiling a card")
	}
	lives, myLife := lifeOfOpponents(g), me.Life
	if err := g.ActivateManaAbility(me.ID, rouser, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{dead}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool holds %d mana, want {R}", me.ManaPool.Total())
	}
	if me.Graveyard.Contains(dead) {
		t.Error("the card is still in the graveyard")
	}
	if !g.Exile.Contains(dead) {
		t.Error("the card was not exiled")
	}
	if c, _ := battlefieldCard(g, rouser); !c.Tapped {
		t.Error("the Rouser did not tap")
	}
	// The "When you do" damage is a reflexive trigger on the stack (CR
	// 603.12, CR 605.1b), so nobody has taken it yet.
	for i, now := range lifeOfOpponents(g) {
		if now != lives[i] {
			t.Errorf("opponent %d took damage before the trigger resolved: %d -> %d", i, lives[i], now)
		}
	}
	passPriorityAroundTable(t, g)
	for i, now := range lifeOfOpponents(g) {
		if now != lives[i]-1 {
			t.Errorf("opponent %d life %d -> %d, want -1", i, lives[i], now)
		}
	}
	if me.Life != myLife {
		t.Errorf("the controller's life changed %d -> %d", myLife, me.Life)
	}
	// Tapped now, and with an empty graveyard: neither activation is possible.
	other := socB11Graveyard(me, "Another", socB34BearTypeLine, "{1}{G}")
	if err := g.ActivateManaAbility(me.ID, rouser, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{other}}); err == nil {
		t.Error("activated while tapped")
	}
}

func TestSocB34RubbleRouserCannotExileACardFromYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rouser := b12Push(g, me.ID, "Rubble Rouser", "Creature — Dwarf Sorcerer", socB34RouserOracle, 1, 4)
	inHand := pushHandCard(g, me)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, rouser, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{inHand}}); err == nil {
		t.Fatal("exiled a card from the hand to pay for it")
	}
	if me.ManaPool.Total() != 0 {
		t.Error("mana appeared without the cost being paid")
	}
}

// --- Run Behind ---------------------------------------------------------------

func TestSocB34RunBehindCostsOneLessOnlyWhenItTargetsAnAttackingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", socB34BearTypeLine, 2, 2)
	idle := b12Creature(g, opp.ID, "Idle", socB34BearTypeLine, 2, 2)
	advanceToMain(t, g)
	card := game.Card{Name: "Run Behind", TypeLine: socB34InstantTypeLine, ManaCost: "{3}{U}", OracleID: socB34RunBehindOracle}
	// Not attacking: it costs the full {3}{U}, so three mana cannot pay for it.
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("cast for {2}{U} at a creature that is not attacking")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	socB34AttackWith(t, g, opp.ID, attacker)
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: cardRefs(attacker)}); err != nil {
		t.Fatalf("cast for {2}{U} at an attacking creature: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left over, want all three spent", me.ManaPool.Total())
	}
}

func TestSocB34RunBehindLetsTheOwnerChooseTopOrBottom(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bottom bool
	}{{"top", false}, {"bottom", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			bear := b12Creature(g, opp.ID, "Their Bear", socB34BearTypeLine, 2, 2)
			castCatalogSpell(t, g, "Run Behind", socB34InstantTypeLine, socB34RunBehindOracle, cardRefs(bear))
			passPriorityAroundTable(t, g)
			if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
				t.Fatal("the caster was asked; the owner chooses")
			}
			c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
			if c == nil {
				t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
			}
			var top, bottom []uuid.UUID
			if tc.bottom {
				bottom = []uuid.UUID{bear}
			} else {
				top = []uuid.UUID{bear}
			}
			if err := g.ResolvePutInLibrary(c.ID, opp.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if g.Battlefield.Contains(bear) {
				t.Fatal("the creature is still on the battlefield")
			}
			cards := opp.Library.Cards
			// The bottom of a library is Cards[0], the top the last entry.
			if tc.bottom && cards[0].InstanceID != bear {
				t.Errorf("the bottom card is %q, want the creature", cards[0].Name)
			}
			if !tc.bottom && cards[len(cards)-1].InstanceID != bear {
				t.Errorf("the top card is %q, want the creature", cards[len(cards)-1].Name)
			}
		})
	}
}

// --- Scathing Shadelock // Venomous Words ----------------------------------------

func socB34ShadelockCard(owner uuid.UUID) game.Card {
	c := sos1Prep(owner, socB34ShadelockOracle, nil, "Scathing Shadelock", "Creature — Snake Warlock", "{4}{B}", 4, 6,
		socB34VenomousWordsName, socB34SorceryTypeLine, "{B}")
	c.InstanceID = uuid.New()
	return c
}

func TestSocB34ShadelockBecomesPreparedAtYourFirstMainAndVenomousWordsPumpsAndGrantsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	shade := socB34ShadelockCard(me.ID)
	castFromHandAndResolve(t, g, me, shade)
	if g.IsPreparedForEffect(shade.InstanceID) {
		t.Fatal("prepared the turn it was cast, after the first main phase began")
	}
	bear := b12Creature(g, me.ID, "My Bear", socB34BearTypeLine, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", socB34BearTypeLine, 2, 2)
	advanceTo(t, g, game.StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(shade.InstanceID) {
		t.Fatal("not prepared at the beginning of its controller's first main phase")
	}
	cp, ok := prepareCopyOf(g, socB34VenomousWordsName)
	if !ok {
		t.Fatal("no copy of Venomous Words in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("Venomous Words targeted a creature the caster does not control")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast the copy of Venomous Words: %v", err)
	}
	if g.IsPreparedForEffect(shade.InstanceID) {
		t.Error("the Shadelock is still prepared once its copy was cast")
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 2 {
		t.Errorf("target is %d/%d, want 4/2", p, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("target did not gain deathtouch")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "deathtouch") || effectivePower(t, g, bear) != 2 {
		t.Error("Venomous Words outlasted the turn")
	}
}

// --- Scolding Administrator -----------------------------------------------------

func TestSocB34AdministratorHasMenace(t *testing.T) {
	spec, ok := Lookup(socB34AdminOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "menace" {
		t.Fatalf("printed keywords %v, want [menace]", spec.PrintedKeywords)
	}
}

func TestSocB34AdministratorGetsACounterForAnInstantTargetingACreatureAndOnlyThen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	admin := b12Push(g, me.ID, "Scolding Administrator", "Creature — Dwarf Cleric", socB34AdminOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", socB34BearTypeLine, 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Interjection", socB34InstantTypeLine, socB34InterjectOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := countersOn(g, admin, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters after a creature-targeting instant, want 1", got)
	}
	// A spell that targets only a player is not Repartee.
	castCatalogSpell(t, g, "Stream", socB34SorceryTypeLine, "", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, admin, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after a spell with a player target, want still 1", got)
	}
}

func TestSocB34AdministratorMovesItsCountersToAnotherCreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	admin := b12Push(g, me.ID, "Scolding Administrator", "Creature — Dwarf Cleric", socB34AdminOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", socB34BearTypeLine, 2, 2)
	if err := g.AddCounterForEffect(admin, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	if err := g.AddCounterForEffect(admin, "charge", 1); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	advanceToMain(t, g)
	b27Damage(g, bear, admin, 10)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(admin) {
		t.Fatal("the Administrator did not die")
	}
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("target has %d +1/+1 counters, want the 2 it had", got)
	}
	if got := countersOn(g, bear, "charge"); got != 1 {
		t.Errorf("target has %d charge counters, want the 1 it had", got)
	}
}

func TestSocB34AdministratorWithoutCountersDoesNotTriggerWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	admin := b12Push(g, me.ID, "Scolding Administrator", "Creature — Dwarf Cleric", socB34AdminOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", socB34BearTypeLine, 2, 2)
	advanceToMain(t, g)
	b27Damage(g, bear, admin, 10)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(admin) {
		t.Fatal("the Administrator did not die")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for though it had no counters")
	}
	if n := triggersOnStackFrom(g, admin); n != 0 {
		t.Errorf("%d triggers on the stack, want none", n)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("target has %d counters from nothing", got)
	}
}

// --- Send in the Pest -----------------------------------------------------------

func TestSocB34SendInThePestMakesEachOpponentDiscardAndMakesAPest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opps := g.Seats[1:]
	cards := make([]uuid.UUID, len(opps))
	for i, p := range opps {
		cards[i] = pushHandCard(g, p)
	}
	myHand := me.Hand.Size()
	castCatalogSpell(t, g, "Send in the Pest", socB34SorceryTypeLine, socB34SendInOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range opps {
		if discardOwed(g, p.ID) != 1 {
			t.Fatalf("opponent %d is not asked to discard one card", i)
		}
		answerDiscard(t, g, p.ID, cards[i])
		if !p.Graveyard.Contains(cards[i]) {
			t.Errorf("opponent %d's chosen card was not discarded", i)
		}
	}
	if me.Hand.Size() != myHand {
		t.Errorf("the caster's hand changed: %d -> %d", myHand, me.Hand.Size())
	}
	pests := socB26Tokens(g, me.ID, "Pest")
	if len(pests) != 1 {
		t.Fatalf("%d Pests, want 1", len(pests))
	}
	p := pests[0]
	if p.CurrentPower() != 1 || p.CurrentToughness() != 1 || len(p.Colors) != 2 {
		t.Errorf("token %d/%d colors %v, want a 1/1 black and green Pest", p.CurrentPower(), p.CurrentToughness(), p.Colors)
	}
}

func TestSocB34SendInThePestPestGainsLifeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Send in the Pest", socB34SorceryTypeLine, socB34SendInOracle, nil)
	passPriorityAroundTable(t, g)
	pests := socB26Tokens(g, me.ID, "Pest")
	if len(pests) != 1 {
		t.Fatalf("%d Pests, want 1", len(pests))
	}
	for _, p := range g.Seats[1:] {
		discardFromHand(t, g, p.ID)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	life := me.Life
	socB34AttackWith(t, g, opp.ID, pests[0].InstanceID)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1 from the Pest's attack trigger", life, me.Life)
	}
}

func TestSocB34SendInThePestDoesNotMakeTheCasterDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushHandCard(g, me)
	castCatalogSpell(t, g, "Send in the Pest", socB34SorceryTypeLine, socB34SendInOracle, nil)
	passPriorityAroundTable(t, g)
	if discardOwed(g, me.ID) != 0 {
		t.Error("the caster was asked to discard")
	}
}

// --- Shattered Acolyte --------------------------------------------------------------

func TestSocB34AcolyteHasLifelink(t *testing.T) {
	spec, ok := Lookup(socB34AcolyteOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "lifelink" {
		t.Fatalf("printed keywords %v, want [lifelink]", spec.PrintedKeywords)
	}
}

func TestSocB34AcolyteSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	acolyte := b12Push(g, me.ID, "Shattered Acolyte", "Creature — Dwarf Warlock", socB34AcolyteOracle, 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	ench := b12Permanent(g, opp.ID, "Aura", "Enchantment")
	bear := b12Creature(g, opp.ID, "Bear", socB34BearTypeLine, 2, 2)
	advanceToMain(t, g)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a plain creature is not an artifact or enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(acolyte) {
		t.Fatal("the Acolyte was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("the other target was destroyed too")
	}
	// The same ability can take an enchantment.
	acolyte2 := b12Push(g, me.ID, "Shattered Acolyte", "Creature — Dwarf Warlock", socB34AcolyteOracle, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, acolyte2, 0, game.ActivateAbilityParams{Targets: cardRefs(ench)}); err != nil {
		t.Fatalf("ActivateCatalogAbility at an enchantment: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
}

func TestSocB34AcolyteNeedsItsGenericMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	acolyte := b12Push(g, me.ID, "Shattered Acolyte", "Creature — Dwarf Warlock", socB34AcolyteOracle, 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	advanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Targets: cardRefs(rock), Strict: true})
	if err == nil {
		t.Fatal("activated for no mana")
	}
	if !g.Battlefield.Contains(acolyte) {
		t.Errorf("the Acolyte was sacrificed by a refused activation (%v)", err)
	}
}

// --- Shopkeeper's Bane ---------------------------------------------------------------

func TestSocB34BaneHasTrample(t *testing.T) {
	spec, ok := Lookup(socB34BaneOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "trample" {
		t.Fatalf("printed keywords %v, want [trample]", spec.PrintedKeywords)
	}
}

func TestSocB34BaneGainsTwoLifeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bane := b12Push(g, me.ID, "Shopkeeper's Bane", "Creature — Badger Pest", socB34BaneOracle, 4, 2)
	life := me.Life
	socB34AttackWith(t, g, opp.ID, bane)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

// --- Silverquill Charm -----------------------------------------------------------------

func TestSocB34CharmPutsTwoCountersOnATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", socB34BearTypeLine, 5, 5)
	castModal(t, g, "Silverquill Charm", socB34InstantTypeLine, socB34CharmOracle, []int{0}, []game.TargetRef{modeTargetRef(bear, 0)})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
}

func TestSocB34CharmExilesACreatureWithPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", socB34BearTypeLine, 2, 2)
	castModal(t, g, "Silverquill Charm", socB34InstantTypeLine, socB34CharmOracle, []int{1}, []game.TargetRef{modeTargetRef(small, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the creature is still on the battlefield")
	}
	if !g.Exile.Contains(small) {
		t.Error("the creature was not exiled")
	}
}

func TestSocB34CharmCannotTargetACreatureWithPowerThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", socB34BearTypeLine, 3, 3)
	advanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Silverquill Charm", TypeLine: socB34InstantTypeLine, ManaCost: "{W}{B}",
		OracleID: socB34CharmOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{modeTargetRef(big, 0)}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("cast at a power-3 creature: %v, want ErrIllegalTarget", err)
	}
}

func TestSocB34CharmDrainsEachOpponentForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lives := lifeOfOpponents(g)
	life := me.Life
	castModal(t, g, "Silverquill Charm", socB34InstantTypeLine, socB34CharmOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	for i, now := range lifeOfOpponents(g) {
		if now != lives[i]-3 {
			t.Errorf("opponent %d life %d -> %d, want -3", i, lives[i], now)
		}
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
}

// --- Slumbering Trudge -------------------------------------------------------------------

func socB34CastTrudge(t *testing.T, x int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Slumbering Trudge", TypeLine: "Creature — Plant Beast",
		ManaCost: "{X}{G}", OracleID: socB34TrudgeOracle, Power: 6, Toughness: 6}
	if err := socB19CastWithPool(g, me, card, "G", x+1, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast with X=%d: %v", x, err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Fatalf("the Trudge did not enter (X=%d)", x)
	}
	return g, card.InstanceID
}

func TestSocB34TrudgeEntersWithThreeMinusXStunCountersAndTappedWhenXIsTwoOrLess(t *testing.T) {
	for _, tc := range []struct {
		x      int
		stun   int
		tapped bool
	}{{0, 3, true}, {1, 2, true}, {2, 1, true}, {3, 0, false}, {4, 0, false}} {
		g, id := socB34CastTrudge(t, tc.x)
		c, _ := battlefieldCard(g, id)
		if got := countersOn(g, id, game.CounterStun); got != tc.stun {
			t.Errorf("X=%d: %d stun counters, want %d", tc.x, got, tc.stun)
		}
		if c.Tapped != tc.tapped {
			t.Errorf("X=%d: tapped = %v, want %v", tc.x, c.Tapped, tc.tapped)
		}
	}
}

func TestSocB34TrudgeIsSixBySix(t *testing.T) {
	g, id := socB34CastTrudge(t, 3)
	if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 6 || tg != 6 {
		t.Errorf("Trudge is %d/%d, want 6/6", p, tg)
	}
}
