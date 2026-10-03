package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// talents_test.go — ADR 0109 §2, owner decision 2 (Delivery PR 4): a
// loyalty ability one permanent grants another. The five Talents print
// "Enchanted planeswalker has '[±N]: …'", so each is an Aura whose
// layer-6 grant puts a loyalty row on the walker it enchants.
//
// What the engine has to hold, and where each is pinned:
//
//   - CR 606.3 counts per PERMANENT: the granted row and the walker's
//     own rows share one activation a turn, in either order.
//   - CR 606.6 holds on the granted row: a −N needs N loyalty counters.
//   - The stack item names its grantor, so "you get an emblem" makes
//     the GRANTOR's emblem (CR 114.2): the text is the Talent's.
//   - A restore point holding the granted ability on the stack names it
//     by its grant ref and keeps the grantor (AGENTS.md §5).
//   - A CR 707.10 copy of the granted ability keeps the grantor too.

const (
	elspethsTalentOracle    = "969b81d2-e95b-4a3f-9548-9394e3a6727e"
	lilianasTalentOracle    = "8c55039d-1303-420c-8547-ebdffb63bf89"
	rowansTalentOracle      = "0af6f3ce-59e8-4797-aa1b-dbdb5288fe3d"
	viviensTalentOracle     = "968f0edb-b81f-465a-84b2-f077af61e51b"
	talentElspethSunsOracle = "05e6b243-48a6-4a42-bc5f-413441de9c33"
)

// talentWalker seats a planeswalker the catalog prints no abilities for,
// so the only loyalty rows it has are the ones a Talent grants it.
func talentWalker(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blank Walker", TypeLine: "Legendary Planeswalker — Blank",
		Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// talentOn attaches a Talent to `walker`, with the zone-move event that
// gives its static a timestamp.
func talentOn(g *game.Game, owner uuid.UUID, name, oracle string, walker uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Aura",
		OracleID: oracle, Owner: owner, Controller: owner,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: walker},
	})
}

// grantedLoyaltyRow is the index and ref of the walker's granted row
// whose label starts with `prefix`.
func grantedLoyaltyRow(t *testing.T, g *game.Game, walker uuid.UUID, prefix string) (int, string) {
	t.Helper()
	abs, origins := game.ActivatedAbilitiesWithOrigins(gaLayered(t, g, walker))
	for i, ab := range abs {
		if origins.At(i).Granted() && strings.HasPrefix(ab.Label, prefix) {
			if ab.Cost.Loyalty == nil {
				t.Fatalf("the granted row %q has no loyalty cost", ab.Label)
			}
			return i, origins.Ref(i)
		}
	}
	t.Fatalf("the walker has no granted row starting %q (rows: %d)", prefix, len(abs))
	return -1, ""
}

// talentStackItem is the activated item on the stack from `source`.
func talentStackItem(g *game.Game, source uuid.UUID) *game.StackItem {
	var out *game.StackItem
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemActivated && it.SourceCardID == source && !it.IsCopy {
				out = it
			}
		}
	})
	return out
}

// --- the engine half --------------------------------------------------

