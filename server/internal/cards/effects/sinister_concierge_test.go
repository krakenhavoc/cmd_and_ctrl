package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sinister_concierge_test.go — Sinister Concierge (#5879) on the engine's
// granted suspend (Card.SuspendGranted, #4455). Each test plays the real
// catalog entry.

const scOracle = "9b14c20f-c4ee-42ea-99cd-099b4eb25883"

func TestSinisterConciergeIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(scOracle)
	if !ok {
		t.Fatal("Sinister Concierge is not registered")
	}
	if spec.Name != "Sinister Concierge" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// scBoard seats the Concierge under seat 0 and a vanilla creature under
// seat 1, then destroys the Concierge so its dies trigger is announced.
func scBoard(t *testing.T) (g *game.Game, me, opp *game.Player, concierge, goblin uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	concierge = pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sinister Concierge",
		TypeLine: "Creature — Human Wizard", OracleID: scOracle, ManaCost: "{1}{U}", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID})
	goblin = pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Goblin",
		TypeLine: "Creature — Goblin", ManaCost: "{2}{R}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	p1Destroy(t, g, concierge)
	b04WaitForPick(t, g, me.ID)
	return g, me, opp, concierge, goblin
}

func scCheckSuspended(t *testing.T, g *game.Game, id uuid.UUID, owner uuid.UUID, what string) {
	t.Helper()
	c := gsExiled(g, id)
	if c == nil {
		t.Fatalf("%s is not in exile", what)
	}
	if c.Counters[game.CounterTime] != 3 || !c.SuspendGranted {
		t.Fatalf("%s has %d time counters, suspend granted %v; want 3 and true", what, c.Counters[game.CounterTime], c.SuspendGranted)
	}
	if c.Owner != owner {
		t.Errorf("%s is owned by %s, want %s", what, c.Owner, owner)
	}
}

func scBattlefield(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// Yes, with an opponent's creature targeted: both are exiled with three
// time counters and suspend. Each counts down on its OWNER's upkeep only,
// and after three of them its owner casts it free and it has haste.
func TestSinisterConciergeSuspendsItselfAndAnOpponentsCreatureForTheirOwners(t *testing.T) {
	g, me, opp, concierge, goblin := scBoard(t)
	pickCard(t, g, me.ID, goblin)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	scCheckSuspended(t, g, concierge, me.ID, "the Concierge")
	scCheckSuspended(t, g, goblin, opp.ID, "the target creature")

	// Seat 1's upkeep ticks only the goblin; seat 0's only the Concierge.
	gsNextUpkeep(t, g, 1)
	if n := gsExiled(g, goblin).Counters[game.CounterTime]; n != 2 {
		t.Fatalf("the goblin after its owner's upkeep: %d time counters, want 2", n)
	}
	if n := gsExiled(g, concierge).Counters[game.CounterTime]; n != 3 {
		t.Fatalf("the Concierge after an opponent's upkeep: %d time counters, want 3", n)
	}
	gsNextUpkeep(t, g, 0)
	if n := gsExiled(g, concierge).Counters[game.CounterTime]; n != 2 {
		t.Fatalf("the Concierge after its owner's upkeep: %d time counters, want 2", n)
	}
	if n := gsExiled(g, goblin).Counters[game.CounterTime]; n != 2 {
		t.Fatalf("the goblin after seat 0's upkeep: %d time counters, want 2", n)
	}
	gsNextUpkeep(t, g, 1)
	gsNextUpkeep(t, g, 0)
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) != nil {
		t.Fatal("a cast was offered before the last counter came off")
	}

	// The goblin's last counter: its owner, the opponent, casts it.
	gsNextUpkeep(t, g, 1)
	offer := latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, opp.ID)
	if offer == nil {
		t.Fatal("the goblin's owner was offered no cast")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID) != nil {
		t.Fatal("the Concierge's controller was offered the opponent's card")
	}
	answerKeywordCast(t, g, offer, true)
	passPriorityAroundTable(t, g)
	gob := scBattlefield(g, goblin)
	if gob == nil {
		t.Fatal("the goblin was not cast from exile")
	}
	if gob.Controller != opp.ID {
		t.Errorf("the goblin entered under %s, want its owner %s", gob.Controller, opp.ID)
	}
	if !game.HasKeyword(gob, "haste") {
		t.Error("the goblin cast this way has no haste")
	}

	gsNextUpkeep(t, g, 0)
	offer = latestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID)
	if offer == nil {
		t.Fatal("the Concierge's owner was offered no cast")
	}
	answerKeywordCast(t, g, offer, true)
	passPriorityAroundTable(t, g)
	con := scBattlefield(g, concierge)
	if con == nil {
		t.Fatal("the Concierge was not cast from exile")
	}
	if con.Controller != me.ID || !game.HasKeyword(con, "haste") {
		t.Errorf("the Concierge: controller %s, haste %v", con.Controller, game.HasKeyword(con, "haste"))
	}
	if con.SuspendGranted || con.Counters[game.CounterTime] != 0 {
		t.Errorf("the permanent kept the exiled card's state: %+v", con)
	}
}

// "You may": declined, nothing is exiled, the target included.
func TestSinisterConciergeDeclinedExilesNothing(t *testing.T) {
	g, me, _, concierge, goblin := scBoard(t)
	pickCard(t, g, me.ID, goblin)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(concierge) {
		t.Error("a declined Concierge left the graveyard")
	}
	if !g.Battlefield.Contains(goblin) {
		t.Error("the target was exiled though the Concierge was not")
	}
}

// "If you do": a Concierge that left the graveyard in response cannot be
// exiled by its trigger, so the target is not exiled either.
func TestSinisterConciergeGoneFromTheGraveyardExilesNoTarget(t *testing.T) {
	g, me, _, concierge, goblin := scBoard(t)
	pickCard(t, g, me.ID, goblin)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, concierge); err != nil {
		t.Fatalf("moving the Concierge to hand: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(concierge) {
		t.Error("the Concierge left its owner's hand")
	}
	if !g.Battlefield.Contains(goblin) {
		t.Error("the target was exiled though the Concierge was not")
	}
	if gsExiled(g, goblin) != nil {
		t.Error("the target is in exile")
	}
}

// "Up to one": with no target chosen the Concierge still suspends itself.
func TestSinisterConciergeWithNoTargetSuspendsOnlyItself(t *testing.T) {
	g, me, _, concierge, goblin := scBoard(t)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("declining the target: %v", err)
		}
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	scCheckSuspended(t, g, concierge, me.ID, "the Concierge")
	if !g.Battlefield.Contains(goblin) {
		t.Error("an untargeted creature was exiled")
	}
}

// The target is judged again as the trigger resolves (CR 608.2b): a
// creature that left the battlefield in response is an illegal target, and
// with its only target illegal the trigger does not resolve, so the
// Concierge stays in the graveyard and the bounced creature stays in hand.
func TestSinisterConciergeWithItsTargetGoneDoesNotResolve(t *testing.T) {
	g, me, opp, concierge, goblin := scBoard(t)
	pickCard(t, g, me.ID, goblin)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, goblin); err != nil {
		t.Fatalf("bouncing the target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(concierge) {
		t.Error("the Concierge left the graveyard though its trigger lost its only target")
	}
	if !opp.Hand.Contains(goblin) {
		t.Error("the bounced target left its owner's hand")
	}
}
