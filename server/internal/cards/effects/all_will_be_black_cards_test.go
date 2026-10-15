package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// all_will_be_black_cards_test.go — #2953, the seven cards the
// All Will Be Black deck request (#2910) was missing: Liliana of the
// Dark Realms, Oathsworn Vampire, Phyresis, Pontiff of Blight,
// Timeline Culler, Tragic Slip and Yawgmoth's Vile Offering.

const (
	lilianaDarkRealmsOracle = "92301066-8904-41bc-84bd-4aa286cbab9c"
	oathswornVampireOracle  = "e1fb696f-1df8-4571-9710-0d79001326d3"
	phyresisOracle          = "958628af-243c-4eb0-8620-5d483b6da768"
	pontiffOfBlightOracle   = "f5a95092-fb71-4b55-bc54-782bd1c603e1"
	timelineCullerOracle    = "3c307122-bc18-41f1-87d5-ae6b3ca9b8b9"
	tragicSlipOracle        = "fa6e2a47-a8cb-4224-ae3e-f61c90477285"
	vileOfferingOracle      = "740f9ab2-fc94-4abc-82b4-6b53116bf1a4"
)

// awbLibraryOrder is the library's cards top to bottom, by instance,
// leaving out `except`.
func awbLibraryOrder(p *game.Player, except uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range p.Library.Cards {
		if c.InstanceID != except {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// awbSameOrder reports whether two orders are identical.
func awbSameOrder(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// awbSearches counts the library searches `player` made since event `from`.
func awbSearches(g *game.Game, player uuid.UUID, from int) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == player {
			n++
		}
	}
	return n
}

// awbPadLibrary seeds filler so a shuffle is observable.
func awbPadLibrary(p *game.Player) {
	for i := 0; i < 12; i++ {
		pushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
}

// awbSwamps puts n Swamps onto the battlefield under `owner`.
func awbSwamps(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, gaPut(g, owner, "Swamp", "Basic Land — Swamp", ""))
	}
	return ids
}

// awbLiliana puts Liliana on the battlefield and walks to her
// controller's main phase.
func awbLiliana(t *testing.T, g *game.Game, loyalty int) (*game.Player, uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	lili := pushWalkerForTest(g, me.ID, "Liliana of the Dark Realms", lilianaDarkRealmsOracle, loyalty)
	advanceToMainOf(t, g, seat)
	return me, lili
}

// --- Liliana of the Dark Realms ---------------------------------------

// The +1 finds any Swamp card, nonbasic included, reveals it, puts it
// into the hand and shuffles.
func TestLilianaDarkRealmsPlusOneFindsANonbasicSwampAndShuffles(t *testing.T) {
	g := newCatalogGame(t)
	me, lili := awbLiliana(t, g, 3)
	bayou := pushLibraryCardForTest(me, game.Card{Name: "Scrubland Swamp", TypeLine: "Land — Plains Swamp"})
	pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	awbPadLibrary(me)
	handBefore := me.Hand.Size()
	orderBefore := awbLibraryOrder(me, bayou)
	from := len(g.Events)

	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchByID(t, g, me.ID, bayou)
	}

	if !me.Hand.Contains(bayou) || me.Hand.Size() != handBefore+1 {
		t.Fatalf("the Swamp card is not in hand (hand %d, was %d)", me.Hand.Size(), handBefore)
	}
	if got := loyaltyOf(g, lili); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
	if n := awbSearches(g, me.ID, from); n != 1 {
		t.Errorf("%d searches, want 1", n)
	}
	if awbSameOrder(awbLibraryOrder(me, bayou), orderBefore) {
		t.Error("the library was not shuffled")
	}
}

