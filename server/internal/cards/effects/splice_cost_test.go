package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// splice_cost_test.go — a splice cost that is not mana (#4055, ADR 0202's
// 2026-10-19 amendment, CR 702.47a): Veil of Secrecy, Horobi's Whisper, Roar
// of Jukai, Hundred-Talon Strike and Torrent of Stone. The host is Consuming
// Vortex ({1}{U}, "Return target creature to its owner's hand"), so a spell
// with a main text and a spliced text of its own is read side by side. Every
// refusal is paired with an acceptance on the same board.

const (
	veilOfSecrecyOracle     = "abfde51c-dfe7-4382-8aa2-f0a7ac1b7048"
	horobisWhisperOracle    = "2447f1f4-3569-4afb-a119-2ad9778eaf39"
	roarOfJukaiOracle       = "d611c594-dd2d-4f4b-a958-1f05f22f7f36"
	hundredTalonStrikeOracl = "fbfaf009-8276-4687-a3db-5a71dc13f174"
	torrentOfStoneOracle    = "bc6ee33a-4245-4edb-a6b3-0bdcc2883fbc"
)

// castPaid casts `spell` for exactly {1}{U} with the splice costs paid by
// `pays`. A refused cast empties the pool, so the next attempt starts from
// the same board.
func (b spliceBoard) castPaid(spell uuid.UUID, splices []uuid.UUID, pays []game.SplicePayment, targets ...game.TargetRef) error {
	b.me.ManaPool.EmptyPool()
	if err := b.g.AddManaForEffect(b.me.ID, uuid.Nil, "{C}{U}"); err != nil {
		return err
	}
	err := b.g.CastSpell(b.me.ID, spell, game.CastSpellParams{
		Strict: true, Splices: splices, SplicePayments: pays, Targets: targets,
	})
	if err != nil {
		b.me.ManaPool.EmptyPool()
	}
	return err
}

func (b spliceBoard) land(owner uuid.UUID, subtype string) uuid.UUID {
	return pushPermanentForTest(b.g, owner, subtype, "", "Basic Land — "+subtype)
}

func (b spliceBoard) offers(t *testing.T, spell uuid.UUID) []protocol.SpliceOfferView {
	t.Helper()
	for i := range b.me.Hand.Cards {
		b.me.Hand.Cards[i].KnownBy = map[uuid.UUID]bool{b.me.ID: true}
	}
	v := protocol.ViewOfGameFor(b.g, b.me.ID.String())
	for _, s := range v.Seats {
		for _, c := range s.Hand.Cards {
			if c.InstanceID == spell.String() {
				return c.SpliceOffers
			}
		}
	}
	t.Fatalf("card %s not in the view", spell)
	return nil
}

func offerFor(offers []protocol.SpliceOfferView, card uuid.UUID) *protocol.SpliceOfferView {
	for i := range offers {
		if offers[i].CardID == card.String() {
			return &offers[i]
		}
	}
	return nil
}

func inBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

// badAnnouncement asserts the cast was refused. The reason is the validator
// of the component that refused it (an illegal target, a card not found, a
// permanent the caster does not control), so only the refusal is pinned.
func badAnnouncement(t *testing.T, what string, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: accepted", what)
	}
}

func (b spliceBoard) vortexTarget() []game.TargetRef { return []game.TargetRef{spCardRef(b.bear)} }

// --- Torrent of Stone: sacrifice two Mountains ------------------------------

