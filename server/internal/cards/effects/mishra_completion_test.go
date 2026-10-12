package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSaheeliUltimateFinishesAfterCopiedArtifactsEntryChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	source := pushCatalogPermanent(g, me.ID, "Adaptive Automaton", "Artifact Creature — Construct", "53c730c6-2f8c-4af8-b400-b9d573a71e60", false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == source {
			g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
		}
	}
	item := &game.StackItem{Controller: me.ID}
	g.WithWriteLock(func() {
		if err := saheeliGiftedCopyArtifacts(g, item); err != nil {
			t.Fatal(err)
		}
	})
	if pendingOfKind(g, game.PendingChoiceCreatureType) == nil {
		t.Fatal("copy skipped its entry choice")
	}
	snapshot := g.Clone()
	for run := 0; run < 2; run++ {
		if run > 0 {
			g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
		}
		answerCreatureType(t, g, me.ID, "Construct")
		var token uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if IsToken(c) {
				token = c.InstanceID
				if !game.HasKeyword(&c, "haste") {
					t.Fatal("resumed token lacks haste")
				}
				for _, kw := range c.Keywords {
					if kw == "haste" {
						t.Fatal("granted haste became copiable printed text")
					}
				}
			}
		}
		if token == uuid.Nil || len(g.DelayedTriggers) == 0 {
			t.Fatal("resumed token missed its delayed exile")
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(token) {
			t.Fatal("resumed copy survived its exile trigger")
		}
	}
}

func TestPortalAddsPhyrexianAfterEntryChoiceAndUndo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := welderGraveyardArtifact(opp, "Adaptive Automaton", "53c730c6-2f8c-4af8-b400-b9d573a71e60")
	for i := range opp.Graveyard.Cards {
		c := &opp.Graveyard.Cards[i]
		if c.InstanceID == id {
			c.TypeLine = "Artifact Creature — Construct"
			c.Power, c.Toughness = 2, 2
		}
	}
	item := &game.StackItem{Controller: me.ID, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
	g.WithWriteLock(func() {
		if err := portalToPhyrexiaReanimate(g, item); err != nil {
			t.Fatal(err)
		}
	})
	snapshot := g.Clone()
	for run := 0; run < 2; run++ {
		if run > 0 {
			g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
		}
		answerCreatureType(t, g, me.ID, "Construct")
		c, ok := battlefieldCard(g, id)
		if !ok || c.Controller != me.ID || c.Owner != opp.ID || !c.HasSubtype("Phyrexian") {
			t.Fatalf("returned creature: %+v", c)
		}
	}
}

func TestDeadlyRollickCommanderAlternativeCost(t *testing.T) {
	for _, where := range []string{"command", "battlefield", "stolen"} {
		t.Run(where, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			commander := game.NewCard("Commander", me.ID)
			commander.TypeLine, commander.Controller = "Legendary Creature", me.ID
			commander.IsCommander = true
			if where == "stolen" {
				commander.Owner = opp.ID
			}
			if where == "command" {
				me.Command.PushTop(commander)
			} else {
				g.Battlefield.PushTop(commander)
			}
			alt := game.AlternativeCostByKey(deadlyRollickOracle, "free")
			if alt == nil {
				t.Fatal("free cast is not registered")
			}
			if got := alt.Condition(g, me.ID); got != (where != "command") {
				t.Fatalf("free cast eligibility for %s: %v", where, got)
			}
		})
	}
}
