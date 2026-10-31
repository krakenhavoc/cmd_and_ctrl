package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// baron_helmut_zemo_test.go — #4455: Baron Helmut Zemo's connive trigger
// and his boast, whose cost is "exile any number of black cards from your
// graveyard with fifteen or more black mana symbols among their mana
// costs" and whose effect copies those cards and casts up to three of the
// copies for free (CR 707.12).

const (
	baronZemoOracle    = "0fe6c042-1182-4392-8938-fee36149d74f"
	zemoNightsWhisper  = "7ffae8f8-3006-4969-a339-6d30678f87ea"
	zemoBlackFillerTxt = "Creature — Test"
)

// zemoTable is Zemo on the battlefield (not summoning sick), with a
// graveyard the tests fill.
type zemoTable struct {
	g        *game.Game
	me, opp  *game.Player
	zemo     uuid.UUID
	fillers  []uuid.UUID // three {B}{B}{B}{B} creature cards
	whispers []uuid.UUID // four Night's Whisper, {1}{B} each
}

func newZemoTable(t *testing.T) *zemoTable {
	t.Helper()
	g := newCatalogGame(t)
	x := &zemoTable{g: g, me: g.Seats[0], opp: g.Seats[1]}
	x.zemo = pushCatalogPermanent(g, x.me.ID, "Baron Helmut Zemo", "Legendary Creature — Human Noble Villain", baronZemoOracle, false)
	for range 3 {
		x.fillers = append(x.fillers, gyCard(x.me, "Black Filler", zemoBlackFillerTxt, "{B}{B}{B}{B}", ""))
	}
	for range 4 {
		x.whispers = append(x.whispers, gyCard(x.me, "Night's Whisper", "Sorcery", "{1}{B}", zemoNightsWhisper))
	}
	return x
}

func (x *zemoTable) attack(t *testing.T) {
	t.Helper()
	attackWith(t, x.g, x.opp.ID, x.zemo)
}

func (x *zemoTable) boast(ids ...uuid.UUID) error {
	return x.g.ActivateCatalogAbility(x.me.ID, x.zemo, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: ids})
}

func zemoIDs(parts ...[]uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func TestBaronHelmutZemoIsRegisteredWithItsBoastCost(t *testing.T) {
	var cost *game.ExileCost
	for _, ab := range game.CatalogActivatedAbilities(baronZemoOracle) {
		if ab.Cost.ExileCards != nil {
			cost = ab.Cost.ExileCards
			if !ab.Boast {
				t.Error("the exile ability is not a boast ability")
			}
		}
	}
	if cost == nil {
		t.Fatal("no exile cost on Zemo's activated ability")
	}
	if !cost.AnyNumber || cost.N != 0 || cost.From != game.ZoneGraveyard {
		t.Errorf("cost = %+v, want any number from the graveyard", cost)
	}
	if st := cost.SymbolTotal; st == nil || st.Color != "B" || st.AtLeast != 15 {
		t.Errorf("symbol floor = %+v, want 15 black", st)
	}
	if !cost.Matches(game.Card{ManaCost: "{B}"}) || cost.Matches(game.Card{ManaCost: "{G}"}) {
		t.Error("the cost takes black cards only")
	}
}

// --- the cost -----------------------------------------------------------

func TestBaronHelmutZemoBoastNeedsFifteenBlackSymbols(t *testing.T) {
	x := newZemoTable(t)
	x.attack(t)
	before := x.me.Graveyard.Size()

	// 12 + 1 = 13, and 12 + 2 = 14: short.
	for n := 1; n <= 2; n++ {
		err := x.boast(zemoIDs(x.fillers, x.whispers[:n])...)
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Fatalf("%d black symbols: %v, want ErrInvalidParam", 12+n, err)
		}
		if x.me.Graveyard.Size() != before {
			t.Fatalf("a refused activation exiled cards: graveyard %d -> %d", before, x.me.Graveyard.Size())
		}
	}
	if err := x.boast(); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("naming nothing: %v, want ErrInvalidParam", err)
	}
	// Exactly fifteen pays it.
	if err := x.boast(zemoIDs(x.fillers, x.whispers[:3])...); err != nil {
		t.Fatalf("fifteen black symbols: %v", err)
	}
	for _, id := range zemoIDs(x.fillers, x.whispers[:3]) {
		if !inExile(x.g, id) {
			t.Errorf("%v was named as the cost and is not exiled", id)
		}
	}
	if x.me.Graveyard.Size() != before-6 || !x.me.Graveyard.Contains(x.whispers[3]) {
		t.Errorf("graveyard %d -> %d: only the six named cards leave", before, x.me.Graveyard.Size())
	}
}