func TestTorrentOfStoneSplicedSacrificesTwoMountainsAndDealsFour(t *testing.T) {
	b := newSpliceBoard(t)
	m1, m2, m3 := b.land(b.me.ID, "Mountain"), b.land(b.me.ID, "Mountain"), b.land(b.me.ID, "Mountain")
	forest := b.land(b.me.ID, "Forest")
	theirs := b.land(b.opp.ID, "Mountain")
	vortex := b.vortex()
	torrent := b.hand("Torrent of Stone", "Instant — Arcane", "{3}{R}", torrentOfStoneOracle, "R")
	splices := []uuid.UUID{torrent}
	targets := []game.TargetRef{spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))}

	for _, tc := range []struct {
		name string
		pay  game.SplicePayment
	}{
		{"no payment", game.SplicePayment{}},
		{"one Mountain", game.SplicePayment{SacrificeIDs: []uuid.UUID{m1}}},
		{"three Mountains", game.SplicePayment{SacrificeIDs: []uuid.UUID{m1, m2, m3}}},
		{"the same Mountain twice", game.SplicePayment{SacrificeIDs: []uuid.UUID{m1, m1}}},
		{"a Forest", game.SplicePayment{SacrificeIDs: []uuid.UUID{m1, forest}}},
		{"an opponent's Mountain", game.SplicePayment{SacrificeIDs: []uuid.UUID{m1, theirs}}},
		{"a payment for a component the cost lacks", game.SplicePayment{SacrificeIDs: []uuid.UUID{m1, m2}, ReturnIDs: []uuid.UUID{m3}}},
	} {
		badAnnouncement(t, tc.name, b.castPaid(vortex, splices, []game.SplicePayment{tc.pay}, targets...))
	}
	if !inBattlefield(b.g, m1) || !inBattlefield(b.g, m2) {
		t.Fatal("a refused cast sacrificed a Mountain")
	}
	if err := b.castPaid(vortex, splices, []game.SplicePayment{{SacrificeIDs: []uuid.UUID{m1, m2}}}, targets...); err != nil {
		t.Fatalf("two Mountains: %v", err)
	}
	if inBattlefield(b.g, m1) || inBattlefield(b.g, m2) {
		t.Error("a Mountain named to the splice cost is still on the battlefield")
	}
	if !inBattlefield(b.g, m3) || !inBattlefield(b.g, forest) {
		t.Error("a land that was not named was sacrificed")
	}
	if !spInHand(b.me, torrent) {
		t.Error("Torrent of Stone left the hand: a spliced card stays there (CR 702.47a)")
	}
	passPriorityAroundTable(t, b.g)
	if !spInHand(b.opp, b.bear) {
		t.Error("the main text did not return the bear")
	}
	if inBattlefield(b.g, b.wolf) {
		t.Error("the spliced text did not deal 4 damage to the wolf")
	}
}

// A card whose cost the caster can't pay is not offered (CR 118.3), and is
// offered with the Mountains as its options the moment they can.
func TestTorrentOfStoneIsOfferedOnlyWithTwoMountains(t *testing.T) {
	b := newSpliceBoard(t)
	vortex := b.vortex()
	torrent := b.hand("Torrent of Stone", "Instant — Arcane", "{3}{R}", torrentOfStoneOracle, "R")
	b.land(b.me.ID, "Mountain")
	if o := offerFor(b.offers(t, vortex), torrent); o != nil {
		t.Fatalf("offered with one Mountain: %+v", o)
	}
	m2 := b.land(b.me.ID, "Mountain")
	o := offerFor(b.offers(t, vortex), torrent)
	if o == nil {
		t.Fatal("not offered with two Mountains")
	}
	if o.ManaCost != "" || o.CostLabel != "Sacrifice two Mountains" || o.Label != "Splice onto Arcane—Sacrifice two Mountains." {
		t.Errorf("offer words: mana %q, cost %q, label %q", o.ManaCost, o.CostLabel, o.Label)
	}
	if o.SacrificeOptions == nil || o.SacrificeOptions.Min != 2 || o.SacrificeOptions.Max != 2 ||
		len(o.SacrificeOptions.Cards) != 2 || !slices.Contains(o.SacrificeOptions.Cards, m2.String()) {
		t.Errorf("sacrifice options %+v, want the two Mountains, 2 to 2", o.SacrificeOptions)
	}
}

// --- Horobi's Whisper: exile four cards from your graveyard -----------------

