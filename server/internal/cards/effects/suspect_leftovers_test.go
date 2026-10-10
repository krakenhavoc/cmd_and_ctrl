package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// suspect_leftovers_test.go — #2733: the four suspect cards #2698 left
// out (Airtight Alibi, Frantic Scapegoat, Nelly Borca, Hot Pursuit), and
// the two Impetus Auras whose goad caveat the same change retires.

const (
	slAirtightAlibi    = "cb666ea8-55e3-4e27-ae6d-806677dfc17a"
	slFranticScapegoat = "d181310f-404b-4f26-8024-b7b537d1fd90"
	slNellyBorca       = "7ef5b2b6-86da-4e4a-9456-dcbb878936c4"
	slHotPursuit       = "00bf9859-d5bf-455a-a4be-965ea4250c7e"
	slShinyImpetus     = "aae76e5c-f5e0-4d18-b465-e6a829be908a"
)

func TestSuspectLeftoversAreFull(t *testing.T) {
	for _, id := range []string{slAirtightAlibi, slFranticScapegoat, slNellyBorca, slHotPursuit, slShinyImpetus, sbIncriminatingImp} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) > 0 {
			t.Errorf("%s ships %v with caveats %q, want full", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// slGoaded reports whether the creature reads as goaded, by anyone.
func slGoaded(t *testing.T, g *game.Game, id uuid.UUID) (bool, []uuid.UUID) {
	t.Helper()
	var goaded bool
	var by []uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("%s not found", id)
		}
		goaded, by = c.Goaded(), c.AllGoaders()
	})
	return goaded, by
}

// --- Airtight Alibi ------------------------------------------------

func TestAirtightAlibiUntapsHexproofsAndClearsTheSuspicion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := sbBear(g, me.ID)
	sbSuspect(g, host)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(host) })
	auraCast(t, g, "Airtight Alibi", slAirtightAlibi, host)
	if sbSuspected(g, host) {
		t.Error("the enchanted creature is still suspected")
	}
	c, _ := battlefieldCard(g, host)
	if c.Tapped {
		t.Error("the enchanted creature is still tapped")
	}
	assertKeywords(t, g, host, "hexproof")
	if got := effectivePower(t, g, host); got != 4 {
		t.Errorf("host power %d, want 4 (+2/+2)", got)
	}
}

func TestAirtightAlibiStopsTheCreatureBecomingSuspected(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := sbBear(g, me.ID)
	aura := auraCast(t, g, "Airtight Alibi", slAirtightAlibi, host)
	if got := auraRestrictions(t, g, host); got&game.CantBecomeSuspected == 0 {
		t.Fatalf("restrictions %b, want can't become suspected", got)
	}
	sbSuspect(g, host)
	if sbSuspected(g, host) {
		t.Fatal("a creature under Airtight Alibi became suspected")
	}
	// The restriction is the Aura's: it goes with the Aura.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	passPriorityAroundTable(t, g)
	sbSuspect(g, host)
	if !sbSuspected(g, host) {
		t.Error("with the Aura gone the creature still can't become suspected")
	}
}

// --- Frantic Scapegoat ---------------------------------------------

// slEnterTogether puts n Goblin tokens onto the battlefield under
// `owner` in one simultaneous entry.
func slEnterTogether(t *testing.T, g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() {
		if err := (CreateToken{Controller: owner, Template: TokenCard("1/1 red Goblin"), N: n}).Apply(ctxFor(g, &game.StackItem{})); err != nil {
			t.Fatalf("CreateToken: %v", err)
		}
	})
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestFranticScapegoatSuspectsItselfWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	id := sbCast(t, g, "Frantic Scapegoat", slFranticScapegoat)
	if !sbSuspected(g, id) {
		t.Fatal("Frantic Scapegoat is not suspected after its enters trigger")
	}
	assertKeywords(t, g, id, "haste", "menace")
}

func TestFranticScapegoatPassesTheSuspicionToOneThatEnteredTogether(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goat := b27Push(g, me.ID, "Frantic Scapegoat", "Creature — Goat", slFranticScapegoat, "{R}", 1, 1, "R")
	sbSuspect(g, goat)
	already := sbBear(g, me.ID) // on the battlefield before: not one of them
	tokens := slEnterTogether(t, g, me.ID, 3)
	sbSuspect(g, tokens[2]) // can't become suspected again: not offered
	slEnterTogether(t, g, opp.ID, 1)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("no choice of a creature to suspect: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1 (\"you may suspect one\")", pick.ChooseMin, pick.ChooseMax)
	}
	if len(pick.ChooseCards) != 2 || !hasID(pick.ChooseCards, tokens[0]) || !hasID(pick.ChooseCards, tokens[1]) {
		t.Errorf("offered %v, want exactly the two unsuspected tokens %v", pick.ChooseCards, tokens[:2])
	}
	if hasID(pick.ChooseCards, already) || hasID(pick.ChooseCards, goat) {
		t.Error("a creature that did not enter in the batch was offered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{tokens[1]}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sbSuspected(g, tokens[1]) {
		t.Error("the chosen creature is not suspected")
	}
	if sbSuspected(g, goat) {
		t.Error("Frantic Scapegoat is still suspected after passing the blame")
	}
}

