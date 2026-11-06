package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_b2_cards_test.go — Outlaws of Thunder Junction (OTJ), batch 2 (#4468):
// the D–F cards that need no new engine seam. Each test plays the real catalog
// entry through the engine and asserts what the printed text says.

const (
	otjB2ThunderThiefOracle = "0f5401c2-0189-4a07-8e7b-a765ccd7c954"
	otjB2DeadeyeOracle      = "29ef90ac-2396-4488-970d-d440297ba5ee"
	otjB2DeepmuckOracle     = "b75ffa62-a291-44d7-a950-b50c8394e066"
	otjB2DesertsDueOracle   = "909eacf3-05ce-4c2b-97a8-49b6ea597b64"
	otjB2BloodseekerOracle  = "de8d7cad-02c7-43f1-896e-eaa73f016ada"
	otjB2PeddlerOracle      = "a6584af7-583d-4d89-9b46-d61be8236413"
	otjB2LullabyOracle      = "b134a9e6-b004-4833-b298-10529f5a85b8"
	otjB2FordingOracle      = "fe4bbe82-1d6b-4f3d-8cc0-f6df22d9e9fb"
	otjB2GulchOracle        = "9d3b60af-3e38-4d36-95fc-11b31c38f955"
	otjB2FlatsOracle        = "ebb3e2ff-2214-4e11-88fb-e0fa84288cf1"

	otjB2DesertTypeLine = "Land — Desert"
)

func TestOtjB2EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		otjB2ThunderThiefOracle: "Daring Thunder-Thief",
		otjB2DeadeyeOracle:      "Deadeye Duelist",
		otjB2DeepmuckOracle:     "Deepmuck Desperado",
		otjB2DesertsDueOracle:   "Desert's Due",
		otjB2BloodseekerOracle:  "Desperate Bloodseeker",
		otjB2PeddlerOracle:      "Discerning Peddler",
		otjB2LullabyOracle:      "Eriette's Lullaby",
		otjB2FordingOracle:      "Failed Fording",
		otjB2GulchOracle:        "Festering Gulch",
		otjB2FlatsOracle:        "Forlorn Flats",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// otjB2Desert seeds a plain Desert land under `owner`.
func otjB2Desert(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Desert Land", otjB2DesertTypeLine, "", false)
}

// --- Daring Thunder-Thief ----------------------------------------------------

func TestOtjB2DaringThunderThiefHasFlashAndEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// An end step is no sorcery window: only flash lets a creature be cast.
	advanceTo(t, g, game.StepEnd)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Daring Thunder-Thief", TypeLine: "Creature — Turtle Rogue",
		OracleID: otjB2ThunderThiefOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash: cast in the end step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Thief did not resolve onto the battlefield")
	}
	if !b16Tapped(t, g, id) {
		t.Error("the Thief enters tapped")
	}
}

// --- Deadeye Duelist ---------------------------------------------------------

func TestOtjB2DeadeyeDuelistPaysOneAndTapsToPingAnOpponentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duelist := b12Push(g, me.ID, "Deadeye Duelist", "Creature — Human Assassin", otjB2DeadeyeOracle, 1, 3)
	if !hasEffectiveKeyword(t, g, duelist, "reach") {
		t.Error("reach is printed")
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, duelist, 0, game.ActivateAbilityParams{
		Targets: b16TargetPlayer(opp.ID), Strict: true,
	}); err == nil {
		t.Fatal("the {1} must be paid")
	}
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, duelist, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(me.ID)}); err == nil {
		t.Fatal("you are not a legal target: it says target opponent")
	}
	before := opp.Life
	b16Activate(t, g, me.ID, duelist, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("target opponent takes 1: %d → %d", before, opp.Life)
	}
	if !b16Tapped(t, g, duelist) {
		t.Error("the Duelist taps to activate")
	}
}

// --- Deepmuck Desperado ------------------------------------------------------

