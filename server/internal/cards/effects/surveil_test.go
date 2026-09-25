package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// surveil_test.go — CR 701.25 surveil, and the surveil-land cycle
// that finally does what its type line says.
//
// Surveil is scry with one word changed, so the tests that matter are
// the ones that would still pass if it had been built as scry:
//
//   - the cards that leave the top land in the GRAVEYARD, not under
//     the library. That is the entire keyword. A reanimator deck
//     surveils to fill a graveyard, and bottoming instead is the
//     difference between a resource and a card that is gone.
//   - they emit EventMill on the way, the same event an ordinary mill
//     emits, because it is the same zone change. A "whenever a card is
//     put into your graveyard from your library" payoff must not care
//     which keyword moved it.
//   - surveil is LOOK AT, not reveal — same privacy rule as scry.
//   - the scry resolver must refuse a surveil prompt and vice versa,
//     because both answers carry a top_order and only the other key
//     tells them apart.

const undercitySewersOracle = "08d80efc-9542-4ba2-824c-c8615d8d07f2"

func surveilChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceSurveil && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// graveyardNames reads the graveyard bottom-to-top.
func graveyardNames(p *game.Player) []string {
	out := make([]string, 0, p.Graveyard.Size())
	for _, c := range p.Graveyard.Cards {
		out = append(out, c.Name)
	}
	return out
}

// --- the surveil lands --------------------------------------------

func TestSurveilLandEntersTappedAndSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Next Draw")

	id := playLandFromHand(t, g, "Undercity Sewers", undercitySewersOracle)
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the land isn't on the battlefield")
	}
	if !card.Tapped {
		t.Error("the surveil land entered untapped")
	}
	// The enters-tapped half is a CR 614 replacement, so nothing was
	// ever tapped — same discriminator the Temples use.
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("playing the surveil land did not queue a surveil")
	}
}

// TestRaucousTheaterAndThunderingFallsEnterTappedAndSurveil pins the
// two newest rows in the table — same shape as
// TestSurveilLandEntersTappedAndSurveils, the risk being the row's own
// oracle ID and colours rather than the shared machinery.
func TestRaucousTheaterAndThunderingFallsEnterTappedAndSurveil(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
	}{
		{"Raucous Theater", "04e5e84f-8fd4-43ab-8f9d-5b24646f7ae5"},
		{"Thundering Falls", "d2bcff58-7a8a-46ef-b6b3-39501d4c8e6e"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibrary(me, "Next Draw")

			id := playLandFromHand(t, g, tc.name, tc.oracle)
			passPriorityAroundTable(t, g)

			card, ok := battlefieldCard(g, id)
			if !ok {
				t.Fatal("the land isn't on the battlefield")
			}
			if !card.Tapped {
				t.Error("the surveil land entered untapped")
			}
			if n := tapEventsFor(g, id); n != 0 {
				t.Errorf("%d tap events; the land should have ENTERED tapped", n)
			}
			if surveilChoiceFor(g, me.ID) == nil {
				t.Fatal("playing the surveil land did not queue a surveil")
			}
		})
	}
}

// TestSurveilBinsToGraveyardNotBottom is the assertion the whole
// primitive exists for.
func TestSurveilBinsToGraveyardNotBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Fodder", "Keeper")
	libBefore := me.Library.Size()

	playLandFromHand(t, g, "Undercity Sewers", undercitySewersOracle)
	passPriorityAroundTable(t, g)

	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("looked at %d cards, want 1 (surveil 1)", len(c.ScryCards))
	}
	// Nothing moves until the answer arrives.
	if got := libraryTopNames(me, 1); got[0] != "Fodder" {
		t.Errorf("library top is %q before the answer; surveil must not move anything early", got[0])
	}

	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := graveyardNames(me); len(got) != 1 || got[0] != "Fodder" {
		t.Errorf("graveyard is %v, want [Fodder] — surveil bins, it does not bottom", got)
	}
	if got := libraryTopNames(me, 1); got[0] != "Keeper" {
		t.Errorf("library top is %q, want Keeper", got[0])
	}
	if bottom, _ := me.Library.Bottom(); bottom.Name == "Fodder" {
		t.Error("the surveilled card went to the BOTTOM of the library; that is scry, not surveil")
	}
	if me.Library.Size() != libBefore-1 {
		t.Errorf("library is %d, want %d — the binned card must leave it", me.Library.Size(), libBefore-1)
	}
}

