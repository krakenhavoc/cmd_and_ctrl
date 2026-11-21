package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gitrog_fortune_saddle_test.go — #3547, ADR 0186 slice 5: the two cards
// that read "a creature that saddled it this turn" after the saddle
// record landed (slice 1).

const (
	GitrogRavenousRideOracle = "7739311a-682b-48a6-8db2-c07f44141751"
	FortuneLoyalSteedOracle  = "dc440b62-0bf0-4f23-9210-3550ff0a8d4d"
)

// pushHandLands puts n basic lands into the player's hand and returns them.
func PushHandLands(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Test Forest", TypeLine: "Basic Land — Forest",
			Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	return ids
}

// gitrogConnects saddles The Gitrog with `saddler`, attacks the second
// seat and runs combat damage until the sacrifice prompt is up.
func GitrogConnects(t *testing.T, g *game.Game, me, opp *game.Player, gitrog, saddler uuid.UUID) *game.PendingChoice {
	t.Helper()
	ToMain(t, g)
	if err := SaddleAbility(t, g, me.ID, gitrog, saddler); err != nil {
		t.Fatalf("saddle: %v", err)
	}
	PassPriorityAroundTable(t, g)
	DeclareAttack(t, g, opp.ID, gitrog)
	AdvanceTo(t, g, game.StepCombatDamage)
	PassPriorityAroundTable(t, g)
	choice := LatestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt for the creature that saddled The Gitrog")
	}
	return choice
}

// Fortune's returned ETB scries; answer it if it is up.
func AnswerScryIfAsked(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	PassPriorityAroundTable(t, g)
	if ScryChoiceFor(g, me.ID) != nil {
		AnswerScryKeepAll(t, g, me.ID)
	}
}
