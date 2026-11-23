package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m4_cards_test.go — Marvel Super Heroes Commander, M–Z batch 4 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscM4RescueOracle    = "54917578-9493-499e-b200-51d5cf7978ff"
	mscM4RobotOracle     = "89115ed5-95bf-4b81-b352-5481c3addbed"
	mscM4ShotOracle      = "0c0bb9b0-e7e0-465e-826b-2e56032aed1b"
	mscM4SoldierOracle   = "f20ec8e2-d6e4-4707-9b86-107dbc09ac63"
	mscM4CaptainOracle   = "ece77a64-8d99-430e-8407-bdb321828d63"
	mscM4ScoutOracle     = "ecb8e283-5cdb-4e6a-bdff-546c1617363c"
	mscM4NinjaOracle     = "3c02d725-a05e-4086-9caa-ddb005d7ae18"
	mscM4TauntOracle     = "9138a768-d01d-4603-9bae-d3bff6cd2f61"
	mscM4RevivalOracle   = "19cab267-3c51-4322-b1af-22941bbbdadf"
	mscM4CommandosOracle = "49325445-0825-405e-be86-2a814699ce03"
	mscM4ThanosOracle    = "3c067370-1e3d-4262-b7d6-f2bb2656b1d8"
)

func TestMscM4EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscM4RescueOracle:    "Rescue, Pepper Potts",
		mscM4RobotOracle:     "Panther Robot",
		mscM4ShotOracle:      "Stunning Shot",
		mscM4SoldierOracle:   "Shi'ar Soldier",
		mscM4CaptainOracle:   "Selfless Police Captain",
		mscM4ScoutOracle:     "Scout the City",
		mscM4NinjaOracle:     "Undead Hand Ninja",
		mscM4TauntOracle:     "Taunt from the Rampart",
		mscM4RevivalOracle:   "Wondrous Revival",
		mscM4CommandosOracle: "The Howling Commandos",
		mscM4ThanosOracle:    "Thanos, Death's Consort",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func mscM4Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

func mscM4Kill(t *testing.T, g *game.Game, source, victim uuid.UUID, damage int) {
	t.Helper()
	effectstest.B27Damage(g, source, victim, damage)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Rescue, Pepper Potts ---

func TestMscM4RescueReturnsAnArtifactAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gizmo := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Gizmo", TypeLine: "Artifact"})
	id := effectstest.CastHoldCreature(t, g, "Rescue, Pepper Potts", "Legendary Artifact Creature — Human Hero", mscM4RescueOracle, 2, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
		t.Fatal("Rescue offered itself as the other permanent")
	}
	effectstest.AnswerPickTarget(t, g, gizmo)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gizmo) || !me.Hand.Contains(gizmo) {
		t.Error("the artifact was not returned to its owner's hand")
	}
	if effectstest.PlusOnes(g, id) != 1 {
		t.Errorf("Rescue has %d counters, want 1 (it returned an artifact)", effectstest.PlusOnes(g, id))
	}
}

func TestMscM4RescueReturnsANonArtifactCreatureWithoutGrowing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := effectstest.CastHoldCreature(t, g, "Rescue, Pepper Potts", "Legendary Artifact Creature — Human Hero", mscM4RescueOracle, 2, 1)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
	if effectstest.PlusOnes(g, id) != 0 {
		t.Errorf("Rescue has %d counters, want 0 (it returned a plain creature)", effectstest.PlusOnes(g, id))
	}
}

func TestMscM4RescueCountsAnArtifactCreatureAndRefusesOpposingPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	robot := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Robot", TypeLine: "Artifact Creature — Robot", Power: 1, Toughness: 1})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.CastHoldCreature(t, g, "Rescue, Pepper Potts", "Legendary Artifact Creature — Human Hero", mscM4RescueOracle, 2, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was a legal target")
	}
	effectstest.AnswerPickTarget(t, g, robot)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(robot) || effectstest.PlusOnes(g, id) != 1 {
		t.Errorf("artifact creature returned %v, Rescue counters %d; want true and 1", me.Hand.Contains(robot), effectstest.PlusOnes(g, id))
	}
}

// --- Panther Robot ---

