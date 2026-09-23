package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// waterbend_cost_cards_test.go — the proof cards for #1310 (waterbend
// as an activated ability's cost) and #1311 (a ward whose cost is a
// waterbend cost). The engine shapes are pinned in
// game/waterbend_cost_test.go; these pin each card's printed text.

const (
	kataraWaterTribesHopeOracle = "234fb291-0b62-4092-9071-81311c71bd53"
	theUnagiOfKyoshiIslandOracl = "be922360-ee9f-4cfb-bf10-d481c5de5cb1"
)

func isTapped(g *game.Game, id uuid.UUID) bool {
	c := battlefieldCardFor(g, id)
	return c != nil && c.Tapped
}

// Aang's "Waterbend {8}" paid with no mana at all: seven helpers and
// Aang himself — summoning sick, and still a legal waterbender,
// because tapping to pay a waterbend cost is not the {T} symbol
// (CR 302.6), and the cost prints no {T} of its own.
func TestAangWaterbendsEightByTappingTheTeamAndHimself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := aangSwiftSaviorCard(me.ID)
	aang.SummonedThisTurn = true
	g.Battlefield.PushTop(aang)
	helpers := []uuid.UUID{aang.InstanceID}
	for i := 0; i < 5; i++ {
		helpers = append(helpers, pushVanillaCreature(g, me.ID, "Ally", 1, 1))
	}
	for i := 0; i < 2; i++ {
		helpers = append(helpers, pushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false))
	}

	spec, _ := Lookup(aangSwiftSaviorOracle)
	if wb := spec.Activated[0].Cost.Waterbend; wb == nil || wb.Extra != "{8}" {
		t.Fatalf("Aang's transform cost has no waterbend {8} clause: %+v", spec.Activated[0].Cost)
	}
	if err := g.ActivateCatalogAbility(me.ID, aang.InstanceID, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers,
		Strict:       true,
	}); err != nil {
		t.Fatalf("Waterbend {8} with eight taps and an empty pool: %v", err)
	}
	for _, id := range helpers {
		if !isTapped(g, id) {
			t.Errorf("%v was not tapped to waterbend", id)
		}
	}
	passPriorityAroundTable(t, g)
	if c := battlefieldCardFor(g, aang.InstanceID); c == nil || c.ActiveFace != 1 {
		t.Fatal("Aang did not transform")
	}
}

// Nine taps for an {8} is refused, and nothing taps — a ninth creature
// would be tapped for nothing.
func TestAangRefusesANinthWaterbender(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := aangSwiftSaviorCard(me.ID)
	g.Battlefield.PushTop(aang)
	helpers := []uuid.UUID{aang.InstanceID}
	for i := 0; i < 8; i++ {
		helpers = append(helpers, pushVanillaCreature(g, me.ID, "Ally", 1, 1))
	}
	err := g.ActivateCatalogAbility(me.ID, aang.InstanceID, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers, Strict: true,
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	for _, id := range helpers {
		if isTapped(g, id) {
			t.Fatalf("a refused activation tapped %v", id)
		}
	}
}

// Katara, Water Tribe's Hope: the ETB Ally, then "Waterbend {X}" at
// X=3 paid with two taps and one mana — and every creature you control
// becomes a base 3/3 until end of turn, the tapped ones included.
func TestKataraWaterbendsXAndSetsTheTeamToXX(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	allies := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Ally" {
				n++
			}
		}
		return n
	}
	before := allies()
	katara := castCatalogSpell(t, g, "Katara, Water Tribe's Hope", "Legendary Creature — Human Warrior Ally",
		kataraWaterTribesHopeOracle, nil)
	passPriorityAroundTable(t, g)
	if allies() != before+1 {
		t.Fatalf("Katara's ETB made %d Ally tokens, want 1", allies()-before)
	}
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	ox := pushVanillaCreature(g, me.ID, "Ox", 4, 4)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, katara, 0, game.ActivateAbilityParams{
		XValue: 3, WaterbendIDs: []uuid.UUID{bear, ox}, Strict: true,
	}); err != nil {
		t.Fatalf("Waterbend {X} at X=3: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{katara, bear, ox} {
		if p, tt := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 3 || tt != 3 {
			t.Errorf("%v is %d/%d, want the base 3/3", id, p, tt)
		}
	}
	if !isTapped(g, bear) || !isTapped(g, ox) {
		t.Error("the waterbenders were not tapped")
	}
}

