package game

import (
	"testing"

	"github.com/google/uuid"
)

// madness_test.go — madness (CR 702.35), #657.
//
// Every case here drives a synthetic card through the CatalogTriggers
// / CatalogReplacements hooks, the way the rest of the keyword tests
// in this package do. The real cards (Fiery Temper, Big Game Hunter,
// and Olivia's Dragoon as the cost-discard enabler) are proved in
// cards/effects/madness_cards_test.go.
//
// What is pinned here is the keyword's own contract: the replacement
// runs for every CAUSE of discard and for nothing that is not a
// discard, the trigger offers the cast from exile, declining puts the
// card into the graveyard, the grant is priced and keyed and
// flash-timed, and an accepted-but-uncast card does not live in exile
// forever.

const madnessOracle = "test-madness-oracle"

// withMadnessCard declares madness at `cost` on one oracle key, wiring
// both halves the way buildDef does in production: the discard
// replacement and the exile-zone trigger.
func withMadnessCard(t *testing.T, cost string) {
	t.Helper()
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		madnessOracle: {MadnessReplacement()},
	})
	withZoneTriggerCard(t, madnessOracle, []TriggeredAbility{MadnessTrigger(cost)})
}

// seedMadnessCard puts a madness card in p's hand.
func seedMadnessCard(p *Player, name, typeLine, manaCost string) uuid.UUID {
	return seedHandCard(p, name, madnessOracle, typeLine, manaCost).InstanceID
}

// madnessOffer returns the queued may-cast prompt, or nil.
func madnessOffer(g *Game) *PendingChoice {
	return pendingChoiceOfKind(g, PendingChoiceMayCast)
}

// inGraveyard reports whether the card is in p's graveyard.
func inGraveyard(p *Player, id uuid.UUID) bool {
	return p.Graveyard != nil && p.Graveyard.Contains(id)
}

// --- the replacement (CR 702.35a) --------------------------------------

// The cause does not matter (Fiery Temper ruling, 2022-12-08): an
// effect's instruction and a cost pay the same road, because the
// replacement reads no cause at all.
func TestMadnessExilesTheCardWhateverCausedTheDiscard(t *testing.T) {
	for _, cause := range []DiscardCause{DiscardCauseEffect, DiscardCauseCost} {
		t.Run(string(cause), func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			withMadnessCard(t, "{R}")
			id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

			discardOne(t, g, me, id, cause)

			if inGraveyard(me, id) {
				t.Fatal("the madness card went to the graveyard; CR 702.35a exiles it instead")
			}
			c := exiledCardByIDLocked(g, id)
			if c == nil {
				t.Fatal("the madness card is not in exile")
			}
			if c.FaceDown {
				t.Error("madness exiled the card face down; CR 702.35a exiles it face up")
			}
			settleStack(t, g)
			if madnessOffer(g) == nil {
				t.Error("no madness cast was offered (CR 702.35a)")
			}
		})
	}
}

// A card that leaves the hand some OTHER way is not touched: the
// replacement watches the discard event and CR 701.8a defines a
// discard by the move out of the hand, so a plain hand → graveyard
// move is not one.
func TestMadnessIgnoresANonDiscardExitFromTheHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	g.WithWriteLock(func() {
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:   id,
			Dst:      ZoneGraveyard,
			DstOwner: me.ID,
			Actor:    me.ID,
		}); err != nil {
			t.Fatalf("routeCardToZoneLocked: %v", err)
		}
	})
	settleStack(t, g)

	if !inGraveyard(me, id) {
		t.Error("a non-discard exit from the hand was replaced by madness")
	}
	if madnessOffer(g) != nil {
		t.Error("a non-discard exit offered the madness cast")
	}
}

