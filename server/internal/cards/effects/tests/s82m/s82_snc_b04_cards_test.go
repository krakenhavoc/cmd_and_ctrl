package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b04_cards_test.go — Streets of New Capenna, batch 4 (#5491, S82):
// the connive creatures and attachments. Each test plays the real catalog
// entry.

const (
	sncInformantOracle   = "838c4831-7d5c-4025-92b3-46eee59d6b34"
	sncSnoopOracle       = "a4e8d2ad-cce2-45de-b9a2-7ba3b1546efb"
	sncInspectorOracle   = "d7ae0009-3747-4f3a-a613-3e640b0219a0"
	sncRuinerOracle      = "adfb28ca-2ba6-4104-b731-f8fd9367f7ec"
	sncGrifterOracle     = "5df80355-b12a-4989-94e5-1e0fa27350e8"
	sncSilencerOracle    = "2901948c-2a5a-4135-9135-27406e9d8652"
	sncVirtuosoOracle    = "ed2af15e-955b-43a1-80af-2484be373db3"
	sncBypassOracle      = "cb660c6d-dc8b-4779-8613-6acf19418baa"
	sncPickpocketOracle  = "60ead8a5-11fb-4eeb-99bf-0caaf1de1731"
	sncInterceptorOracle = "b0870f91-c544-446a-8f7b-a8d03fdb0a2a"
	sncMaskOracle        = "584f9c45-9737-47bb-8ba4-a7cd9b0b2c1f"
)

func TestSncB04CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncInformantOracle: "Raffine's Informant", sncSnoopOracle: "Psionic Snoop",
		sncInspectorOracle: "Echo Inspector", sncRuinerOracle: "Revel Ruiner",
		sncGrifterOracle: "Hypnotic Grifter", sncSilencerOracle: "Raffine's Silencer",
		sncVirtuosoOracle: "Illuminator Virtuoso", sncBypassOracle: "Security Bypass",
		sncPickpocketOracle: "Psychic Pickpocket", sncInterceptorOracle: "Obscura Interceptor",
		sncMaskOracle: "Mask of the Schemer",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// sncPrinted is each creature's printed power and toughness, which the
// hand-seeded card carries since the test builds it by hand.
var sncPrinted = map[string][2]int{
	sncInformantOracle: {2, 1}, sncSnoopOracle: {0, 3}, sncInspectorOracle: {2, 3},
	sncRuinerOracle: {3, 1}, sncSilencerOracle: {1, 1}, sncPickpocketOracle: {3, 2},
	sncInterceptorOracle: {3, 1},
}

// sncDiscard answers the open connive discard prompt with `card`.
func sncDiscard(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID) {
	t.Helper()
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, p.ID)
	if pick == nil {
		t.Fatal("no discard prompt is open")
	}
	if err := g.ResolveChooseCards(pick.ID, p.ID, []uuid.UUID{card}); err != nil {
		t.Fatalf("discard: %v", err)
	}
}

// sncEnters casts the creature, lets its enters trigger resolve and answers
// the connive discard with a nonland card (or a land when discardLand).
// It returns the creature and the board it entered.
func sncEnters(t *testing.T, g *game.Game, name, typeLine, oracle string, discardLand bool) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := effectstest.HandCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	land := effectstest.HandCardFull(me, "Land B", "Basic Land — Forest", "", "", nil)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: sncPrinted[oracle][0], Toughness: sncPrinted[oracle][1], Owner: me.ID, Controller: me.ID,
	})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) == nil {
		t.Fatal("the connive owes a discard")
	}
	if got := me.Hand.Size(); got != 3 {
		t.Fatalf("hand at the discard: %d, want 3 (two seeded and the draw)", got)
	}
	if discardLand {
		sncDiscard(t, g, me, land)
	} else {
		sncDiscard(t, g, me, nonland)
	}
	return id
}

func TestSncB04PlainConniveCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power                  int
	}{
		{"Raffine's Informant", "Creature — Human Wizard", sncInformantOracle, 2},
		{"Psionic Snoop", "Creature — Human Rogue", sncSnoopOracle, 0},
		{"Echo Inspector", "Creature — Bird Rogue", sncInspectorOracle, 2},
		{"Revel Ruiner", "Creature — Octopus Rogue", sncRuinerOracle, 3},
	} {
		for _, land := range []bool{false, true} {
			g := effectstest.NewCatalogGame(t)
			id := sncEnters(t, g, tc.name, tc.typeLine, tc.oracle, land)
			want := 0
			if !land {
				want = 1
			}
			if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != want {
				t.Errorf("%s (discard land %v): %d counters, want %d", tc.name, land, got, want)
			}
			if got := g.Seats[0].Hand.Size(); got != 2 {
				t.Errorf("%s: hand %d after the discard, want 2", tc.name, got)
			}
			if !effectstest.OnBattlefieldByID(g, id) {
				t.Errorf("%s is not on the battlefield", tc.name)
			}
		}
	}
}