// "X can't be 0" and "Activate only during your turn".
func TestKataraRefusesXZeroAndOtherPlayersTurns(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	katara := pushCatalogPermanent(g, me.ID, "Katara, Water Tribe's Hope",
		"Legendary Creature — Human Warrior Ally", kataraWaterTribesHopeOracle, false)
	theirs := pushCatalogPermanent(g, other.ID, "Katara, Water Tribe's Hope",
		"Legendary Creature — Human Warrior Ally", kataraWaterTribesHopeOracle, false)

	if err := g.ActivateCatalogAbility(me.ID, katara, 0, game.ActivateAbilityParams{XValue: 0}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("X=0: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(other.ID, theirs, 0, game.ActivateAbilityParams{XValue: 1}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Errorf("off-turn: err = %v, want ErrConditionNotMet", err)
	}
}

// The Unagi's ward—waterbend {4} paid with four taps and no mana: the
// removal resolves, and the payer's four permanents are tapped.
func TestUnagiWardPaidByWaterbending(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unagi := pushCatalogPermanent(g, me.ID, "The Unagi of Kyoshi Island",
		"Legendary Creature — Serpent", theUnagiOfKyoshiIslandOracl, false)
	var helpers []uuid.UUID
	for i := 0; i < 4; i++ {
		helpers = append(helpers, pushVanillaCreature(g, opp.ID, "Helper", 1, 1))
	}
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, unagi)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == opp.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("no ward prompt for the caster")
	}
	if tc := prompt.PayTapCost(); tc == nil || tc.Extra != "{4}" {
		t.Fatalf("the ward prompt carries no waterbend {4} clause: %+v", tc)
	}
	if err := g.ResolvePayUnlessWithTaps(prompt.ID, opp.ID, true, helpers); err != nil {
		t.Fatalf("pay with four taps: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(unagi) {
		t.Error("the ward was paid — the removal should have resolved")
	}
	for _, id := range helpers {
		if !isTapped(g, id) {
			t.Errorf("%v was not tapped to pay the ward", id)
		}
	}
}

// Declined, the ward counters the spell and the Unagi lives.
func TestUnagiWardDeclinedCountersTheSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unagi := pushCatalogPermanent(g, me.ID, "The Unagi of Kyoshi Island",
		"Legendary Creature — Serpent", theUnagiOfKyoshiIslandOracl, false)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := castAtWardedCreature(t, g, opp, unagi)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(unagi) {
		t.Error("a declined ward must counter the spell")
	}
	if !opp.Graveyard.Contains(blade) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
}

// "Whenever an opponent draws their second card each turn, you draw
// two cards."
func TestUnagiDrawsTwoOffAnOpponentsSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Unagi of Kyoshi Island",
		"Legendary Creature — Serpent", theUnagiOfKyoshiIslandOracl, false)
	pushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Filler 1", TypeLine: "Land"})
	pushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Filler 2", TypeLine: "Land"})
	handBefore := me.Hand.Size()

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("first draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Fatalf("the FIRST draw triggered: hand %d, want %d", me.Hand.Size(), handBefore)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("second draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+2 {
		t.Fatalf("hand = %d, want %d — the second draw draws TWO", me.Hand.Size(), handBefore+2)
	}
}

// WaterbendCost puts the whole cost in the mana component, and Plus
// SUMS a waterbend's mana with another mana component (CR 701.67b —
// the waterbend is part of the total) rather than letting the later
// one win, while the clause still names only the waterbend's part.
// Plus also keeps the tap-another component it used to drop (#758).
func TestWaterbendCostComposesWithPlus(t *testing.T) {
	c := WaterbendCost("{8}")
	if c.Mana != "{8}" || c.Waterbend == nil || c.Waterbend.Extra != "{8}" {
		t.Fatalf("WaterbendCost(\"{8}\") = %+v", c)
	}
	both := Plus(ManaCost("{1}{U}"), WaterbendCost("{2}"))
	if both.Mana != "{1}{U}{2}" || both.Waterbend.Extra != "{2}" {
		t.Errorf("composite = %q / %q, want {1}{U}{2} / {2}", both.Mana, both.Waterbend.Extra)
	}
	x := Plus(WaterbendCost("{X}"), MinX(1))
	if x.Mana != "{X}" || x.MinX != 1 || !x.DemandsX() {
		t.Errorf("Katara's cost = %+v", x)
	}
	plain := Plus(ManaCost("{1}"), ManaCost("{2}"))
	if plain.Mana != "{2}" {
		t.Errorf("two plain mana components: %q, want the later one as before", plain.Mana)
	}
	tap := &game.TapOthersCost{Count: 1, Filter: TargetCreature("a creature"), Label: "Tap a creature"}
	if got := Plus(TapCost(), game.AbilityCost{TapOthers: tap}); got.TapOthers != tap || !got.Tap {
		t.Errorf("Plus dropped the tap-another component: %+v", got)
	}
}

// Register refuses, at boot, a waterbend clause that could not be paid
// as printed.
func TestRegisterRefusesAMalformedAbilityWaterbend(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	bigger := WaterbendCost("{4}")
	bigger.Mana = "{2}"
	noPool := WaterbendCost("{2}")
	noPool.Waterbend.Spec = nil
	xWithoutX := WaterbendCost("{X}")
	xWithoutX.Mana = "{3}"
	for _, tc := range []struct {
		name string
		cost game.AbilityCost
		want string
	}{
		{"a clause larger than the mana", bigger, "does not contain it"},
		{"a clause with no pool", noPool, "no key, pool or cost"},
		{"an {X} clause on a fixed mana cost", xWithoutX, "does not contain it"},
	} {
		msg := registerPanics(Spec{
			OracleID: "waterbend-malformed-" + tc.name,
			Name:     "Malformed Waterbender",
			Activated: []ActivatedAbility{{
				Label: "Waterbend: nothing", Cost: tc.cost, Effect: noop,
			}},
		})
		if msg == "" || !strings.Contains(msg, tc.want) {
			t.Errorf("%s: panic = %q, want one mentioning %q", tc.name, msg, tc.want)
		}
	}
}