func TestFranticScapegoatDecliningKeepsItSuspected(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goat := b27Push(g, me.ID, "Frantic Scapegoat", "Creature — Goat", slFranticScapegoat, "{R}", 1, 1, "R")
	sbSuspect(g, goat)
	tokens := slEnterTogether(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no choice of a creature to suspect")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sbSuspected(g, goat) || sbSuspected(g, tokens[0]) {
		t.Error("declining moved the suspicion")
	}
}

func TestFranticScapegoatSkipsACreatureThatCantBecomeSuspected(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goat := b27Push(g, me.ID, "Frantic Scapegoat", "Creature — Goat", slFranticScapegoat, "{R}", 1, 1, "R")
	sbSuspect(g, goat)
	tokens := slEnterTogether(t, g, me.ID, 1)
	// An Airtight Alibi already on the token as the trigger resolves.
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Airtight Alibi", TypeLine: auraTypeLine, OracleID: slAirtightAlibi,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: tokens[0]},
	})
	passPriorityAroundTable(t, g)
	if pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); pick != nil {
		t.Errorf("offered %v, want no choice", pick.ChooseCards)
	}
	if !sbSuspected(g, goat) {
		t.Error("the Goat stopped being suspected with nothing suspected in its place")
	}
}

func TestFranticScapegoatDoesNotTriggerUnsuspected(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Frantic Scapegoat", "Creature — Goat", slFranticScapegoat, "{R}", 1, 1, "R")
	slEnterTogether(t, g, me.ID, 2)
	passPriorityAroundTable(t, g)
	if pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); pick != nil {
		t.Error("an unsuspected Goat asked to pass the blame")
	}
}

// --- Nelly Borca ---------------------------------------------------

func TestNellyBorcaSuspectsThenGoadsEverySuspectedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nelly := b27Push(g, me.ID, "Nelly Borca, Impulsive Accuser", "Legendary Creature — Human Detective", slNellyBorca, "{2}{R}{W}", 2, 4, "R", "W")
	foe := sbBear(g, opp.ID)
	other := sbBear(g, g.Seats[2].ID)
	sbSuspect(g, other)
	plain := sbBear(g, g.Seats[3].ID)
	declareAttack(t, g, opp.ID, nelly)
	pickCard(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	if !sbSuspected(g, foe) {
		t.Fatal("the target is not suspected")
	}
	for _, id := range []uuid.UUID{foe, other} {
		if goaded, by := slGoaded(t, g, id); !goaded || !hasID(by, me.ID) {
			t.Errorf("suspected creature %s is not goaded by Nelly's controller (goaders %v)", id, by)
		}
	}
	if goaded, _ := slGoaded(t, g, plain); goaded {
		t.Error("an unsuspected creature was goaded")
	}
}

func TestNellyBorcaDrawsOncePerOpponentWhoseCreaturesConnect(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	b27Push(g, me.ID, "Nelly Borca, Impulsive Accuser", "Legendary Creature — Human Detective", slNellyBorca, "{2}{R}{W}", 2, 4, "R", "W")
	a1, a2 := sbBear(g, a.ID), sbBear(g, a.ID)
	b1 := sbBear(g, b.ID)
	mine := sbBear(g, me.ID)
	hands := map[uuid.UUID]int{me.ID: me.Hand.Size(), a.ID: a.Hand.Size(), b.ID: b.Hand.Size(), c.ID: c.Hand.Size()}
	// One combat damage step: A's two creatures hit B and C, B's hits C,
	// and Nelly's controller's own creature hits A (not an opponent's
	// creature). Every event names its creature's controller as Actor.
	g.WithWriteLock(func() {
		for _, hit := range []struct{ src, actor, victim uuid.UUID }{
			{a1, a.ID, b.ID}, {a2, a.ID, c.ID}, {b1, b.ID, c.ID}, {mine, me.ID, a.ID},
		} {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: hit.src, Actor: hit.actor, Target: hit.victim, Amount: 2, Combat: true})
		}
		// Damage to Nelly's controller is not damage to an opponent.
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: b1, Actor: b.ID, Target: me.ID, Amount: 2, Combat: true})
	})
	passPriorityAroundTable(t, g)
	want := map[uuid.UUID]int{me.ID: 2, a.ID: 1, b.ID: 1, c.ID: 0}
	for _, p := range []*game.Player{me, a, b, c} {
		if got := p.Hand.Size() - hands[p.ID]; got != want[p.ID] {
			t.Errorf("seat %s drew %d, want %d", p.Name, got, want[p.ID])
		}
	}
}

