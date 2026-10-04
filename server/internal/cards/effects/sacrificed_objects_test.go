package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrificed_objects_test.go — ADR 0113 §1 (#2072): a spell reads what
// its sacrifice cost took. The payment record names each sacrificed
// permanent (PaidCost.SacrificedObjects), and the effect reads it as it
// last existed on the battlefield (CR 400.7j, 608.2h). One test per
// interaction the ADR lists.

const (
	flingOracle         = "24227761-b50e-4b9e-93a2-e82d053b3e3d"
	kazuulsFuryOracle   = "f8410804-632b-4f18-9a73-6dccc7e4582d"
	momentousFallOracle = "6a5ce1b0-ac78-4a74-9eb7-4064e1687bf1"
	tendThePestsOracle  = "9ffce1d5-e29c-4dab-8b97-6b32655785d4"
	corpseCobbleOracle  = "5b0b6c8f-472a-4ee2-8368-3b17a2df498d"
)

// soCreature puts a creature with the given power and toughness onto
// the battlefield under `owner`.
func soCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func soPlayer(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// soCastFling casts Fling (or Kazuul's Fury) at `target`, paying with
// `fodder`, and returns the spell's ID.
func soCastFling(t *testing.T, g *game.Game, name, cost, oracle string, target uuid.UUID, fodder ...uuid.UUID) uuid.UUID {
	t.Helper()
	id, err := castWithTapParams(t, g, name, "Instant", cost, oracle,
		game.CastSpellParams{Targets: soPlayer(target), SacrificeIDs: fodder})
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// The record names the sacrificed object, and the damage is its power
// as it last existed: a printed 2/2 with a +1/+1 counter under Glorious
// Anthem deals 4. The anthem leaving in response changes nothing — the
// creature was gone before anyone could respond (the 2019-10-04 ruling),
// so its last-known power is the one it had as the cost was paid.
func TestFlingReadsTheSacrificedCreaturesLastKnownPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	anthem := pushGloriousAnthemFor(g, me.ID)
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	if got := effectivePower(t, g, bear); got != 3 {
		t.Fatalf("setup: the anthem makes the bear's layered power %d, want 3 (4 with its counter)", got)
	}
	epoch := 0
	if c, ok := battlefieldCard(g, bear); ok {
		epoch = c.ObjectEpoch
	}
	life := lifeOf(g, opp.ID)
	id := soCastFling(t, g, "Fling", "{1}{R}", flingOracle, opp.ID, bear)
	paid := g.StackMeta[id].Paid
	if paid.Sacrificed != 1 || len(paid.SacrificedObjects) != 1 ||
		paid.SacrificedObjects[0] != (game.ObjectRef{ID: bear, Epoch: epoch}) {
		t.Fatalf("Paid = %+v, want one sacrifice naming the bear at epoch %d", paid, epoch)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(anthem); err != nil {
			t.Fatalf("destroy the anthem: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent's life %d → %d, want 4 damage", life, got)
	}
}

// CR 107.1b: a creature with negative power deals no damage.
func TestFlingWithANegativePowerCreatureDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	weak := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Withered", TypeLine: "Creature — Beast",
		Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterMinusOne: 2},
	})
	if c, ok := battlefieldCard(g, weak); !ok || c.PowerForComparison() != -1 {
		t.Fatal("setup: the creature's power is not -1")
	}
	life := lifeOf(g, opp.ID)
	soCastFling(t, g, "Fling", "{1}{R}", flingOracle, opp.ID, weak)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life {
		t.Errorf("opponent's life %d → %d, want no damage", life, got)
	}
}

// Kazuul's Fury is Fling's front face at {2}{R}: the same read.
func TestKazuulsFuryFlingsTheSacrificedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := soCreature(g, me.ID, "Ogre", 5, 4)
	life := lifeOf(g, opp.ID)
	soCastFling(t, g, "Kazuul's Fury", "{2}{R}", kazuulsFuryOracle, opp.ID, ogre)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-5 {
		t.Errorf("opponent's life %d → %d, want 5 damage", life, got)
	}
}

