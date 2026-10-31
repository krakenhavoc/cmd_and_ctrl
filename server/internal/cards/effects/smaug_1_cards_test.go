package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// smaug_1_cards_test.go — the Smaug Unchained deck request (#4724), batch 1:
// Flame Slash, Shivan Meteor, Mox Jasper, Dragonspeaker Shaman, Cut Propulsion
// and Self-Destruct. Each test plays the real catalog entry and asserts what
// its text says.

const (
	smg1FlameSlashOracle    = "8d98d674-6811-4d45-b22a-63792e272a2b"
	smg1MeteorOracle        = "6cf5edd6-1ca2-46a5-83b4-7ddff590ef96"
	smg1MoxJasperOracle     = "d5fca77c-377f-41b4-a9e9-e91a89d351cc"
	smg1ShamanOracle        = "6a29f36b-7911-4d84-a3ae-546b131d5159"
	smg1CutPropulsionOracle = "cf3234f2-09d7-4d8b-b871-3bf0fcb75b6d"
	smg1SelfDestructOracle  = "98bf6311-bc9b-4bf8-ac83-c041973b86d0"
)

func TestSmg1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		smg1FlameSlashOracle:    "Flame Slash",
		smg1MeteorOracle:        "Shivan Meteor",
		smg1MoxJasperOracle:     "Mox Jasper",
		smg1ShamanOracle:        "Dragonspeaker Shaman",
		smg1CutPropulsionOracle: "Cut Propulsion",
		smg1SelfDestructOracle:  "Self-Destruct",
	} {
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

// smg1Creature seeds a creature with extra keywords for a test.
func smg1Creature(g *game.Game, owner uuid.UUID, typeLine string, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Creature", TypeLine: typeLine,
		Keywords: keywords, Power: power, Toughness: toughness,
		Owner: owner, Controller: owner,
	})
}

// smg1Pump puts n +1/+1 counters on a creature, standing in for a pump
// spell cast in response.
func smg1Pump(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddCounterForEffect(id, "+1/+1", n) })
	if err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
}

// smg1LoseKeyword strips a printed keyword from a battlefield creature and
// recomputes the layers, standing in for "loses flying" in response.
func smg1LoseKeyword(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Keywords = nil
			}
		}
		g.BumpLayerVersionForTest()
		g.RecomputeLayersIfStaleLocked()
	})
}

// --- Flame Slash -------------------------------------------------------------

func TestSmg1FlameSlashDealsFourToACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	four := b12Creature(g, opp.ID, "Four Toughness", "Creature — Wall", 0, 4)
	five := b12Creature(g, opp.ID, "Five Toughness", "Creature — Wall", 0, 5)

	castCatalogSpell(t, g, "Flame Slash", "Sorcery", smg1FlameSlashOracle, b16TargetCard(four))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(four) {
		t.Error("a creature with toughness 4 survived 4 damage")
	}

	castCatalogSpell(t, g, "Flame Slash", "Sorcery", smg1FlameSlashOracle, b16TargetCard(five))
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, five); !ok || c.DamageMarked != 4 {
		t.Errorf("toughness 5 creature: on battlefield %v, damage %d, want 4", ok, c.DamageMarked)
	}
}

func TestSmg1FlameSlashHitsYourOwnCreatureToo(t *testing.T) {
	g := newCatalogGame(t)
	mine := b12Creature(g, g.Seats[0].ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Flame Slash", "Sorcery", smg1FlameSlashOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the Bear survived: any creature is a legal target")
	}
}

// "Target creature": not a player, not a planeswalker.
func TestSmg1FlameSlashCannotTargetAPlayerOrAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	walker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 5)
	life := opp.Life

	if err := castCatalogSpellErr(t, g, "Flame Slash", "Sorcery", smg1FlameSlashOracle, b16TargetPlayer(opp.ID)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a player target: err %v, want ErrIllegalTarget", err)
	}
	if err := castCatalogSpellErr(t, g, "Flame Slash", "Sorcery", smg1FlameSlashOracle, b16TargetCard(walker)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a planeswalker target: err %v, want ErrIllegalTarget", err)
	}
	if opp.Life != life || loyaltyOf(g, walker) != 5 {
		t.Errorf("a refused announcement dealt damage: life %d→%d, loyalty %d", life, opp.Life, loyaltyOf(g, walker))
	}
}

