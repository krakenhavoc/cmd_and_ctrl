package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_seams_4707_test.go — the three small seams Marvel Super Heroes Commander
// (#4455) needed, each driven through its card (#4707): Black Bolt's "that
// player controls" (the event's Actor), Council of Reeds' creature-only legend
// rule exemption, and Flying Drone's coloured activation discount.

const (
	msc4707BlackBoltOracle    = "a30f30e3-dbbf-4cb2-b3d0-76f57dffb2dd"
	msc4707CouncilOfReeds     = "33fa949c-f29e-4a98-9381-ecfa9e5c3f6d"
	msc4707FlyingDroneOracle  = "541f7525-a9a4-493d-b724-354ae6d37de5"
	msc4707ParapetThrasherOID = "722e23d5-4f2a-43cd-bbd5-baf56b76e68d"
)

func TestMsc4707EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		msc4707BlackBoltOracle:   "Black Bolt, Inhuman King",
		msc4707CouncilOfReeds:    "Council of Reeds",
		msc4707FlyingDroneOracle: "Flying Drone",
	} {
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

// --- Black Bolt, Inhuman King ------------------------------------------

type blackBoltTable struct {
	g                        *game.Game
	me, them, third          *game.Player
	bolt                     uuid.UUID
	myRelic                  uuid.UUID
	theirRelicA, theirRelicB uuid.UUID
	theirLand, thirdRelic    uuid.UUID
}

func newBlackBoltTable(t *testing.T) blackBoltTable {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	tb := blackBoltTable{g: g, me: g.Seats[0], them: g.Seats[1], third: g.Seats[2]}
	tb.bolt = effectstest.B12Push(g, tb.me.ID, "Black Bolt, Inhuman King", "Legendary Creature — Inhuman Noble Hero", msc4707BlackBoltOracle, 3, 3)
	tb.myRelic = effectstest.PushPermanentForTest(g, tb.me.ID, "My Relic", "", "Artifact")
	tb.theirRelicA = effectstest.PushPermanentForTest(g, tb.them.ID, "Their Relic A", "", "Artifact")
	tb.theirRelicB = effectstest.PushPermanentForTest(g, tb.them.ID, "Their Relic B", "", "Artifact")
	tb.theirLand = effectstest.PushPermanentForTest(g, tb.them.ID, "Their Forest", "", "Basic Land — Forest")
	tb.thirdRelic = effectstest.PushPermanentForTest(g, tb.third.ID, "Third Relic", "", "Artifact")
	effectstest.AdvanceToMain(t, g)
	return tb
}

// targetBoltWith casts an uncatalogued instant for `who` at Black Bolt.
func (tb blackBoltTable) targetBoltWith(t *testing.T, who *game.Player) {
	t.Helper()
	spell := game.NewCard("Test Removal", who.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{R}"
	who.Hand.PushTop(spell)
	if err := tb.g.CastSpell(who.ID, spell.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tb.bolt}},
	}); err != nil {
		t.Fatalf("cast the targeting spell: %v", err)
	}
}