// Momentous Fall reads power AND toughness off the same record, the
// counters included: a 3/4 with a +1/+1 counter draws four, then gains
// five.
func TestMomentousFallReadsPowerAndToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Beast", TypeLine: "Creature — Beast",
		Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	if _, err := castWithTapParams(t, g, "Momentous Fall", "Instant", "{2}{G}{G}", momentousFallOracle,
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{beast}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand, life := me.Hand.Size(), lifeOf(g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4", got)
	}
	if got := lifeOf(g, me.ID) - life; got != 5 {
		t.Errorf("gained %d, want 5", got)
	}
}

// Tend the Pests makes one Pest per point of the sacrificed creature's
// power.
func TestTendThePestsMakesAPestPerPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := soCreature(g, me.ID, "Beast", 3, 3)
	if _, err := castWithTapParams(t, g, "Tend the Pests", "Instant", "{B}{G}", tendThePestsOracle,
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{beast}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Pest"); n != 3 {
		t.Errorf("%d Pests, want 3", n)
	}
}

// soCopyTheSpell copies spell `id` for `controller` (CR 707.10) and
// returns the copy's stack item.
func soCopyTheSpell(t *testing.T, g *game.Game, id, controller uuid.UUID) *game.StackItem {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for k := range g.StackMeta {
		before[k] = true
	}
	if err := g.CopySpellForEffect(id, controller, false, nil); err != nil {
		t.Fatalf("copy: %v", err)
	}
	for k, item := range g.StackMeta {
		if !before[k] && item.IsCopy {
			return item
		}
	}
	t.Fatal("no copy on the stack")
	return nil
}

// CR 707.10 and the 2021-04-16 ruling: a copy of Tend the Pests uses
// the creature sacrificed for the ORIGINAL, so the two make 3 + 3.
func TestACopiedTendThePestsUsesTheOriginalsSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := soCreature(g, me.ID, "Beast", 3, 3)
	id, err := castWithTapParams(t, g, "Tend the Pests", "Instant", "{B}{G}", tendThePestsOracle,
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{beast}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	cp := soCopyTheSpell(t, g, id, me.ID)
	if cp.Paid.Sacrificed != 1 || len(cp.Paid.SacrificedObjects) != 1 || cp.Paid.SacrificedObjects[0].ID != beast {
		t.Fatalf("copy's Paid = %+v, want the original's sacrifice", cp.Paid)
	}
	if len(cp.Paid.Mana) != 0 {
		t.Error("the copy carries mana; nothing was spent to cast it")
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Pest"); n != 6 {
		t.Errorf("%d Pests, want 6 (3 from the original, 3 from the copy)", n)
	}
}

// The count travels with the copy now: a copied Vicious Betrayal over
// two sacrificed creatures pumps +4/+4 twice, where it used to read 0.
func TestACopiedViciousBetrayalReadsTheOriginalsCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := vsCreatures(g, me.ID, 2, "Creature — Goblin")
	hero := vsPermanent(g, me.ID, "Hero", "Creature — Human")
	id, err := castWithTapParams(t, g, "Vicious Betrayal", "Sorcery", "{3}{B}{B}", viciousBetrayalOracle,
		game.CastSpellParams{Targets: vsTarget(hero), SacrificeIDs: fodder})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if cp := soCopyTheSpell(t, g, id, me.ID); cp.Paid.Sacrificed != 2 {
		t.Fatalf("copy's Sacrificed = %d, want 2", cp.Paid.Sacrificed)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCardByID(g, hero)
	if c.CurrentPower() != 10 || c.CurrentToughness() != 10 {
		t.Errorf("Hero is %d/%d, want 10/10 (2/2, +4/+4 twice)", c.CurrentPower(), c.CurrentToughness())
	}
}

// soCobble casts Corpse Cobble from hand over `fodder` and resolves it.
func soCobble(t *testing.T, g *game.Game, fodder ...uuid.UUID) uuid.UUID {
	t.Helper()
	id, err := castWithTapParams(t, g, "Corpse Cobble", "Instant", "{U}{B}", corpseCobbleOracle,
		game.CastSpellParams{SacrificeIDs: fodder})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// soZombie is the one Zombie token `controller` has on the battlefield.
func soZombie(t *testing.T, g *game.Game, controller uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && IsToken(c) && c.Name == "Zombie" {
			return c
		}
	}
	t.Fatal("no Zombie token")
	return game.Card{}
}

func soTokensCreated(g *game.Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTokenCreated {
			n++
		}
	}
	return n
}

