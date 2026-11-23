package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_limit_test.go — #3927, CR 121.2b: "can't draw more than one card
// each turn" and "can't draw cards".

const (
	dlNarset   = "ab26fbe2-e808-48b9-8d0d-3fbb6c3d554f"
	dlSpirit   = "1463795b-ec0c-44d6-ae1a-55f78d9843ec"
	dlLeovold  = "d5d91377-fd66-4dbe-a092-07f2ea379ca7"
	dlMaralen  = "0b40032b-711d-499f-8bb9-87855f1fce9e"
	dlMornsong = "8ca2f668-2a4d-4a1e-abcc-093ccbc29295"
)

func dlNarsetOn(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushCatalogWalkerWithLine(g, owner, "Narset, Parter of Veils", "Legendary Planeswalker — Narset", dlNarset, 5)
}

func dlPut(g *game.Game, owner uuid.UUID, name, oracle, typeLine string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: oracle, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func dlDrawEvents(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventDrawCard && ev.Actor == player {
			n++
		}
	}
	return n
}

// Narset: an opponent told to draw three draws one; the other two are
// not performed — no card moves, no draw event, nothing to log. Narset's
// controller is not limited.
func TestNarsetLetsAnOpponentDrawOnlyOneCardEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dlNarsetOn(g, me.ID)

	hand, lib, events := opp.Hand.Size(), opp.Library.Size(), dlDrawEvents(g, opp.ID)
	effectstest.DrawN(g, opp, 3)
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Fatalf("the opponent drew %d cards, want 1", got)
	}
	if got := lib - opp.Library.Size(); got != 1 {
		t.Fatalf("the opponent's library lost %d cards, want 1", got)
	}
	if got := dlDrawEvents(g, opp.ID) - events; got != 1 {
		t.Fatalf("%d draw events for the opponent, want 1 (no phantom draws)", got)
	}
	effectstest.DrawN(g, opp, 1)
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Fatalf("a second instruction drew more: %d cards this turn, want 1", got)
	}

	mine := me.Hand.Size()
	effectstest.DrawN(g, me, 3)
	if got := me.Hand.Size() - mine; got != 3 {
		t.Fatalf("Narset's controller drew %d of three, want all three", got)
	}
}

// A draw made before Narset arrived counts: the opponent who drew this
// turn already can't draw again (the 2019-05-03 ruling). The next turn
// the count starts over.
func TestNarsetSeesDrawsMadeBeforeSheArrived(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DrawN(g, opp, 2)
	hand := opp.Hand.Size()
	dlNarsetOn(g, me.ID)
	effectstest.DrawN(g, opp, 1)
	if opp.Hand.Size() != hand {
		t.Fatal("an opponent who had drawn this turn drew again after Narset entered")
	}

	// The opponent's own turn: their draw-step draw is their one draw.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	hand = opp.Hand.Size()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Fatalf("the draw step drew %d, want 1 on a new turn", got)
	}
	effectstest.DrawN(g, opp, 1)
	if got := opp.Hand.Size() - hand; got != 1 {
		t.Fatalf("a second draw on the opponent's own turn went through (%d drawn)", got)
	}
}

// A draw Narset forbids is not attempted: an empty library loses no one
// (no CR 704.5b), and the "then" of the instruction still happens.
func TestNarsetForbiddenDrawFromAnEmptyLibraryLosesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dlNarsetOn(g, me.ID)
	effectstest.DrawN(g, opp, 1)
	trimLibrary(t, opp, 0)
	effectstest.StartLoot(g, opp, 2, 1)
	if opp.AttemptedEmptyDraw {
		t.Fatal("a draw Narset forbids drew from the empty library")
	}
	if effectstest.DiscardPromptFor(g, opp) == nil {
		t.Fatal("the discard after \"then\" did not happen")
	}
}

// A draw Narset forbids can't be replaced: dredge is not offered. But a
// first draw that dredge replaced was not a draw, so the next one is
// allowed (the 2019-05-03 ruling).
func TestNarsetAndDredge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dlNarsetOn(g, me.ID)
	effectstest.GyCardWithOracle(opp, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)

	// First draw: dredge replaces it.
	effectstest.DrawN(g, opp, 1)
	effectstest.AnswerOptional(t, g, opp, true)
	if got := len(g.CardsDrawnThisTurnFor(opp.ID)); got != 0 {
		t.Fatalf("a dredged draw counted as %d draws", got)
	}
	// Second draw: allowed, because nothing was drawn yet. Decline the
	// dredge and draw.
	effectstest.GyCardWithOracle(opp, "Stinkweed Imp", "Creature — Imp", effectstest.StinkweedImpOracle)
	hand := opp.Hand.Size()
	effectstest.DrawN(g, opp, 1)
	effectstest.AnswerOptional(t, g, opp, false)
	if opp.Hand.Size() != hand+1 {
		t.Fatal("the first real draw of the turn was refused")
	}
	// Third draw: forbidden, and so not offered to dredge.
	effectstest.DrawN(g, opp, 1)
	if effectstest.OptionalOffer(g, opp) != nil {
		t.Fatal("dredge was offered on a draw Narset forbids")
	}
	if opp.Hand.Size() != hand+1 {
		t.Fatal("a second real draw went through")
	}
}