// The target is re-checked as the spell resolves (CR 608.2b): a creature that
// left in response takes nothing, and the spell does not blow up.
func TestSmg1FlameSlashFizzlesWhenTheCreatureLeaves(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	spell := castCatalogSpell(t, g, "Flame Slash", "Sorcery", smg1FlameSlashOracle, b16TargetCard(bear))
	removeInResponse(t, g, bear, true)
	passPriorityAroundTable(t, g)
	if !g.Seats[0].Graveyard.Contains(spell) {
		t.Error("the fizzled Flame Slash is not in its owner's graveyard")
	}
	if c, _ := battlefieldCard(g, other); c.DamageMarked != 0 {
		t.Errorf("an unrelated creature took %d", c.DamageMarked)
	}
}

// --- Shivan Meteor -----------------------------------------------------------

func TestSmg1ShivanMeteorDealsThirteen(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	thirteen := b12Creature(g, opp.ID, "Thirteen Toughness", "Creature — Wall", 0, 13)
	fourteen := b12Creature(g, opp.ID, "Fourteen Toughness", "Creature — Wall", 0, 14)

	castCatalogSpell(t, g, "Shivan Meteor", "Sorcery", smg1MeteorOracle, b16TargetCard(thirteen))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(thirteen) {
		t.Error("a creature with toughness 13 survived 13 damage")
	}
	castCatalogSpell(t, g, "Shivan Meteor", "Sorcery", smg1MeteorOracle, b16TargetCard(fourteen))
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, fourteen); !ok || c.DamageMarked != 13 {
		t.Errorf("toughness 14 creature: on battlefield %v, damage %d, want 13", ok, c.DamageMarked)
	}
}

func TestSmg1ShivanMeteorCannotTargetAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	if err := castCatalogSpellErr(t, g, "Shivan Meteor", "Sorcery", smg1MeteorOracle, b16TargetPlayer(opp.ID)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a player target: err %v, want ErrIllegalTarget", err)
	}
	if opp.Life != life {
		t.Errorf("life %d→%d", life, opp.Life)
	}
}

func TestSmg1ShivanMeteorDeclaresSuspendTwoForOneRedRed(t *testing.T) {
	sa := suspendOf(smg1MeteorOracle)
	if sa == nil {
		t.Fatal("Shivan Meteor declares no suspend special action")
	}
	if sa.Counters != 2 || sa.Cost != "{1}{R}{R}" {
		t.Errorf("suspend %d—%s, want 2—{1}{R}{R}", sa.Counters, sa.Cost)
	}
	var upkeep, lastCounter int
	for _, tr := range game.CatalogTriggers(smg1MeteorOracle) {
		if !game.TriggerWatchesFromZone(tr, game.ZoneExile) {
			continue
		}
		for _, kind := range tr.Watches {
			switch kind {
			case game.EventBeginUpkeep:
				upkeep++
			case game.EventCounterPlaced:
				lastCounter++
			}
		}
	}
	if upkeep != 1 || lastCounter != 1 {
		t.Errorf("%d upkeep and %d last-counter exile triggers, want one of each (CR 702.62b)", upkeep, lastCounter)
	}
}

