package testscr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every werewolf in the slice turns over with the designation, its two
// faces are the right cards, and each face's printed keywords ride.
func TestDayNightBWerewolvesTurnOverAndKeepTheirKeywords(t *testing.T) {
	for _, w := range effectstest.DnbWerewolves {
		t.Run(w.Front, func(t *testing.T) {
			spec, ok := effects.Lookup(w.Oracle)
			if !ok || spec.Name != w.Front {
				t.Fatalf("front face is not registered as %q: %+v", w.Front, spec)
			}
			back, ok := effects.Lookup(w.Oracle + "#1")
			if !ok || back.Name != w.Back {
				t.Fatalf("back face is not registered as %q: %+v", w.Back, back)
			}
			if !slices.Contains(spec.PrintedKeywords, "daybound") || !slices.Contains(back.PrintedKeywords, "nightbound") {
				t.Errorf("keywords: front %v back %v, want daybound / nightbound", spec.PrintedKeywords, back.PrintedKeywords)
			}
			g, _, _, id := dnaTable(t, w)
			if c, _ := effectstest.BattlefieldCard(g, id); c.Name != w.Front || c.ActiveFace != 0 {
				t.Fatalf("by day: %s face %d, want %s", c.Name, c.ActiveFace, w.Front)
			}
			for _, k := range w.FrontKeywords {
				if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), k) {
					t.Errorf("front lacks %q: %v", k, effectstest.EffectiveAbilities(t, g, id))
				}
			}
			dnaNight(g)
			if c, _ := effectstest.BattlefieldCard(g, id); c.Name != w.Back || c.ActiveFace != 1 {
				t.Fatalf("by night: %s face %d, want %s", c.Name, c.ActiveFace, w.Back)
			}
			for _, k := range w.BackKeys {
				if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), k) {
					t.Errorf("back lacks %q: %v", k, effectstest.EffectiveAbilities(t, g, id))
				}
			}
			dnaDay(g)
			if c, _ := effectstest.BattlefieldCard(g, id); c.Name != w.Front {
				t.Errorf("back by day: %s, want %s", c.Name, w.Front)
			}
		})
	}
}

// The vanilla and keyword werewolves carry exactly their printed
// numbers on each face.
func TestDayNightBTableWerewolvesHaveTheirPrintedPower(t *testing.T) {
	for _, tc := range []struct {
		oracle        string
		front, back   int
		frontT, backT int
	}{
		{effectstest.DnbRuffian, 2, 6, 5, 5},
		{effectstest.DnbHauler, 4, 6, 5, 6},
		{effectstest.DnbShady, 2, 4, 3, 4},
	} {
		w := effectstest.DnbWerewolf(tc.oracle)
		g, _, _, id := dnaTable(t, w)
		if got := effectstest.EffectivePower(t, g, id); got != tc.front {
			t.Errorf("%s power = %d, want %d", w.Front, got, tc.front)
		}
		if got := effectstest.EffectiveToughness(t, g, id); got != tc.frontT {
			t.Errorf("%s toughness = %d, want %d", w.Front, got, tc.frontT)
		}
		dnaNight(g)
		if got := effectstest.EffectivePower(t, g, id); got != tc.back {
			t.Errorf("%s power = %d, want %d", w.Back, got, tc.back)
		}
		if got := effectstest.EffectiveToughness(t, g, id); got != tc.backT {
			t.Errorf("%s toughness = %d, want %d", w.Back, got, tc.backT)
		}
	}
}

// Oakshade Stalker's flash-for-{2}-more is an optional cost that opens
// the window.
func TestOakshadeStalkerOffersItsFlashSurcharge(t *testing.T) {
	// #3172: the flash-for-{2}-more clause landed (FlashForMore);
	// cav_b_test.go casts it. #5207: it is an optional additional cost,
	// not an alternative cost, so the front face is Full.
	spec, ok := effects.Lookup(effectstest.DnbOakshade)
	if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || len(spec.AlternativeCosts) != 0 ||
		len(spec.OptionalCosts) != 1 || !spec.OptionalCosts[0].AsThoughFlash || spec.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("Oakshade Stalker: %+v, want Full with the {2} optional cost that opens the window", spec)
	}
	if back, _ := effects.Lookup(effectstest.DnbOakshade + "#1"); back.Completeness != effects.CompletenessFull {
		t.Errorf("Moonlit Ambusher is %v, want full", back.Completeness)
	}
}

