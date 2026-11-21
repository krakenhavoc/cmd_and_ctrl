package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestE5LandsAndStaffAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.LushPorticoOracle:         "Lush Portico",
		effectstest.UndergroundMortuaryOracle: "Underground Mortuary",
		effectstest.IndathaTriomeOracle:       "Indatha Triome",
		effectstest.StaffOfCompleationOracle:  "Staff of Compleation",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- Lush Portico / Underground Mortuary ---------------------------

func TestLushPorticoEntersTappedAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Next Draw")

	id := effectstest.PlayLandFromHand(t, g, "Lush Portico", effectstest.LushPorticoOracle)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.Top100AssertEnteredTapped(t, g, id, "Lush Portico")
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("playing Lush Portico did not queue a surveil")
	}
}

func TestUndergroundMortuaryEntersTappedAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Next Draw")

	id := effectstest.PlayLandFromHand(t, g, "Underground Mortuary", effectstest.UndergroundMortuaryOracle)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.Top100AssertEnteredTapped(t, g, id, "Underground Mortuary")
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("playing Underground Mortuary did not queue a surveil")
	}
}

func TestLushPorticoTapsForGreenOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Lush Portico", "Land", effectstest.LushPorticoOracle)

	for _, want := range []string{"G", "W"} {
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := effectstest.RiderLatestManaPick(g, me.ID)
		if pick == nil || len(pick.ColorOptions) != 2 {
			t.Fatalf("colour options %+v, want G and W", pick)
		}
		effectstest.RiderAnswerManaPicks(t, g, me.ID, want)
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != want {
			t.Errorf("pool %v, want [%s]", got, want)
		}
		me.ManaPool.EmptyPool()
		effectstest.B08Untap(g, land)
	}
}

func TestUndergroundMortuaryTapsForBlackOrGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Underground Mortuary", "Land", effectstest.UndergroundMortuaryOracle)

	for _, want := range []string{"B", "G"} {
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := effectstest.RiderLatestManaPick(g, me.ID)
		if pick == nil || len(pick.ColorOptions) != 2 {
			t.Fatalf("colour options %+v, want B and G", pick)
		}
		effectstest.RiderAnswerManaPicks(t, g, me.ID, want)
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != want {
			t.Errorf("pool %v, want [%s]", got, want)
		}
		me.ManaPool.EmptyPool()
		effectstest.B08Untap(g, land)
	}
}

// --- Indatha Triome --------------------------------------------------

func TestIndathaTriomeEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Indatha Triome", effectstest.IndathaTriomeOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Indatha Triome")
}

func TestIndathaTriomeTapsForWhiteBlackOrGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Indatha Triome", "Land", effectstest.IndathaTriomeOracle)

	for _, want := range []string{"W", "B", "G"} {
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := effectstest.RiderLatestManaPick(g, me.ID)
		if pick == nil || len(pick.ColorOptions) != 3 {
			t.Fatalf("colour options %+v, want W, B and G", pick)
		}
		effectstest.RiderAnswerManaPicks(t, g, me.ID, want)
		if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != want {
			t.Errorf("pool %v, want [%s]", got, want)
		}
		me.ManaPool.EmptyPool()
		effectstest.B08Untap(g, land)
	}
}

// TestIndathaTriomeCyclesForACard mirrors
// TestKetriaTriomeCyclesForACard (cycling_test.go) one colour wheel
// over.
func TestIndathaTriomeCyclesForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.CycleFromHand(t, g, "Indatha Triome", "Land — Plains Swamp Forest", effectstest.IndathaTriomeOracle, "{C}{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Triome is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
}

// --- Staff of Compleation --------------------------------------------

// staffOfCompleation seeds Staff of Compleation onto the battlefield,
// free of summoning sickness (irrelevant to an artifact, but the
// helper still asks).
func staffOfCompleation(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, "Staff of Compleation", "Artifact", effectstest.StaffOfCompleationOracle, false)
}

// The mana ability: pays 2 life, offers all five colours, does not
// use the stack.
func TestStaffOfCompleationManaAbilityPaysTwoLifeAndSkipsTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := staffOfCompleation(g, me.ID)
	before := me.Life

	if err := g.ActivateManaAbility(me.ID, staff, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-2 {
		t.Errorf("life %d -> %d, want %d", before, me.Life, before-2)
	}
	if g.Stack.Size() != 0 {
		t.Error("the mana ability used the stack — mana abilities never do (CR 605.3b)")
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("colour options %+v, want all five", pick)
	}
	effectstest.RiderAnswerManaPicks(t, g, me.ID, "U")
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
}

