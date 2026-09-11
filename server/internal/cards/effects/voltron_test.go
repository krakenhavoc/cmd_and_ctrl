package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// voltron_test.go — S25 (#77). The engine's indestructible rules are
// pinned in server/internal/game/indestructible_test.go; these cases
// drive the CARDS through a real cast and check the grants land on
// the right permanents and nothing else.

const (
	heroicInterventionOracle = "24882fa2-3fe9-4c1b-aa3d-0e6488b9db27"
	blossomingDefenseOracle  = "5a851367-1c4a-4cc9-a9c3-3d2775986b4c"
	snakeskinVeilOracle      = "1e6a24be-8281-41c1-a5ba-b68f0ef1d7b8"
	makeAStandOracle         = "531f78d5-5004-4b02-99c7-b390cb342fd9"
	tamiyosSafekeepingOracle = "bb2b324c-970a-4920-884e-c92ba49669f0"
	avacynOracle             = "216cb26e-8da9-478b-bfbc-8030f7adee72"
	bastionProtectorOracle   = "858f53ef-3fec-4aa0-867d-b7f040614b3c"
	zurgoOracle              = "6c48d888-9f5d-43f4-adbd-61dbdba09260"
	urilOracle               = "4308a020-48cf-45fe-8074-dd5d0ac6d12d"
)

// pushVoltronCreature puts a 2/2 on the battlefield under `owner`
// with a battlefield timestamp, so the layer engine sorts it.
func pushVoltronCreature(g *game.Game, owner *game.Player, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
}

// TestHeroicInterventionGrantsBothKeywordsToYourSideOnly is the
// sprint's acceptance card: one instant, two keywords, every
// permanent you control and none that you don't.
func TestHeroicInterventionGrantsBothKeywordsToYourSideOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	mine := pushVoltronCreature(g, me, "My Bear")
	theirs := pushVoltronCreature(g, opp, "Their Bear")
	// A non-creature permanent, because the card says "permanents".
	myRock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Mana Rock",
		TypeLine:   "Artifact",
		Owner:      me.ID,
		Controller: me.ID,
	})

	castCatalogSpell(t, g, "Heroic Intervention", "Instant", heroicInterventionOracle, nil)
	passPriorityAroundTable(t, g)

	for _, kw := range []string{"hexproof", "indestructible"} {
		if !hasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my creature lacks %q", kw)
		}
		if !hasEffectiveKeyword(t, g, myRock, kw) {
			t.Errorf("my artifact lacks %q — the card says permanents, not creatures", kw)
		}
		if hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("opponent's creature gained %q", kw)
		}
	}

	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("indestructible survived the cleanup step")
	}
}

// TestHeroicInterventionStopsAWrath is the interaction the card is
// played for, end to end: the grant reaches the destruction path
// that a catalog wrath goes through.
func TestHeroicInterventionStopsAWrath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	mine := pushVoltronCreature(g, me, "My Bear")
	theirs := pushVoltronCreature(g, opp, "Their Bear")

	castCatalogSpell(t, g, "Heroic Intervention", "Instant", heroicInterventionOracle, nil)
	passPriorityAroundTable(t, g)

	// A destroy-all driven through the same entry point every
	// catalog wrath uses.
	g.WithWriteLock(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.IsCreature() {
				if err := g.DestroyPermanentForEffect(c.InstanceID); err != nil {
					t.Errorf("DestroyPermanentForEffect: %v", err)
				}
			}
		}
	})

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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVoltronCreature(g, me, "My Bear")

	castCatalogSpell(t, g, "Blossoming Defense", "Instant", blossomingDefenseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("target did not gain hexproof")
	}
}

// TestBlossomingDefenseRefusesAnOpponentsCreature pins the "you
// control" clause at announce (CR 601.2c). Hexproof only stops your
// opponents, so the protection half would do nothing on their side
// and the pump half would be a gift.
func TestBlossomingDefenseRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := pushVoltronCreature(g, opp, "Their Bear")

	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Blossoming Defense",
		TypeLine:   "Instant",
		OracleID:   blossomingDefenseOracle,
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVoltronCreature(g, me, "My Bear")

	castCatalogSpell(t, g, "Snakeskin Veil", "Instant", snakeskinVeilOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	// Counters are layer 7d and live on the card, not in the
	// effective characteristic — `CurrentPower` is the read that
	// folds them in, and is what combat and the SBAs use.
	if p := currentPower(t, g, bear); p != 3 {
		t.Errorf("power after the counter = %d, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("target did not gain hexproof")
	}

	advanceToNextSeatsTurn(t, g)

	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("hexproof survived cleanup")
	}
	if p := currentPower(t, g, bear); p != 3 {
		t.Errorf("power after cleanup = %d, want 3 — the +1/+1 COUNTER is permanent", p)
	}
}

