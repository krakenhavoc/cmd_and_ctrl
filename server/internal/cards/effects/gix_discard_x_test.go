package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gix_discard_x_test.go — #2527: "{4}{B}{B}{B}, Discard X cards: Exile
// the top X cards of target opponent's library. You may play lands and
// cast spells from among cards exiled this way without paying their
// mana costs." The variable-count discard cost (game.DiscardCost.
// CountFromX, ADR 0113's 2026-10-07 amendment) and the free play it
// pays for.

const gixYawgmothPraetorOracle = "928d977e-cff0-4e0e-83bb-16d73a754f35"

// gixTable is a table with Gix on seat 0's battlefield, seven black
// mana in the pool, three cards in seat 0's hand and a three-card
// opposing library (top first: a creature, a land, a spell).
func gixTable(t *testing.T) (g *game.Game, me, opp *game.Player, gix uuid.UUID, hand []uuid.UUID, top []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	gix = pushCatalogPermanent(g, me.ID, "Gix, Yawgmoth Praetor", "Legendary Creature — Phyrexian Praetor", gixYawgmothPraetorOracle, false)
	for i := 0; i < 7; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	}
	for _, n := range []string{"Fodder A", "Fodder B", "Fodder C"} {
		hand = append(hand, ccHandCard(me, n, "Sorcery", "{2}"))
	}
	opp.Library.Cards = nil
	// Pushed bottom-up: the LAST pushed is the top of the library.
	for _, c := range []game.Card{
		{Name: "Their Spell", TypeLine: "Sorcery", ManaCost: "{5}{R}{R}"},
		{Name: "Their Island", TypeLine: "Basic Land — Island"},
		{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{6}{G}", Power: 2, Toughness: 2},
	} {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = opp.ID, opp.ID
		opp.Library.PushTop(c)
		top = append([]uuid.UUID{c.InstanceID}, top...)
	}
	return g, me, opp, gix, hand, top
}

func gixTarget(opp *game.Player) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
}

func gixInExile(g *game.Game, id uuid.UUID) bool { return g.Exile.Contains(id) }

// X = 2: two cards leave the hand as the cost is paid, the top two
// cards of the opponent's library go to exile, the mana is the printed
// seven whatever X is, and the third library card stays put.
func TestGixDiscardsXAndExilesTheTopX(t *testing.T) {
	g, me, opp, gix, hand, top := gixTable(t)
	if err := g.ActivateCatalogAbility(me.ID, gix, 0, game.ActivateAbilityParams{
		XValue:     2,
		DiscardIDs: hand[:2],
		Targets:    gixTarget(opp),
	}); err != nil {
		t.Fatalf("activate at X=2: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %d, want all seven spent — X buys cards, not mana", len(me.ManaPool))
	}
	for _, id := range hand[:2] {
		if me.Hand.Contains(id) || !me.Graveyard.Contains(id) {
			t.Errorf("discarded card %v is not in the graveyard", id)
		}
	}
	if !me.Hand.Contains(hand[2]) {
		t.Error("the card that was not named left the hand")
	}
	// The discards are paid with the ability on the stack, before it resolves.
	if gixInExile(g, top[0]) {
		t.Fatal("the library was exiled before the ability resolved")
	}
	passPriorityAroundTable(t, g)
	if !gixInExile(g, top[0]) || !gixInExile(g, top[1]) {
		t.Fatal("the top two cards of the opponent's library are not in exile")
	}
	if gixInExile(g, top[2]) || !opp.Library.Contains(top[2]) {
		t.Error("the third card should still be in the library")
	}
}

// The announcement and the payment have to agree, in both directions,
// and a refusal leaves everything where it was.
func TestGixDiscardCountMustEqualX(t *testing.T) {
	g, me, opp, gix, hand, _ := gixTable(t)
	handBefore := me.Hand.Size()
	for _, tc := range []struct {
		why string
		p   game.ActivateAbilityParams
	}{
		{"X=2 with one card", game.ActivateAbilityParams{XValue: 2, DiscardIDs: hand[:1]}},
		{"X=1 with two cards", game.ActivateAbilityParams{XValue: 1, DiscardIDs: hand[:2]}},
		{"X=2 naming one card twice", game.ActivateAbilityParams{XValue: 2, DiscardIDs: []uuid.UUID{hand[0], hand[0]}}},
		{"X=1 naming a card not in hand", game.ActivateAbilityParams{XValue: 1, DiscardIDs: []uuid.UUID{uuid.New()}}},
		{"X=1 with no cards", game.ActivateAbilityParams{XValue: 1}},
		{"X=-1", game.ActivateAbilityParams{XValue: -1}},
	} {
		tc.p.Targets = gixTarget(opp)
		err := g.ActivateCatalogAbility(me.ID, gix, 0, tc.p)
		if err == nil {
			t.Errorf("%s: activation accepted", tc.why)
			continue
		}
		if len(me.ManaPool) != 7 || me.Hand.Size() != handBefore {
			t.Fatalf("%s: a refused activation spent something (pool %d, hand %d)", tc.why, len(me.ManaPool), me.Hand.Size())
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, gix, 0, game.ActivateAbilityParams{
		XValue: 1, DiscardIDs: []uuid.UUID{opp.Library.Cards[0].InstanceID}, Targets: gixTarget(opp),
	}); !errors.Is(err, game.ErrCardNotFound) && !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("discarding a library card: err = %v", err)
	}
}