// A card with no madness is discarded the ordinary way and offered
// nothing — the keyword is per card, not per player.
func TestANonMadnessCardIsOfferedNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	id := seedHandCard(me, "Plain Sorcery", "test-no-madness-oracle", "Sorcery", "{1}{R}").InstanceID

	discardOne(t, g, me, id, DiscardCauseEffect)
	settleStack(t, g)

	if !inGraveyard(me, id) {
		t.Error("a card without madness did not reach the graveyard")
	}
	if madnessOffer(g) != nil {
		t.Error("a card without madness was offered a madness cast")
	}
}

// CR 701.8a: the discard HAPPENED, wherever the card ended up. Megrim,
// Marauding Mako and every other discard payoff still see it.
func TestTheDiscardEventStillFiresForAMadnessDiscard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	w := watchDiscards(g)
	discardOne(t, g, me, id, DiscardCauseEffect)

	var seen []Event
	for _, ev := range w.discards {
		if ev.CardID == id {
			seen = append(seen, ev)
		}
	}
	if len(seen) != 1 {
		t.Fatalf("EventDiscardCard fired %d times, want once", len(seen))
	}
	if seen[0].OldZone != ZoneHand || seen[0].NewZone != ZoneExile {
		t.Errorf("discard event zones: %s → %s, want hand → exile", seen[0].OldZone, seen[0].NewZone)
	}
}

// --- the offer (CR 702.35a) --------------------------------------------

// Taking the offer stamps the cast on that one exiled object, priced
// at the madness cost, keyed "madness" and flash-timed (CR 608.2g).
func TestTakingTheMadnessOfferGrantsThePricedCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	discardOne(t, g, me, id, DiscardCauseEffect)
	settleStack(t, g)
	offer := madnessOffer(g)
	if offer == nil {
		t.Fatal("no madness offer")
	}
	if offer.Chooser != me.ID {
		t.Errorf("offer chooser = %s, want the owner %s", offer.Chooser, me.ID)
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}

	perm := grantOn(g, me.ID, id, ZoneExile)
	if perm == nil {
		t.Fatal("accepting the madness offer granted no cast permission")
	}
	if perm.Cost != "{R}" {
		t.Errorf("granted price: got %q, want the madness cost {R}", perm.Cost)
	}
	if perm.AltCostKey != AltCostKeyMadness {
		t.Errorf("granted key: got %q, want %q", perm.AltCostKey, AltCostKeyMadness)
	}
	if perm.Timing != TimingFlash {
		t.Errorf("granted timing: got %q, want flash (CR 608.2g)", perm.Timing)
	}
	if !perm.CastOnly {
		t.Error("the madness grant is not cast-only; CR 702.35a says cast it")
	}

	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: AltCostKeyMadness}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Error("the madness cast did not reach the stack")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool after the cast: %d tokens left, want 0 — the madness cost was not charged", len(me.ManaPool))
	}
}

// CR 608.2g, through TimingFlash: the cast ignores the card's own
// timing. A SORCERY discarded on an opponent's turn can be cast on
// that turn.
func TestAMadnessSorceryCastsOnAnOpponentsTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	id := seedMadnessCard(me, "Test Madness Sorcery", "Sorcery", "{1}{R}{R}")
	// Walk to the opponent's turn, then discard.
	upkeepFor(t, g, 1)
	discardOne(t, g, me, id, DiscardCauseEffect)
	settleStack(t, g)

	offer := madnessOffer(g)
	if offer == nil {
		t.Fatal("no madness offer on the opponent's turn")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: AltCostKeyMadness}); err != nil {
		t.Fatalf("a madness sorcery on the opponent's turn: %v", err)
	}
}

// CR 702.35b: "if that player doesn't, they put this card into their
// graveyard" — immediately, inside the trigger's resolution.
func TestDecliningMadnessPutsTheCardInTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	discardOne(t, g, me, id, DiscardCauseEffect)
	settleStack(t, g)
	offer := madnessOffer(g)
	if offer == nil {
		t.Fatal("no madness offer")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveMayCast(decline): %v", err)
	}

	if exiledCardByIDLocked(g, id) != nil {
		t.Error("a declined madness card is still in exile; CR 702.35b puts it into the graveyard")
	}
	if !inGraveyard(me, id) {
		t.Error("the declined madness card is not in its owner's graveyard")
	}
	if perm := grantOn(g, me.ID, id, ZoneExile); perm != nil {
		t.Error("a declined offer still granted a cast permission")
	}
}

