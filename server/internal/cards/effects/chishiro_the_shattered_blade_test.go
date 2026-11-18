package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const chishiroOracle = "c98a4af3-3e58-4927-9a60-d2664cf0b139"

func TestChishiroIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(chishiroOracle)
	if !ok || spec.Name != "Chishiro, the Shattered Blade" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestChishiroMakesASpiritWhenYourAuraOrEquipmentEntersButNotOtherPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Chishiro, the Shattered Blade", "Legendary Creature — Snake Samurai", chishiroOracle, 4, 4)
	bear := auraBear(g, me.ID)
	castCatalogSpell(t, g, "Plain Artifact", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("a plain artifact made %d Spirits", n)
	}
	castCatalogSpell(t, g, "Test Equipment", "Artifact — Equipment", "", nil)
	passPriorityAroundTable(t, g)
	auraCast(t, g, "One with the Kami", oneWithTheKamiOracle, bear)
	passPriorityAroundTable(t, g)
	var spirits []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Spirit") {
			spirits = append(spirits, c)
		}
	}
	if len(spirits) != 2 {
		t.Fatalf("Spirits = %d, want 2 (one for the Equipment, one for the Aura)", len(spirits))
	}
	if s := spirits[0]; s.Power != 2 || s.Toughness != 2 || !game.HasKeyword(&s, "menace") {
		t.Errorf("token = %d/%d, want a 2/2 with menace", s.Power, s.Toughness)
	}
}

func TestChishiroIgnoresAnOpponentsEquipment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	b12Push(g, g.Seats[0].ID, "Chishiro, the Shattered Blade", "Legendary Creature — Snake Samurai", chishiroOracle, 4, 4)
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Equipment", TypeLine: "Artifact — Equipment", Owner: opp.ID, Controller: opp.ID})
	// Put it onto the battlefield under the opponent's control through a real entry.
	id := opp.Hand.Cards[len(opp.Hand.Cards)-1].InstanceID
	advanceToMain(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("move: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, g.Seats[0].ID, "Spirit"); n != 0 {
		t.Errorf("an opponent's Equipment made %d Spirits", n)
	}
}

func TestChishiroEndStepGrowsOnlyModifiedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chishiro := b12Push(g, me.ID, "Chishiro, the Shattered Blade", "Legendary Creature — Snake Samurai", chishiroOracle, 4, 4)
	plain := bigBear(g, me.ID, 2, false)
	mod := modifiedBear(g, me.ID)
	theirs := bigBear(g, g.Seats[1].ID, 2, true)
	var done uuid.UUID
	_ = done
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, plain, game.CounterPlusOne); n != 0 {
		t.Errorf("plain bear counters = %d", n)
	}
	if n := countersOn(g, mod, game.CounterPlusOne); n != 2 {
		t.Errorf("modified bear counters = %d, want 2", n)
	}
	if n := countersOn(g, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("opponent's bear counters = %d, want unchanged 1", n)
	}
	if n := countersOn(g, chishiro, game.CounterPlusOne); n != 0 {
		t.Errorf("unmodified Chishiro counters = %d, want 0", n)
	}
}
