package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestHeroicInterventionGrantsBothKeywordsToYourSideOnly is the
// sprint's acceptance card: one instant, two keywords, every
// permanent you control and none that you don't.
func TestHeroicInterventionGrantsBothKeywordsToYourSideOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	mine := effectstest.PushVoltronCreature(g, me, "My Bear")
	theirs := effectstest.PushVoltronCreature(g, opp, "Their Bear")
	// A non-creature permanent, because the card says "permanents".
	myRock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Mana Rock",
		TypeLine:   "Artifact",
		Owner:      me.ID,
		Controller: me.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Heroic Intervention", "Instant", effectstest.HeroicInterventionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my creature lacks %q", kw)
		}
		if !effectstest.HasEffectiveKeyword(t, g, myRock, kw) {
			t.Errorf("my artifact lacks %q — the card says permanents, not creatures", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("opponent's creature gained %q", kw)
		}
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("indestructible survived the cleanup step")
	}
}

// TestHeroicInterventionStopsAWrath is the interaction the card is
// played for, end to end.
//
// S30 (#470 / #446): this used to drive a hand-rolled loop over
// DestroyPermanentForEffect and describe it as "the same entry point
// every catalog wrath uses". It was not — a catalog wrath goes
// through DestroyAllMatching → DestroyPermanentsForEffect, which had
// no indestructible guard at all, so the card's headline interaction
// passed its own test while failing in a real game. The spell is now
// cast for real.
func TestHeroicInterventionStopsAWrath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	mine := effectstest.PushVoltronCreature(g, me, "My Bear")
	theirs := effectstest.PushVoltronCreature(g, opp, "Their Bear")

	effectstest.CastCatalogSpell(t, g, "Heroic Intervention", "Instant", effectstest.HeroicInterventionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastCatalogSpell(t, g, "Wrath of God", "Sorcery", effectstest.WrathOfGodOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(mine) {
		t.Error("protected creature died to a wrath")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("unprotected creature survived a wrath — control case broken")
	}
}

// TestBlossomingDefensePumpsAndProtects covers the two-layer shape:
// a 7c boost and a 6 grant from one card.
func TestBlossomingDefensePumpsAndProtects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVoltronCreature(g, me, "My Bear")

	effectstest.CastCatalogSpell(t, g, "Blossoming Defense", "Instant", effectstest.BlossomingDefenseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("target did not gain hexproof")
	}
}

// TestBlossomingDefenseRefusesAnOpponentsCreature pins the "you
// control" clause at announce (CR 601.2c). Hexproof only stops your
// opponents, so the protection half would do nothing on their side
// and the pump half would be a gift.
func TestBlossomingDefenseRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.PushVoltronCreature(g, opp, "Their Bear")

	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Blossoming Defense",
		TypeLine:   "Instant",
		OracleID:   effectstest.BlossomingDefenseOracle,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	})
	if err == nil {
		t.Fatal("casting Blossoming Defense at an opponent's creature was allowed")
	}
}

// TestSnakeskinVeilLeavesThePumpBehind is the difference between
// this card and Ranger's Guile: the counter is permanent, the
// hexproof is not.
func TestSnakeskinVeilLeavesThePumpBehind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVoltronCreature(g, me, "My Bear")

	effectstest.CastCatalogSpell(t, g, "Snakeskin Veil", "Instant", effectstest.SnakeskinVeilOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	// Counters are layer 7d and live on the card, not in the
	// effective characteristic — `CurrentPower` is the read that
	// folds them in, and is what combat and the SBAs use.
	if p := effectstest.CurrentPower(t, g, bear); p != 3 {
		t.Errorf("power after the counter = %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("target did not gain hexproof")
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)

	if effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof survived cleanup")
	}
	if p := effectstest.CurrentPower(t, g, bear); p != 3 {
		t.Errorf("power after cleanup = %d, want 3 — the +1/+1 COUNTER is permanent", p)
	}
}

// TestMakeAStandPumpsPowerOnly guards the asymmetric boost: +1/+0
// means the toughness must not move.
func TestMakeAStandPumpsPowerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVoltronCreature(g, me, "My Bear")

	effectstest.CastCatalogSpell(t, g, "Make a Stand", "Instant", effectstest.MakeAStandOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("creature did not gain indestructible")
	}
}

