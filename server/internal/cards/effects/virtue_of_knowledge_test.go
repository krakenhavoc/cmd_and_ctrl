package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// virtue_of_knowledge_test.go — Virtue of Knowledge // Vantress Visions.
// The enchantment is Panharmonicon without the artifact-or-creature
// filter; the Adventure is Lithoform Engine's ability copy as an
// instant.

const vokRuinCrabOracle = "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7"

// virtueOfKnowledge builds the printed card the way deck.toGameCard
// does: per-face data, then SetFace(0).
func virtueOfKnowledge(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   virtueOfKnowledgeOracleID,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:     "Virtue of Knowledge",
				TypeLine: "Enchantment",
				ManaCost: "{4}{U}",
				Colors:   []string{"U"},
			},
			{
				Name:     "Vantress Visions",
				TypeLine: "Instant — Adventure",
				ManaCost: "{1}{U}",
				Colors:   []string{"U"},
			},
		},
	}
	c.SetFace(0)
	return c
}

// A LAND entering doubles your landfall trigger. Panharmonicon's own
// test pins that a land does not trip it; the Virtue's "a permanent"
// is the whole difference between the two cards.
func TestVirtueOfKnowledgeDoublesALandEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	virtue := virtueOfKnowledge(me.ID)
	g.Battlefield.PushTop(virtue)
	pushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", vokRuinCrabOracle, false)

	before := opp.Library.Size()
	enterFromHand(t, g, me.ID, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 6 {
		t.Errorf("Ruin Crab milled %d with Virtue of Knowledge out, want 6 (the trigger twice)", got)
	}
}

// "A triggered ability of a permanent YOU control": an opponent's
// watcher is left at one trigger.
func TestVirtueOfKnowledgeLeavesAnOpponentsTriggerAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.Battlefield.PushTop(virtueOfKnowledge(me.ID))
	pushCatalogPermanent(g, opp.ID, "Ruin Crab", "Creature — Crab", vokRuinCrabOracle, false)

	before := me.Library.Size()
	enterFromHand(t, g, opp.ID, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 3 {
		t.Errorf("an opponent's Ruin Crab milled %d, want 3 — the Virtue doubles only your triggers", got)
	}
}

// Vantress Visions offers an activated or triggered ability you
// control, not an opponent's, and resolving it puts a copy on the
// stack under your control.
func TestVantressVisionsCopiesAnAbilityYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	card := virtueOfKnowledge(me.ID)
	me.Hand.PushTop(card)
	other := pushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Artifact", "", false)

	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "an activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if err := g.AnnounceTrigger(me.ID, other, game.AbilityParams{Label: "a trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	if err := g.AnnounceTrigger(opp.ID, theirs, game.AbilityParams{Label: "their trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	act, trig, notMine := acItemLabelled(g, "an activation"), acItemLabelled(g, "a trigger"), acItemLabelled(g, "their trigger")

	spec := game.TargetSpecFor(virtueOfKnowledgeOracleID + "#1")
	if spec == nil || !spec.Abilities {
		t.Fatalf("Vantress Visions must target an ability item: %+v", spec)
	}
	offered := map[uuid.UUID]bool{}
	g.WithWriteLock(func() {
		for _, id := range g.LegalTargetsForEffect(game.SourceChooser(me.ID), spec).Cards {
			offered[id] = true
		}
	})
	if !offered[act] || !offered[trig] {
		t.Errorf("'activated or triggered ability you control' must offer both: act=%v trig=%v", offered[act], offered[trig])
	}
	if offered[notMine] {
		t.Error("an opponent's trigger is not 'you control'")
	}

	floatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{
		Face:    1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: act}},
	}); err != nil {
		t.Fatalf("cast Vantress Visions: %v", err)
	}
	acPassUntilCopy(t, g)
	cp := acCopyOnStack(g)
	if cp == nil {
		t.Fatal("Vantress Visions made no copy")
	}
	if cp.Kind != game.StackItemActivated {
		t.Errorf("the copy of an activated ability is an activated ability, got %v", cp.Kind)
	}
	if cp.Controller != me.ID {
		t.Error("the copy is controlled by the player who created it (CR 707.10b)")
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("CR 715.3d: the resolved Adventure spell is exiled")
	}
}
