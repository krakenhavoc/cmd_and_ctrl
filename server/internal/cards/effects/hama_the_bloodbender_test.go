package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hama_the_bloodbender_test.go — #3989: a granted cast paid by
// waterbending {X}, X the card's mana value (CR 701.67a, 118.9).

const hamaOracle = "ef80e174-78e8-412d-a6fe-89c7af746669"

type hamaTable struct {
	g        *game.Game
	me, opp  *game.Player
	hama     uuid.UUID
	rite     uuid.UUID // a sorcery, {2}{U}: mana value 3
	bear     uuid.UUID // a creature card in the same graveyard
	helpers  []uuid.UUID
	libStart int
}

// hamaEnter seats Hama for the active player in their main phase, with
// four creatures of theirs to tap, resolves the enters trigger at the
// opponent and returns the pick prompt still open.
func hamaEnter(t *testing.T) *hamaTable {
	t.Helper()
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	h := &hamaTable{g: g, me: g.Seats[seat], opp: g.Seats[(seat+1)%len(g.Seats)]}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	h.rite, h.bear = uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		h.opp.Graveyard.PushTop(game.Card{InstanceID: h.rite, Name: "Their Rite", TypeLine: "Sorcery",
			ManaCost: "{2}{U}", Owner: h.opp.ID, Controller: h.opp.ID})
		h.opp.Graveyard.PushTop(game.Card{InstanceID: h.bear, Name: "Their Bear", TypeLine: "Creature — Bear",
			ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: h.opp.ID, Controller: h.opp.ID})
	})
	for _, name := range []string{"Helper A", "Helper B", "Helper C", "Helper D"} {
		h.helpers = append(h.helpers, b12Creature(g, h.me.ID, name, "Creature — Human", 1, 1))
	}
	h.libStart = h.opp.Library.Size()
	h.hama = b12Push(g, h.me.ID, "Hama, the Bloodbender", "Legendary Creature — Human Warlock", hamaOracle, 3, 3)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, Actor: h.me.ID, CardID: h.hama}) })
	b04WaitForPick(t, g, h.me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, h.me.ID).ID, h.me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: h.opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	return h
}

// hamaExiling enters Hama and exiles the sorcery.
func hamaExiling(t *testing.T) *hamaTable {
	t.Helper()
	h := hamaEnter(t)
	answerChooseCards(t, h.g, h.me.ID, h.rite)
	passPriorityAroundTable(t, h.g)
	if !h.g.Exile.Contains(h.rite) {
		t.Fatal("the chosen card was not exiled")
	}
	return h
}

func (h *hamaTable) cast(params game.CastSpellParams) error {
	// Strict: a refused payment is an error, not a shortfall waived
	// as paid on paper.
	params.FromZone, params.Strict = "exile", true
	return h.g.CastSpell(h.me.ID, h.rite, params)
}

func (h *hamaTable) untapped() int {
	n := 0
	for _, id := range h.helpers {
		if c, ok := battlefieldCard(h.g, id); ok && !c.Tapped {
			n++
		}
	}
	return n
}

func TestHamaMillsThreeAndOffersOnlyNoncreatureNonlandCards(t *testing.T) {
	h := hamaEnter(t)
	if got := h.libStart - h.opp.Library.Size(); got != 3 {
		t.Errorf("the opponent milled %d, want 3", got)
	}
	pick := chooseCardsChoiceFor(h.g, h.me.ID)
	if pick == nil {
		t.Fatal("Hama's controller was not asked to choose a card")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("pick bounds %d..%d, want up to one", pick.ChooseMin, pick.ChooseMax)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
		c, ok := h.g.LookupCardForEffect(id)
		if !ok || c.IsCreature() || c.IsLand() || c.Owner != h.opp.ID {
			t.Errorf("offered %v, which is not a noncreature, nonland card of the opponent's", id)
		}
	}
	if !offered[h.rite] || offered[h.bear] {
		t.Errorf("offered %v: want the sorcery and not the creature", pick.ChooseCards)
	}
}

func TestHamaGrantsAWaterbendCastWhileYouControlHama(t *testing.T) {
	h := hamaExiling(t)
	perm := exiledPermission(h.g, h.rite)
	if perm == nil || perm.Player != h.me.ID || !perm.CastOnly || !perm.WaterbendManaValue || perm.Cost != "" ||
		perm.Timing != game.TimingYourTurnOnly || !permissionLive(h.g, perm, h.me.ID) {
		t.Fatalf("permission = %+v: want Hama's controller to cast it by waterbending, on their turn", perm)
	}
	h.g.WithWriteLock(func() { _ = h.g.DestroyPermanentForEffect(h.hama) })
	if permissionLive(h.g, exiledPermission(h.g, h.rite), h.me.ID) {
		t.Error("still castable after Hama left the battlefield")
	}
}

// Waterbend {3} paid wholly by tapping three creatures, with no mana at
// all: the printed {2}{U} is not charged.
func TestHamaCastIsPaidByTappingCreatures(t *testing.T) {
	h := hamaExiling(t)
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:3]}); err != nil {
		t.Fatalf("cast by tapping three: %v", err)
	}
	if !h.g.Stack.Contains(h.rite) {
		t.Fatal("the spell is not on the stack")
	}
	if got := h.untapped(); got != 1 {
		t.Errorf("%d helpers untapped, want 1 (three tapped to pay)", got)
	}
}

