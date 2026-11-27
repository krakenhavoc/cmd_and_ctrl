package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b786_cards_test.go — Visions: Damage, batch 786 (S82 older sets,
// #5631). Every test plays the real catalog entry through the engine.

const (
	oldB786RockSlide     = "0b2e7c1a-dc07-4138-a560-dc80fa723860"
	oldB786LightningCld  = "4a574140-0657-4f49-b497-331447f17b29"
	oldB786Kookus        = "4c8196e0-c392-4402-b77c-e3d009e9899c"
	oldB786Triangle      = "6d4daf9c-e2f4-4d0e-b2e4-ded04d644b11"
	oldB786SwineRider    = "9ebe32fc-3b6e-4cb9-9af0-1a6c3c6ad885"
	oldB786Desolation    = "a80c3af8-470e-4e95-9579-cfdf18fbdbfa"
	oldB786InfernalHarv  = "b8c8cde9-34db-4185-86f5-4289969880c5"
	oldB786Wildebeests   = "c270c770-91ed-4540-b281-c4b60b943b34"
	oldB786BearTypeLine  = "Creature — Bear"
	oldB786BasicSwamp    = "Basic Land — Swamp"
	oldB786BasicPlains   = "Basic Land — Plains"
	oldB786BasicMountain = "Basic Land — Mountain"
)

func TestOldB786EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB786RockSlide: "Rock Slide", oldB786LightningCld: "Lightning Cloud", oldB786Kookus: "Kookus",
		oldB786Triangle: "Triangle of War", oldB786SwineRider: "Goblin Swine-Rider", oldB786Desolation: "Desolation",
		oldB786InfernalHarv: "Infernal Harvest", oldB786Wildebeests: "Stampeding Wildebeests",
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

func oldB786Wall(g *game.Game, owner uuid.UUID, name string, keywords ...string) uuid.UUID {
	return effectstest.ReqCreature(g, owner, name, "", 0, 30, keywords...)
}

// --- Rock Slide -------------------------------------------------------------

func TestOldB786RockSlideDividesXAmongNonflyingAttackersAndBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	a := oldB786Wall(g, me.ID, "Ground Attacker")
	flier := oldB786Wall(g, me.ID, "Flying Attacker", "flying")
	idle := oldB786Wall(g, me.ID, "Idle")
	blocker := oldB786Wall(g, opp.ID, "Ground Blocker")
	flyBlocker := oldB786Wall(g, opp.ID, "Flying Blocker", "flying")
	effectstest.Pr7bAttack(t, g, opp.ID, a, flier)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blocker: a, flyBlocker: flier})
	id := effectstest.HandCardFull(me, "Rock Slide", "Instant", "{X}{R}", oldB786RockSlide, []string{"R"})
	effectstest.B10AddMana(me, "R", "R", "R", "R")
	for name, target := range map[string]uuid.UUID{"a flying attacker": flier, "a flying blocker": flyBlocker, "a creature not in combat": idle} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(target),
			Distribution: map[uuid.UUID]int{target: 3}}); err == nil {
			t.Fatalf("Rock Slide targeted %s", name)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(a, blocker),
		Distribution: map[uuid.UUID]int{a: 1, blocker: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if gotA, gotB := effectstest.DamageMarkedOn(g, a), effectstest.DamageMarkedOn(g, blocker); gotA != 1 || gotB != 2 {
		t.Errorf("damage %d/%d, want 1/2", gotA, gotB)
	}
	for _, untouched := range []uuid.UUID{flier, flyBlocker, idle} {
		if n := effectstest.DamageMarkedOn(g, untouched); n != 0 {
			t.Errorf("a creature that was not a target took %d", n)
		}
	}
}

// --- Infernal Harvest --------------------------------------------------------

func TestOldB786InfernalHarvestReturnsXSwampsAndDividesXDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	s1 := effectstest.CcLand(g, me.ID, "Swamp", oldB786BasicSwamp)
	s2 := effectstest.CcLand(g, me.ID, "Swamp", oldB786BasicSwamp)
	s3 := effectstest.CcLand(g, me.ID, "Swamp", oldB786BasicSwamp)
	forest := effectstest.CcLand(g, me.ID, "Forest", "Basic Land — Forest")
	x := oldB786Wall(g, opp.ID, "Wall A")
	y := oldB786Wall(g, opp.ID, "Wall B")
	id := effectstest.HandCardFull(me, "Infernal Harvest", "Sorcery", "{1}{B}", oldB786InfernalHarv, []string{"B"})
	effectstest.B10AddMana(me, "B", "B")
	params := func(ret ...uuid.UUID) game.CastSpellParams {
		return game.CastSpellParams{Strict: true, XValue: 2, ReturnIDs: ret, Targets: effectstest.CardRefs(x, y),
			Distribution: map[uuid.UUID]int{x: 1, y: 1}}
	}
	for name, ret := range map[string][]uuid.UUID{"one Swamp for X=2": {s1}, "a Forest": {s1, forest}, "the same Swamp twice": {s1, s1}} {
		if err := g.CastSpell(me.ID, id, params(ret...)); err == nil {
			t.Fatalf("%s: the cast was accepted", name)
		}
		if !me.Hand.Contains(id) || me.Hand.Contains(s1) {
			t.Fatalf("%s: a refused cast moved something", name)
		}
	}
	// X is also the damage: 3 damage cannot be divided as 1+1 under X=2.
	bad := params(s1, s2)
	bad.Distribution = map[uuid.UUID]int{x: 1, y: 2}
	if err := g.CastSpell(me.ID, id, bad); err == nil {
		t.Fatal("a division that does not sum to X was accepted")
	}
	if err := g.CastSpell(me.ID, id, params(s1, s2)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !me.Hand.Contains(s1) || !me.Hand.Contains(s2) || me.Hand.Contains(s3) {
		t.Error("exactly the two named Swamps should have returned to hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if gx, gy := effectstest.DamageMarkedOn(g, x), effectstest.DamageMarkedOn(g, y); gx != 1 || gy != 1 {
		t.Errorf("damage %d/%d, want 1/1", gx, gy)
	}
}

func TestOldB786InfernalHarvestForZeroIsLegalAndDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	swamp := effectstest.CcLand(g, me.ID, "Swamp", oldB786BasicSwamp)
	wall := oldB786Wall(g, opp.ID, "Wall")
	id := effectstest.HandCardFull(me, "Infernal Harvest", "Sorcery", "{1}{B}", oldB786InfernalHarv, []string{"B"})
	effectstest.B10AddMana(me, "B", "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 0}); err != nil {
		t.Fatalf("X=0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(swamp) || effectstest.DamageMarkedOn(g, wall) != 0 {
		t.Error("X=0 should return nothing and damage nothing")
	}
}

// --- Goblin Swine-Rider ------------------------------------------------------

func TestOldB786SwineRiderDamagesEveryAttackerAndBlockerOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	rider := effectstest.B12Push(g, me.ID, "Goblin Swine-Rider", "Creature — Goblin", oldB786SwineRider, 1, 1)
	other := oldB786Wall(g, me.ID, "Other Attacker")
	idle := oldB786Wall(g, me.ID, "Idle")
	b1 := oldB786Wall(g, opp.ID, "Blocker One")
	b2 := oldB786Wall(g, opp.ID, "Blocker Two")
	b3 := oldB786Wall(g, opp.ID, "Blocker Three")
	effectstest.Pr7bAttack(t, g, opp.ID, rider, other)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: rider, b2: rider, b3: other})
	for name, id := range map[string]uuid.UUID{"other attacker": other, "first blocker": b1, "second blocker": b2, "third blocker": b3} {
		if n := effectstest.DamageMarkedOn(g, id); n != 2 {
			t.Errorf("%s took %d, want 2 (one trigger however many block the Swine-Rider)", name, n)
		}
	}
	if n := effectstest.DamageMarkedOn(g, idle); n != 0 {
		t.Errorf("a creature not in combat took %d", n)
	}
	if g.Battlefield.Contains(rider) {
		t.Error("the 1/1 Swine-Rider survived its own 2 damage")
	}
}

func TestOldB786SwineRiderDoesNothingUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	rider := effectstest.B12Push(g, me.ID, "Goblin Swine-Rider", "Creature — Goblin", oldB786SwineRider, 1, 1)
	wall := oldB786Wall(g, opp.ID, "Wall")
	effectstest.Pr7bAttack(t, g, opp.ID, rider)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{})
	if !g.Battlefield.Contains(rider) || effectstest.DamageMarkedOn(g, wall) != 0 {
		t.Error("an unblocked Swine-Rider triggered")
	}
}

// --- Triangle of War ---------------------------------------------------------

func TestOldB786TriangleOfWarSacrificesAndFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	tri := effectstest.PushCatalogPermanent(g, me.ID, "Triangle of War", "Artifact", oldB786Triangle, false)
	mine := effectstest.B12Creature(g, me.ID, oldB786BearTypeLine, oldB786BearTypeLine, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB786BearTypeLine, 2, 2)
	refs := func(a, b uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1}}
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, tri, 0, game.ActivateAbilityParams{Strict: true, Targets: refs(theirs, mine)}); err == nil {
		t.Fatal("the opponent's creature was accepted as the creature I control")
	}
	if !g.Battlefield.Contains(tri) {
		t.Fatal("a refused activation sacrificed the artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, tri, 0, game.ActivateAbilityParams{Strict: true, Targets: refs(mine, theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(tri) {
		t.Error("the artifact was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature survived 3 damage")
	}
	if n := effectstest.DamageMarkedOn(g, mine); n != 2 {
		t.Errorf("my creature has %d damage, want 2", n)
	}
}

