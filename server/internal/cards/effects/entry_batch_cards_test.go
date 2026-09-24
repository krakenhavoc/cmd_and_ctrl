package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_batch_cards_test.go — the proof cards for #1322, #1324 and
// #1327 (tracker #1306, "Aang is so flashy"): the entry paths a
// shockland, a pay-life MDFC land back, Sword of Hearth and Home and
// Phelia, Exuberant Shepherd each waited on. The engine half is
// server/internal/game/entry_batch_test.go and
// exile_return_then_test.go.

const (
	ebSwordOfHearthAndHomeOracle = "913e6182-706a-4872-8c8a-e146b0ae0738"
	ebPheliaOracle               = "5d86a59a-ba1f-45f7-b829-dca6f9f3e624"
	ebLoyalWarhoundOracle        = "cc6a83c7-e645-4a53-9550-be79b42cd851"
	ebSinkIntoStuporOracle       = "bcc6eece-75ea-494c-b33a-d4477d504e0b"
)

// ebLandsOf is the lands `controller` controls on the battlefield.
func ebLandsOf(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsLand() && c.Controller == controller {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- #1322: Hallowed Fountain put by a spell ---------------------------

// TestPutHallowedFountainAsksAndTheRestWaits is Genesis Wave's shape:
// "put any number of permanent cards from among them onto the
// battlefield. Put the rest into your graveyard." The Fountain's pay-2
// question is ASKED (it used to take the un-asked branch and enter
// tapped), nothing moves while it is open — not the Fountain, not the
// Forest put beside it, and not "the rest" either, which is the card
// side's half of the fix: the rest is computed once the entry is
// complete.
func TestPutHallowedFountainAsksAndTheRestWaits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lifeBefore := me.Life
	chaff := plTop(me, "Opt", "Instant", "{U}")
	forest := plTop(me, "Forest", "Basic Land — Forest", "")
	fountain := pushLibraryCardForTest(me, game.Card{
		Name: "Hallowed Fountain", TypeLine: "Land — Plains Island", OracleID: hallowedFountainOracle,
	})

	var res PutFromLibraryResult
	thenRan := 0
	g.WithWriteLock(func() {
		err := PutFromLibraryOntoBattlefield{
			Player: me.ID,
			Cards:  []uuid.UUID{fountain, forest, chaff},
			All:    true,
			Then: func(g *game.Game, r PutFromLibraryResult) error {
				thenRan++
				res = r
				return PutRestIntoGraveyard(g, r)
			},
		}.Apply(NewContext(g, nil))
		if err != nil {
			t.Fatalf("put: %v", err)
		}
	})

	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("Hallowed Fountain put onto the battlefield by a spell was not offered its payment")
	}
	if g.Battlefield.Contains(fountain) || g.Battlefield.Contains(forest) {
		t.Fatal("a card of the batch entered while the question was open")
	}
	if thenRan != 0 || me.Graveyard.Contains(chaff) {
		t.Fatal("the rest was put away while a card of the batch was still waiting to enter")
	}

	answerEntryPayLife(t, g, me.ID, true)

	card, ok := battlefieldCard(g, fountain)
	if !ok || card.Tapped {
		t.Fatalf("the paid Fountain should be on the battlefield untapped (ok=%v)", ok)
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("the Forest enters with it")
	}
	if me.Life != lifeBefore-2 {
		t.Errorf("life = %d, want %d", me.Life, lifeBefore-2)
	}
	if thenRan != 1 || len(res.Entered) != 2 {
		t.Fatalf("Then ran %d times with %v entered; want once with both", thenRan, res.Entered)
	}
	if len(res.Rest) != 1 || res.Rest[0] != chaff || !me.Graveyard.Contains(chaff) {
		t.Errorf("rest %v; want the Opt, in the graveyard", res.Rest)
	}
	if n := tapEventsFor(g, fountain); n != 0 {
		t.Errorf("%d tap events on a paid shockland", n)
	}
}

// --- #1322: the pay-life MDFC land backs -------------------------------