func TestOtjB2DeepmuckDesperadoMillsEachOpponentThreeOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		msc18Library(p, 10)
	}
	b12Push(g, me.ID, "Deepmuck Desperado", "Creature — Homarid Mercenary", otjB2DeepmuckOracle, 2, 4)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	graveyards := func() []int {
		out := make([]int, len(g.Seats))
		for i, p := range g.Seats {
			out[i] = p.Graveyard.Size()
		}
		return out
	}

	// Targeting your own creature is not a crime.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	start := graveyards()
	for i := 1; i < len(g.Seats); i++ {
		if g.Seats[i].Graveyard.Size() != 0 {
			t.Fatalf("seat %d milled with no crime committed", i)
		}
	}

	// Targeting an opponent is: every opponent mills three, you do not.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	after := graveyards()
	for i := 1; i < len(g.Seats); i++ {
		if after[i]-start[i] != 3 {
			t.Errorf("opponent at seat %d milled %d, want 3", i, after[i]-start[i])
		}
	}
	if after[0]-start[0] != 1 {
		t.Errorf("your graveyard grew by %d, want only the Bolt", after[0]-start[0])
	}

	// A second crime the same turn does not trigger it again.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	again := graveyards()
	for i := 1; i < len(g.Seats); i++ {
		if again[i] != after[i] {
			t.Errorf("seat %d milled again on the same turn: %d → %d", i, after[i], again[i])
		}
	}

	// A new turn of yours, a new trigger.
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	next := graveyards()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	last := graveyards()
	for i := 1; i < len(g.Seats); i++ {
		if last[i]-next[i] != 3 {
			t.Errorf("next turn: seat %d milled %d, want 3", i, last[i]-next[i])
		}
	}
}

// --- Desert's Due ------------------------------------------------------------

func TestOtjB2DesertsDueGivesMinusTwoAndMinusOneMorePerDesert(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	otjB2Desert(g, me.ID)
	otjB2Desert(g, me.ID)
	otjB2Desert(g, opp.ID) // an opponent's Desert is not yours
	pushCatalogPermanent(g, me.ID, "Plain Forest", "Basic Land — Forest", "", false)
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	castCatalogSpell(t, g, "Desert's Due", "Instant", otjB2DesertsDueOracle, b16TargetCard(giant))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, giant); p != 2 || tt != 2 {
		t.Errorf("6/6 with two Deserts of yours: %d/%d, want 2/2", p, tt)
	}
}

func TestOtjB2DesertsDueWithNoDesertIsMinusTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Desert's Due", "Instant", otjB2DesertsDueOracle, b16TargetCard(ogre))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, ogre); p != 1 || tt != 1 {
		t.Errorf("3/3 with no Desert: %d/%d, want 1/1", p, tt)
	}
}

// --- Desperate Bloodseeker ---------------------------------------------------

func TestOtjB2DesperateBloodseekerMakesTargetPlayerMillTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(opp, 5)
	library, graveyard := opp.Library.Size(), opp.Graveyard.Size()
	seeker := msc17Play(t, g, me, game.Card{Name: "Desperate Bloodseeker", TypeLine: "Creature — Vampire",
		ManaCost: "{1}{B}", OracleID: otjB2BloodseekerOracle, Power: 2, Toughness: 2})
	if !hasEffectiveKeyword(t, g, seeker, "lifelink") {
		t.Error("lifelink is printed")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil || !hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the enters trigger asks for any player, you included: %+v", p)
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - graveyard; got != 2 {
		t.Errorf("target player milled %d, want 2", got)
	}
	if got := opp.Library.Size(); got != library-2 {
		t.Errorf("target player's library %d, want %d", got, library-2)
	}
}

// --- Discerning Peddler ------------------------------------------------------

func otjB2PlayPeddler(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	msc17Play(t, g, me, game.Card{Name: "Discerning Peddler", TypeLine: "Creature — Human Rogue", ManaCost: "{1}{R}",
		OracleID: otjB2PeddlerOracle, Power: 2, Toughness: 2})
}

func TestOtjB2DiscerningPeddlerMayDiscardAndIfSoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn")
	discard := pushHandCard(g, me)
	otjB2PlayPeddler(t, g, me)
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, discard)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand after discard one, draw one: %d, want %d", got, hand)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the chosen card was not discarded")
	}
}

func TestOtjB2DiscerningPeddlerDeclinedDiscardDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Drawn")
	pushHandCard(g, me)
	otjB2PlayPeddler(t, g, me)
	hand, library := me.Hand.Size(), me.Library.Size()
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Library.Size() != library {
		t.Errorf("declining drew anyway: hand %d -> %d, library %d -> %d", hand, me.Hand.Size(), library, me.Library.Size())
	}
}

