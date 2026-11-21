package testss82v

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsp "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/p"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSingleGraveyardCardsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		oracle, name string
		full         bool
	}{
		{effectstest.OracleDigsiteConservator, "Digsite Conservator", true},
		{effectstest.OracleCarrionBeetles, "Carrion Beetles", true},
		{effectstest.OracleRagDealer, "Rag Dealer", true},
		{effectstest.OracleDecompose, "Decompose", true},
		{effectstest.OracleFamishedGhoul, "Famished Ghoul", true},
		{effectstest.OracleGriffnautTracker, "Griffnaut Tracker", true},
		{effectstest.OracleScarabFeast, "Scarab Feast", true},
		{effectstest.OracleRapidDecay, "Rapid Decay", true},
		{effectstest.OracleShredMemory, "Shred Memory", true},
		{effectstest.OracleArashinSunshield, "Arashin Sunshield", true},
		{effectstest.OracleGravegouger, "Gravegouger", true},
		{effectstest.OracleSoulShackledZombie, "Soul-Shackled Zombie", true},
		{effectstest.OracleEbonyCharm, "Ebony Charm", true},
		{effectstest.OracleRatsFeast, "Rats' Feast", true},
		{effectstest.OracleSereneRemembrance, "Serene Remembrance", true},
		{effectstest.OracleUnlicensedHearse, "Unlicensed Hearse", true},
		{effectstest.OracleWasteManagement, "Waste Management", true},
		{effectstest.OracleKayaOrzhovUsurper, "Kaya, Orzhov Usurper", true},
		{effectstest.OracleCeaseDesist, "Cease", true},
		{effectstest.OracleCeaseDesist + "#1", "Desist", true},
		{effectstest.OraclePushPull, "Push", true},
		{effectstest.OraclePushPull + "#1", "Pull", true},
		{cardsp.PestilentCauldronOracleID, "Pestilent Cauldron", true},
		{cardsp.PestilentCauldronOracleID + "#1", "Restorative Burst", true},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Errorf("%s: registered = %v as %q", tc.name, ok, spec.Name)
			continue
		}
		if full := spec.Completeness == effects.CompletenessFull; full != tc.full {
			t.Errorf("%s: Completeness = %s, want full=%v", tc.name, spec.Completeness, tc.full)
		}
	}
}

// --- spells -------------------------------------------------------

// Decompose: three from one graveyard is legal and all three go; one
// from each of two graveyards is refused with the rule's sentence.
func TestDecomposeExilesFromOneGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "A", "Creature — Bear")
	b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
	c := effectstest.Batch01GraveyardCard(opp, "C", "Instant")
	mine := effectstest.Batch01GraveyardCard(me, "Mine", "Creature — Bear")

	err := effectstest.CastCatalogSpellErr(t, g, "Decompose", "Sorcery", effectstest.OracleDecompose, effectstest.SgCards(a, mine))
	if !errors.Is(err, game.ErrIllegalTarget) || !strings.Contains(err.Error(), "come from a single graveyard") {
		t.Fatalf("two graveyards: err = %v, want the single-graveyard refusal", err)
	}

	effectstest.CastCatalogSpell(t, g, "Decompose", "Sorcery", effectstest.OracleDecompose, effectstest.SgCards(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SgInExile(g, a, b, c) {
		t.Error("all three cards from the one graveyard are exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("the other graveyard is untouched")
	}
}

// A card that leaves in response is skipped; the rest still go
// (CR 608.2b).
func TestScarabFeastSkipsATargetThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "A", "Creature — Bear")
	b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Scarab Feast", "Instant", effectstest.OracleScarabFeast, effectstest.SgCards(a, b))
	g.WithWriteLock(func() { opp.Graveyard.Remove(b) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) {
		t.Error("the survivor is exiled")
	}
}

// Rats' Feast: X targets, all from one graveyard.
func TestRatsFeastExilesXFromOneGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
	b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
	mine := effectstest.Batch01GraveyardCard(me, "Mine", "Sorcery")
	effectstest.B10AddMana(me, "B", "B", "B")
	if err := effectstest.CastX(t, g, "Rats' Feast", "Sorcery", effectstest.OracleRatsFeast, 2, effectstest.SgCards(a, mine)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("X=2 across two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if err := effectstest.CastX(t, g, "Rats' Feast", "Sorcery", effectstest.OracleRatsFeast, 2, effectstest.SgCards(a, b)); err != nil {
		t.Fatalf("X=2 from one graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SgInExile(g, a, b) {
		t.Error("both cards are exiled")
	}
}

// Serene Remembrance goes into its owner's library with its targets,
// each into its owner's.
func TestSereneRemembranceShufflesItselfAndTheTargetsIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "A", "Creature — Bear")
	b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
	spell := effectstest.CastCatalogSpell(t, g, "Serene Remembrance", "Sorcery", effectstest.OracleSereneRemembrance, effectstest.SgCards(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Library.Contains(a) || !opp.Library.Contains(b) {
		t.Error("the targets are in their owner's library")
	}
	if !me.Library.Contains(spell) {
		t.Errorf("Serene Remembrance is in its owner's library, not the graveyard (graveyard has it: %v)", me.Graveyard.Contains(spell))
	}
}

// Shred Memory's transmute finds a card with the same mana value (2),
// at sorcery speed, from the hand.
func TestShredMemoryTransmutesForAManaValueTwoCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	shred := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: shred, Name: "Shred Memory", TypeLine: "Instant", ManaCost: "{1}{B}",
		OracleID: effectstest.OracleShredMemory, Owner: me.ID, Controller: me.ID,
	})
	two := effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID})
	effectstest.B10AddMana(me, "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, shred, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transmute: %v", err)
	}
	if !me.Graveyard.Contains(shred) {
		t.Fatal("Shred Memory is discarded to pay for transmute")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{two}); err != nil {
				t.Fatalf("ResolveSearchLibrary: %v", err)
			}
		}
	}
	if !me.Hand.Contains(two) {
		t.Error("the mana value 2 card is in hand")
	}
}

