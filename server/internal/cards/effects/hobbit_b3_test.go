package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB / HOC) cards, half B, batch 3. Every test drives
// the registered catalog row through the engine.

// hobB3Cast puts the named spell in `who`'s hand, floats `mana`, and casts it
// under the strict mana gate, so a cast that costs more than the pool holds
// is refused. It returns the spell's id and the cast's error.
func hobB3Cast(t *testing.T, g *game.Game, who *game.Player, name, typeLine, cost, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	id := handCardFull(who, name, typeLine, cost, hobOracle(t, name), nil)
	if mana != "" {
		if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	params.Strict = true
	return id, g.CastSpell(who.ID, id, params)
}

// hobB3Price is what the named spell costs to cast with `targets` chosen.
func hobB3Price(t *testing.T, g *game.Game, who *game.Player, name, typeLine, cost string, targets ...uuid.UUID) int {
	t.Helper()
	var refs []game.TargetRef
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return selfPriced(t, g, who, hobOracle(t, name), typeLine, cost, game.ZoneHand, refs).ManaValue()
}

// hobB3Activate floats `mana` and activates ability 0 of `id` under the
// strict mana gate.
func hobB3Activate(t *testing.T, g *game.Game, who, id uuid.UUID, mana string) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	return g.ActivateCatalogAbility(who, id, 0, game.ActivateAbilityParams{Strict: true})
}

func hobB3Perm(g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	return pushPermanent(g, owner, c)
}

// --- Magnificent End --------------------------------------------------

