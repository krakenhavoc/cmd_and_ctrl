package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shocklands_test.go — "as this land enters, you may pay 2 life. If
// you don't, it enters tapped."
//
// The assertion that earns its keep in every one of these is the tap
// EVENT count, not the Tapped flag. The implementation this replaced
// entered tapped and untapped itself for 2 life, which produces the
// same final Tapped state in both branches and is therefore invisible
// to a state assertion. Events are not fooled:
//
//	replacement      → 0 tap events, 0 untap events, ever
//	ETB-tap patch-up → 1 tap event (or 1 untap on the pay branch)
//
// The prompt itself is the second half of the proof: nothing is on
// the battlefield while it is open, because the entry has not
// happened yet.

const (
	bloodCryptOracle       = "43985bbc-a0f6-4812-984e-392bc8562633"
	hallowedFountainOracle = "f1750962-a87c-49f6-b731-02ae971ac6ea"
	steamVentsOracle       = "17039058-822d-409f-938c-b727a366ba63"
)

// shocklandOracleIDs is the whole cycle, name → oracle ID, verified
// against Scryfall's default-cards dump.
var shocklandOracleIDs = map[string]string{
	"Blood Crypt":       bloodCryptOracle,
	"Breeding Pool":     "20283c4a-f1f0-42f0-bc08-6da87474426b",
	"Godless Shrine":    "73864fcc-1bde-4bc0-831e-2b93e546e417",
	"Hallowed Fountain": hallowedFountainOracle,
	"Overgrown Tomb":    "975ec9a3-6f20-4177-8211-82526e092538",
	"Sacred Foundry":    "45181cb8-2090-4471-ba90-e5a8f04d525f",
	"Steam Vents":       steamVentsOracle,
	"Stomping Ground":   "16052b52-ade1-406f-a06b-ce7ea607fb63",
	"Temple Garden":     "f413a83d-a40d-434c-b20a-4c707c0527fa",
	"Watery Grave":      "fc9ec820-4245-4a96-b009-5308a818ca58",
}

// entryPayLifeChoiceFor returns the open "you may pay N life" prompt
// addressed to chooser, or nil.
func entryPayLifeChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceEntryPayLife && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// answerEntryPayLife answers the open entry prompt for chooser.
func answerEntryPayLife(t *testing.T, g *game.Game, chooser uuid.UUID, pay bool) {
	t.Helper()
	c := entryPayLifeChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no PendingChoiceEntryPayLife addressed to %s", chooser)
	}
	if err := g.ResolveEntryPayLife(c.ID, chooser, pay); err != nil {
		t.Fatalf("ResolveEntryPayLife: %v", err)
	}
}

// untapEventsFor counts EventUntapCard entries for one card — the
// tell of the enter-tapped-then-untap implementation this replaced.
func untapEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventUntapCard && ev.CardID == id {
			n++
		}
	}
	return n
}

func TestShocklandPromptPausesTheEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	id := playLandFromHand(t, g, "Steam Vents", steamVentsOracle)

	c := entryPayLifeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("playing a shockland queued no pay-life prompt")
	}
	if c.PayCost != "2 life" {
		t.Errorf("prompt cost label: got %q, want %q", c.PayCost, "2 life")
	}
	if c.Source != id {
		t.Errorf("prompt source: got %s, want the entering land %s", c.Source, id)
	}
	// The entry is what's being replaced, so nothing has entered.
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("the land reached the battlefield before the choice was made")
	}
	if !me.Hand.Contains(id) {
		t.Error("the land left the hand while the entry prompt was still open")
	}

	answerEntryPayLife(t, g, me.ID, false)
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("answering the prompt did not resume the entry")
	}
}

func TestShocklandPaidTwoLifeEntersUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lifeBefore := me.Life

	id := playLandFromHand(t, g, "Hallowed Fountain", hallowedFountainOracle)
	answerEntryPayLife(t, g, me.ID, true)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Hallowed Fountain is not on the battlefield")
	}
	if card.Tapped {
		t.Error("paid 2 life and the land still entered tapped")
	}
	if got := lifeBefore - me.Life; got != 2 {
		t.Errorf("life paid: got %d, want 2", got)
	}
	// The pay branch of the old implementation entered tapped and
	// then untapped: one tap event and one untap event. A replaced
	// entry emits neither.
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; a paid shockland is never tapped at all", n)
	}
	if n := untapEventsFor(g, id); n != 0 {
		t.Errorf("%d untap events; the land entered untapped, it was not untapped", n)
	}
}

func TestShocklandDeclinedEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lifeBefore := me.Life

	id := playLandFromHand(t, g, "Blood Crypt", bloodCryptOracle)
	answerEntryPayLife(t, g, me.ID, false)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Blood Crypt is not on the battlefield")
	}
	if !card.Tapped {
		t.Error("declined the payment and the land entered untapped anyway")
	}
	// The discriminator: the land ENTERED tapped, so nothing tapped
	// it. An OnETB tap would show up here as one event.
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
	if me.Life != lifeBefore {
		t.Errorf("life changed on a decline: got %d, want %d", me.Life, lifeBefore)
	}
}

// TestShocklandBelowTheCostIsNotPrompted pins CR 118.4: a player may
// pay 2 life only with a life total of at least 2. At 1 life there is
// no decision to offer, so the land just enters tapped — and it must
// not cost the player their last life either.
func TestShocklandBelowTheCostIsNotPrompted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Life = 1

	id := playLandFromHand(t, g, "Godless Shrine", shocklandOracleIDs["Godless Shrine"])

	if c := entryPayLifeChoiceFor(g, me.ID); c != nil {
		t.Error("prompted for a payment the player cannot legally make")
	}
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the land did not enter when the prompt was skipped")
	}
	if !card.Tapped {
		t.Error("unpaid shockland entered untapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
	if me.Life != 1 {
		t.Errorf("life: got %d, want 1 — an unaffordable cost must not be charged", me.Life)
	}
}

// TestShocklandExactlyTheCostMayPay is the other side of CR 118.4:
// at exactly 2 life the payment is legal, and taking it to 0 is the
// player's business (state-based actions handle the rest).
func TestShocklandExactlyTheCostMayPay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Life = 2

	id := playLandFromHand(t, g, "Watery Grave", shocklandOracleIDs["Watery Grave"])
	if c := entryPayLifeChoiceFor(g, me.ID); c == nil {
		t.Fatal("no prompt at exactly the cost — paying to 0 is legal")
	}
	answerEntryPayLife(t, g, me.ID, true)

	if me.Life != 0 {
		t.Errorf("life: got %d, want 0", me.Life)
	}
	// "State-based actions handle the rest" is the whole tail of this
	// test now: at 0 life the player loses, and CR 800.4a (#769) takes
	// their objects out of the game with them — this land included, at
	// a four-seat table where the game goes on. That a PAID shockland
	// enters untapped is TestShocklandPaidTwoLifeEntersUntapped's
	// claim; this test is about the payment being legal at exactly the
	// cost.
	if !me.Eliminated {
		t.Error("a player at 0 life should have lost to the state-based action")
	}
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("the land stayed on the battlefield after its owner left the game (CR 800.4a)")
	}
}

// TestShocklandNoStackTrip pins the other half of the superseded
// simplification: the choice is not a triggered ability, so it never
// touches the stack and never hands an opponent priority.
func TestShocklandNoStackTrip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	id := playLandFromHand(t, g, "Temple Garden", shocklandOracleIDs["Temple Garden"])
	answerEntryPayLife(t, g, me.ID, true)

	if n := g.Stack.Size(); n != 0 {
		t.Errorf("stack holds %d items; the entry choice is a replacement, not a trigger", n)
	}
	if len(g.PendingTriggers) != 0 {
		t.Errorf("%d pending triggers; the entry choice is not a trigger", len(g.PendingTriggers))
	}
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("Temple Garden is not on the battlefield")
	}
}

// TestShocklandUnderKismetStillOffersThePayment covers the CR 616
// case: an opponent's Kismet replaces the same entry, so the
// affected player orders the two replacements first. The payment
// decision has to survive that — an effect with a choice inside it
// must not be fired blind just because it arrived through the
// ordering prompt.
//
// Paying under a Kismet is a bad deal (the land enters tapped
// either way), which is exactly why the player, not the engine,
// should be the one to decline it.
func TestShocklandUnderKismetStillOffersThePayment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opponent := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lifeBefore := me.Life
	seedReplacementPermanent(g, "81fdd1c4-d43b-4f8b-8712-7c2bf45a3e0b", "Kismet", opponent.ID)

	id := playLandFromHand(t, g, "Sacred Foundry", shocklandOracleIDs["Sacred Foundry"])

	var order *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			order = c
		}
	}
	if order == nil {
		t.Fatal("two enters-tapped replacements applied but no CR 616 ordering prompt")
	}
	if order.Chooser != me.ID {
		t.Errorf("ordering chooser = %s, want the entering land's controller %s", order.Chooser, me.ID)
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, order.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}

	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the ordering prompt swallowed the pay-life decision")
	}
	answerEntryPayLife(t, g, me.ID, false)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Sacred Foundry never entered")
	}
	if !card.Tapped {
		t.Error("entered untapped with a Kismet out and no payment")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; both replacements act on the ENTRY", n)
	}
	if me.Life != lifeBefore {
		t.Errorf("life changed on a decline: got %d, want %d", me.Life, lifeBefore)
	}
}