func TestMsc4707BlackBoltDestroysAPermanentOfTheOpponentWhoTargetedIt(t *testing.T) {
	tb := newBlackBoltTable(t)
	tb.targetBoltWith(t, tb.them)

	ch := effectstest.PickTargetPrompt(t, tb.g)
	if ch.Chooser != tb.me.ID {
		t.Errorf("the trigger's target is chosen by %s, want Black Bolt's controller", ch.Chooser)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range ch.PickTargetCards {
		offered[id] = true
	}
	for id, want := range map[uuid.UUID]bool{
		tb.theirRelicA: true, tb.theirRelicB: true,
		tb.theirLand: false, tb.myRelic: false, tb.thirdRelic: false, tb.bolt: false,
	} {
		if offered[id] != want {
			t.Errorf("%s offered = %v, want %v", id, offered[id], want)
		}
	}
	effectstest.AnswerPickTarget(t, tb.g, tb.theirRelicA)
	effectstest.PassPriorityAroundTable(t, tb.g)

	if effectstest.OnBoard(tb.g, tb.theirRelicA) {
		t.Error("the targeting opponent's permanent was not destroyed")
	}
	for name, id := range map[string]uuid.UUID{
		"Black Bolt": tb.bolt, "my artifact": tb.myRelic, "their other artifact": tb.theirRelicB,
		"their land": tb.theirLand, "a third player's artifact": tb.thirdRelic,
	} {
		if !effectstest.OnBoard(tb.g, id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

func TestMsc4707BlackBoltNamesTheTargetingPlayerNotJustAnyOpponent(t *testing.T) {
	tb := newBlackBoltTable(t)
	tb.targetBoltWith(t, tb.third)
	// The third player controls only one nonland permanent: still a prompt or an
	// automatic pick, and never one of seat 1's.
	effectstest.PassPriorityAroundTable(t, tb.g)
	for _, pc := range tb.g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoicePickTarget {
			effectstest.AnswerPickTarget(t, tb.g, tb.thirdRelic)
			effectstest.PassPriorityAroundTable(t, tb.g)
			break
		}
	}
	if effectstest.OnBoard(tb.g, tb.thirdRelic) {
		t.Error("the third player's artifact survived their own targeting")
	}
	for _, id := range []uuid.UUID{tb.theirRelicA, tb.theirRelicB, tb.myRelic} {
		if !effectstest.OnBoard(tb.g, id) {
			t.Error("a permanent of a player who did not target Black Bolt was destroyed")
		}
	}
}

func TestMsc4707BlackBoltIgnoresItsOwnControllersTargeting(t *testing.T) {
	tb := newBlackBoltTable(t)
	tb.targetBoltWith(t, tb.me)
	effectstest.PassPriorityAroundTable(t, tb.g)
	for _, pc := range tb.g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoicePickTarget {
			t.Fatal("Black Bolt triggered off its own controller's spell")
		}
	}
	for _, id := range []uuid.UUID{tb.myRelic, tb.theirRelicA, tb.theirRelicB, tb.thirdRelic} {
		if !effectstest.OnBoard(tb.g, id) {
			t.Error("a permanent was destroyed by a spell Black Bolt's controller cast")
		}
	}
}

func TestMsc4707BlackBoltGetsPlusTwoPlusTwoForNoncreatureSpellsOnly(t *testing.T) {
	tb := newBlackBoltTable(t)
	effectstest.Msc12CastAndSettle(t, tb.g, "Test Bear", "Creature — Bear")
	if p := effectstest.EffectivePower(t, tb.g, tb.bolt); p != 3 {
		t.Fatalf("a creature spell pumped Black Bolt to power %d", p)
	}
	effectstest.Msc12CastAndSettle(t, tb.g, "Test Relic", "Artifact")
	if c := effectstest.GaLayered(t, tb.g, tb.bolt); c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Errorf("Black Bolt is %d/%d after a noncreature spell, want 5/5", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Council of Reeds ---------------------------------------------------

func TestMsc4707CouncilKeepsTwoLegendaryCreaturesButNotOtherLegends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Council of Reeds", "Legendary Creature — Human Scientist Hero", msc4707CouncilOfReeds, 2, 2)
	creatureA := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	creatureB := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	relicA := effectstest.B12Permanent(g, me.ID, "Legend Relic", "Legendary Artifact")
	relicB := effectstest.B12Permanent(g, me.ID, "Legend Relic", "Legendary Artifact")
	theirA := effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	theirB := effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)

	g.RunStateChecksForTest()

	if !effectstest.OnBoard(g, creatureA) || !effectstest.OnBoard(g, creatureB) {
		t.Error(effectstest.BothLegendsMustStay + " among creatures you control")
	}
	ch := effectstest.LegendPrompt(g, me.ID)
	if ch == nil {
		t.Fatal("two legendary artifacts you control were exempted by a creature-only clause")
	}
	if len(ch.PickTargetCards) != 2 || !effectstest.OnBoard(g, relicA) || !effectstest.OnBoard(g, relicB) {
		t.Errorf("the prompt should be about the two artifacts only: %v", ch.PickTargetCards)
	}
	for _, id := range ch.PickTargetCards {
		if id == creatureA || id == creatureB {
			t.Error("a creature was offered to the legend rule")
		}
	}
	oppCh := effectstest.LegendPrompt(g, opp.ID)
	if oppCh == nil {
		t.Fatal("an opponent's duplicate legendary creatures were exempted")
	}
	if len(oppCh.PickTargetCards) != 2 || !effectstest.OnBoard(g, theirA) || !effectstest.OnBoard(g, theirB) {
		t.Errorf("opponent prompt malformed: %v", oppCh.PickTargetCards)
	}
}

func TestMsc4707CouncilLeavingRestoresTheLegendRuleForCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	council := effectstest.B12Push(g, me.ID, "Council of Reeds", "Legendary Creature — Human Scientist Hero", msc4707CouncilOfReeds, 2, 2)
	effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) != nil {
		t.Fatal("prompted while Council of Reeds was in play")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(council); err != nil {
			t.Fatalf("destroy Council: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) == nil {
		t.Error("the legend rule did not apply again after Council of Reeds left")
	}
}

func councilCopies(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Council of Reeds" {
			n++
		}
	}
	return n
}

