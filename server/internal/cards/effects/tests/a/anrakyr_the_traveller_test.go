package testsa

import (
	"slices"

	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const anrakyrOracle = "32476743-c9d4-49ca-bec2-0669c215841b"

func TestAnrakyrTheTravellerIsRegisteredFull(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, anrakyrOracle, "Anrakyr the Traveller")
}

// anrakyrAttack puts Anrakyr, an artifact in hand, an artifact in the
// graveyard and a sorcery in hand in place, attacks, and resolves the
// trigger up to its offer.
func anrakyrAttack(t *testing.T) (g *game.Game, me *game.Player, handArt, yardArt, sorcery uuid.UUID, prompt *game.PendingChoice) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	anrakyr := effectstest.PushCatalogPermanent(g, me.ID, "Anrakyr the Traveller", "Legendary Artifact Creature — Necron", anrakyrOracle, false)
	handArt = effectstest.HandSpell(me, "Hand Relic", "Artifact", "{3}")
	sorcery = effectstest.HandSpell(me, "Hand Sorcery", "Sorcery", "{1}")
	yardArt = uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: yardArt, Name: "Yard Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{5}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(anrakyr, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.ResolutionCastOfferFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer to cast an artifact spell")
	}
	return
}

// The offer names the artifact cards in the hand and the graveyard, not
// the sorcery. Casting the graveyard artifact costs life equal to its
// mana value, and spends the one cast: the hand artifact is no longer
// castable this way.
func TestAnrakyrCastsAnArtifactFromTheGraveyardForLife(t *testing.T) {
	g, me, handArt, yardArt, sorcery, prompt := anrakyrAttack(t)
	if !slices.Contains(prompt.DiscardOptions, handArt) || !slices.Contains(prompt.DiscardOptions, yardArt) {
		t.Fatalf("offer %v, want both artifact cards", prompt.DiscardOptions)
	}
	if slices.Contains(prompt.DiscardOptions, sorcery) {
		t.Fatal("the sorcery is offered")
	}
	life := me.Life
	if err := g.CastForResolutionChoice(me.ID, prompt.ID, yardArt, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "anrakyr_the_traveller"}); err != nil {
		t.Fatalf("cast from the graveyard for life: %v", err)
	}
	if me.Life != life-5 {
		t.Errorf("life %d -> %d, want 5 paid", life, me.Life)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("the offer is still open after the one cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, handArt, game.CastSpellParams{AlternativeCost: "anrakyr_the_traveller", Strict: true}); err == nil {
		t.Error("a second artifact was cast for life")
	}
}

// From the hand, the same: life equal to the mana value, no mana.
func TestAnrakyrCastsAnArtifactFromTheHandForLife(t *testing.T) {
	g, me, handArt, _, _, prompt := anrakyrAttack(t)
	life := me.Life
	if err := g.CastForResolutionChoice(me.ID, prompt.ID, handArt, game.CastSpellParams{AlternativeCost: "anrakyr_the_traveller"}); err != nil {
		t.Fatalf("cast from the hand for life: %v", err)
	}
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want 3 paid", life, me.Life)
	}
}

// Declining leaves both cards where they are.
func TestAnrakyrDeclineLeavesTheCards(t *testing.T) {
	g, me, handArt, yardArt, _, _ := anrakyrAttack(t)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if z := effectstest.E2Zone(g, handArt); z != game.ZoneHand {
		t.Errorf("the hand artifact is in %q", z)
	}
	if z := effectstest.E2Zone(g, yardArt); z != game.ZoneGraveyard {
		t.Errorf("the graveyard artifact is in %q", z)
	}
	if err := g.CastSpell(me.ID, yardArt, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "anrakyr_the_traveller", Strict: true}); err == nil {
		t.Error("cast from the graveyard after declining")
	}
}