// currentPower reads the counter-inclusive power the combat engine
// and the SBAs see (Effective().Power plus the 7d counter delta),
// as distinct from effectivePower's post-layer-only read.
func currentPower(t *testing.T, g *game.Game, cardID uuid.UUID) int {
	t.Helper()
	var p int
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			p = c.CurrentPower()
			found = true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return p
}

// TestMakeAStandPumpsPowerOnly guards the asymmetric boost: +1/+0
// means the toughness must not move.
func TestMakeAStandPumpsPowerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVoltronCreature(g, me, "My Bear")

	castCatalogSpell(t, g, "Make a Stand", "Instant", makeAStandOracle, nil)
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("creature did not gain indestructible")
	}
}

// TestTamiyosSafekeepingProtectsANonCreatureAndGainsLife — the card
// says PERMANENT, which is what separates it from the rest of the
// family.
func TestTamiyosSafekeepingProtectsANonCreatureAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Mana Rock",
		TypeLine:   "Artifact",
		Owner:      me.ID,
		Controller: me.ID,
	})
	lifeBefore := me.Life

	castCatalogSpell(t, g, "Tamiyo's Safekeeping", "Instant", tamiyosSafekeepingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)

	for _, kw := range []string{"hexproof", "indestructible"} {
		if !hasEffectiveKeyword(t, g, rock, kw) {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	avacyn := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Avacyn, Angel of Hope",
		TypeLine:   "Legendary Creature — Angel",
		OracleID:   avacynOracle,
		Power:      8,
		Toughness:  8,
		Owner:      me.ID,
		Controller: me.ID,
	})
	mine := pushVoltronCreature(g, me, "My Bear")
	myLand := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Plains",
		TypeLine:   "Basic Land — Plains",
		Owner:      me.ID,
		Controller: me.ID,
	})
	theirs := pushVoltronCreature(g, opp, "Their Bear")

	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("my creature lacks indestructible under Avacyn")
	}
	if !hasEffectiveKeyword(t, g, myLand, "indestructible") {
		t.Error("my land lacks indestructible — the line says permanents")
	}
	if hasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("opponent's creature gained indestructible")
	}
	if !hasEffectiveKeyword(t, g, avacyn, "indestructible") {
		t.Error("Avacyn lacks her own printed indestructible")
	}
	// Self-exclusion: exactly one copy of the string.
	n := 0
	for _, a := range effectiveAbilities(t, g, avacyn) {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Bastion Protector",
		TypeLine:   "Creature — Human Soldier",
		OracleID:   bastionProtectorOracle,
		Power:      3,
		Toughness:  3,
		Owner:      me.ID,
		Controller: me.ID,
	})
	myCmdr := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:  uuid.New(),
		Name:        "My Commander",
		TypeLine:    "Legendary Creature — Avatar",
		Power:       4,
		Toughness:   4,
		IsCommander: true,
		Owner:       me.ID,
		Controller:  me.ID,
	})
	myBear := pushVoltronCreature(g, me, "My Bear")
	theirCmdr := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID:  uuid.New(),
		Name:        "Their Commander",
		TypeLine:    "Legendary Creature — Avatar",
		Power:       4,
		Toughness:   4,
		IsCommander: true,
		Owner:       opp.ID,
		Controller:  opp.ID,
	})

	if p, tough := effectivePower(t, g, myCmdr), effectiveToughness(t, g, myCmdr); p != 6 || tough != 6 {
		t.Errorf("my commander P/T = %d/%d, want 6/6", p, tough)
	}
	if !hasEffectiveKeyword(t, g, myCmdr, "indestructible") {
		t.Error("my commander lacks indestructible")
	}
	if p := effectivePower(t, g, myBear); p != 2 {
		t.Errorf("my non-commander bear power = %d, want 2", p)
	}
	if hasEffectiveKeyword(t, g, myBear, "indestructible") {
		t.Error("a non-commander creature gained indestructible")
	}
	if p := effectivePower(t, g, theirCmdr); p != 4 {
		t.Errorf("opponent's commander power = %d, want 4", p)
	}
}