// --- Eriette's Lullaby -------------------------------------------------------

func TestOtjB2EriettesLullabyDestroysATappedCreatureAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Their Untapped Bear", "Creature — Bear", 2, 2)
	tapped := b12Creature(g, opp.ID, "Their Tapped Bear", "Creature — Bear", 2, 2)
	b16Tap(g, tapped)
	if err := castCatalogSpellErr(t, g, "Eriette's Lullaby", "Sorcery", otjB2LullabyOracle, b16TargetCard(untapped)); err == nil {
		t.Fatal("an untapped creature was accepted as the target tapped creature")
	}
	before := me.Life
	castCatalogSpell(t, g, "Eriette's Lullaby", "Sorcery", otjB2LullabyOracle, b16TargetCard(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !opp.Graveyard.Contains(tapped) {
		t.Error("the tapped creature is destroyed")
	}
	if me.Life != before+2 {
		t.Errorf("you gain 2 life: %d → %d", before, me.Life)
	}
}

func TestOtjB2EriettesLullabyDoesNothingIfTheTargetUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b16Tap(g, bear)
	before := me.Life
	castCatalogSpell(t, g, "Eriette's Lullaby", "Sorcery", otjB2LullabyOracle, b16TargetCard(bear))
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bear) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature that untapped in response is no longer a legal target")
	}
	if me.Life != before {
		t.Errorf("with no legal target the spell does nothing, life %d → %d", before, me.Life)
	}
}

// --- Failed Fording ----------------------------------------------------------

func TestOtjB2FailedFordingBouncesThenSurveilsWithADesert(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 6)
	otjB2Desert(g, me.ID)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := pushCatalogPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", false)
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	hand := opp.Hand.Size()

	if err := castCatalogSpellErr(t, g, "Failed Fording", "Instant", otjB2FordingOracle, cardRefs(land)); err == nil {
		t.Fatal("a land was accepted as the target nonland permanent")
	}
	castCatalogSpell(t, g, "Failed Fording", "Instant", otjB2FordingOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || opp.Hand.Size() != hand+1 {
		t.Fatal("the Bear was not returned to its owner's hand")
	}
	prompt := surveilChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("you control a Desert: no surveil prompt")
	}
	if err := g.ResolveSurveil(prompt.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("surveil away the top card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not reach the graveyard")
	}
}

func TestOtjB2FailedFordingWithoutADesertOnlyBounces(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 6)
	otjB2Desert(g, opp.ID) // an opponent's Desert is not yours
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Failed Fording", "Instant", otjB2FordingOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Fatal("the Bear was not returned to its owner's hand")
	}
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("no Desert of yours: no surveil")
	}
}

// --- Festering Gulch, Forlorn Flats ------------------------------------------

func TestOtjB2DesertDualsEnterTappedPingAnOpponentAndTapForTheirColours(t *testing.T) {
	for _, row := range []struct{ name, oracle, first, second string }{
		{"Festering Gulch", otjB2GulchOracle, "B", "G"},
		{"Forlorn Flats", otjB2FlatsOracle, "W", "B"},
	} {
		for _, colour := range []string{row.first, row.second} {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[2]
			before := opp.Life
			land := b12PlayFromHand(t, g, row.name, otjB2DesertTypeLine, row.oracle, game.CastSpellParams{})
			if !b16Tapped(t, g, land) {
				t.Fatalf("%s: the Desert enters tapped", row.name)
			}
			if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
				t.Fatalf("%s: the trigger asks for an opponent, never you: %+v", row.name, p)
			}
			b16PickPlayer(t, g, me.ID, opp.ID)
			passPriorityAroundTable(t, g)
			if opp.Life != before-1 {
				t.Errorf("%s: target opponent takes 1: %d → %d", row.name, before, opp.Life)
			}
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
			b28TapForMana(t, g, me.ID, land, colour)
			if got := poolColors(me); len(got) != 1 || got[0] != colour {
				t.Errorf("%s: tapped for {%s}: pool %v", row.name, colour, got)
			}
		}
	}
}
