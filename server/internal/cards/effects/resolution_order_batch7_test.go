package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_order_batch7_test.go — #4155 (CR 608.2c): the seventh and last
// batch of hand-written bodies written as steps. It adds a real card for each
// shape batch 7 converted, each with the prompt open:
//
//   - a spell's ResolveSteps behind Khârn the Betrayer's question, on a table
//     that is a restore point, so it is run live and on a restored table
//     (Searing Barb, Awaken the Inferno, Cut In, No Admittance, Ribbons of
//     Night);
//   - a spell's ResolveSteps on a CR 616 counter order (Tam's Resistance);
//   - a spell's ResolveSteps on a dredge offer, the later sentence a skipped
//     turn, a queued discard, or a ModeSteps bullet's life (Meditate, Timeline
//     Inquiry, Surgical Precision);
//   - a spell that takes a card from outside the game and exiles itself after
//     the question (Death Wish);
//   - a continuation run from a "you may" (The Legend of Yangchen's chapter II)
//     and a row's Do behind a saga chapter (The Great Synthesis);
//   - a step that reads the damage dealt back from the event log, where one
//     piece cannot carry a Go variable to the next (Violent Echoes).
//
// A dredge offer and a CR 616 order resume through closures, so a table that
// holds one is not a restore point; an undo (a clone) stands in for the restore
// there, as in resolution_order_batch1_test.go.

// b7TokenCount is the tokens a player controls.
func b7TokenCount(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && IsToken(c) {
			n++
		}
	}
	return n
}

// Each spell below deals damage to Khârn the Betrayer, whose prevention asks its
// controller for an opponent (#2040), and then does something more. The
// something more is the next instruction, so it waits for the answer, live and
// on a table restored while the question is open.
func TestKharnsQuestionHoldsTheSentencesBehindADamageSpell(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		second                 bool // a second target: a creature of the caster's
		observe                func(g *game.Game, me *game.Player, mine uuid.UUID) int
		delta                  int
	}{
		// Searing Barb — "deals 2 damage to any target. Incubate 1."
		{"Searing Barb", "Instant", searingBarbOracle, false,
			func(g *game.Game, me *game.Player, _ uuid.UUID) int { return b7TokenCount(g, me.ID) }, 1},
		// Awaken the Inferno — "deals 6 damage to target creature or planeswalker an
		// opponent controls. Put a +1/+1 counter on up to one target creature you control."
		{"Awaken the Inferno", "Sorcery", rfaAwakenInferno, true,
			func(g *game.Game, _ *game.Player, mine uuid.UUID) int {
				c, _ := battlefieldCard(g, mine)
				return c.Counters["+1/+1"]
			}, 1},
		// Cut In — "deals 4 damage to target creature. Create a Young Hero Role token
		// attached to up to one target creature you control."
		{"Cut In", "Sorcery", cutInOracle, true,
			func(g *game.Game, me *game.Player, _ uuid.UUID) int { return b7TokenCount(g, me.ID) }, 1},
		// No Admittance — "deals 3 damage to any target. Empower Jace 1."
		{"No Admittance", "Sorcery", oracleNoAdmittance, false,
			func(g *game.Game, me *game.Player, _ uuid.UUID) int { return len(jaceTokensOf(g, me.ID)) }, 1},
		// Ribbons of Night — "deals 4 damage to target creature and you gain 4 life."
		{"Ribbons of Night", "Sorcery", wishRibbonsOracle, false,
			func(_ *game.Game, me *game.Player, _ uuid.UUID) int { return me.Life }, 4},
	}
	for _, tc := range cases {
		for _, restored := range []bool{false, true} {
			g := newCatalogGame(t)
			toMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
			kharn := kharnPush(g, owner.ID, owner.ID)
			mine := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
			targets := cardRefs(kharn)
			if tc.second {
				targets = append(targets, cardRefs(mine)...)
			}
			before := tc.observe(g, me, mine)
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, targets)
			passPriorityAroundTable(t, g)
			if latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, owner.ID) == nil {
				t.Fatalf("%s restored=%v: Khârn's question is not open", tc.name, restored)
			}
			if got := tc.observe(g, me, mine); got != before {
				t.Fatalf("%s restored=%v: %d with the question open, want %d: the next sentence ran before the answer",
					tc.name, restored, got, before)
			}
			if restored {
				g = restoreRoundTrip(t, g, true)
				me = g.Seats[g.Turn.ActiveSeat]
			}
			kharnAnswer(t, g, owner.ID, third.ID)
			if got := tc.observe(g, me, mine); got != before+tc.delta {
				t.Fatalf("%s restored=%v: %d after the answer, want %d", tc.name, restored, got, before+tc.delta)
			}
		}
	}
}

