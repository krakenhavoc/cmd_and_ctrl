package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choices_at_resolution_cav_a_test.go — the #3172 caveat slice "choices
// made at the right time": each card here used to settle a choice when
// its ability went on the stack, or approximate one with a fixed rule, and
// now asks as it resolves. The per-card tests beside each card prove the
// ordinary path; these cover the corners the move opened.

const (
	cavATechnomancerOracle = "4e58ad76-37c7-4531-b207-6890b39a2679"
	cavAGlissaOracle       = "8859a692-d107-4de2-9110-4681fc2761c7"
)

// With no other creature there is no sacrifice to offer, so no question
// is asked and the trigger does nothing.
func TestCavARuthlessTechnomancerWithNoOtherCreatureAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Ruthless Technomancer", "Creature — Human Wizard", cavATechnomancerOracle)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("the Technomancer alone is no candidate: no prompt")
	}
	if b16CountNamed(g, "Treasure") != 0 {
		t.Error("nothing was sacrificed: no Treasures")
	}
}

// Thopter Spy Network's "if you control an artifact" is checked again as
// the trigger resolves (CR 603.4): lose the last artifact in response and
// there is no Thopter.
func TestCavAThopterSpyNetworkRechecksTheArtifactOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushCatalogPermanent(g, owner.ID, "Thopter Spy Network", "Enchantment", b08ThopterSpyNetworkOracle, false)
	rock := seedPermanentFor(g, owner.ID, "Rock", "Artifact")
	advanceToUpkeepOf(t, g, 1)
	b14Kill(g, rock)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, owner.ID, "Thopter"); n != 0 {
		t.Errorf("%d Thopters, want 0: the artifact was gone as the trigger resolved", n)
	}
}

// Emeria's "seven or more Plains" is checked again as the trigger
// resolves: lose a Plains in response and the creature stays put.
func TestCavAEmeriaRechecksThePlainsOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushCatalogPermanent(g, owner.ID, "Emeria, the Sky Ruin", "Land", b08EmeriaOracle, false)
	var plains []uuid.UUID
	for i := 0; i < 7; i++ {
		plains = append(plains, seedLandOnBattlefield(g, owner.ID, "Plains", "Basic Land — Plains"))
	}
	dead := pushGraveyardCardForTest(owner, "Dead Bear")
	advanceToUpkeepOf(t, g, 1)
	b04WaitForPick(t, g, owner.ID)
	pickCard(t, g, owner.ID, dead)
	b14Kill(g, plains[0])
	answerLatestTriggerPrompt(t, g, owner.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) {
		t.Error("six Plains as the trigger resolved: the creature must not return")
	}
	if !owner.Graveyard.Contains(dead) {
		t.Error("the creature card stays in the graveyard")
	}
}

// "Up to three creature cards" is a real choice even when the library
// holds fewer: with two creatures, burying only one is legal.
func TestCavABuriedAliveLetsYouBuryFewerThanAllOfTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, n := range []string{"Body A", "Body B"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Creature — Zombie",
			Owner: me.ID, Controller: me.ID,
		})
	}
	castCatalogSpell(t, g, "Buried Alive", "Sorcery", b02bBuriedAliveOracle, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("two creature cards are still a choice: the search asks")
	}
	answerSearchNamed(t, g, me.ID, "Body A")
	if !b02bGraveyardHasNamed(me, "Body A") || b02bGraveyardHasNamed(me, "Body B") {
		t.Error("only the chosen creature is buried")
	}
	if spec, _ := Lookup(b02bBuriedAliveOracle); spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the caveat was stale: Buried Alive is complete")
	}
}