// End to end on the real card: suspend it for {1}{R}{R}, wait out two of its
// controller's upkeeps, take the free cast and kill a creature with it. The
// target is picked at that cast, so a creature that arrived while it was
// suspended is a legal one.
func TestSmg1ShivanMeteorSuspendsTwoUpkeepsThenCastsFree(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	active := g.Seats[seat]
	opp := g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	id := handCardForTest(active, "Shivan Meteor", "Sorcery", smg1MeteorOracle)

	// {1}{R}{R} is what it takes: two red is not enough.
	active.ManaPool.AddMana(game.ManaToken{Color: "R"})
	active.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("suspended for {R}{R}: the suspend cost is {1}{R}{R}")
	}
	active.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.PerformSpecialAction(active.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend Shivan Meteor: %v", err)
	}
	if active.ManaPool.Total() != 0 {
		t.Errorf("pool %v after paying {1}{R}{R}", batch01PoolColors(active))
	}
	timeCounters := func() int {
		var n int
		g.ReadSnapshot(func() {
			if c, ok := g.LookupCardForEffect(id); ok {
				n = c.Counters[game.CounterTime]
			}
		})
		return n
	}
	if got := timeCounters(); got != 2 {
		t.Fatalf("%d time counters after suspending, want 2", got)
	}

	// First upkeep: one counter off, no cast yet.
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := timeCounters(); got != 1 {
		t.Fatalf("%d time counters after the first upkeep, want 1", got)
	}
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) != nil {
		t.Fatal("offered the cast with a time counter left")
	}

	// Second upkeep: the last counter comes off and the cast is offered.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) == nil {
		t.Fatal("the last time counter coming off offered no cast")
	}

	victim := b12Creature(g, opp.ID, "Arrived While Suspended", "Creature — Wall", 0, 13)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "exile", Targets: b16TargetCard(victim),
	}); err != nil {
		t.Fatalf("the free cast of a sorcery in an upkeep: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived the Meteor's 13 damage")
	}
	if !active.Graveyard.Contains(id) {
		t.Error("the Meteor is not in its owner's graveyard after resolving")
	}
}

// --- Mox Jasper --------------------------------------------------------------

func TestSmg1MoxJasperNeedsADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mox := seedPermanentWithOracle(g, me.ID, "Mox Jasper", "Legendary Artifact", smg1MoxJasperOracle)

	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet: no Dragon", err)
	}
	if me.ManaPool.Total() != 0 || manaPickFor(g, me.ID) != nil {
		t.Fatalf("a failed gate paid something: pool %v", batch01PoolColors(me))
	}
	if c, _ := battlefieldCard(g, mox); c.Tapped {
		t.Fatal("a failed gate tapped the Mox")
	}

	// A creature that is not a Dragon does not open it.
	smg1Creature(g, me.ID, "Creature — Elf", 1, 1)
	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("an Elf opened it: err = %v", err)
	}

	smg1Creature(g, me.ID, "Creature — Dragon", 4, 4, "flying")
	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility with a Dragon: %v", err)
	}
	pick := manaPickFor(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("pick = %v, want a full five-colour choice", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool = %v, want one {U}", got)
	}
	if c, _ := battlefieldCard(g, mox); !c.Tapped {
		t.Error("the Mox did not tap")
	}
}

// "If you control a Dragon": an opponent's Dragon is not yours.
func TestSmg1MoxJasperIgnoresAnOpponentsDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mox := seedPermanentWithOracle(g, me.ID, "Mox Jasper", "Legendary Artifact", smg1MoxJasperOracle)
	smg1Creature(g, opp.ID, "Creature — Dragon", 4, 4, "flying")
	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("an opponent's Dragon opened it: err = %v", err)
	}
}

// A changeling is every creature type (CR 702.73a), so it is a Dragon.
func TestSmg1MoxJasperCountsAChangeling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mox := seedPermanentWithOracle(g, me.ID, "Mox Jasper", "Legendary Artifact", smg1MoxJasperOracle)
	b34Changeling(g, me.ID, "Shapeshifter")
	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("a changeling is a Dragon: %v", err)
	}
}

// The condition is checked when the ability is activated and never again: a
// mana ability does not use the stack, so the Dragon dying afterwards does not
// take the mana back (CR 605.3b).
func TestSmg1MoxJasperKeepsItsManaWhenTheDragonLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mox := seedPermanentWithOracle(g, me.ID, "Mox Jasper", "Legendary Artifact", smg1MoxJasperOracle)
	dragon := smg1Creature(g, me.ID, "Creature — Dragon", 4, 4)
	if err := g.ActivateManaAbility(me.ID, mox, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	removeInResponse(t, g, dragon, false)
	b10ResolveAllManaPicks(t, g, me.ID, "R")
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool = %v, want one {R}", got)
	}
}