func TestHorobisWhisperSplicedExilesFourAndDestroysWithASwamp(t *testing.T) {
	b := newSpliceBoard(t)
	b.land(b.me.ID, "Swamp")
	vortex := b.vortex()
	whisper := b.hand("Horobi's Whisper", "Instant — Arcane", "{1}{B}{B}", horobisWhisperOracle, "B")
	var yard []uuid.UUID
	for i := 0; i < 5; i++ {
		yard = append(yard, gyCard(b.me, "Filler", "Sorcery", "{1}", ""))
	}
	theirs := gyCard(b.opp, "Theirs", "Sorcery", "{1}", "")
	splices := []uuid.UUID{whisper}
	targets := []game.TargetRef{spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))}

	for _, tc := range []struct {
		name string
		pay  game.SplicePayment
	}{
		{"three cards", game.SplicePayment{ExileIDs: yard[:3]}},
		{"five cards", game.SplicePayment{ExileIDs: yard[:5]}},
		{"a card twice", game.SplicePayment{ExileIDs: []uuid.UUID{yard[0], yard[0], yard[1], yard[2]}}},
		{"an opponent's card", game.SplicePayment{ExileIDs: []uuid.UUID{yard[0], yard[1], yard[2], theirs}}},
		{"a card in the hand", game.SplicePayment{ExileIDs: []uuid.UUID{yard[0], yard[1], yard[2], whisper}}},
	} {
		badAnnouncement(t, tc.name, b.castPaid(vortex, splices, []game.SplicePayment{tc.pay}, targets...))
	}
	if err := b.castPaid(vortex, splices, []game.SplicePayment{{ExileIDs: yard[:4]}}, targets...); err != nil {
		t.Fatalf("four cards: %v", err)
	}
	for _, id := range yard[:4] {
		if !inExile(b.g, id) {
			t.Errorf("%v was named to the splice cost and is not exiled", id)
		}
	}
	if inExile(b.g, yard[4]) {
		t.Error("a card that was not named was exiled")
	}
	passPriorityAroundTable(t, b.g)
	if inBattlefield(b.g, b.wolf) {
		t.Error("with a Swamp, the spliced text did not destroy the nonblack wolf")
	}
}

func TestHorobisWhisperSplicedWithoutASwampStillPaysItsCost(t *testing.T) {
	b := newSpliceBoard(t)
	vortex := b.vortex()
	whisper := b.hand("Horobi's Whisper", "Instant — Arcane", "{1}{B}{B}", horobisWhisperOracle, "B")
	var yard []uuid.UUID
	for i := 0; i < 4; i++ {
		yard = append(yard, gyCard(b.me, "Filler", "Sorcery", "{1}", ""))
	}
	if err := b.castPaid(vortex, []uuid.UUID{whisper}, []game.SplicePayment{{ExileIDs: yard}},
		spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))); err != nil {
		t.Fatalf("splice: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	if !inBattlefield(b.g, b.wolf) {
		t.Error("without a Swamp the spliced text destroyed the wolf")
	}
	for _, id := range yard {
		if !inExile(b.g, id) {
			t.Error("the cost was not paid: the condition is the text's, not the cost's")
		}
	}
}

func TestHorobisWhisperIsOfferedOnlyWithFourCardsInTheGraveyard(t *testing.T) {
	b := newSpliceBoard(t)
	vortex := b.vortex()
	whisper := b.hand("Horobi's Whisper", "Instant — Arcane", "{1}{B}{B}", horobisWhisperOracle, "B")
	for i := 0; i < 3; i++ {
		gyCard(b.me, "Filler", "Sorcery", "{1}", "")
	}
	gyCard(b.opp, "Theirs", "Sorcery", "{1}", "")
	if o := offerFor(b.offers(t, vortex), whisper); o != nil {
		t.Fatalf("offered with three cards in the graveyard: %+v", o)
	}
	gyCard(b.me, "Filler", "Sorcery", "{1}", "")
	o := offerFor(b.offers(t, vortex), whisper)
	if o == nil {
		t.Fatal("not offered with four cards in the graveyard")
	}
	if o.ExileCostN != 4 || len(o.ExileCostOptions) != 4 || o.CostLabel != "Exile four cards from your graveyard" {
		t.Errorf("offer %+v, want four exile options", o)
	}
}