// With two commanders in the command zone the player chooses which one
// Command Beacon brings back.
func TestCavACommandBeaconAsksWhichOfTwoCommanders(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beacon := seedPermanentWithOracle(g, me.ID, "Command Beacon", "Land", commandBeaconOracle)
	first, second := uuid.New(), uuid.New()
	for _, c := range []struct {
		id   uuid.UUID
		name string
	}{{first, "Tymna the Weaver"}, {second, "Thrasios, Triton Hero"}} {
		me.Command.PushTop(game.Card{
			InstanceID: c.id, Name: c.name, TypeLine: "Legendary Creature — Human",
			Owner: me.ID, Controller: me.ID, IsCommander: true,
		})
	}
	if err := g.ActivateCatalogAbility(me.ID, beacon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || !hasID(pick.ChooseCards, first) || !hasID(pick.ChooseCards, second) {
		t.Fatalf("both commanders are offered: %+v", pick)
	}
	answerChooseCards(t, g, me.ID, second)
	if offer := latestChoiceOfKind(g, game.PendingChoiceOptionalReplacement); offer != nil {
		if err := g.ResolveOptionalReplacement(offer.ID, me.ID, false); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
	}
	if !me.Hand.Contains(second) || !me.Command.Contains(first) || me.Command.Contains(second) {
		t.Error("the chosen commander is in hand and the other stays in the command zone")
	}
}

// Glissa's third mode asks how many counters of each kind: "up to three"
// in all, so taking two of the first kind leaves one for the second.
func TestCavAGlissaChoosesHowManyCountersOfEachKind(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glissa := b12Push(g, me.ID, "Glissa Sunslayer", "Legendary Creature — Phyrexian Zombie Elf",
		cavAGlissaOracle, 3, 3)
	b64ReadyToAttack(g, glissa)
	prey := pushCounterCreature(g, opp.ID, "Counter Bear", "+1/+1", 2)
	if err := g.AddCounter(prey, "charge", 2); err != nil {
		t.Fatal(err)
	}

	declareAttack(t, g, opp.ID, glissa)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("combat damage to a player asks for the mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil || !hasID(p.PickTargetCards, prey) {
		t.Fatalf("a permanent with counters is a legal target: %+v", p)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: prey}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Kinds are asked in a fixed order: "+1/+1" (two of them), then charge.
	first := payAmountPrompt(g, me.ID)
	if first == nil || first.PayAmount.Max != 2 {
		t.Fatalf("first the +1/+1 counters, up to the two it has: %+v", first)
	}
	if err := g.ResolvePayAmount(first.ID, me.ID, 2); err != nil {
		t.Fatalf("answer 2: %v", err)
	}
	second := payAmountPrompt(g, me.ID)
	if second == nil || second.PayAmount.Max != 1 {
		t.Fatalf("one of the three is left for the charge counters: %+v", second)
	}
	if err := g.ResolvePayAmount(second.ID, me.ID, 1); err != nil {
		t.Fatalf("answer 1: %v", err)
	}
	if got := b12Counter(t, g, prey, "+1/+1"); got != 0 {
		t.Errorf("+1/+1 counters left: %d, want 0", got)
	}
	if got := b12Counter(t, g, prey, "charge"); got != 1 {
		t.Errorf("charge counters left: %d, want 1 (three removed in all)", got)
	}
}

// Resourceful Defense moves exactly the counters chosen, kind by kind.
func TestCavAResourcefulDefenseMovesTheChosenCountersOfEachKind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	def := b12Push(g, me.ID, "Resourceful Defense", "Enchantment", b14ResourcefulDefenseOracle, 0, 0)
	from := pushCounterCreature(g, me.ID, "From", "+1/+1", 2)
	to := b12Creature(g, me.ID, "To", "Creature — Bear", 2, 2)
	if err := g.AddCounter(from, "charge", 1); err != nil {
		t.Fatal(err)
	}
	advanceToMain(t, g)
	b10AddMana(me, "C", "C", "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, def, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			modeRef(game.TargetCard, from, 0, 0),
			modeRef(game.TargetCard, to, 0, 1),
		},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// "+1/+1" sorts before "charge": move one of two, then the charge.
	if err := g.ResolvePayAmount(payAmountPrompt(g, me.ID).ID, me.ID, 1); err != nil {
		t.Fatalf("answer 1: %v", err)
	}
	if err := g.ResolvePayAmount(payAmountPrompt(g, me.ID).ID, me.ID, 1); err != nil {
		t.Fatalf("answer 1: %v", err)
	}
	if b12Counter(t, g, from, "+1/+1") != 1 || b12Counter(t, g, from, "charge") != 0 {
		t.Errorf("from keeps one +1/+1 and loses its charge: %v", b12Card(t, g, from).Counters)
	}
	if b12Counter(t, g, to, "+1/+1") != 1 || b12Counter(t, g, to, "charge") != 1 {
		t.Errorf("to gains exactly what was chosen: %v", b12Card(t, g, to).Counters)
	}
}

// A token that leaves the battlefield ceases to exist, but Resourceful
// Defense still sees the counters it had and who controlled it.
func TestCavAResourcefulDefenseTriggersForATokenThatLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Resourceful Defense", "Enchantment", b14ResourcefulDefenseOracle, 0, 0)
	keeper := b12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 green Squirrel"), 1) })
	token := findBattlefieldByName(g, "Squirrel")
	if token == uuid.Nil {
		t.Fatal("setup: no Squirrel token")
	}
	if err := g.AddCounter(token, "charge", 2); err != nil {
		t.Fatal(err)
	}
	b14Kill(g, token)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, keeper)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, keeper, "charge"); got != 2 {
		t.Errorf("the token's counters move: %d, want 2", got)
	}
}
