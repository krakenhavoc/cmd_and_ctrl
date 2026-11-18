package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	afrB18DragonsFire   = "7237d607-2613-4786-96d6-55792c2e02b9"
	afrB18Disciple      = "3e5d3574-cc92-434a-a210-fcda906391c8"
	afrB18BlueDragon    = "1d00e54a-5897-43dc-9c38-a0e2ce6a7433"
	afrB18GreenDragon   = "c2a4c6c7-e5b4-4cc8-91d8-5de8cc95ed00"
	afrB18Iymrith       = "d5efba91-8400-41a3-91c6-cb5e81f07976"
	afrB18Inferno       = "cd54438d-7f9d-4d80-b15b-0f0bf4becc15"
	afrB18Regent        = "2df30a75-4957-4eab-8b0b-4c973f00642d"
	afrB18Tarrasque     = "bfd66214-e576-4a7e-8559-4e4e51d56b98"
	afrB18Storvald      = "5d7802a7-df6e-47fb-9065-c7461ea24a19"
	afrB18Klauth        = "8cefa5d6-5151-4e75-8b72-f210a4c751ea"
	afrB18KlauthsWill   = "639926c1-c4b4-4a49-b125-0e1b584c5bb8"
	afrB18DragonType    = "Creature — Dragon"
	afrB18LegendDragon  = "Legendary Creature — Dragon"
	afrB18BeholdOfferAt = 0
)

func TestAfrB18CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		afrB18DragonsFire: "Dragon's Fire", afrB18Disciple: "Dragon's Disciple",
		afrB18BlueDragon: "Blue Dragon", afrB18GreenDragon: "Green Dragon",
		afrB18Iymrith: "Iymrith, Desert Doom", afrB18Inferno: "Inferno of the Star Mounts",
		afrB18Regent: "Westgate Regent", afrB18Tarrasque: "The Tarrasque",
		afrB18Storvald: "Storvald, Frost Giant Jarl", afrB18Klauth: "Klauth, Unrivaled Ancient",
		afrB18KlauthsWill: "Klauth's Will",
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

func afrB18CastFire(t *testing.T, g *game.Game, params game.CastSpellParams) error {
	t.Helper()
	_, err := castWithTapParams(t, g, "Dragon's Fire", "Instant", "{1}{R}", afrB18DragonsFire, params)
	return err
}

// Declined, Dragon's Fire deals 3.
func TestAfrB18DragonsFireDealsThreeWithoutBehold(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	_ = me
	if err := afrB18CastFire(t, g, game.CastSpellParams{Targets: slotted(foe)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, foe); got != 3 {
		t.Errorf("foe took %d, want 3", got)
	}
}

// Choosing a Dragon you control: damage equals that Dragon's power as it
// resolves, a pump in response included.
func TestAfrB18DragonsFireUsesTheChosenDragonsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	dragon := apaPush(g, me.ID, me.ID, game.Card{Name: "Big Dragon", TypeLine: afrB18DragonType, Power: 6, Toughness: 6})
	if err := afrB18CastFire(t, g, game.CastSpellParams{
		OptionalCosts: []int{afrB18BeholdOfferAt},
		RevealIDs:     []uuid.UUID{dragon},
		Targets:       slotted(foe),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, foe); got != 6 {
		t.Errorf("foe took %d, want 6 (the Dragon's power)", got)
	}
}

// Revealing a Dragon card from hand: the card's printed power.
func TestAfrB18DragonsFireUsesTheRevealedCardsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Hand Dragon", TypeLine: afrB18DragonType, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	if err := afrB18CastFire(t, g, game.CastSpellParams{
		OptionalCosts: []int{afrB18BeholdOfferAt},
		RevealIDs:     []uuid.UUID{card},
		Targets:       slotted(foe),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !me.Hand.Contains(card) {
		t.Error("revealing moved the card out of the hand")
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, foe); got != 5 {
		t.Errorf("foe took %d, want 5", got)
	}
}

// A non-Dragon is not a legal pick.
func TestAfrB18DragonsFireRefusesANonDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := afrB18CastFire(t, g, game.CastSpellParams{
		Strict:        true,
		OptionalCosts: []int{afrB18BeholdOfferAt},
		RevealIDs:     []uuid.UUID{bear},
		Targets:       slotted(foe),
	}); err == nil {
		t.Fatal("a Bear was accepted as the Dragon to choose")
	}
}

// afrB18Ready lets a pushed creature attack and tap.
func afrB18Ready(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, id); c != nil {
			c.SummonedThisTurn = false
		}
	})
}