// The same {3} paid with colourless mana: no blue is needed.
func TestHamaCastIsPaidWithGenericMana(t *testing.T) {
	h := hamaExiling(t)
	b06AddMana(h.me, "C", "C", "C")
	if err := h.cast(game.CastSpellParams{}); err != nil {
		t.Fatalf("cast with {3}: %v", err)
	}
	if !h.g.Stack.Contains(h.rite) {
		t.Fatal("the spell is not on the stack")
	}
	if h.untapped() != 4 {
		t.Error("a mana payment tapped creatures")
	}
}

// Taps and mana together; and the taps pay only the waterbend {3}, so a
// fourth tap is refused, as is a cast that can't be paid.
func TestHamaWaterbendBudgetIsTheManaValue(t *testing.T) {
	h := hamaExiling(t)
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers}); err == nil {
		t.Fatal("four taps were accepted for a waterbend {3}")
	}
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:2]}); err == nil {
		t.Fatal("two taps and no mana paid a waterbend {3}")
	}
	if h.untapped() != 4 {
		t.Fatal("a refused cast tapped creatures")
	}
	b06AddMana(h.me, "B")
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:2]}); err != nil {
		t.Fatalf("two taps and {1}: %v", err)
	}
	if got := h.untapped(); got != 2 {
		t.Errorf("%d helpers untapped, want 2", got)
	}
}

// "During your turn": not on an opponent's.
func TestHamaCastIsOnlyDuringYourTurn(t *testing.T) {
	h := hamaExiling(t)
	advanceToMainOf(t, h.g, (h.g.Turn.ActiveSeat+1)%len(h.g.Seats))
	b06AddMana(h.me, "C", "C", "C")
	if err := h.cast(game.CastSpellParams{}); err == nil {
		t.Fatal("cast on an opponent's turn")
	}
}

// Declining the pick exiles nothing and grants nothing.
func TestHamaMayExileNothing(t *testing.T) {
	h := hamaEnter(t)
	answerChooseCards(t, h.g, h.me.ID)
	passPriorityAroundTable(t, h.g)
	if h.g.Exile.Contains(h.rite) || !h.opp.Graveyard.Contains(h.rite) {
		t.Error("a declined pick moved the card")
	}
	if perm := h.g.CastPermissionOnCardByIDForEffect(h.rite); perm != nil {
		t.Errorf("a declined pick granted %+v", perm)
	}
}

// CR 118.9d: a cost increase applies to the waterbend, and the taps pay
// only the waterbend's own {3} (CR 701.67b). Under an opponent's Thalia,
// Guardian of Thraben the cast costs {4}: three taps leave {1} for mana,
// and a fourth tap is refused.
func TestHamaTaxIsPaidWithManaNotTaps(t *testing.T) {
	h := hamaExiling(t)
	b12Push(h.g, h.opp.ID, "Thalia, Guardian of Thraben", "Legendary Creature — Human Soldier", thaliaGuardianOracle, 2, 1)
	h.g.BumpLayerVersionForTest()
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers}); err == nil {
		t.Fatal("a fourth tap paid Thalia's tax")
	}
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:3]}); err == nil {
		t.Fatal("three taps and no mana paid a taxed waterbend {3}")
	}
	b06AddMana(h.me, "C")
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:3]}); err != nil {
		t.Fatalf("three taps and {1} under Thalia: %v", err)
	}
	if got := h.untapped(); got != 1 {
		t.Errorf("%d helpers untapped, want 1", got)
	}
}

// CR 118.9d: a cost reduction applies to the waterbend too. Under Goblin
// Electromancer the {3} costs {2}, and a third tap would pay for nothing.
func TestHamaReductionComesOffTheWaterbend(t *testing.T) {
	h := hamaExiling(t)
	b12Push(h.g, h.me.ID, "Goblin Electromancer", "Creature — Goblin Wizard", "81f06f84-1580-43c0-89d5-08d34541a519", 2, 2)
	h.g.BumpLayerVersionForTest()
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:3]}); err == nil {
		t.Fatal("three taps accepted for a waterbend reduced to {2}")
	}
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:2]}); err != nil {
		t.Fatalf("two taps for the reduced {2}: %v", err)
	}
	if got := h.untapped(); got != 2 {
		t.Errorf("%d helpers untapped, want 2", got)
	}
}

// CR 107.3b: the cast pays neither the card's mana cost nor an
// alternative cost that includes its X, so X is 0, and the waterbend is
// the rest of its mana value: {X}{U}{U} is a waterbend {2}.
func TestHamaCastOfAnXSpellLocksXAtZero(t *testing.T) {
	h := hamaExiling(t)
	h.g.WithWriteLock(func() {
		for i := range h.g.Exile.Cards {
			if h.g.Exile.Cards[i].InstanceID == h.rite {
				h.g.Exile.Cards[i].ManaCost = "{X}{U}{U}"
			}
		}
	})
	if err := h.cast(game.CastSpellParams{XValue: 2, TapIDs: h.helpers[:2]}); err == nil {
		t.Fatal("X announced as 2 on a cast that does not pay the mana cost")
	}
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:3]}); err == nil {
		t.Fatal("three taps accepted for a waterbend {2}")
	}
	if err := h.cast(game.CastSpellParams{TapIDs: h.helpers[:2]}); err != nil {
		t.Fatalf("two taps for a waterbend {2}: %v", err)
	}
}