// Corpse Cobble over zero creatures makes a 0/0 Zombie, which dies to
// the state-based actions (the 2021-09-24 ruling).
func TestCorpseCobbleWithNothingMakesAZombieThatDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := soTokensCreated(g)
	soCobble(t, g)
	if soTokensCreated(g) == before {
		t.Fatal("no token was created")
	}
	if n, _ := countTokensNamed(g, me.ID, "Zombie"); n != 0 {
		t.Errorf("%d Zombies on the battlefield, want the 0/0 gone", n)
	}
}

// One creature: an X/X equal to its power, blue and black, with menace.
func TestCorpseCobbleWithOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	soCobble(t, g, soCreature(g, me.ID, "Beast", 3, 1))
	z := soZombie(t, g, me.ID)
	if z.CurrentPower() != 3 || z.CurrentToughness() != 3 {
		t.Errorf("Zombie is %d/%d, want 3/3", z.CurrentPower(), z.CurrentToughness())
	}
	if len(z.Colors) != 2 || z.Colors[0] != "U" || z.Colors[1] != "B" {
		t.Errorf("Zombie colours %v, want blue and black", z.Colors)
	}
	if !game.HasKeyword(&z, "menace") {
		t.Error("the Zombie has no menace")
	}
}

// Three creatures, one of them at -2 power: the total is summed with
// the negative included (CR 107.1b floors only the result) — 4 + 3 - 2.
func TestCorpseCobbleSumsTheTotalPowerNegativesIncluded(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shrunk := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shrunk", TypeLine: "Creature — Beast",
		Power: 1, Toughness: 5, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterMinusOne: 3},
	})
	soCobble(t, g, soCreature(g, me.ID, "A", 4, 4), soCreature(g, me.ID, "B", 3, 3), shrunk)
	z := soZombie(t, g, me.ID)
	if z.CurrentPower() != 5 || z.CurrentToughness() != 5 {
		t.Errorf("Zombie is %d/%d, want 5/5", z.CurrentPower(), z.CurrentToughness())
	}
}

// A negative TOTAL is zero (CR 107.1b): the Zombie is a 0/0.
func TestCorpseCobbleFloorsANegativeTotalAtZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shrunk := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shrunk", TypeLine: "Creature — Beast",
		Power: 0, Toughness: 5, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterMinusOne: 2},
	})
	before := soTokensCreated(g)
	soCobble(t, g, shrunk)
	if soTokensCreated(g) == before {
		t.Fatal("no token was created")
	}
	if n, _ := countTokensNamed(g, me.ID, "Zombie"); n != 0 {
		t.Errorf("%d Zombies on the battlefield, want the 0/0 gone", n)
	}
}

// Cast with flashback, the additional cost is still paid (the
// 2021-09-24 ruling) and still read; the card is exiled.
func TestCorpseCobbleFlashbackStillSacrificesAndReads(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := soCobble(t, g)
	if !me.Graveyard.Contains(id) {
		t.Fatal("Corpse Cobble is not in the graveyard")
	}
	beast := soCreature(g, me.ID, "Beast", 2, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", SacrificeIDs: []uuid.UUID{beast},
	}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	if onBattlefield(g, beast) {
		t.Fatal("the flashback cast did not sacrifice the beast")
	}
	passPriorityAroundTable(t, g)
	z := soZombie(t, g, me.ID)
	if z.CurrentPower() != 2 {
		t.Errorf("Zombie power %d, want 2", z.CurrentPower())
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back Corpse Cobble is not exiled")
	}
}

// A sacrificed token is read: it ceased to exist (CR 704.5d), but its
// record lasts the turn.
func TestFlingReadsASacrificedToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tok := pushToken(g, me.ID, game.Card{
		Name: "Beast", TypeLine: "Token Creature — Beast", Power: 3, Toughness: 3, PrintedPTKnown: true,
	})
	life := lifeOf(g, opp.ID)
	soCastFling(t, g, "Fling", "{1}{R}", flingOracle, opp.ID, tok)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent's life %d → %d, want 3 damage", life, got)
	}
}