// A library with no Swamp finds nothing, and still shuffles.
func TestLilianaDarkRealmsPlusOneWithNoSwampStillShuffles(t *testing.T) {
	g := newCatalogGame(t)
	me, lili := awbLiliana(t, g, 3)
	pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	awbPadLibrary(me)
	handBefore := me.Hand.Size()
	orderBefore := awbLibraryOrder(me, uuid.Nil)
	from := len(g.Events)

	if err := g.ActivateCatalogAbility(me.ID, lili, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchFailToFind(t, g, me.ID)
	}

	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d, want %d: nothing to find", me.Hand.Size(), handBefore)
	}
	if got := loyaltyOf(g, lili); got != 4 {
		t.Errorf("loyalty %d, want 4 (the +1 is paid either way)", got)
	}
	if n := awbSearches(g, me.ID, from); n != 1 {
		t.Errorf("%d searches, want 1", n)
	}
	if awbSameOrder(awbLibraryOrder(me, uuid.Nil), orderBefore) {
		t.Error("the library was not shuffled")
	}
}

// The −3 asks +X/+X or -X/-X as it resolves, and X is the Swamps you
// control then, not when it was activated.
func TestLilianaDarkRealmsMinusThreeBothSignsAndLiveCount(t *testing.T) {
	for _, tc := range []struct {
		name   string
		answer int
		want   int
	}{
		{"plus", 0, 3},
		{"minus", 1, -3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, lili := awbLiliana(t, g, 5)
			awbSwamps(g, me.ID, 2)
			bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 5, 5)

			if err := g.ActivateCatalogAbility(me.ID, lili, 1, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
			}); err != nil {
				t.Fatalf("−3: %v", err)
			}
			if got := loyaltyOf(g, lili); got != 2 {
				t.Errorf("loyalty %d, want 2", got)
			}
			// A third Swamp arrives in response: X counts it.
			awbSwamps(g, me.ID, 1)
			passPriorityAroundTable(t, g)
			answerOptionPick(t, g, me.ID, tc.answer)

			if got := effectivePower(t, g, bear); got != 5+tc.want {
				t.Errorf("power %d, want %d", got, 5+tc.want)
			}
			if got := effectiveToughness(t, g, bear); got != 5+tc.want {
				t.Errorf("toughness %d, want %d", got, 5+tc.want)
			}
		})
	}
}

// The −3 does nothing, and asks nothing, once its target is gone.
func TestLilianaDarkRealmsMinusThreeWithAnIllegalTargetAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, lili := awbLiliana(t, g, 5)
	awbSwamps(g, me.ID, 2)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, lili, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("the sign question was asked about a creature that is gone")
	}
}

// The −6 emblem gives every Swamp {T}: Add {B}{B}{B}{B}, for as long as
// the emblem lasts: Liliana leaving does not end it, and a Swamp that
// arrives later has it too.
func TestLilianaDarkRealmsEmblemGrantsSwampsFourBlackAfterSheLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, lili := awbLiliana(t, g, 6)
	swamp := awbSwamps(g, me.ID, 1)[0]
	forest := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	theirs := gaPut(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp", "")

	if err := g.ActivateCatalogAbility(me.ID, lili, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("emblems %d, want 1", me.Emblems.Size())
	}
	// Liliana is at 0 loyalty and dies to state-based actions.
	if g.Battlefield.Contains(lili) {
		t.Fatal("Liliana should have left the battlefield at 0 loyalty")
	}

	if n := grantedManaRowsOf(g, swamp); n != 1 {
		t.Fatalf("your Swamp has %d granted mana rows, want 1", n)
	}
	if n := grantedManaRowsOf(g, forest); n != 0 {
		t.Errorf("a Forest has %d granted rows, want 0", n)
	}
	if n := grantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's Swamp has %d granted rows, want 0", n)
	}
	later := awbSwamps(g, me.ID, 1)[0]
	if n := grantedManaRowsOf(g, later); n != 1 {
		t.Errorf("a Swamp that arrived later has %d granted rows, want 1", n)
	}

	if err := tapGrantedMana(t, g, me.ID, swamp, "B"); err != nil {
		t.Fatalf("tap the Swamp: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 4 {
		t.Errorf("pool %v, want BBBB", got)
	}
}