func afrB18CounterCount(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

// Westgate Regent puts as many +1/+1 counters on itself as the combat damage
// it dealt to a player.
func TestAfrB18RegentGrowsByTheCombatDamageDealt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	regent := b12Push(g, me.ID, "Westgate Regent", "Creature — Vampire", afrB18Regent, 4, 4)
	afrB18Ready(g, regent)
	dealCombatDamageToPlayer(g, regent, opp.ID, 4)
	passPriorityAroundTable(t, g)
	if got := afrB18CounterCount(g, regent); got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", got)
	}
	// Noncombat damage does not trigger it.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: regent, Target: opp.ID, Amount: 3})
	})
	passPriorityAroundTable(t, g)
	if got := afrB18CounterCount(g, regent); got != 4 {
		t.Errorf("noncombat damage changed the counters to %d", got)
	}
}

func TestAfrB18RegentWardIsDiscardACard(t *testing.T) {
	g, _, opp, regent := wardTable(t, "Westgate Regent", "Creature — Vampire", afrB18Regent)
	hand := fillHand(opp, "Filler A")
	boltWardedCreature(t, g, opp, regent)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(hand[0]) {
		t.Error("declining a discard ward discarded anyway")
	}
	if boltResolved(g, regent) {
		t.Error("the bolt resolved despite the declined ward")
	}
}

// Iymrith: ward {4} only while untapped.
func TestAfrB18IymrithHasWardFourOnlyWhileUntapped(t *testing.T) {
	g, _, opp, dragon := wardTable(t, "Iymrith, Desert Doom", afrB18LegendDragon, afrB18Iymrith)
	castAtWardedCreature(t, g, opp, dragon)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("an untapped Iymrith did not ask for ward {4}")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, dragon); c != nil {
			c.Tapped = true
		}
		g.RecomputeLayersIfStaleLocked()
	})
	castAtWardedCreature(t, g, opp, dragon)
	if hasPayUnlessFor(g, opp.ID) {
		t.Fatal("a tapped Iymrith asked for ward")
	}
}

func TestAfrB18IymrithDrawsUpToThree(t *testing.T) {
	for _, tc := range []struct{ hand, want int }{{0, 3}, {1, 3}, {2, 3}, {3, 4}, {5, 6}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		for i := 0; i < 10; i++ {
			seedLibraryTop(me, "Filler", "Sorcery", "")
		}
		g.WithWriteLock(func() { me.Hand.Cards = nil })
		fillHand(me, make([]string, tc.hand)...)
		dragon := b12Push(g, me.ID, "Iymrith, Desert Doom", afrB18LegendDragon, afrB18Iymrith, 5, 5)
		afrB18Ready(g, dragon)
		dealCombatDamageToPlayer(g, dragon, opp.ID, 5)
		passPriorityAroundTable(t, g)
		if got := len(me.Hand.Cards); got != tc.want {
			t.Errorf("hand of %d: ended with %d cards, want %d", tc.hand, got, tc.want)
		}
	}
}

// Dragon's Disciple.
func TestAfrB18DiscipleEntersWithACounterOnlyWithADragon(t *testing.T) {
	// Controls a Dragon: the counter, with no reveal needed.
	g := newCatalogGame(t)
	me := g.Seats[0]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Dragon", TypeLine: afrB18DragonType, Power: 3, Toughness: 3})
	id := castHoldCreature(t, g, "Dragon's Disciple", "Creature — Human Monk", afrB18Disciple, 1, 3)
	afrB18DeclineReveal(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := afrB18CounterCount(g, id); got != 1 {
		t.Errorf("with a Dragon in play: counters = %d, want 1", got)
	}

	// No Dragon anywhere: no counter.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	id2 := castHoldCreature(t, g2, "Dragon's Disciple", "Creature — Human Monk", afrB18Disciple, 1, 3)
	afrB18DeclineReveal(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	if got := afrB18CounterCount(g2, id2); got != 0 {
		t.Errorf("with no Dragon: counters = %d, want 0", got)
	}

	// A Dragon card in hand, revealed: the counter.
	g3 := newCatalogGame(t)
	me3 := g3.Seats[0]
	card := handCard(me3, "Hand Dragon", afrB18DragonType)
	id3 := castHoldCreature(t, g3, "Dragon's Disciple", "Creature — Human Monk", afrB18Disciple, 1, 3)
	afrB18RevealCard(t, g3, me3.ID, card)
	passPriorityAroundTable(t, g3)
	if got := afrB18CounterCount(g3, id3); got != 1 {
		t.Errorf("after revealing a Dragon card: counters = %d, want 1", got)
	}

	// A Dragon card in hand, declined, and none in play: no counter.
	g4 := newCatalogGame(t)
	me4 := g4.Seats[0]
	handCard(me4, "Hand Dragon", afrB18DragonType)
	id4 := castHoldCreature(t, g4, "Dragon's Disciple", "Creature — Human Monk", afrB18Disciple, 1, 3)
	afrB18DeclineReveal(t, g4, me4.ID)
	passPriorityAroundTable(t, g4)
	if got := afrB18CounterCount(g4, id4); got != 0 {
		t.Errorf("declining the reveal with no Dragon: counters = %d, want 0", got)
	}
}

func TestAfrB18DiscipleGrantsWardOneToDragonsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Dragon's Disciple", "Creature — Human Monk", afrB18Disciple, false)
	dragon := apaPush(g, me.ID, me.ID, game.Card{Name: "Dragon", TypeLine: afrB18DragonType, Power: 3, Toughness: 3})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToMain(t, g)
	castAtWardedCreature(t, g, opp, bear)
	if hasPayUnlessFor(g, opp.ID) {
		t.Fatal("a Bear got ward")
	}
	passPriorityAroundTable(t, g)
	castAtWardedCreature(t, g, opp, dragon)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("a Dragon you control has no ward {1}")
	}
}