// Tam's Resistance — "Put a +1/+1 counter on up to one target creature. It
// gains vigilance until end of turn. Empower Jace 4." With Doubling Season and
// Hardened Scales out the counter waits on the CR 616 order, and the vigilance
// and the Jace token wait behind it.
func TestTamsResistanceVigilanceAndJaceWaitForTheCounterOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	castCatalogSpell(t, g, "Tam's Resistance", "Sorcery", oracleTamsResistance,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("want the one CR 616 prompt open, got %d choices", len(g.PendingChoices))
	}
	if effectiveAbilitiesContain(t, g, bear, "vigilance") || len(jaceTokensOf(g, me.ID)) != 0 {
		t.Fatal("the vigilance or the Jace token came before the counter order was answered")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		answerOrderBySource(t, table, ds, hs)
		if got := b12Counter(t, table, bear, "+1/+1"); got != 3 {
			t.Fatalf("+1/+1 counters %d, want 3 ((1×2)+1, Doubling Season first)", got)
		}
		if !effectiveAbilitiesContain(t, table, bear, "vigilance") {
			t.Fatal("the creature did not gain vigilance after the order")
		}
		if len(jaceTokensOf(table, me.ID)) == 0 {
			t.Fatal("Jace was not empowered after the order")
		}
	}
}

// Meditate — "Draw four cards. You skip your next turn." The skip waits for the
// dredge offers the draws make.
func TestMeditateSkipWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	castCatalogSpell(t, g, "Meditate", "Instant", meditateOracle, nil)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if g.TurnSkipsFor(me.ID) != 0 {
		t.Fatal("the skip was owed with dredge open: it ran before the answer")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		for optionalOffer(table, p) != nil {
			answerOptional(t, table, p, false)
		}
		table.SettleResolution()
		if got := table.TurnSkipsFor(p.ID); got != 1 {
			t.Fatalf("owes %d skips after the draws, want 1", got)
		}
	}
}

// Timeline Inquiry — "Draw three cards. If you didn't cast this spell using
// teamwork, discard a card." The discard waits for the dredge offers the draws
// make.
func TestTimelineInquiryDiscardWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	if _, err := castPaying(t, g, "Timeline Inquiry", "Instant", timelineInquiryOracle, nil, nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if discardOwed(g, me.ID) != 0 {
		t.Fatal("the discard was owed with dredge open: it ran before the answer")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		for optionalOffer(table, p) != nil {
			answerOptional(t, table, p, false)
		}
		table.SettleResolution()
		if got := discardOwed(table, p.ID); got != 1 {
			t.Fatalf("discard owed %d after the draws, want 1", got)
		}
	}
}

// Surgical Precision's second mode — "You draw a card and gain 2 life." The
// life waits for the dredge offer the draw makes.
func TestSurgicalPrecisionLifeWaitsForDredge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	life := me.Life
	castModal(t, g, "Surgical Precision", "Sorcery", rfbSurgicalPrecisionOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if me.Life != life {
		t.Fatalf("life %d with dredge open, want %d: the gain ran before the answer", me.Life, life)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		for optionalOffer(table, p) != nil {
			answerOptional(t, table, p, false)
		}
		table.SettleResolution()
		if p.Life != life+2 {
			t.Fatalf("life %d after the draw, want %d", p.Life, life+2)
		}
	}
}