// TestPayLifeMDFCBacksAreCompleteAndStillAsk holds Sea Gate, Reborn and
// Soporific Springs — the two back faces the deck plays — to the
// catalog and to the engine.
//
// Their caveat said a spell PUTTING the back face onto the battlefield
// skipped the 3 life. For these two that case cannot arise at all: a
// double-faced card put onto the battlefield from anywhere but the
// stack enters FRONT face up (CR 712.14), and outside the battlefield
// and the stack it has only its front face's characteristics
// (CR 712.8a) — a sorcery and an instant, which are not permanent
// cards (CR 110.4). So the back face enters by being PLAYED, and
// playing it asks. With the put batch resumable the caveat is untrue
// for every back face in the cycle, and is gone.
func TestPayLifeMDFCBacksAreCompleteAndStillAsk(t *testing.T) {
	for _, tc := range []struct {
		front, frontType, back string
		oracle                 string
	}{
		{"Sea Gate Restoration", "Sorcery", "Sea Gate, Reborn", seaGateRestorationOracle},
		{"Sink into Stupor", "Instant", "Soporific Springs", ebSinkIntoStuporOracle},
	} {
		t.Run(tc.back, func(t *testing.T) {
			spec, ok := Lookup(game.CatalogKeyForFace(tc.oracle, 1))
			if !ok || spec.Name != tc.back {
				t.Fatalf("%s is not registered as the back face", tc.back)
			}
			if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s: completeness %q with caveats %v", tc.back, spec.Completeness, spec.Caveats)
			}
			build := func(owner uuid.UUID) game.Card {
				return mdfcCard(owner, tc.oracle, game.LayoutModalDFC,
					game.Face{Name: tc.front, TypeLine: tc.frontType, ManaCost: "{1}{U}{U}", Colors: []string{"U"}},
					game.Face{Name: tc.back, TypeLine: "Land"},
				)
			}

			// Played as a land: asked, paid, untapped.
			g, me, id := handWithMDFC(t, build)
			lifeBefore := me.Life
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
				t.Fatalf("play %s: %v", tc.back, err)
			}
			if c := entryPayLifeChoiceFor(g, me.ID); c == nil || c.PayCost != "3 life" {
				t.Fatalf("playing %s did not ask for 3 life", tc.back)
			}
			answerEntryPayLife(t, g, me.ID, true)
			if card, ok := battlefieldCard(g, id); !ok || card.Tapped || card.Name != tc.back {
				t.Fatalf("%s should be on the battlefield untapped", tc.back)
			}
			if me.Life != lifeBefore-3 {
				t.Errorf("life = %d, want %d", me.Life, lifeBefore-3)
			}

			// Put by a spell: CR 712.14 / 110.4 — the front face is not a
			// permanent card, so the batch refuses it and the card stays
			// where it is.
			other := build(me.ID)
			me.Library.PushTop(other)
			g.WithWriteLock(func() {
				if _, err := g.PutCardsFromLibraryOntoBattlefieldForEffect(
					[]uuid.UUID{other.InstanceID}, game.LibraryEntryOptions{Controller: me.ID}); err == nil {
					t.Error("a spell put a double-faced card whose front is not a permanent onto the battlefield")
				}
			})
			if !me.Library.Contains(other.InstanceID) || entryPayLifeChoiceFor(g, me.ID) != nil {
				t.Fatal("the refused put moved the card or asked a question")
			}
		})
	}
}

// --- #1322: a manifested shockland is a 2/2 and asks nothing -----------

// TestManifestedShocklandAsksNothing: a face-down permanent has no
// abilities (CR 708.2a), and CR 614.12 decides which entry replacements
// apply from the permanent as it would exist on the battlefield — so
// the shockland's own "pay 2 life" does not apply to a manifest of it.
// Before #1322 the question could not be asked on the put batch, so the
// manifest entered TAPPED instead; with the batch resumable it would
// have asked for life on a card the table cannot see. Neither: it
// enters untapped, face down, with no prompt.
func TestManifestedShocklandAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lifeBefore := me.Life
	// On TOP: manifest takes the top card, and pushLibraryCardForTest
	// puts its card on the bottom.
	fountain := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: fountain, Name: "Hallowed Fountain", TypeLine: "Land — Plains Island",
		OracleID: hallowedFountainOracle, Owner: me.ID, Controller: me.ID,
	})
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		if id, err = g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("ManifestForEffect: %v", err)
		}
	})
	if entryPayLifeChoiceFor(g, me.ID) != nil {
		t.Fatal("a manifested shockland asked for its life: the face-down permanent has no abilities")
	}
	if id != fountain {
		t.Fatalf("manifested %s, want the Hallowed Fountain on top", id)
	}
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the manifest did not enter")
	}
	if card.Tapped {
		t.Error("the manifest entered tapped: the shockland's own replacement applied to a face-down 2/2")
	}
	if me.Life != lifeBefore {
		t.Errorf("life = %d, want %d", me.Life, lifeBefore)
	}
}