func TestSncB04ConniveLeavesTheKeywordsAlone(t *testing.T) {
	for oracle, kws := range map[string][]string{
		sncSnoopOracle: {"flash"}, sncInspectorOracle: {"flying"}, sncRuinerOracle: {"menace"},
		sncInterceptorOracle: {"flash", "lifelink"}, sncVirtuosoOracle: {"double strike"},
	} {
		spec, _ := effects.Lookup(oracle)
		for i, k := range kws {
			if i >= len(spec.PrintedKeywords) || spec.PrintedKeywords[i] != k {
				t.Errorf("%s: keywords %v, want %v", spec.Name, spec.PrintedKeywords, kws)
			}
		}
	}
}

func TestSncB04HypnoticGrifterConnivesForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grifter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hypnotic Grifter", TypeLine: "Creature — Human Rogue",
		OracleID: sncGrifterOracle, Power: 1, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := effectstest.HandCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	effectstest.ToMainForCost(t, g)
	// A pool that cannot pay refuses (strict).
	if err := g.ActivateCatalogAbility(me.ID, grifter, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FloatForTest(g, me, "UUU")
	if err := g.ActivateCatalogAbility(me.ID, grifter, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sncDiscard(t, g, me, nonland)
	if got := effectstest.CounterCount(g, grifter, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand %d, want 1 (drew one, discarded one, one seeded gone)", got)
	}
}

// When it dies the Silencer shrinks an opposing creature by its own power as
// it last was, the +1/+1 counter from its connive included.
func TestSncB04SilencerDeathShrinksByItsLastPower(t *testing.T) {
	for _, tc := range []struct {
		discardLand   bool
		wantBearAlive bool
		wantPower     int
	}{
		{false, false, 0}, // counter: power 2, the 2/2 dies
		{true, true, 1},   // no counter: power 1, the 2/2 is a 1/1
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
		theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		silencer := sncEnters(t, g, "Raffine's Silencer", "Creature — Human Assassin", sncSilencerOracle, tc.discardLand)
		effectstest.KillCreature(t, g, me.ID, silencer)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if effectstest.HasID(p.PickTargetCards, mine) {
				t.Error("its own controller's creature offered as a target")
			}
			effectstest.PickCard(t, g, me.ID, theirs)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.OnBattlefieldByID(g, theirs); got != tc.wantBearAlive {
			t.Fatalf("discard land %v: their bear alive = %v, want %v", tc.discardLand, got, tc.wantBearAlive)
		}
		if tc.wantBearAlive {
			if got := effectstest.CurrentPower(t, g, theirs); got != tc.wantPower {
				t.Errorf("their bear power %d, want %d", got, tc.wantPower)
			}
		}
		if !effectstest.OnBattlefieldByID(g, mine) {
			t.Error("my own creature was hit")
		}
	}
}

// The reflexive "up to one target nonland permanent" is chosen after the
// connive, and the permanent goes back to its owner's hand.
func TestSncB04PickpocketBouncesANonlandPermanentAfterConniving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID,
	})
	pp := sncEnters(t, g, "Psychic Pickpocket", "Creature — Octopus Rogue", sncPickpocketOracle, false)
	if got := effectstest.CounterCount(g, pp, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters %d, want 1", got)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		effectstest.PassPriorityAroundTable(t, g)
		p = effectstest.LatestPickTarget(g, me.ID)
	}
	if p == nil {
		t.Fatal("the reflexive trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, land) {
		t.Error("a land is offered")
	}
	if !effectstest.HasID(p.PickTargetCards, bear) {
		t.Error("their creature is not offered")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, bear) {
		t.Error("their bear stayed on the battlefield")
	}
	if !effectstest.Mh3b1HandNames(opp)["Their Bear"] {
		t.Error("their bear is not in its owner's hand")
	}
}

// With nothing worth bouncing chosen ("up to one"), the permanent stays.
func TestSncB04PickpocketUpToOneMayChooseNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	sncEnters(t, g, "Psychic Pickpocket", "Creature — Octopus Rogue", sncPickpocketOracle, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, g.Seats[0].ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, g.Seats[0].ID, nil); err != nil {
			t.Fatalf("choose no target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, bear) {
		t.Error("their bear was bounced though no target was chosen")
	}
}

// Flash lets the Interceptor come down above a spell; after its connive the
// reflexive trigger targets that spell and returns it to its owner's hand
// (not a counter: it is not put into the graveyard).
func TestSncB04InterceptorReturnsASpellAfterConniving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := effectstest.HandCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	victim := effectstest.CastCatalogSpell(t, g, "Doomed Spell", "Sorcery", "", nil)
	if g.Stack.Size() != 1 {
		t.Fatalf("stack %d, want the one spell", g.Stack.Size())
	}
	interceptor := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: interceptor, Name: "Obscura Interceptor", TypeLine: "Creature — Octopus Wizard",
		OracleID: sncInterceptorOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, interceptor, game.CastSpellParams{Strict: false}); err != nil {
		t.Fatalf("flash it in: %v", err)
	}
	for i := 0; i < 8 && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	sncDiscard(t, g, me, nonland)
	if got := effectstest.CounterCount(g, interceptor, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell left the stack before the reflexive trigger resolved")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the spell is still on the stack")
	}
	if !effectstest.Mh3b1HandNames(me)["Doomed Spell"] {
		t.Error("the spell did not return to its owner's hand")
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Doomed Spell" {
			t.Error("the spell was countered into the graveyard")
		}
	}
}