// Waste Management unkicked: up to two from one graveyard, a Rogue per
// creature card exiled. Kicked: the whole of one player's graveyard.
func TestWasteManagementMakesARoguePerCreatureCardOnEitherBranch(t *testing.T) {
	t.Run("unkicked", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.Batch01GraveyardCard(opp, "Bear", "Creature — Bear")
		sorc := effectstest.Batch01GraveyardCard(opp, "Sorc", "Sorcery")
		effectstest.CastCatalogSpell(t, g, "Waste Management", "Instant", effectstest.OracleWasteManagement, effectstest.SgCards(bear, sorc))
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.SgInExile(g, bear, sorc) {
			t.Fatal("both cards are exiled")
		}
		if n := effectstest.OnBattlefieldNamed(g, "Rogue"); n != 1 {
			t.Errorf("Rogues = %d, want 1 for the one creature card", n)
		}
		_ = me
	})
	t.Run("kicked", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		for i := 0; i < 3; i++ {
			effectstest.Batch01GraveyardCard(opp, "Bear", "Creature — Bear")
		}
		effectstest.Batch01GraveyardCard(opp, "Sorc", "Sorcery")
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Waste Management", TypeLine: "Instant",
			OracleID: effectstest.OracleWasteManagement, Owner: me.ID, Controller: me.ID,
		})
		effectstest.B10AddMana(me, "B", "B", "B", "B")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{
			OptionalCosts: []int{0},
			Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("kicked cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Graveyard.Size() != 0 {
			t.Errorf("the target player's graveyard is exiled, %d cards left", opp.Graveyard.Size())
		}
		if n := effectstest.OnBattlefieldNamed(g, "Rogue"); n != 3 {
			t.Errorf("Rogues = %d, want 3", n)
		}
	})
}

// Ebony Charm's second mode is the family's clause on a bullet.
func TestEbonyCharmSecondModeExilesFromOneGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
	mine := effectstest.Batch01GraveyardCard(me, "Mine", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	cast := func(targets ...uuid.UUID) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ebony Charm", TypeLine: "Instant", OracleID: effectstest.OracleEbonyCharm, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.SgCards(targets...)})
	}
	if err := cast(a, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards on the mode: err = %v, want ErrIllegalTarget", err)
	}
	if err := cast(a); err != nil {
		t.Fatalf("one graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || g.Exile.Contains(mine) {
		t.Error("the chosen card is exiled and nothing else")
	}
}

