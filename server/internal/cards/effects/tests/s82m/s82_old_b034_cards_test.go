package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 older sets, batch 34 (Modern Horizons 2: life gain and loss, #5501).

const (
	oldB034DiscerningTaste = "1896c513-ec48-4c8f-9acb-30bc6357fce2"
	oldB034CabalInitiate   = "2fc19fa3-c6c2-42c0-85fa-8d13e1baa145"
	oldB034Enforcer        = "3dcbbbeb-4efe-4363-bfb4-b7d7f803bafa"
	oldB034SmellFear       = "c9739382-9fe0-4d24-b58e-1e45a88491b6"
	oldB034Daggertooth     = "cc545c68-829f-4c57-94d2-cd1edf2667a7"
	oldB034Familiar        = "ce8833ea-5f58-4ab9-b33d-8972f2215574"
)

func TestOldB034CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB034DiscerningTaste: "Discerning Taste",
		oldB034CabalInitiate:   "Cabal Initiate",
		oldB034Enforcer:        "Glorious Enforcer",
		oldB034SmellFear:       "Smell Fear",
		oldB034Daggertooth:     "Urban Daggertooth",
		oldB034Familiar:        "Necromancer's Familiar",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB034Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var out game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			out = z.Kind
		}
	})
	return out
}

// oldB034Mill puts n fresh cards into the player's graveyard through real moves,
// so the layers are told.
func oldB034Mill(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		id := effectstest.PushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneLibrary, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
}

// --- Discerning Taste -------------------------------------------------------

// oldB034Top stacks the top four cards, first listed on top.
func oldB034Top(me *game.Player, cards ...game.Card) []uuid.UUID {
	me.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.Owner, c.Controller = me.ID, me.ID
		ids[i] = effectstest.PushLibraryCardForTest(me, c)
	}
	return ids
}

func TestOldB034DiscerningTasteTakesOneAndGainsTheGreatestPowerBinned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := oldB034Top(me,
		game.Card{Name: "Big", TypeLine: "Creature — Giant", Power: 5, Toughness: 5},
		game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 3, Toughness: 3},
		game.Card{Name: "Field", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Zap", TypeLine: "Instant"},
	)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Discerning Taste", "Sorcery", oldB034DiscerningTaste, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if oldB034Zone(g, ids[2]) != game.ZoneHand {
		t.Errorf("the chosen card is in %q, want the hand", oldB034Zone(g, ids[2]))
	}
	for _, i := range []int{0, 1, 3} {
		if oldB034Zone(g, ids[i]) != game.ZoneGraveyard {
			t.Errorf("card %d is in %q, want the graveyard", i, oldB034Zone(g, ids[i]))
		}
	}
	if me.Life != life+5 {
		t.Errorf("life %d, want %d (the greatest power, 5)", me.Life, life+5)
	}
}

func TestOldB034DiscerningTasteCountsOnlyCreaturesPutIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := oldB034Top(me,
		game.Card{Name: "Big", TypeLine: "Creature — Giant", Power: 5, Toughness: 5},
		game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 3, Toughness: 3},
		game.Card{Name: "Field", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Zap", TypeLine: "Instant"},
	)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Discerning Taste", "Sorcery", oldB034DiscerningTaste, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Taking the biggest creature leaves the 3-power one to be binned.
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (the creature put into the graveyard, power 3)", me.Life, life+3)
	}
}

func TestOldB034DiscerningTasteGainsNothingWithoutAGraveyardCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := oldB034Top(me,
		game.Card{Name: "Field", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Zap", TypeLine: "Instant"},
		game.Card{Name: "Rock", TypeLine: "Artifact"},
		game.Card{Name: "Big", TypeLine: "Creature — Giant", Power: 5, Toughness: 5},
	)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Discerning Taste", "Sorcery", oldB034DiscerningTaste, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[3])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// --- Cabal Initiate ---------------------------------------------------------

func TestOldB034CabalInitiateDiscardsForLifelinkAndThresholdGrowsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ci := effectstest.PushCatalogPermanent(g, me.ID, "Cabal Initiate", "Creature — Human Warlock", oldB034CabalInitiate, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, ci)
		c.Power, c.Toughness = 2, 1
	})
	if effectstest.EffectivePower(t, g, ci) != 2 || effectstest.EffectiveToughness(t, g, ci) != 1 {
		t.Fatal("the Initiate is not a 2/1 without threshold")
	}
	pitch := effectstest.PushCatalogGraveyardCard(me, "Pitched", "Sorcery", "", 0, 0)
	_ = pitch
	card := effectstest.HandSpell(me, "Pitch", "Sorcery", "")
	if effectstest.HasEffectiveKeyword(t, g, ci, "lifelink") {
		t.Fatal("the Initiate has lifelink before the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, ci, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, ci, "lifelink") {
		t.Error("the Initiate did not gain lifelink")
	}
	if oldB034Zone(g, card) != game.ZoneGraveyard {
		t.Errorf("the discarded card is in %q", oldB034Zone(g, card))
	}
	// Threshold: seven cards in the graveyard make it 3/3.
	oldB034Mill(t, g, me, 5)
	if got := effectstest.EffectivePower(t, g, ci); got != 3 {
		t.Errorf("power %d at 7 cards in the graveyard, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, ci); got != 3 {
		t.Errorf("toughness %d at 7 cards in the graveyard, want 3", got)
	}
}

func TestOldB034CabalInitiateHasNoThresholdBelowSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ci := effectstest.PushCatalogPermanent(g, me.ID, "Cabal Initiate", "Creature — Human Warlock", oldB034CabalInitiate, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, ci)
		c.Power, c.Toughness = 2, 1
	})
	oldB034Mill(t, g, me, 6)
	if got := effectstest.EffectivePower(t, g, ci); got != 2 {
		t.Errorf("power %d with six cards, want 2", got)
	}
}