// CR 606.3: "only if no player has previously activated a loyalty
// ability of that permanent that turn". The granted row IS a loyalty
// ability of the walker, so it and the walker's own rows share one
// activation a turn — in both orders.
func TestGrantedLoyaltyAbilitySharesTheWalkersOncePerTurn(t *testing.T) {
	t.Run("own first, then granted", func(t *testing.T) {
		g := newCatalogGame(t)
		toMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		elspeth := pushCatalogWalker(g, me.ID, "Elspeth, Sun's Champion", talentElspethSunsOracle, 4)
		talentOn(g, me.ID, "Elspeth's Talent", elspethsTalentOracle, elspeth)
		idx, ref := grantedLoyaltyRow(t, g, elspeth, "+1: Create three")

		// Resolved before the next ask, so the stack is empty and the
		// refusal is CR 606.3's rather than the sorcery-speed window's.
		b16Activate(t, g, me.ID, elspeth, 0, game.ActivateAbilityParams{})
		if err := g.ActivateCatalogAbility(me.ID, elspeth, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrLoyaltyAlreadyActivated {
			t.Errorf("the granted +1 after her own: got %v, want ErrLoyaltyAlreadyActivated", err)
		}
	})
	t.Run("granted first, then own", func(t *testing.T) {
		g := newCatalogGame(t)
		toMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		elspeth := pushCatalogWalker(g, me.ID, "Elspeth, Sun's Champion", talentElspethSunsOracle, 4)
		talentOn(g, me.ID, "Elspeth's Talent", elspethsTalentOracle, elspeth)
		idx, ref := grantedLoyaltyRow(t, g, elspeth, "+1: Create three")

		b16Activate(t, g, me.ID, elspeth, idx, game.ActivateAbilityParams{Ref: ref})
		if got := loyaltyCount(g, elspeth); got != 5 {
			t.Errorf("loyalty after the granted +1: got %d, want 5 — the cost is paid by the walker", got)
		}
		for _, own := range []int{0, 1, 2} {
			if err := g.ActivateCatalogAbility(me.ID, elspeth, own, game.ActivateAbilityParams{}); err != game.ErrLoyaltyAlreadyActivated {
				t.Errorf("her own row %d after the granted +1: got %v, want ErrLoyaltyAlreadyActivated", own, err)
			}
		}
	})
}

// CR 606.6: a −N ability can't be activated with fewer than N loyalty
// counters, granted or not. And CR 606.3's sorcery timing holds for it.
func TestGrantedMinusTwelveNeedsTwelveLoyaltyAndSorceryTiming(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := talentWalker(g, me.ID, 11)
	talentOn(g, me.ID, "Teferi's Talent", teferisTalentOracle, walker)
	idx, ref := grantedLoyaltyRow(t, g, walker, "−12")

	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrInsufficientLoyalty {
		t.Fatalf("−12 with 11 loyalty: got %v, want ErrInsufficientLoyalty", err)
	}
	if got := loyaltyCount(g, walker); got != 11 {
		t.Fatalf("a refused activation moved the loyalty: %d", got)
	}

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(walker, game.CounterLoyalty, 1) })
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepEnd)
	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != game.ErrSorcerySpeedRequired {
		t.Fatalf("−12 on an opponent's end step: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// The −12's stack item names the Talent as its grantor, and the emblem
// it makes is the Talent's (CR 114.2) — it opens instant-speed loyalty
// for every planeswalker its owner controls, and the walker that paid
// the 12 dies to CR 704.5i.
func TestTeferisTalentMinusTwelveMakesTheTalentsEmblem(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := talentWalker(g, me.ID, 12)
	talent := talentOn(g, me.ID, "Teferi's Talent", teferisTalentOracle, walker)
	other := pushCatalogWalker(g, me.ID, "Teferi, Hero of Dominaria", teferiHeroOracleForTiming, 4)
	idx, ref := grantedLoyaltyRow(t, g, walker, "−12")

	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate the granted −12: %v", err)
	}
	item := talentStackItem(g, walker)
	if item == nil {
		t.Fatal("the −12 is not on the stack")
	}
	if item.GrantedBy != talent {
		t.Errorf("the item's grantor = %s, want the Talent %s", item.GrantedBy, talent)
	}
	if item.Params.Ability == nil || !strings.HasPrefix(item.Params.Ability.Ref, "grant:") {
		t.Errorf("the item is not named by a grant ref: %+v", item.Params.Ability)
	}
	passPriorityAroundTable(t, g)

	emblems := emblemsOfPlayer(g, me.ID)
	if len(emblems) != 1 || emblems[0].Label != "Teferi's Talent emblem" {
		t.Fatalf("emblems after the −12: %+v, want the Talent's", emblems)
	}
	if g.Battlefield.Contains(walker) {
		t.Error("the walker paid all 12 loyalty and is still on the battlefield (CR 704.5i)")
	}

	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepEnd)
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{}); err != nil {
		t.Errorf("another walker's +1 on an opponent's end step under the Talent's emblem: %v", err)
	}
}

// AGENTS.md §5: a restore point holding the granted −12 on the stack
// names it by its grant ref, keeps the grantor, and still makes the
// Talent's emblem in the restored game.
func TestTeferisTalentMinusTwelveRestoresFromASnapshot(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := talentWalker(g, me.ID, 13)
	talent := talentOn(g, me.ID, "Teferi's Talent", teferisTalentOracle, walker)
	idx, ref := grantedLoyaltyRow(t, g, walker, "−12")
	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate the granted −12: %v", err)
	}

	restored := restoreRoundTrip(t, g, true)
	item := talentStackItem(restored, walker)
	if item == nil {
		t.Fatal("the −12 did not survive the restore")
	}
	if item.GrantedBy != talent {
		t.Errorf("restored grantor = %s, want the Talent %s", item.GrantedBy, talent)
	}
	if item.Effect == nil {
		t.Fatal("the restored −12 has no effect — the grant ref was not resolved")
	}
	passPriorityAroundTable(t, restored)
	emblems := emblemsOfPlayer(restored, me.ID)
	if len(emblems) != 1 || emblems[0].Label != "Teferi's Talent emblem" {
		t.Fatalf("restored emblems after the −12: %+v, want the Talent's", emblems)
	}
	if got := loyaltyCount(restored, walker); got != 1 {
		t.Errorf("restored walker loyalty: got %d, want 1", got)
	}
}

