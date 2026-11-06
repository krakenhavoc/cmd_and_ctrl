package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// champion_of_the_clachan_test.go — #4687: "As an additional cost to cast
// this spell, behold a Kithkin and exile it" (CR 701.4a, CR 601.2h) and
// "When this creature leaves the battlefield, return the exiled card to
// its owner's hand" (CR 607.2q), on Champion of the Clachan.

const (
	clachanOracle   = "4a220e0c-e463-4264-9a96-658cf842b82e"
	clachanName     = "Champion of the Clachan"
	clachanTypeLine = "Creature — Kithkin Knight"
	clachanManaCost = "{3}{W}"
	kithkinTypeLine = "Creature — Kithkin Soldier"
)

func castClachan(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, clachanName, clachanTypeLine, clachanManaCost, clachanOracle, params)
}

// clachanBoard is a main phase with one Kithkin creature on the active
// seat's battlefield, one Kithkin card in its hand and a Goblin of each.
type clachanBoard struct {
	g          *game.Game
	me, opp    *game.Player
	onBoard    uuid.UUID // a Kithkin creature I control
	inHand     uuid.UUID // a Kithkin card in my hand
	goblin     uuid.UUID // a Goblin creature I control
	goblinCard uuid.UUID // a Goblin card in my hand
	theirs     uuid.UUID // a Kithkin creature an opponent controls
}

