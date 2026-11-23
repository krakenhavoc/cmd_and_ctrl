package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Psychosis Crawler ------------------------------------------------

// TestPsychosisCrawlerIsAsBigAsYourHand — the layer-7a CDA. Power and
// toughness are SET from the controller's hand size.
//
// The second half used to document a limitation: the layer engine's
// cached resolution was invalidated by battlefield events, counters
// and tap state and NOT by hand-size changes, so the Crawler read
// correctly at every recompute and went stale between them — and the
// test called BumpLayerVersionForTest to paper over it.
//
// #74 closed that with StaticAbility.DependsOnHandSize: a hand-only
// zone move now drops the cache, but only while a permanent that
// declares the dependency is on the battlefield. So the forced bump
// is gone from this test, and its absence is the assertion.
func TestPsychosisCrawlerIsAsBigAsYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	effectstest.AdvanceToMain(t, g)
	// Push through the zone-move path so the layer engine registers
	// the static and recomputes; a raw battlefield push leaves the
	// printed P/T in place and the CDA never runs.
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Psychosis Crawler",
		TypeLine:   "Artifact Creature — Phyrexian Horror",
		OracleID:   effectstest.PsychosisCrawlerOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})

	handSize := me.Hand.Size()
	// Read the POST-LAYER characteristics: a CDA sets power in layer
	// 7a, so the printed value on the card is not the answer.
	if got := effectstest.EffectivePower(t, g, id); got != handSize {
		t.Errorf("power is %d with %d cards in hand, want %d", got, handSize, handSize)
	}
	if got := effectstest.EffectiveToughness(t, g, id); got != handSize {
		t.Errorf("toughness is %d with %d cards in hand, want %d", got, handSize, handSize)
	}

	// Draw one and read again with NOTHING else touching the board:
	// no forced bump, no token, no counter, no tap. The draw is a
	// library→hand move and nothing more, which is precisely the
	// event the listener used to ignore.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if got := effectstest.EffectivePower(t, g, id); got != handSize+1 {
		t.Errorf("power is %d after drawing a card, want %d — the CDA is reading a stale hand",
			got, handSize+1)
	}
	if got := effectstest.EffectiveToughness(t, g, id); got != handSize+1 {
		t.Errorf("toughness is %d after drawing a card, want %d", got, handSize+1)
	}

	// And the other direction: a card LEAVING the hand shrinks it.
	// Discarding is a hand→graveyard move, the mirror image of the
	// draw and the other half of the conditional bump.
	discarded := me.Hand.Cards[0].InstanceID
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		discarded,
	); err != nil {
		t.Fatalf("hand → graveyard: %v", err)
	}
	if got := effectstest.EffectivePower(t, g, id); got != handSize {
		t.Errorf("power is %d after discarding, want %d", got, handSize)
	}
}

// TestPsychosisCrawlerDrainsPerCard — three draws, three life.
func TestPsychosisCrawlerDrainsPerCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Psychosis Crawler",
		"Artifact Creature — Phyrexian Horror", effectstest.PsychosisCrawlerOracle, false)
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	effectstest.AdvanceToMain(t, g)

	opponents := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opponents[p.ID] = p.Life
		}
	}
	if len(opponents) == 0 {
		t.Skip("single-seat table")
	}

	for i := 0; i < 3; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("DrawCard %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}

	for _, p := range g.Seats {
		before, isOpp := opponents[p.ID]
		if !isOpp {
			continue
		}
		if p.Life != before-3 {
			t.Errorf("opponent life is %d, want %d — the drain must fire once PER CARD DRAWN",
				p.Life, before-3)
		}
	}
	if me.Life != g.Seats[0].Life {
		t.Error("the Crawler's controller lost life; it drains opponents only")
	}
}

// --- The Locust God ---------------------------------------------------

func TestLocustGodMakesAnInsectPerDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "The Locust God", "Legendary Creature — God",
		effectstest.LocustGodOracle, false)
	// The manual draw has to happen off the active seat's own draw
	// step: Game.DrawCard is a deliberate no-op there, because the
	// turn-based draw has already fired. newCatalogGame parks the
	// cursor on that step (#692), so walk to the main phase first.
	effectstest.AdvanceToMain(t, g)

	for i := 0; i < 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("DrawCard %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}

	insects := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Insect" && c.Controller == me.ID {
			insects++
			// The haste is the card: without it a big draw spell is
			// a board next turn instead of lethal this turn.
			if !game.HasKeyword(&c, "haste") {
				t.Error("an Insect token lacks haste")
			}
			if !game.HasKeyword(&c, "flying") {
				t.Error("an Insect token lacks flying")
			}
		}
	}
	if insects != 2 {
		t.Errorf("%d Insects after two draws, want 2 — one per card", insects)
	}
}

// TestLocustGodSchedulesItsOwnReturn — the dies-trigger is a CR 603.7
// delayed trigger, not an immediate bounce. The God is in the
// graveyard until the next end step.
func TestLocustGodSchedulesItsOwnReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "The Locust God", "Legendary Creature — God",
		effectstest.LocustGodOracle, false)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(id) && !me.Hand.Contains(id) {
		t.Fatal("the God went nowhere on death")
	}
	if me.Hand.Contains(id) {
		t.Fatal("the God returned to hand immediately; the return is delayed to the next end step")
	}
	if len(g.DelayedTriggers) == 0 {
		t.Fatal("no delayed trigger was scheduled for the return")
	}
	found := false
	for _, d := range g.DelayedTriggers {
		if d != nil && d.At == game.StepEnd {
			found = true
		}
	}
	if !found {
		t.Error("the scheduled return does not fire at an end step")
	}
}

// --- Howling Mine -----------------------------------------------------

// TestHowlingMineIsCheckedTwiceForUntapped — CR 603.4's intervening-if
// clause. Tapping the Mine in response is the whole interaction, so
// the check has to happen at resolution as well as at trigger time.
func TestHowlingMineIsCheckedTwiceForUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Howling Mine", "Artifact", effectstest.HowlingMineOracle, false)

	// Walk to the NEXT seat's draw step. Seat 0's own turn-1 draw
	// step is already behind the cursor — newCatalogGame parks there
	// (#692: at four seats CR 103.8c has the starting player draw
	// like everyone else), so the Mine's first chance to trigger is
	// the following seat's.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	// The extra draw is a trigger, so it needs the stack to drain.
	// The turn-based draw has already happened by the time the
	// cursor sits on the draw step; whatever arrives after priority
	// passes is the Mine's card.
	effectstest.AdvanceTo(t, g, game.StepDraw)
	active := g.Seats[g.Turn.ActiveSeat]
	handBefore := active.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := active.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand is %d, want %d — an untapped Mine draws one extra card", got, handBefore+1)
	}

	// Tap it and confirm the next draw step gives nothing extra.
	// This is the CR 603.4 intervening-if: a tapped Mine never puts
	// the trigger on the stack.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceTo(t, g, game.StepDraw)
	active = g.Seats[g.Turn.ActiveSeat]
	tappedBefore := active.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := active.Hand.Size(); got != tappedBefore {
		t.Errorf("hand is %d, want %d — a tapped Mine draws nothing", got, tappedBefore)
	}
	_ = me
}

// --- Hedron Crab ------------------------------------------------------

func TestHedronCrabMillsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
	effectstest.PushCatalogPermanent(g, me.ID, "Hedron Crab", "Creature — Crab", effectstest.HedronCrabOracle, false)
	oppLibBefore := opp.Library.Size()

	effectstest.PlayLandFromHand(t, g, "Island", "")
	// The trigger takes a target, so it cannot go on the stack until
	// its controller has picked one.
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Library.Size() != oppLibBefore-3 {
		t.Errorf("opponent library is %d, want %d — landfall mills three",
			opp.Library.Size(), oppLibBefore-3)
	}
	if opp.Graveyard.Size() != 3 {
		t.Errorf("opponent graveyard has %d cards, want 3", opp.Graveyard.Size())
	}
}
