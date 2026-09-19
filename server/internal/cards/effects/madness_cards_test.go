package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// madness_cards_test.go — the three cards #657 ships: the two madness
// cards and the enabler that discards as a COST.
//
// The keyword's own contract is pinned in game/madness_test.go. What
// is proved here is that a card file declaring one string gets both
// halves of it, and that the two discard roads a real game takes —
// an effect's instruction (Faithless Looting) and an activation cost
// (Olivia's Dragoon) — both end in a castable card in exile.

const (
	fieryTemperOracle     = "f07bd49d-8e71-4d56-be2a-638514011318"
	bigGameHunterOracle   = "ab55834f-c935-4773-89c6-bec9712284eb"
	oliviasDragoonOracle  = "a5019399-91a8-4233-b16f-399718c4be9c"
	faithlessLootingOracl = "3d6fa57a-aa53-4b5c-b8af-a7612c823117"
)

// madnessAbilitiesFor counts the two halves the keyword must grow
// from a card's `Madness` string: the discard replacement and the
// exile trigger that watches a discard.
func madnessAbilitiesFor(oracle string) (replacements, triggers int) {
	for _, r := range game.CatalogReplacements(oracle) {
		for _, kind := range r.Watches {
			if kind == game.EventDiscardCard {
				replacements++
			}
		}
	}
	for _, tr := range game.CatalogTriggers(oracle) {
		if !game.TriggerWatchesFromZone(tr, game.ZoneExile) {
			continue
		}
		for _, kind := range tr.Watches {
			if kind == game.EventDiscardCard {
				triggers++
			}
		}
	}
	return replacements, triggers
}

// One string on the Spec, both abilities in the CardDef — and no card
// file wrote either. Big Game Hunter also keeps the ETB trigger it
// declares itself, which is what the append (rather than an
// overwrite) in buildDef is for.
func TestMadnessCardsGrowBothHalvesFromOneDeclaration(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		cost   string
	}{
		{"Fiery Temper", fieryTemperOracle, "{R}"},
		{"Big Game Hunter", bigGameHunterOracle, "{B}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reps, trigs := madnessAbilitiesFor(tc.oracle)
			if reps != 1 {
				t.Errorf("discard replacements: got %d, want 1 (CR 702.35a)", reps)
			}
			if trigs != 1 {
				t.Errorf("exile discard triggers: got %d, want 1 (CR 702.35a)", trigs)
			}
		})
	}
	// The Hunter's own ETB is still there beside the keyword's trigger.
	etb := 0
	for _, tr := range game.CatalogTriggers(bigGameHunterOracle) {
		for _, kind := range tr.Watches {
			if kind == game.EventETB {
				etb++
			}
		}
	}
	if etb != 1 {
		t.Errorf("Big Game Hunter's own ETB triggers: got %d, want 1", etb)
	}
}

// Fiery Temper through an EFFECT discard, end to end on the real
// cards: Faithless Looting loots it away, the madness trigger offers
// the cast, and three damage costs one red mana.
func TestFieryTemperIsLootedAwayAndCastForItsMadnessCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	temper := pushCatalogHandCard(me, "Fiery Temper", "Instant", fieryTemperOracle)
	spare := pushCatalogHandCard(me, "Spare", "Sorcery", "")

	castCatalogSpell(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracl, nil)
	passPriorityAroundTable(t, g)

	// "then discard two cards" — the Temper and the spare.
	answerChooseCards(t, g, me.ID, temper, spare)
	passPriorityAroundTable(t, g)

	if me.Graveyard.Contains(temper) {
		t.Fatal("Fiery Temper was discarded to the graveyard; madness exiles it (CR 702.35a)")
	}
	if !g.Exile.Contains(temper) {
		t.Fatal("Fiery Temper is not in exile")
	}
	if !me.Graveyard.Contains(spare) {
		t.Error("the card without madness did not reach the graveyard")
	}

	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("the madness trigger offered no cast")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	before := victim.Life
	if err := g.CastSpell(me.ID, temper, game.CastSpellParams{
		Strict:          true,
		FromZone:        "exile",
		AlternativeCost: game.AltCostKeyMadness,
		Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool after the cast: %d tokens left, want 0 — {R} is the whole price", len(me.ManaPool))
	}
	passPriorityAroundTable(t, g)
	if got := victim.Life; got != before-3 {
		t.Errorf("life after Fiery Temper: %d → %d, want %d", before, got, before-3)
	}
	if !me.Graveyard.Contains(temper) {
		t.Error("the resolved Fiery Temper is not in its owner's graveyard")
	}
}

// Big Game Hunter through a COST discard, on Olivia's Dragoon: the
// cost is paid at announce (CR 601.2h / 602.2b, one indivisible
// step), madness exiles the card anyway, and the 1/1 that comes back
// for {B} kills something four times its size.
func TestOliviasDragoonPitchesBigGameHunterAsACost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	dragoon := pushCatalogPermanent(g, me.ID, "Olivia's Dragoon",
		"Creature — Vampire Berserker", oliviasDragoonOracle, false)
	fatty := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: fatty, Name: "Fatty", TypeLine: "Creature — Giant",
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	hunter := pushCatalogHandCard(me, "Big Game Hunter",
		"Creature — Human Rebel Assassin", bigGameHunterOracle)

	if err := g.ActivateCatalogAbility(me.ID, dragoon, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{hunter},
	}); err != nil {
		t.Fatalf("activate Olivia's Dragoon: %v", err)
	}
	if me.Graveyard.Contains(hunter) {
		t.Fatal("the pitched Hunter went to the graveyard; madness exiles a COST discard too")
	}
	if !g.Exile.Contains(hunter) {
		t.Fatal("the pitched Hunter is not in exile")
	}

	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("a cost discard offered no madness cast")
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, hunter, game.CastSpellParams{
		Strict:          true,
		FromZone:        "exile",
		AlternativeCost: game.AltCostKeyMadness,
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	// The ETB is a real targeted trigger (CR 603.3d).
	var pick *game.PendingChoice
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoicePickTarget && ch.Source == hunter {
			pick = ch
		}
	}
	if pick == nil {
		t.Fatal("Big Game Hunter's ETB queued no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetCard, ID: fatty}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(fatty) {
		t.Error("the 5/5 survived Big Game Hunter's ETB")
	}
	if !g.Battlefield.Contains(hunter) {
		t.Error("Big Game Hunter is not on the battlefield")
	}
}

// The Dragoon's own ability still does what it prints, and the
// discard it takes is a cost the keyword-free card in hand pays
// ordinarily: to the graveyard.
func TestOliviasDragoonGainsFlyingForAnOrdinaryDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	dragoon := pushCatalogPermanent(g, me.ID, "Olivia's Dragoon",
		"Creature — Vampire Berserker", oliviasDragoonOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	fodder := pushCatalogHandCard(me, "Fodder", "Sorcery", "")

	if err := g.ActivateCatalogAbility(me.ID, dragoon, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate Olivia's Dragoon: %v", err)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the discarded card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, dragoon)
	if !ok {
		t.Fatal("Olivia's Dragoon left the battlefield")
	}
	if !game.HasKeyword(&c, "flying") {
		t.Error("Olivia's Dragoon did not gain flying")
	}
}