func afrB18DeclineReveal(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && entryRevealChoiceFor(g, who) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			return
		}
	}
	if entryRevealChoiceFor(g, who) != nil {
		answerEntryReveal(t, g, who)
	}
}

func afrB18RevealCard(t *testing.T, g *game.Game, who, card uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && entryRevealChoiceFor(g, who) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	answerEntryReveal(t, g, who, card)
}

func afrB18BasePT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return effectivePTOf(t, g, id)
}

// Storvald: both bullets, each on its own target.
func TestAfrB18StorvaldBothModesSetBasePT(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	small := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Storvald, Frost Giant Jarl", "Legendary Creature — Giant", afrB18Storvald, 7, 7)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, small)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if p, tt := afrB18BasePT(t, g, small); p != 7 || tt != 7 {
		t.Errorf("7/7 mode target is %d/%d", p, tt)
	}
	if p, tt := afrB18BasePT(t, g, big); p != 1 || tt != 1 {
		t.Errorf("1/1 mode target is %d/%d", p, tt)
	}
}

func TestAfrB18StorvaldGrantsWardThreeToOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Storvald, Frost Giant Jarl", "Legendary Creature — Giant", afrB18Storvald, false)
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToMain(t, g)
	castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("another creature you control has no ward {3}")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	// The opponent's own creature, targeted by its controller, is not warded.
	castAtWardedCreature(t, g, me, theirs)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's creature got ward from Storvald")
	}
}

// The Tarrasque: haste and ward {10} only if it was cast.
func TestAfrB18TarrasqueHasHasteAndWardOnlyIfCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cast := castHoldCreature(t, g, "The Tarrasque", "Legendary Creature — Dinosaur", afrB18Tarrasque, 10, 10)
	if !hasKeywordOnBattlefield(t, g, cast, "haste") {
		t.Error("a cast Tarrasque has no haste")
	}
	put := b12Push(g, me.ID, "The Tarrasque", "Legendary Creature — Dinosaur", afrB18Tarrasque, 10, 10)
	if hasKeywordOnBattlefield(t, g, put, "haste") {
		t.Error("a Tarrasque that was not cast has haste")
	}
}

func TestAfrB18TarrasqueWardTenAsksTheCaster(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	cast := castHoldCreature(t, g, "The Tarrasque", "Legendary Creature — Dinosaur", afrB18Tarrasque, 10, 10)
	castAtWardedCreature(t, g, opp, cast)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("a cast Tarrasque has no ward")
	}
}

func TestAfrB18TarrasqueFightsADefendersCreatureWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tar := b12Push(g, me.ID, "The Tarrasque", "Legendary Creature — Dinosaur", afrB18Tarrasque, 10, 10)
	afrB18Ready(g, tar)
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToDeclareAttackersOf(t, g, 0)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: tar, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	lockInAttacks(t, g)
	b04WaitForPickOrSkip(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, foe) {
		t.Error("the Foe survived the fight")
	}
	if got := pr6Marked(g, tar); got != 2 {
		t.Errorf("the Tarrasque took %d, want 2", got)
	}
}

func b04WaitForPickOrSkip(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if latestPickTarget(g, chooser) != nil {
		pickCard(t, g, chooser, target)
	}
}