// The auto-tapper reads the same gate: with no Dragon the Mox is not a source,
// with one it pays for a spell.
func TestSmg1MoxJasperIsAnAutoTapSourceOnlyWithADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	mox := seedPermanentWithOracle(g, me.ID, "Mox Jasper", "Legendary Artifact", smg1MoxJasperOracle)
	cost, err := game.ParseCost("{R}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Fatalf("a plan %v for {R} off a Mox with no Dragon", plan)
	}
	smg1Creature(g, me.ID, "Creature — Dragon", 4, 4)
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 || plan[0] != mox {
		t.Fatalf("with a Dragon the Mox pays {R}: plan %v, ok %v", plan, ok)
	}
}

// --- Dragonspeaker Shaman ----------------------------------------------------

func TestSmg1DragonspeakerShamanDiscountsDragonSpellsByTwoGeneric(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Dragonspeaker Shaman", "Creature — Human Barbarian Shaman", smg1ShamanOracle, false)

	if got := priceInHand(t, g, me, "Big Dragon", "Creature — Dragon", "{5}{B}{R}"); got != 5 {
		t.Errorf("{5}{B}{R} Dragon: %d, want 5 ({3}{B}{R})", got)
	}
	if got := priceInHand(t, g, me, "Mid Dragon", "Creature — Dragon", "{2}{R}{R}"); got != 2 {
		t.Errorf("{2}{R}{R} Dragon: %d, want 2 ({R}{R})", got)
	}
	// CR 601.2f: only the generic part is reduced, and it stops at zero.
	if got := priceInHand(t, g, me, "One Generic Dragon", "Creature — Dragon", "{1}{R}"); got != 1 {
		t.Errorf("{1}{R} Dragon: %d, want 1 ({R})", got)
	}
	if got := priceInHand(t, g, me, "Coloured Dragon", "Creature — Dragon", "{R}{R}"); got != 2 {
		t.Errorf("{R}{R} Dragon: %d, want 2 (coloured pips untouched)", got)
	}
	if got := priceInHand(t, g, me, "Kindred Dragon", "Kindred Creature — Dragon", "{4}{R}"); got != 3 {
		t.Errorf("a Dragon with another type: %d, want 3", got)
	}
	// Only Dragon spells, and only yours.
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{2}{G}"); got != 3 {
		t.Errorf("a non-Dragon creature: %d, want 3 (undiscounted)", got)
	}
	if got := priceInHand(t, g, me, "Shock", "Instant", "{R}"); got != 1 {
		t.Errorf("a noncreature spell: %d, want 1 (undiscounted)", got)
	}
	if got := priceInHand(t, g, opp, "Their Dragon", "Creature — Dragon", "{5}{R}"); got != 6 {
		t.Errorf("an opponent's Dragon: %d, want 6 (undiscounted)", got)
	}
}

// The Shaman is a Human Barbarian Shaman: it does not discount itself.
func TestSmg1DragonspeakerShamanDoesNotDiscountItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Dragonspeaker Shaman", "Creature — Human Barbarian Shaman", smg1ShamanOracle, false)
	if got := priceInHand(t, g, me, "Dragonspeaker Shaman", "Creature — Human Barbarian Shaman", "{1}{R}{R}"); got != 3 {
		t.Errorf("a second Shaman: %d, want 3", got)
	}
}

// A changeling spell is a Dragon spell (CR 702.73a).
func TestSmg1DragonspeakerShamanDiscountsAChangelingSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Dragonspeaker Shaman", "Creature — Human Barbarian Shaman", smg1ShamanOracle, false)
	base, err := game.ParseCost("{3}{R}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(), Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter",
			ManaCost: "{3}{R}", Keywords: []string{game.KeywordChangeling},
			Owner: me.ID, Controller: me.ID,
		},
		Controller: me.ID,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	if got := out.ManaValue(); got != 2 {
		t.Errorf("a changeling spell: %d, want 2 ({1}{R})", got)
	}
}

