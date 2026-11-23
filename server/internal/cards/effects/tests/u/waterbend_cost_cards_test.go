package testsu

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aang's "Waterbend {8}" paid with no mana at all: seven helpers and
// Aang himself — summoning sick, and still a legal waterbender,
// because tapping to pay a waterbend cost is not the {T} symbol
// (CR 302.6), and the cost prints no {T} of its own.
func TestAangWaterbendsEightByTappingTheTeamAndHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := effectstest.AangSwiftSaviorCard(me.ID)
	aang.SummonedThisTurn = true
	g.Battlefield.PushTop(aang)
	helpers := []uuid.UUID{aang.InstanceID}
	for i := 0; i < 5; i++ {
		helpers = append(helpers, effectstest.PushVanillaCreature(g, me.ID, "Ally", 1, 1))
	}
	for i := 0; i < 2; i++ {
		helpers = append(helpers, effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false))
	}

	spec, _ := effects.Lookup(effects.AangSwiftSaviorOracle)
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
		if !effectstest.IsTapped(g, id) {
			t.Errorf("%v was not tapped to waterbend", id)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.BattlefieldCardFor(g, aang.InstanceID); c == nil || c.ActiveFace != 1 {
		t.Fatal("Aang did not transform")
	}
}

// Nine taps for an {8} is refused, and nothing taps — a ninth creature
// would be tapped for nothing.
func TestAangRefusesANinthWaterbender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := effectstest.AangSwiftSaviorCard(me.ID)
	g.Battlefield.PushTop(aang)
	helpers := []uuid.UUID{aang.InstanceID}
	for i := 0; i < 8; i++ {
		helpers = append(helpers, effectstest.PushVanillaCreature(g, me.ID, "Ally", 1, 1))
	}
	err := g.ActivateCatalogAbility(me.ID, aang.InstanceID, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers, Strict: true,
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	for _, id := range helpers {
		if effectstest.IsTapped(g, id) {
			t.Fatalf("a refused activation tapped %v", id)
		}
	}
}

// Katara, Water Tribe's Hope: the ETB Ally, then "Waterbend {X}" at
// X=3 paid with two taps and one mana — and every creature you control
// becomes a base 3/3 until end of turn, the tapped ones included.
func TestKataraWaterbendsXAndSetsTheTeamToXX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
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
	katara := effectstest.CastCatalogSpell(t, g, "Katara, Water Tribe's Hope", "Legendary Creature — Human Warrior Ally",
		effectstest.KataraWaterTribesHopeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if allies() != before+1 {
		t.Fatalf("Katara's ETB made %d Ally tokens, want 1", allies()-before)
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	ox := effectstest.PushVanillaCreature(g, me.ID, "Ox", 4, 4)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, katara, 0, game.ActivateAbilityParams{
		XValue: 3, WaterbendIDs: []uuid.UUID{bear, ox}, Strict: true,
	}); err != nil {
		t.Fatalf("Waterbend {X} at X=3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{katara, bear, ox} {
		if p, tt := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tt != 3 {
			t.Errorf("%v is %d/%d, want the base 3/3", id, p, tt)
		}
	}
	if !effectstest.IsTapped(g, bear) || !effectstest.IsTapped(g, ox) {
		t.Error("the waterbenders were not tapped")
	}
}

// "X can't be 0" and "Activate only during your turn".
func TestKataraRefusesXZeroAndOtherPlayersTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	katara := effectstest.PushCatalogPermanent(g, me.ID, "Katara, Water Tribe's Hope",
		"Legendary Creature — Human Warrior Ally", effectstest.KataraWaterTribesHopeOracle, false)
	theirs := effectstest.PushCatalogPermanent(g, other.ID, "Katara, Water Tribe's Hope",
		"Legendary Creature — Human Warrior Ally", effectstest.KataraWaterTribesHopeOracle, false)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unagi := effectstest.PushCatalogPermanent(g, me.ID, "The Unagi of Kyoshi Island",
		"Legendary Creature — Serpent", effectstest.TheUnagiOfKyoshiIslandOracl, false)
	var helpers []uuid.UUID
	for i := 0; i < 4; i++ {
		helpers = append(helpers, effectstest.PushVanillaCreature(g, opp.ID, "Helper", 1, 1))
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, unagi)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
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
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(unagi) {
		t.Error("the ward was paid — the removal should have resolved")
	}
	for _, id := range helpers {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("%v was not tapped to pay the ward", id)
		}
	}
}