func TestOldB786TriangleOfWarDoesNothingWhenATargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	tri := effectstest.PushCatalogPermanent(g, me.ID, "Triangle of War", "Artifact", oldB786Triangle, false)
	mine := effectstest.B12Creature(g, me.ID, oldB786BearTypeLine, oldB786BearTypeLine, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB786BearTypeLine, 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, tri, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirs); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, mine); n != 0 {
		t.Errorf("my creature took %d from a fight that had no second creature", n)
	}
}

// --- Stampeding Wildebeests --------------------------------------------------

func oldB786Green(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: oldB786BearTypeLine, Power: 2, Toughness: 2,
		Colors: []string{"G"}, Owner: owner, Controller: owner,
	})
}

func TestOldB786WildebeestsReturnsAGreenCreatureEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beasts := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stampeding Wildebeests", TypeLine: "Creature — Antelope Beast", OracleID: oldB786Wildebeests,
		Power: 5, Toughness: 4, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID,
	})
	green := oldB786Green(g, me.ID, "Green Bear")
	red := effectstest.B12Creature(g, me.ID, "Red Bear", oldB786BearTypeLine, 2, 2)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no return pick")
	}
	if p.ChooseMin != 1 {
		t.Error("the return is mandatory")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{red}); err == nil {
		t.Fatal("a non-green creature was accepted")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{green}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(green) || !me.Hand.Contains(green) {
		t.Error("the green Bear did not return to hand")
	}
	if !g.Battlefield.Contains(beasts) || !g.Battlefield.Contains(red) {
		t.Error("something other than the chosen creature left")
	}
}

func TestOldB786WildebeestsReturnsItselfWhenItIsTheOnlyGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beasts := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stampeding Wildebeests", TypeLine: "Creature — Antelope Beast", OracleID: oldB786Wildebeests,
		Power: 5, Toughness: 4, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID,
	})
	red := effectstest.B12Creature(g, me.ID, "Red Bear", oldB786BearTypeLine, 2, 2)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); p != nil {
		if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{beasts}); err != nil {
			t.Fatalf("ResolveOwnPermanents: %v", err)
		}
	}
	if g.Battlefield.Contains(beasts) || !me.Hand.Contains(beasts) {
		t.Error("the Wildebeests did not return itself")
	}
	if !g.Battlefield.Contains(red) {
		t.Error("the non-green creature left")
	}
}

func TestOldB786WildebeestsHasTrample(t *testing.T) {
	spec, ok := effects.Lookup(oldB786Wildebeests)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "trample" {
		t.Fatalf("printed keywords = %v, want [trample]", spec.PrintedKeywords)
	}
}

// --- Kookus ------------------------------------------------------------------

func oldB786KookusTable(t *testing.T, withKeeper bool) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Kookus", "Creature — Djinn", oldB786Kookus, 3, 5)
	if withKeeper {
		effectstest.B12Creature(g, me.ID, "Keeper of Kookus", "Creature — Goblin", 0, 2)
	}
	return g, me, k
}

func TestOldB786KookusHurtsYouAndMustAttackWithoutItsKeeper(t *testing.T) {
	g, me, k := oldB786KookusTable(t, false)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, me.Life)
	}
	if !effectstest.FaMustAttack(t, g) {
		t.Error("Kookus is not required to attack this turn")
	}
	_ = k
}

func TestOldB786KookusIsQuietWithItsKeeper(t *testing.T) {
	g, me, _ := oldB786KookusTable(t, true)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d with Keeper of Kookus on the battlefield", life, me.Life)
	}
	if effectstest.FaMustAttack(t, g) {
		t.Error("Kookus must attack although its Keeper is out")
	}
}

func TestOldB786KookusFirebreathes(t *testing.T) {
	g, me, k := oldB786KookusTable(t, true)
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "R", "R")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("pump %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.PtOf(t, g, k); p != 5 || tough != 5 {
		t.Errorf("Kookus is %d/%d, want 5/5", p, tough)
	}
}