func newClachanBoard(t *testing.T) *clachanBoard {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b := &clachanBoard{g: g, me: me, opp: opp}
	b.onBoard = apaPush(g, me.ID, me.ID, game.Card{Name: "Goldmeadow Guard", TypeLine: kithkinTypeLine, Power: 2, Toughness: 2})
	b.inHand = handCard(me, "Kinsbaile Scout", kithkinTypeLine)
	b.goblin = apaPush(g, me.ID, me.ID, game.Card{Name: "Boggart", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	b.goblinCard = handCard(me, "Boggart Card", "Creature — Goblin")
	b.theirs = apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Kithkin", TypeLine: kithkinTypeLine, Power: 1, Toughness: 1})
	return b
}

// linkedTo reports whether the card `id` sits in exile linked to the
// permanent `perm` as the object it is now.
func linkedTo(g *game.Game, id, perm uuid.UUID) bool {
	linked := false
	g.ReadSnapshot(func() {
		c := findBattlefieldCardForTest(g, perm)
		if c == nil {
			return
		}
		for _, e := range g.ExiledWithForEffect(game.PermissionCardRef{ID: perm, Epoch: c.ObjectEpoch}) {
			if e.InstanceID == id {
				linked = true
			}
		}
	})
	return linked
}

// Beholding a Kithkin you control exiles it with the spell on the stack;
// the Champion resolves linked to it, pumps the other Kithkin, and when it
// dies the exiled card goes to its owner's hand, not the battlefield.
func TestChampionOfTheClachanBeholdsAPermanentAndReturnsItWhenItDies(t *testing.T) {
	b := newClachanBoard(t)
	other := apaPush(b.g, b.me.ID, b.me.ID, game.Card{Name: "Second Kithkin", TypeLine: kithkinTypeLine, Power: 1, Toughness: 1})
	champ, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: []uuid.UUID{b.onBoard}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !inExile(b.g, b.onBoard) {
		t.Fatal("the beheld Kithkin was not exiled as the cost was paid")
	}
	passPriorityAroundTable(t, b.g)
	if !b.g.Battlefield.Contains(champ) {
		t.Fatal("the Champion did not resolve")
	}
	if !linkedTo(b.g, b.onBoard, champ) {
		t.Fatal("the exiled Kithkin is not linked to the Champion (CR 607.2q)")
	}
	if got := effectivePower(t, b.g, other); got != 2 {
		t.Errorf("another Kithkin's power = %d, want 2 under the anthem", got)
	}
	if got := effectivePower(t, b.g, b.goblin); got != 1 {
		t.Errorf("a Goblin's power = %d, want its own 1", got)
	}
	b27Kill(b.g, champ)
	passPriorityAroundTable(t, b.g)
	if !b.me.Hand.Contains(b.onBoard) {
		t.Fatal("the exiled Kithkin did not return to its owner's hand when the Champion died")
	}
	if b.g.Battlefield.Contains(b.onBoard) {
		t.Error("the exiled card returns to the hand, not the battlefield")
	}
}

// A Kithkin card in hand pays it too; bouncing the Champion returns it.
func TestChampionOfTheClachanBeholdsAHandCardAndReturnsItWhenBounced(t *testing.T) {
	b := newClachanBoard(t)
	champ, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: []uuid.UUID{b.inHand}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !inExile(b.g, b.inHand) || b.me.Hand.Contains(b.inHand) {
		t.Fatal("the beheld Kithkin card was not exiled from the hand")
	}
	passPriorityAroundTable(t, b.g)
	b.g.WithWriteLock(func() { _ = b.g.BounceToHandForEffect(champ) })
	passPriorityAroundTable(t, b.g)
	if !b.me.Hand.Contains(b.inHand) {
		t.Fatal("the exiled card did not return when the Champion was bounced")
	}
}

// Exiling the Champion is leaving the battlefield too.
func TestChampionOfTheClachanReturnsTheCardWhenExiled(t *testing.T) {
	b := newClachanBoard(t)
	champ, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: []uuid.UUID{b.onBoard}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	b.g.WithWriteLock(func() { _ = b.g.ExileCardForEffect(champ) })
	passPriorityAroundTable(t, b.g)
	if !b.me.Hand.Contains(b.onBoard) {
		t.Fatal("the exiled card did not return when the Champion was exiled")
	}
	if !inExile(b.g, champ) {
		t.Error("the Champion itself stays in exile")
	}
}

// A countered Champion paid its cost: the card stays in exile and nothing
// ever returns it.
func TestChampionOfTheClachanCounteredLeavesTheCardInExile(t *testing.T) {
	b := newClachanBoard(t)
	champ, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: []uuid.UUID{b.onBoard}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	b.g.WithWriteLock(func() {
		if err := b.g.CounterTargetForEffect(champ); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	passPriorityAroundTable(t, b.g)
	if !b.me.Graveyard.Contains(champ) {
		t.Fatal("the countered Champion is not in its owner's graveyard")
	}
	if !inExile(b.g, b.onBoard) {
		t.Fatal("countering the Champion gave the beheld Kithkin back")
	}
}

// Each Champion returns only the card exiled to pay for it: the first
// Champion's death returns its own card and leaves the second's in exile.
func TestChampionOfTheClachanReturnsOnlyItsOwnCard(t *testing.T) {
	b := newClachanBoard(t)
	first, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: []uuid.UUID{b.onBoard}})
	if err != nil {
		t.Fatalf("first CastSpell: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	second, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: []uuid.UUID{b.inHand}})
	if err != nil {
		t.Fatalf("second CastSpell: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	if !linkedTo(b.g, b.inHand, second) || linkedTo(b.g, b.inHand, first) {
		t.Fatal("the second Champion's card must be linked to the second Champion alone")
	}
	b27Kill(b.g, first)
	passPriorityAroundTable(t, b.g)
	if !b.me.Hand.Contains(b.onBoard) {
		t.Fatal("the first Champion's card did not return")
	}
	if !inExile(b.g, b.inHand) {
		t.Fatal("the first Champion returned the second Champion's card")
	}
	b27Kill(b.g, second)
	passPriorityAroundTable(t, b.g)
	if !b.me.Hand.Contains(b.inHand) {
		t.Fatal("the second Champion's card did not return when it died")
	}
}

// A card that has left exile since is not pulled back from where it went.
func TestChampionOfTheClachanLeavesACardThatLeftExileAlone(t *testing.T) {
	b := newClachanBoard(t)
	champ, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: []uuid.UUID{b.onBoard}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, b.g)
	b.g.WithWriteLock(func() {
		if _, err := game.MoveCard(b.g.Exile, b.me.Graveyard, b.onBoard); err != nil {
			t.Fatalf("move out of exile: %v", err)
		}
	})
	b27Kill(b.g, champ)
	passPriorityAroundTable(t, b.g)
	if !b.me.Graveyard.Contains(b.onBoard) || b.me.Hand.Contains(b.onBoard) {
		t.Fatal("the leave trigger pulled a card that had left exile")
	}
}

// A changeling has every creature type, so it pays any Champion's behold;
// a creature of another type, an opponent's Kithkin and nothing at all do
// not.
func TestChampionOfTheClachanWhoCanPay(t *testing.T) {
	b := newClachanBoard(t)
	mimic := uuid.New()
	b.me.Hand.PushTop(game.Card{InstanceID: mimic, Name: "Mimic", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{"changeling"}, Owner: b.me.ID, Controller: b.me.ID})
	b.g.ReadSnapshot(func() {
		for _, subtype := range []string{"Kithkin", "Elemental", "Goblin", "Elf", "Merfolk"} {
			rc := &game.RevealCost{Subtype: subtype, Behold: true, Exile: true}
			found := false
			for _, id := range b.g.RevealCostOptionsForEffect(b.me.ID, uuid.Nil, rc) {
				if id == mimic {
					found = true
				}
				if id == b.theirs {
					t.Errorf("an opponent's Kithkin is offered to behold a %s", subtype)
				}
			}
			if !found {
				t.Errorf("the changeling is not offered to behold a %s", subtype)
			}
		}
	})
	for _, bad := range [][]uuid.UUID{nil, {b.goblin}, {b.goblinCard}, {b.theirs}, {b.onBoard, b.inHand}} {
		if _, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: bad}); err == nil {
			t.Fatalf("casting with reveal_ids %v was accepted", bad)
		}
		if stackLen(b.g) != 0 {
			t.Fatalf("a refused cast with %v left the stack non-empty", bad)
		}
	}
	if !b.g.Battlefield.Contains(b.onBoard) || !b.me.Hand.Contains(b.inHand) {
		t.Fatal("a refused cast moved a Kithkin")
	}
	if _, err := castClachan(t, b.g, game.CastSpellParams{RevealIDs: []uuid.UUID{mimic}}); err != nil {
		t.Fatalf("beholding the changeling: %v", err)
	}
	if !inExile(b.g, mimic) {
		t.Fatal("the changeling was not exiled")
	}
}

// Champions of the Perfect: the same cost over Elves, and its cast trigger
// draws for each creature spell its controller casts afterwards.
func TestChampionsOfThePerfectBeholdsAnElfAndDrawsOnCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := handCard(me, "Llanowar Elf Card", "Creature — Elf Druid")
	champ, err := castWithTapParams(t, g, "Champions of the Perfect", "Creature — Elf Warrior", "{3}{G}",
		"9cb7f23f-322d-4a87-b6c6-15b9d446aef3", game.CastSpellParams{RevealIDs: []uuid.UUID{elf}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !inExile(g, elf) {
		t.Fatal("the beheld Elf card was not exiled")
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if _, err := castWithTapParams(t, g, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{}); err != nil {
		t.Fatalf("casting a creature spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The Bears were put into the hand and cast (net 0); the trigger drew one.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand size after a creature spell = %d, want %d (one card drawn)", got, hand+1)
	}
	b27Kill(g, champ)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Fatal("the exiled Elf did not return when the Champions died")
	}
}

// With no Kithkin to behold the Champion can't be cast (CR 601.2h).
func TestChampionOfTheClachanNeedsAKithkin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := apaPush(g, me.ID, me.ID, game.Card{Name: "Boggart", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	champ, err := castClachan(t, g, game.CastSpellParams{RevealIDs: []uuid.UUID{goblin}})
	if err == nil {
		t.Fatal("cast with nothing to behold was accepted")
	}
	payable := true
	g.ReadSnapshot(func() {
		// The refused Champion is still in hand, and it is not a Kithkin
		// it could behold: the spell is never one of its own options.
		card := game.Card{InstanceID: champ, Name: clachanName, OracleID: clachanOracle, TypeLine: clachanTypeLine}
		payable = g.AnyAdditionalCostBranchPayableLocked(me.ID, card)
	})
	if payable {
		t.Error("the cost reads payable with no Kithkin to behold")
	}
}
