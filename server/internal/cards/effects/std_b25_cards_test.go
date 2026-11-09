package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b25_cards_test.go — Standard-set batch 25 (Duskmourn leftovers,
// #4452, #4466). Each test plays the real catalog entry and asserts what its
// text says.

const (
	b25CoordinatedClobberingOracle = "b2b65b17-1fb4-42d6-9c6a-376d3f18851b"
	b25CryptidInspectorOracle      = "c36f787e-d8d4-4bd5-9790-64766878fe46"
	b25MonstrousEmergenceOracle    = "5542ea4d-34ab-4f89-8894-793a7038f501"
	b25OmnivorousFlytrapOracle     = "0aa2ab46-c5ad-462b-b441-dbbc4b2c3532"
	b25OvergrownZealotOracle       = "5146b291-1170-481f-a4b9-709dd5c4b09c"
	b25RootwiseSurvivorOracle      = "e42a6880-7ca6-498d-abc9-fb7ccd08b111"
	b25BaseballBatOracle           = "9b5a3765-943b-4925-b008-fc734b2934f7"
	b25BeastieBeatdownOracle       = "3da5fbe2-e47b-4a1d-9e73-393f37314fa3"
	b25BroodspinnerOracle          = "ad50214c-4d0e-4f95-9098-fd7d64dbf7f4"
)

func TestB25EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b25CoordinatedClobberingOracle: "Coordinated Clobbering",
		b25CryptidInspectorOracle:      "Cryptid Inspector",
		b25MonstrousEmergenceOracle:    "Monstrous Emergence",
		b25OmnivorousFlytrapOracle:     "Omnivorous Flytrap",
		b25OvergrownZealotOracle:       "Overgrown Zealot",
		b25RootwiseSurvivorOracle:      "Rootwise Survivor",
		b25BaseballBatOracle:           "Baseball Bat",
		b25BeastieBeatdownOracle:       "Beastie Beatdown",
		b25BroodspinnerOracle:          "Broodspinner",
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

// b25SeedTypes fills p's graveyard with one card of each of the first n
// card types: creature, instant, sorcery, land, artifact, enchantment.
func b25SeedTypes(g *game.Game, p *game.Player, n int) {
	types := []string{"Creature — Bear", "Instant", "Sorcery", "Basic Land — Forest", "Artifact", "Enchantment"}
	g.WithWriteLock(func() {
		for i, tl := range types[:n] {
			p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old " + tl, TypeLine: tl, Owner: p.ID, Controller: p.ID})
			_ = i
		}
		g.BumpLayerVersionForTest()
	})
}

func b25Damage(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return e2Card(t, g, id).DamageMarked
}

// --- Coordinated Clobbering ------------------------------------------------

func b25ClobberRefs(mine []uuid.UUID, victim uuid.UUID) []game.TargetRef {
	var refs []game.TargetRef
	for _, id := range mine {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: 0})
	}
	return append(refs, game.TargetRef{Kind: game.TargetCard, ID: victim, Slot: 1})
}

func TestB25CoordinatedClobberingTapsBothAndEachDealsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	b := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Coordinated Clobbering", "Sorcery", b25CoordinatedClobberingOracle, b25ClobberRefs([]uuid.UUID{a, b}, victim))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !e2Card(t, g, id).Tapped {
			t.Errorf("%s was not tapped", e2Card(t, g, id).Name)
		}
	}
	if got := b25Damage(t, g, victim); got != 5 {
		t.Errorf("the victim took %d damage, want 3 + 2 = 5", got)
	}
}

func TestB25CoordinatedClobberingOneCreatureIsEnough(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Coordinated Clobbering", "Sorcery", b25CoordinatedClobberingOracle, b25ClobberRefs([]uuid.UUID{a}, victim))
	passPriorityAroundTable(t, g)
	if got := b25Damage(t, g, victim); got != 4 {
		t.Errorf("the victim took %d damage, want 4", got)
	}
}