// TestTamiyosSafekeepingProtectsANonCreatureAndGainsLife — the card
// says PERMANENT, which is what separates it from the rest of the
// family.
func TestTamiyosSafekeepingProtectsANonCreatureAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Mana Rock",
		TypeLine:   "Artifact",
		Owner:      me.ID,
		Controller: me.ID,
	})
	lifeBefore := me.Life

	effectstest.CastCatalogSpell(t, g, "Tamiyo's Safekeeping", "Instant", effectstest.TamiyosSafekeepingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)

	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, rock, kw) {
			t.Errorf("artifact lacks %q", kw)
		}
	}
	if got, want := me.Life, lifeBefore+2; got != want {
		t.Errorf("life = %d, want %d", got, want)
	}
}

// TestAvacynGrantsIndestructibleToOthersNotHerselfTwice covers both
// halves of the card, including the self-exclusion that keeps the
// client's badge row from doubling up.
func TestAvacynGrantsIndestructibleToOthersNotHerselfTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	avacyn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Avacyn, Angel of Hope",
		TypeLine:   "Legendary Creature — Angel",
		OracleID:   effectstest.AvacynOracle,
		Power:      8,
		Toughness:  8,
		Owner:      me.ID,
		Controller: me.ID,
	})
	mine := effectstest.PushVoltronCreature(g, me, "My Bear")
	myLand := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Plains",
		TypeLine:   "Basic Land — Plains",
		Owner:      me.ID,
		Controller: me.ID,
	})
	theirs := effectstest.PushVoltronCreature(g, opp, "Their Bear")

	if !effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("my creature lacks indestructible under Avacyn")
	}
	if !effectstest.HasEffectiveKeyword(t, g, myLand, "indestructible") {
		t.Error("my land lacks indestructible — the line says permanents")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("opponent's creature gained indestructible")
	}
	if !effectstest.HasEffectiveKeyword(t, g, avacyn, "indestructible") {
		t.Error("Avacyn lacks her own printed indestructible")
	}
	// Self-exclusion: exactly one copy of the string.
	n := 0
	for _, a := range effectstest.EffectiveAbilities(t, g, avacyn) {
		if a == "indestructible" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("Avacyn carries %d copies of \"indestructible\", want 1", n)
	}
}

// TestBastionProtectorOnlyBuffsCommanders is the card that makes
// Card.IsCommander load-bearing.
func TestBastionProtectorOnlyBuffsCommanders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Bastion Protector",
		TypeLine:   "Creature — Human Soldier",
		OracleID:   effectstest.BastionProtectorOracle,
		Power:      3,
		Toughness:  3,
		Owner:      me.ID,
		Controller: me.ID,
	})
	myCmdr := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:  uuid.New(),
		Name:        "My Commander",
		TypeLine:    "Legendary Creature — Avatar",
		Power:       4,
		Toughness:   4,
		IsCommander: true,
		Owner:       me.ID,
		Controller:  me.ID,
	})
	myBear := effectstest.PushVoltronCreature(g, me, "My Bear")
	theirCmdr := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:  uuid.New(),
		Name:        "Their Commander",
		TypeLine:    "Legendary Creature — Avatar",
		Power:       4,
		Toughness:   4,
		IsCommander: true,
		Owner:       opp.ID,
		Controller:  opp.ID,
	})

	if p, tough := effectstest.EffectivePower(t, g, myCmdr), effectstest.EffectiveToughness(t, g, myCmdr); p != 6 || tough != 6 {
		t.Errorf("my commander P/T = %d/%d, want 6/6", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, myCmdr, "indestructible") {
		t.Error("my commander lacks indestructible")
	}
	if p := effectstest.EffectivePower(t, g, myBear); p != 2 {
		t.Errorf("my non-commander bear power = %d, want 2", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, myBear, "indestructible") {
		t.Error("a non-commander creature gained indestructible")
	}
	if p := effectstest.EffectivePower(t, g, theirCmdr); p != 4 {
		t.Errorf("opponent's commander power = %d, want 4", p)
	}
}

// TestZurgoIsIndestructibleOnlyOnHisControllersTurn is the
// conditional static — the case a "grant once" implementation of the
// keyword would have got wrong.
//
// #1599: Zurgo's own "attacks each combat if able" is now enforced
// (AttacksEachCombat, attack_requirements.go), so this test declares
// him as an attacker at his controller's declare-attackers step
// before moving on — otherwise the free ride through
// advanceToNextSeatsTurn hits the CR 508.1d refusal that is this
// card's whole point elsewhere (attack_requirement_cards_test.go).
func TestZurgoIsIndestructibleOnlyOnHisControllersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]

	zurgo := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Zurgo Helmsmasher",
		TypeLine:   "Legendary Creature — Orc Warrior",
		OracleID:   effectstest.ZurgoOracle,
		Power:      7,
		Toughness:  2,
		Owner:      me.ID,
		Controller: me.ID,
	})

	if !effectstest.HasEffectiveKeyword(t, g, zurgo, "indestructible") {
		t.Fatal("Zurgo lacks indestructible on his controller's turn")
	}
	if !effectstest.HasEffectiveKeyword(t, g, zurgo, "haste") {
		t.Error("Zurgo lacks his printed haste")
	}

	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	if err := g.DeclareAttacker(zurgo, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Zurgo): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with Zurgo attacking: %v", err)
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)

	if effectstest.HasEffectiveKeyword(t, g, zurgo, "indestructible") {
		t.Error("Zurgo kept indestructible on someone else's turn")
	}
	if !effectstest.HasEffectiveKeyword(t, g, zurgo, "haste") {
		t.Error("Zurgo lost his printed haste along with the conditional grant")
	}
}

