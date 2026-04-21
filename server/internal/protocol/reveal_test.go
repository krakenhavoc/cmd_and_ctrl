package protocol

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestThoughtseizeRevealStickyAfterResolveChoice drives the live game
// through a Thoughtseize cast + resolve + resolve_choice sequence,
// then checks that the remaining opponent hand cards still project
// through FilterViewFor as face-up (KnownByYou=true) from the
// caster's perspective. This is the S14 sub-PR 5 sticky-reveal
// regression guard.
func TestThoughtseizeRevealStickyAfterResolveChoice(t *testing.T) {
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer %d: %v", i, err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand %v: %v", p.ID, err)
		}
	}

	caster := g.Seats[0]
	target := g.Seats[1]
	// Seed three named cards into target's hand.
	targetHand := []uuid.UUID{}
	for _, name := range []string{"Ponder", "Counterspell", "Terror"} {
		id := uuid.New()
		target.Hand.PushTop(game.Card{
			InstanceID: id,
			Name:       name,
			TypeLine:   "Instant",
			ScryfallID: "scry-" + name,
			Owner:      target.ID,
			Controller: target.ID,
		})
		targetHand = append(targetHand, id)
	}

	// Advance to main, cast Thoughtseize.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	thoughtseizeID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: thoughtseizeID,
		Name:       "Thoughtseize",
		TypeLine:   "Sorcery",
		OracleID:   "edd8d1e8-be43-4c38-bb3a-83081fbaf0b5",
		Owner:      caster.ID,
		Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, thoughtseizeID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// Pass priority until stack empty (Thoughtseize resolves).
	for i := 0; i < 16 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	// Caster should now be a knower of every target hand card.
	for _, id := range targetHand {
		for i := range target.Hand.Cards {
			if target.Hand.Cards[i].InstanceID == id && !target.Hand.Cards[i].IsKnownTo(caster.ID) {
				t.Errorf("caster not a knower of %v after Thoughtseize resolve", id)
			}
		}
	}

	// Before resolve_choice: FilterViewFor from caster's POV.
	v1 := FilterViewFor(ViewOfGame(g), caster.ID.String())
	var targetSeat *PlayerView
	for i := range v1.Seats {
		if v1.Seats[i].ID == target.ID.String() {
			targetSeat = &v1.Seats[i]
		}
	}
	if targetSeat == nil {
		t.Fatalf("target seat missing")
	}
	t.Logf("pre-choice: opponent hand cards=%d count=%d", len(targetSeat.Hand.Cards), targetSeat.Hand.Count)
	for _, c := range targetSeat.Hand.Cards {
		if !c.KnownByYou {
			t.Errorf("pre-choice: card %s KnownByYou=false", c.InstanceID)
		}
		if c.Name == "" {
			t.Errorf("pre-choice: card %s Name empty", c.InstanceID)
		}
	}

	// Resolve the choice — caster picks Ponder.
	if len(g.PendingChoices) != 1 {
		t.Fatalf("expected 1 pending choice, got %d", len(g.PendingChoices))
	}
	choice := g.PendingChoices[0]
	if err := g.ResolvePendingChoice(choice.ID, caster.ID, []uuid.UUID{targetHand[0]}); err != nil {
		t.Fatalf("ResolvePendingChoice: %v", err)
	}

	// After resolve_choice: remaining hand cards should STILL be
	// visible + KnownByYou=true.
	v2 := FilterViewFor(ViewOfGame(g), caster.ID.String())
	targetSeat = nil
	for i := range v2.Seats {
		if v2.Seats[i].ID == target.ID.String() {
			targetSeat = &v2.Seats[i]
		}
	}
	if targetSeat == nil {
		t.Fatalf("target seat missing post-resolve")
	}
	t.Logf("post-choice: opponent hand cards=%d count=%d", len(targetSeat.Hand.Cards), targetSeat.Hand.Count)
	if len(targetSeat.Hand.Cards) == 0 {
		t.Errorf("post-choice: all opponent hand cards dropped from wire")
	}
	for _, c := range targetSeat.Hand.Cards {
		if !c.KnownByYou {
			t.Errorf("post-choice: card %s KnownByYou=false (regression — sticky reveal lost)", c.InstanceID)
		}
	}
}
