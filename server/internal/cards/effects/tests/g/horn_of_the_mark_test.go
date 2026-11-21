package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// horn_of_the_mark_test.go — #952's proof card. Horn of the Mark is
// "look at the top five, you may reveal a creature card from among
// them and put it into your hand, put the rest on the bottom in a
// random order": the whole of the primitive in one sentence.

const hornOfTheMarkOracle = "9b836c32-84f0-41ae-b7d9-67b92c743c60"

// hornSwing puts the Horn and `attackers` 2/2s under the active seat,
// attacks `defender` with all of them and locks the declaration in.
func hornSwing(t *testing.T, g *game.Game, me *game.Player, defender uuid.UUID, attackers int) {
	t.Helper()
	effectstest.PushCatalogPermanent(g, me.ID, "Horn of the Mark", "Legendary Artifact", hornOfTheMarkOracle, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for i := 0; i < attackers; i++ {
		id := effectstest.AcPermanent(g, me.ID, "Rider", "Creature — Human Knight", 2, 2)
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker %d: %v", i, err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

// TestHornOfTheMarkTakesACreatureAndBottomsTheRest is the whole card.
func TestHornOfTheMarkTakesACreatureAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := effectstest.HornTopFive(me, effectstest.HornFive)
	hornSwing(t, g, me, opp.ID, 2)

	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("two attackers raised no Horn of the Mark prompt")
	}
	// Only the CREATURE cards of the five are offered, and only one
	// may be taken — "you may reveal A creature card".
	want := map[uuid.UUID]bool{top[1]: true, top[3]: true}
	if len(prompt.ChooseCards) != len(want) {
		t.Fatalf("offered %d cards, want the %d creature cards of the five", len(prompt.ChooseCards), len(want))
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered a card that is not a creature card: %s", id)
		}
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("bounds = [%d,%d], want [0,1] for an optional single take", prompt.ChooseMin, prompt.ChooseMax)
	}
	// The look is PRIVATE: nobody else is asked, and nobody else is
	// shown the five.
	if effectstest.LatestChooseCards(g, opp.ID) != nil {
		t.Error("the opponent was asked about the Horn's look")
	}

	handBefore := me.Hand.Size()
	libBefore := me.Library.Size()
	taken := top[3]
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{taken}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	if !me.Hand.Contains(taken) {
		t.Error("the chosen creature card is not in hand")
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want exactly one card taken", handBefore, me.Hand.Size())
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library %d → %d, want exactly one card gone", libBefore, me.Library.Size())
	}
	// The card taken is REVEALED; the four it was chosen from are not.
	if !effectstest.HornRevealed(g, taken) {
		t.Error("the taken card was not revealed — the card prints \"you may reveal\"")
	}
	for _, id := range top {
		if id != taken && effectstest.HornRevealed(g, id) {
			t.Errorf("a card that was only LOOKED at was revealed: %s", id)
		}
	}
	// The rest went to the bottom: none of them is in the top five any
	// more, and all four are still in the library.
	// Zone.Cards is bottom-first (PushTop appends), so the bottom four
	// are the first four entries.
	bottom := me.Library.Cards[:4]
	rest := map[uuid.UUID]bool{}
	for _, id := range top {
		if id != taken {
			rest[id] = true
		}
	}
	for _, c := range bottom {
		if !rest[c.InstanceID] {
			t.Errorf("%q is on the bottom and was not one of the looked-at cards", c.Name)
		}
		delete(rest, c.InstanceID)
	}
	if len(rest) != 0 {
		t.Errorf("%d of the looked-at cards did not reach the bottom", len(rest))
	}
}

// Declining takes nothing and still bottoms the five. "You may".
func TestHornOfTheMarkMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	top := effectstest.HornTopFive(me, effectstest.HornFive)
	hornSwing(t, g, me, opp.ID, 2)

	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no Horn of the Mark prompt")
	}
	handBefore, libBefore := me.Hand.Size(), me.Library.Size()
	if err := g.ResolveChooseCards(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d → %d after declining", handBefore, me.Hand.Size())
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library %d → %d after declining — nothing should leave it", libBefore, me.Library.Size())
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:5] {
		bottom[c.InstanceID] = true
	}
	for _, id := range top {
		if !bottom[id] {
			t.Errorf("a looked-at card did not reach the bottom: %s", id)
		}
	}
}