// Outland Liberator sacrifices to destroy an artifact or enchantment.
func TestOutlandLiberatorSacrificesToDestroyAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.B12Push(g, me.ID, "Outland Liberator", "Creature — Human Werewolf", effectstest.DnbLiberator, 2, 2)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, lib, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("the ability accepted a creature that is neither artifact nor enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, lib, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, lib); ok {
		t.Error("the sacrifice was not paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, relic); ok {
		t.Error("the Relic survived")
	}
}

// Frenzied Trapbreaker's attack trigger reaches only what the defending
// player controls.
func TestFrenziedTrapbreakerDestroysADefendersArtifactOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	breaker := effectstest.B12Push(g, me.ID, "Frenzied Trapbreaker", "Creature — Werewolf", effectstest.DnbLiberator+"#1", 3, 3)
	mine := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	other := effectstest.B12Permanent(g, third.ID, "Third Relic", "Enchantment")
	effectstest.DeclareAttack(t, g, opp.ID, breaker)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	for _, bad := range []uuid.UUID{mine, other} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Error("the trigger accepted a permanent the defending player does not control")
		}
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("the defending player's Relic survived")
	}
	for _, id := range []uuid.UUID{mine, other} {
		if _, ok := effectstest.BattlefieldCard(g, id); !ok {
			t.Error("a permanent the defending player does not control was destroyed")
		}
	}
}

// Reckless Stormseeker boosts and hastes a creature at the beginning of
// combat on its controller's turn; the back face adds trample.
func TestRecklessStormseekerBoostsAtBeginningOfCombat(t *testing.T) {
	for _, night := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		oracle, name, power := effectstest.DnbStormseek, "Reckless Stormseeker", 1
		if night {
			oracle, name, power = effectstest.DnbStormseek+"#1", "Storm-Charged Slasher", 2
		}
		effectstest.B12Push(g, me.ID, name, "Creature — Werewolf", oracle, 2, 3)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.AnswerPickTarget(t, g, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.EffectivePower(t, g, bear); got != 2+power {
			t.Errorf("night=%v: Bear power = %d, want %d", night, got, 2+power)
		}
		abil := effectstest.EffectiveAbilities(t, g, bear)
		if !slices.Contains(abil, "haste") {
			t.Errorf("night=%v: Bear lacks haste: %v", night, abil)
		}
		if night != slices.Contains(abil, "trample") {
			t.Errorf("night=%v: trample = %v", night, slices.Contains(abil, "trample"))
		}
	}
}

// Spellrune Painter grows for an instant or sorcery and for nothing else.
func TestSpellruneWerewolfGrowsForInstantsAndSorceries(t *testing.T) {
	for _, night := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		oracle, name, bonus := effectstest.DnbPainter, "Spellrune Painter", 1
		if night {
			oracle, name, bonus = effectstest.DnbPainter+"#1", "Spellrune Howler", 2
		}
		id := effectstest.B12Push(g, me.ID, name, "Creature — Werewolf", oracle, 2, 3)
		effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.EffectivePower(t, g, id); got != 2 {
			t.Fatalf("night=%v: a creature spell pumped it to %d", night, got)
		}
		effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.EffectivePower(t, g, id); got != 2+bonus {
			t.Errorf("night=%v: power = %d, want %d", night, got, 2+bonus)
		}
		if got := effectstest.EffectiveToughness(t, g, id); got != 3+bonus {
			t.Errorf("night=%v: toughness = %d, want %d", night, got, 3+bonus)
		}
	}
}

// Suspicious Stowaway can't be blocked, and Seafaring Werewolf draws.
func TestSuspiciousStowawayCantBeBlockedAndSeafaringWerewolfDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := effectstest.B12Push(g, me.ID, "Seafaring Werewolf", "Creature — Werewolf", effectstest.DnbStowaway+"#1", 2, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	handBefore := len(me.Hand.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, wolf)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, wolf); err == nil {
		t.Fatal("a creature blocked an unblockable werewolf")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d: one card drawn", got, handBefore+1)
	}
}

// The front face loots: draws one, discards one.
func TestSuspiciousStowawayLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	front := effectstest.B12Push(g, me.ID, "Suspicious Stowaway", "Creature — Human Rogue Werewolf", effectstest.DnbStowaway, 1, 1)
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Drawn Card", TypeLine: "Sorcery", Owner: me.ID})
	handBefore := len(me.Hand.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, front)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) == handBefore+1 {
		effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("hand = %d, want %d after draw then discard", got, handBefore)
	}
}

