package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aura_host_cards_test.go — CR 303.4f / 303.4g (#3158): an Aura put onto
// the battlefield other than by resolving as an Aura spell chooses what
// it enchants as it enters, and stays where it was when nothing is
// legal. The engine half is game/aura_host.go; these are the cards.

// ahGraveyardAura puts a catalogued Aura card into p's graveyard.
func ahGraveyardAura(p *game.Player, name, oracle, cost, color string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: AuraTypeLine, OracleID: oracle,
		ManaCost: cost, Colors: []string{color}, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func AhPacifism(p *game.Player) uuid.UUID {
	return ahGraveyardAura(p, "Pacifism", PacifismOracle, "{1}{W}", "W")
}

func AhAnimateDead(p *game.Player) uuid.UUID {
	return ahGraveyardAura(p, "Animate Dead", AnimateDeadOracle, "{1}{B}", "B")
}

// auraHostChoiceFor is the open aura_host prompt addressed to chooser.
func AuraHostChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceAuraHost && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func AnswerAuraHost(t *testing.T, g *game.Game, chooser uuid.UUID, host game.TargetRef) {
	t.Helper()
	c := AuraHostChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no aura_host prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveAuraHost(c.ID, chooser, host); err != nil {
		t.Fatalf("ResolveAuraHost: %v", err)
	}
}

// sunTitanReturns casts Sun Titan and returns `card` from the caster's
// graveyard with its enters trigger, stopping at whatever the entry asks.
func SunTitanReturns(t *testing.T, g *game.Game, card uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	titan := CastAndResolveCreature(t, g, "Sun Titan", "Creature — Giant", SunTitanOracle)
	PickCard(t, g, me.ID, card)
	AnswerLatestTriggerPrompt(t, g, me.ID, true)
	PassPriorityAroundTable(t, g)
	return titan
}
