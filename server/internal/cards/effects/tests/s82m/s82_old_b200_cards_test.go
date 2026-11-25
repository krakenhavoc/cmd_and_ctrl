package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b200_cards_test.go — Core Set 2019, batch 200 (S82 older sets,
// #5527). Every test plays the real catalog entry through the engine.

const (
	oldB200LilianasSpoilsOracle   = "751da0fa-0ab1-4b8c-81bb-d86a0a280ac4"
	oldB200GhostformOracle        = "8d813b72-8ead-4a61-88ff-e93c8d843196"
	oldB200TezzeretsStriderOracle = "cac79a2b-1de3-4461-95ac-fa0f1129cd6b"

	oldB200BearType = "Creature — Bear"
)

func TestOldB200EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB200LilianasSpoilsOracle, oldB200GhostformOracle, oldB200TezzeretsStriderOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB200LilianasSpoilsDiscardsAndTakesABlackCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	// Stack the library so the first listed is on top.
	cards := []game.Card{
		{Name: "Red Card", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}},
		{Name: "Black Card", TypeLine: "Sorcery", ManaCost: "{B}", Colors: []string{"B"}},
		{Name: "Land", TypeLine: "Land"}, {Name: "Land", TypeLine: "Land"}, {Name: "Land", TypeLine: "Land"},
		{Name: "Deep Black", TypeLine: "Sorcery", ManaCost: "{B}", Colors: []string{"B"}},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		cards[i].InstanceID = uuid.New()
		cards[i].Owner, cards[i].Controller = me.ID, me.ID
		me.Library.PushTop(cards[i])
	}
	red, black, deep := cards[0].InstanceID, cards[1].InstanceID, cards[5].InstanceID
	spare := effectstest.PushCatalogHandCard(opp, "Spare", "Sorcery", "")
	hand := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Liliana's Spoils", "Sorcery", oldB200LilianasSpoilsOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, spare)
	if opp.Hand.Size() != hand-1 || !opp.Graveyard.Contains(spare) {
		t.Errorf("the opponent did not discard the card they chose (hand %d -> %d)", hand, opp.Hand.Size())
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to take a black card")
	}
	for _, id := range pick.ChooseCards {
		if id == red || id == deep {
			t.Errorf("offered a card it may not take: %s", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, black)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(black) {
		t.Error("the black card is not in hand")
	}
}

func TestOldB200GhostformMakesUpToTwoCreaturesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", oldB200BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", oldB200BearType, 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", oldB200BearType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB200BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ghostform", "Sorcery", oldB200GhostformOracle, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b, c} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.BlockRefusal(t, g, blocker, id); got == "" {
			t.Errorf("the unblockable creature %s was blockable", id)
		}
	}
	if got := effectstest.BlockRefusal(t, g, blocker, c); got != "" {
		t.Errorf("the third creature was refused a block: %q", got)
	}
}

func TestOldB200GhostformTakesNoTargetsToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Ghostform", "Sorcery", oldB200GhostformOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB200TezzeretsStriderHasMenaceWithATezzeret(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	strider := effectstest.B12Push(g, me.ID, "Tezzeret's Strider", "Artifact Creature — Golem", oldB200TezzeretsStriderOracle, 3, 1)
	if effectstest.HasEffectiveKeyword(t, g, strider, "menace") {
		t.Error("menace with no planeswalker")
	}
	for _, tc := range []struct {
		owner   uuid.UUID
		subtype string
		want    bool
	}{{me.ID, "Ajani", false}, {opp.ID, "Tezzeret", false}, {me.ID, "Tezzeret", true}} {
		pw := effectstest.B12Permanent(g, tc.owner, tc.subtype+" Test", "Legendary Planeswalker — "+tc.subtype)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(pw, game.CounterLoyalty, 4) })
		if got := effectstest.HasEffectiveKeyword(t, g, strider, "menace"); got != tc.want {
			t.Errorf("after %s's %s: menace = %v, want %v", "a player", tc.subtype, got, tc.want)
		}
	}
}