func TestB25CoordinatedClobberingRefusesTappedAndWrongControllerTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tapped := b12Creature(g, me.ID, "Tapped", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	ready := b12Creature(g, me.ID, "Ready", "Creature — Bear", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 9)
	mine2 := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	if err := castCatalogSpellErr(t, g, "Coordinated Clobbering", "Sorcery", b25CoordinatedClobberingOracle,
		b25ClobberRefs([]uuid.UUID{tapped}, theirs)); err == nil {
		t.Error("a tapped creature was accepted as the source")
	}
	if err := castCatalogSpellErr(t, g, "Coordinated Clobbering", "Sorcery", b25CoordinatedClobberingOracle,
		b25ClobberRefs([]uuid.UUID{theirs}, theirs)); err == nil {
		t.Error("an opponent's creature was accepted as a source")
	}
	if err := castCatalogSpellErr(t, g, "Coordinated Clobbering", "Sorcery", b25CoordinatedClobberingOracle,
		b25ClobberRefs([]uuid.UUID{ready}, mine2)); err == nil {
		t.Error("my own creature was accepted as the victim")
	}
	if err := castCatalogSpellErr(t, g, "Coordinated Clobbering", "Sorcery", b25CoordinatedClobberingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 1}}); err == nil {
		t.Error("a cast with no creature of mine was accepted")
	}
}

func TestB25CoordinatedClobberingSkipsASourceTappedInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	b := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Coordinated Clobbering", "Sorcery", b25CoordinatedClobberingOracle, b25ClobberRefs([]uuid.UUID{a, b}, victim))
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(a) })
	passPriorityAroundTable(t, g)
	if got := b25Damage(t, g, victim); got != 2 {
		t.Errorf("the victim took %d damage, want only the untapped creature's 2", got)
	}
}

func TestB25CoordinatedClobberingDealsNothingWhenTheVictimIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Coordinated Clobbering", "Sorcery", b25CoordinatedClobberingOracle, b25ClobberRefs([]uuid.UUID{a}, victim))
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	passPriorityAroundTable(t, g)
	if !e2Card(t, g, a).Tapped {
		t.Error("the source was not tapped although the spell still had a legal target")
	}
	if got := b25Damage(t, g, a); got != 0 {
		t.Errorf("my creature took %d damage", got)
	}
}

// --- Beastie Beatdown ------------------------------------------------------

func TestB25BeastieBeatdownPutsCountersOnlyWithDelirium(t *testing.T) {
	for _, types := range []int{3, 4} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
		b25SeedTypes(g, me, types)
		castCatalogSpell(t, g, "Beastie Beatdown", "Sorcery", b25BeastieBeatdownOracle, b25ClobberRefs([]uuid.UUID{mine}, victim))
		passPriorityAroundTable(t, g)
		wantCounters, wantDamage := 0, 2
		if types == 4 {
			wantCounters, wantDamage = 2, 4
		}
		if got := countersOn(g, mine, game.CounterPlusOne); got != wantCounters {
			t.Errorf("%d types: %d +1/+1 counters, want %d", types, got, wantCounters)
		}
		if got := b25Damage(t, g, victim); got != wantDamage {
			t.Errorf("%d types: the victim took %d damage, want %d", types, got, wantDamage)
		}
	}
}

func TestB25BeastieBeatdownCountsTheGraveyardAsItResolvesNotTheSpell(t *testing.T) {
	// Three types in the graveyard and the Sorcery itself still on the
	// stack: the spell is the fourth type only once it has resolved.
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	b25SeedTypes(g, me, 2) // creature and instant
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	castCatalogSpell(t, g, "Beastie Beatdown", "Sorcery", b25BeastieBeatdownOracle, b25ClobberRefs([]uuid.UUID{mine}, victim))
	passPriorityAroundTable(t, g)
	if got := countersOn(g, mine, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters with three types and the spell on the stack, want none", got)
	}
}

func TestB25BeastieBeatdownNeedsBothTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	if err := castCatalogSpellErr(t, g, "Beastie Beatdown", "Sorcery", b25BeastieBeatdownOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}}); err == nil {
		t.Error("a cast with only my creature was accepted")
	}
	if err := castCatalogSpellErr(t, g, "Beastie Beatdown", "Sorcery", b25BeastieBeatdownOracle,
		b25ClobberRefs([]uuid.UUID{victim}, victim)); err == nil {
		t.Error("an opponent's creature was accepted as mine")
	}
}

func TestB25BeastieBeatdownStillGrowsMyCreatureWhenTheirsIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	b25SeedTypes(g, me, 4)
	castCatalogSpell(t, g, "Beastie Beatdown", "Sorcery", b25BeastieBeatdownOracle, b25ClobberRefs([]uuid.UUID{mine}, victim))
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2: only the creature I control must still be a legal target", got)
	}
}

