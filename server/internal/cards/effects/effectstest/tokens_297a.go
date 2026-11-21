package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castFromHand casts `card` from the active seat's hand with the given
// announcement, after advancing to a main phase.
func CastFromHand(t *testing.T, g *game.Game, card game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = active.ID, active.ID
	active.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, card.InstanceID, params); err != nil {
		t.Fatalf("CastSpell %s: %v", card.Name, err)
	}
	return card.InstanceID
}

// castFlashCreatureFrom casts a creature with flash out of `seat`'s
// hand whoever's turn it is, so a test can have an opponent's creature
// enter during the active player's turn.
func CastFlashCreatureFrom(t *testing.T, g *game.Game, seat *game.Player, card game.Card) uuid.UUID {
	t.Helper()
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = seat.ID, seat.ID
	card.Keywords = append(card.Keywords, "flash")
	seat.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(seat.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s from seat %s: %v", card.Name, seat.ID, err)
	}
	return card.InstanceID
}

// cardsNamed counts battlefield cards with a name under a controller.
func CardsNamed(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

func PickPlayerTarget(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}