// Weaver of Blossoms taps for one colour; Blossom-Clad Werewolf for two
// of ONE colour.
func TestWeaverOfBlossomsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	weaver := effectstest.B12Push(g, me.ID, "Weaver of Blossoms", "Creature — Human Werewolf", effectstest.DnbWeaver, 2, 3)
	clad := effectstest.B12Push(g, me.ID, "Blossom-Clad Werewolf", "Creature — Werewolf", effectstest.DnbWeaver+"#1", 3, 4)
	if err := g.ActivateManaAbility(me.ID, weaver, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Weaver: %v", err)
	}
	if effectstest.B10ResolveAllManaPicks(t, g, me.ID, "G") != 1 {
		t.Fatal("Weaver raised no colour pick")
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Fatalf("pool after Weaver = %d, want 1", got)
	}
	if err := g.ActivateManaAbility(me.ID, clad, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Blossom-Clad: %v", err)
	}
	if effectstest.B10ResolveAllManaPicks(t, g, me.ID, "R") != 1 {
		t.Fatal("Blossom-Clad Werewolf raised no colour pick")
	}
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool after both = %d, want 3", got)
	}
	reds := 0
	for _, m := range me.ManaPool.Units() {
		if m.Color == "R" {
			reds++
		}
	}
	if reds != 2 {
		t.Errorf("red mana = %d, want both of Blossom-Clad Werewolf's to be red", reds)
	}
}

// Sunstreak Phoenix comes back from the graveyard tapped when the
// designation flips and its controller pays {1}{R}.
func TestSunstreakPhoenixReturnsOnAFlipForTheFee(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	phoenix := dnCast(t, g, "Sunstreak Phoenix", "Creature — Phoenix", effectstest.DnbPhoenix)
	if !g.IsDay() {
		t.Fatal("entering did not make it day")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(phoenix); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if _, ok := effectstest.BattlefieldCard(g, phoenix); ok {
		t.Fatal("the Phoenix did not die")
	}
	dnaNight(g)
	effectstest.MonSettle(t, g)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	c, ok := effectstest.BattlefieldCard(g, phoenix)
	if !ok {
		t.Fatal("the Phoenix did not return")
	}
	if !c.Tapped {
		t.Error("it returned untapped")
	}
}

func TestSunstreakPhoenixStaysIfYouDeclineTheFee(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	phoenix := dnCast(t, g, "Sunstreak Phoenix", "Creature — Phoenix", effectstest.DnbPhoenix)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(phoenix); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	dnaNight(g)
	effectstest.MonSettle(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.MonSettle(t, g)
	if _, ok := effectstest.BattlefieldCard(g, phoenix); ok {
		t.Error("it returned without being paid for")
	}
}

// A Phoenix that is still on the battlefield offers nothing on a flip.
func TestSunstreakPhoenixOnTheBattlefieldIgnoresAFlip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dnCast(t, g, "Sunstreak Phoenix", "Creature — Phoenix", effectstest.DnbPhoenix)
	dnaNight(g)
	effectstest.MonSettle(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a payment was offered for a Phoenix that is not in the graveyard")
	}
}

// Vadrik makes your instants and sorceries cheaper by its power, and
// grows on each flip.
func TestVadrikDiscountsByPowerAndGrowsOnAFlip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vadrik := effectstest.B12Push(g, me.ID, "Vadrik, Astral Archmage", "Legendary Creature — Human Wizard", effectstest.DnbVadrik, 1, 2)
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationDay })
	if got := effectstest.DnaPower(t, g, vadrik); got != 1 {
		t.Fatalf("Vadrik power = %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("an instant costs %d, want 1 with Vadrik at power 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("a creature spell costs %d, want 2: Vadrik discounts instants and sorceries", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Their Bolt", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("an opponent's instant costs %d, want 2", got)
	}
	dnaNight(g)
	effectstest.MonSettle(t, g)
	if got := effectstest.DnaPower(t, g, vadrik); got != 2 {
		t.Fatalf("Vadrik power = %d after a flip, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Big Sorcery", "Sorcery", "{3}{U}"); got != 2 {
		t.Errorf("a sorcery costs %d, want 2 with Vadrik at power 2", got)
	}
}

// Wolf Strike: the biter deals its power; at night it first gets +2/+0.
func TestWolfStrikeBitesAndPumpsAtNight(t *testing.T) {
	for _, night := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		biter := effectstest.B12Creature(g, me.ID, "Biter", "Creature — Wolf", 2, 2)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 0, 6)
		if night {
			g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
		} else {
			g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationDay })
		}
		effectstest.CastCatalogSpell(t, g, "Wolf Strike", "Instant", effectstest.DnbWolfStrike, []game.TargetRef{
			{Kind: game.TargetCard, ID: biter}, {Kind: game.TargetCard, ID: victim},
		})
		effectstest.PassPriorityAroundTable(t, g)
		wantPower, wantDamage := 2, 2
		if night {
			wantPower, wantDamage = 4, 4
		}
		if got := effectstest.EffectivePower(t, g, biter); got != wantPower {
			t.Errorf("night=%v: Biter power = %d, want %d", night, got, wantPower)
		}
		c, ok := effectstest.BattlefieldCard(g, victim)
		if !ok || c.DamageMarked != wantDamage {
			t.Errorf("night=%v: Victim damage = %d (present %v), want %d", night, c.DamageMarked, ok, wantDamage)
		}
		if c, _ := effectstest.BattlefieldCard(g, biter); c.DamageMarked != 0 {
			t.Errorf("night=%v: a one-sided bite damaged the biter (%d)", night, c.DamageMarked)
		}
	}
}

