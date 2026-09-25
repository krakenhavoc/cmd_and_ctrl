package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_combat_limits_test.go — #1534's proof cards (ADR 0045
// amendment of 2026-09-24, Decision 47). Mirri, Weatherlight Duelist
// carries the per-defender block limit (her attack trigger) and the
// While-gated attack limit (her static); The Eternal Wanderer carries
// the per-permanent attack scope and three loyalty abilities.
//
// Every shape is pinned the way combat_limits_test.go pins #1507's:
// the verb accepts the legal declaration and refuses the illegal one,
// and the enumerator offers exactly what the verb accepts — checked by
// trying every candidate on a clone.

const (
	mirriWeatherlightDuelistOracle = "60bc9cd2-2e03-4889-b1ef-fdab9a9a6d08"
	theEternalWandererOracle       = "20a1671d-e8a4-4cf1-87a7-f2f6319f4b9e"
)

// clAcceptedAttacksAnywhere is clAcceptedAttacks over every attack
// target the engine lists — players AND planeswalkers — so a limit on
// a planeswalker is checked against the verb too.
func clAcceptedAttacksAnywhere(g *game.Game, seat uuid.UUID) []clAttack {
	var out []clAttack
	targets := g.AttackTargetsForEffect(seat)
	for _, c := range g.Battlefield.Cards {
		if c.Controller != seat || !c.IsCreature() || c.AttackingTarget != uuid.Nil {
			continue
		}
		for _, t := range targets {
			if err := g.Clone().DeclareAttacker(c.InstanceID, t.ID); err == nil {
				out = append(out, clAttack{c.InstanceID, t.ID})
			}
		}
	}
	return clSorted(out)
}

// clAgreeOnAttacksAnywhere asserts the enumerator offers exactly what
// the verb accepts at every target, and returns the offer.
func clAgreeOnAttacksAnywhere(t *testing.T, g *game.Game, seat uuid.UUID) []clAttack {
	t.Helper()
	offered, accepted := clOfferedAttacks(t, g, seat), clAcceptedAttacksAnywhere(g, seat)
	if len(offered) != len(accepted) {
		t.Fatalf("the enumerator offers %d attacks, the verb accepts %d:\noffered  %v\naccepted %v", len(offered), len(accepted), offered, accepted)
	}
	for i := range offered {
		if offered[i] != accepted[i] {
			t.Fatalf("the enumerator and the verb disagree:\noffered  %v\naccepted %v", offered, accepted)
		}
	}
	return offered
}

// --- Mirri, Weatherlight Duelist ----------------------------------------

func pushMirri(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Mirri, Weatherlight Duelist", "Legendary Creature — Cat Warrior", mirriWeatherlightDuelistOracle, 3, 2)
}