// TestShocklandOnAnUnresumableEntryEntersTapped covers an entry site
// that runs the replacement pipeline but has no resume for a paused
// prompt (a direct move, as opposed to playing the land). Prompting
// there would strand the card in its old zone, so the engine takes
// the branch the player didn't pay for: the land enters tapped, for
// free, and the game keeps moving. Weaker than printed, never
// stronger.
func TestShocklandOnAnUnresumableEntryEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lifeBefore := me.Life

	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Overgrown Tomb",
		TypeLine:   "Land",
		OracleID:   shocklandOracleIDs["Overgrown Tomb"],
		Owner:      me.ID,
		Controller: me.ID,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		id,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}

	if c := entryPayLifeChoiceFor(g, me.ID); c != nil {
		t.Error("prompted on an entry the engine cannot resume")
	}
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the land was stranded outside the battlefield")
	}
	if !card.Tapped {
		t.Error("entered untapped without a payment")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; it should have ENTERED tapped", n)
	}
	if me.Life != lifeBefore {
		t.Errorf("life changed without a prompt: got %d, want %d", me.Life, lifeBefore)
	}
}

// TestShocklandCycleRegistered walks all ten: each is registered
// under the oracle ID Scryfall reports, charges 2 life on entry, and
// carries its two-colour mana ability.
func TestShocklandCycleRegistered(t *testing.T) {
	for name, oracleID := range shocklandOracleIDs {
		spec, ok := Lookup(oracleID)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracleID)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered under the name %q", oracleID, spec.Name)
		}
		if len(spec.Replacements) != 1 {
			t.Errorf("%s: %d replacements, want 1", name, len(spec.Replacements))
			continue
		}
		if got := spec.Replacements[0].EntryLifeCost; got != 2 {
			t.Errorf("%s: entry life cost %d, want 2", name, got)
		}
		if len(spec.ManaAbilities) != 1 {
			t.Errorf("%s: %d mana abilities, want 1", name, len(spec.ManaAbilities))
		}
	}
}

// TestShocklandEveryMemberEntersTappedOnDecline plays each of the ten
// through the real path — a typo in one oracle ID would otherwise
// only show up as a card that silently does nothing.
func TestShocklandEveryMemberEntersTappedOnDecline(t *testing.T) {
	for name, oracleID := range shocklandOracleIDs {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]

			id := playLandFromHand(t, g, name, oracleID)
			answerEntryPayLife(t, g, me.ID, false)

			card, ok := battlefieldCard(g, id)
			if !ok {
				t.Fatalf("%s is not on the battlefield", name)
			}
			if !card.Tapped {
				t.Errorf("%s entered untapped without paying", name)
			}
			if n := tapEventsFor(g, id); n != 0 {
				t.Errorf("%s: %d tap events; it should have ENTERED tapped", name, n)
			}
		})
	}
}