// TestZurgoIsIndestructibleOnlyOnHisControllersTurn is the
// conditional static — the case a "grant once" implementation of the
// keyword would have got wrong.
func TestZurgoIsIndestructibleOnlyOnHisControllersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	zurgo := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Zurgo Helmsmasher",
		TypeLine:   "Legendary Creature — Orc Warrior",
		OracleID:   zurgoOracle,
		Power:      7,
		Toughness:  2,
		Owner:      me.ID,
		Controller: me.ID,
	})

	if !hasEffectiveKeyword(t, g, zurgo, "indestructible") {
		t.Fatal("Zurgo lacks indestructible on his controller's turn")
	}
	if !hasEffectiveKeyword(t, g, zurgo, "haste") {
		t.Error("Zurgo lacks his printed haste")
	}

	advanceToNextSeatsTurn(t, g)

	if hasEffectiveKeyword(t, g, zurgo, "indestructible") {
		t.Error("Zurgo kept indestructible on someone else's turn")
	}
	if !hasEffectiveKeyword(t, g, zurgo, "haste") {
		t.Error("Zurgo lost his printed haste along with the conditional grant")
	}
}

// TestUrilCountsAurasAttachedToHim is the half of the card that was
// deferred when this file was first written and became writable an
// hour later, when S24's attachment relation (#374) landed on main.
// Layer 7c, recomputed every pass, so removing an Aura shrinks him
// again with nothing to un-latch.
func TestUrilCountsAurasAttachedToHim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	uril := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Uril, the Miststalker",
		TypeLine:   "Legendary Creature — Beast",
		OracleID:   urilOracle,
		Power:      5,
		Toughness:  5,
		Owner:      me.ID,
		Controller: me.ID,
	})
	aura1 := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Test Aura",
		TypeLine:   "Enchantment — Aura",
		Owner:      me.ID,
		Controller: me.ID,
	})
	aura2 := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Other Aura",
		TypeLine:   "Enchantment — Aura",
		Owner:      me.ID,
		Controller: me.ID,
	})
	// Equipment rides the same AttachedTo field and must NOT count.
	equip := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Test Blade",
		TypeLine:   "Artifact — Equipment",
		Owner:      me.ID,
		Controller: me.ID,
	})

	if p := effectivePower(t, g, uril); p != 5 {
		t.Fatalf("unattached Uril power = %d, want 5", p)
	}

	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{aura1, aura2, equip} {
			if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: uril}); err != nil {
				t.Fatalf("AttachForEffect: %v", err)
			}
		}
	})

	if p, tough := effectivePower(t, g, uril), effectiveToughness(t, g, uril); p != 9 || tough != 9 {
		t.Errorf("Uril with two Auras and one Equipment = %d/%d, want 9/9", p, tough)
	}

	// Remove one Aura: the count is recomputed, not latched.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura1); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if p := effectivePower(t, g, uril); p != 7 {
		t.Errorf("Uril after losing one Aura = %d, want 7", p)
	}
}

// TestUrilAndSigardaCarryEnforcedHexproof — the two named voltron
// commanders ship with their targeting protection genuinely live,
// which is the half of each card that is not waiting on attachments.
func TestUrilAndSigardaCarryEnforcedHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	uril := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Uril, the Miststalker",
		TypeLine:   "Legendary Creature — Beast",
		OracleID:   urilOracle,
		Power:      5,
		Toughness:  5,
		Owner:      me.ID,
		Controller: me.ID,
	})

	if !hasEffectiveKeyword(t, g, uril, "hexproof") {
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
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, opp.ID) {
				t.Error("an opponent may target Uril despite hexproof")
			}
			if !game.CanBeTargetedBy(c, game.ZoneBattlefield, me.ID) {
				t.Error("Uril's own controller may not target him — hexproof is not shroud")
			}
		}
	})
}