// The activation event says whether a loyalty ability was activated, so
// "whenever you activate a loyalty ability" has something to read.
func TestActivationEventCarriesTheLoyaltyBit(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := talentWalker(g, me.ID, 3)
	talentOn(g, me.ID, "Elspeth's Talent", elspethsTalentOracle, walker)
	idx, ref := grantedLoyaltyRow(t, g, walker, "+1")
	before := len(g.Events)
	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	var found bool
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventActivateAbility {
			found = true
			if !ev.Loyalty {
				t.Errorf("the loyalty activation's event has no loyalty bit: %+v", ev)
			}
		}
	}
	if !found {
		t.Fatal("no activation event")
	}
}

// --- the cards --------------------------------------------------------

// Teferi's Talent is whole now: its catalog entry declares the grant
// and the emblem, and carries no caveat.
func TestTeferisTalentIsComplete(t *testing.T) {
	spec, ok := Lookup(teferisTalentOracle)
	if !ok {
		t.Fatal("Teferi's Talent is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v, caveats %v — want Full", spec.Completeness, spec.Caveats)
	}
	if spec.Emblem == nil || spec.Emblem.Label != "Teferi's Talent emblem" {
		t.Errorf("Teferi's Talent declares emblem %+v", spec.Emblem)
	}
}

// Elspeth's Talent: the granted +1 makes three Soldiers, and every
// loyalty activation of the enchanted walker — granted or its own —
// gives your creatures +2/+2 and vigilance until end of turn.
func TestElspethsTalentPumpsOnEveryLoyaltyActivation(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := talentWalker(g, me.ID, 3)
	talentOn(g, me.ID, "Elspeth's Talent", elspethsTalentOracle, walker)
	bear := ctrlPushCreature(g, me.ID, "My Bear")
	theirs := ctrlPushCreature(g, opp.ID, "Their Bear")
	idx, ref := grantedLoyaltyRow(t, g, walker, "+1: Create three")

	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate the granted +1: %v", err)
	}
	passPriorityAroundTable(t, g)

	soldiers := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && strings.Contains(c.Name, "Soldier") {
			soldiers++
			if got := effectivePower(t, g, c.InstanceID); got != 1 {
				t.Errorf("a Soldier made by the +1 is %d power — it entered after the pump resolved", got)
			}
		}
	}
	if soldiers != 3 {
		t.Errorf("the +1 made %d Soldiers, want 3", soldiers)
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("my creature's power: got %d, want 4", got)
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Errorf("my creature's toughness: got %d, want 4", got)
	}
	if !hasKeywordForTest(gaLayered(t, g, bear).Effective().Abilities, "vigilance") {
		t.Error("my creature did not gain vigilance")
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature was pumped: power %d", got)
	}

	advancePastCleanupForTest(t, g)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("the pump outlasted the turn: power %d", got)
	}
}

// The trigger reads "enchanted planeswalker": a loyalty ability of
// another walker you control does not trigger it.
func TestElspethsTalentIgnoresAnotherWalker(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := talentWalker(g, me.ID, 3)
	talentOn(g, me.ID, "Elspeth's Talent", elspethsTalentOracle, walker)
	other := pushCatalogWalker(g, me.ID, "Elspeth, Sun's Champion", talentElspethSunsOracle, 4)
	bear := ctrlPushCreature(g, me.ID, "My Bear")

	b16Activate(t, g, me.ID, other, 0, game.ActivateAbilityParams{})
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("another walker's loyalty ability pumped my creature: power %d", got)
	}
}

// Liliana's Talent: the granted −8 puts every creature card from every
// graveyard onto the battlefield under your control, and a creature
// that deals damage to the enchanted walker is destroyed.
func TestLilianasTalentReanimatesEveryGraveyardAndPunishesDamage(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := talentWalker(g, me.ID, 10)
	talentOn(g, me.ID, "Liliana's Talent", lilianasTalentOracle, walker)

	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Dead Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Dead Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Their Sorcery", TypeLine: "Sorcery",
		Owner: opp.ID, Controller: opp.ID})

	idx, ref := grantedLoyaltyRow(t, g, walker, "−8")
	b16Activate(t, g, me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref})
	for _, id := range []uuid.UUID{mine, theirs} {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("creature card %s did not return", id)
		}
		if c.Controller != me.ID {
			t.Errorf("%s returned under %s, want the −8's controller", c.Name, c.Controller)
		}
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("a noncreature card left the graveyard")
	}

	attacker := ctrlPushCreature(g, opp.ID, "Pinger")
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(attacker, walker, 1) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the creature that damaged the enchanted walker survived")
	}
	// Damage from a source that is not a creature does nothing.
	before := len(g.Events)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, walker, 1) })
	passPriorityAroundTable(t, g)
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventTrigger && strings.HasPrefix(ev.Label, "Liliana's Talent") {
			t.Errorf("damage from no creature triggered it: %+v", ev)
		}
	}
}