func TestSurveilKeepingOnTopLeavesTheLibraryAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Keeper", "Second")
	libBefore := me.Library.Size()

	playLandFromHand(t, g, "Meticulous Archive", "ccfb8b4d-651c-418a-aa19-cb23105b3f2f")
	passPriorityAroundTable(t, g)

	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := graveyardNames(me); len(got) != 0 {
		t.Errorf("graveyard is %v, want empty — nothing was binned", got)
	}
	if got := libraryTopNames(me, 2); got[0] != "Keeper" || got[1] != "Second" {
		t.Errorf("top two are %v, want [Keeper Second]", got)
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library is %d, want %d unchanged", me.Library.Size(), libBefore)
	}
}

// TestSurveilEmitsMillAndSurveilEvents — the binned card takes the
// ordinary library-to-graveyard route, so graveyard payoffs see it;
// the completed keyword gets its own event on top, so surveil payoffs
// see that and an ordinary scry does not fire them.
func TestSurveilEmitsMillAndSurveilEvents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Fodder")

	playLandFromHand(t, g, "Shadowy Backstreet", "216a2a92-9ca3-4ca3-8af7-686c13b04290")
	passPriorityAroundTable(t, g)

	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	binned := c.ScryCards[0]
	before := len(g.Events)
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}

	var mills, surveils, scries int
	for _, ev := range g.Events[before:] {
		switch ev.Kind {
		case game.EventMill:
			if ev.CardID == binned && ev.OldZone == game.ZoneLibrary && ev.NewZone == game.ZoneGraveyard {
				mills++
			}
		case game.EventSurveil:
			surveils++
			if ev.Amount != 1 {
				t.Errorf("EventSurveil amount = %d, want 1 (one card binned)", ev.Amount)
			}
		case game.EventScry:
			scries++
		}
	}
	if mills != 1 {
		t.Errorf("%d library-to-graveyard mill events, want 1", mills)
	}
	if surveils != 1 {
		t.Errorf("%d surveil events, want 1", surveils)
	}
	if scries != 0 {
		t.Error("a surveil emitted EventScry; a scry payoff must not fire on a surveil")
	}
}

// --- the "then" continuation ---------------------------------------

// TestSurveilThenRunsAfterTheAnswer — the same ordering guarantee
// Preordain needs from Scry.Then. A continuation that ran eagerly
// would see the pre-surveil library.
func TestSurveilThenRunsAfterTheAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Fodder", "Drawn")
	handBefore := me.Hand.Size()

	g.WithWriteLock(func() {
		g.SurveilThenForEffect(me.ID, uuid.Nil, 1, func(gg *game.Game) error {
			return gg.DrawNForEffect(me.ID, 1)
		})
	})

	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if me.Hand.Size() != handBefore {
		t.Fatal("the continuation drew a card BEFORE the player answered")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if me.Hand.Size() != handBefore+1 {
		t.Fatalf("hand is %d, want %d — the continuation never ran", me.Hand.Size(), handBefore+1)
	}
	// Fodder was binned, so the draw is the card under it.
	top, _ := me.Hand.Top()
	if top.Name != "Drawn" {
		t.Errorf("drew %q, want Drawn — the bin must happen before the draw", top.Name)
	}
}

// --- rejections -----------------------------------------------------

