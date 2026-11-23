package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b178_test.go — GAP batch 178 (#5948): Jumpstart 2022 sorceries.

const (
	gapB178InterpretTheSigns = "834554f6-3b65-44fd-b1d1-8495b740dbfc"
	gapB178MultipleChoice    = "b4d23acf-ff1f-4a6a-9ede-2667a99e9665"
	gapB178NotForgotten      = "e0fa6d08-1808-4a29-b1e4-5cec9d31d798"
	gapB178SettleBeyond      = "2debd9f1-6d3a-4f0a-8488-603447c44490"
)

func TestGapB178EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{gapB178InterpretTheSigns, gapB178MultipleChoice, gapB178NotForgotten, gapB178SettleBeyond} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestGapB178InterpretTheSignsDrawsTheRevealedCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	revealed := effectstest.SeedLibraryTop(me, "Three Drop", "Creature — Bear", "{2}{U}")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Interpret the Signs", "Sorcery", gapB178InterpretTheSigns, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand grew by %d, want 3 (the revealed card's mana value)", got-hand)
	}
	if !me.Hand.Contains(revealed) {
		t.Error("the revealed card should be the first one drawn")
	}
}

func TestGapB178InterpretTheSignsScryDecidesWhatIsRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	big := effectstest.SeedLibraryTop(me, "Five Drop", "Creature — Bear", "{4}{U}")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Interpret the Signs", "Sorcery", gapB178InterpretTheSigns, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	// Put the five-drop on the bottom: the new top card is a 0.
	var rest []uuid.UUID
	for _, id := range c.ScryCards {
		if id != big {
			rest = append(rest, id)
		}
	}
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{big}, rest); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand grew by %d, want 0 (the new top card has mana value 0)", got-hand)
	}
}

func TestGapB178NotForgottenPutsAGraveyardCardOnTopAndMakesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	card := effectstest.PushGraveyardCard(g, opp, "Their Card")
	before := effectstest.CountNamed(g, "Spirit")
	effectstest.CastCatalogSpell(t, g, "Not Forgotten", "Sorcery", gapB178NotForgotten, effectstest.CardTarget(card))
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("the caster was not asked for the top or bottom")
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, nil, []uuid.UUID{card}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != card {
		t.Error("the card is not on top of its owner's library")
	}
	if got := effectstest.CountNamed(g, "Spirit"); got != before+1 {
		t.Errorf("Spirit tokens %d, want %d", got, before+1)
	}
}

func TestGapB178NotForgottenCanPutItOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.SeedLibrary(opp, "Theirs")
	card := effectstest.PushGraveyardCard(g, me, "My Card")
	effectstest.CastCatalogSpell(t, g, "Not Forgotten", "Sorcery", gapB178NotForgotten, effectstest.CardTarget(card))
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if ask == nil {
		t.Fatal("the caster was not asked for the top or bottom")
	}
	if err := g.ResolvePutInLibrary(ask.ID, me.ID, []uuid.UUID{card}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if bottom, err := me.Library.Bottom(); err != nil || bottom.InstanceID != card {
		t.Error("the card is not on the bottom of its owner's library")
	}
}

func TestGapB178NotForgottenRefusesABattlefieldCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Not Forgotten", "Sorcery", gapB178NotForgotten, effectstest.CardTarget(bear)); err == nil {
		t.Error("Not Forgotten targeted a permanent")
	}
}

func TestGapB178SettleBeyondRealityExilesAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.P1g2Cast(t, g, "Settle Beyond Reality", "Sorcery", gapB178SettleBeyond, "", game.CastSpellParams{
		Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0, Slot: 0}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the opposing creature")
}

func TestGapB178SettleBeyondRealityBothBulletsExileAndBlink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.StdB9AddCounter(t, g, mine, game.CounterPlusOne, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.P1g2Cast(t, g, "Settle Beyond Reality", "Sorcery", gapB178SettleBeyond, "", game.CastSpellParams{
		Modes: []int{0, 1}, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: theirs, Mode: 0, Slot: 0},
			{Kind: game.TargetCard, ID: mine, Mode: 1, Slot: 0},
		},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, theirs, game.ZoneExile, "the opposing creature")
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mine" {
			if c.Counters[game.CounterPlusOne] != 0 {
				t.Error("the blinked creature kept its counter: it is not a new object")
			}
			return
		}
	}
	t.Error("the creature I control did not return to the battlefield")
}

func TestGapB178SettleBeyondRealityRefusesAFriendlyExileTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Settle Beyond Reality", TypeLine: "Sorcery", OracleID: gapB178SettleBeyond,
		Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Modes: []int{0}, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine, Mode: 0, Slot: 0}},
	})
	if err == nil {
		t.Error("the exile bullet accepted a creature I control")
	}
}

// --- Multiple Choice ----------------------------------------------------

func gapB178CastMultipleChoice(t *testing.T, g *game.Game, x int) {
	t.Helper()
	effectstest.CastXSpell(t, g, "Multiple Choice", "Sorcery", gapB178MultipleChoice, "{X}{U}", x, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB178MultipleChoiceXOneScriesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	tokens := effectstest.CountNamed(g, "Elemental")
	gapB178CastMultipleChoice(t, g, 1)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand grew by %d, want 1", got-hand)
	}
	if got := effectstest.CountNamed(g, "Elemental"); got != tokens {
		t.Error("X=1 made a token")
	}
}

func TestGapB178MultipleChoiceXTwoBouncesTheChosenPlayersCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	gapB178CastMultipleChoice(t, g, 2)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
	if pick == nil {
		t.Fatal("the chosen player was not asked which creature to return")
	}
	if err := g.ResolveOwnPermanents(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
	if got := me.Hand.Size(); got != hand {
		t.Error("X=2 drew a card")
	}
}

func TestGapB178MultipleChoiceXTwoMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	gapB178CastMultipleChoice(t, g, 2)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was returned although the caster declined")
	}
}

func TestGapB178MultipleChoiceXThreeMakesAnElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hand := effectstest.Top100ActiveSeat(g).Hand.Size()
	before := effectstest.CountNamed(g, "Elemental")
	gapB178CastMultipleChoice(t, g, 3)
	if got := effectstest.CountNamed(g, "Elemental"); got != before+1 {
		t.Errorf("Elemental tokens %d, want %d", got, before+1)
	}
	if got := effectstest.Top100ActiveSeat(g).Hand.Size(); got != hand {
		t.Error("X=3 drew a card")
	}
}

func TestGapB178MultipleChoiceXFourDoesAllOfIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	before := effectstest.CountNamed(g, "Elemental")
	gapB178CastMultipleChoice(t, g, 4)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.AnswerChoosePlayer(t, g, me.ID, me)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("the chosen player was not asked which creature to return")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	_ = opp
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand grew by %d, want 2 (the drawn card and the bounced creature)", got-hand)
	}
	if got := effectstest.CountNamed(g, "Elemental"); got != before+1 {
		t.Errorf("Elemental tokens %d, want %d", got, before+1)
	}
}