// The sandbox draw is refused out loud.
func TestManualDrawIsRefusedUnderTheLimit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dlNarsetOn(g, me.ID)
	effectstest.DrawN(g, opp, 1)
	if err := g.DrawCard(opp.ID); !errors.Is(err, game.ErrDrawRestricted) {
		t.Fatalf("DrawCard = %v, want ErrDrawRestricted", err)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("Narset's controller's manual draw: %v", err)
	}
}

// Narset's −2: four cards looked at, only a noncreature, nonland card
// may be taken, the rest go to the bottom.
func TestNarsetMinusTwoTakesANoncreatureNonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	narset := dlNarsetOn(g, me.ID)
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Opt", TypeLine: "Instant"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
	)
	if err := g.ActivateCatalogAbility(me.ID, narset, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate −2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no −2 prompt")
	}
	want := map[uuid.UUID]bool{ids[2]: true, ids[3]: true}
	if len(prompt.ChooseCards) != len(want) {
		t.Fatalf("offered %d cards, want the 2 noncreature, nonland cards", len(prompt.ChooseCards))
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered a creature or land card: %s", id)
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("take Opt: %v", err)
	}
	if !effectstest.ZoneHas(me.Hand, ids[2]) {
		t.Fatal("Opt did not reach the hand")
	}
	if c := effectstest.GaLayered(t, g, narset); c.Counters[game.CounterLoyalty] != 3 {
		t.Fatalf("loyalty %d, want 3", c.Counters[game.CounterLoyalty])
	}
}

// Spirit of the Labyrinth limits every player, its controller too. The
// active player (seat 0) has made their draw-step draw already, so they
// draw nothing more; seats 1 and 2 (the Spirit's controller) draw one.
func TestSpiritOfTheLabyrinthLimitsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active, me, other := g.Seats[0], g.Seats[1], g.Seats[2]
	dlPut(g, me.ID, "Spirit of the Labyrinth", dlSpirit, "Enchantment Creature — Spirit")
	for _, tc := range []struct {
		name string
		p    *game.Player
	}{{"the active player", active}, {"the Spirit's controller", me}, {"another player", other}} {
		want := 1 - len(g.CardsDrawnThisTurnFor(tc.p.ID))
		if want < 0 {
			want = 0
		}
		hand := tc.p.Hand.Size()
		effectstest.DrawN(g, tc.p, 2)
		if got := tc.p.Hand.Size() - hand; got != want {
			t.Errorf("%s drew %d of two, want %d", tc.name, got, want)
		}
	}
}

// Leovold: the static is Narset's; an opponent's spell targeting Leovold's
// controller offers a draw.
func TestLeovoldLimitsOpponentsAndDrawsWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	leovold := dlPut(g, me.ID, "Leovold, Emissary of Trest", dlLeovold, "Legendary Creature — Elf Advisor")

	other := g.Seats[2]
	hand := other.Hand.Size()
	effectstest.DrawN(g, other, 2)
	if got := other.Hand.Size() - hand; got != 1 {
		t.Fatalf("Leovold's opponent drew %d of two, want 1", got)
	}

	mine := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	prompt := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if prompt.Source != leovold {
		t.Fatalf("the trigger prompt is from %v, want Leovold", prompt.Source)
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("answer Leovold: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - mine; got != 1 {
		t.Fatalf("Leovold's controller drew %d when targeted by an opponent, want 1", got)
	}

	// Leovold's controller targeting themselves does not trigger it.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	l2 := dlPut(g2, me2.ID, "Leovold, Emissary of Trest", dlLeovold, "Legendary Creature — Elf Advisor")
	effectstest.CastCatalogSpell(t, g2, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me2.ID}})
	if effectstest.TriggerOnStack(g2, l2) != nil || effectstest.LatestTriggerPrompt(g2, me2.ID) != nil {
		t.Fatal("Leovold triggered on its controller's own spell")
	}
}

// Maralen of the Mornsong: nobody draws, the draw step's draw included;
// instead the active player loses 3 life and searches for a card.
func TestMaralenOfTheMornsongReplacesTheDrawWithATutor(t *testing.T) {
	for _, tc := range []struct{ name, oracle, typeLine string }{
		{"Maralen of the Mornsong", dlMaralen, "Legendary Creature — Elf Wizard"},
		{"Mornsong Aria", dlMornsong, "Legendary Enchantment"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dlPut(g, me.ID, tc.name, tc.oracle, tc.typeLine)

		mine := me.Hand.Size()
		effectstest.DrawN(g, me, 1)
		if me.Hand.Size() != mine {
			t.Fatalf("%s: its controller drew a card", tc.name)
		}

		effectstest.AdvanceToUpkeepOf(t, g, 1)
		hand, life := opp.Hand.Size(), opp.Life
		for g.Turn.Step != game.StepDraw {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("%s: advance: %v", tc.name, err)
			}
		}
		if opp.Hand.Size() != hand {
			t.Fatalf("%s: the draw step drew a card", tc.name)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := life - opp.Life; got != 3 {
			t.Fatalf("%s: the active player lost %d life, want 3", tc.name, got)
		}
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, opp.ID) == nil {
			t.Fatalf("%s: the active player was not asked to search", tc.name)
		}
	}
}

// Mornsong Aria also stops life gain.
func TestMornsongAriaStopsLifeGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dlPut(g, me.ID, "Mornsong Aria", dlMornsong, "Legendary Enchantment")
	var cant bool
	g.ReadSnapshot(func() { cant = g.PlayerCantGainLifeLocked(me) })
	if !cant {
		t.Fatal("Mornsong Aria's controller can gain life")
	}
}