func TestMscM4PantherRobotHasReachTrampleAndAffinityForArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	const typeLine = "Artifact Creature — Cat Robot"
	if got := effectstest.SelfPricedMV(t, g, me, mscM4RobotOracle, typeLine, "{10}"); got != 10 {
		t.Fatalf("with no artifacts it costs %d, want 10", got)
	}
	for i := 0; i < 3; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Gizmo", TypeLine: "Artifact"})
	}
	effectstest.PushPermanent(g, g.Seats[1].ID, game.Card{Name: "Their Gizmo", TypeLine: "Artifact"})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if got := effectstest.SelfPricedMV(t, g, me, mscM4RobotOracle, typeLine, "{10}"); got != 7 {
		t.Errorf("with three artifacts of mine it costs %d, want 7", got)
	}
	robot := effectstest.B12Push(g, me.ID, "Panther Robot", typeLine, mscM4RobotOracle, 8, 8)
	effectstest.AssertKeywords(t, g, robot, "reach", "trample")
}

// --- Stunning Shot ---

func TestMscM4StunningShotGrowsMineAndStunsTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.MscM4TryCast(t, g, "Stunning Shot", "Sorcery", mscM4ShotOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, mine) != 2 {
		t.Errorf("my creature has %d counters, want 2", effectstest.PlusOnes(g, mine))
	}
	tc := mscM4Live(t, g, theirs)
	if !tc.Tapped || tc.Counters[game.CounterStun] != 1 {
		t.Errorf("their creature tapped %v, stun %d; want tapped with one stun counter", tc.Tapped, tc.Counters[game.CounterStun])
	}
	if effectstest.PlusOnes(g, theirs) != 0 {
		t.Errorf("their creature got %d +1/+1 counters", effectstest.PlusOnes(g, theirs))
	}
}

func TestMscM4StunningShotEachHalfIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.MscM4TryCast(t, g, "Stunning Shot", "Sorcery", mscM4ShotOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	}); err != nil {
		t.Fatalf("CastSpell with only the second clause: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tc := mscM4Live(t, g, theirs); !tc.Tapped || tc.Counters[game.CounterStun] != 1 {
		t.Errorf("their creature tapped %v, stun %d; want stunned", tc.Tapped, tc.Counters[game.CounterStun])
	}

	g = effectstest.NewCatalogGame(t)
	mine := effectstest.B12Creature(g, g.Seats[0].ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.MscM4TryCast(t, g, "Stunning Shot", "Sorcery", mscM4ShotOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
	}); err != nil {
		t.Fatalf("CastSpell with only the first clause: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, mine) != 2 {
		t.Errorf("my creature has %d counters, want 2", effectstest.PlusOnes(g, mine))
	}
}

func TestMscM4StunningShotRefusesTheWrongSides(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// #4497: stamped, so the zero slot is the caller's word and the pick
	// is refused rather than filed under the clause it could answer.
	active, shot := g.Seats[g.Turn.ActiveSeat], uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: shot, Name: "Stunning Shot", TypeLine: "Sorcery", OracleID: mscM4ShotOracle,
		Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, shot, game.CastSpellParams{
		Targets:        []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 0}},
		TargetsStamped: true,
	}); err == nil {
		t.Error("an opponent's creature was accepted for the counters")
	}
	if _, err := effectstest.MscM4TryCast(t, g, "Stunning Shot", "Sorcery", mscM4ShotOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 1},
	}); err == nil {
		t.Error("my own creature was accepted for the stun")
	}
}

// --- Shi'ar Soldier ---

func mscM4Soldier(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, "Shi'ar Soldier", "Creature — Shi'ar Soldier", mscM4SoldierOracle, false)
}

func TestMscM4ShiarSoldierBouncesAnotherPermanentOfMineOnMyTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soldier := mscM4Soldier(g, me.ID)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, soldier, "flying")
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, soldier, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(soldier)}); err == nil {
		t.Fatal("the Soldier targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, soldier, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's permanent was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, soldier, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the bear was not returned to my hand")
	}
	if c := mscM4Live(t, g, soldier); !c.Tapped {
		t.Error("the {T} cost was not paid")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left floating, the {U} cost should have spent it", me.ManaPool.Total())
	}
}