// Declined, the ward counters the spell and the Unagi lives.
func TestUnagiWardDeclinedCountersTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unagi := effectstest.PushCatalogPermanent(g, me.ID, "The Unagi of Kyoshi Island",
		"Legendary Creature — Serpent", effectstest.TheUnagiOfKyoshiIslandOracl, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	blade := effectstest.CastAtWardedCreature(t, g, opp, unagi)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "The Unagi of Kyoshi Island",
		"Legendary Creature — Serpent", effectstest.TheUnagiOfKyoshiIslandOracl, false)
	effectstest.PushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Filler 1", TypeLine: "Land"})
	effectstest.PushLibraryCardForTest(opp, game.Card{InstanceID: uuid.New(), Name: "Filler 2", TypeLine: "Land"})
	handBefore := me.Hand.Size()

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("first draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore {
		t.Fatalf("the FIRST draw triggered: hand %d, want %d", me.Hand.Size(), handBefore)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("second draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	c := effects.WaterbendCost("{8}")
	if c.Mana != "{8}" || c.Waterbend == nil || c.Waterbend.Extra != "{8}" {
		t.Fatalf("WaterbendCost(\"{8}\") = %+v", c)
	}
	both := effects.Plus(effects.ManaCost("{1}{U}"), effects.WaterbendCost("{2}"))
	if both.Mana != "{1}{U}{2}" || both.Waterbend.Extra != "{2}" {
		t.Errorf("composite = %q / %q, want {1}{U}{2} / {2}", both.Mana, both.Waterbend.Extra)
	}
	x := effects.Plus(effects.WaterbendCost("{X}"), effects.MinX(1))
	if x.Mana != "{X}" || x.MinX != 1 || !x.DemandsX() {
		t.Errorf("Katara's cost = %+v", x)
	}
	plain := effects.Plus(effects.ManaCost("{1}"), effects.ManaCost("{2}"))
	if plain.Mana != "{2}" {
		t.Errorf("two plain mana components: %q, want the later one as before", plain.Mana)
	}
	tap := &game.TapOthersCost{Count: 1, Filter: effects.TargetCreature("a creature"), Label: "Tap a creature"}
	if got := effects.Plus(effects.TapCost(), game.AbilityCost{TapOthers: tap}); got.TapOthers != tap || !got.Tap {
		t.Errorf("Plus dropped the tap-another component: %+v", got)
	}
}

// Register refuses, at boot, a waterbend clause that could not be paid
// as printed.
func TestRegisterRefusesAMalformedAbilityWaterbend(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	bigger := effects.WaterbendCost("{4}")
	bigger.Mana = "{2}"
	noPool := effects.WaterbendCost("{2}")
	noPool.Waterbend.Spec = nil
	xWithoutX := effects.WaterbendCost("{X}")
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
		msg := effectstest.RegisterPanics(effects.Spec{
			OracleID: "waterbend-malformed-" + tc.name,
			Name:     "Malformed Waterbender",
			Activated: []effects.ActivatedAbility{{
				Label: "Waterbend: nothing", Cost: tc.cost, Effect: noop,
			}},
		})
		if msg == "" || !strings.Contains(msg, tc.want) {
			t.Errorf("%s: panic = %q, want one mentioning %q", tc.name, msg, tc.want)
		}
	}
}