// --- Oathsworn Vampire ------------------------------------------------

func awbVampireInGraveyard(t *testing.T, g *game.Game) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Oathsworn Vampire", TypeLine: "Creature — Vampire Knight",
		ManaCost: "{1}{B}", OracleID: oathswornVampireOracle, Owner: me.ID, Controller: me.ID,
		Power: 2, Toughness: 2,
	})
	advanceToMain(t, g)
	return me, id
}

// Without life gained this turn the graveyard is closed to it; with it,
// the cast costs the printed {1}{B}, and the creature enters tapped.
func TestOathswornVampireGraveyardCastNeedsLifeGainedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, id := awbVampireInGraveyard(t, g)
	b06AddMana(me, "B", "B")

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast from the graveyard with no life gained this turn")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("a refused cast must leave the card in the graveyard")
	}

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast after gaining life: %v", err)
	}
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the Vampire did not resolve onto the battlefield")
	}
	if !card.Tapped {
		t.Error("the Vampire enters tapped")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("the printed {1}{B} was not charged: pool %v", me.ManaPool)
	}
}

// Life gained on an earlier turn, life lost, and an opponent's gain do
// not open the graveyard.
func TestOathswornVampireIgnoresLifeLostEarlierTurnsAndOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, id := awbVampireInGraveyard(t, g)
	b06AddMana(me, "B", "B")
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, 4)
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Fatal("life lost or an opponent's gain opened the graveyard")
	}

	// Gain life, then move to the next turn of this seat's: the tally resets.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	turn := g.Turn.Seq
	for g.Turn.Seq == turn {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Fatal("life gained on an earlier turn opened the graveyard")
	}
}

// A hand cast is never restricted by the graveyard clause.
func TestOathswornVampireHandCastIsNotRestricted(t *testing.T) {
	g := newCatalogGame(t)
	id := castCatalogSpell(t, g, "Oathsworn Vampire", "Creature — Vampire Knight", oathswornVampireOracle, nil)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the hand cast did not resolve")
	}
	if !card.Tapped {
		t.Error("the Vampire enters tapped from the hand too")
	}
}

// --- Phyresis ---------------------------------------------------------

func TestPhyresisGivesInfectWhileAttached(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectiveAbilitiesContain(t, g, bear, "infect") {
		t.Fatal("the Bear has infect before the Aura")
	}
	aura := enchant(t, g, "Phyresis", phyresisOracle, bear)
	if !effectiveAbilitiesContain(t, g, bear, "infect") {
		t.Fatal("the enchanted creature has no infect")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("destroy the Aura: %v", err)
		}
	})
	if effectiveAbilitiesContain(t, g, bear, "infect") {
		t.Error("infect outlived the Aura")
	}
}

// --- Pontiff of Blight ------------------------------------------------