func TestMscM4ShiarSoldierCannotBeActivatedOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	soldier := mscM4Soldier(g, opp.ID)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, opp, "{U}")
	if err := g.ActivateCatalogAbility(opp.ID, soldier, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Error("the ability was activated during another player's turn")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the bear left the battlefield")
	}
}

// --- Selfless Police Captain ---

func TestMscM4CaptainEntersWithACounterAndPassesItOnWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	captain := effectstest.CastHoldCreature(t, g, "Selfless Police Captain", "Creature — Human Detective", mscM4CaptainOracle, 1, 1)
	if got := effectstest.PlusOnes(g, captain); got != 1 {
		t.Fatalf("the Captain entered with %d counters, want 1", got)
	}
	// Two more counters: all three move.
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: captain, Kind: game.CounterPlusOne, N: 2}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: captain}))
	})
	mscM4Kill(t, g, theirs, captain, 9)
	if g.Battlefield.Contains(captain) {
		t.Fatal("the Captain survived lethal damage")
	}
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was a legal target")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 3 {
		t.Errorf("my bear has %d counters, want the Captain's 3", got)
	}
	if effectstest.PlusOnes(g, theirs) != 0 {
		t.Errorf("their bear got %d counters", effectstest.PlusOnes(g, theirs))
	}
}

func TestMscM4CaptainBouncedStillPassesItsCounterOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	captain := effectstest.CastHoldCreature(t, g, "Selfless Police Captain", "Creature — Human Detective", mscM4CaptainOracle, 1, 1)
	g.WithWriteLock(func() {
		g.BounceCardsToHandForEffect([]uuid.UUID{captain})
	})
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("my bear has %d counters, want 1", got)
	}
}

// --- Scout the City ---

func mscM4StackScoutLibrary(me *game.Player) (land, bear, spell uuid.UUID) {
	effectstest.PlTop(me, "Deep Card", "Instant", "{U}")
	spell = effectstest.PlTop(me, "Milled Spell", "Sorcery", "{R}")
	land = effectstest.PlTop(me, "Milled Land", "Basic Land — Forest", "")
	bear = effectstest.PlTop(me, "Milled Bear", "Creature — Bear", "{1}{G}")
	return land, bear, spell
}

func TestMscM4ScoutTheCityMillsThreeTakesAPermanentAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land, bear, spell := mscM4StackScoutLibrary(me)
	lib, life := me.Library.Size(), me.Life
	if _, err := effectstest.MscM4TryCast(t, g, "Scout the City", "Sorcery", mscM4ScoutOracle, []int{0}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-3 {
		t.Errorf("library %d -> %d, want three milled", lib, me.Library.Size())
	}
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no permanent card was offered from among the milled cards")
	}
	if len(pick.ChooseCards) != 2 || !effectstest.HasID(pick.ChooseCards, land) || !effectstest.HasID(pick.ChooseCards, bear) || effectstest.HasID(pick.ChooseCards, spell) {
		t.Errorf("offered %v, want the milled land and creature only", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !effectstest.MiInHand(me, land) || effectstest.MiInHand(me, bear) {
		t.Error("only the chosen land should have reached the hand")
	}
}

func TestMscM4ScoutTheCityTakingIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, bear, _ := mscM4StackScoutLibrary(me)
	if _, err := effectstest.MscM4TryCast(t, g, "Scout the City", "Sorcery", mscM4ScoutOracle, []int{0}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if effectstest.MiInHand(me, bear) || !me.Graveyard.Contains(bear) {
		t.Error("the creature left the graveyard although the player declined")
	}
}

func TestMscM4ScoutTheCityDestroysOnlyACreatureWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	flier := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 1)
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.BattlefieldCardPtr(t, g, flier).Keywords = append(effectstest.BattlefieldCardPtr(t, g, flier).Keywords, "flying")
	if _, err := effectstest.MscM4TryCast(t, g, "Scout the City", "Sorcery", mscM4ScoutOracle, []int{1}, effectstest.B16TargetCard(ground)); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	if _, err := effectstest.MscM4TryCast(t, g, "Scout the City", "Sorcery", mscM4ScoutOracle, []int{1}, effectstest.B16TargetCard(flier)); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the flier survived")
	}
	if !g.Battlefield.Contains(ground) {
		t.Error("the ground creature was destroyed")
	}
}

