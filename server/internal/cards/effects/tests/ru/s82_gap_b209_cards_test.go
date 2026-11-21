package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB209Ground    = "13613f6e-a6e4-4493-b029-8ec11610c29f"
	gapB209Guided    = "f4b30f6c-09ac-407e-9c72-da64d212ce93"
	gapB209Justice   = "333809cb-e196-45f2-8a67-31374438e56e"
	gapB209Gambit    = "c8df3d1f-130e-4391-9bec-3f2d345a13fb"
	gapB209Breakthru = "c6b26c1f-121f-4b7c-921e-429b623ca934"
)

func TestGapB209EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB209Ground: "Ground Assault", gapB209Guided: "Guided Passage", gapB209Justice: "Fiery Justice",
		gapB209Gambit: "Sparkmage's Gambit", gapB209Breakthru: "Breakthrough",
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

func TestGapB209GroundAssaultDealsDamageEqualToYourLands(t *testing.T) {
	for _, tc := range []struct {
		name  string
		extra int
		dies  bool
	}{{"toughness equal to your lands dies", 0, true}, {"one more toughness survives", 1, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			for i := 0; i < 3; i++ {
				effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
			}
			for i := 0; i < 4; i++ { // not yours: not counted
				effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
			}
			target := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 1, 3+tc.extra)
			effectstest.CastCatalogSpell(t, g, "Ground Assault", "Sorcery", gapB209Ground, effectstest.CardRefs(target))
			effectstest.PassPriorityAroundTable(t, g)
			if died := !g.Battlefield.Contains(target); died != tc.dies {
				t.Errorf("died=%v, want %v", died, tc.dies)
			}
		})
	}
}

func TestGapB209GroundAssaultRefusesAPlayerTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Ground Assault", "Sorcery", gapB209Ground,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if err == nil {
		t.Error("Ground Assault targeted a player")
	}
}

func TestGapB209FieryJusticeDividesFiveAndTheOpponentGainsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp2.ID, "Ogre", "Creature — Ogre", 3, 3)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fiery Justice", TypeLine: "Sorcery", ManaCost: "{R}{G}{W}", OracleID: gapB209Justice, Owner: me.ID, Controller: me.ID})
	life, life2 := opp.Life, opp2.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: ogre, Slot: 0},
			{Kind: game.TargetPlayer, ID: opp2.ID, Slot: 0},
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
		},
		Distribution: map[uuid.UUID]int{ogre: 3, opp2.ID: 2},
	}); err != nil {
		t.Fatalf("cast Fiery Justice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived 3 damage")
	}
	if opp2.Life != life2-2 {
		t.Errorf("damaged opponent life %d, want %d", opp2.Life, life2-2)
	}
	if opp.Life != life+5 {
		t.Errorf("target opponent life %d, want %d", opp.Life, life+5)
	}
}

func TestGapB209FieryJusticeNeedsAnOpponentToGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fiery Justice", TypeLine: "Sorcery", ManaCost: "{R}{G}{W}", OracleID: gapB209Justice, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID, Slot: 0}, {Kind: game.TargetPlayer, ID: me.ID, Slot: 1}},
		Distribution: map[uuid.UUID]int{g.Seats[1].ID: 5},
		Strict:       true,
	})
	if err == nil {
		t.Error("Fiery Justice let you gain the life yourself")
	}
}

func TestGapB209SparkmagesGambitPingsTwoAndStopsThemBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	atk := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Mouse", "Creature — Mouse", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sparkmage's Gambit", "Sorcery", gapB209Gambit, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the 1/1 survived 1 damage")
	}
	if bear, _ := effectstest.BattlefieldCard(g, b); bear.DamageMarked != 1 {
		t.Errorf("Bear has %d damage, want 1", bear.DamageMarked)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(b, atk); err == nil {
		t.Error("the targeted Bear can't block this turn")
	}
	if err := g.DeclareBlocker(c, atk); err != nil {
		t.Errorf("an untargeted creature still blocks: %v", err)
	}
}

func TestGapB209BreakthroughKeepsXCardsOfYourWholeHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Breakthrough", TypeLine: "Sorcery", ManaCost: "{X}{U}", OracleID: gapB209Breakthru, Owner: me.ID, Controller: me.ID})
	before := me.Hand.Size() - 1 // cards left in hand once it is cast
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast Breakthrough: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if owed := effectstest.DiscardOwed(g, me.ID); owed != before+4-2 {
		t.Fatalf("discard owed %d, want %d", owed, before+4-2)
	}
}

func TestGapB209BreakthroughDiscardsNothingWhenTheHandIsSmall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Breakthrough", TypeLine: "Sorcery", ManaCost: "{X}{U}", OracleID: gapB209Breakthru, Owner: me.ID, Controller: me.ID})
	size := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: size + 4}); err != nil {
		t.Fatalf("cast Breakthrough: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if owed := effectstest.DiscardOwed(g, me.ID); owed != 0 {
		t.Errorf("discard owed %d, want 0", owed)
	}
	if me.Hand.Size() != size-1+4 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), size-1+4)
	}
}

func TestGapB209GuidedPassageOpponentChoosesOneOfEachKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear1 := effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
	bear2 := effectstest.SeedLibraryTop(me, "Elk", "Creature — Elk", "{1}{G}")
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	bolt := effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	ring := effectstest.SeedLibraryTop(me, "Ring", "Artifact", "{1}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Guided Passage", TypeLine: "Sorcery", ManaCost: "{G}{U}{R}", OracleID: gapB209Guided, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Guided Passage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.AnswerChooseCards(t, g, opp.ID, bear2)
	effectstest.AnswerChooseCards(t, g, opp.ID, land)
	effectstest.AnswerChooseCards(t, g, opp.ID, ring)
	for _, want := range []uuid.UUID{bear2, land, ring} {
		if !me.Hand.Contains(want) {
			t.Errorf("chosen card %v is not in hand", want)
		}
	}
	for _, left := range []uuid.UUID{bear1, bolt} {
		if me.Hand.Contains(left) || !me.Library.Contains(left) {
			t.Errorf("unchosen card %v left the library", left)
		}
	}
}

func TestGapB209GuidedPassageSkipsAKindThatIsNotThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	me.Library.Cards = nil // no noncreature, nonland card left to choose
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	bear := effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Guided Passage", TypeLine: "Sorcery", ManaCost: "{G}{U}{R}", OracleID: gapB209Guided, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Guided Passage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.AnswerChooseCards(t, g, opp.ID, bear)
	effectstest.AnswerChooseCards(t, g, opp.ID, land)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(land) {
		t.Error("the two available kinds were not put into hand")
	}
}