// TestUrilCountsAurasAttachedToHim is the half of the card that was
// deferred when this file was first written and became writable an
// hour later, when S24's attachment relation (#374) landed on main.
// Layer 7c, recomputed every pass, so removing an Aura shrinks him
// again with nothing to un-latch.
func TestUrilCountsAurasAttachedToHim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	uril := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Uril, the Miststalker",
		TypeLine:   "Legendary Creature — Beast",
		OracleID:   effectstest.UrilOracle,
		Power:      5,
		Toughness:  5,
		Owner:      me.ID,
		Controller: me.ID,
	})
	aura1 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Test Aura",
		TypeLine:   "Enchantment — Aura",
		Owner:      me.ID,
		Controller: me.ID,
	})
	aura2 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Other Aura",
		TypeLine:   "Enchantment — Aura",
		Owner:      me.ID,
		Controller: me.ID,
	})
	// Equipment rides the same AttachedTo field and must NOT count.
	equip := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Test Blade",
		TypeLine:   "Artifact — Equipment",
		Owner:      me.ID,
		Controller: me.ID,
	})

	if p := effectstest.EffectivePower(t, g, uril); p != 5 {
		t.Fatalf("unattached Uril power = %d, want 5", p)
	}

	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{aura1, aura2, equip} {
			if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: uril}); err != nil {
				t.Fatalf("AttachForEffect: %v", err)
			}
		}
	})

	if p, tough := effectstest.EffectivePower(t, g, uril), effectstest.EffectiveToughness(t, g, uril); p != 9 || tough != 9 {
		t.Errorf("Uril with two Auras and one Equipment = %d/%d, want 9/9", p, tough)
	}

	// Remove one Aura: the count is recomputed, not latched.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura1); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if p := effectstest.EffectivePower(t, g, uril); p != 7 {
		t.Errorf("Uril after losing one Aura = %d, want 7", p)
	}
}

// TestUrilAndSigardaCarryEnforcedHexproof — the two named voltron
// commanders ship with their targeting protection genuinely live,
// which is the half of each card that is not waiting on attachments.
func TestUrilAndSigardaCarryEnforcedHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	uril := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Uril, the Miststalker",
		TypeLine:   "Legendary Creature — Beast",
		OracleID:   effectstest.UrilOracle,
		Power:      5,
		Toughness:  5,
		Owner:      me.ID,
		Controller: me.ID,
	})

	if !effectstest.HasEffectiveKeyword(t, g, uril, "hexproof") {
		t.Error("Uril lacks hexproof")
	}
	// The targeting gate is the consumer that makes it mean
	// something: an opponent may not choose him.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != uril {
				continue
			}
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(opp.ID)) {
				t.Error("an opponent may target Uril despite hexproof")
			}
			if !game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceChooser(me.ID)) {
				t.Error("Uril's own controller may not target him — hexproof is not shroud")
			}
		}
	})
}