// Two Shamans stack, still on the generic half only.
func TestSmg1TwoDragonspeakerShamansStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Dragonspeaker Shaman", "Creature — Human Barbarian Shaman", smg1ShamanOracle, false)
	pushCatalogPermanent(g, me.ID, "Dragonspeaker Shaman", "Creature — Human Barbarian Shaman", smg1ShamanOracle, false)
	if got := priceInHand(t, g, me, "Big Dragon", "Creature — Dragon", "{5}{B}{R}"); got != 3 {
		t.Errorf("{5}{B}{R} Dragon under two Shamans: %d, want 3 ({1}{B}{R})", got)
	}
	if got := priceInHand(t, g, me, "Mid Dragon", "Creature — Dragon", "{2}{R}{R}"); got != 2 {
		t.Errorf("{2}{R}{R} Dragon under two Shamans: %d, want 2 (generic spent, {R}{R} stays)", got)
	}
}

// The discount is real at the till: a {2}{R}{R} Dragon is cast off two
// Mountains with the Shaman out, and refused without it.
func TestSmg1DragonspeakerShamanDiscountIsPaidAtTheTill(t *testing.T) {
	cast := func(t *testing.T, withShaman bool) error {
		g := newCatalogGame(t)
		me := g.Seats[0]
		toMainForCost(t, g)
		if withShaman {
			pushCatalogPermanent(g, me.ID, "Dragonspeaker Shaman", "Creature — Human Barbarian Shaman", smg1ShamanOracle, false)
		}
		seedManaLand(g, me.ID, "Mountain", "Basic Land — Mountain", "R")
		seedManaLand(g, me.ID, "Mountain", "Basic Land — Mountain", "R")
		id := handCardFull(me, "Test Dragon", "Creature — Dragon", "{2}{R}{R}", "", []string{"R"})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"})
	}
	if err := cast(t, true); err != nil {
		t.Errorf("{2}{R}{R} Dragon off two Mountains with the Shaman: %v", err)
	}
	if err := cast(t, false); !errors.Is(err, game.ErrInsufficientMana) {
		t.Errorf("{2}{R}{R} Dragon off two Mountains without the Shaman: err %v, want ErrInsufficientMana", err)
	}
}

// Smaug is a Dragon commander: the discount applies to a cast from the command
// zone, after the commander tax (CR 601.2f), still on the generic half only.
func TestSmg1DragonspeakerShamanDiscountsADragonCommanderAfterTax(t *testing.T) {
	cast := func(t *testing.T, tokens int) error {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushCatalogPermanent(g, me.ID, "Dragonspeaker Shaman", "Creature — Human Barbarian Shaman", smg1ShamanOracle, false)
		cmdr := game.NewCommander("Test Dragon Commander", me.ID)
		cmdr.TypeLine = "Legendary Creature — Dragon"
		cmdr.ManaCost = "{5}{B}{R}"
		cmdr.ColorIdentity = []string{"B", "R"}
		cmdr.Power, cmdr.Toughness = 8, 7
		me.Command.PushTop(cmdr)
		me.CommanderCasts = map[uuid.UUID]int{cmdr.InstanceID: 1} // one prior cast: {2} tax
		toMainForCost(t, g)
		me.ManaPool.AddMana(game.ManaToken{Color: "B"})
		me.ManaPool.AddMana(game.ManaToken{Color: "R"})
		for i := 2; i < tokens; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "R"})
		}
		return g.CastSpell(me.ID, cmdr.InstanceID, game.CastSpellParams{FromZone: "command", Strict: true})
	}
	// {5}{B}{R} + {2} tax - {2} = {5}{B}{R}: seven mana, not nine.
	if err := cast(t, 6); !errors.Is(err, game.ErrInsufficientMana) {
		t.Errorf("six mana for a taxed, discounted {5}{B}{R}: err %v, want ErrInsufficientMana", err)
	}
	if err := cast(t, 7); err != nil {
		t.Errorf("seven mana for a taxed, discounted {5}{B}{R}: %v", err)
	}
}

// --- Cut Propulsion ----------------------------------------------------------

