package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// With no other creature there is no sacrifice to offer, so no question
// is asked and the trigger does nothing.
func TestCavARuthlessTechnomancerWithNoOtherCreatureAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Ruthless Technomancer", "Creature — Human Wizard", effectstest.CavATechnomancerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("the Technomancer alone is no candidate: no prompt")
	}
	if effectstest.B16CountNamed(g, "Treasure") != 0 {
		t.Error("nothing was sacrificed: no Treasures")
	}
}

// Thopter Spy Network's "if you control an artifact" is checked again as
// the trigger resolves (CR 603.4): lose the last artifact in response and
// there is no Thopter.
func TestCavAThopterSpyNetworkRechecksTheArtifactOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Thopter Spy Network", "Enchantment", effectstest.B08ThopterSpyNetworkOracle, false)
	rock := effectstest.SeedPermanentFor(g, owner.ID, "Rock", "Artifact")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.B14Kill(g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, owner.ID, "Thopter"); n != 0 {
		t.Errorf("%d Thopters, want 0: the artifact was gone as the trigger resolved", n)
	}
}

// Emeria's "seven or more Plains" is checked again as the trigger
// resolves: lose a Plains in response and the creature stays put.
func TestCavAEmeriaRechecksThePlainsOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Emeria, the Sky Ruin", "Land", effectstest.B08EmeriaOracle, false)
	var plains []uuid.UUID
	for i := 0; i < 7; i++ {
		plains = append(plains, effectstest.SeedLandOnBattlefield(g, owner.ID, "Plains", "Basic Land — Plains"))
	}
	dead := effectstest.PushGraveyardCardForTest(owner, "Dead Bear")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.B04WaitForPick(t, g, owner.ID)
	effectstest.PickCard(t, g, owner.ID, dead)
	effectstest.B14Kill(g, plains[0])
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, owner.ID) != nil {
		t.Error("a trigger whose condition failed asked its \"you may\" (#3907)")
	}
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, n := range []string{"Body A", "Body B"} {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: n, TypeLine: "Creature — Zombie",
			Owner: me.ID, Controller: me.ID,
		})
	}
	effectstest.CastCatalogSpell(t, g, "Buried Alive", "Sorcery", effectstest.B02bBuriedAliveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("two creature cards are still a choice: the search asks")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Body A")
	if !effectstest.B02bGraveyardHasNamed(me, "Body A") || effectstest.B02bGraveyardHasNamed(me, "Body B") {
		t.Error("only the chosen creature is buried")
	}
	if spec, _ := effects.Lookup(effectstest.B02bBuriedAliveOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the caveat was stale: Buried Alive is complete")
	}
}

// With two commanders in the command zone the player chooses which one
// Command Beacon brings back.
func TestCavACommandBeaconAsksWhichOfTwoCommanders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beacon := effectstest.SeedPermanentWithOracle(g, me.ID, "Command Beacon", "Land", effectstest.CommandBeaconOracle)
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
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, first) || !effectstest.HasID(pick.ChooseCards, second) {
		t.Fatalf("both commanders are offered: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	if offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionalReplacement); offer != nil {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glissa := effectstest.B12Push(g, me.ID, "Glissa Sunslayer", "Legendary Creature — Phyrexian Zombie Elf",
		effectstest.CavAGlissaOracle, 3, 3)
	effectstest.B64ReadyToAttack(g, glissa)
	prey := effectstest.PushCounterCreature(g, opp.ID, "Counter Bear", "+1/+1", 2)
	if err := g.AddCounter(prey, "charge", 2); err != nil {
		t.Fatal(err)
	}

	effectstest.DeclareAttack(t, g, opp.ID, glissa)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("combat damage to a player asks for the mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, prey) {
		t.Fatalf("a permanent with counters is a legal target: %+v", p)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: prey}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Kinds are asked in a fixed order: "+1/+1" (two of them), then charge.
	first := effectstest.PayAmountPrompt(g, me.ID)
	if first == nil || first.PayAmount.Max != 2 {
		t.Fatalf("first the +1/+1 counters, up to the two it has: %+v", first)
	}
	if err := g.ResolvePayAmount(first.ID, me.ID, 2); err != nil {
		t.Fatalf("answer 2: %v", err)
	}
	second := effectstest.PayAmountPrompt(g, me.ID)
	if second == nil || second.PayAmount.Max != 1 {
		t.Fatalf("one of the three is left for the charge counters: %+v", second)
	}
	if err := g.ResolvePayAmount(second.ID, me.ID, 1); err != nil {
		t.Fatalf("answer 1: %v", err)
	}
	if got := effectstest.B12Counter(t, g, prey, "+1/+1"); got != 0 {
		t.Errorf("+1/+1 counters left: %d, want 0", got)
	}
	if got := effectstest.B12Counter(t, g, prey, "charge"); got != 1 {
		t.Errorf("charge counters left: %d, want 1 (three removed in all)", got)
	}
}

// Resourceful Defense moves exactly the counters chosen, kind by kind.
func TestCavAResourcefulDefenseMovesTheChosenCountersOfEachKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	def := effectstest.B12Push(g, me.ID, "Resourceful Defense", "Enchantment", effectstest.B14ResourcefulDefenseOracle, 0, 0)
	from := effectstest.PushCounterCreature(g, me.ID, "From", "+1/+1", 2)
	to := effectstest.B12Creature(g, me.ID, "To", "Creature — Bear", 2, 2)
	if err := g.AddCounter(from, "charge", 1); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "C", "C", "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, def, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			effectstest.ModeRef(game.TargetCard, from, 0, 0),
			effectstest.ModeRef(game.TargetCard, to, 0, 1),
		},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// "+1/+1" sorts before "charge": move one of two, then the charge.
	if err := g.ResolvePayAmount(effectstest.PayAmountPrompt(g, me.ID).ID, me.ID, 1); err != nil {
		t.Fatalf("answer 1: %v", err)
	}
	if err := g.ResolvePayAmount(effectstest.PayAmountPrompt(g, me.ID).ID, me.ID, 1); err != nil {
		t.Fatalf("answer 1: %v", err)
	}
	if effectstest.B12Counter(t, g, from, "+1/+1") != 1 || effectstest.B12Counter(t, g, from, "charge") != 0 {
		t.Errorf("from keeps one +1/+1 and loses its charge: %v", effectstest.B12Card(t, g, from).Counters)
	}
	if effectstest.B12Counter(t, g, to, "+1/+1") != 1 || effectstest.B12Counter(t, g, to, "charge") != 1 {
		t.Errorf("to gains exactly what was chosen: %v", effectstest.B12Card(t, g, to).Counters)
	}
}

// A token that leaves the battlefield ceases to exist, but Resourceful
// Defense still sees the counters it had and who controlled it.
func TestCavAResourcefulDefenseTriggersForATokenThatLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Resourceful Defense", "Enchantment", effectstest.B14ResourcefulDefenseOracle, 0, 0)
	keeper := effectstest.B12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 green Squirrel"), 1) })
	token := effectstest.FindBattlefieldByName(g, "Squirrel")
	if token == uuid.Nil {
		t.Fatal("setup: no Squirrel token")
	}
	if err := g.AddCounter(token, "charge", 2); err != nil {
		t.Fatal(err)
	}
	effectstest.B14Kill(g, token)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, keeper)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, keeper, "charge"); got != 2 {
		t.Errorf("the token's counters move: %d, want 2", got)
	}
}