func TestB25BeastieBeatdownDoesNothingToMineWhenItIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	b25SeedTypes(g, me, 4)
	castCatalogSpell(t, g, "Beastie Beatdown", "Sorcery", b25BeastieBeatdownOracle, b25ClobberRefs([]uuid.UUID{mine}, victim))
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	passPriorityAroundTable(t, g)
	if got := b25Damage(t, g, victim); got != 0 {
		t.Errorf("the victim took %d damage from a creature that was gone", got)
	}
}

// --- Cryptid Inspector -----------------------------------------------------

func TestB25CryptidInspectorGrowsForAFaceDownEntryAndAFaceUpTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	inspector := b12Push(g, me.ID, "Cryptid Inspector", "Creature — Elf Warrior", b25CryptidInspectorOracle, 2, 3)
	// A face-up creature entering does nothing.
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, inspector, game.CounterPlusOne); got != 0 {
		t.Fatalf("a face-up creature entering gave %d counters", got)
	}
	spy := mkmMPCastFaceDown(t, g, "Mistway Spy", mkmMP2SpyTypeLine, mkmMP2SpyOracle, 1, 1)
	b30SettleOrder(t, g)
	if got := countersOn(g, inspector, game.CounterPlusOne); got != 1 {
		t.Fatalf("a face-down permanent entered and the Inspector has %d counters, want 1", got)
	}
	mkmMPTurnFaceUp(t, g, me, spy, "U", "C")
	b30SettleOrder(t, g)
	if got := countersOn(g, inspector, game.CounterPlusOne); got != 2 {
		t.Errorf("a permanent turned face up and the Inspector has %d counters, want 2", got)
	}
	if p, tough := ptOf(t, g, inspector); p != 4 || tough != 5 {
		t.Errorf("the Inspector is %d/%d, want 4/5", p, tough)
	}
}

func TestB25CryptidInspectorIgnoresAnOpponentsFaceDownPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inspector := b12Push(g, me.ID, "Cryptid Inspector", "Creature — Elf Warrior", b25CryptidInspectorOracle, 2, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventTurnedFaceUp, CardID: theirs, Actor: opp.ID})
	})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, inspector, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's permanent turned face up gave %d counters", got)
	}
}

// --- Baseball Bat ----------------------------------------------------------

func TestB25BaseballBatAttachesOnEntryAndTapsOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	bat := castCatalogSpell(t, g, "Baseball Bat", "Artifact — Equipment", b25BaseballBatOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := e2Card(t, g, bat); c.AttachedTo.ID != bear {
		t.Fatalf("the Bat is attached to %v, want the Bear", c.AttachedTo.ID)
	}
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tough)
	}
	declareAttack(t, g, opp.ID, bear)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, blocker)
	passPriorityAroundTable(t, g)
	if !e2Card(t, g, blocker).Tapped {
		t.Error("the targeted creature was not tapped")
	}
}

func TestB25BaseballBatsAttackClauseIsUpToOneTargetCreature(t *testing.T) {
	spec, ok := Lookup(b25BaseballBatOracle)
	if !ok || len(spec.Triggered) != 2 || spec.Triggered[1].Targets == nil || spec.Triggered[1].Targets.Min != 0 || spec.Triggered[1].Targets.Max != 1 {
		t.Fatalf("the attack trigger's clause is not \"up to one target creature\": %+v", spec.Triggered)
	}
}

// --- Overgrown Zealot ------------------------------------------------------

func TestB25OvergrownZealotMakesTwoRestrictedManaOfOneColor(t *testing.T) {
	spec, ok := Lookup(b25OvergrownZealotOracle)
	if !ok || len(spec.ManaAbilities) != 2 {
		t.Fatalf("Overgrown Zealot has %d mana abilities, want 2", len(spec.ManaAbilities))
	}
	any, two := spec.ManaAbilities[0], spec.ManaAbilities[1]
	if any.Produced != "{W|U|B|R|G}" || len(any.Restrictions) != 0 {
		t.Errorf("the first ability produces %q with restrictions %v", any.Produced, any.Restrictions)
	}
	if len(two.Restrictions) != 1 || two.Restrictions[0] != game.ManaRestrictTurnFaceUp {
		t.Errorf("the second ability's restrictions are %v, want turn-face-up only", two.Restrictions)
	}
	g := newCatalogGame(t)
	amount := two.ProducedAmount(g, g.Seats[0].ID, uuid.New())
	if amount.N != 2 || !amount.OneColor || len(amount.Options) != 5 {
		t.Errorf("the second ability produces %+v, want two mana of any one color", amount)
	}
}