// The mana ability's life is a COST — same discriminator Mana
// Confluence's tests use: at 1 life, paying 2 is rejected and the
// artifact does NOT tap.
func TestStaffOfCompleationManaAbilityRefusesWithoutEnoughLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := staffOfCompleation(g, me.ID)
	g.WithWriteLock(func() { me.Life = 1 })

	if err := g.ActivateManaAbility(me.ID, staff, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated the mana ability while unable to pay 2 life")
	}
	card, ok := effectstest.BattlefieldCard(g, staff)
	if !ok {
		t.Fatal("the staff left the battlefield")
	}
	if card.Tapped {
		t.Error("the rejected activation tapped the staff anyway — a cost must be validated before any of it is paid")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v after a rejected activation, want empty", me.ManaPool)
	}
}

// {T}, Pay 1 life: Destroy target permanent you own — the ordinary
// case, an owned-and-controlled permanent.
func TestStaffOfCompleationDestroysAnOwnedPermanentAndPaysOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := staffOfCompleation(g, me.ID)
	victim := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	before := me.Life

	effectstest.B16Activate(t, g, me.ID, staff, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})

	if me.Life != before-1 {
		t.Errorf("life %d -> %d, want %d", before, me.Life, before-1)
	}
	if z := effectstest.B12ZoneOf(g, victim); z != game.ZoneGraveyard {
		t.Errorf("the victim is in %s, want the graveyard", z)
	}
}

// The printed clause says "you OWN", not "you control": a permanent
// this seat controls but does not own (a stolen one) is not a legal
// target.
func TestStaffOfCompleationCannotDestroyAPermanentItOnlyControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := staffOfCompleation(g, me.ID)
	stolenFromOpp := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Stolen Rock", TypeLine: "Artifact",
		Owner: opp.ID, Controller: me.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: stolenFromOpp}},
	}); err == nil {
		t.Fatal("activated Staff of Compleation targeting a permanent I control but do not own")
	}
	card, ok := effectstest.BattlefieldCard(g, staff)
	if !ok || card.Tapped {
		t.Error("the rejected activation tapped the staff anyway")
	}
}

// The flip side: a permanent this seat OWNS but does not control (an
// opponent stole it) IS a legal target — this is the card's real use
// case, taking a permanent back from an opponent's control by
// destroying it.
func TestStaffOfCompleationCanDestroyItsOwnPermanentEvenIfSomeoneElseControlsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := staffOfCompleation(g, me.ID)
	mineButTheirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Stolen Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: opp.ID,
	})

	effectstest.B16Activate(t, g, me.ID, staff, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mineButTheirs}},
	})

	if z := effectstest.B12ZoneOf(g, mineButTheirs); z != game.ZoneGraveyard {
		t.Errorf("the reclaimed permanent is in %s, want the graveyard", z)
	}
}

// {T}, Pay 3 life: Proliferate.
func TestStaffOfCompleationProliferatesForThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := staffOfCompleation(g, me.ID)
	mine := effectstest.PushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 2)
	before := me.Life

	effectstest.B16Activate(t, g, me.ID, staff, 1, game.ActivateAbilityParams{})

	if me.Life != before-3 {
		t.Errorf("life %d -> %d, want %d", before, me.Life, before-3)
	}
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}

// {T}, Pay 4 life: Draw a card.
func TestStaffOfCompleationDrawsACardForFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := staffOfCompleation(g, me.ID)
	effectstest.SeedLibrary(me, "Drawn")
	before := me.Life
	handBefore := len(me.Hand.Cards)

	effectstest.B16Activate(t, g, me.ID, staff, 2, game.ActivateAbilityParams{})

	if me.Life != before-4 {
		t.Errorf("life %d -> %d, want %d", before, me.Life, before-4)
	}
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand %d -> %d, want %d", handBefore, got, handBefore+1)
	}
}

// {5}: Untap this artifact — lets a second ability activate the same
// turn, exactly what Fain, the Broker's "{3}{B}: Untap Fain" buys.
func TestStaffOfCompleationUntapsForFiveAndAllowsASecondActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	staff := staffOfCompleation(g, me.ID)
	effectstest.SeedLibrary(me, "First Draw", "Second Draw")

	// First activation: draw a card, taps the staff.
	effectstest.B16Activate(t, g, me.ID, staff, 2, game.ActivateAbilityParams{})
	if card, ok := effectstest.BattlefieldCard(g, staff); !ok || !card.Tapped {
		t.Fatal("the staff did not tap for its first activation")
	}

	// {5}: untap.
	effectstest.FloatForTest(g, me, "CCCCC")
	effectstest.B16Activate(t, g, me.ID, staff, 3, game.ActivateAbilityParams{})
	if card, ok := effectstest.BattlefieldCard(g, staff); !ok || card.Tapped {
		t.Fatal("the staff did not untap")
	}

	// Second activation the same turn.
	handBefore := len(me.Hand.Cards)
	effectstest.B16Activate(t, g, me.ID, staff, 2, game.ActivateAbilityParams{})
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("second draw did not happen: hand %d -> %d, want %d", handBefore, got, handBefore+1)
	}
}