func TestMsc4707CouncilCopiesItselfAtCombatAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Council of Reeds", "Legendary Creature — Human Scientist Hero", msc4707CouncilOfReeds, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc12CastAndSettle(t, g, "Test Bear", "Creature — Bear")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := councilCopies(g); n != 1 {
		t.Fatalf("%d Councils after only a creature spell, want 1", n)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Council of Reeds", "Legendary Creature — Human Scientist Hero", msc4707CouncilOfReeds, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc12CastAndSettle(t, g, "Test Relic", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := councilCopies(g); n != 2 {
		t.Fatalf("%d Councils after a noncreature spell, want the original and one token copy", n)
	}
	// The legend rule doesn't apply to the copy, which is a legendary creature.
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) != nil {
		t.Error("the token copy of Council of Reeds triggered the legend rule")
	}
}

// --- Flying Drone ---------------------------------------------------------

func newDroneTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drone := effectstest.PushCatalogPermanent(g, me.ID, "Flying Drone", "Artifact Creature — Robot Scout", msc4707FlyingDroneOracle, false)
	effectstest.AdvanceToMain(t, g)
	return g, me, drone
}

// activateDrone tries the loot ability under strict mana with `mana` floating.
func activateDrone(t *testing.T, g *game.Game, drone uuid.UUID, mana string) error {
	t.Helper()
	return effectstest.HobActivateStrict(t, g, g.Seats[0].ID, drone, mana)
}

func TestMsc4707FlyingDroneCostsFullPriceWithoutAnotherFlier(t *testing.T) {
	g, _, drone := newDroneTable(t)
	if err := activateDrone(t, g, drone, ""); err == nil {
		t.Fatal("activated for {T} alone with no flier entered")
	}
	if effectstest.BattlefieldTapped(g, drone) {
		t.Fatal("a refused activation tapped the Drone")
	}
	if err := activateDrone(t, g, drone, "{C}{U}"); err != nil {
		t.Fatalf("activation for {1}{U}: %v", err)
	}
	if !effectstest.BattlefieldTapped(g, drone) {
		t.Error("the Drone is not tapped")
	}
}

func TestMsc4707FlyingDroneCostsOnlyTapAfterAnotherFlierEntersUnderYourControl(t *testing.T) {
	g, me, drone := newDroneTable(t)
	effectstest.CastCatalogSpell(t, g, "Parapet Thrasher", "Creature — Dragon", msc4707ParapetThrasherOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := activateDrone(t, g, drone, ""); err != nil {
		t.Fatalf("activation with no mana after your flier entered: %v", err)
	}
	if !effectstest.BattlefieldTapped(g, drone) {
		t.Error("the Drone is not tapped")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("mana was spent or left over: %d", me.ManaPool.Total())
	}
}

func TestMsc4707FlyingDroneIgnoresItsOwnEntryAndAnOpponentsFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Drone itself entering: a flier, but not "another".
	droneID := effectstest.CastCatalogSpell(t, g, "Flying Drone", "Artifact Creature — Robot Scout", msc4707FlyingDroneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(droneID) {
		t.Fatal("the Drone did not resolve")
	}
	// Past summoning sickness, so only the price can refuse its {T} ability.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == droneID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	// An opponent's flier entering under their control.
	effectstest.PushCatalogPermanent(g, opp.ID, "Parapet Thrasher", "Creature — Dragon", msc4707ParapetThrasherOID, false)
	g.WithWriteLock(func() {
		thrasher := g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: thrasher.InstanceID, Actor: opp.ID})
	})
	var theirs bool
	g.ReadSnapshot(func() { theirs = g.EnteredFlierUnderThisTurn(opp.ID, uuid.Nil) })
	if !theirs {
		t.Fatal("test setup: the opponent's flier entry was not recorded")
	}
	if err := effectstest.HobActivateStrict(t, g, me.ID, droneID, ""); err == nil {
		t.Fatal("the Drone's own entry or an opponent's flier made the ability cost {T} only")
	}
	if err := effectstest.HobActivateStrict(t, g, me.ID, droneID, "{C}{U}"); err != nil {
		t.Fatalf("full-price activation: %v", err)
	}
}