func TestSmg1CutPropulsionHitsAGroundCreatureForItsPower(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	survivor := smg1Creature(g, opp.ID, "Creature — Ogre", 3, 4)
	doomed := smg1Creature(g, opp.ID, "Creature — Ogre", 3, 3)

	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(survivor))
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, survivor); !ok || c.DamageMarked != 3 {
		t.Errorf("3/4 ground creature: on battlefield %v, damage %d, want 3", ok, c.DamageMarked)
	}
	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(doomed))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(doomed) {
		t.Error("a 3/3 survived dealing 3 damage to itself")
	}
}

func TestSmg1CutPropulsionDoublesForAFlyingCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	survivor := smg1Creature(g, opp.ID, "Creature — Bird", 2, 5, "flying")
	doomed := smg1Creature(g, opp.ID, "Creature — Dragon", 3, 6, "flying")

	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(survivor))
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, survivor); !ok || c.DamageMarked != 4 {
		t.Errorf("2/5 flier: on battlefield %v, damage %d, want 4 (twice its power, not 2 + 4)", ok, c.DamageMarked)
	}
	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(doomed))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(doomed) {
		t.Error("a 3/6 flier survived dealing 6 damage to itself")
	}
}

// "Instead": a flier takes twice its power, never its power on top.
func TestSmg1CutPropulsionHitsYourOwnCreatureToo(t *testing.T) {
	g := newCatalogGame(t)
	mine := smg1Creature(g, g.Seats[0].ID, "Creature — Bird", 2, 3, "flying")
	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("a 2/3 flier survived dealing 4 damage to itself")
	}
}

// The creature is the damage source, so its lifelink gains ITS controller life
// (CR 702.15b), not the caster.
func TestSmg1CutPropulsionUsesTheCreaturesLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := smg1Creature(g, opp.ID, "Creature — Angel", 3, 9, "lifelink")
	myLife, theirLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(target))
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife+3 {
		t.Errorf("the lifelinker's controller: %d → %d, want +3", theirLife, opp.Life)
	}
	if me.Life != myLife {
		t.Errorf("the caster's life changed: %d → %d", myLife, me.Life)
	}
}

// Deathtouch on the source (CR 702.2b): one point of damage from itself kills
// a creature with any toughness.
func TestSmg1CutPropulsionUsesTheCreaturesDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	target := smg1Creature(g, g.Seats[1].ID, "Creature — Snake", 1, 10, "deathtouch")
	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(target))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Error("a deathtouch 1/10 survived dealing 1 damage to itself")
	}
}

// Power is read as the spell resolves, not when it was cast.
func TestSmg1CutPropulsionReadsPowerAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	target := smg1Creature(g, g.Seats[1].ID, "Creature — Ogre", 2, 9)
	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(target))
	smg1Pump(t, g, target, 3)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, target); !ok || c.DamageMarked != 5 {
		t.Errorf("pumped to 5 in response: on battlefield %v, damage %d, want 5", ok, c.DamageMarked)
	}
}

// "If that creature has flying" is asked as the spell resolves too: a creature
// that lost flying in response takes its power once.
func TestSmg1CutPropulsionAsksAboutFlyingAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	target := smg1Creature(g, g.Seats[1].ID, "Creature — Bird", 2, 9, "flying")
	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(target))
	smg1LoseKeyword(g, target)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, target); !ok || c.DamageMarked != 2 {
		t.Errorf("flying lost in response: on battlefield %v, damage %d, want 2", ok, c.DamageMarked)
	}
}

// And the other way round: a creature that is a flier when the spell resolves
// is doubled, whatever it was when the spell was cast.
func TestSmg1CutPropulsionDoublesAPumpedFlierFromItsCurrentPower(t *testing.T) {
	g := newCatalogGame(t)
	target := smg1Creature(g, g.Seats[1].ID, "Creature — Bird", 1, 20, "flying")
	castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(target))
	smg1Pump(t, g, target, 2)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, target); c.DamageMarked != 6 {
		t.Errorf("a 3-power flier: damage %d, want 6", c.DamageMarked)
	}
}