// "Up to one": with no spell on the stack the trigger has no target and
// does nothing, and the connive still happened.
func TestSncB04InterceptorWithNoSpellStillConnives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := sncEnters(t, g, "Obscura Interceptor", "Creature — Octopus Wizard", sncInterceptorOracle, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
	if !effectstest.OnBattlefieldByID(g, id) {
		t.Error("the Interceptor is not on the battlefield")
	}
}

// sncBypassOn casts Security Bypass from `p`'s hand onto `host` and settles
// the stack, returning the Aura.
func sncBypassOn(t *testing.T, g *game.Game, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.CastCatalogSpell(t, g, "Security Bypass", "Enchantment — Aura", sncBypassOracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	return aura
}

func TestSncB04VirtuosoConnivesWhenYourSpellTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	virtuoso := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Illuminator Virtuoso", TypeLine: "Creature — Human Rogue",
		OracleID: sncVirtuosoOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := effectstest.HandCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	// The Aura is a spell I control that targets it.
	effectstest.CastCatalogSpell(t, g, "Security Bypass", "Enchantment — Aura", sncBypassOracle, effectstest.CardTarget(virtuoso))
	effectstest.PassPriorityAroundTable(t, g)
	sncDiscard(t, g, me, nonland)
	if got := effectstest.CounterCount(g, virtuoso, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
}

// An ability that targets it (equip) is not a spell, and an opponent's spell
// is not "a spell you control": neither connives it.
func TestSncB04VirtuosoIgnoresAbilitiesAndOpposingSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	virtuoso := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Illuminator Virtuoso", TypeLine: "Creature — Human Rogue",
		OracleID: sncVirtuosoOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	mask := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mask of the Schemer", TypeLine: "Artifact — Equipment",
		OracleID: sncMaskOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	handBefore := me.Hand.Size()
	effectstest.FloatForTest(g, me, "UU")
	effectstest.EquipTo(t, g, me.ID, mask, virtuoso)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil || me.Hand.Size() != handBefore {
		t.Error("equip, an ability, made the Virtuoso connive")
	}
}

func TestSncB04SecurityBypassIsUnblockableAloneAndConnivesOnDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bypassed Bear", 2, 2)
	wall := effectstest.BrCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := effectstest.HandCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	sncBypassOn(t, g, bear)

	// Attacking alone: no block, and the damage connives it.
	effectstest.BrAttack(t, g, bear)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, bear), game.BlockReasonCantBeBlocked)
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("opponent life %d -> %d, want 2 damage", life, opp.Life)
	}
	sncDiscard(t, g, me, nonland)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
	// Beside another attacker it can be blocked.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	bear2 := effectstest.PushVanillaCreature(g2, me2.ID, "Bypassed Bear", 2, 2)
	friend2 := effectstest.PushVanillaCreature(g2, me2.ID, "Friend", 2, 2)
	wall2 := effectstest.BrCreature(g2, opp2.ID, "Wall", "Creature — Wall", 0, 8)
	sncBypassOn(t, g2, bear2)
	effectstest.BrAttack(t, g2, bear2, friend2)
	if err := g2.DeclareBlocker(wall2, bear2); err != nil {
		t.Errorf("a creature attacking beside another could not be blocked: %v", err)
	}
}

// The Aura gone, the creature loses both the evasion and the trigger.
func TestSncB04SecurityBypassGrantEndsWithTheAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bypassed Bear", 2, 2)
	wall := effectstest.BrCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	aura := sncBypassOn(t, g, bear)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.KillCreature(t, g, me.ID, aura)
	effectstest.BrAttack(t, g, bear)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Errorf("the creature stayed unblockable after the Aura left: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("the creature connived without the Aura")
	}
}

// Mask of the Schemer: connive X for X combat damage dealt.
func TestSncB04MaskConnivesForTheDamageDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mask := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mask of the Schemer", TypeLine: "Artifact — Equipment",
		OracleID: sncMaskOracle, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	a := effectstest.HandCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	b := effectstest.HandCardFull(me, "Land B", "Basic Land — Forest", "", "", nil)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UU")
	effectstest.EquipTo(t, g, me.ID, mask, bear)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("opponent life %d -> %d, want 2 damage", life, opp.Life)
	}
	// X is 2: two cards drawn, two discarded in one prompt, a counter for the nonland one.
	if got := me.Hand.Size(); got != 4 {
		t.Fatalf("hand at the discard: %d, want 4 (two seeded and two drawn)", got)
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("discard two: %v", err)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1 (one nonland among the two discarded)", got)
	}
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d after discarding, want 2", got)
	}
}
