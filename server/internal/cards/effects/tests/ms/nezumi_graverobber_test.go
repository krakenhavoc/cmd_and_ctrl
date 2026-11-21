package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsn "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/n"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func nezumiGraverobberRow() cards.Card {
	return effectstest.FlipRow(cardsn.NezumiGraverobberOracleID, "{1}{B}", []string{"B"}, nil,
		cards.CardFace{Name: "Nezumi Graverobber", TypeLine: "Creature — Rat Rogue",
			OracleText: "{1}{B}: Exile target card from an opponent's graveyard. If no cards are in that graveyard, flip this creature.",
			Power:      "2", Toughness: "1"},
		cards.CardFace{Name: "Nighteyes the Desecrator", TypeLine: "Legendary Creature — Rat Wizard",
			OracleText: "{4}{B}: Put target creature card from a graveyard onto the battlefield under your control.",
			Power:      "4", Toughness: "2"})
}

// The Graverobber exiles a card from an opponent's graveyard, and flips
// only if that graveyard is then empty.
func TestNezumiGraverobberFlipsWhenTheGraveyardIsEmpty(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cards    int
		wantFlip bool
	}{{"last card", 1, true}, {"two cards", 2, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.FlipMain(t)
			id := effectstest.FlipEntered(t, g, nezumiGraverobberRow(), me)
			g.WithWriteLock(func() { opp.Graveyard.Cards = nil })
			var first = effectstest.GraveCard(opp, "Dead One", "Sorcery")
			for i := 1; i < tc.cards; i++ {
				effectstest.GraveCard(opp, "Dead Extra", "Sorcery")
			}
			effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(first)})
			if !g.Exile.Contains(first) {
				t.Fatal("the targeted card was not exiled")
			}
			if tc.wantFlip {
				effectstest.WantFlipped(t, g, id, "Nighteyes the Desecrator", []string{"B"}, 2)
			} else {
				effectstest.WantUnflipped(t, g, id, "Nezumi Graverobber")
			}
		})
	}
}

// You can't aim it at your own graveyard.
func TestNezumiGraverobberCannotTargetYourOwnGraveyard(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, nezumiGraverobberRow(), me)
	mine := effectstest.GraveCard(me, "My Dead One", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)}); err == nil {
		t.Error("exiled a card from my own graveyard")
	}
}

// Nighteyes puts a creature card from ANY graveyard onto the battlefield
// under the activator's control.
func TestNighteyesReanimatesFromAnyGraveyard(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.FlipEntered(t, g, nezumiGraverobberRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.WantFlipped(t, g, id, "Nighteyes the Desecrator", []string{"B"}, 2)
	theirs := effectstest.GraveCard(opp, "Their Dead Bear", "Creature — Bear")
	mine := effectstest.GraveCard(me, "My Dead Bear", "Creature — Bear")

	for _, target := range []struct {
		id    uuid.UUID
		label string
	}{{theirs, "an opponent's"}, {mine, "my own"}} {
		effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target.id)})
		c := effectstest.FindBattlefieldCardForTest(g, target.id)
		if c == nil || c.Controller != me.ID {
			t.Errorf("the creature from %s graveyard did not come back under my control: %+v", target.label, c)
		}
	}
}