func TestSurveilRejectsBadAnswers(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.PendingChoice) {
		t.Helper()
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seedLibrary(me, "Top", "Second")
		playLandFromHand(t, g, "Undercity Sewers", undercitySewersOracle)
		passPriorityAroundTable(t, g)
		c := surveilChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no surveil prompt")
		}
		return g, me, c
	}

	t.Run("a card listed twice", func(t *testing.T) {
		g, me, c := setup(t)
		if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, c.ScryCards); err == nil {
			t.Error("the same card was accepted in both lists")
		}
		if me.Graveyard.Size() != 0 {
			t.Error("a rejected answer moved cards anyway")
		}
	})

	t.Run("a card left out", func(t *testing.T) {
		g, me, c := setup(t)
		if err := g.ResolveSurveil(c.ID, me.ID, nil, nil); err == nil {
			t.Error("an answer that named none of the looked-at cards was accepted")
		}
		if me.Graveyard.Size() != 0 {
			t.Error("a rejected answer moved cards anyway")
		}
	})

	t.Run("someone else answering", func(t *testing.T) {
		g, me, c := setup(t)
		other := g.Seats[1]
		if other.ID == me.ID {
			t.Skip("single-seat table")
		}
		if err := g.ResolveSurveil(c.ID, other.ID, c.ScryCards, nil); err == nil {
			t.Error("another player answered my surveil")
		}
	})

	// The two resolvers must not accept each other's prompts. Both
	// answers carry a top_order, so a routing slip would silently
	// bottom cards that should have been binned.
	t.Run("the scry resolver refuses a surveil", func(t *testing.T) {
		g, me, c := setup(t)
		if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err == nil {
			t.Error("ResolveScry answered a surveil prompt")
		}
		if me.Graveyard.Size() != 0 {
			t.Error("the misrouted answer moved cards")
		}
	})
}

func TestScryResolverRefusesSurveilAndViceVersa(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top")
	playLandFromHand(t, g, "Temple of Silence", templeOfSilenceOracle)
	passPriorityAroundTable(t, g)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err == nil {
		t.Error("ResolveSurveil answered a scry prompt — that would bin a card meant for the bottom")
	}
	if me.Graveyard.Size() != 0 {
		t.Error("the misrouted answer put a card in the graveyard")
	}
}

// --- privacy --------------------------------------------------------

// TestSurveilIsLookAtNotReveal — same rule as scry. The chooser sees
// faces; everyone else sees backs.
func TestSurveilIsLookAtNotReveal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opp *game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opp = p
			break
		}
	}
	if opp == nil {
		t.Skip("single-seat table")
	}
	seedLibrary(me, "Secret Top")

	playLandFromHand(t, g, "Undercity Sewers", undercitySewersOracle)
	passPriorityAroundTable(t, g)

	find := func(v protocol.GameView) *protocol.PendingChoiceView {
		for i := range v.PendingChoices {
			if v.PendingChoices[i].Kind == "surveil" && v.PendingChoices[i].Chooser == me.ID.String() {
				return &v.PendingChoices[i]
			}
		}
		return nil
	}

	mine := find(protocol.ViewOfGameFor(g, me.ID.String()))
	if mine == nil {
		t.Fatal("the surveilling player can't see their own prompt")
	}
	if len(mine.Options) != 1 {
		t.Fatalf("chooser sees %d options, want 1", len(mine.Options))
	}
	if mine.Options[0].Name != "Secret Top" {
		t.Errorf("chooser sees %q, want the real card name", mine.Options[0].Name)
	}

	theirs := find(protocol.ViewOfGameFor(g, opp.ID.String()))
	if theirs != nil {
		for _, o := range theirs.Options {
			if o.Name == "Secret Top" {
				t.Error("an opponent can read the top of my library; surveil is look-at, not reveal")
			}
			if o.KnownByYou {
				t.Error("an opponent is marked a knower of a surveilled card")
			}
		}
	}
}