// The grant-instead-of-inline-cast simplification must not be
// STRONGER than printed: a card accepted and then never cast is in
// the graveyard by the next end step, not parked in exile forever.
func TestAnUncastMadnessCardGoesToTheGraveyard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	discardOne(t, g, me, id, DiscardCauseEffect)
	settleStack(t, g)
	offer := madnessOffer(g)
	if offer == nil {
		t.Fatal("no madness offer")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if exiledCardByIDLocked(g, id) == nil {
		t.Fatal("the accepted madness card left exile before it was cast")
	}

	advanceTo(t, g, StepEnd)
	settleStack(t, g)

	if exiledCardByIDLocked(g, id) != nil {
		t.Error("an accepted-but-uncast madness card is still in exile at the end step")
	}
	if !inGraveyard(me, id) {
		t.Error("an accepted-but-uncast madness card did not reach its owner's graveyard")
	}
}

// --- the cleanup discard (CR 514.1) ------------------------------------

// The hand-size discard is a turn-based action, not an effect, and
// madness works for it too. CR 514.3a is what makes the trigger
// playable: it gets priority in the SAME cleanup step rather than
// waiting for the next player's upkeep.
func TestAMadnessCardDiscardedToHandSize(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	fillHandTo(t, g, me, 7)
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	advanceIntoCleanup(t, g)
	if g.Turn.Step != StepCleanup || g.DiscardPending[me.ID] != 1 {
		t.Fatalf("expected a cleanup discard pause, at %s pending=%v", g.Turn.Step, g.DiscardPending)
	}
	if err := g.DiscardSelection(me.ID, []uuid.UUID{id}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}

	if exiledCardByIDLocked(g, id) == nil {
		t.Fatal("the cleanup discard did not exile the madness card")
	}
	settleStack(t, g)
	offer := madnessOffer(g)
	if offer == nil {
		t.Fatal("the cleanup discard offered no madness cast (CR 514.3a)")
	}
	if g.Turn.Step != StepCleanup {
		t.Errorf("the offer arrived in %s, want the same cleanup step (CR 514.3a)", g.Turn.Step)
	}
}

// The reclamation reaches across a turn. A cleanup discard is taken
// after this turn's end step has gone by, so the graveyard trigger
// fires at the NEXT turn's end step — and the card that was accepted
// and never cast is in a graveyard rather than parked in exile for
// the rest of the game. The trigger is guarded by CR 400.7's object
// epoch, so a gap this long costs it nothing.
func TestACleanupMadnessCardIsReclaimedAtTheNextEndStep(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	fillHandTo(t, g, me, 7)
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	advanceIntoCleanup(t, g)
	if err := g.DiscardSelection(me.ID, []uuid.UUID{id}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	settleStack(t, g)
	offer := madnessOffer(g)
	if offer == nil {
		t.Fatal("the cleanup discard offered no madness cast")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if exiledCardByIDLocked(g, id) == nil {
		t.Fatal("the accepted card left exile before it was cast")
	}

	// A whole turn later: the grant is long gone.
	advanceTo(t, g, StepEnd)
	settleStack(t, g)

	if perm := grantOn(g, me.ID, id, ZoneExile); perm != nil {
		t.Error("the madness grant outlived the turn it was made in")
	}
	if exiledCardByIDLocked(g, id) != nil {
		t.Error("the uncast madness card is stranded in exile")
	}
	if !inGraveyard(me, id) {
		t.Error("the uncast madness card never reached its owner's graveyard")
	}
}