// Wolf Strike cannot name your own creature as the victim.
func TestWolfStrikeRefusesAnOwnVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Wolf", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Wolf Strike", "Instant", effectstest.DnbWolfStrike, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	if err == nil {
		t.Error("Wolf Strike accepted a creature you control as the second target")
	}
}

// Wolfkin Outcast costs {2} less with a Wolf or Werewolf, and only then.
func TestWolfkinOutcastCostsLessWithAWolfOrWerewolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.DnbOutcast, "Creature — Human Werewolf", "{5}{G}"); got != 6 {
		t.Fatalf("with no Wolf the Outcast costs %d, want 6", got)
	}
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.DnbOutcast, "Creature — Human Werewolf", "{5}{G}"); got != 6 {
		t.Errorf("with a Bear the Outcast costs %d, want 6", got)
	}
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Wolf", "Creature — Wolf", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.DnbOutcast, "Creature — Human Werewolf", "{5}{G}"); got != 6 {
		t.Errorf("with only an opponent's Wolf the Outcast costs %d, want 6", got)
	}
	effectstest.B12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.DnbOutcast, "Creature — Human Werewolf", "{5}{G}"); got != 4 {
		t.Errorf("with a Wolf the Outcast costs %d, want 4", got)
	}
}

// Wedding Crasher draws when it or another Wolf or Werewolf you control
// dies, and for nothing else.
func TestWeddingCrasherDrawsForWolvesAndWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crasher := effectstest.B12Push(g, me.ID, "Wedding Crasher", "Creature — Werewolf", effectstest.DnbOutcast+"#1", 6, 5)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	kill := func(id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	hand := len(me.Hand.Cards)
	kill(bear)
	kill(theirs)
	if got := len(me.Hand.Cards); got != hand {
		t.Fatalf("a Bear or an opponent's Wolf drew %d cards", got-hand)
	}
	kill(wolf)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("your Wolf dying drew %d cards, want 1", got-hand)
	}
	kill(crasher)
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("Wedding Crasher dying drew %d more cards, want 1", got-hand-1)
	}
}

// Unnatural Moonrise: night, +1/+0, trample and a draw on combat damage.
func TestUnnaturalMoonriseMakesNightAndGrantsTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Unnatural Moonrise", "Sorcery", effectstest.DnbMoonrise, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsNight() {
		t.Fatalf("designation = %q, want night", effectstest.DnDesignation(g))
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power = %d, want 3", got)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, bear), "trample") {
		t.Error("the Bear did not gain trample")
	}
	hand := len(me.Hand.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand = %d, want %d: the granted trigger draws", got, hand+1)
	}
}

// The spell's only target is the creature, so when it is gone the spell
// does not resolve at all (CR 608.2b) and nothing becomes night.
func TestUnnaturalMoonriseFizzlesWhenTheTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Unnatural Moonrise", "Sorcery", effectstest.DnbMoonrise, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.IsNight() {
		t.Error("the spell resolved with no legal target")
	}
}