// --- Roar of Jukai: an opponent gains 5 life -------------------------------

func TestRoarOfJukaiSplicedGivesTheChosenOpponentFiveLife(t *testing.T) {
	b := newSpliceBoard(t)
	third := b.g.Seats[2]
	vortex := b.vortex()
	roar := b.hand("Roar of Jukai", "Instant — Arcane", "{2}{G}", roarOfJukaiOracle, "G")
	splices := []uuid.UUID{roar}
	life := map[uuid.UUID]int{}
	for _, p := range b.g.Seats {
		life[p.ID] = p.Life
	}
	for _, tc := range []struct {
		name string
		pay  game.SplicePayment
	}{
		{"no opponent", game.SplicePayment{}},
		{"the caster", game.SplicePayment{Opponent: b.me.ID}},
		{"an unseated player", game.SplicePayment{Opponent: uuid.New()}},
		{"an opponent and a card", game.SplicePayment{Opponent: b.opp.ID, SacrificeIDs: []uuid.UUID{b.bear}}},
	} {
		badAnnouncement(t, tc.name, b.castPaid(vortex, splices, []game.SplicePayment{tc.pay}, b.vortexTarget()...))
	}
	for _, p := range b.g.Seats {
		if p.Life != life[p.ID] {
			t.Fatalf("a refused cast moved %s's life", p.Name)
		}
	}
	if err := b.castPaid(vortex, splices, []game.SplicePayment{{Opponent: third.ID}}, b.vortexTarget()...); err != nil {
		t.Fatalf("choosing an opponent: %v", err)
	}
	if got := third.Life - life[third.ID]; got != 5 {
		t.Errorf("the chosen opponent gained %d, want 5", got)
	}
	if b.opp.Life != life[b.opp.ID] || b.me.Life != life[b.me.ID] {
		t.Error("someone other than the chosen opponent changed life")
	}
	if !spInHand(b.me, roar) {
		t.Error("Roar of Jukai left the hand")
	}
}

func TestRoarOfJukaiSplicedPumpsBlockedCreaturesWithAForest(t *testing.T) {
	for _, forest := range []bool{true, false} {
		b := newSpliceBoard(t)
		if forest {
			b.land(b.me.ID, "Forest")
		}
		blocked := twPut(b.g, b.me.ID, "Blocked", "Creature — Test", []string{"R"})
		free := twPut(b.g, b.me.ID, "Free", "Creature — Test", []string{"R"})
		blocker := twPut(b.g, b.opp.ID, "Blocker", "Creature — Test", []string{"G"})
		vortex := b.vortex()
		roar := b.hand("Roar of Jukai", "Instant — Arcane", "{2}{G}", roarOfJukaiOracle, "G")
		pr7bAttack(t, b.g, b.opp.ID, blocked, free)
		pr7bBlock(t, b.g, map[uuid.UUID]uuid.UUID{blocker: blocked})
		if err := b.castPaid(vortex, []uuid.UUID{roar}, []game.SplicePayment{{Opponent: b.opp.ID}}, b.vortexTarget()...); err != nil {
			t.Fatalf("forest=%v: %v", forest, err)
		}
		passPriorityAroundTable(t, b.g)
		want := 2
		if forest {
			want = 4
		}
		if got := effectivePower(t, b.g, blocked); got != want {
			t.Errorf("forest=%v: the blocked attacker has power %d, want %d", forest, got, want)
		}
		if got := effectivePower(t, b.g, free); got != 2 {
			t.Errorf("forest=%v: the unblocked attacker has power %d, want 2", forest, got)
		}
	}
}