// Blue Dragon: three clauses, -3/-0, -2/-0, -1/-0.
func TestAfrB18BlueDragonShrinksThreeCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 5, 5)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 5, 5)
	c := b12Creature(g, me.ID, "C", "Creature — Bear", 5, 5)
	castAndResolveCreature(t, g, "Blue Dragon", afrB18DragonType, afrB18BlueDragon)
	for _, id := range []uuid.UUID{a, b, c} {
		b04WaitForPick(t, g, me.ID)
		pickCard(t, g, me.ID, id)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 2, b: 3, c: 4} {
		if p, _ := afrB18BasePT(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}

// Green Dragon: an opponent's creature dealt damage this turn is destroyed.
func TestAfrB18GreenDragonDestroysDamagedOpponentCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Bear", Power: 1, Toughness: 9})
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 1, Toughness: 9})
	castAndResolveCreature(t, g, "Green Dragon", afrB18DragonType, afrB18GreenDragon)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: mine, Target: mine, Amount: 1})
	})
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, mine) {
		t.Error("damage to my own creature destroyed it")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: mine, Target: foe, Amount: 1})
	})
	passPriorityAroundTable(t, g)
	if onBattlefield(g, foe) {
		t.Error("the damaged opponent creature survived")
	}
}

// Klauth: X mana, X the total power of attackers, spell-only, kept all turn.
func TestAfrB18KlauthAddsRestrictedManaForTotalAttackingPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	k := b12Push(g, me.ID, "Klauth, Unrivaled Ancient", afrB18LegendDragon, afrB18Klauth, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	afrB18Ready(g, k)
	afrB18Ready(g, bear)
	advanceToDeclareAttackersOf(t, g, 0)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: k, Target: opp.ID}, {Attacker: bear, Target: opp.ID},
	}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	lockInAttacks(t, g)
	for i := 0; i < 6 && me.ManaPool.Total() == 0; i++ {
		if ch := latestChoiceForAfr(g, me.ID); ch {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if got := me.ManaPool.Total(); got != 7 {
		t.Errorf("pool holds %d mana, want 7", got)
	}
}

func latestChoiceForAfr(g *game.Game, who uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == who {
			return true
		}
	}
	return false
}

// Klauth's Will.
func afrB18CastWill(t *testing.T, g *game.Game, x int, modes []int, targets []game.TargetRef) error {
	t.Helper()
	_, err := castWithTapParams(t, g, "Klauth's Will", "Instant", "{X}{R}{R}{G}", afrB18KlauthsWill,
		game.CastSpellParams{XValue: x, Modes: modes, Targets: targets})
	return err
}

func TestAfrB18KlauthsWillBreatheFlameSparesFliers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ground := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Ground", TypeLine: "Creature — Bear", Power: 1, Toughness: 9})
	flier := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 9, Keywords: []string{"flying"}})
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 1, Toughness: 9})
	if err := afrB18CastWill(t, g, 3, []int{0}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if pr6Marked(g, ground) != 3 || pr6Marked(g, mine) != 3 {
		t.Errorf("ground creatures took %d and %d, want 3", pr6Marked(g, ground), pr6Marked(g, mine))
	}
	if pr6Marked(g, flier) != 0 {
		t.Errorf("the flier took %d", pr6Marked(g, flier))
	}
}

func TestAfrB18KlauthsWillSmashRelicsDestroysUpToX(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	e := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Aura", TypeLine: "Enchantment"})
	c := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Gem", TypeLine: "Artifact"})
	if err := afrB18CastWill(t, g, 2, []int{1}, slotted(a, e)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, a) || onBattlefield(g, e) {
		t.Error("a chosen relic survived")
	}
	if !onBattlefield(g, c) {
		t.Error("an unchosen relic was destroyed")
	}
}

func TestAfrB18KlauthsWillRefusesTwoModesWithoutACommander(t *testing.T) {
	g := newCatalogGame(t)
	if err := afrB18CastWill(t, g, 1, []int{0, 1}, nil); err == nil {
		t.Fatal("both modes were allowed without a commander")
	}
}

// Inferno: the 20th point of power raises the 20-damage trigger.
func TestAfrB18InfernoDealsTwentyWhenPowerBecomesTwenty(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	inf := b12Push(g, me.ID, "Inferno of the Star Mounts", afrB18LegendDragon, afrB18Inferno, 19, 6)
	afrB18Ready(g, inf)
	life := opp.Life
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, inf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if latestPickTarget(g, me.ID) != nil {
		p := latestPickTarget(g, me.ID)
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-20 {
		t.Errorf("life %d -> %d, want -20", life, opp.Life)
	}
}

func TestAfrB18InfernoBelowTwentyDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	inf := b12Push(g, me.ID, "Inferno of the Star Mounts", afrB18LegendDragon, afrB18Inferno, 6, 6)
	afrB18Ready(g, inf)
	life := opp.Life
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, inf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := afrB18BasePT(t, g, inf); p != 7 {
		t.Errorf("power %d, want 7", p)
	}
	if opp.Life != life {
		t.Errorf("life changed to %d", opp.Life)
	}
}