// --- Undead Hand Ninja ---

func TestMscM4UndeadHandNinjaDrainsOncePerBatchOfCreatureCardsLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ninja := effectstest.B12Push(g, me.ID, "Undead Hand Ninja", "Creature — Zombie Ninja Villain", mscM4NinjaOracle, 2, 1)
	effectstest.AssertKeywords(t, g, ninja, "deathtouch")
	a := effectstest.PushGraveyardTyped(me, "Hero A", "Legendary Creature — Human Hero")
	b := effectstest.PushGraveyardTyped(me, "Hero B", "Legendary Creature — Human Hero")
	myLife, oppLife := me.Life, opp.Life
	if _, err := effectstest.MscM4TryCast(t, g, "Wondrous Revival", "Sorcery", mscM4RevivalOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("the Heroes did not return")
	}
	if me.Life != myLife+1 || opp.Life != oppLife-1 {
		t.Errorf("two creature cards leaving in one batch: my life %d -> %d, theirs %d -> %d; want one drain of 1", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestMscM4UndeadHandNinjaIgnoresLandsAndOpposingGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Undead Hand Ninja", "Creature — Zombie Ninja Villain", mscM4NinjaOracle, 2, 1)
	land := effectstest.PushGraveyardTyped(me, "Old Forest", "Basic Land — Forest")
	theirs := effectstest.PushGraveyardTyped(opp, "Their Hero", "Creature — Human Hero")
	myLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(land)
		_ = g.ExileCardForEffect(theirs)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(land) || opp.Graveyard.Contains(theirs) {
		t.Fatal("the cards did not leave their graveyards")
	}
	if me.Life != myLife || opp.Life != oppLife {
		t.Errorf("a land card or an opposing graveyard drained: my life %d -> %d, theirs %d -> %d", myLife, me.Life, oppLife, opp.Life)
	}
	creature := effectstest.PushGraveyardTyped(me, "My Hero", "Creature — Human Hero")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(creature) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != oppLife-1 {
		t.Errorf("my creature card leaving: my life %d -> %d, theirs %d -> %d; want a drain of 1", myLife, me.Life, oppLife, opp.Life)
	}
}

// --- Taunt from the Rampart ---

func TestMscM4TauntGoadsAndStopsOnlyTheOpposingCreaturesThereNow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Taunt from the Rampart", "Sorcery", mscM4TauntOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c := mscM4Live(t, g, id); c.LatestGoader() != me.ID {
			t.Errorf("%s was not goaded by me", c.Name)
		}
		effectstest.AssertRestrictions(t, g, id, game.CantBlock)
	}
	if c := mscM4Live(t, g, mine); c.LatestGoader() == me.ID {
		t.Error("my own creature was goaded")
	}
	effectstest.AssertRestrictions(t, g, mine, 0)
	late := effectstest.B12Creature(g, opp.ID, "Late Bear", "Creature — Bear", 2, 2)
	effectstest.AssertRestrictions(t, g, late, 0)
}

func TestMscM4TauntLastsUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Taunt from the Rampart", "Sorcery", mscM4TauntOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, a, game.CantBlock)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	effectstest.AssertRestrictions(t, g, a, game.CantBlock)
	// The goaded creature must attack this combat; let it.
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AssertRestrictions(t, g, a, 0)
}

// --- Wondrous Revival ---