func TestBaronHelmutZemoBoastTakesOnlyBlackCardsFromYourOwnGraveyard(t *testing.T) {
	x := newZemoTable(t)
	x.attack(t)
	green := gyCard(x.me, "Green Bear", "Creature — Bear", "{1}{G}", "")
	theirs := gyCard(x.opp, "Their Black Filler", zemoBlackFillerTxt, "{B}{B}{B}{B}", "")
	all := zemoIDs(x.fillers, x.whispers)

	if err := x.boast(append(all[:len(all):len(all)], green)...); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("naming a green card: %v, want ErrInvalidParam", err)
	}
	if err := x.boast(append(all[:len(all):len(all)], theirs)...); err == nil {
		t.Error("naming a card in an opponent's graveyard paid the cost")
	}
	if err := x.boast(append(all[:len(all):len(all)], x.fillers[0])...); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("naming one card twice: %v, want ErrInvalidParam", err)
	}
	if x.me.Graveyard.Size() != 3+4+1 || x.opp.Graveyard.Size() != 1 {
		t.Fatal("a refused activation moved cards")
	}
}

// A hybrid, a Phyrexian and a {2/B} symbol each offer black, and count
// once (CR 700.5, CR 107.4).
func TestBaronHelmutZemoCountsHybridAndPhyrexianBlackSymbols(t *testing.T) {
	x := newZemoTable(t)
	x.attack(t)
	hybrid := gyCard(x.me, "Hybrid", zemoBlackFillerTxt, "{B/G}{B/G}{B/G}{B/G}{B/G}", "")
	phyrexian := gyCard(x.me, "Phyrexian", zemoBlackFillerTxt, "{B/P}{B/P}{B/P}{B/P}{B/P}", "")
	twobrid := gyCard(x.me, "Twobrid", zemoBlackFillerTxt, "{2/B}{2/B}{2/B}{2/B}{2/B}", "")

	if err := x.boast(hybrid, phyrexian); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("ten symbols: %v, want ErrInvalidParam", err)
	}
	if err := x.boast(hybrid, phyrexian, twobrid); err != nil {
		t.Fatalf("fifteen hybrid, Phyrexian and two-brid black symbols: %v", err)
	}
	for _, id := range []uuid.UUID{hybrid, phyrexian, twobrid} {
		if !inExile(x.g, id) {
			t.Errorf("%v not exiled", id)
		}
	}
}

// A symbol of another colour does not count toward the floor, even on a
// black card: {B}{G} is one black symbol.
func TestBaronHelmutZemoDoesNotCountOtherColoursSymbols(t *testing.T) {
	x := newZemoTable(t)
	x.attack(t)
	gold := gyCard(x.me, "Golgari Card", zemoBlackFillerTxt, "{B}{G}{G}{G}{G}", "")
	if err := x.boast(x.fillers[0], x.fillers[1], x.fillers[2], gold); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("12 + 1 black symbols: %v, want ErrInvalidParam (the green symbols do not count)", err)
	}
}

func TestBaronHelmutZemoBoastNeedsAnAttackAndIsOncePerTurn(t *testing.T) {
	x := newZemoTable(t)
	all := zemoIDs(x.fillers, x.whispers)
	if err := x.boast(all...); !errors.Is(err, game.ErrBoastNotAttacked) {
		t.Fatalf("before attacking: %v, want ErrBoastNotAttacked", err)
	}
	if x.me.Graveyard.Size() != 7 {
		t.Fatal("a refused boast exiled cards")
	}
	x.attack(t)
	if err := x.boast(all[:6]...); err != nil {
		t.Fatalf("boast after attacking: %v", err)
	}
	if err := x.boast(all[6]); !errors.Is(err, game.ErrBoastSpent) {
		t.Errorf("second boast: %v, want ErrBoastSpent", err)
	}
}

// --- the effect ---------------------------------------------------------

