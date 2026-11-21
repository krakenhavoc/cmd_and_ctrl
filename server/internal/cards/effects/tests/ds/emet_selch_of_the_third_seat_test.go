package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func emetCast(g *game.Game, me *game.Player, id uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
	})
}

// A loss offers the cast at the card's mana cost less {2}; the spell
// is exiled instead of going to the graveyard.
func TestEmetSelchCastsTheTargetAtItsCostLessTwoAndExilesIt(t *testing.T) {
	g, me, opp, emet := effectstest.EmetSetup(t)
	div := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	effectstest.EmetOffer(t, g, me, opp, emet, div)

	if err := emetCast(g, me, div); err == nil {
		t.Fatal("the cast is not free: Divination costs {U} after the discount")
	}
	effectstest.FloatForTest(g, me, "U")
	handBefore := len(me.Hand.Cards)
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("cast Divination for {U}: %v", err)
	}
	// #4250: cast during the trigger's resolution.
	effectstest.CastInsideTheResolution(t, g, div)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(div) || !effectstest.InExile(g, div) {
		t.Error("Divination goes to exile instead of the graveyard")
	}
	if got := len(me.Hand.Cards) - handBefore; got != 2 {
		t.Errorf("drew %d, want Divination's two", got)
	}
}

// The discount reads the zone the cast comes from: a hand cast pays
// the full price.
func TestEmetSelchDiscountsOnlyGraveyardCasts(t *testing.T) {
	g, me, _, _ := effectstest.EmetSetup(t)
	div := effectstest.HandCardFull(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle, []string{"U"})

	effectstest.FloatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a hand cast of a {2}{U} spell was discounted")
	}
	effectstest.FloatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, div, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for the full {2}{U}: %v", err)
	}
}

// "Do this only once each turn" counts the cast: after one is made a
// later loss puts nothing on the stack.
func TestEmetSelchOffersOnlyOnceEachTurn(t *testing.T) {
	g, me, opp, emet := effectstest.EmetSetup(t)
	div := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)
	effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	effectstest.EmetOffer(t, g, me, opp, emet, div)
	effectstest.FloatForTest(g, me, "U")
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.EmetOpponentLoses(t, g, emet, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TriggerOnStack(g, emet) != nil || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a second trigger fired the turn the first cast was made")
	}
}

// A declined offer spends nothing: the decline (#4250, the resolution_cast's
// empty answer) closes it, and the next loss offers again.
func TestEmetSelchDeclinedOfferLeavesTheTurnsUse(t *testing.T) {
	g, me, opp, emet := effectstest.EmetSetup(t)
	div := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	effectstest.EmetOffer(t, g, me, opp, emet, div)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, div)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if !me.Graveyard.Contains(div) {
		t.Fatal("the declined card left the graveyard")
	}
	effectstest.FloatForTest(g, me, "U")
	if err := emetCast(g, me, div); err == nil {
		t.Fatal("the lapsed offer still allowed a cast")
	}

	effectstest.EmetOffer(t, g, me, opp, emet, div)
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("a declined offer must not spend the turn's use: %v", err)
	}
}

// A card with its own flashback is cast for its MANA cost here, and
// the cast still spends the turn's use.
func TestEmetSelchCastOfAFlashbackCardCountsAndPaysManaCost(t *testing.T) {
	g, me, opp, emet := effectstest.EmetSetup(t)
	souls := effectstest.GyCard(me, "Lingering Souls", "Sorcery", "{2}{W}", effectstest.GyThemeLingeringSoulsOracle)
	// A second legal target, so a trigger that wrongly fires has one.
	effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	effectstest.EmetOffer(t, g, me, opp, emet, souls)
	effectstest.FloatForTest(g, me, "W")
	if err := emetCast(g, me, souls); err != nil {
		t.Fatalf("cast Lingering Souls for {W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, souls) {
		t.Error("the spell is exiled")
	}

	effectstest.EmetOpponentLoses(t, g, emet, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TriggerOnStack(g, emet) != nil || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the cast of a flashback card did not spend the turn's use")
	}
}

// Every opponent losing life at once is ONE trigger (CR 603.2c).
func TestEmetSelchOneOrMoreOpponentsIsOneTrigger(t *testing.T) {
	g, me, _, emet := effectstest.EmetSetup(t)
	effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	effectstest.CastCatalogSpell(t, g, "Cauldron Familiar", "Creature — Bat Skeleton", effectstest.CauldronFamiliarOra, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4 && effectstest.TriggerOnStack(g, emet) == nil && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == emet {
			n++
		}
	}
	prompts := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == me.ID {
			prompts++
		}
	}
	if n+prompts != 1 {
		t.Fatalf("three opponents lost life in one resolution: %d triggers + %d prompts, want exactly one", n, prompts)
	}
}

// The deck's own life gain or loss to the controller does not trigger.
func TestEmetSelchIgnoresItsControllersLifeLoss(t *testing.T) {
	g, me, _, emet := effectstest.EmetSetup(t)
	effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(emet, me.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TriggerOnStack(g, emet) != nil || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("Emet-Selch triggered on its controller's own life loss")
	}
}

// Only an instant or sorcery card is a legal target.
func TestEmetSelchTargetsOnlyInstantsAndSorceries(t *testing.T) {
	g, me, opp, emet := effectstest.EmetSetup(t)
	effectstest.GyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")

	effectstest.EmetOpponentLoses(t, g, emet, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TriggerOnStack(g, emet) != nil || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a trigger with no legal target went on the stack (CR 603.3d)")
	}
}

// #3863: "If that spell would be put into your graveyard, exile it
// instead" replaces a graveyard destination on every path, a counter
// included: a countered spell is exiled.
func TestEmetSelchExilesTheSpellWhenItIsCountered(t *testing.T) {
	g, me, opp, emet := effectstest.EmetSetup(t)
	div := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)
	effectstest.EmetOffer(t, g, me, opp, emet, div)
	effectstest.FloatForTest(g, me, "U")
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("cast Divination: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(div); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if me.Graveyard.Contains(div) || !effectstest.InExile(g, div) {
		t.Error("the countered spell goes to exile instead of the graveyard")
	}
}

// #3863: and ONLY a graveyard destination. A spell returned to its
// owner's hand goes there; flashback's wider "anywhere else" would
// have exiled it.
func TestEmetSelchLetsABouncedSpellGoToHand(t *testing.T) {
	g, me, opp, emet := effectstest.EmetSetup(t)
	div := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)
	effectstest.EmetOffer(t, g, me, opp, emet, div)
	effectstest.FloatForTest(g, me, "U")
	if err := emetCast(g, me, div); err != nil {
		t.Fatalf("cast Divination: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.ReturnSpellToHandForEffect(div); err != nil {
			t.Fatalf("return to hand: %v", err)
		}
	})
	if effectstest.InExile(g, div) || !me.Hand.Contains(div) {
		t.Error("the spell returned to its owner's hand was exiled")
	}
}
