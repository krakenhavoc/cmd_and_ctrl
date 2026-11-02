package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_batch2_crew_test.go — Marvel's Spider-Man cards, batch 2 (#4460).

// spm2Graveyard seeds a graveyard card with a mana cost.
func spm2Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

const damageControlCrewOracle = "c0f9e604-a8bf-4c7d-8480-487e8076e1d8"

func TestSpm2DamageControlCrewRepairReturnsACardOfManaValueFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := spm2Graveyard(me, "Big Spell", "Sorcery", "{2}{G}{G}")
	small := spm2Graveyard(me, "Small Spell", "Sorcery", "{1}{G}")
	castAndResolveCreature(t, g, "Damage Control Crew", "Creature — Human Citizen", damageControlCrewOracle)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt from the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("choosing Repair: %v", err)
	}
	if err := sgAnswerTargets(t, g, small); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("mana value 2 card: err = %v, want ErrIllegalTarget", err)
	}
	if err := sgAnswerTargets(t, g, big); err != nil {
		t.Fatalf("mana value 4 card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(big) || !me.Graveyard.Contains(small) {
		t.Error("only the mana value 4 card returns to hand")
	}
}

func TestSpm2DamageControlCrewImpoundExilesAnArtifactOrEnchantmentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1-g.Turn.ActiveSeat]
	rock := pushTypedCard(g, opp.ID, "Rock", "Artifact", "{1}")
	bear := pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	castAndResolveCreature(t, g, "Damage Control Crew", "Creature — Human Citizen", damageControlCrewOracle)
	mode := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("choosing Impound: %v", err)
	}
	if err := sgAnswerTargets(t, g, bear); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a creature: err = %v, want ErrIllegalTarget", err)
	}
	if err := sgAnswerTargets(t, g, rock); err != nil {
		t.Fatalf("an artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !g.Battlefield.Contains(bear) {
		t.Error("only the artifact is exiled")
	}
}