// Cease: the cards from one graveyard, then any player gains 2 and
// draws.
func TestCeaseExilesThenTheTargetPlayerGainsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	card := effectstest.MdfcCard(me.ID, effectstest.OracleCeaseDesist, game.LayoutSplit,
		game.Face{Name: "Cease", TypeLine: "Instant", ManaCost: "{0}"},
		game.Face{Name: "Desist", TypeLine: "Sorcery", ManaCost: "{0}"},
	)
	me.Hand.PushTop(card)
	life, hand := me.Life, me.Hand.Size()
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a},
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 1},
	}}); err != nil {
		t.Fatalf("cast Cease: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) {
		t.Error("the card is exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	// The hand lost Cease and drew one.
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want the same (Cease left, a card was drawn)", hand, me.Hand.Size())
	}
}

// Pull: two creature cards from one graveyard enter under the caster's
// control with haste, and are sacrificed at the next end step.
func TestPullReanimatesFromOneGraveyardWithHasteThenSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "Bear A", "Creature — Bear")
	b := effectstest.Batch01GraveyardCard(opp, "Bear B", "Creature — Bear")
	mine := effectstest.Batch01GraveyardCard(me, "My Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	cast := func(targets ...uuid.UUID) error {
		card := effectstest.MdfcCard(me.ID, effectstest.OraclePushPull, game.LayoutSplit,
			game.Face{Name: "Push", TypeLine: "Sorcery", ManaCost: "{0}"},
			game.Face{Name: "Pull", TypeLine: "Sorcery", ManaCost: "{0}"},
		)
		me.Hand.PushTop(card)
		return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.SgCards(targets...)})
	}
	if err := cast(a, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if err := cast(a, b); err != nil {
		t.Fatalf("cast Pull: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !effects.OnBattlefield(g, id) || c.Controller != me.ID {
			t.Fatalf("%s is on the battlefield under the caster's control", id)
		}
		if !effectstest.BearHasKeyword(t, g, id, "haste") {
			t.Errorf("%s has haste", c.Name)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, a) || effects.OnBattlefield(g, b) {
		t.Error("both are sacrificed at the end step")
	}
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("sacrificed, they go to their owner's graveyard")
	}
}

// Push destroys only a TAPPED creature.
func TestPushDestroysATappedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	effectstest.AdvanceToMain(t, g)
	cast := func(target uuid.UUID) error {
		card := effectstest.MdfcCard(me.ID, effectstest.OraclePushPull, game.LayoutSplit,
			game.Face{Name: "Push", TypeLine: "Sorcery", ManaCost: "{0}"},
			game.Face{Name: "Pull", TypeLine: "Sorcery", ManaCost: "{0}"},
		)
		me.Hand.PushTop(card)
		return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Targets: effectstest.SgCards(target)})
	}
	if err := cast(untapped); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an untapped creature: err = %v, want ErrIllegalTarget", err)
	}
	if err := cast(tapped); err != nil {
		t.Fatalf("cast Push: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, tapped) {
		t.Error("the tapped creature is destroyed")
	}
}

// Restorative Burst: up to two cards back to hand, everyone gains 4,
// and it exiles itself.
func TestRestorativeBurstReturnsGivesLifeAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.Batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	card := effectstest.MdfcCard(me.ID, cardsp.PestilentCauldronOracleID, game.LayoutModalDFC,
		game.Face{Name: "Pestilent Cauldron", TypeLine: "Artifact", ManaCost: "{0}"},
		game.Face{Name: "Restorative Burst", TypeLine: "Sorcery", ManaCost: "{0}"},
	)
	me.Hand.PushTop(card)
	myLife, oppLife := me.Life, opp.Life
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.SgCards(land)}); err != nil {
		t.Fatalf("cast Restorative Burst: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card returns to hand")
	}
	if me.Life != myLife+4 || opp.Life != oppLife+4 {
		t.Errorf("life: me %d → %d, opponent %d → %d, want +4 each", myLife, me.Life, oppLife, opp.Life)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("Restorative Burst exiles itself")
	}
}

// --- activated abilities -----------------------------------------

func TestCarrionBeetlesAndRagDealerExileUpToThree(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Carrion Beetles", effectstest.OracleCarrionBeetles},
		{"Rag Dealer", effectstest.OracleRagDealer},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Creature — Insect", tc.oracle, false)
			a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
			b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
			mine := effectstest.Batch01GraveyardCard(me, "Mine", "Sorcery")
			effectstest.AdvanceToMain(t, g)
			effectstest.B10AddMana(me, "B", "B", "B")
			err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(a, mine)})
			if !errors.Is(err, game.ErrIllegalTarget) {
				t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
			}
			if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(a, b)}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.SgInExile(g, a, b) {
				t.Error("both are exiled")
			}
		})
	}
}

