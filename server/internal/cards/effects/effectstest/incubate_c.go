package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incubate_c_test.go — the incubate cards that needed small pieces
// (#3568, slice inc-c): the repeat on Incubate, the Incubator-token
// predicate, "transform all Incubator tokens", and the seven cards on
// them.

const (
	EleshNornEtchingsOracle  = "650ed75c-20b0-45b8-a4d4-813e0369aaf1"
	ElvishVatkeeperOracle    = "b881d61c-d7cf-49af-bed7-3bc657ea1208"
	ProgenitorExarchOracle   = "4052f710-6ba7-406f-8fb7-653a684a80e0"
	glissaHeraldOracle       = "9b10e0f7-9414-4769-85d7-c1cd1c74a9a6"
	GlisteningDawnOracle     = "13ee85c3-a16a-41ac-bbaa-c7549c700a03"
	NornsInquisitorOracle    = "89a59a6d-e9d4-405c-8518-4da86dae30d1"
	SunderTheGatewayOracle   = "88d49309-e7d7-486d-8696-cd66bec2fe33"
	IncCPhyrexianCreatureTyp = "Creature — Phyrexian Test"
)

// makeIncubator puts an Incubator with n +1/+1 counters under owner's
// control through the real creation path.
func MakeIncubator(t *testing.T, g *game.Game, owner uuid.UUID, n int) uuid.UUID {
	t.Helper()
	opts := game.TokenEntryOptions{}
	if n > 0 {
		opts.Counters = map[string]int{"+1/+1": n}
	}
	var ids []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		ids, err = g.CreateTokensForEffect(owner, effects.IncubatorToken(), 1, opts)
	})
	if err != nil || len(ids) != 1 {
		t.Fatalf("make an Incubator: %v (%d tokens)", err, len(ids))
	}
	return ids[0]
}

// pushLands puts n basic lands under owner's control.
func PushLands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: owner, Controller: owner,
		})
	}
}

func SameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func FaceUp(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.ActiveFace
		}
	}
	return -1
}

func ActivateOn(t *testing.T, g *game.Game, who, source uuid.UUID, idx int, target uuid.UUID) error {
	t.Helper()
	params := game.ActivateAbilityParams{Strict: true}
	if target != uuid.Nil {
		params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	}
	return g.ActivateCatalogAbility(who, source, idx, params)
}

// --- Glissa, Herald of Predation -------------------------------------

// glissaAtCombat puts Glissa under seat 0's control and walks to the
// beginning of combat, returning the open mode prompt.
func GlissaAtCombat(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	PushCatalogPermanent(g, me.ID, "Glissa, Herald of Predation", "Legendary Creature — Phyrexian Zombie Elf", glissaHeraldOracle, false)
	AdvanceTo(t, g, game.StepBeginCombat)
	pick := LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt at the beginning of combat: %+v", g.PendingChoices)
	}
	return pick
}