// X may be zero: a legal, empty payment that exiles nothing.
func TestGixXZeroPaysNothingAndExilesNothing(t *testing.T) {
	g, me, opp, gix, _, top := gixTable(t)
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gix, 0, game.ActivateAbilityParams{
		XValue: 0, Targets: gixTarget(opp),
	}); err != nil {
		t.Fatalf("activate at X=0: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range top {
		if gixInExile(g, id) {
			t.Error("X=0 exiled a card")
		}
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand = %d, want %d", me.Hand.Size(), handBefore)
	}
}

// Only an opponent is a legal target, and a library shorter than X
// exiles what is there.
func TestGixTargetsAnOpponentAndExilesAShortLibrary(t *testing.T) {
	g, me, opp, gix, hand, top := gixTable(t)
	if err := g.ActivateCatalogAbility(me.ID, gix, 0, game.ActivateAbilityParams{
		XValue: 1, DiscardIDs: hand[:1], Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("Gix's controller is not a legal target")
	}
	opp.Library.Cards = opp.Library.Cards[2:] // only the top card is left
	if err := g.ActivateCatalogAbility(me.ID, gix, 0, game.ActivateAbilityParams{
		XValue: 3, DiscardIDs: hand, Targets: gixTarget(opp),
	}); err != nil {
		t.Fatalf("activate at X=3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !gixInExile(g, top[0]) || opp.Library.Size() != 0 {
		t.Errorf("exile has the top card = %v, library = %d, want the one card exiled and nothing left", gixInExile(g, top[0]), opp.Library.Size())
	}
}

// "Without paying their mana costs": a spell and a land from among the
// exiled cards are both playable, free (strict payment and an empty
// pool), and the land is a land play.
func TestGixFreePlayOfASpellAndALand(t *testing.T) {
	g, me, opp, gix, hand, top := gixTable(t)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, gix, 0, game.ActivateAbilityParams{
		XValue: 2, DiscardIDs: hand[:2], Targets: gixTarget(opp),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	bear, island := top[0], top[1]
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled {6}{G} creature for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, bear); !ok {
		t.Fatal("the free-cast creature did not enter")
	}
	drops := g.LandDropsRemainingFor(me.ID)
	if err := g.CastSpell(me.ID, island, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	if _, ok := battlefieldCard(g, island); !ok {
		t.Fatal("the exiled land did not enter")
	}
	if got := g.LandDropsRemainingFor(me.ID); got != drops-1 {
		t.Errorf("land drops = %d, want %d — playing a land from exile is a land play", got, drops-1)
	}
	// The grant is the activator's: another seat cannot play the card.
	if err := g.CastSpell(g.Seats[2].ID, top[2], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a card nobody exiled this way is playable")
	}
}

// The printed text has no "this turn": the permission lasts as long as
// the card stays exiled, so it is still free on a later turn.
func TestGixPermissionOutlastsTheTurn(t *testing.T) {
	g, me, opp, gix, hand, top := gixTable(t)
	if err := g.ActivateCatalogAbility(me.ID, gix, 0, game.ActivateAbilityParams{
		XValue: 1, DiscardIDs: hand[:1], Targets: gixTarget(opp),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		g.Turn.Seq++
		me.TurnsBegun += 2
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, top[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("a later turn, the exiled card should still be free to cast: %v", err)
	}
}

// Gix's other half still works beside the new one.
func TestGixIsCatalogedFull(t *testing.T) {
	spec, ok := Lookup(gixYawgmothPraetorOracle)
	if !ok {
		t.Fatal("Gix is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness = %v, caveats = %v, want Full with none", spec.Completeness, spec.Caveats)
	}
	if !spec.XMatters {
		t.Error("Gix's whole effect is X, so XMatters keeps the bot off the X=0 no-op")
	}
	if len(spec.Activated) != 1 || !spec.Activated[0].Cost.DemandsX() {
		t.Error("the ability should demand an announced X")
	}
}

// --- the registry refuses the shapes that cannot work ------------------

func TestRegisterRejectsDiscardXMisuse(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	cases := []struct {
		name string
		spec Spec
		want string
	}{
		{"{X} in the mana cost beside it", Spec{OracleID: "dx-guard-two-x", Name: "dx-guard-two-x", Activated: []ActivatedAbility{{
			Label: "x", Cost: Plus(ManaCost("{X}{B}"), DiscardX("X cards")), Effect: noop,
		}}}, "one announced X cannot pay both"},
		{"sacrifice X beside it", Spec{OracleID: "dx-guard-sac-x", Name: "dx-guard-sac-x", Activated: []ActivatedAbility{{
			Label: "x", Cost: Plus(SacrificeX("X creatures", Creature()), DiscardX("X cards")), Effect: noop,
		}}}, "one announced X cannot pay both"},
		{"another card out of the hand beside it", Spec{OracleID: "dx-guard-hand", Name: "dx-guard-hand", Activated: []ActivatedAbility{{
			Label: "x", Cost: Plus(PutACardFromHandOnTop(), DiscardX("X cards")), Effect: noop,
		}}}, "another card from the hand"},
		{"a mana ability", Spec{OracleID: "dx-guard-mana", Name: "dx-guard-mana", ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{DiscardCards: DiscardX("X cards").DiscardCards}, Produced: "{C}",
		}}}, "announces no X"},
		{"a count as well", Spec{OracleID: "dx-guard-count", Name: "dx-guard-count", Activated: []ActivatedAbility{{
			Label: "x", Cost: game.AbilityCost{DiscardCards: &game.DiscardCost{N: 2, CountFromX: true, Label: "x"}}, Effect: noop,
		}}}, "also declares a count"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectRegisterPanic(t, tc.want, func() { Register(tc.spec) })
		})
	}
}