// --- Lightning Cloud ---------------------------------------------------------

func oldB786CloudCast(t *testing.T, g *game.Game, me *game.Player, color, cost string) {
	t.Helper()
	spell := effectstest.DstHand(me, "Spell", "Sorcery", cost, color)
	effectstest.B10AddMana(me, color)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the %s spell: %v", color, err)
	}
}

func TestOldB786LightningCloudPingsAnyTargetWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Lightning Cloud", "Enchantment", oldB786LightningCld, false)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB786BearTypeLine, 2, 2)
		oldB786CloudCast(t, g, me, "R", "{R}")
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.FloatForTest(g, me, "R")
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		if got := effectstest.DamageMarkedOn(g, bear); got != want {
			t.Errorf("pay=%v: the Bear took %d, want %d", pay, got, want)
		}
	}
}

func TestOldB786LightningCloudCanHitAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Lightning Cloud", "Enchantment", oldB786LightningCld, false)
	oldB786CloudCast(t, g, me, "R", "{R}")
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	life := opp.Life
	effectstest.FloatForTest(g, me, "R")
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent's life %d -> %d, want -1", life, opp.Life)
	}
}

func TestOldB786LightningCloudIgnoresOtherColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Lightning Cloud", "Enchantment", oldB786LightningCld, false)
	oldB786CloudCast(t, g, me, "U", "{U}")
	if effectstest.TriggerOnStackFrom(g, "Lightning Cloud") {
		t.Error("a blue spell triggered Lightning Cloud")
	}
}

// --- Desolation --------------------------------------------------------------

// oldB786DesolationEndStep walks to the end step of the active seat's turn and
// lets Desolation's trigger resolve; sacrifice answers each seat's pick from
// `picks` as it is asked.
func oldB786DesolationEndStep(t *testing.T, g *game.Game, picks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 6; i++ {
		answered := false
		for _, seat := range g.Seats {
			if p := effectstest.LatestChooseCards(g, seat.ID); p != nil {
				effectstest.AnswerChooseCards(t, g, seat.ID, picks[seat.ID])
				answered = true
			}
		}
		if !answered {
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB786DesolationMakesTappersSacrificeAndHurtsPlainsSacrificers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Desolation", "Enchantment", oldB786Desolation, false)
	myMountain := effectstest.CcLand(g, me.ID, "Mountain", oldB786BasicMountain)
	myPlains := effectstest.CcLand(g, me.ID, "Plains", oldB786BasicPlains)
	oppPlains := effectstest.CcLand(g, opp.ID, "Plains", oldB786BasicPlains)
	oppSwamp := effectstest.CcLand(g, opp.ID, "Swamp", oldB786BasicSwamp)
	if err := g.ActivateManaAbility(me.ID, myMountain, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Mountain: %v", err)
	}
	// The opponent taps a land too (CR 106.12).
	if err := g.ActivateManaAbility(opp.ID, oppSwamp, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the opponent taps the Swamp: %v", err)
	}
	myLife, oppLife := me.Life, opp.Life
	oldB786DesolationEndStep(t, g, map[uuid.UUID]uuid.UUID{me.ID: myPlains, opp.ID: oppSwamp})
	if g.Battlefield.Contains(myPlains) || !g.Battlefield.Contains(myMountain) {
		t.Error("I should have sacrificed the Plains and kept the Mountain")
	}
	if g.Battlefield.Contains(oppSwamp) || !g.Battlefield.Contains(oppPlains) {
		t.Error("the opponent should have sacrificed the Swamp and kept the Plains")
	}
	if me.Life != myLife-2 {
		t.Errorf("my life %d -> %d, want -2 for sacrificing a Plains", myLife, me.Life)
	}
	if opp.Life != oppLife {
		t.Errorf("the opponent's life %d -> %d, want no change for sacrificing a Swamp", oppLife, opp.Life)
	}
}

func TestOldB786DesolationSparesPlayersWhoTappedNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Desolation", "Enchantment", oldB786Desolation, false)
	myPlains := effectstest.CcLand(g, me.ID, "Plains", oldB786BasicPlains)
	oppPlains := effectstest.CcLand(g, opp.ID, "Plains", oldB786BasicPlains)
	myLife, oppLife := me.Life, opp.Life
	oldB786DesolationEndStep(t, g, nil)
	if !g.Battlefield.Contains(myPlains) || !g.Battlefield.Contains(oppPlains) {
		t.Error("a land was sacrificed though nobody tapped a land for mana")
	}
	if me.Life != myLife || opp.Life != oppLife {
		t.Error("damage was dealt though nobody sacrificed")
	}
}