// A sacrificed commander is read, whichever zone its owner then sends
// it to: it is sacrificed into the graveyard as the cost is paid, and
// CR 903.9a (ADR 0115) offers the command zone afterwards; here, yes.
func TestFlingReadsASacrificedCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmd := pushBattlefieldCardWithTimestamp(g, costCommander(me.ID, "My Commander"))
	life := lifeOf(g, opp.ID)
	if _, err := castWithTapParams(t, g, "Fling", "Instant", "{1}{R}", flingOracle,
		game.CastSpellParams{Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !me.Graveyard.Contains(cmd) {
		t.Fatal("the sacrificed commander is not in its owner's graveyard")
	}
	answerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(cmd) {
		t.Fatal("the commander did not go to the command zone")
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent's life %d → %d, want 3 damage", life, got)
	}
}

// The activated path writes the same record: Greater Good's item names
// the creature, and the draw reads it.
func TestGreaterGoodReadsTheRecord(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gg := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Greater Good", TypeLine: "Enchantment",
		OracleID: greaterGoodOracle, Owner: me.ID, Controller: me.ID,
	})
	beast := soCreature(g, me.ID, "Beast", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, gg, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{beast}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	item := soOnlyAbilityOnStack(t, g)
	if len(item.Paid.SacrificedObjects) != 1 || item.Paid.SacrificedObjects[0].ID != beast {
		t.Fatalf("Paid = %+v, want the beast named", item.Paid)
	}
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

// Jarad's "the sacrificed creature's power" reads the record too.
func TestJaradReadsTheRecord(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jarad := b12Push(g, me.ID, "Jarad, Golgari Lich Lord", "Legendary Creature — Zombie Elf", b17JaradOracle, 2, 2)
	beast := soCreature(g, me.ID, "Beast", 3, 3)
	advanceToMain(t, g)
	b06AddMana(me, "C", "B", "G")
	before := b17Life(g)
	if err := g.ActivateCatalogAbility(me.ID, jarad, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{beast}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if item := soOnlyAbilityOnStack(t, g); len(item.Paid.SacrificedObjects) != 1 || item.Paid.SacrificedObjects[0].ID != beast {
		t.Fatalf("Paid = %+v, want the beast named", item.Paid)
	}
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if i != 0 && p.Life != before[i]-3 {
			t.Errorf("seat %d life %d, want %d", i, p.Life, before[i]-3)
		}
	}
}

// A restore point written before SacrificedObjects existed carries the
// count and no list. The log-scan fallback still finds the permanent.
func TestAnItemWithACountAndNoListFallsBackToTheLog(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gg := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Greater Good", TypeLine: "Enchantment",
		OracleID: greaterGoodOracle, Owner: me.ID, Controller: me.ID,
	})
	beast := soCreature(g, me.ID, "Beast", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, gg, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{beast}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	item := soOnlyAbilityOnStack(t, g)
	item.Paid.SacrificedObjects = nil
	if item.Paid.Sacrificed != 1 {
		t.Fatalf("Sacrificed = %d, want 1", item.Paid.Sacrificed)
	}
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

// soOnlyAbilityOnStack is the one activated ability on the stack.
func soOnlyAbilityOnStack(t *testing.T, g *game.Game) *game.StackItem {
	t.Helper()
	var found *game.StackItem
	for _, item := range g.StackMeta {
		if item.Kind == game.StackItemActivated {
			if found != nil {
				t.Fatal("more than one activated ability on the stack")
			}
			found = item
		}
	}
	if found == nil {
		t.Fatal("no activated ability on the stack")
	}
	return found
}

// A restore point with Fling on the stack keeps the record, and the
// restored game reads the sacrificed creature's last-known power.
func TestFlingOnTheStackSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := soCreature(g, me.ID, "Beast", 4, 4)
	id := soCastFling(t, g, "Fling", "{1}{R}", flingOracle, opp.ID, beast)
	restored := restoreThroughJSON(t, g)
	item := restored.StackMeta[id]
	if item == nil || len(item.Paid.SacrificedObjects) != 1 || item.Paid.SacrificedObjects[0].ID != beast {
		t.Fatalf("restored Paid = %+v, want the beast named", item)
	}
	life := lifeOf(restored, opp.ID)
	passPriorityAroundTable(t, restored)
	if got := lifeOf(restored, opp.ID); got != life-4 {
		t.Errorf("opponent's life %d → %d after the restore, want 4 damage", life, got)
	}
}