// An opponent who can't gain life can't be chosen (CR 614.17b), and is not
// offered; with nobody left the card is not offered at all.
func TestRoarOfJukaiOpponentsWhoCannotGainLifeAreNotOffered(t *testing.T) {
	b := newSpliceBoard(t)
	vortex := b.vortex()
	roar := b.hand("Roar of Jukai", "Instant — Arcane", "{2}{G}", roarOfJukaiOracle, "G")
	o := offerFor(b.offers(t, vortex), roar)
	if o == nil || !o.ChoosesOpponent || len(o.OpponentOptions) != 3 {
		t.Fatalf("offer %+v, want a choice of the three opponents", o)
	}
	b.g.WithWriteLock(func() {
		b.g.PlayerCantGainLifeForEffect(uuid.Nil, b.opp.ID, game.IndefiniteDuration(), "test")
	})
	o = offerFor(b.offers(t, vortex), roar)
	if o == nil || slices.Contains(o.OpponentOptions, b.opp.ID.String()) || len(o.OpponentOptions) != 2 {
		t.Fatalf("offer %+v, want the two opponents who can gain life", o)
	}
	badAnnouncement(t, "an opponent who can't gain life",
		b.castPaid(vortex, []uuid.UUID{roar}, []game.SplicePayment{{Opponent: b.opp.ID}}, b.vortexTarget()...))
	b.g.Seats[2].Eliminated, b.g.Seats[3].Eliminated = true, true
	if o := offerFor(b.offers(t, vortex), roar); o != nil {
		t.Errorf("offered with nobody to give the life to: %+v", o)
	}
}

// --- Hundred-Talon Strike: tap an untapped white creature ------------------

func TestHundredTalonStrikeSplicedTapsAWhiteCreature(t *testing.T) {
	b := newSpliceBoard(t)
	white := twPut(b.g, b.me.ID, "Knight", "Creature — Knight", []string{"W"})
	green := twPut(b.g, b.me.ID, "Druid", "Creature — Druid", []string{"G"})
	theirsWhite := twPut(b.g, b.opp.ID, "Their Knight", "Creature — Knight", []string{"W"})
	tapped := pushBattlefieldCardWithTimestamp(b.g, game.Card{
		InstanceID: uuid.New(), Name: "Tired Knight", TypeLine: "Creature — Knight",
		Power: 2, Toughness: 2, Colors: []string{"W"}, Tapped: true, Owner: b.me.ID, Controller: b.me.ID,
	})
	vortex := b.vortex()
	strike := b.hand("Hundred-Talon Strike", "Instant — Arcane", "{W}", hundredTalonStrikeOracl, "W")
	splices := []uuid.UUID{strike}
	targets := []game.TargetRef{spCardRef(b.bear), splicedRef(1, spCardRef(green))}

	for _, tc := range []struct {
		name string
		pay  game.SplicePayment
	}{
		{"no payment", game.SplicePayment{}},
		{"a green creature", game.SplicePayment{TapIDs: []uuid.UUID{green}}},
		{"an opponent's creature", game.SplicePayment{TapIDs: []uuid.UUID{theirsWhite}}},
		{"a tapped creature", game.SplicePayment{TapIDs: []uuid.UUID{tapped}}},
		{"two creatures", game.SplicePayment{TapIDs: []uuid.UUID{white, tapped}}},
	} {
		badAnnouncement(t, tc.name, b.castPaid(vortex, splices, []game.SplicePayment{tc.pay}, targets...))
	}
	if c, _ := battlefieldCard(b.g, white); c.Tapped {
		t.Fatal("a refused cast tapped the Knight")
	}
	if err := b.castPaid(vortex, splices, []game.SplicePayment{{TapIDs: []uuid.UUID{white}}}, targets...); err != nil {
		t.Fatalf("tapping the Knight: %v", err)
	}
	if c, _ := battlefieldCard(b.g, white); !c.Tapped {
		t.Error("the Knight named to the splice cost is untapped")
	}
	passPriorityAroundTable(t, b.g)
	if got := effectivePower(t, b.g, green); got != 3 {
		t.Errorf("the target has power %d, want 3 (+1/+0)", got)
	}
	if !effectiveAbilitiesContain(t, b.g, green, "first strike") {
		t.Error("the target did not gain first strike")
	}
}