func TestMscM4WondrousRevivalReturnsUpToThreeHeroCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h1 := effectstest.PushGraveyardTyped(me, "Hero One", "Legendary Creature — Human Hero")
	h2 := effectstest.PushGraveyardTyped(me, "Hero Two", "Creature — Mutant Hero")
	h3 := effectstest.PushGraveyardTyped(me, "Hero Three", "Creature — Human Hero")
	h4 := effectstest.PushGraveyardTyped(me, "Hero Four", "Creature — Human Hero")
	villain := effectstest.PushGraveyardTyped(me, "Villain", "Creature — Human Villain")
	spell := effectstest.PushGraveyardTyped(me, "Hero Spell", "Instant — Hero")
	theirs := effectstest.PushGraveyardTyped(opp, "Their Hero", "Creature — Human Hero")

	for _, bad := range []uuid.UUID{villain, spell, theirs} {
		if _, err := effectstest.MscM4TryCast(t, g, "Wondrous Revival", "Sorcery", mscM4RevivalOracle, nil, effectstest.B16TargetCard(bad)); err == nil {
			t.Error("a card that is not one of my Hero creature cards was a legal target")
		}
	}
	if _, err := effectstest.MscM4TryCast(t, g, "Wondrous Revival", "Sorcery", mscM4RevivalOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: h1}, {Kind: game.TargetCard, ID: h2}, {Kind: game.TargetCard, ID: h3}, {Kind: game.TargetCard, ID: h4},
	}); err == nil {
		t.Error("four targets were accepted for up to three")
	}
	if _, err := effectstest.MscM4TryCast(t, g, "Wondrous Revival", "Sorcery", mscM4RevivalOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: h1}, {Kind: game.TargetCard, ID: h2}, {Kind: game.TargetCard, ID: h3},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{h1, h2, h3} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("target %s did not return to the battlefield", id)
		}
	}
	if g.Battlefield.Contains(h4) || !me.Graveyard.Contains(h4) {
		t.Error("the untargeted Hero returned")
	}
}

// --- The Howling Commandos ---

func TestMscM4HowlingCommandosPumpsEveryCreatureAndGivesSoldiersVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	commandos := effectstest.B12Push(g, me.ID, "The Howling Commandos", "Legendary Creature — Human Soldier Hero", mscM4CommandosOracle, 3, 1)
	soldier := effectstest.B12Creature(g, me.ID, "A Soldier", "Creature — Human Soldier", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Soldier", "Creature — Human Soldier", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, commandos, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left floating, the {5} cost should have spent it", me.ManaPool.Total())
	}
	for id, want := range map[uuid.UUID][2]int{commandos: {4, 2}, soldier: {3, 3}, bear: {3, 3}, theirs: {2, 2}} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != want[0] || tough != want[1] {
			t.Errorf("%s is %d/%d, want %d/%d", mscM4Live(t, g, id).Name, p, tough, want[0], want[1])
		}
	}
	effectstest.AssertKeywords(t, g, commandos, "vigilance")
	effectstest.AssertKeywords(t, g, soldier, "vigilance")
	if c := mscM4Live(t, g, bear); game.HasKeyword(&c, "vigilance") {
		t.Error("a non-Soldier gained vigilance")
	}
	if c := mscM4Live(t, g, theirs); game.HasKeyword(&c, "vigilance") {
		t.Error("an opposing Soldier gained vigilance")
	}
}

func TestMscM4HowlingCommandosNeedsFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	commandos := effectstest.B12Push(g, me.ID, "The Howling Commandos", "Legendary Creature — Human Soldier Hero", mscM4CommandosOracle, 3, 1)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, commandos, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated with four mana")
	}
}

// --- Thanos, Death's Consort ---

func TestMscM4ThanosGrowsWhenAnotherCreatureDiesButNotWhenItDoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thanos := effectstest.B12Push(g, me.ID, "Thanos, Death's Consort", "Legendary Creature — Eternal Villain", mscM4ThanosOracle, 3, 3)
	effectstest.AssertKeywords(t, g, thanos, "lifelink")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	mscM4Kill(t, g, theirs, theirs, 1)
	if g.Battlefield.Contains(theirs) || effectstest.PlusOnes(g, thanos) != 1 {
		t.Fatalf("after an opposing creature died Thanos has %d counters, want 1", effectstest.PlusOnes(g, thanos))
	}
	mscM4Kill(t, g, theirs, mine, 1)
	if effectstest.PlusOnes(g, thanos) != 2 {
		t.Errorf("after my own creature died Thanos has %d counters, want 2", effectstest.PlusOnes(g, thanos))
	}
	// A token dying is a creature dying too; Thanos dying grows nothing.
	mscM4Kill(t, g, theirs, thanos, 99)
	if g.Battlefield.Contains(thanos) {
		t.Error("Thanos survived lethal damage")
	}
}
