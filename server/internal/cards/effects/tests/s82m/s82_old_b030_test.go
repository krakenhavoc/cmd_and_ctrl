package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 30 (S82, #5501): Modern Horizons 2 — Gilt-Blade Prowler,
// Ethersworn Sphinx, Flame Blitz, Lucid Dreams, Step Through, Solitary
// Confinement and Recalibrate.

const (
	oldB030Prowler    = "63e86054-5e9c-48fe-92a1-b71aeb6ee57f"
	oldB030Sphinx     = "71366469-04ea-4526-bf24-7abc02fdc5ab"
	oldB030FlameBlitz = "817eb91a-e3b4-40c3-9a6e-e4d0387e0f4f"
	oldB030Lucid      = "bd716715-15a4-4356-ad7e-c642fafbfd18"
	oldB030StepThru   = "c4fa89c7-f00d-42ef-bd67-407a288cd054"
	oldB030Solitary   = "d41ab41d-07c8-4f6d-be5e-7aec9f5f6365"
	oldB030Recal      = "e6331b3a-81a8-48e8-8a28-237d381f9062"
)

func TestOldB030CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB030Prowler:    "Gilt-Blade Prowler",
		oldB030Sphinx:     "Ethersworn Sphinx",
		oldB030FlameBlitz: "Flame Blitz",
		oldB030Lucid:      "Lucid Dreams",
		oldB030StepThru:   "Step Through",
		oldB030Solitary:   "Solitary Confinement",
		oldB030Recal:      "Recalibrate",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB030Library(p *game.Player, n int) {
	p.Library.Cards = nil
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// oldB030Price prices a card for `caster` cast from hand, through the cost
// modifiers on the table.
func oldB030Price(t *testing.T, g *game.Game, caster uuid.UUID, card game.Card) string {
	t.Helper()
	card.Owner, card.Controller = caster, caster
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	var out game.ParsedCost
	g.WithWriteLock(func() {
		var err error
		out, err = g.ApplyCostModifiersForEffect(base, game.CostQuery{Game: g, Card: card, Controller: caster, FromZone: game.ZoneHand})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
	})
	return out.String()
}

// --- Gilt-Blade Prowler ----------------------------------------------------

func TestOldB030ProwlerDrawsOnlyAfterADiscardThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB030Library(me, 10)
	prowler := effectstest.PushCatalogPermanent(g, me.ID, "Gilt-Blade Prowler", "Creature — Human Rogue", oldB030Prowler, false)
	if err := effectstest.HobB4Activate(t, g, me.ID, prowler, 0, "{C}", nil); err == nil {
		t.Fatal("activated before discarding a card this turn")
	}
	effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	effectstest.B30DiscardOne(t, g, me.ID)
	hand, life := me.Hand.Size(), me.Life
	if err := effectstest.HobB4Activate(t, g, me.ID, prowler, 0, "{C}", nil); err != nil {
		t.Fatalf("activate after a discard: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d after paying 1", me.Life, life-1)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, prowler); !tapped {
		t.Error("the Prowler was not tapped as a cost")
	}
}

// --- Ethersworn Sphinx -----------------------------------------------------

func TestOldB030SphinxHasAffinityForArtifactsAndCascades(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 7; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	}
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Their Rock", "Artifact", "", false)
	sphinx := game.Card{InstanceID: uuid.New(), Name: "Ethersworn Sphinx", TypeLine: "Artifact Creature — Sphinx", ManaCost: "{7}{W}{U}", OracleID: oldB030Sphinx}
	if got := oldB030Price(t, g, me.ID, sphinx); got != "{W}{U}" {
		t.Fatalf("with seven artifacts it costs %q, want {W}{U}", got)
	}
	effectstest.FloatForTest(g, me, "WU")
	id := effectstest.CastFromHandForTest(t, g, me, "Ethersworn Sphinx", "Artifact Creature — Sphinx", "{7}{W}{U}", oldB030Sphinx, game.CastSpellParams{Strict: true})
	if effectstest.TriggersOnStackFrom(g, id) != 1 {
		t.Errorf("%d triggers on the stack from the Sphinx, want its cascade", effectstest.TriggersOnStackFrom(g, id))
	}
}

// --- Flame Blitz -----------------------------------------------------------

func TestOldB030FlameBlitzDamagesEveryPlaneswalkerAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Flame Blitz", "Enchantment", oldB030FlameBlitz, false)
	mine := effectstest.GuffbWalker(g, me.ID, "My Walker", "Legendary Planeswalker — Test", "", 7)
	theirs := effectstest.GuffbWalker(g, opp.ID, "Their Walker", "Legendary Planeswalker — Test", "", 4)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, mine); !ok || c.Counters[game.CounterLoyalty] != 2 {
		t.Errorf("my planeswalker: on battlefield %v, loyalty %d, want 2", ok, c.Counters[game.CounterLoyalty])
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("the opposing 4-loyalty planeswalker survived 5 damage")
	}
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.DamageMarked != 0 {
		t.Error("a creature was damaged by Flame Blitz")
	}
}

func TestOldB030FlameBlitzCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB030Library(me, 5)
	id := effectstest.PushCatalogHandCard(me, "Flame Blitz", "Enchantment", oldB030FlameBlitz)
	hand := me.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("cycling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) || me.Hand.Size() != hand {
		t.Errorf("in graveyard %v, hand %d (was %d): want it discarded and one card drawn", me.Graveyard.Contains(id), me.Hand.Size(), hand)
	}
}