// TestMirriLimitsEachOpponentToOneBlockerOnHerAttack — four seats. Mirri
// attacks one opponent and two other creatures attack two more. Before
// her trigger resolves nothing is limited; once it has, every opponent
// may block with exactly one creature, counted on their own, and the
// enumerator offers exactly what the verb accepts for each of them.
func TestMirriLimitsEachOpponentToOneBlockerOnHerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mirri := pushMirri(g, me.ID)
	ogreA := b12Creature(g, me.ID, "Ogre A", "Creature — Ogre", 4, 4)
	ogreB := b12Creature(g, me.ID, "Ogre B", "Creature — Ogre", 4, 4)
	p1a := b12Creature(g, p1.ID, "P1 Bear A", "Creature — Bear", 2, 2)
	p1b := b12Creature(g, p1.ID, "P1 Bear B", "Creature — Bear", 2, 2)
	p2a := b12Creature(g, p2.ID, "P2 Bear A", "Creature — Bear", 2, 2)
	p2b := b12Creature(g, p2.ID, "P2 Bear B", "Creature — Bear", 2, 2)
	if !game.HasKeyword(mustBattlefieldCard(t, g, mirri), "first strike") {
		t.Error("Mirri has first strike")
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []game.AttackDeclaration{{Attacker: mirri, Target: p1.ID}, {Attacker: ogreA, Target: p1.ID}, {Attacker: ogreB, Target: p2.ID}} {
		if err := g.DeclareAttacker(d.Attacker, d.Target); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	if triggerOnStack(g, mirri) == nil {
		t.Fatal("Mirri's attack trigger is not on the stack")
	}
	if n := scopedBlockRuleCount(g); n != 0 {
		t.Fatalf("%d block rules before the trigger resolved", n)
	}
	passPriorityAroundTable(t, g)
	if n := scopedBlockRuleCount(g); n != 1 {
		t.Fatalf("%d block rules after the trigger resolved, want 1", n)
	}
	p2Late := b12Creature(g, p2.ID, "P2 Flashed-in Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareBlockers)

	// P1: two blockers at once is refused; one is legal; then nothing
	// more is offered, and the verb agrees.
	if clAgreeOnBlocks(t, g, p1.ID) == 0 {
		t.Fatal("P1 is offered no block at all")
	}
	refused := brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: p1a, Attacker: mirri}, {Blocker: p1b, Attacker: ogreA},
	}), game.BlockReasonDeclarationLimit)
	if want := "Each opponent can't block with more than one creature this combat (Mirri, Weatherlight Duelist)."; refused.Sentence(p1.ID) != want {
		t.Errorf("P1 reads %q, want %q", refused.Sentence(p1.ID), want)
	}
	if err := g.DeclareBlocker(p1a, ogreA); err != nil {
		t.Fatalf("P1's one block: %v", err)
	}
	if n := clAgreeOnBlocks(t, g, p1.ID); n != 0 {
		t.Errorf("P1 is offered %d blocks past its one", n)
	}
	// P2's one is its own.
	if clAgreeOnBlocks(t, g, p2.ID) == 0 {
		t.Fatal("P1's block used up P2's")
	}
	if err := g.DeclareBlocker(p2a, ogreB); err != nil {
		t.Fatalf("P2's one block: %v", err)
	}
	brRefusal(t, g.DeclareBlocker(p2b, ogreB), game.BlockReasonDeclarationLimit)
	// A creature that arrived after the trigger resolved is bound too:
	// the rule is about players, not a snapshot of creatures.
	brRefusal(t, g.DeclareBlocker(p2Late, ogreB), game.BlockReasonDeclarationLimit)
	if n := clAgreeOnBlocks(t, g, p2.ID); n != 0 {
		t.Errorf("P2 is offered %d blocks past its one", n)
	}
}

// TestMirriBlockLimitEndsWithHerTurn — "this combat": the rule is gone
// by the next player's turn, where the same two creatures may block
// together.
func TestMirriBlockLimitEndsWithHerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mirri := pushMirri(g, me.ID)
	declareAttack(t, g, p1.ID, mirri)
	passPriorityAroundTable(t, g)
	if scopedBlockRuleCount(g) != 1 {
		t.Fatal("Mirri's trigger registered no rule")
	}

	// P1's turn: its creature attacks P2, who double-blocks.
	atk := b12Creature(g, p1.ID, "Ogre", "Creature — Ogre", 4, 4)
	a := b12Creature(g, p2.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, p2.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, 1)
	if scopedBlockRuleCount(g) != 0 {
		t.Fatal("Mirri's rule outlived her turn")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, p2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: a, Attacker: atk}, {Blocker: b, Attacker: atk}}); err != nil {
		t.Errorf("a double block on another player's turn was refused: %v", err)
	}
}