func TestB25OvergrownZealotsTwoManaPayATurnFaceUpCostOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castWithAltCost(t, g, "Willbender", "Creature — Human Wizard", willbenderOracle, "morph")
	passPriorityAroundTable(t, g)
	if perm := findCardAnywhere(t, g, id); !perm.FaceDownIsPermanent() {
		t.Fatal("the Willbender did not enter face down")
	}
	zealot := pushCatalogPermanent(g, me.ID, "Overgrown Zealot", "Creature — Elf Druid", b25OvergrownZealotOracle, false)
	if err := g.ActivateManaAbility(me.ID, zealot, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Zealot for its second ability: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["U"] != 2 {
		t.Fatalf("pick = %+v, want two mana of the one color chosen", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("the pool holds %d mana, want 2", me.ManaPool.Total())
	}
	for _, u := range me.ManaPool.Units() {
		if len(u.Restrictions) == 0 {
			t.Fatalf("the Zealot's two mana are unrestricted: %+v", me.ManaPool)
		}
	}
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up with the Zealot's mana: %v", err)
	}
	if findCardAnywhere(t, g, id).FaceDown {
		t.Error("the Willbender is still face down")
	}
}

// --- Omnivorous Flytrap ----------------------------------------------------

func TestB25OmnivorousFlytrapDistributesOnEntryWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b25SeedTypes(g, me, 4)
	trap := castCatalogSpell(t, g, "Omnivorous Flytrap", "Creature — Plant", b25OmnivorousFlytrapOracle, nil)
	passPriorityAroundTable(t, g)
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{trap: 1, bear: 1}, trap, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, trap, game.CounterPlusOne); got != 1 {
		t.Errorf("the Flytrap has %d counters, want 1", got)
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1 (four types do not double)", got)
	}
}

func TestB25OmnivorousFlytrapDoublesTheCountersWithSixTypes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, bear); c != nil {
			c.Counters = map[string]int{game.CounterPlusOne: 1}
		}
	})
	b25SeedTypes(g, me, 6)
	trap := castCatalogSpell(t, g, "Omnivorous Flytrap", "Creature — Plant", b25OmnivorousFlytrapOracle, nil)
	passPriorityAroundTable(t, g)
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{trap: 1, bear: 1}, trap, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, trap, game.CounterPlusOne); got != 2 {
		t.Errorf("the Flytrap has %d counters, want 1 doubled to 2", got)
	}
	// The Bear had one, gets one (2) and doubles to 4: "those creatures" are
	// all the targets, with every counter they hold.
	if got := countersOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("the Bear has %d counters, want (1 + 1) doubled to 4", got)
	}
}

func TestB25OmnivorousFlytrapDoesNotTriggerWithoutDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b25SeedTypes(g, me, 3)
	trap := castCatalogSpell(t, g, "Omnivorous Flytrap", "Creature — Plant", b25OmnivorousFlytrapOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, trap) != nil {
		t.Error("the Flytrap triggered with three card types")
	}
}

func TestB25OmnivorousFlytrapAlsoTriggersOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	trap := b12Push(g, me.ID, "Omnivorous Flytrap", "Creature — Plant", b25OmnivorousFlytrapOracle, 2, 4)
	b25SeedTypes(g, me, 4)
	declareAttack(t, g, opp.ID, trap)
	b04WaitForPick(t, g, me.ID)
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{trap: 2}, trap)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, trap, game.CounterPlusOne); got != 2 {
		t.Errorf("the attacking Flytrap has %d counters, want 2", got)
	}
}

func TestB25OmnivorousFlytrapRechecksDeliriumOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b25SeedTypes(g, me, 4)
	trap := castCatalogSpell(t, g, "Omnivorous Flytrap", "Creature — Plant", b25OmnivorousFlytrapOracle, nil)
	passPriorityAroundTable(t, g)
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{trap: 1, bear: 1}, trap, bear)
	// A card leaves the graveyard in response: three types remain.
	g.WithWriteLock(func() {
		cards := me.Graveyard.Cards
		me.Graveyard.Cards = cards[:len(cards)-1]
	})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the Bear has %d counters although delirium was gone on resolution", got)
	}
}