// --- Glorious Enforcer ---------------------------------------------------------

func TestOldB034EnforcerGainsDoubleStrikeOnlyWhenAheadOnLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	enf := effectstest.PushCatalogPermanent(g, me.ID, "Glorious Enforcer", "Creature — Angel", oldB034Enforcer, false)
	me.Life, opp.Life = 30, 40
	for _, p := range g.Seats[2:] {
		p.Life = 40
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, enf, "double strike") {
		t.Fatal("double strike while behind on life")
	}
	// Next turn of the same table: ahead of one opponent is enough.
	me.Life, opp.Life = 41, 40
	for _, p := range g.Seats[2:] {
		p.Life = 50
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, enf, "double strike") {
		t.Error("no double strike with more life than one opponent (every player's combat counts)")
	}
}

// --- Smell Fear ------------------------------------------------------------------

func TestOldB034SmellFearProliferatesThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 3, 3)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	effectstest.PutCounters(t, g, mine, game.CounterPlusOne, 1)
	effectstest.CastCatalogSpell(t, g, "Smell Fear", "Sorcery", oldB034SmellFear, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2 (proliferated)", got)
	}
	if oldB034Zone(g, theirs) != game.ZoneGraveyard {
		t.Errorf("the opposing creature is in %q, want the graveyard (4 damage)", oldB034Zone(g, theirs))
	}
	if effectstest.DamageMarkedOn(g, mine) != 2 {
		t.Errorf("my creature has %d damage, want 2", effectstest.DamageMarkedOn(g, mine))
	}
}

func TestOldB034SmellFearWithNoSecondTargetOnlyProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 3, 3)
	effectstest.PutCounters(t, g, mine, game.CounterPlusOne, 1)
	effectstest.CastCatalogSpell(t, g, "Smell Fear", "Sorcery", oldB034SmellFear, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("my creature took damage with nothing to fight")
	}
}

func TestOldB034SmellFearRefusesYourOwnCreatureAsTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 3, 3)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 3, 3)
	err := effectstest.CastCatalogSpellErr(t, g, "Smell Fear", "Sorcery", oldB034SmellFear, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	if err == nil {
		t.Error("the second target must be a creature you don't control")
	}
}

// --- Urban Daggertooth ---------------------------------------------------------------

func TestOldB034DaggertoothProliferatesWhenDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	dino := effectstest.PushCatalogPermanent(g, me.ID, "Urban Daggertooth", "Creature — Dinosaur", oldB034Daggertooth, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, dino)
		c.Power, c.Toughness = 4, 3
	})
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 2, 2)
	effectstest.PutCounters(t, g, other, game.CounterPlusOne, 1)
	effectstest.Pr6Damage(t, g, other, dino, 1)
	effectstest.SettleAnsweringProliferate(t, g)
	if got := effectstest.CounterCount(g, other, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2 (proliferated)", got)
	}
}

// --- Necromancer's Familiar --------------------------------------------------------------

func TestOldB034FamiliarIsLifelinkOnlyWithAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	fam := effectstest.PushCatalogPermanent(g, me.ID, "Necromancer's Familiar", "Creature — Bird Spirit", oldB034Familiar, false)
	me.Hand.Cards = nil
	if !effectstest.HasEffectiveKeyword(t, g, fam, "lifelink") {
		t.Error("an empty hand should give lifelink")
	}
	drawn := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneLibrary, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, drawn); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if effectstest.HasEffectiveKeyword(t, g, fam, "lifelink") {
		t.Error("a card in hand should turn lifelink off")
	}
}

func TestOldB034FamiliarDiscardsForIndestructibleAndTapsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	fam := effectstest.PushCatalogPermanent(g, me.ID, "Necromancer's Familiar", "Creature — Bird Spirit", oldB034Familiar, false)
	card := effectstest.HandSpell(me, "Pitch", "Sorcery", "")
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, fam, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, fam, "indestructible") {
		t.Error("no indestructible after the ability")
	}
	if !effectstest.TappedOnBattlefield(t, g, fam) {
		t.Error("the Familiar was not tapped")
	}
}

func TestOldB034FamiliarNeedsAPoolAndACardToActivate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	fam := effectstest.PushCatalogPermanent(g, me.ID, "Necromancer's Familiar", "Creature — Bird Spirit", oldB034Familiar, false)
	card := effectstest.HandSpell(me, "Pitch", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, fam, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}}); err == nil {
		t.Error("activating without {B} must be refused")
	}
}