// "TWO OR MORE creatures". One attacker is not a trigger.
func TestHornOfTheMarkNeedsTwoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.HornTopFive(me, effectstest.HornFive)
	hornSwing(t, g, me, opp.ID, 1)

	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("one attacker triggered Horn of the Mark")
	}
}

// "attack A PLAYER": two attackers split across two opponents is not
// two creatures attacking a player, so nothing triggers. The same key
// is what makes three attackers on one player trigger once rather than
// three times.
func TestHornOfTheMarkIsPerDefendingPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	second := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.HornTopFive(me, effectstest.HornFive)

	effectstest.PushCatalogPermanent(g, me.ID, "Horn of the Mark", "Legendary Artifact", hornOfTheMarkOracle, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	a := effectstest.AcPermanent(g, me.ID, "Rider A", "Creature — Human Knight", 2, 2)
	b := effectstest.AcPermanent(g, me.ID, "Rider B", "Creature — Human Knight", 2, 2)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("one attacker per opponent triggered Horn of the Mark")
	}
}

// Three attackers at one player is ONE trigger, not three.
func TestHornOfTheMarkTriggersOncePerDeclaration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.HornTopFive(me, effectstest.HornFive)
	hornSwing(t, g, me, opp.ID, 3)

	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("three attackers raised %d Horn prompts, want 1", n)
	}
}

// hornAttackCases are #4864's three declarations for "two or more
// creatures you control attack a player": each names where the two
// attackers go, given the player, their planeswalker and a battle they
// protect, and whether the ability triggers. Only creatures whose attack
// target is the player itself count (CR 508.3a).
var hornAttackCases = []struct {
	name string
	at   func(player, walker, battle uuid.UUID) [2]uuid.UUID
	want bool
}{
	{"two at the player trigger", func(p, _, _ uuid.UUID) [2]uuid.UUID { return [2]uuid.UUID{p, p} }, true},
	{"one at the player and one at their planeswalker do not", func(p, w, _ uuid.UUID) [2]uuid.UUID { return [2]uuid.UUID{p, w} }, false},
	{"two at their battle do not", func(_, _, b uuid.UUID) [2]uuid.UUID { return [2]uuid.UUID{b, b} }, false},
}

// hornAttackTargets gives `opp` a planeswalker and makes them the
// protector of a battle `other` controls, and returns both.
func hornAttackTargets(g *game.Game, opp, other uuid.UUID) (walker, battle uuid.UUID) {
	walker = effectstest.P1g2Planeswalker(g, opp, 5)
	battle = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Owner: other, Controller: other, ProtectorPlayerID: opp,
		Counters: map[string]int{game.CounterDefense: 5},
	})
	return walker, battle
}

// #4864: an attack on a player's planeswalker or on a battle they protect
// is not an attack on that player, so it does not make up the two.
func TestHornOfTheMarkCountsOnlyAttackersAtThePlayer(t *testing.T) {
	for _, tc := range hornAttackCases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
			effectstest.HornTopFive(me, effectstest.HornFive)
			effectstest.PushCatalogPermanent(g, me.ID, "Horn of the Mark", "Legendary Artifact", hornOfTheMarkOracle, false)
			walker, battle := hornAttackTargets(g, opp.ID, other.ID)
			at := tc.at(opp.ID, walker, battle)
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			for i := range at {
				id := effectstest.AcPermanent(g, me.ID, "Rider", "Creature — Human Knight", 2, 2)
				if err := g.DeclareAttacker(id, at[i]); err != nil {
					t.Fatalf("DeclareAttacker %d: %v", i, err)
				}
			}
			effectstest.LockInAttacks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.LatestChooseCards(g, me.ID) != nil; got != tc.want {
				t.Errorf("Horn of the Mark triggered = %v, want %v", got, tc.want)
			}
		})
	}
}