// --- #1324: Sword of Hearth and Home -----------------------------------

// ebSwordSetup equips the Sword to a bear under seat 0, blinks-to-be
// Loyal Warhound beside it, and gives the opponent exactly one more land
// than seat 0 — so Loyal Warhound's intervening-if ("if an opponent
// controls more lands than you") is TRUE before the Sword's basic land
// arrives and FALSE after.
func ebSwordSetup(t *testing.T) (g *game.Game, me, opp *game.Player, bear, warhound, plains, spare uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear = seedBear(g, me.ID)
	sword := seedEquipment(g, me.ID, "Sword of Hearth and Home", ebSwordOfHearthAndHomeOracle)
	equipTo(t, g, me.ID, sword, bear)
	warhound = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Loyal Warhound", TypeLine: "Creature — Dog",
		OracleID: ebLoyalWarhoundOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 2; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
			Owner: me.ID, Controller: me.ID,
		})
	}
	for i := 0; i < 3; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
			Owner: opp.ID, Controller: opp.ID,
		})
	}
	plains = plTop(me, "Plains", "Basic Land — Plains", "")
	plTop(me, "Opt", "Instant", "{U}")
	spare = plTop(me, "Plains", "Basic Land — Plains", "")
	return g, me, opp, bear, warhound, plains, spare
}

// TestSwordOfHearthAndHomePutsBothCardsAsOneEntry is #1324's proof. The
// Warhound comes back from exile and a basic Plains comes out of the
// library as ONE event, so when the Warhound's ETB is checked the
// Plains is already on the battlefield (CR 603.6a): seat 0 now has as
// many lands as the opponent, the intervening-if is false (CR 603.4),
// and the Warhound does not trigger. Two entries with the creature
// first would trigger it and fetch the spare Plains — stronger than
// printed.
func TestSwordOfHearthAndHomePutsBothCardsAsOneEntry(t *testing.T) {
	g, me, opp, bear, warhound, _, spare := ebSwordSetup(t)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Fatalf("equipped power %d, want 4", got)
	}

	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g) // the trigger announces and asks for its target
	pickCard(t, g, me.ID, warhound)
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchByID(t, g, me.ID, c.SearchCards[0])
		passPriorityAroundTable(t, g)
	}

	hounds := battlefieldIDsNamed(g, "Loyal Warhound")
	if len(hounds) != 1 || hounds[0] == warhound {
		t.Fatalf("Loyal Warhound should be back as a new object; battlefield has %v", hounds)
	}
	if exileHas(g, warhound) {
		t.Error("the Warhound is still in exile")
	}
	if n := len(ebLandsOf(g, me.ID)); n != 3 {
		t.Errorf("seat 0 controls %d lands, want 3 — the two it had and the one the Sword put", n)
	}
	if g.Battlefield.Contains(spare) {
		t.Error("Loyal Warhound triggered and fetched a Plains: its ETB saw the board before the Sword's land " +
			"arrived, which is two entries rather than one (CR 603.6a)")
	}
	// The sharper half: the ability must not TRIGGER at all. Checked on
	// the event log rather than the stack, because the intervening-if
	// is checked again on resolution (CR 603.4) — a Warhound announced
	// before the Plains landed would trigger, go on the stack, give
	// every player a window, and only then do nothing.
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && ev.Source == hounds[0] {
			t.Fatalf("Loyal Warhound triggered (%q): its ETB was checked before the Sword's Plains "+
				"had landed, which is two entries rather than one (CR 603.6a)", ev.Label)
		}
	}
}