// A creature with no power deals no damage, doubled or not.
func TestSmg1CutPropulsionDealsNothingWithNoPower(t *testing.T) {
	g := newCatalogGame(t)
	wall := smg1Creature(g, g.Seats[1].ID, "Creature — Wall", 0, 4)
	bird := smg1Creature(g, g.Seats[1].ID, "Creature — Bird", 0, 4, "flying")
	for _, id := range []uuid.UUID{wall, bird} {
		castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(id))
		passPriorityAroundTable(t, g)
		if c, ok := battlefieldCard(g, id); !ok || c.DamageMarked != 0 {
			t.Errorf("a 0-power creature: on battlefield %v, damage %d, want 0", ok, c.DamageMarked)
		}
	}
}

func TestSmg1CutPropulsionCannotTargetAPlayerOrAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	walker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 5)
	if err := castCatalogSpellErr(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetPlayer(opp.ID)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a player target: err %v, want ErrIllegalTarget", err)
	}
	if err := castCatalogSpellErr(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(walker)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a planeswalker target: err %v, want ErrIllegalTarget", err)
	}
}

// A creature that left in response is an illegal target: nothing happens.
func TestSmg1CutPropulsionFizzlesWhenTheCreatureLeaves(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := smg1Creature(g, opp.ID, "Creature — Ogre", 3, 3)
	spell := castCatalogSpell(t, g, "Cut Propulsion", "Instant", smg1CutPropulsionOracle, b16TargetCard(target))
	removeInResponse(t, g, target, true)
	passPriorityAroundTable(t, g)
	if !g.Seats[0].Graveyard.Contains(spell) {
		t.Error("the fizzled Cut Propulsion is not in its owner's graveyard")
	}
}

// --- Self-Destruct -----------------------------------------------------------

// smg1SelfDestructTargets is the two-clause announcement: the creature, then
// whatever it hits.
func smg1SelfDestructTargets(creature uuid.UUID, other game.TargetRef) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: creature}, other}
}

func smg1CardRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func TestSmg1SelfDestructDealsItsPowerToTheTargetAndToItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Ogre", 3, 6)
	theirs := smg1Creature(g, opp.ID, "Creature — Ogre", 2, 3)

	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, smg1CardRef(theirs)))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("their 2/3 survived 3 damage")
	}
	if c, ok := battlefieldCard(g, mine); !ok || c.DamageMarked != 3 {
		t.Errorf("my 3/6: on battlefield %v, damage %d, want 3 to itself", ok, c.DamageMarked)
	}
}

func TestSmg1SelfDestructKillsTheCreatureThatHitsItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Ogre", 4, 4)
	theirs := smg1Creature(g, opp.ID, "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, smg1CardRef(theirs)))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("a 4/4 survived dealing 4 damage to itself")
	}
	if c, _ := battlefieldCard(g, theirs); c.DamageMarked != 4 {
		t.Errorf("their wall took %d, want 4", c.DamageMarked)
	}
}

// "Any other target": a player, a planeswalker or a battle, not just a creature.
func TestSmg1SelfDestructHitsAPlayerOrAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Ogre", 2, 20)
	walker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 5)
	life := opp.Life

	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}))
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("target player: life %d → %d, want -2", life, opp.Life)
	}

	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, smg1CardRef(walker)))
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, walker); got != 3 {
		t.Errorf("target planeswalker: loyalty %d, want 3", got)
	}
	if c, _ := battlefieldCard(g, mine); c.DamageMarked != 4 {
		t.Errorf("my creature took %d across two spells, want 4", c.DamageMarked)
	}
}

// "Any OTHER target": the creature cannot be both of its own targets.
func TestSmg1SelfDestructCannotTargetTheCreatureTwice(t *testing.T) {
	g := newCatalogGame(t)
	mine := smg1Creature(g, g.Seats[0].ID, "Creature — Ogre", 3, 3)
	err := castCatalogSpellErr(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, smg1CardRef(mine)))
	// A clause marked Distinct refuses a pick an earlier clause already holds.
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("the creature as its own other target: err %v, want ErrInvalidParam", err)
	}
	if c, _ := battlefieldCard(g, mine); c.DamageMarked != 0 {
		t.Errorf("a refused announcement dealt %d", c.DamageMarked)
	}
}