func TestUnnaturalMoonriseHasFlashback(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.DnbMoonrise)
	if !ok || !slices.Contains(spec.CastableZones, game.ZoneGraveyard) || len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "flashback" {
		t.Fatalf("Unnatural Moonrise: %+v, want flashback from the graveyard", spec)
	}
}

// Tovolar's Huntmaster makes two Wolves; cast at night, Packleader's
// enters trigger makes two as well.
func TestTovolarsHuntmasterMakesTwoWolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ImportAndCast(t, g, effectstest.DnbWerewolf(effectstest.DnbHuntmaster).Row(), me)
	effectstest.MonSettle(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Wolf"); n != 2 {
		t.Errorf("Wolf tokens by day = %d, want 2", n)
	}
	g2 := effectstest.NewCatalogGame(t)
	g2.WithWriteLock(func() { g2.DayNight.Designation = game.DesignationNight })
	me2 := g2.Seats[0]
	id := effectstest.ImportAndCast(t, g2, effectstest.DnbWerewolf(effectstest.DnbHuntmaster).Row(), me2)
	effectstest.MonSettle(t, g2)
	if c, _ := effectstest.BattlefieldCard(g2, id); c.Name != "Tovolar's Packleader" {
		t.Fatalf("cast at night: %s, want Tovolar's Packleader", c.Name)
	}
	if n, _ := effectstest.CountTokensNamed(g2, me2.ID, "Wolf"); n != 2 {
		t.Errorf("Wolf tokens at night = %d, want 2 (the back face's enters trigger)", n)
	}
}

// Tovolar's Packleader's fight: another Wolf or Werewolf of yours fights
// a creature you don't control.
func TestTovolarsPackleaderFightsWithAnotherWolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	leader := effectstest.B12Push(g, me.ID, "Tovolar's Packleader", "Creature — Werewolf", effectstest.DnbHuntmaster+"#1", 7, 7)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 4)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}")
	// Not itself, and not a non-Wolf.
	for _, bad := range []uuid.UUID{leader, bear} {
		if err := g.ActivateCatalogAbility(me.ID, leader, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}, {Kind: game.TargetCard, ID: victim}},
		}); err == nil {
			t.Errorf("the fight accepted %s as the fighter", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, leader, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wolf}, {Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, wolf); !ok || c.DamageMarked != 2 {
		t.Errorf("Wolf damage = %d (present %v), want 2", c.DamageMarked, ok)
	}
	if c, ok := effectstest.BattlefieldCard(g, victim); !ok || c.DamageMarked != 3 {
		t.Errorf("Victim damage = %d (present %v), want 3", c.DamageMarked, ok)
	}
}

// Tovolar draws once for each Wolf or Werewolf that connects, not for a
// Bear.
func TestTovolarDrawsForEachWolfOrWerewolfThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Tovolar, Dire Overlord", "Legendary Creature — Human Werewolf", effectstest.DnbTovolar, 3, 3)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	were := effectstest.B12Creature(g, me.ID, "Werewolf", "Creature — Werewolf", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := len(me.Hand.Cards)
	effectstest.DeclareAttack(t, g, opp.ID, wolf, were, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("cards drawn = %d, want 2 (Wolf and Werewolf, not the Bear)", got-hand)
	}
}

// dnbAfterASpellTurn plays through to the last seat's turn and has it
// cast one spell, so the CR 502.2 check at the next untap step changes
// nothing and any change to the designation at seat 0's upkeep is the
// card's own.
func dnbAfterASpellTurn(t *testing.T, g *game.Game) {
	t.Helper()
	last := len(g.Seats) - 1
	effectstest.AdvanceToMainOf(t, g, last)
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The turns on the way here cast nothing and ran the check, so it is
	// night and the werewolves are turned over; make it day again the way
	// the game does, so they turn back, for the turn under test.
	g.WithWriteLock(func() {
		g.DayNight.Designation = game.DesignationNight
		g.BecomeDayForEffect()
	})
	if c := g.CastTallyFor(g.Seats[last].ID); c.Total != 1 {
		t.Fatalf("the last seat has cast %d spells, want 1", c.Total)
	}
}

