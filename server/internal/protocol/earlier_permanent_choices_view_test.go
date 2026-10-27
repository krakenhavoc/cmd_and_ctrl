package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEarlierPermanentChoicesArePublicAndUseFilteredCardIdentities(t *testing.T) {
	g := buildThreeSeatGame(t)
	first, second, current := g.Seats[0], g.Seats[1], g.Seats[2]
	publicID, publicAltID := uuid.New(), uuid.New()
	faceDownID, faceDownAltID := uuid.New(), uuid.New()
	currentID, currentAltID := uuid.New(), uuid.New()
	var setupErr error
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: publicID, Name: "Sunlit Meadow", TypeLine: "Land", Owner: first.ID, Controller: first.ID,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: publicAltID, Name: "Dawnfield", TypeLine: "Land", Owner: first.ID, Controller: first.ID,
		})
		hidden := game.Card{
			InstanceID: faceDownID, Name: "Secret Assassin", TypeLine: "Creature — Human Rogue",
			Owner: second.ID, Controller: second.ID,
		}
		hidden.SetFaceDown(game.FaceDownMorphed)
		g.Battlefield.PushTop(hidden)
		g.Battlefield.PushTop(game.Card{
			InstanceID: faceDownAltID, Name: "Fallback Rogue", TypeLine: "Creature — Rogue", Owner: second.ID, Controller: second.ID,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: currentID, Name: "Third Picker's Bear", TypeLine: "Creature — Bear", Owner: current.ID, Controller: current.ID,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: currentAltID, Name: "Third Picker's Wolf", TypeLine: "Creature — Wolf", Owner: current.ID, Controller: current.ID,
		})

		legs := []game.ChooseCardsPrompt{
			{Chooser: first.ID, FromPlayer: first.ID, Cards: []uuid.UUID{publicID, publicAltID}, Min: 1, Max: 1, Zone: game.ZoneBattlefield},
			{Chooser: second.ID, FromPlayer: second.ID, Cards: []uuid.UUID{faceDownID, faceDownAltID}, Min: 1, Max: 1, Zone: game.ZoneBattlefield},
			{Chooser: current.ID, FromPlayer: current.ID, Cards: []uuid.UUID{currentID, currentAltID}, Min: 1, Max: 1, Zone: game.ZoneBattlefield},
		}
		// #4361 S7b: the keyed collector (the closure one is gone).
		then := game.RegisterPromptRunThen("test/4361-s7b-earlier-"+uuid.NewString(), func(*game.Game, game.RunDone) error { return nil })
		setupErr = g.ChoosePermanentsInOrderKeyedForEffect(legs, then, uuid.Nil, game.RunData{})
	})
	if setupErr != nil {
		t.Fatalf("ChoosePermanentsInOrderKeyedForEffect: %v", setupErr)
	}

	for _, pick := range []struct {
		player *game.Player
		card   uuid.UUID
	}{{first, publicID}, {second, faceDownID}} {
		choice := choiceFor(t, ViewOfGameFor(g, pick.player.ID.String()), string(game.PendingChoiceChooseCards))
		if err := g.ResolveChooseCards(uuid.MustParse(choice.ID), pick.player.ID, []uuid.UUID{pick.card}); err != nil {
			t.Fatalf("ResolvePendingChoice for %s: %v", pick.player.Name, err)
		}
	}

	for _, viewer := range g.Seats {
		frame := ViewOfGameFor(g, viewer.ID.String())
		choice := choiceFor(t, frame, string(game.PendingChoiceChooseCards))
		if len(choice.EarlierChoices) != 2 {
			t.Fatalf("%s sees %d earlier selectors, want 2: %+v", viewer.Name, len(choice.EarlierChoices), choice.EarlierChoices)
		}
		if choice.EarlierChoices[0].PlayerID != first.ID.String() || len(choice.EarlierChoices[0].CardIDs) != 1 || choice.EarlierChoices[0].CardIDs[0] != publicID.String() {
			t.Errorf("%s earlier choice 0 = %+v, want %s choosing %s", viewer.Name, choice.EarlierChoices[0], first.ID, publicID)
		}
		if choice.EarlierChoices[1].PlayerID != second.ID.String() || len(choice.EarlierChoices[1].CardIDs) != 1 || choice.EarlierChoices[1].CardIDs[0] != faceDownID.String() {
			t.Errorf("%s earlier choice 1 = %+v, want %s choosing %s", viewer.Name, choice.EarlierChoices[1], second.ID, faceDownID)
		}

		var faceDown *CardView
		for i := range frame.Battlefield.Cards {
			if frame.Battlefield.Cards[i].InstanceID == faceDownID.String() {
				faceDown = &frame.Battlefield.Cards[i]
				break
			}
		}
		if faceDown == nil || !faceDown.FaceDown {
			t.Errorf("%s did not receive the public face-down object with its identity hidden: %+v", viewer.Name, faceDown)
		}
		encoded, err := json.Marshal(frame)
		if err != nil {
			t.Fatalf("marshal %s's frame: %v", viewer.Name, err)
		}
		if !strings.Contains(string(encoded), faceDownID.String()) {
			t.Errorf("%s did not receive the public selected-object ID", viewer.Name)
		}
		if viewer.ID != second.ID && (strings.Contains(string(encoded), "Secret Assassin") || (faceDown != nil && faceDown.Name != "")) {
			t.Errorf("%s's frame leaked the face-down object's identity", viewer.Name)
		}
	}
}