// --- Hot Pursuit ---------------------------------------------------

// slCastHotPursuit casts Hot Pursuit and answers its enters trigger's
// target.
func slCastHotPursuit(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Hot Pursuit", "Enchantment", slHotPursuit, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, g.Seats[0].ID, target)
	passPriorityAroundTable(t, g)
	return id
}

func TestHotPursuitSuspectsAndGoadsWhileItRemains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := sbBear(g, opp.ID)
	hp := slCastHotPursuit(t, g, foe)
	if !sbSuspected(g, foe) {
		t.Error("the target is not suspected")
	}
	if goaded, by := slGoaded(t, g, foe); !goaded || len(by) != 1 || by[0] != me.ID {
		t.Fatalf("the target is not goaded by Hot Pursuit's controller (goaded %v, by %v)", goaded, by)
	}
	// A goad from a spell ends at the goader's next turn; this one ends
	// only with the enchantment.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hp) })
	passPriorityAroundTable(t, g)
	if goaded, _ := slGoaded(t, g, foe); goaded {
		t.Error("the creature is still goaded after Hot Pursuit left")
	}
	if !sbSuspected(g, foe) {
		t.Error("the suspicion went with the enchantment; it is a status")
	}
}

func TestHotPursuitGoadsATargetThatCantBecomeSuspected(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	foe := sbBear(g, opp.ID)
	sbSuspect(g, foe) // already suspected: the suspect does nothing, the goad still holds
	slCastHotPursuit(t, g, foe)
	if goaded, _ := slGoaded(t, g, foe); !goaded {
		t.Error("the creature is not goaded")
	}
}

func TestHotPursuitTakesEveryGoadedOrSuspectedCreatureOnceTwoPlayersHaveLost(t *testing.T) {
	for _, lost := range []int{1, 2} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b27Push(g, me.ID, "Hot Pursuit", "Enchantment", slHotPursuit, "{1}{R}", 0, 0, "R")
		suspected := sbBear(g, opp.ID)
		sbSuspect(g, suspected)
		goaded := sbBear(g, opp.ID)
		g.WithWriteLock(func() { g.GoadForEffect(goaded, g.Seats[2].ID) })
		// An Aura's goad counts too (#2733).
		shiny := sbBear(g, opp.ID)
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(shiny) })
		plain := sbBear(g, opp.ID)
		for i := 0; i < lost; i++ {
			g.Seats[2+i].Eliminated = true
		}
		// Shiny Impetus cast by the active seat onto an opponent's creature.
		auraCast(t, g, "Shiny Impetus", slShinyImpetus, shiny)
		advanceToStepOf(t, g, 0, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		for _, id := range []uuid.UUID{suspected, goaded, shiny} {
			c, _ := battlefieldCard(g, id)
			if want := lost >= 2; (c.Controller == me.ID) != want {
				t.Errorf("lost=%d: %s controlled by me = %v, want %v", lost, id, c.Controller == me.ID, want)
			}
			if lost >= 2 {
				if c.Tapped {
					t.Errorf("%s was not untapped", id)
				}
				assertKeywords(t, g, id, "haste")
			}
		}
		if c, _ := battlefieldCard(g, plain); c.Controller == me.ID {
			t.Errorf("lost=%d: a creature neither goaded nor suspected was taken", lost)
		}
	}
}

// --- The Impetus Auras: a static goad reads as goaded --------------

func TestShinyImpetusGoadReadsAsGoaded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := sbBear(g, opp.ID)
	auraCast(t, g, "Shiny Impetus", slShinyImpetus, host)
	goaded, by := slGoaded(t, g, host)
	if !goaded || len(by) != 1 || by[0] != me.ID {
		t.Errorf("goaded %v by %v, want goaded by the Aura's controller", goaded, by)
	}
	c, _ := battlefieldCard(g, host)
	if c.IsGoaded() {
		t.Error("the static goad was written into the marker the snapshot carries")
	}
}