// Famished Ghoul is sacrificed as the cost and its ability still
// resolves (CR 113.7a).
func TestFamishedGhoulResolvesAfterItsSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := effectstest.PushCatalogPermanent(g, me.ID, "Famished Ghoul", "Creature — Zombie", effectstest.OracleFamishedGhoul, false)
	a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
	b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(a, b)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effects.OnBattlefield(g, ghoul) {
		t.Fatal("the Ghoul is sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SgInExile(g, a, b) {
		t.Error("both are exiled")
	}
}

// Digsite Conservator: the sorcery-speed sacrifice exiles four from one
// graveyard; dying asks for {4}, and declining discovers nothing.
func TestDigsiteConservatorExilesFourFromOneGraveyardAndAsksToDiscover(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gnome := effectstest.PushCatalogPermanent(g, me.ID, "Digsite Conservator", "Artifact Creature — Gnome", effectstest.OracleDigsiteConservator, false)
	var four []uuid.UUID
	for i := 0; i < 4; i++ {
		four = append(four, effectstest.Batch01GraveyardCard(opp, "Card", "Sorcery"))
	}
	mine := effectstest.Batch01GraveyardCard(me, "Mine", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, gnome, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(four[0], four[1], four[2], mine)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if !effects.OnBattlefield(g, gnome) {
		t.Fatal("a refused activation pays nothing")
	}
	if err := g.ActivateCatalogAbility(me.ID, gnome, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(four...)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.AnswerPayUnless(t, g, me.ID, false)
	if ask.PayCost != "{4}" {
		t.Errorf("the dies trigger asks for {4}, got %q", ask.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SgInExile(g, four...) {
		t.Error("all four are exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("the other graveyard is untouched")
	}
}

// The sacrifice is sorcery-speed only (CR 602.5d).
func TestDigsiteConservatorActivatesOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gnome := effectstest.PushCatalogPermanent(g, me.ID, "Digsite Conservator", "Artifact Creature — Gnome", effectstest.OracleDigsiteConservator, false)
	a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
	// The harness parks on the turn-1 draw step: not a main phase.
	if err := g.ActivateCatalogAbility(me.ID, gnome, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(a)}); err == nil {
		t.Fatal("activated outside a main phase")
	}
}

// Pestilent Cauldron's exact four: four from one graveyard exiles and
// draws; four split over two graveyards is refused.
func TestPestilentCauldronNeedsFourFromOneGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cauldron := effectstest.PushCatalogPermanent(g, me.ID, "Pestilent Cauldron", "Artifact", cardsp.PestilentCauldronOracleID, false)
	var theirs []uuid.UUID
	for i := 0; i < 4; i++ {
		theirs = append(theirs, effectstest.Batch01GraveyardCard(opp, "Card", "Sorcery"))
	}
	mine := effectstest.Batch01GraveyardCard(me, "Mine", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "B", "B", "B", "B")
	err := g.ActivateCatalogAbility(me.ID, cauldron, 2, game.ActivateAbilityParams{Targets: effectstest.SgCards(theirs[0], theirs[1], theirs[2], mine)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("four over two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 2, game.ActivateAbilityParams{Targets: effectstest.SgCards(theirs...)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SgInExile(g, theirs...) {
		t.Error("all four are exiled")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a card drawn", hand, me.Hand.Size())
	}
}

// Unlicensed Hearse is as big as the number of cards it has exiled.
func TestUnlicensedHearseCountsTheCardsExiledWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hearse := effectstest.PushCatalogPermanent(g, me.ID, "Unlicensed Hearse", "Artifact — Vehicle", effectstest.OracleUnlicensedHearse, false)
	a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
	b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	if p := effectstest.EffectivePower(t, g, hearse); p != 0 {
		t.Errorf("power before anything is exiled = %d, want 0", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, hearse, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(a, b)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, hearse), effectstest.EffectiveToughness(t, g, hearse); p != 2 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 2/2", p, tough)
	}
	// A card that leaves exile no longer counts (CR 607.2a).
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(a) })
	if p := effectstest.EffectivePower(t, g, hearse); p != 1 {
		t.Errorf("power after one card left exile = %d, want 1", p)
	}
}

// Kaya's +1 gains 2 only when a creature card was exiled; the −5 deals
// and gains the number of cards the player owns in exile.
func TestKayaOrzhovUsurper(t *testing.T) {
	t.Run("+1 with a creature card", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kaya := effectstest.PushCatalogWalker(g, me.ID, "Kaya, Orzhov Usurper", effectstest.OracleKayaOrzhovUsurper, 3)
		bear := effectstest.Batch01GraveyardCard(opp, "Bear", "Creature — Bear")
		effectstest.AdvanceToMain(t, g)
		life := me.Life
		if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(bear)}); err != nil {
			t.Fatalf("+1: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+2 {
			t.Errorf("life %d → %d, want +2", life, me.Life)
		}
	})
	t.Run("+1 without one", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kaya := effectstest.PushCatalogWalker(g, me.ID, "Kaya, Orzhov Usurper", effectstest.OracleKayaOrzhovUsurper, 3)
		sorc := effectstest.Batch01GraveyardCard(opp, "Sorc", "Sorcery")
		effectstest.AdvanceToMain(t, g)
		life := me.Life
		if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(sorc)}); err != nil {
			t.Fatalf("+1: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(sorc) || me.Life != life {
			t.Errorf("the sorcery is exiled and no life is gained (life %d → %d)", life, me.Life)
		}
	})
	t.Run("−5", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kaya := effectstest.PushCatalogWalker(g, me.ID, "Kaya, Orzhov Usurper", effectstest.OracleKayaOrzhovUsurper, 6)
		for i := 0; i < 3; i++ {
			g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Exiled", Owner: opp.ID, Controller: opp.ID})
		}
		g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine", Owner: me.ID, Controller: me.ID})
		effectstest.AdvanceToMain(t, g)
		myLife, oppLife := me.Life, opp.Life
		if err := g.ActivateCatalogAbility(me.ID, kaya, 2, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		}); err != nil {
			t.Fatalf("−5: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != oppLife-3 || me.Life != myLife+3 {
			t.Errorf("opponent %d → %d (want −3), me %d → %d (want +3)", oppLife, opp.Life, myLife, me.Life)
		}
	})
}