// A creature that came under the caster's control this turn can pay: the
// cost taps it, and it is not the {T} symbol (CR 302.6).
func TestHundredTalonStrikeCostIsNotTheTapSymbol(t *testing.T) {
	b := newSpliceBoard(t)
	fresh := pushBattlefieldCardWithTimestamp(b.g, game.Card{
		InstanceID: uuid.New(), Name: "Fresh Knight", TypeLine: "Creature — Knight",
		Power: 2, Toughness: 2, Colors: []string{"W"}, SummonedThisTurn: true, Owner: b.me.ID, Controller: b.me.ID,
	})
	vortex := b.vortex()
	strike := b.hand("Hundred-Talon Strike", "Instant — Arcane", "{W}", hundredTalonStrikeOracl, "W")
	if err := b.castPaid(vortex, []uuid.UUID{strike}, []game.SplicePayment{{TapIDs: []uuid.UUID{fresh}}},
		spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))); err != nil {
		t.Fatalf("a summoning-sick Knight pays the tap: %v", err)
	}
}

// --- Veil of Secrecy: return a blue creature you control -------------------

func TestVeilOfSecrecySplicedReturnsABlueCreature(t *testing.T) {
	b := newSpliceBoard(t)
	blue := twPut(b.g, b.me.ID, "Merfolk", "Creature — Merfolk", []string{"U"})
	red := twPut(b.g, b.me.ID, "Ogre", "Creature — Ogre", []string{"R"})
	theirsBlue := twPut(b.g, b.opp.ID, "Their Merfolk", "Creature — Merfolk", []string{"U"})
	vortex := b.vortex()
	veil := b.hand("Veil of Secrecy", "Instant — Arcane", "{1}{U}", veilOfSecrecyOracle, "U")
	splices := []uuid.UUID{veil}
	targets := []game.TargetRef{spCardRef(b.bear), splicedRef(1, spCardRef(red))}

	for _, tc := range []struct {
		name string
		pay  game.SplicePayment
	}{
		{"no payment", game.SplicePayment{}},
		{"a red creature", game.SplicePayment{ReturnIDs: []uuid.UUID{red}}},
		{"an opponent's blue creature", game.SplicePayment{ReturnIDs: []uuid.UUID{theirsBlue}}},
		{"two creatures", game.SplicePayment{ReturnIDs: []uuid.UUID{blue, red}}},
	} {
		badAnnouncement(t, tc.name, b.castPaid(vortex, splices, []game.SplicePayment{tc.pay}, targets...))
	}
	if !inBattlefield(b.g, blue) {
		t.Fatal("a refused cast returned the Merfolk")
	}
	if err := b.castPaid(vortex, splices, []game.SplicePayment{{ReturnIDs: []uuid.UUID{blue}}}, targets...); err != nil {
		t.Fatalf("returning the Merfolk: %v", err)
	}
	if inBattlefield(b.g, blue) || !spInHand(b.me, blue) {
		t.Error("the Merfolk named to the splice cost did not return to its owner's hand")
	}
	passPriorityAroundTable(t, b.g)
	if !effectiveAbilitiesContain(t, b.g, red, "shroud") {
		t.Error("the target did not gain shroud")
	}
	if restrictionsOf(t, b.g, red)&game.CantBeBlocked == 0 {
		t.Error("the target can be blocked")
	}
}

// --- the cost belongs to its card ------------------------------------------