// "Target creature you control": another player's creature is not a legal
// first target, and a player is not a creature.
func TestSmg1SelfDestructNeedsACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Ogre", 3, 3)
	theirs := smg1Creature(g, opp.ID, "Creature — Ogre", 3, 3)
	err := castCatalogSpellErr(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(theirs, smg1CardRef(mine)))
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("an opponent's creature as the source: err %v, want ErrIllegalTarget", err)
	}
	err = castCatalogSpellErr(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}, smg1CardRef(theirs)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a player as the source: err %v, want ErrIllegalTarget", err)
	}
	for _, id := range []uuid.UUID{mine, theirs} {
		if c, _ := battlefieldCard(g, id); c.DamageMarked != 0 {
			t.Errorf("a refused announcement dealt %d", c.DamageMarked)
		}
	}
}

// X is the creature's power as the spell resolves, one number for both hits.
func TestSmg1SelfDestructReadsPowerAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Ogre", 2, 20)
	theirs := smg1Creature(g, opp.ID, "Creature — Wall", 0, 20)
	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, smg1CardRef(theirs)))
	smg1Pump(t, g, mine, 3)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, theirs); c.DamageMarked != 5 {
		t.Errorf("the other target took %d, want 5 (power after the pump)", c.DamageMarked)
	}
	if c, _ := battlefieldCard(g, mine); c.DamageMarked != 5 {
		t.Errorf("the creature took %d, want the same 5", c.DamageMarked)
	}
}

// The creature is the source of both hits: its lifelink counts both (CR 702.15b).
func TestSmg1SelfDestructUsesTheCreaturesLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Angel", 3, 9, "lifelink")
	life := me.Life
	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}))
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6 (3 to the player and 3 to itself)", life, me.Life)
	}
}

// Deathtouch (CR 702.2b) kills the other target with 1 damage, and the
// creature with the 1 it deals itself.
func TestSmg1SelfDestructUsesTheCreaturesDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Snake", 1, 10, "deathtouch")
	theirs := smg1Creature(g, opp.ID, "Creature — Wall", 0, 12)
	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, smg1CardRef(theirs)))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("a deathtouch 1-power hit did not kill the 0/12")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the deathtouch creature survived dealing 1 damage to itself")
	}
}

// The other target left in response: the creature, still legal, still deals X
// damage to itself, because that half of the sentence names no illegal target.
func TestSmg1SelfDestructStillHurtsTheCreatureWhenTheOtherTargetLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Ogre", 3, 9)
	theirs := smg1Creature(g, opp.ID, "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, smg1CardRef(theirs)))
	removeInResponse(t, g, theirs, true)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, mine); !ok || c.DamageMarked != 3 {
		t.Errorf("my creature: on battlefield %v, damage %d, want 3 to itself", ok, c.DamageMarked)
	}
}

// The creature left in response: it is an illegal target and deals nothing, so
// the other target is not hit either.
func TestSmg1SelfDestructDealsNothingWhenTheCreatureLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Ogre", 5, 5)
	theirs := smg1Creature(g, opp.ID, "Creature — Wall", 0, 9)
	life := opp.Life
	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, smg1CardRef(theirs)))
	removeInResponse(t, g, mine, false)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, theirs); c.DamageMarked != 0 {
		t.Errorf("the other target took %d from a creature that was gone", c.DamageMarked)
	}
	if opp.Life != life {
		t.Errorf("life changed: %d → %d", life, opp.Life)
	}
}

// A creature with no power deals nothing to anyone, itself included.
func TestSmg1SelfDestructDealsNothingWithNoPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := smg1Creature(g, me.ID, "Creature — Wall", 0, 4)
	life := opp.Life
	castCatalogSpell(t, g, "Self-Destruct", "Instant", smg1SelfDestructOracle,
		smg1SelfDestructTargets(mine, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}))
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d → %d, want no change", life, opp.Life)
	}
	if c, ok := battlefieldCard(g, mine); !ok || c.DamageMarked != 0 {
		t.Errorf("0-power creature: on battlefield %v, damage %d", ok, c.DamageMarked)
	}
}