// At your upkeep with three Wolves and/or Werewolves it becomes night,
// and you may transform Human Werewolves that are not daybound.
func TestTovolarsUpkeepMakesNight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tov := effectstest.ImportAndCast(t, g, effectstest.DnbWerewolf(effectstest.DnbTovolar).Row(), me)
	effectstest.MonSettle(t, g)
	effectstest.B12Creature(g, me.ID, "Wolf A", "Creature — Wolf", 2, 2)
	effectstest.B12Creature(g, me.ID, "Wolf B", "Creature — Wolf", 2, 2)
	dnbAfterASpellTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAllPrompts(t, g, me.ID)
	if !g.IsNight() {
		t.Fatalf("designation = %q, want night", effectstest.DnDesignation(g))
	}
	if c, _ := effectstest.BattlefieldCard(g, tov); c.ActiveFace != 1 {
		t.Errorf("Tovolar is on face %d, want its night face", c.ActiveFace)
	}
}

// Two Wolves are not enough.
func TestTovolarsUpkeepNeedsThreeWolvesOrWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ImportAndCast(t, g, effectstest.DnbWerewolf(effectstest.DnbTovolar).Row(), me)
	effectstest.MonSettle(t, g)
	effectstest.B12Creature(g, me.ID, "Wolf A", "Creature — Wolf", 2, 2)
	dnbAfterASpellTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsDay() {
		t.Errorf("designation = %q, want day: only two Wolves/Werewolves", effectstest.DnDesignation(g))
	}
}

// The back face's X pump: +X/+0 and trample to a Wolf or Werewolf.
func TestTovolarTheMidnightScourgePumpsAWolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tov := effectstest.B12Push(g, me.ID, "Tovolar, the Midnight Scourge", "Legendary Creature — Werewolf", effectstest.DnbTovolar+"#1", 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}{R}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, tov, 0, game.ActivateAbilityParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("the pump accepted a Bear")
	}
	effectstest.FloatMana(t, g, me, "{R}{R}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, tov, 0, game.ActivateAbilityParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tov}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, tov); got != 6 {
		t.Errorf("power = %d, want 6", got)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, tov), "trample") {
		t.Error("no trample")
	}
}

// Volatile Arsonist deals its damage to each chosen kind of target.
func TestVolatileArsonistDamagesACreatureAndAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arsonist := effectstest.B12Push(g, me.ID, "Dire-Strain Anarchist", "Creature — Werewolf", effectstest.DnbArsonist+"#1", 5, 5)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 4)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, arsonist)
	// One prompt per clause: the creature, the player, no planeswalker.
	for i, answer := range [][]game.TargetRef{
		{{Kind: game.TargetCard, ID: victim}},
		{{Kind: game.TargetPlayer, ID: opp.ID}},
		nil,
	} {
		pick := effectstest.LatestPickTarget(g, me.ID)
		if pick == nil && answer == nil {
			continue // no planeswalker on the table: no prompt for that clause
		}
		if pick == nil {
			t.Fatalf("no target prompt for clause %d of the attack trigger", i)
		}
		if err := g.ResolvePickTargets(pick.ID, me.ID, answer); err != nil {
			t.Fatalf("clause %d: ResolvePickTargets: %v", i, err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, victim); !ok || c.DamageMarked != 2 {
		t.Errorf("Victim damage = %d (present %v), want 2", c.DamageMarked, ok)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-2)
	}
}

// With nothing chosen the trigger does nothing and does not error.
func TestVolatileArsonistMayTargetNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arsonist := effectstest.B12Push(g, me.ID, "Volatile Arsonist", "Creature — Human Werewolf", effectstest.DnbArsonist, 4, 4)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, arsonist)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("declining every target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life = %d, want %d", opp.Life, life)
	}
}

// Wrathful Jailbreaker has to attack; Weary Prisoner is a defender.
func TestWrathfulJailbreakerAttacksEachCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jail := effectstest.B12Push(g, me.ID, "Wrathful Jailbreaker", "Creature — Werewolf", effectstest.DnbPrisoner+"#1", 6, 6)
	var reqs []game.AttackRequirement
	g.ReadSnapshot(func() {
		if c, ok := effectstest.BattlefieldCard(g, jail); ok {
			reqs = c.Effective().AttackRequirements
		}
	})
	if len(reqs) == 0 {
		t.Error("Wrathful Jailbreaker carries no attack requirement")
	}
	prisoner := effectstest.B12Push(g, me.ID, "Weary Prisoner", "Creature — Human Werewolf", effectstest.DnbPrisoner, 2, 6)
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, prisoner), "defender") {
		t.Error("Weary Prisoner is not a defender")
	}
}
