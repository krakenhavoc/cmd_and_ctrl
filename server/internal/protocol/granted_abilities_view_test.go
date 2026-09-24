package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_abilities_view_test.go — ADR 0093 Decision 8 on the wire:
// `ref` on every ability row, `granted_by` on a granted row, and
// `granted_abilities` on the card, for every viewer.

const (
	viewFixtureRite  = "view-fixture-rite"
	viewGrantMana    = "view-fixture/tap-for-green"
	viewGrantTrigger = "view-fixture/dies-draw"
)

func stubViewGrants(t *testing.T) {
	t.Helper()
	fixtures := map[string]*game.CardDef{
		viewFixtureRite: {Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			GrantAbilities: []string{viewGrantMana, viewGrantTrigger},
		}}},
		game.GrantKey(viewGrantMana): {
			ManaAbilities: []game.ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
			GrantText:     "{T}: Add {G}.",
		},
		game.GrantKey(viewGrantTrigger): {
			Triggered: []game.TriggeredAbility{{Watches: []game.EventKind{game.EventLTB}}},
			GrantText: "When this creature dies, draw a card.",
		},
	}
	prev := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef {
		if d, ok := fixtures[key]; ok {
			return d
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
	t.Cleanup(func() { game.CatalogLookup = prev })
}

func TestGrantedAbilitiesReachTheWireForEveryViewer(t *testing.T) {
	stubViewGrants(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	rite := game.NewCard("Rite", me.ID)
	rite.TypeLine = "Enchantment"
	rite.OracleID = viewFixtureRite
	forest := game.NewCard("Forest", me.ID)
	forest.TypeLine = "Basic Land — Forest"
	g.WithWriteLock(func() {
		for _, c := range []game.Card{bear, rite, forest} {
			c.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
			g.Battlefield.PushTop(c)
		}
	})
	g.BumpLayerVersionForTest()

	v := ViewOfGame(g)
	for _, viewer := range []struct {
		name string
		id   string
	}{{"controller", me.ID.String()}, {"opponent", opp.ID.String()}, {"spectator", ""}} {
		t.Run(viewer.name, func(t *testing.T) {
			fv := FilterViewFor(v, viewer.id)
			var gotBear, gotForest *CardView
			for i := range fv.Battlefield.Cards {
				switch fv.Battlefield.Cards[i].InstanceID {
				case bear.InstanceID.String():
					gotBear = &fv.Battlefield.Cards[i]
				case forest.InstanceID.String():
					gotForest = &fv.Battlefield.Cards[i]
				}
			}
			if gotBear == nil || gotForest == nil {
				t.Fatal("fixture cards missing from the view")
			}
			if len(gotBear.ManaAbilities) != 1 {
				t.Fatalf("bear mana rows = %+v, want the granted {G}", gotBear.ManaAbilities)
			}
			row := gotBear.ManaAbilities[0]
			if row.Ref != game.GrantedAbilityRef(viewGrantMana, 0, 0) {
				t.Errorf("row ref = %q", row.Ref)
			}
			if row.GrantedBy == nil || row.GrantedBy.ID != rite.InstanceID.String() || row.GrantedBy.Name != "Rite" {
				t.Errorf("granted_by = %+v, want the Rite by id and name", row.GrantedBy)
			}
			if len(gotBear.GrantedAbilities) != 2 {
				t.Fatalf("granted_abilities = %+v, want the mana ability and the trigger", gotBear.GrantedAbilities)
			}
			if ga := gotBear.GrantedAbilities[1]; ga.Text != "When this creature dies, draw a card." || ga.SourceName != "Rite" || ga.SourceID != rite.InstanceID.String() {
				t.Errorf("granted trigger = %+v", ga)
			}
			// An own / intrinsic row carries its ref and no grantor.
			if len(gotForest.ManaAbilities) != 1 || gotForest.ManaAbilities[0].Ref != "land:G" || gotForest.ManaAbilities[0].GrantedBy != nil {
				t.Errorf("forest rows = %+v", gotForest.ManaAbilities)
			}
			if len(gotForest.GrantedAbilities) != 0 {
				t.Errorf("forest granted_abilities = %+v, want none", gotForest.GrantedAbilities)
			}
		})
	}
}