// One object pays one cost of a cast (CR 118.3): two splice cards may not
// name the same creature, and a splice payment may not name an object the
// cast's own costs spend.
func TestSpliceCostsOfTwoCardsNeverShareAnObject(t *testing.T) {
	b := newSpliceBoard(t)
	both := twPut(b.g, b.me.ID, "Azorius Knight", "Creature — Knight", []string{"W", "U"})
	other := twPut(b.g, b.me.ID, "Merfolk", "Creature — Merfolk", []string{"U"})
	vortex := b.vortex()
	veil := b.hand("Veil of Secrecy", "Instant — Arcane", "{1}{U}", veilOfSecrecyOracle, "U")
	strike := b.hand("Hundred-Talon Strike", "Instant — Arcane", "{W}", hundredTalonStrikeOracl, "W")
	splices := []uuid.UUID{veil, strike}
	targets := []game.TargetRef{spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf)), splicedRef(2, spCardRef(b.wolf))}

	badAnnouncement(t, "one creature returned and tapped", b.castPaid(vortex, splices, []game.SplicePayment{
		{ReturnIDs: []uuid.UUID{both}}, {TapIDs: []uuid.UUID{both}}}, targets...))
	if err := b.castPaid(vortex, splices, []game.SplicePayment{
		{ReturnIDs: []uuid.UUID{other}}, {TapIDs: []uuid.UUID{both}}}, targets...); err != nil {
		t.Fatalf("two cards paid by two creatures: %v", err)
	}
	if inBattlefield(b.g, other) {
		t.Error("the Merfolk was not returned")
	}
	if c, _ := battlefieldCard(b.g, both); !c.Tapped {
		t.Error("the Azorius Knight was not tapped")
	}
}

// Payments are aligned with the cards: a payment for a card nothing was
// spliced for is refused, and a missing payment pays nothing (so a card
// with a cost is refused, one without is not).
func TestSplicePaymentsAreAlignedWithTheSplices(t *testing.T) {
	b := newSpliceBoard(t)
	m1, m2 := b.land(b.me.ID, "Mountain"), b.land(b.me.ID, "Mountain")
	vortex, ray := b.vortex(), b.ray()
	torrent := b.hand("Torrent of Stone", "Instant — Arcane", "{3}{R}", torrentOfStoneOracle, "R")
	pay := game.SplicePayment{SacrificeIDs: []uuid.UUID{m1, m2}}
	opp := game.TargetRef{Kind: game.TargetPlayer, ID: b.opp.ID}

	// Glacial Ray's cost is mana: a payment for it is a client bug.
	badAnnouncement(t, "a payment for a mana splice", b.castPaid(vortex, []uuid.UUID{ray},
		[]game.SplicePayment{pay}, spCardRef(b.bear), splicedRef(1, opp)))
	// A payment list longer than the splices.
	badAnnouncement(t, "more payments than splices", b.castPaid(vortex, []uuid.UUID{torrent},
		[]game.SplicePayment{pay, {}}, spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))))
	// The payment of the second card sits at index 1.
	b.me.ManaPool.EmptyPool()
	if err := b.g.AddManaForEffect(b.me.ID, uuid.Nil, "{C}{U}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	err := b.g.CastSpell(b.me.ID, vortex, game.CastSpellParams{
		Strict: true, Splices: []uuid.UUID{ray, torrent},
		SplicePayments: []game.SplicePayment{{}, pay},
		Targets:        []game.TargetRef{spCardRef(b.bear), splicedRef(1, opp), splicedRef(2, spCardRef(b.wolf))},
	})
	if err != nil {
		t.Fatalf("a mana splice then a sacrifice splice: %v", err)
	}
	if inBattlefield(b.g, m1) || inBattlefield(b.g, m2) {
		t.Error("the second card's payment was not read at its own index")
	}
}