// TestWhichShocklandEntrySitesOfferThePayment is the CAVEAT, held
// against the engine (#1051).
//
// The caveat used to read "put onto the battlefield by another spell
// always enters tapped", and that stopped being true at #478: the
// effect-side entries that go through
// enterBattlefieldThroughPipelineLocked are entryResumable, so the
// biggest "another spell" there is — a SEARCH, which is every
// fetchland, Farseek, Circuitous Route and Expedition Map — pauses the
// entry and asks. What survives is the entry site that runs the CR 614
// pipeline with NO resume behind it:
// putOntoBattlefieldFromZoneLocked, the hand / library "put onto the
// battlefield" batch (server/internal/game/battlefield_put.go:94),
// whose simultaneity a per-card resume would break. There the engine
// takes the un-paid branch — weaker than printed, never stronger.
//
// One table over the entry SITES rather than a test per card, because
// what the caveat claims is a DIFFERENCE: either half asserted alone
// would keep passing if the other moved, and it is the boundary
// between them that the caveat's words have to track.
//
// The search row is deliberately thin. What it pins is that this
// family is asked at all; the search's own side of the pause —
// nothing moves while the question is open, EventSearchLibrary and
// the shuffle wait for the answer — is
// TestFetchedShocklandOffersItsPaymentAndTheSearchWaits
// (search_chooser_test.go), and is not copied here.
func TestWhichShocklandEntrySitesOfferThePayment(t *testing.T) {
	for _, tc := range []struct {
		site       string
		wantPrompt bool
		enter      func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID
	}{
		{
			// A fetchland cracking for it, Farseek, a Circuitous
			// Route: the search path, entryResumable since #478.
			site:       "fetched by a search",
			wantPrompt: true,
			enter: func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
				t.Helper()
				id := pushLibraryCardForTest(me, game.Card{
					Name:     "Steam Vents",
					TypeLine: "Land — Island Mountain",
					OracleID: steamVentsOracle,
				})
				g.WithWriteLock(func() {
					if err := g.SearchLibraryForEffectWithOptions(me.ID,
						func(c game.Card) bool { return c.Name == "Steam Vents" },
						game.ZoneBattlefield, 1, false, false, false); err != nil {
						t.Fatalf("SearchLibraryForEffectWithOptions: %v", err)
					}
				})
				return id
			},
		},
		{
			// Warp World, Genesis Wave, Coiling Oracle: the library
			// half of the batch that cannot resume.
			site:       "put from the library",
			wantPrompt: false,
			enter: func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
				t.Helper()
				id := pushLibraryCardForTest(me, game.Card{
					Name:     "Blood Crypt",
					TypeLine: "Land — Swamp Mountain",
					OracleID: bloodCryptOracle,
				})
				g.WithWriteLock(func() {
					if _, err := g.PutCardsFromLibraryOntoBattlefieldForEffect(
						[]uuid.UUID{id}, game.LibraryEntryOptions{Controller: me.ID},
					); err != nil {
						t.Fatalf("PutCardsFromLibraryOntoBattlefieldForEffect: %v", err)
					}
				})
				return id
			},
		},
		{
			// Arboreal Grazer: the hand half of the same batch, and
			// not a land drop.
			site:       "put from the hand",
			wantPrompt: false,
			enter: func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
				t.Helper()
				id := uuid.New()
				me.Hand.PushTop(game.Card{
					InstanceID: id,
					Name:       "Hallowed Fountain",
					TypeLine:   "Land — Plains Island",
					OracleID:   hallowedFountainOracle,
					Owner:      me.ID,
					Controller: me.ID,
				})
				g.WithWriteLock(func() {
					if _, err := g.PutFromHandOntoBattlefieldForEffect(
						id, game.HandEntryOptions{Controller: me.ID},
					); err != nil {
						t.Fatalf("PutFromHandOntoBattlefieldForEffect: %v", err)
					}
				})
				return id
			},
		},
	} {
		t.Run(tc.site, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			lifeBefore := me.Life

			id := tc.enter(t, g, me)

			prompt := entryPayLifeChoiceFor(g, me.ID)
			switch {
			case tc.wantPrompt && prompt == nil:
				t.Fatal("no pay-life prompt: this entry site is entryResumable, so the " +
					"land's own replacement is allowed to ask")
			case !tc.wantPrompt && prompt != nil:
				t.Fatal("prompted on an entry the engine cannot resume; pausing there would " +
					"strand the card in its old zone")
			}

			wantLife, wantTapped := lifeBefore, true
			if tc.wantPrompt {
				if prompt.PayCost != "2 life" {
					t.Errorf("prompt cost label: got %q, want %q", prompt.PayCost, "2 life")
				}
				// The ENTRY is what is being replaced, so nothing has
				// entered while the question is open.
				if _, ok := battlefieldCard(g, id); ok {
					t.Error("the land entered before the choice was made")
				}
				answerEntryPayLife(t, g, me.ID, true)
				wantLife, wantTapped = lifeBefore-2, false
			}

			card, ok := battlefieldCard(g, id)
			if !ok {
				t.Fatal("the land is not on the battlefield")
			}
			if card.Tapped != wantTapped {
				t.Errorf("tapped = %v, want %v", card.Tapped, wantTapped)
			}
			if me.Life != wantLife {
				t.Errorf("life: got %d, want %d", me.Life, wantLife)
			}
			// The whole cycle's discriminator, at every site: the
			// entry is what the replacement acts on, so nothing is
			// ever tapped or untapped after the fact.
			if n := tapEventsFor(g, id); n != 0 {
				t.Errorf("%d tap events; the land should have ENTERED as it is", n)
			}
			if n := untapEventsFor(g, id); n != 0 {
				t.Errorf("%d untap events; there was no tapped window to undo", n)
			}
		})
	}
}