// awbSettle drives the stack to empty, answering each extort offer with
// pay(i) for the i-th offer (adding the {B} when it pays). It returns
// how many offers were made.
func awbSettle(t *testing.T, g *game.Game, me *game.Player, pay func(i int) bool) int {
	t.Helper()
	offers := 0
	for i := 0; i < 96; i++ {
		if ch := triggerOrderPrompt(g); ch != nil {
			if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if hasPayUnlessFor(g, me.ID) {
			p := pay(offers)
			offers++
			if p {
				me.ManaPool.AddMana(game.ManaToken{Color: "B"})
			}
			answerPayUnless(t, g, me.ID, p)
			continue
		}
		if stackFullyEmpty(g) {
			return offers
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
	return offers
}

func awbPontiffBoard(t *testing.T, g *game.Game, others int) (*game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	pontiff := pushPermanentForTest(g, me.ID, "Pontiff of Blight", pontiffOfBlightOracle, "Creature — Zombie Cleric")
	var out []uuid.UUID
	for i := 0; i < others; i++ {
		out = append(out, gaPut(g, me.ID, "Bear", "Creature — Bear", ""))
	}
	return me, pontiff, out
}

// Every other creature has its own extort, and each instance (the
// Pontiff's own included) triggers and is paid for separately.
func TestPontiffOfBlightEachInstanceOffersItsOwnExtort(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := awbPontiffBoard(t, g, 2)
	lives := b17Life(g)

	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	offers := awbSettle(t, g, me, func(i int) bool { return i < 2 })
	if offers != 3 {
		t.Fatalf("%d extort offers, want 3 (the Pontiff and two creatures)", offers)
	}
	opponents := len(g.Seats) - 1
	for i, p := range g.Seats {
		want := lives[i] - 2
		if i == 0 {
			want = lives[i] + 2*opponents
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d (two paid, one declined)", i, p.Life, want)
		}
	}
}

// Declining every offer changes nothing.
func TestPontiffOfBlightDeclinedExtortChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := awbPontiffBoard(t, g, 1)
	lives := b17Life(g)
	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	if offers := awbSettle(t, g, me, func(int) bool { return false }); offers != 2 {
		t.Fatalf("%d offers, want 2", offers)
	}
	for i, p := range g.Seats {
		if p.Life != lives[i] {
			t.Errorf("seat %d life %d, want %d", i, p.Life, lives[i])
		}
	}
}

// Two Pontiffs: a creature has two granted extorts, and the other
// Pontiff is "another creature" for the first.
func TestPontiffOfBlightTwoPontiffsStackTheGrant(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := awbPontiffBoard(t, g, 1)
	pushPermanentForTest(g, me.ID, "Pontiff of Blight", pontiffOfBlightOracle, "Creature — Zombie Cleric")
	// Pontiff A: own + granted by B; Pontiff B: own + granted by A; the
	// Bear: granted by both.
	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	if offers := awbSettle(t, g, me, func(int) bool { return false }); offers != 6 {
		t.Fatalf("%d offers, want 6", offers)
	}
}

// The drain gains only the life actually lost: an opponent whose life
// total can't change gives nothing.
func TestPontiffOfBlightGainsOnlyTheLifeActuallyLost(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := awbPontiffBoard(t, g, 0)
	pushPermanentForTest(g, g.Seats[1].ID, "Platinum Emperion", platinumEmperionOracle, "Artifact Creature — Golem")
	lives := b17Life(g)

	e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
	awbSettle(t, g, me, func(int) bool { return true })

	if got := me.Life - lives[0]; got != 2 {
		t.Errorf("gained %d, want 2 (one opponent lost nothing)", got)
	}
	if g.Seats[1].Life != lives[1] {
		t.Errorf("the locked opponent's life changed: %d -> %d", lives[1], g.Seats[1].Life)
	}
	if g.Seats[2].Life != lives[2]-1 || g.Seats[3].Life != lives[3]-1 {
		t.Errorf("the other opponents did not each lose 1: %d %d", g.Seats[2].Life, g.Seats[3].Life)
	}
}

// The grant is the creature's: the Pontiff leaving takes it from every
// creature, and a creature that loses all abilities loses its extort.
func TestPontiffOfBlightGrantEndsWithThePontiffOrTheCreatureAbilities(t *testing.T) {
	t.Run("pontiff leaves", func(t *testing.T) {
		g := newCatalogGame(t)
		me, pontiff, _ := awbPontiffBoard(t, g, 2)
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(pontiff); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
		e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
		if offers := awbSettle(t, g, me, func(int) bool { return false }); offers != 0 {
			t.Errorf("%d offers after the Pontiff left, want 0", offers)
		}
	})
	t.Run("creature loses abilities", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _, others := awbPontiffBoard(t, g, 1)
		// Casting the Aura is a spell cast: both extort offers are declined.
		castCatalogSpell(t, g, "Darksteel Mutation", auraTypeLine, darksteelMutationOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: others[0]}})
		awbSettle(t, g, me, func(int) bool { return false })
		e2CastSpellOfColor(t, g, "Some Spell", "Sorcery", "U")
		if offers := awbSettle(t, g, me, func(int) bool { return false }); offers != 1 {
			t.Errorf("%d offers, want 1 (only the Pontiff's own)", offers)
		}
	})
}

