package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const patternBufferOracle = "690d0d8f-48b9-40a6-b881-6f17ac19bf3b"

func TestPatternBufferPreservationIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(patternBufferOracle)
	if !ok {
		t.Fatal("Pattern Buffer Preservation is not registered")
	}
	if spec.Name != "Pattern Buffer Preservation" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// patternBufferTake resolves the pending look-at-two pick, taking
// `take`, and checks the hand, graveyard and library afterwards.
func patternBufferTake(t *testing.T, g *game.Game, me *game.Player, take, other uuid.UUID) {
	t.Helper()
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick between the top two cards")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top two", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	lib := me.Library.Size()
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{take}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(take) || me.Hand.Size() != hand+1 {
		t.Error("the chosen card is not the one new card in hand")
	}
	if !me.Graveyard.Contains(other) {
		t.Error("the other of the top two should be put into the graveyard")
	}
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d -> %d, want exactly the top two gone", lib, me.Library.Size())
	}
}

func TestPatternBufferPreservationTakesOneOfTheTopTwoAndBinsTheOther(t *testing.T) {
	g, me, _ := trekB8Main(t)
	seedLibrary(me, "Filler")
	lower := topLibraryCard(me, "Lower", "Sorcery")
	upper := topLibraryCard(me, "Upper", "Sorcery")
	trekB8Float(t, g, me, "{1}{B}")
	castFromHandForTest(t, g, me, "Pattern Buffer Preservation", "Sorcery", "{1}{B}", patternBufferOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	patternBufferTake(t, g, me, upper, lower)
}

func patternBufferInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	return pushToGraveyard(g, me, game.Card{Name: "Pattern Buffer Preservation", TypeLine: "Sorcery", ManaCost: "{1}{B}", OracleID: patternBufferOracle})
}

func TestPatternBufferPreservationBeamsUpFromTheGraveyardAndIsExiled(t *testing.T) {
	g, me, _ := trekB8Main(t)
	lower := topLibraryCard(me, "Lower", "Sorcery")
	upper := topLibraryCard(me, "Upper", "Sorcery")
	crew := b12Creature(g, me.ID, "Away Team", "Creature — Human", 2, 2)
	card := patternBufferInGraveyard(g, me)
	trekB8Float(t, g, me, "{1}{B}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{crew}, Strict: true,
	}); err != nil {
		t.Fatalf("beam me up cast: %v", err)
	}
	if !me.Hand.Contains(crew) || g.Battlefield.Contains(crew) {
		t.Fatal("the creature was not returned to its owner's hand as the cost")
	}
	if pool := me.ManaPool.Total(); pool != 0 {
		t.Errorf("%d mana left floating, want {1}{B} spent", pool)
	}
	passPriorityAroundTable(t, g)
	patternBufferTake(t, g, me, lower, upper)
	if me.Graveyard.Contains(card) || !g.Exile.Contains(card) {
		t.Error("a beamed-up spell must be exiled as it leaves the stack")
	}
}

func TestPatternBufferPreservationBeamMeUpNeedsACreatureYouControl(t *testing.T) {
	g, me, opp := trekB8Main(t)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	card := patternBufferInGraveyard(g, me)
	trekB8Float(t, g, me, "{1}{B}")
	for name, ids := range map[string][]uuid.UUID{"no creature": nil, "an opponent's creature": {theirs}} {
		if err := g.CastSpell(me.ID, card, game.CastSpellParams{
			FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: ids, Strict: true,
		}); err == nil {
			t.Fatalf("beam me up was cast returning %s", name)
		}
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("the card was cast from the graveyard without beam me up")
	}
	if !me.Graveyard.Contains(card) || !g.Battlefield.Contains(theirs) {
		t.Error("a refused cast moved a card")
	}
}

func TestPatternBufferPreservationBeamMeUpNeedsItsMana(t *testing.T) {
	g, me, _ := trekB8Main(t)
	crew := b12Creature(g, me.ID, "Away Team", "Creature — Human", 2, 2)
	card := patternBufferInGraveyard(g, me)
	trekB8Float(t, g, me, "{B}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{crew}, Strict: true,
	}); err == nil {
		t.Fatal("beam me up was cast with only {B}")
	}
	if !g.Battlefield.Contains(crew) || !me.Graveyard.Contains(card) {
		t.Error("a refused cast returned the creature or moved the card")
	}
}