func TestHobMagnificentEndCostsThreeLessAtATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := hobB3Perm(g, opp.ID, game.Card{Name: "Tired", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	fresh := hobB3Perm(g, opp.ID, game.Card{Name: "Fresh", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if got := hobB3Price(t, g, me, "Magnificent End", "Instant", "{4}{W}", tapped); got != 2 {
		t.Errorf("at a tapped creature: %d, want 2", got)
	}
	if got := hobB3Price(t, g, me, "Magnificent End", "Instant", "{4}{W}", fresh); got != 5 {
		t.Errorf("at an untapped creature: %d, want 5", got)
	}
	if got := hobB3Price(t, g, me, "Magnificent End", "Instant", "{4}{W}"); got != 5 {
		t.Errorf("with no target: %d, want 5", got)
	}
}

func TestHobMagnificentEndKillsATappedCreatureForTwoMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	big := hobB3Perm(g, opp.ID, game.Card{Name: "Tired Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5, Tapped: true})
	if _, err := hobB3Cast(t, g, me, "Magnificent End", "Instant", "{4}{W}", "{C}{W}", game.CastSpellParams{Targets: b16TargetCard(big)}); err != nil {
		t.Fatalf("cast for {1}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !opp.Graveyard.Contains(big) {
		t.Error("five damage should kill a 5/5")
	}
}

func TestHobMagnificentEndAtAnUntappedCreatureCostsFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	big := hobB3Perm(g, opp.ID, game.Card{Name: "Fresh Giant", TypeLine: "Creature — Giant", Power: 6, Toughness: 6})
	if _, err := hobB3Cast(t, g, me, "Magnificent End", "Instant", "{4}{W}", "{C}{W}", game.CastSpellParams{Targets: b16TargetCard(big)}); err == nil {
		t.Fatal("cast at an untapped creature for {1}{W}")
	}
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	toMainForCost(t, g2)
	big2 := hobB3Perm(g2, opp2.ID, game.Card{Name: "Fresh Giant", TypeLine: "Creature — Giant", Power: 6, Toughness: 6})
	if _, err := hobB3Cast(t, g2, me2, "Magnificent End", "Instant", "{4}{W}", "{C}{C}{C}{C}{W}", game.CastSpellParams{Targets: b16TargetCard(big2)}); err != nil {
		t.Fatalf("full price: %v", err)
	}
	passPriorityAroundTable(t, g2)
	c := cardByID(g2, big2)
	if !g2.Battlefield.Contains(big2) || c.DamageMarked != 5 {
		t.Errorf("a 6/6 should survive with 5 damage marked, has %d", c.DamageMarked)
	}
}

// --- Bitter Downfall --------------------------------------------------

func TestHobBitterDownfallCostsThreeLessAtADamagedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := hobB3Perm(g, opp.ID, game.Card{Name: "Hurt", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	fresh := hobB3Perm(g, opp.ID, game.Card{Name: "Fresh", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := hobB3Price(t, g, me, "Bitter Downfall", "Instant", "{3}{B}", hurt); got != 1 {
		t.Errorf("at a damaged creature: %d, want 1", got)
	}
	if got := hobB3Price(t, g, me, "Bitter Downfall", "Instant", "{3}{B}", fresh); got != 4 {
		t.Errorf("at an undamaged creature: %d, want 4", got)
	}
}

func TestHobBitterDownfallDestroysAndItsControllerLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim := hobB3Perm(g, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	life, mine := opp.Life, me.Life
	if _, err := hobB3Cast(t, g, me, "Bitter Downfall", "Instant", "{3}{B}", "{C}{C}{C}{B}", game.CastSpellParams{Targets: b16TargetCard(victim)}); err != nil {
		t.Fatalf("full price: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
		t.Error("the creature should be destroyed")
	}
	if opp.Life != life-2 {
		t.Errorf("its controller's life %d -> %d, want -2", life, opp.Life)
	}
	if me.Life != mine {
		t.Errorf("the caster's life changed: %d -> %d", mine, me.Life)
	}
}

func TestHobBitterDownfallIsOneManaAfterDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	hurt := hobB3Perm(g, opp.ID, game.Card{Name: "Hurt", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	life := opp.Life
	if _, err := hobB3Cast(t, g, me, "Bitter Downfall", "Instant", "{3}{B}", "{B}", game.CastSpellParams{Targets: b16TargetCard(hurt)}); err != nil {
		t.Fatalf("cast for {B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) || opp.Life != life-2 {
		t.Errorf("on battlefield %v, life %d -> %d; want it destroyed and -2", g.Battlefield.Contains(hurt), life, opp.Life)
	}
}

// --- Stern Scolding ---------------------------------------------------

// hobB3SternScold casts a creature spell with the given numbers (it stays
// on the stack) and answers it with Stern Scolding. It returns the creature
// spell's id and the Scolding's cast error.
func hobB3SternScold(t *testing.T, g *game.Game, typeLine string, power, toughness int) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[0]
	toMainForCost(t, g)
	spell := rfdHandCard(me, "Target Creature", typeLine, "", "", power, toughness)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature spell: %v", err)
	}
	_, err := hobB3Cast(t, g, me, "Stern Scolding", "Instant", "{U}", "{U}", game.CastSpellParams{Targets: b16TargetCard(spell)})
	return spell, err
}

func TestHobSternScoldingCountersALowPowerCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	spell, err := hobB3SternScold(t, g, "Creature — Bear", 2, 5)
	if err != nil {
		t.Fatalf("a 2/5 spell is a legal target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spell) || !g.Seats[0].Graveyard.Contains(spell) {
		t.Error("the 2/5 creature spell should be countered")
	}
}

func TestHobSternScoldingCountersALowToughnessCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	spell, err := hobB3SternScold(t, g, "Creature — Bear", 6, 2)
	if err != nil {
		t.Fatalf("a 6/2 spell is a legal target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spell) || !g.Seats[0].Graveyard.Contains(spell) {
		t.Error("the 6/2 creature spell should be countered")
	}
}

func TestHobSternScoldingRefusesABigCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := hobB3SternScold(t, g, "Creature — Bear", 3, 3); err == nil {
		t.Fatal("Stern Scolding accepted a 3/3 creature spell")
	}
}

func TestHobSternScoldingRefusesANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	if _, err := hobB3Cast(t, g, me, "Stern Scolding", "Instant", "{U}", "{U}", game.CastSpellParams{Targets: b16TargetCard(bolt)}); err == nil {
		t.Fatal("Stern Scolding accepted an instant")
	}
}

// --- Uneasy Partings --------------------------------------------------

// hobB3Attack has seat 1 attack seat 2 with the given creatures and stops at
// the declare attackers step, where seat 0 holds an instant.
func hobB3Attack(t *testing.T, g *game.Game, attackers ...uuid.UUID) {
	t.Helper()
	advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, g.Seats[2].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
}

func TestHobUneasyPartingsCostsOneLessAtAnAttackingNontokenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	real := hobB3Perm(g, opp.ID, game.Card{Name: "Raider", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	token := hobB3Perm(g, opp.ID, game.Card{Name: "Raider Token", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1})
	home := hobB3Perm(g, opp.ID, game.Card{Name: "Homebody", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	hobB3Attack(t, g, real, token)
	if got := hobB3Price(t, g, me, "Uneasy Partings", "Instant", "{3}{U}", real); got != 3 {
		t.Errorf("at an attacking nontoken creature: %d, want 3", got)
	}
	if got := hobB3Price(t, g, me, "Uneasy Partings", "Instant", "{3}{U}", token); got != 4 {
		t.Errorf("at an attacking token: %d, want 4", got)
	}
	if got := hobB3Price(t, g, me, "Uneasy Partings", "Instant", "{3}{U}", home); got != 4 {
		t.Errorf("at a creature that is not attacking: %d, want 4", got)
	}
}

func TestHobUneasyPartingsLetsTheOwnerChooseTopOrBottom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := hobB3Perm(g, opp.ID, game.Card{Name: "Raider", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	hobB3Attack(t, g, raider)
	if _, err := hobB3Cast(t, g, me, "Uneasy Partings", "Instant", "{3}{U}", "{C}{C}{U}", game.CastSpellParams{Targets: b16TargetCard(raider)}); err != nil {
		t.Fatalf("cast for {2}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := putInLibraryChoiceFor(g, opp.ID)
	if c == nil || c.LibraryPlacement != game.LibraryPlaceTopOrBottom {
		t.Fatalf("the owner was not asked top or bottom: %+v", c)
	}
	if putInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("the caster must not choose")
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{raider}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(raider) {
		t.Error("the creature is still on the battlefield")
	}
	if got := opp.Library.Cards[0].InstanceID; got != raider {
		t.Error("the creature should be on the bottom of its owner's library")
	}
}

func TestHobUneasyPartingsCanPutItOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := hobB3Perm(g, opp.ID, game.Card{Name: "Raider", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if _, err := hobB3Cast(t, g, me, "Uneasy Partings", "Instant", "{3}{U}", "{C}{C}{C}{U}", game.CastSpellParams{Targets: b16TargetCard(raider)}); err != nil {
		t.Fatalf("full price: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := putInLibraryChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the owner was not asked")
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, nil, []uuid.UUID{raider}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; top != raider {
		t.Error("the creature should be on top of its owner's library")
	}
}

// --- Ravenhill Flock --------------------------------------------------

func TestHobRavenhillFlockFliesAndGrowsOnEachDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	flock := hobPush(t, g, me.ID, "Ravenhill Flock", "Creature — Bird", 1, 2)
	if !hasEffectiveKeyword(t, g, flock, "flying") {
		t.Error("Ravenhill Flock lacks flying")
	}
	advanceToMain(t, g)
	for i := 0; i < 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if c := gaLayered(t, g, flock); c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Errorf("after two draws it is %d/%d, want 3/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobRavenhillFlockIgnoresAnOpponentsDraw(t *testing.T) {
	g := newCatalogGame(t)
	flock := hobPush(t, g, g.Seats[0].ID, "Ravenhill Flock", "Creature — Bird", 1, 2)
	advanceToMain(t, g)
	if err := g.DrawCard(g.Seats[1].ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, flock); c.CurrentPower() != 1 || c.CurrentToughness() != 2 {
		t.Errorf("it is %d/%d after an opponent's draw, want 1/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Attercop ---------------------------------------------------------

func TestHobAttercopHasReachAndDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	id := hobPush(t, g, g.Seats[0].ID, "Attercop", "Creature — Spider", 2, 1)
	for _, kw := range []string{"reach", "deathtouch"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Attercop lacks %s", kw)
		}
	}
}

func TestHobAttercopGrowsWhenALandEntersUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := hobPush(t, g, me.ID, "Attercop", "Creature — Spider", 2, 1)
	toMainForCost(t, g)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("after a land it is %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	aangAdvanceToMain(t, g, 1)
	if c := gaLayered(t, g, id); c.CurrentPower() != 2 || c.CurrentToughness() != 1 {
		t.Errorf("next turn it is %d/%d, want the boost gone (2/1)", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Boughside Wanderers ----------------------------------------------

func TestHobBoughsideWanderersTakesAPermanentCardAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	top := hornTopFive(me, []struct{ name, typeLine string }{
		{"Shock", "Instant"},
		{"Divination", "Sorcery"},
		{"Sol Ring", "Artifact"},
		{"Forest", "Basic Land — Forest"},
	})
	hobCast(t, g, "Boughside Wanderers", "Creature — Elf Scout", 4, 4)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt for the look at the top four")
	}
	want := map[uuid.UUID]bool{top[2]: true, top[3]: true}
	if len(prompt.ChooseCards) != len(want) {
		t.Fatalf("offered %d cards, want the 2 permanent cards", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered a card that is not a permanent card: %s", id)
		}
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds = [%d,%d], want [0,1]", prompt.ChooseMin, prompt.ChooseMax)
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[3]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[3]) || me.Hand.Size() != hand+1 {
		t.Error("the chosen land should be the one card added to hand")
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, id := range top[:3] {
		if !bottom[id] {
			t.Errorf("a card that was not taken is not on the bottom: %s", id)
		}
	}
}

func TestHobBoughsideWanderersMayTakeNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hornTopFive(me, []struct{ name, typeLine string }{
		{"Shock", "Instant"}, {"Divination", "Sorcery"}, {"Sol Ring", "Artifact"}, {"Forest", "Basic Land — Forest"},
	})
	hobCast(t, g, "Boughside Wanderers", "Creature — Elf Scout", 4, 4)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if me.Hand.Size() != hand {
		t.Error("declining should add nothing to hand")
	}
}

func TestHobBoughsideWanderersGetsTwoTwoPerLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := hobPush(t, g, me.ID, "Boughside Wanderers", "Creature — Elf Scout", 4, 4)
	toMainForCost(t, g)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, id); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("after a land it is %d/%d, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Desolation Prowler -----------------------------------------------

func TestHobDesolationProwlerPumpsForTwoLifeOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := hobPush(t, g, me.ID, "Desolation Prowler", "Creature — Wolf", 2, 2)
	life := me.Life
	if err := hobB3Activate(t, g, me.ID, id, ""); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
	if c := gaLayered(t, g, id); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("it is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if err := hobB3Activate(t, g, me.ID, id, ""); err == nil {
		t.Fatal("activated a second time in one turn")
	}
	if me.Life != life-2 {
		t.Errorf("the refused activation cost life: %d", me.Life)
	}
	// A new turn of its controller's: it may be activated again.
	aangAdvanceToMain(t, g, 1)
	aangAdvanceToMain(t, g, 0)
	if err := hobB3Activate(t, g, me.ID, id, ""); err != nil {
		t.Fatalf("activation on the next turn: %v", err)
	}
}

func TestHobDesolationProwlerCantPayLifeYouDontHave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := hobPush(t, g, me.ID, "Desolation Prowler", "Creature — Wolf", 2, 2)
	me.Life = 1
	if err := hobB3Activate(t, g, me.ID, id, ""); err == nil {
		t.Fatal("paid 2 life at 1 life")
	}
	if me.Life != 1 {
		t.Errorf("life = %d, want it untouched", me.Life)
	}
}

// --- Olog-hai Crusher -------------------------------------------------

// hobB3CrusherBlocks has the active seat attack seat 1 with a bear, and
// returns the block attempt of seat 1's Crusher, plus the Crusher's id.
func hobB3Block(t *testing.T, g *game.Game, crusher, bear uuid.UUID) error {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	return g.DeclareBlocker(crusher, bear)
}

func TestHobOlogHaiCrusherHasTrample(t *testing.T) {
	g := newCatalogGame(t)
	id := hobPush(t, g, g.Seats[0].ID, "Olog-hai Crusher", "Creature — Troll Soldier", 4, 4)
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Olog-hai Crusher lacks trample")
	}
}

func TestHobOlogHaiCrusherCantBlockWithoutAGoblinOrOrc(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, g.Seats[0].ID, "Bear", 2, 2)
	crusher := hobPush(t, g, opp.ID, "Olog-hai Crusher", "Creature — Troll Soldier", 4, 4)
	brRefusal(t, hobB3Block(t, g, crusher, bear), game.BlockReasonCantBlockAttacker)
}

func TestHobOlogHaiCrusherBlocksWithAGoblin(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, g.Seats[0].ID, "Bear", 2, 2)
	crusher := hobPush(t, g, opp.ID, "Olog-hai Crusher", "Creature — Troll Soldier", 4, 4)
	b12Creature(g, opp.ID, "Goblin Friend", "Creature — Goblin Warrior", 1, 1)
	if err := hobB3Block(t, g, crusher, bear); err != nil {
		t.Fatalf("block with a Goblin in play: %v", err)
	}
}

func TestHobOlogHaiCrusherBlocksWithAnOrc(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, g.Seats[0].ID, "Bear", 2, 2)
	crusher := hobPush(t, g, opp.ID, "Olog-hai Crusher", "Creature — Troll Soldier", 4, 4)
	b12Creature(g, opp.ID, "Orc Friend", "Creature — Orc Soldier", 1, 1)
	if err := hobB3Block(t, g, crusher, bear); err != nil {
		t.Fatalf("block with an Orc in play: %v", err)
	}
}

func TestHobOlogHaiCrusherIgnoresAnOpponentsGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	b12Creature(g, me.ID, "Their Goblin", "Creature — Goblin Warrior", 1, 1)
	crusher := hobPush(t, g, opp.ID, "Olog-hai Crusher", "Creature — Troll Soldier", 4, 4)
	brRefusal(t, hobB3Block(t, g, crusher, bear), game.BlockReasonCantBlockAttacker)
}

// --- Goblin Cratermaker -----------------------------------------------

func TestHobGoblinCratermakerDealsTwoToACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maker := hobPush(t, g, me.ID, "Goblin Cratermaker", "Creature — Goblin Warrior", 2, 2)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	toMainForCost(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, maker, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(victim)}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if g.Battlefield.Contains(maker) {
		t.Error("the Cratermaker should be sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("two damage should kill a 2/2")
	}
}

func TestHobGoblinCratermakerDestroysAColorlessNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maker := hobPush(t, g, me.ID, "Goblin Cratermaker", "Creature — Goblin Warrior", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	toMainForCost(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, maker, 1, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(rock)}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(maker) {
		t.Error("the artifact is destroyed and the Cratermaker sacrificed")
	}
}

func TestHobGoblinCratermakerRefusesColoredAndLandTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maker := hobPush(t, g, me.ID, "Goblin Cratermaker", "Creature — Goblin Warrior", 2, 2)
	green := acPermanent(g, opp.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	toMainForCost(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	for _, target := range []uuid.UUID{green, land} {
		if err := g.ActivateCatalogAbility(me.ID, maker, 1, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(target)}); err == nil {
			t.Fatal("destroy mode accepted a colored creature or a land")
		}
	}
	if !g.Battlefield.Contains(maker) {
		t.Error("a refused activation must not sacrifice the Cratermaker")
	}
}

// --- Improvised Club --------------------------------------------------

func TestHobImprovisedClubSacrificesAndDealsFourToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	life := opp.Life
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if _, err := hobB3Cast(t, g, me, "Improvised Club", "Instant", "{1}{R}", "{C}{R}", game.CastSpellParams{Targets: target, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature should be gone")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life %d -> %d, want -4", life, opp.Life)
	}
}

func TestHobImprovisedClubCanSacrificeAnArtifactAndHitACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	victim := pushVanillaCreature(g, opp.ID, "Victim", 4, 4)
	if _, err := hobB3Cast(t, g, me, "Improvised Club", "Instant", "{1}{R}", "{C}{R}", game.CastSpellParams{Targets: b16TargetCard(victim), SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(victim) {
		t.Error("the artifact is sacrificed and the 4/4 dies")
	}
}

func TestHobImprovisedClubNeedsASacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if _, err := hobB3Cast(t, g, me, "Improvised Club", "Instant", "{1}{R}", "{C}{R}", game.CastSpellParams{Targets: target}); err == nil {
		t.Fatal("cast without sacrificing anything")
	}
}

func TestHobImprovisedClubRefusesALandAsFodder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	land := b12Permanent(g, me.ID, "My Land", "Land")
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if _, err := hobB3Cast(t, g, me, "Improvised Club", "Instant", "{1}{R}", "{C}{R}", game.CastSpellParams{Targets: target, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("sacrificed a land")
	}
}
