package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_gl_cards_test.go — Bloomburrow cards starting with G and I to L
// (#4467). Each test plays the real catalog entry and asserts what its
// text says.

const (
	BlbGLMooseOracle        = "b0747a4f-d041-4750-aef5-a5d101696f65"
	BlbGLGevOracle          = "77b832a8-a3af-44fe-bc67-72510d292667"
	BlbGLGlidediveOracle    = "33c06c11-8d0a-45ce-adea-c9a7a8ac8779"
	BlbGLRabbitOracle       = "7ed4afc9-959a-4868-8f4a-2221bea1ba32"
	BlbGLVinelasherOracle   = "4f8c4927-46c2-4c18-b5af-c6288dbaff6d"
	BlbGLBruiserOracle      = "0a2fbdd3-4940-4bbf-ae19-f6d00d7ff294"
	BlbGLKindlesparkOracle  = "980cd7d5-a511-49ab-bc52-08b71c65db76"
	BlbGLKitnapOracle       = "a1bd1e83-40ac-449f-838c-ef308f275a5d"
	BlbGLKitsaOracle        = "16414460-11ce-481d-ab54-6f0e0ea7d83d"
	BlbGLKnightfisherOracle = "cb2d222c-df42-41a7-a517-0a41e47a4280"
	BlbGLLightshellOracle   = "ef8fab12-d7d2-4c21-8b8c-41627048ac8b"
	BlbGLMentorOracle       = "20194297-f2a4-4473-95bc-7c6835e5b6b3"
	BlbGLBrawlOracle        = "adc8976a-0871-4fd6-9c4c-822ba9549b4b"
	BlbGLVillageOracle      = "b6c7c708-5212-4100-b954-b77855b27915"
	BlbGLRabbitTypeLine     = "Creature — Rabbit Soldier"
	BlbGLVinelasherType     = "Creature — Lizard Assassin"
)

// blbGLCastCreature casts a creature with real P/T from the active
// player's hand, paying the optional costs at `optional`, and resolves
// it, leaving its entry triggers on the stack.
func BlbGLCastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, optional []int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{OptionalCosts: optional}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

// blbGLSettleTriggerOrder answers an open trigger-order prompt for
// chooser in the order offered (CR 603.3b), if there is one. It passes
// no priority, so nothing else moves.
func BlbGLSettleTriggerOrder(g *game.Game, chooser uuid.UUID) {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			_ = g.ResolveTriggerOrder(c.ID, chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...))
			return
		}
	}
}