// --- Lucid Dreams ----------------------------------------------------------

func TestOldB030LucidDreamsDrawsOnePerCardTypeInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB030Library(me, 20)
	effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}")
	effectstest.PushGraveyardPermanent(me, "Other Bear", "Creature — Bear", "{1}")
	effectstest.PushGraveyardPermanent(me, "Golem", "Artifact Creature — Golem", "{3}")
	effectstest.PushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Lucid Dreams", "Sorcery", oldB030Lucid, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// creature, artifact and instant: three types, the Dreams themselves not counted.
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
}

// --- Step Through ----------------------------------------------------------

func TestOldB030StepThroughReturnsTwoCreaturesAndNeedsBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Step Through", "Sorcery", oldB030StepThru, effectstest.CardTargets(a)); err == nil {
		t.Fatal("cast with one target")
	}
	effectstest.CastCatalogSpell(t, g, "Step Through", "Sorcery", oldB030StepThru, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Errorf("returned to their owners' hands: A %v, B %v", opp.Hand.Contains(a), me.Hand.Contains(b))
	}
}

func TestOldB030StepThroughWizardcyclingFindsAWizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Step Through", "Sorcery", oldB030StepThru)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Apprentice", TypeLine: "Creature — Human Wizard"},
	)
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("wizardcycling: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || me.Hand.Contains(ids[0]) || !me.Graveyard.Contains(id) {
		t.Errorf("Wizard in hand %v, Bear in hand %v, Step Through in graveyard %v", me.Hand.Contains(ids[1]), me.Hand.Contains(ids[0]), me.Graveyard.Contains(id))
	}
}

// --- Recalibrate -----------------------------------------------------------

func TestOldB030RecalibrateDrawsOnlyIfYouDiscardedThisTurn(t *testing.T) {
	for _, discarded := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		oldB030Library(me, 10)
		bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
		if discarded {
			effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
			effectstest.B30DiscardOne(t, g, me.ID)
		}
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Recalibrate", "Instant", oldB030Recal, effectstest.CardTargets(bear))
		effectstest.PassPriorityAroundTable(t, g)
		if !opp.Hand.Contains(bear) {
			t.Fatalf("discarded=%v: the creature was not returned", discarded)
		}
		want := hand - 1 // the cast spell leaves the hand
		if discarded {
			want++
		}
		// CastCatalogSpell pushes the card to hand first, so the cast takes it back out.
		if got := me.Hand.Size(); got != want+1 {
			t.Errorf("discarded=%v: hand %d, want %d", discarded, got, want+1)
		}
	}
}

// --- Solitary Confinement --------------------------------------------------

func TestOldB030SolitaryConfinementGivesShroudAndPreventsAllDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Solitary Confinement", "Enchantment", oldB030Solitary, false)
	if !effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, me), effects.KeywordShroud) {
		t.Error("no shroud on the controller")
	}
	if effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, opp), effects.KeywordShroud) {
		t.Error("an opponent has shroud")
	}
	life, oppLife := me.Life, opp.Life
	effectstest.Pr7Hit(t, g, uuid.New(), me.ID, 5)
	effectstest.Pr7Hit(t, g, uuid.New(), opp.ID, 3)
	if me.Life != life {
		t.Errorf("life %d, want %d: all damage to you is prevented", me.Life, life)
	}
	if opp.Life != oppLife-3 {
		t.Errorf("an opponent's life %d, want %d: their damage is not prevented", opp.Life, oppLife-3)
	}
}

func TestOldB030SolitaryConfinementSkipsYourDrawStepAndAsksForADiscardAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	grove := effectstest.PushCatalogPermanent(g, owner.ID, "Solitary Confinement", "Enchantment", oldB030Solitary, false)
	oldB030Library(owner, 10)
	for i := 0; i < 3; i++ {
		effectstest.HandCardFull(owner, "Card", "Sorcery", "{1}", "", nil)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, owner.ID, owner.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	hand := owner.Hand.Size()
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if got := owner.Hand.Size(); got != hand {
		t.Errorf("hand went from %d to %d through the draw step, want no draw", hand, got)
	}
	if _, ok := effectstest.BattlefieldCard(g, grove); !ok {
		t.Error("Solitary Confinement was sacrificed after a discard")
	}
}

func TestOldB030SolitaryConfinementIsSacrificedWithoutADiscard(t *testing.T) {
	for _, emptyHand := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		owner := g.Seats[1]
		grove := effectstest.PushCatalogPermanent(g, owner.ID, "Solitary Confinement", "Enchantment", oldB030Solitary, false)
		oldB030Library(owner, 10)
		owner.Hand.Cards = nil
		if !emptyHand {
			effectstest.HandCardFull(owner, "Card", "Sorcery", "{1}", "", nil)
		}
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if !emptyHand {
			c := effectstest.DiscardChoiceFor(g, owner.ID)
			if c == nil {
				t.Fatal("no discard prompt with a card in hand")
			}
			if err := g.ResolveChooseCards(c.ID, owner.ID, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		if _, ok := effectstest.BattlefieldCard(g, grove); ok {
			t.Errorf("emptyHand=%v: Solitary Confinement stayed without a discard", emptyHand)
		}
	}
}