// --- Rootwise Survivor -----------------------------------------------------

func b25SurvivorTable(t *testing.T, tapped bool) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	survivor := b12Push(g, me.ID, "Rootwise Survivor", "Creature — Human Survivor", b25RootwiseSurvivorOracle, 3, 4)
	forest := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	if tapped {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(survivor) })
	}
	return g, me, survivor, forest
}

func TestB25RootwiseSurvivorAnimatesALandIntoAThreeThreeElemental(t *testing.T) {
	g, me, _, forest := b25SurvivorTable(t, true)
	advanceTo(t, g, game.StepPostcombatMain)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, forest, game.CounterPlusOne); got != 3 {
		t.Fatalf("the land has %d +1/+1 counters, want 3", got)
	}
	c := e2Card(t, g, forest)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Elemental") {
		t.Errorf("the land is not a land Elemental creature: %q", c.TypeLine)
	}
	if p, tough := ptOf(t, g, forest); p != 3 || tough != 3 {
		t.Errorf("the land is %d/%d, want 3/3 (0/0 with three counters)", p, tough)
	}
	if !hasKeywordOf(g, forest, "haste") {
		t.Error("the land has no haste")
	}
}

func TestB25RootwiseSurvivorsLandStaysACreatureAfterTheHasteEnds(t *testing.T) {
	g, me, _, forest := b25SurvivorTable(t, true)
	advanceTo(t, g, game.StepPostcombatMain)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	// Walk to the controller's next turn: "until your next turn" ends the haste.
	for i := 0; i < 80; i++ {
		if g.Turn.ActiveSeat == (g.Turn.ActiveSeat+0) && i > 0 && g.Turn.Step == game.StepUntap && g.Seats[g.Turn.ActiveSeat].ID == me.ID {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Seats[g.Turn.ActiveSeat].ID == me.ID && g.Turn.Step == game.StepUpkeep {
			break
		}
	}
	if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		t.Fatal("setup: never reached my next turn")
	}
	if hasKeywordOf(g, forest, "haste") {
		t.Error("the land still has haste on my next turn")
	}
	if c := e2Card(t, g, forest); !c.IsCreature() {
		t.Error("the land stopped being a creature with the haste")
	}
	if p, tough := ptOf(t, g, forest); p != 3 || tough != 3 {
		t.Errorf("the land is %d/%d, want 3/3", p, tough)
	}
}

func TestB25RootwiseSurvivorUntappedDoesNotTrigger(t *testing.T) {
	g, me, _, forest := b25SurvivorTable(t, false)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an untapped Rootwise Survivor asked for a target")
	}
	if got := countersOn(g, forest, game.CounterPlusOne); got != 0 {
		t.Errorf("the land got %d counters", got)
	}
}

func TestB25RootwiseSurvivorMayChooseNoLandAndOnlyMyLands(t *testing.T) {
	spec, ok := Lookup(b25RootwiseSurvivorOracle)
	if !ok || len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil {
		t.Fatalf("Rootwise Survivor has no targeting trigger: %+v", spec.Triggered)
	}
	if ts := spec.Triggered[0].Targets; ts.Min != 0 || ts.Max != 1 {
		t.Errorf("the clause counts %d..%d, want up to one", ts.Min, ts.Max)
	}
	g, me, _, forest := b25SurvivorTable(t, true)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := pushPermanentForTest(g, opp.ID, "Island", "", "Basic Land — Island")
	advanceTo(t, g, game.StepPostcombatMain)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's land was accepted")
	}
	pickCard(t, g, me.ID, forest)
}

// --- Broodspinner ----------------------------------------------------------

func TestB25BroodspinnerSurveilsTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Broodspinner", "Creature — Spider", b25BroodspinnerOracle, nil)
	passPriorityAroundTable(t, g)
	eoeQZKeepOnTop(t, g, me.ID, eoeQZSurveil(t, g, me.ID, 2))
}

func TestB25BroodspinnerMakesAnInsectPerCardTypeIncludingItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spinner := b12Push(g, me.ID, "Broodspinner", "Creature — Spider", b25BroodspinnerOracle, 2, 3)
	b25SeedTypes(g, me, 3) // creature, instant, sorcery
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("B", "G", "B", "G", "B", "G")...) })
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, spinner, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	insects := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Insect") {
				insects++
			}
		}
	})
	// The graveyard held creature, instant and sorcery already; the
	// sacrificed Broodspinner is a creature card, so it adds nothing new.
	if insects != 3 {
		t.Errorf("%d Insect tokens, want 3", insects)
	}
	if g.Battlefield.Contains(spinner) {
		t.Error("the Broodspinner was not sacrificed")
	}
}