// Griffnaut Tracker's entry trigger refuses two graveyards, then
// exiles two from one.
func TestGriffnautTrackerEntryExilesFromOneGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
	b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
	mine := effectstest.Batch01GraveyardCard(me, "Mine", "Sorcery")
	effectstest.CastHoldCreature(t, g, "Griffnaut Tracker", "Creature — Human Detective", effectstest.OracleGriffnautTracker, 2, 3)
	if err := effectstest.SgAnswerTargets(t, g, a, mine); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two graveyards: err = %v, want ErrIllegalTarget", err)
	}
	if err := effectstest.SgAnswerTargets(t, g, a, b); err != nil {
		t.Fatalf("one graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SgInExile(g, a, b) || !me.Graveyard.Contains(mine) {
		t.Error("the two picked cards are exiled and nothing else")
	}
}

// Soul-Shackled Zombie drains only when a creature card was exiled.
func TestSoulShackledZombieDrainsOnlyForACreatureCard(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		drain    bool
	}{
		{"a creature card", "Creature — Bear", true},
		{"no creature card", "Sorcery", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			card := effectstest.Batch01GraveyardCard(opp, "Card", tc.typeLine)
			myLife, oppLife := me.Life, opp.Life
			effectstest.CastHoldCreature(t, g, "Soul-Shackled Zombie", "Creature — Zombie", effectstest.OracleSoulShackledZombie, 4, 2)
			if err := effectstest.SgAnswerTargets(t, g, card); err != nil {
				t.Fatalf("answer: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			drained := opp.Life == oppLife-2 && me.Life == myLife+2
			if drained != tc.drain {
				t.Errorf("drained = %v (me %d → %d, opponent %d → %d), want %v", drained, myLife, me.Life, oppLife, opp.Life, tc.drain)
			}
		})
	}
}

// Gravegouger returns the cards it exiled when it leaves.
func TestGravegougerReturnsTheExiledCardsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.Batch01GraveyardCard(opp, "A", "Sorcery")
	b := effectstest.Batch01GraveyardCard(opp, "B", "Sorcery")
	gouger := effectstest.CastHoldCreature(t, g, "Gravegouger", "Creature — Nightmare Horror", effectstest.OracleGravegouger, 3, 3)
	if err := effectstest.SgAnswerTargets(t, g, a, b); err != nil {
		t.Fatalf("answer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SgInExile(g, a, b) {
		t.Fatal("both are exiled")
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(gouger) })
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("the exiled cards are back in their owner's graveyard")
	}
}

// Arashin Sunshield's tapper.
func TestArashinSunshieldTapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shield := effectstest.PushCatalogPermanent(g, me.ID, "Arashin Sunshield", "Creature — Human Warrior", effectstest.OracleArashinSunshield, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B10AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, shield, 0, game.ActivateAbilityParams{Targets: effectstest.SgCards(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(bear)
	if !c.Tapped {
		t.Error("the creature is tapped")
	}
}