// --- Timeline Culler --------------------------------------------------

func awbCullerCard(owner *game.Player) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Timeline Culler", TypeLine: "Creature — Drix Warlock",
		ManaCost: "{B}{B}", OracleID: timelineCullerOracle, Owner: owner.ID, Controller: owner.ID,
		Power: 2, Toughness: 2,
	}
}

func TestTimelineCullerWarpFromHandOrGraveyardPaysManaAndLife(t *testing.T) {
	for _, zone := range []string{"hand", "graveyard"} {
		t.Run(zone, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			c := awbCullerCard(me)
			if zone == "hand" {
				me.Hand.PushTop(c)
			} else {
				me.Graveyard.PushTop(c)
			}
			advanceToMain(t, g)
			key := "warp"
			params := game.CastSpellParams{Strict: true, AlternativeCost: key}
			if zone == "graveyard" {
				params.FromZone = "graveyard"
				params.AlternativeCost = "warp-graveyard"
			}
			life := me.Life

			// Short on mana: refused, life untouched.
			if err := g.CastSpell(me.ID, c.InstanceID, params); err == nil {
				t.Fatal("warp with no mana")
			}
			b06AddMana(me, "B")
			if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
				t.Fatalf("warp cast: %v", err)
			}
			if me.Life != life-2 {
				t.Errorf("life %d, want %d: warp costs 2 life", me.Life, life-2)
			}
			if me.ManaPool.Total() != 0 {
				t.Errorf("pool %v, want {B} spent", me.ManaPool)
			}
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(c.InstanceID) {
				t.Fatal("the warped creature did not enter")
			}
			if !effectiveAbilitiesContain(t, g, c.InstanceID, "haste") {
				t.Error("Timeline Culler has haste")
			}

			// Exiled at the next end step, castable from exile on a later turn only.
			advanceThroughEndStep(t, g)
			if !g.Exile.Contains(c.InstanceID) && !awbCardInExile(g, "Timeline Culler") {
				t.Fatal("the creature was not exiled at the next end step")
			}
		})
	}
}

func awbCardInExile(g *game.Game, name string) bool {
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

// Life can't be paid below what the player has, and a player who
// can't pay it is not offered warp.
func TestTimelineCullerWarpNeedsTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := awbCullerCard(me)
	me.Hand.PushTop(c)
	advanceToMain(t, g)
	me.Life = 1
	b06AddMana(me, "B")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, AlternativeCost: "warp"}); err == nil {
		t.Fatal("warp paid 2 life at 1 life")
	}
	if me.Life != 1 {
		t.Errorf("life %d, want 1", me.Life)
	}
}

// The graveyard opens only through warp: the printed {B}{B}, with or
// without the hand's warp key, is refused.
func TestTimelineCullerGraveyardCastAtOrdinaryPriceIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := awbCullerCard(me)
	me.Graveyard.PushTop(c)
	advanceToMain(t, g)
	b06AddMana(me, "B", "B")

	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast from the graveyard at the ordinary price")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{FromZone: "graveyard", Strict: true, AlternativeCost: "warp"}); err == nil {
		t.Fatal("the hand's warp offer was claimed from the graveyard")
	}
	if !me.Graveyard.Contains(c.InstanceID) {
		t.Fatal("a refused cast left the graveyard")
	}
}