// The cost exiled the named cards (a cost: nobody can respond to it); on
// resolution a copy of each exists and three may be cast for free.
func TestBaronHelmutZemoCopiesTheExiledCardsAndCastsUpToThree(t *testing.T) {
	x := newZemoTable(t)
	x.attack(t)
	all := zemoIDs(x.fillers, x.whispers) // 3*4 + 4*1 = 16 symbols
	if err := x.boast(all...); err != nil {
		t.Fatalf("boast: %v", err)
	}
	if n := len(cardCopies(x.g)); n != 0 {
		t.Fatalf("%d copies before the ability resolved, want none", n)
	}
	passPriorityAroundTable(t, x.g)

	if n := len(cardCopies(x.g)); n != 7 {
		t.Fatalf("%d copies after resolution, want 7 (one per exiled card)", n)
	}
	for _, id := range all {
		if !inExile(x.g, id) {
			t.Errorf("the original %v left exile: copying does not move it", id)
		}
	}
	var whisperCopies []game.Card
	for _, c := range cardCopies(x.g) {
		if c.Name == "Night's Whisper" {
			whisperCopies = append(whisperCopies, c)
		}
	}
	if len(whisperCopies) != 4 {
		t.Fatalf("%d Night's Whisper copies, want 4", len(whisperCopies))
	}

	hand, life := x.me.Hand.Size(), x.me.Life
	for i := 0; i < 3; i++ {
		// Strict mana with an empty pool: only a free cast gets through.
		if err := castCopy(x.g, x.me.ID, whisperCopies[i].InstanceID, true); err != nil {
			t.Fatalf("free cast %d of a copy: %v", i+1, err)
		}
	}
	// The fourth is past the limit of three.
	if err := castCopy(x.g, x.me.ID, whisperCopies[3].InstanceID, true); err == nil {
		t.Fatal("a fourth copy was cast: the limit is up to three")
	}
	// The third cast spent the last of the three: nothing is left to offer,
	// so the resolution is no longer waiting on the caster.
	if resolutionCastFor(x.g, x.me.ID) != nil {
		t.Error("the resolution still waits on a cast after the third")
	}
	passPriorityAroundTable(t, x.g)

	if got := x.me.Hand.Size() - hand; got != 6 {
		t.Errorf("hand grew by %d, want 6 (three Night's Whispers resolved)", got)
	}
	if x.me.Life != life-6 {
		t.Errorf("life %d -> %d, want -6", life, x.me.Life)
	}
	if n := len(cardCopies(x.g)); n != 0 {
		t.Errorf("%d copies left: the cast ones resolved and the rest cease to exist", n)
	}
	for _, id := range all {
		if !inExile(x.g, id) {
			t.Errorf("original %v is not still exiled", id)
		}
	}
	// A black spell cast from exile is not cast from the hand.
	if latestChoiceOfKindFor(x.g, game.PendingChoiceChooseCards, x.me.ID) != nil {
		t.Error("Zemo connived on a spell cast from exile")
	}
}

// "Up to": casting fewer, or none, is fine.
func TestBaronHelmutZemoMayCastFewerThanThree(t *testing.T) {
	x := newZemoTable(t)
	x.attack(t)
	if err := x.boast(zemoIDs(x.fillers, x.whispers[:3])...); err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, x.g)
	whisper := copyOf(t, x.g, "Night's Whisper")
	hand := x.me.Hand.Size()
	if err := castCopy(x.g, x.me.ID, whisper.InstanceID, true); err != nil {
		t.Fatalf("cast one copy: %v", err)
	}
	declineResolutionCasts(t, x.g, x.me.ID)
	passPriorityAroundTable(t, x.g)
	if got := x.me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
	if n := len(cardCopies(x.g)); n != 0 {
		t.Errorf("%d copies left after declining the rest", n)
	}
}

// Only the controller may cast the copies.
func TestBaronHelmutZemoOpponentsCannotCastTheCopies(t *testing.T) {
	x := newZemoTable(t)
	x.attack(t)
	if err := x.boast(zemoIDs(x.fillers, x.whispers[:3])...); err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, x.g)
	whisper := copyOf(t, x.g, "Night's Whisper")
	if err := castCopy(x.g, x.opp.ID, whisper.InstanceID, false); err == nil {
		t.Error("an opponent cast Zemo's copy")
	}
}

// --- the trigger ----------------------------------------------------------

func TestBaronHelmutZemoConnivesOnABlackSpellFromYourHand(t *testing.T) {
	x := newZemoTable(t)
	g, me := x.g, x.me
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	// The connive draws this nonland card, and discarding it puts a +1/+1
	// counter on Zemo (CR 701.50a).
	top := topLibraryCard(me, "Top Spell", "Sorcery")
	spell := handCardFull(me, "Night's Whisper", "Sorcery", "{1}{B}", zemoNightsWhisper, []string{"B"})
	toMainForCost(t, g)
	floatForTest(g, me, "CB")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("Zemo did not connive on a black spell cast from the hand")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top}); err != nil {
		t.Fatalf("connive discard: %v", err)
	}
	if got := counterCount(g, x.zemo, game.CounterPlusOne); got != 1 {
		t.Errorf("Zemo has %d +1/+1 counters, want 1 (he discarded a nonland card)", got)
	}
}

func TestBaronHelmutZemoDoesNotConniveOnARedSpell(t *testing.T) {
	x := newZemoTable(t)
	g, me := x.g, x.me
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	red := handCardFull(me, "Red Bear", "Creature — Bear", "{1}{R}", "", []string{"R"})
	toMainForCost(t, g)
	floatForTest(g, me, "CR")
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast red: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("Zemo connived on a red spell")
	}
	if got := counterCount(g, x.zemo, game.CounterPlusOne); got != 0 {
		t.Errorf("Zemo has %d +1/+1 counters, want 0", got)
	}
}
