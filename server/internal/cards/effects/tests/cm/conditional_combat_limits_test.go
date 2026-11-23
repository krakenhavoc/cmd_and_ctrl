package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Mirri, Weatherlight Duelist ----------------------------------------

func pushMirri(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Mirri, Weatherlight Duelist", "Legendary Creature — Cat Warrior", effectstest.MirriWeatherlightDuelistOracle, 3, 2)
}

// TestMirriLimitsEachOpponentToOneBlockerOnHerAttack — four seats. Mirri
// attacks one opponent and two other creatures attack two more. Before
// her trigger resolves nothing is limited; once it has, every opponent
// may block with exactly one creature, counted on their own, and the
// enumerator offers exactly what the verb accepts for each of them.
func TestMirriLimitsEachOpponentToOneBlockerOnHerAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mirri := pushMirri(g, me.ID)
	ogreA := effectstest.B12Creature(g, me.ID, "Ogre A", "Creature — Ogre", 4, 4)
	ogreB := effectstest.B12Creature(g, me.ID, "Ogre B", "Creature — Ogre", 4, 4)
	p1a := effectstest.B12Creature(g, p1.ID, "P1 Bear A", "Creature — Bear", 2, 2)
	p1b := effectstest.B12Creature(g, p1.ID, "P1 Bear B", "Creature — Bear", 2, 2)
	p2a := effectstest.B12Creature(g, p2.ID, "P2 Bear A", "Creature — Bear", 2, 2)
	p2b := effectstest.B12Creature(g, p2.ID, "P2 Bear B", "Creature — Bear", 2, 2)
	if !game.HasKeyword(effectstest.MustBattlefieldCard(t, g, mirri), "first strike") {
		t.Error("Mirri has first strike")
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []game.AttackDeclaration{{Attacker: mirri, Target: p1.ID}, {Attacker: ogreA, Target: p1.ID}, {Attacker: ogreB, Target: p2.ID}} {
		if err := g.DeclareAttacker(d.Attacker, d.Target); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.TriggerOnStack(g, mirri) == nil {
		t.Fatal("Mirri's attack trigger is not on the stack")
	}
	if n := effectstest.ScopedBlockRuleCount(g); n != 0 {
		t.Fatalf("%d block rules before the trigger resolved", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ScopedBlockRuleCount(g); n != 1 {
		t.Fatalf("%d block rules after the trigger resolved, want 1", n)
	}
	p2Late := effectstest.B12Creature(g, p2.ID, "P2 Flashed-in Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)

	// P1: two blockers at once is refused; one is legal; then nothing
	// more is offered, and the verb agrees.
	if clAgreeOnBlocks(t, g, p1.ID) == 0 {
		t.Fatal("P1 is offered no block at all")
	}
	refused := effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
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
	effectstest.BrRefusal(t, g.DeclareBlocker(p2b, ogreB), game.BlockReasonDeclarationLimit)
	// A creature that arrived after the trigger resolved is bound too:
	// the rule is about players, not a snapshot of creatures.
	effectstest.BrRefusal(t, g.DeclareBlocker(p2Late, ogreB), game.BlockReasonDeclarationLimit)
	if n := clAgreeOnBlocks(t, g, p2.ID); n != 0 {
		t.Errorf("P2 is offered %d blocks past its one", n)
	}
}

// TestMirriBlockLimitEndsWithHerTurn — "this combat": the rule is gone
// by the next player's turn, where the same two creatures may block
// together.
func TestMirriBlockLimitEndsWithHerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, p1, p2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mirri := pushMirri(g, me.ID)
	effectstest.DeclareAttack(t, g, p1.ID, mirri)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScopedBlockRuleCount(g) != 1 {
		t.Fatal("Mirri's trigger registered no rule")
	}

	// P1's turn: its creature attacks P2, who double-blocks.
	atk := effectstest.B12Creature(g, p1.ID, "Ogre", "Creature — Ogre", 4, 4)
	a := effectstest.B12Creature(g, p2.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, p2.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.ScopedBlockRuleCount(g) != 0 {
		t.Fatal("Mirri's rule outlived her turn")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, p2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
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
		g := effectstest.NewCatalogGame(t)
		me, mirriSeat, other := g.Seats[0], g.Seats[1], g.Seats[2]
		mirri := pushMirri(g, mirriSeat.ID)
		effectstest.MustBattlefieldCard(t, g, mirri).Tapped = tapped
		var bears []uuid.UUID
		for i := 0; i < 3; i++ {
			bears = append(bears, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if at := effectstest.ClTargets(effectstest.ClAgreeOnAttacks(t, g, me.ID)); at[mirriSeat.ID] != 3 {
			t.Fatalf("tapped=%v: before any attack, %d attacks at Mirri's controller, want 3", tapped, at[mirriSeat.ID])
		}
		if err := g.DeclareAttacker(bears[0], mirriSeat.ID); err != nil {
			t.Fatalf("tapped=%v: the first attacker: %v", tapped, err)
		}
		at := effectstest.ClTargets(effectstest.ClAgreeOnAttacks(t, g, me.ID))
		if !tapped {
			if at[mirriSeat.ID] != 2 {
				t.Errorf("untapped Mirri limited the attack: %v", at)
			}
			continue
		}
		if at[mirriSeat.ID] != 0 || at[other.ID] != 2 {
			t.Errorf("tapped Mirri: offers after one attacker %v", at)
		}
		le := effectstest.ClAttackRefusal(t, g.DeclareAttacker(bears[1], mirriSeat.ID))
		if le.Source != mirri || le.Defender != mirriSeat.ID || le.Max != 1 {
			t.Errorf("refusal = %+v", le)
		}
		if err := g.DeclareAttacker(bears[1], other.ID); err != nil {
			t.Fatalf("another opponent may still be attacked: %v", err)
		}
	}
}

// TestEternalWandererAllowsOneAttackerAtHerself — one creature may attack
// the Wanderer each combat; her controller and another planeswalker
// they control may be attacked by any number, and the enumerator
// offers exactly what the verb accepts at every target.
func TestEternalWandererAllowsOneAttackerAtHerself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wanderer := effectstest.PushWanderer(g, opp.ID)
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Another Walker", "", 4)
	var bears []uuid.UUID
	for i := 0; i < 4; i++ {
		bears = append(bears, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)

	if at := effectstest.ClTargets(effectstest.ClAgreeOnAttacksAnywhere(t, g, me.ID)); at[wanderer] != 4 || at[opp.ID] != 4 || at[walker] != 4 {
		t.Fatalf("before any attack every creature is offered everywhere: %v", at)
	}
	if err := g.DeclareAttacker(bears[0], wanderer); err != nil {
		t.Fatalf("one attacker at the Wanderer: %v", err)
	}
	at := effectstest.ClTargets(effectstest.ClAgreeOnAttacksAnywhere(t, g, me.ID))
	if at[wanderer] != 0 || at[opp.ID] != 3 || at[walker] != 3 {
		t.Errorf("offers after one attacker at the Wanderer: %v", at)
	}
	le := effectstest.ClAttackRefusal(t, g.DeclareAttacker(bears[1], wanderer))
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := effectstest.PushWanderer(g, me.ID)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Samurai"); n != 1 {
		t.Fatalf("%d Samurai tokens, want 1", n)
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Samurai" && effects.IsToken(*c) {
			if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !game.HasKeyword(c, "double strike") {
				t.Errorf("Samurai is %d/%d double strike=%v", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(c, "double strike"))
			}
		}
	}
	if l := effectstest.LoyaltyOf(g, pw); l != 5 {
		t.Errorf("loyalty %d after a 0 ability, want 5", l)
	}
}

// TestEternalWandererPlusOneBlinksAtTheNextEndStep — the +1 exiles an
// opponent's artifact, and the delayed trigger returns it under its
// owner's control at the beginning of its OWNER's next end step (#1538:
// not the Wanderer controller's end step, which came first).
func TestEternalWandererPlusOneBlinksAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	pw := effectstest.PushWanderer(g, me.ID)
	relic := effectstest.B12Push(g, opp.ID, "Relic", "Artifact", "", 0, 0)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(relic)}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyOf(g, pw) != 6 {
		t.Errorf("loyalty %d after +1, want 6", effectstest.LoyaltyOf(g, pw))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Relic" {
			t.Fatal("the Relic is still on the battlefield")
		}
	}
	effectstest.AdvanceToEndStepOf(t, g, opp.Seat)
	effectstest.PassPriorityAroundTable(t, g)
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
		t.Error("the Relic did not return at its owner's end step")
	}
}