// After a warp the creature is in exile and castable on a LATER turn
// for its printed cost, and not the turn it was exiled.
func TestTimelineCullerWarpedCreatureRecastsFromExileOnALaterTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := awbCullerCard(me)
	me.Graveyard.PushTop(c)
	advanceToMain(t, g)
	b06AddMana(me, "B")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "warp-graveyard"}); err != nil {
		t.Fatalf("warp: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceThroughEndStep(t, g)
	if !awbCardInExile(g, "Timeline Culler") {
		t.Fatal("not exiled at the end step")
	}
	var exiledID uuid.UUID
	for _, e := range g.Exile.Cards {
		if e.Name == "Timeline Culler" {
			exiledID = e.InstanceID
		}
	}
	// Same turn: not yet.
	if err := g.CastSpell(me.ID, exiledID, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Fatal("cast from exile on the turn it was exiled")
	}
	// A later turn of the same seat: for the printed cost.
	turn := g.Turn.Seq
	for g.Turn.Seq == turn || g.Turn.ActiveSeat != me.Seat || g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	b06AddMana(me, "B", "B")
	life := me.Life
	if err := g.CastSpell(me.ID, exiledID, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("recast from exile: %v", err)
	}
	if me.Life != life {
		t.Errorf("the ordinary recast paid life: %d -> %d", life, me.Life)
	}
}

// --- Tragic Slip -----------------------------------------------------

func TestTragicSlipShrinksByOneWithoutMorbid(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Tragic Slip", "Instant", tragicSlipOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p, th := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || th != 2 {
		t.Errorf("%d/%d, want 2/2", p, th)
	}
}

func TestTragicSlipIsMorbidWhenACreatureDiedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 3, 3)
	victim := b12Creature(g, g.Seats[2].ID, "Victim", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(victim); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	castCatalogSpell(t, g, "Tragic Slip", "Instant", tragicSlipOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 3/3 under -13/-13 should have died")
	}
}

// A creature dying in response turns morbid on.
func TestTragicSlipGainsMorbidInResponse(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 5, 5)
	victim := b12Creature(g, g.Seats[2].ID, "Victim", "Creature — Bear", 1, 1)
	castCatalogSpell(t, g, "Tragic Slip", "Instant", tragicSlipOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	// In response, a creature dies.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(victim); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("morbid should have been on when the Slip resolved: the 5/5 survived")
	}
}

// --- Yawgmoth's Vile Offering ---------------------------------------

func awbOffering(t *testing.T, g *game.Game, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Yawgmoth's Vile Offering", TypeLine: "Legendary Sorcery",
		OracleID: vileOfferingOracle, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets})
}

func awbLegend(g *game.Game, owner uuid.UUID) {
	b12Creature(g, owner, "Yawgmoth", "Legendary Creature — Human Wizard", 2, 4)
}

func TestVileOfferingRequiresALegendaryCreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := awbOffering(t, g, nil); err == nil {
		t.Fatal("cast with no legendary creature or planeswalker")
	}
	g2 := newCatalogGame(t)
	pushWalkerForTest(g2, g2.Seats[g2.Turn.ActiveSeat].ID, "Some Walker", "", 3)
	if _, err := awbOffering(t, g2, nil); err != nil {
		t.Fatalf("a legendary planeswalker should open the gate: %v", err)
	}
}

// Both clauses: reanimate under your control, destroy, exile itself.
func TestVileOfferingReanimatesDestroysAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	awbLegend(g, me.ID)
	dead := b17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)

	id, err := awbOffering(t, g, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCard(g, dead)
	if !ok || card.Controller != me.ID {
		t.Errorf("the creature card should be on the battlefield under your control (ok=%v)", ok)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the second target should have been destroyed")
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("Yawgmoth's Vile Offering exiles itself")
	}
}