func TestB25BroodspinnerCountsItselfWhenItIsTheOnlyType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spinner := b12Push(g, me.ID, "Broodspinner", "Creature — Spider", b25BroodspinnerOracle, 2, 3)
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("B", "G", "B", "G", "B", "G")...) })
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, spinner, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	insects := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Insect") {
				insects++
			}
		}
	})
	if insects != 1 {
		t.Errorf("%d Insect tokens, want 1 for the Broodspinner's own creature type", insects)
	}
}

// --- Monstrous Emergence ---------------------------------------------------

func b25EmergenceInHand(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Monstrous Emergence", TypeLine: "Sorcery",
		OracleID: b25MonstrousEmergenceOracle, Owner: p.ID, Controller: p.ID})
	return id
}

func b25HandCard(p *game.Player, name, typeLine string, power int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: power, Owner: p.ID, Controller: p.ID})
	return id
}

func b25CastEmergence(t *testing.T, g *game.Game, p *game.Player, spell uuid.UUID, params game.CastSpellParams) error {
	t.Helper()
	advanceToMain(t, g)
	return g.CastSpell(p.ID, spell, params)
}

func TestB25MonstrousEmergenceDealsTheRevealedCardsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	spell := b25EmergenceInHand(g, me)
	giant := b25HandCard(me, "Hand Giant", "Creature — Giant", 5)
	if err := b25CastEmergence(t, g, me, spell, game.CastSpellParams{Targets: b16TargetCard(victim), RevealIDs: []uuid.UUID{giant}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b25Damage(t, g, victim); got != 5 {
		t.Errorf("the victim took %d damage, want the revealed card's 5", got)
	}
	if !me.Hand.Contains(giant) {
		t.Error("revealing the card moved it out of the hand")
	}
}

func TestB25MonstrousEmergenceReadsTheChosenCreaturesPowerAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spell := b25EmergenceInHand(g, me)
	if err := b25CastEmergence(t, g, me, spell, game.CastSpellParams{Targets: b16TargetCard(victim), RevealIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2) })
	passPriorityAroundTable(t, g)
	if got := b25Damage(t, g, victim); got != 4 {
		t.Errorf("the victim took %d damage, want the pumped creature's 4", got)
	}
}

func TestB25MonstrousEmergenceUsesLastKnownPowerOfAChosenCreatureThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	spell := b25EmergenceInHand(g, me)
	if err := b25CastEmergence(t, g, me, spell, game.CastSpellParams{Targets: b16TargetCard(victim), RevealIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if got := b25Damage(t, g, victim); got != 3 {
		t.Errorf("the victim took %d damage, want the departed creature's last-known 3", got)
	}
}

func TestB25MonstrousEmergenceRefusesAnIllegalChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := b25HandCard(me, "Hand Forest", "Basic Land — Forest", 0)
	creatureCard := b25HandCard(me, "Hand Bear", "Creature — Bear", 2)
	spell := b25EmergenceInHand(g, me)
	advanceToMain(t, g)
	cases := map[string][]uuid.UUID{
		"no choice at all":             nil,
		"a land card":                  {land},
		"an opponent's creature":       {theirs},
		"two objects":                  {creatureCard, creatureCard},
		"a card that is not in hand":   {uuid.New()},
		"the spell itself in the hand": {spell},
	}
	for name, ids := range cases {
		err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: b16TargetCard(victim), RevealIDs: ids})
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: CastSpell = %v, want ErrInvalidParam", name, err)
		}
		if g.Stack.Contains(spell) {
			t.Fatalf("%s: the refused cast left the spell on the stack", name)
		}
	}
}

func TestB25MonstrousEmergenceCannotBeCastWithNothingToChoose(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	spell := b25EmergenceInHand(g, me)
	b25HandCard(me, "Hand Forest", "Basic Land — Forest", 0)
	err := b25CastEmergence(t, g, me, spell, game.CastSpellParams{Targets: b16TargetCard(victim)})
	if err == nil {
		t.Error("cast with no creature to choose and no creature card to reveal was accepted")
	}
}