// TestSwordOfHearthAndHomeWithNoTargetStillFetches: "up to one" — with
// no creature chosen the land is still searched for, put, and the
// library shuffled.
func TestSwordOfHearthAndHomeWithNoTargetStillFetches(t *testing.T) {
	g, me, opp, bear, _, _, _ := ebSwordSetup(t)
	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		answerSearchByID(t, g, me.ID, c.SearchCards[0])
		passPriorityAroundTable(t, g)
	}
	if n := len(ebLandsOf(g, me.ID)); n != 3 {
		t.Errorf("seat 0 controls %d lands, want 3", n)
	}
	if len(battlefieldIDsNamed(g, "Loyal Warhound")) != 1 {
		t.Error("the Warhound was not chosen and stays where it is")
	}
}

// --- #1327: Phelia, Exuberant Shepherd ---------------------------------

// ebPheliaAttack puts Phelia on the battlefield under seat 0, attacks
// with her and exiles `target` with her trigger.
func ebPheliaAttack(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	phelia := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Phelia, Exuberant Shepherd", TypeLine: "Legendary Creature — Dog",
		OracleID: ebPheliaOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(phelia, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no pick_target prompt from Phelia's attack trigger")
	}
	if hasID(prompt.PickTargetCards, phelia) {
		t.Error("Phelia is offered as her own target; the clause says OTHER")
	}
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	if !exileHas(g, target) {
		t.Fatal("Phelia's trigger did not exile the target")
	}
	return phelia
}

func TestPheliaGrowsWhenYourCardComesBackToYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	phelia := ebPheliaAttack(t, g, bear)
	if counterCount(g, phelia, "+1/+1") != 0 {
		t.Fatal("a counter before the return")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	bears := battlefieldIDsNamed(g, "Bear")
	if len(bears) != 1 || bears[0] == bear {
		t.Fatalf("the Bear should be back as a new object; battlefield has %v", bears)
	}
	if got := counterCount(g, phelia, "+1/+1"); got != 1 {
		t.Errorf("Phelia has %d +1/+1 counters, want 1 — the Bear entered under your control", got)
	}
}

func TestPheliaDoesNotGrowWhenTheCardGoesHome(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	theirs := seedBear(g, opp.ID)
	phelia := ebPheliaAttack(t, g, theirs)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	bears := battlefieldIDsNamed(g, "Bear")
	if len(bears) != 1 {
		t.Fatalf("the opponent's Bear should be back; battlefield has %v", bears)
	}
	if c, _ := battlefieldCard(g, bears[0]); c.Controller != opp.ID {
		t.Error("the Bear returns under its owner's control")
	}
	if got := counterCount(g, phelia, "+1/+1"); got != 0 {
		t.Errorf("Phelia has %d counters; the card entered under its owner's control, not yours", got)
	}
}

// TestPheliaWaitsForAPausedReturn is #1327 itself. The returning card's
// entry asks its controller a question (a pay-life entry, the shockland
// shape on a creature). While it is open nothing has entered, and
// Phelia must not have her counter yet — a counter "before an entry
// that could still fail" is stronger than printed. Once it is answered
// the card has entered under your control and the counter arrives.
func TestPheliaWaitsForAPausedReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches:       []game.EventKind{game.EventZoneMove},
			EntryLifeCost: 2,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.CardID == bear &&
					ev.OldZone == game.ZoneExile && ev.NewZone == game.ZoneBattlefield
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = true
				return nil
			},
			Controller: func(*game.ReplacementEvent, *game.Game, *game.Card) uuid.UUID { return me.ID },
			Label:      "test: pay 2 life or the returning Bear enters tapped",
		})
	})
	phelia := ebPheliaAttack(t, g, bear)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	if entryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the returning Bear's entry question was not asked")
	}
	if !exileHas(g, bear) {
		t.Fatal("the Bear left exile before its question was answered")
	}
	if got := counterCount(g, phelia, "+1/+1"); got != 0 {
		t.Fatalf("Phelia has %d counters while the return is still paused", got)
	}

	answerEntryPayLife(t, g, me.ID, false)

	if len(battlefieldIDsNamed(g, "Bear")) != 1 {
		t.Fatal("the Bear did not come back once the question was answered")
	}
	if got := counterCount(g, phelia, "+1/+1"); got != 1 {
		t.Errorf("Phelia has %d counters after the paused return finished, want 1", got)
	}
}