// TestEternalWandererPlusOneWithNoTarget — "up to one": no target is a
// legal activation that only adds loyalty.
func TestEternalWandererPlusOneWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pw := effectstest.PushWanderer(g, me.ID)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1 with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoyaltyOf(g, pw) != 6 {
		t.Errorf("loyalty %d, want 6", effectstest.LoyaltyOf(g, pw))
	}
}

// TestEternalWandererMinusFourKeepsOneCreatureEach — the −4: the
// Wanderer's controller picks one creature on every board that has
// one, and every other creature is sacrificed; noncreatures stay.
func TestEternalWandererMinusFourKeepsOneCreatureEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	pw := effectstest.PushWanderer(g, me.ID)
	var myA, myB, theirA, theirB, theirRelic uuid.UUID
	g.WithWriteLock(func() {
		myA = effectstest.PushNamedPermanent(g, me.ID, "My Bear A", "Creature — Bear")
		myB = effectstest.PushNamedPermanent(g, me.ID, "My Bear B", "Creature — Bear")
		theirA = effectstest.PushNamedPermanent(g, opp.ID, "Their Bear A", "Creature — Bear")
		theirB = effectstest.PushNamedPermanent(g, opp.ID, "Their Bear B", "Creature — Bear")
		theirRelic = effectstest.PushNamedPermanent(g, opp.ID, "Their Relic", "Artifact")
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	if effectstest.LoyaltyOf(g, pw) != 1 {
		t.Errorf("loyalty %d after −4, want 1", effectstest.LoyaltyOf(g, pw))
	}
}