// Rowan's Talent: the granted +1 pumps up to one target creature, and
// every loyalty activation of the enchanted walker is copied. Paired
// with Teferi's Talent on the same walker, the copied −12 is still the
// Teferi's Talent's: a CR 707.10 copy keeps the grantor.
func TestRowansTalentCopiesTheLoyaltyAbilityAndKeepsItsGrantor(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := talentWalker(g, me.ID, 12)
	talentOn(g, me.ID, "Rowan's Talent", rowansTalentOracle, walker)
	teferis := talentOn(g, me.ID, "Teferi's Talent", teferisTalentOracle, walker)
	idx, ref := grantedLoyaltyRow(t, g, walker, "−12")

	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate the granted −12: %v", err)
	}
	passPriorityAroundTable(t, g)
	emblems := emblemsOfPlayer(g, me.ID)
	if len(emblems) != 2 {
		t.Fatalf("the −12 and its copy made %d emblems, want 2: %+v", len(emblems), emblems)
	}
	for _, e := range emblems {
		if e.Label != "Teferi's Talent emblem" {
			t.Errorf("an emblem is %q, want Teferi's Talent's (grantor %s)", e.Label, teferis)
		}
	}
}

func TestRowansTalentPlusOnePumpsAndIsCopied(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := talentWalker(g, me.ID, 3)
	talentOn(g, me.ID, "Rowan's Talent", rowansTalentOracle, walker)
	bear := ctrlPushCreature(g, me.ID, "My Bear")
	idx, ref := grantedLoyaltyRow(t, g, walker, "+1")

	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{
		Ref:     ref,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate the granted +1: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The copy may choose new targets; keep the bear.
	if pick := latestChoiceOfKind(g, game.PendingChoicePickTarget); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	// +2/+0 twice: the ability and its copy.
	if got := effectivePower(t, g, bear); got != 6 {
		t.Errorf("power after the +1 and its copy: got %d, want 6", got)
	}
	ab := gaLayered(t, g, bear).Effective().Abilities
	if !hasKeywordForTest(ab, "first strike") || !hasKeywordForTest(ab, "trample") {
		t.Errorf("the bear's abilities %v, want first strike and trample", ab)
	}
	if got := loyaltyCount(g, walker); got != 4 {
		t.Errorf("loyalty: got %d, want 4 — a copy is not activated and pays nothing", got)
	}
}

// Vivien's Talent: the granted +1 looks at four and may take a creature
// or land, and a nontoken creature entering under your control puts a
// loyalty counter on the enchanted walker.
func TestViviensTalentDigsAndGrowsTheWalker(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := talentWalker(g, me.ID, 3)
	talentOn(g, me.ID, "Vivien's Talent", viviensTalentOracle, walker)

	seeded := seedLibrary(me, "A", "B", "C", "D", "Fifth")
	creature := seeded[2]
	me.Library.Cards[me.Library.Size()-3].TypeLine = testCreatureTypeLine
	handBefore := me.Hand.Size()

	idx, ref := grantedLoyaltyRow(t, g, walker, "+1")
	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate the granted +1: %v", err)
	}
	passPriorityAroundTable(t, g)
	choice := chooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt to take a creature or land card")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(creature) || me.Hand.Size() != handBefore+1 {
		t.Fatal("the creature card did not reach the hand")
	}
	if got := loyaltyCount(g, walker); got != 4 {
		t.Fatalf("loyalty after the +1: got %d, want 4", got)
	}

	arriving := pushCatalogHandCard(me, "Arriving Bear", testCreatureTypeLine, "")
	g.WithWriteLock(func() {
		_, _ = g.PutFromHandOntoBattlefieldForEffect(arriving, game.HandEntryOptions{Controller: me.ID})
	})
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, walker); got != 5 {
		t.Errorf("loyalty after a nontoken creature entered: got %d, want 5", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 1) })
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, walker); got != 5 {
		t.Errorf("a creature TOKEN entering moved the loyalty: got %d", got)
	}
}