// Register refuses a non-mana splice cost the engine cannot pay as printed.
func TestRegisterRefusesAnUnbuildableSpliceCost(t *testing.T) {
	body := func(*game.StackItem, *Context) error { return nil }
	good := SpliceSacrifice(2, "two Mountains", Land(), Subtype("Mountain"))
	for _, tc := range []struct {
		name string
		sa   *game.SpliceAbility
	}{
		{"mana and a cost", &game.SpliceAbility{Onto: game.SpliceOntoArcane, ManaCost: "{1}", Cost: good}},
		{"an empty cost", &game.SpliceAbility{Onto: game.SpliceOntoArcane, Cost: &game.SpliceCost{Label: "Nothing"}}},
		{"no label", &game.SpliceAbility{Onto: game.SpliceOntoArcane, Cost: &game.SpliceCost{Sacrifice: good.Sacrifice}}},
		{"two components", &game.SpliceAbility{Onto: game.SpliceOntoArcane, Cost: &game.SpliceCost{
			Label: "Both", Sacrifice: good.Sacrifice, OpponentGainsLife: 1}}},
		{"a variable sacrifice", &game.SpliceAbility{Onto: game.SpliceOntoArcane, Cost: &game.SpliceCost{
			Label: "Sacrifice X lands", Sacrifice: func() *game.TargetSpec { s := good.Sacrifice; s.CountFromX = true; return s }()}}},
		{"exile from a hand", &game.SpliceAbility{Onto: game.SpliceOntoArcane, Cost: &game.SpliceCost{
			Label: "Exile a card from your hand", ExileCards: ExileCardsFromHand(1, "a card", nil)}}},
	} {
		spec := Spec{Splice: tc.sa, OnResolve: body, OracleID: "splice-cost-test-" + tc.name, Name: "Splice Cost Test " + tc.name}
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted it", tc.name)
				}
			}()
			Register(spec)
		}()
	}
}

// A commander a splice cost returns to a hand is asked about first, as an
// alternative cost's return is (CR 903.9b, ADR 0115): the cast waits on its
// owner, nothing is paid, and either answer then completes the announcement.
func TestVeilOfSecrecyReturningACommanderAsksItsOwnerFirst(t *testing.T) {
	for _, toCommandZone := range []bool{true, false} {
		b := newSpliceBoard(t)
		cmdr := pushBattlefieldCardWithTimestamp(b.g, game.Card{
			InstanceID: uuid.New(), Name: "Blue Commander", TypeLine: "Legendary Creature — Merfolk",
			Power: 2, Toughness: 2, Colors: []string{"U"}, IsCommander: true, Owner: b.me.ID, Controller: b.me.ID,
		})
		vortex := b.vortex()
		veil := b.hand("Veil of Secrecy", "Instant — Arcane", "{1}{U}", veilOfSecrecyOracle, "U")
		if err := b.castPaid(vortex, []uuid.UUID{veil}, []game.SplicePayment{{ReturnIDs: []uuid.UUID{cmdr}}},
			spCardRef(b.bear), splicedRef(1, spCardRef(b.wolf))); err != nil {
			t.Fatalf("cast: %v", err)
		}
		if !inBattlefield(b.g, cmdr) || len(b.g.Stack.Cards) != 0 {
			t.Fatalf("toCommandZone=%v: the cast paid before the owner answered", toCommandZone)
		}
		c := latestChoiceOfKindFor(b.g, game.PendingChoiceOptionalReplacement, b.me.ID)
		if c == nil {
			t.Fatalf("toCommandZone=%v: the owner was not asked", toCommandZone)
		}
		if err := b.g.ResolveOptionalReplacement(c.ID, b.me.ID, toCommandZone); err != nil {
			t.Fatalf("answer: %v", err)
		}
		if inBattlefield(b.g, cmdr) {
			t.Errorf("toCommandZone=%v: the commander is still on the battlefield", toCommandZone)
		}
		if toCommandZone != b.me.Command.Contains(cmdr) || toCommandZone == spInHand(b.me, cmdr) {
			t.Errorf("toCommandZone=%v: command zone %v, hand %v", toCommandZone, b.me.Command.Contains(cmdr), spInHand(b.me, cmdr))
		}
		if len(b.g.Stack.Cards) != 1 {
			t.Errorf("toCommandZone=%v: the announcement was not made after the answer", toCommandZone)
		}
	}
}