// Choosing zero targets is legal and resolves: it exiles itself.
func TestVileOfferingWithZeroTargetsResolvesAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	awbLegend(g, g.Seats[0].ID)
	id, err := awbOffering(t, g, nil)
	if err != nil {
		t.Fatalf("cast with no targets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("a spell with no targets resolves, and exiles itself")
	}
}

// One target only (either clause) is legal.
func TestVileOfferingWithOneTargetOnlyDoesThatHalf(t *testing.T) {
	t.Run("only the destroy", func(t *testing.T) {
		g := newCatalogGame(t)
		awbLegend(g, g.Seats[0].ID)
		victim := b12Creature(g, g.Seats[1].ID, "Victim", "Creature — Bear", 2, 2)
		id, err := awbOffering(t, g, []game.TargetRef{{Kind: game.TargetCard, ID: victim, Slot: 1}})
		if err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) || !g.Exile.Contains(id) {
			t.Errorf("victim on battlefield=%v, offering exiled=%v", g.Battlefield.Contains(victim), g.Exile.Contains(id))
		}
	})
	t.Run("only the reanimation", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		awbLegend(g, me.ID)
		dead := b17GraveyardCard(g.Seats[1], "Dead Bear", "Creature — Bear", "{1}{G}")
		id, err := awbOffering(t, g, []game.TargetRef{{Kind: game.TargetCard, ID: dead, Slot: 0}})
		if err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if c, ok := battlefieldCard(g, dead); !ok || c.Controller != me.ID {
			t.Errorf("not reanimated under your control (ok=%v)", ok)
		}
		if !g.Exile.Contains(id) {
			t.Error("not exiled")
		}
	})
}

// One chosen target going illegal turns only its half off: the other
// half and the exile still happen.
func TestVileOfferingOneIllegalClauseStillDoesTheOther(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	awbLegend(g, me.ID)
	dead := b17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	id, err := awbOffering(t, g, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	// In response the destroy target is exiled.
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(victim); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, dead); !ok || c.Controller != me.ID {
		t.Errorf("the reanimation half should still happen (ok=%v)", ok)
	}
	if !g.Exile.Contains(id) {
		t.Error("it resolved, so it exiles itself")
	}
}

// Every chosen target illegal: it does not resolve and goes to the
// graveyard, not exile.
func TestVileOfferingAllChosenTargetsIllegalFizzlesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	awbLegend(g, me.ID)
	dead := b17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	id, err := awbOffering(t, g, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(dead); err != nil {
			t.Fatalf("exile dead: %v", err)
		}
		if err := g.ExileCardForEffect(victim); err != nil {
			t.Fatalf("exile victim: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(id) {
		t.Error("a fizzled spell goes to the graveyard, not exile")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the countered-by-rules Offering should be in the graveyard")
	}
}

// A reanimated creature whose entry asks a question holds the rest of
// the spell until it is answered.
func TestVileOfferingReplacementPausedReanimationWaitsForTheAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	awbLegend(g, me.ID)
	dead := b17GraveyardCard(opp, "Dead Bear", "Creature — Bear", "{1}{G}")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches:       []game.EventKind{game.EventZoneMove},
			EntryLifeCost: 2,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.CardID == dead &&
					ev.OldZone == game.ZoneGraveyard && ev.NewZone == game.ZoneBattlefield
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(*game.ReplacementEvent, *game.Game, *game.Card) uuid.UUID { return me.ID },
			Label:      "test: pay 2 life or the Bear enters tapped",
		})
	})
	id, err := awbOffering(t, g, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the reanimated creature's entry question was not asked")
	}
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the destroy ran while the reanimation was still paused")
	}
	answerEntryPayLife(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	if c, ok := battlefieldCard(g, dead); !ok || c.Controller != me.ID || c.Tapped {
		t.Errorf("the creature should be on the battlefield, untapped, under your control (ok=%v)", ok)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the destroy should run once the answer arrives")
	}
	if !g.Exile.Contains(id) {
		t.Error("the Offering should end in exile")
	}
}