// TestMirriTappedLimitsAttackersAtHerController — "As long as Mirri is
// tapped, no more than one creature can attack you each combat."
// Untapped, the attacking seat may send everything at her controller;
// tapped, one creature — the others may go at the other opponents, and
// the enumerator offers exactly that.
func TestMirriTappedLimitsAttackersAtHerController(t *testing.T) {
	for _, tapped := range []bool{false, true} {
		g := newCatalogGame(t)
		me, mirriSeat, other := g.Seats[0], g.Seats[1], g.Seats[2]
		mirri := pushMirri(g, mirriSeat.ID)
		mustBattlefieldCard(t, g, mirri).Tapped = tapped
		var bears []uuid.UUID
		for i := 0; i < 3; i++ {
			bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		advanceTo(t, g, game.StepDeclareAttackers)
		if at := clTargets(clAgreeOnAttacks(t, g, me.ID)); at[mirriSeat.ID] != 3 {
			t.Fatalf("tapped=%v: before any attack, %d attacks at Mirri's controller, want 3", tapped, at[mirriSeat.ID])
		}
		if err := g.DeclareAttacker(bears[0], mirriSeat.ID); err != nil {
			t.Fatalf("tapped=%v: the first attacker: %v", tapped, err)
		}
		at := clTargets(clAgreeOnAttacks(t, g, me.ID))
		if !tapped {
			if at[mirriSeat.ID] != 2 {
				t.Errorf("untapped Mirri limited the attack: %v", at)
			}
			continue
		}
		if at[mirriSeat.ID] != 0 || at[other.ID] != 2 {
			t.Errorf("tapped Mirri: offers after one attacker %v", at)
		}
		le := clAttackRefusal(t, g.DeclareAttacker(bears[1], mirriSeat.ID))
		if le.Source != mirri || le.Defender != mirriSeat.ID || le.Max != 1 {
			t.Errorf("refusal = %+v", le)
		}
		if err := g.DeclareAttacker(bears[1], other.ID); err != nil {
			t.Fatalf("another opponent may still be attacked: %v", err)
		}
	}
}

// --- The Eternal Wanderer ------------------------------------------------

func pushWanderer(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := pushWalkerForTest(g, owner, "The Eternal Wanderer", theEternalWandererOracle, 5)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].TypeLine = "Legendary Planeswalker"
		}
	}
	return id
}

// TestEternalWandererAllowsOneAttackerAtHerself — one creature may attack
// the Wanderer each combat; her controller and another planeswalker
// they control may be attacked by any number, and the enumerator
// offers exactly what the verb accepts at every target.
func TestEternalWandererAllowsOneAttackerAtHerself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wanderer := pushWanderer(g, opp.ID)
	walker := pushWalkerForTest(g, opp.ID, "Another Walker", "", 4)
	var bears []uuid.UUID
	for i := 0; i < 4; i++ {
		bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	advanceTo(t, g, game.StepDeclareAttackers)

	if at := clTargets(clAgreeOnAttacksAnywhere(t, g, me.ID)); at[wanderer] != 4 || at[opp.ID] != 4 || at[walker] != 4 {
		t.Fatalf("before any attack every creature is offered everywhere: %v", at)
	}
	if err := g.DeclareAttacker(bears[0], wanderer); err != nil {
		t.Fatalf("one attacker at the Wanderer: %v", err)
	}
	at := clTargets(clAgreeOnAttacksAnywhere(t, g, me.ID))
	if at[wanderer] != 0 || at[opp.ID] != 3 || at[walker] != 3 {
		t.Errorf("offers after one attacker at the Wanderer: %v", at)
	}
	le := clAttackRefusal(t, g.DeclareAttacker(bears[1], wanderer))
	if le.Source != wanderer || le.Max != 1 {
		t.Errorf("refusal = %+v", le)
	}
	if got, want := le.Sentence(me.ID), "No more than one creature can attack The Eternal Wanderer each combat."; got != want {
		t.Errorf("sentence %q, want %q", got, want)
	}
	if err := g.DeclareAttacker(bears[1], opp.ID); err != nil {
		t.Fatalf("the Wanderer's controller: %v", err)
	}
	if err := g.DeclareAttacker(bears[2], opp.ID); err != nil {
		t.Fatalf("the Wanderer's controller, twice: %v", err)
	}
	if err := g.DeclareAttacker(bears[3], walker); err != nil {
		t.Fatalf("another planeswalker: %v", err)
	}
}

// TestEternalWandererZeroMakesADoubleStrikeSamurai — the 0.
func TestEternalWandererZeroMakesADoubleStrikeSamurai(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := pushWanderer(g, me.ID)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Samurai"); n != 1 {
		t.Fatalf("%d Samurai tokens, want 1", n)
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Samurai" && IsToken(*c) {
			if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !game.HasKeyword(c, "double strike") {
				t.Errorf("Samurai is %d/%d double strike=%v", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(c, "double strike"))
			}
		}
	}
	if l := loyaltyOf(g, pw); l != 5 {
		t.Errorf("loyalty %d after a 0 ability, want 5", l)
	}
}

