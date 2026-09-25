package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestStackItemViewCarriesTriggerDoublerAttribution(t *testing.T) {
	doubler := uuid.New()
	item := &game.StackItem{
		ID:            uuid.New(),
		Kind:          game.StackItemTriggered,
		DoubledBy:     doubler,
		DoubledByName: "Panharmonicon",
	}

	got := viewOfStackItem(item)
	if got.DoubledBy != doubler.String() || got.DoubledByName != "Panharmonicon" {
		t.Fatalf("doubler attribution = %+v", got)
	}
	payload, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(payload, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["doubled_by"] != doubler.String() || wire["doubled_by_name"] != "Panharmonicon" {
		t.Fatalf("wire attribution = %#v", wire)
	}
}

func TestTriggerDoublerAttributionSurvivesHiddenChoiceFiltering(t *testing.T) {
	doubler := uuid.New()
	choice := PendingChoiceView{
		Kind:          string(game.PendingChoiceTriggerPrompt),
		Chooser:       uuid.New().String(),
		DoubledBy:     doubler.String(),
		DoubledByName: "Panharmonicon",
		Options: []CardView{{
			InstanceID: uuid.New().String(),
			Name:       "Hidden card",
		}},
	}

	filtered := filterPendingChoices([]PendingChoiceView{choice}, func(CardView) bool { return false }, uuid.New().String())
	if len(filtered) != 1 {
		t.Fatalf("filtered choices = %d, want 1", len(filtered))
	}
	if filtered[0].DoubledBy != choice.DoubledBy || filtered[0].DoubledByName != choice.DoubledByName {
		t.Fatalf("public attribution changed while filtering: %+v", filtered[0])
	}
	if len(filtered[0].Options) != 0 {
		t.Fatalf("hidden option leaked through trigger prompt: %+v", filtered[0].Options)
	}
}

func TestDoubledOptionalTriggerTargetPromptProjection(t *testing.T) {
	g := buildActiveGame(t)
	owner, opponent := g.Seats[0], g.Seats[1]
	sourceID, doublerID := uuid.New(), uuid.New()
	const sourceOracle, doublerOracle = "protocol-trigger-source", "protocol-trigger-doubler"

	previousTriggers := game.CatalogTriggers
	previousDoublers := game.CatalogTriggerDoublers
	t.Cleanup(func() {
		game.CatalogTriggers = previousTriggers
		game.CatalogTriggerDoublers = previousDoublers
	})
	game.CatalogTriggers = func(oracle string) []game.TriggeredAbility {
		if oracle != sourceOracle {
			return nil
		}
		return []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Choose a target?"},
			Targets: &game.TargetSpec{
				Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1,
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "targeted trigger")
			},
		}}
	}
	game.CatalogTriggerDoublers = func(oracle string) []game.TriggerDoubler {
		if oracle != doublerOracle {
			return nil
		}
		return []game.TriggerDoubler{{Label: "Panharmonicon", Applies: func(*game.Game, game.TriggerDoublingQuery) bool {
			return true
		}}}
	}

	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: sourceID, OracleID: sourceOracle, Name: "Source", TypeLine: "Creature", Owner: owner.ID, Controller: owner.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: doublerID, OracleID: doublerOracle, Name: "Panharmonicon", TypeLine: "Artifact", Owner: owner.ID, Controller: owner.ID})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: sourceID, Actor: owner.ID})
	})

	var doubled *game.PendingChoice
	for _, choice := range g.PendingChoices {
		if choice != nil {
			if id, name := choice.TriggerDoubler(); id == doublerID && name == "Panharmonicon" {
				choiceValue := *choice
				doubled = &choiceValue
				break
			}
		}
	}
	if doubled == nil {
		t.Fatal("engine did not produce a doubled optional trigger prompt")
	}
	if err := g.ResolveTriggerPrompt(doubled.ID, owner.ID, true); err != nil {
		t.Fatalf("resolve doubled optional prompt: %v", err)
	}

	view := ViewOfGameFor(g, owner.ID.String())
	var target *PendingChoiceView
	for i := range view.PendingChoices {
		if view.PendingChoices[i].Kind == string(game.PendingChoicePickTarget) {
			target = &view.PendingChoices[i]
			break
		}
	}
	if target == nil {
		t.Fatalf("doubled optional trigger did not become a pick_target prompt: %+v", view.PendingChoices)
	}
	if target.DoubledBy != doublerID.String() || target.DoubledByName != "Panharmonicon" {
		t.Fatalf("pick_target attribution = %q / %q", target.DoubledBy, target.DoubledByName)
	}
	if target.PickTarget == nil {
		t.Fatal("pick_target projection omitted legal target set")
	}
	containsOpponent := false
	for _, id := range target.PickTarget.Players {
		if id == opponent.ID.String() {
			containsOpponent = true
		}
	}
	if !containsOpponent {
		t.Fatalf("target legal set = %+v", target.PickTarget)
	}
}