// Death Wish — "You may reveal a card you own from outside the game and put it
// into your hand. You lose half your life, rounded up. Exile Death Wish." The
// life loss and the exile wait for the question the take asks.
func TestDeathWishLifeLossAndExileWaitForTheQuestion(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Lightning Bolt", "Instant"))
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Death Wish", "Sorcery", deathWishOracle, nil)
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	if me.Life != life || outsideExileHas(g, "Death Wish") {
		t.Fatalf("life %d (want %d) or exiled %v with the question open", me.Life, life, outsideExileHas(g, "Death Wish"))
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[0]
		c := latestChooseCardsFor(table, p.ID)
		if err := table.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{sideCardID(p, "Lightning Bolt")}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		table.SettleResolution()
		if want := life - (life+1)/2; p.Life != want {
			t.Fatalf("life %d after the take, want %d", p.Life, want)
		}
		if !outsideExileHas(table, "Death Wish") || !outsideHandHas(p, "Lightning Bolt") {
			t.Fatal("Death Wish did not exile itself or the card did not come into the hand")
		}
	}
}

// The Legend of Yangchen's chapter II — "You may have target opponent draw three
// cards. If you do, you draw three cards." The yes is a continuation written as
// steps: your draw waits for the dredge offer the opponent's draw makes.
func TestYangchenChapterTwoYourDrawWaitsForTheOpponentsDredge(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	awbPadLibrary(me)
	awbPadLibrary(opp)

	importAndCast(t, g, yangchenRow(), me)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	gyCardWithOracle(opp, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	answerMayChoice(t, g, me.ID, true)
	if optionalOffer(g, opp) == nil {
		t.Fatal("the opponent's dredge is not offered")
	}
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d with the opponent's dredge open, want %d: the draw ran before the answer", me.Hand.Size(), hand)
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		you, them := table.Seats[seat], table.Seats[(seat+1)%len(table.Seats)]
		for optionalOffer(table, them) != nil {
			answerOptional(t, table, them, false)
		}
		table.SettleResolution()
		if you.Hand.Size() != hand+3 {
			t.Fatalf("hand %d after the opponent's draws, want %d", you.Hand.Size(), hand+3)
		}
	}
}

// The Great Synthesis's chapter I — "Draw cards equal to the number of cards in
// your hand. For as long as you control this Saga, you have no maximum hand
// size." The chapter is a row's Do: the lifted maximum waits for the dredge
// offers the draws make.
func TestGreatSynthesisNoMaximumWaitsForTheDrawsDredge(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	greatSynthesisToChapterOne(t, g, me)
	awbPadLibrary(me)
	gyCardWithOracle(me, "Stinkweed Imp", "Creature — Imp", stinkweedImpOracle)
	passPriorityAroundTable(t, g)
	if optionalOffer(g, me) == nil {
		t.Fatal("dredge is not offered")
	}
	if got := mhMax(g, me); got == game.NoMaxHandSize {
		t.Fatal("the maximum was lifted with dredge open: it ran before the answer")
	}
	undo := g.Clone()
	for _, table := range []*game.Game{g, undo} {
		p := table.Seats[table.Turn.ActiveSeat]
		for optionalOffer(table, p) != nil {
			answerOptional(t, table, p, false)
		}
		table.SettleResolution()
		if got := mhMax(table, p); got != game.NoMaxHandSize {
			t.Fatalf("max %d after the draws, want no maximum", got)
		}
	}
}

// Violent Echoes — "deals 6 damage to target creature or planeswalker. Empower
// Jace X, where X is the amount of excess damage dealt this way." The first step
// files the lethal amount and the log's cursor and the second reads the damage
// dealt after the cursor, so the excess survives a pause between the two. A 2/2
// with 1 damage already marked has 1 left to take: 5 is excess.
func TestViolentEchoesEmpowersJaceByTheExcess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
				g.Battlefield.Cards[i].DamageMarked = 1
			}
		}
	})
	castCatalogSpell(t, g, "Violent Echoes", "Instant", fraEchoesOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 5 {
		t.Fatalf("Jace loyalty %d, want 5 (6 damage less the 1 that was lethal)", got)
	}
}