// TestEternalWandererPlusOneBlinksAtTheNextEndStep — the +1 exiles an
// opponent's artifact, and the delayed trigger returns it under its
// owner's control at the next end step (the declared caveat: the next
// one, not its owner's).
func TestEternalWandererPlusOneBlinksAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWanderer(g, me.ID)
	relic := b12Push(g, opp.ID, "Relic", "Artifact", "", 0, 0)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Targets: cardRefs(relic)}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if loyaltyOf(g, pw) != 6 {
		t.Errorf("loyalty %d after +1, want 6", loyaltyOf(g, pw))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Relic" {
			t.Fatal("the Relic is still on the battlefield")
		}
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Relic" {
			back = true
			if c.Controller != opp.ID {
				t.Errorf("the Relic returned under %v, want its owner", c.Controller)
			}
		}
	}
	if !back {
		t.Error("the Relic did not return at the end step")
	}
}

// TestEternalWandererPlusOneWithNoTarget — "up to one": no target is a
// legal activation that only adds loyalty.
func TestEternalWandererPlusOneWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := pushWanderer(g, me.ID)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1 with no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if loyaltyOf(g, pw) != 6 {
		t.Errorf("loyalty %d, want 6", loyaltyOf(g, pw))
	}
}

// TestEternalWandererMinusFourKeepsOneCreatureEach — the −4: the
// Wanderer's controller picks one creature on every board that has
// one, and every other creature is sacrificed; noncreatures stay.
func TestEternalWandererMinusFourKeepsOneCreatureEach(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	pw := pushWanderer(g, me.ID)
	var myA, myB, theirA, theirB, theirRelic uuid.UUID
	g.WithWriteLock(func() {
		myA = pushNamedPermanent(g, me.ID, "My Bear A", "Creature — Bear")
		myB = pushNamedPermanent(g, me.ID, "My Bear B", "Creature — Bear")
		theirA = pushNamedPermanent(g, opp.ID, "Their Bear A", "Creature — Bear")
		theirB = pushNamedPermanent(g, opp.ID, "Their Bear B", "Creature — Bear")
		theirRelic = pushNamedPermanent(g, opp.ID, "Their Relic", "Artifact")
	})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−4: %v", err)
	}
	passPriorityAroundTable(t, g)

	legs := 0
	for _, c := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
		if c == nil {
			continue
		}
		switch c.Kind {
		case game.PendingChoiceOwnPermanents:
			legs++
			if c.ChooseMin != 1 || c.ChooseMax != 1 {
				t.Errorf("own leg bounds %d..%d, want exactly one", c.ChooseMin, c.ChooseMax)
			}
			if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{myA}); err != nil {
				t.Fatalf("ResolveOwnPermanents: %v", err)
			}
		case game.PendingChoiceTheirPermanents:
			legs++
			if c.FromPlayer != opp.ID {
				t.Fatalf("a leg over %s, who has no creatures", c.FromPlayer)
			}
			if err := g.ResolveTheirPermanents(c.ID, me.ID, []uuid.UUID{theirB}); err != nil {
				t.Fatalf("ResolveTheirPermanents: %v", err)
			}
		}
	}
	if legs != 2 {
		t.Fatalf("%d legs, want one per player with a creature (2)", legs)
	}
	on := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		on[c.InstanceID] = true
	}
	for _, id := range []uuid.UUID{myA, theirB, theirRelic} {
		if !on[id] {
			t.Errorf("%s was kept (or is not a creature) and should have survived", id)
		}
	}
	for _, id := range []uuid.UUID{myB, theirA} {
		if on[id] {
			t.Errorf("%s was not chosen and should have been sacrificed", id)
		}
	}
	if loyaltyOf(g, pw) != 1 {
		t.Errorf("loyalty %d after −4, want 1", loyaltyOf(g, pw))
	}
}
