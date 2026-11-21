package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// when_paid_test.go — #1716: a target clause widened by an announced
// optional cost (WhenPaid). One test per proof card, each proving the
// printed clause on a plain cast and the wider one on a paid cast.

const (
	CruelAllianceOracle     = "6897f9e0-f654-4c0a-9fda-2ad4e264bf9a"
	BloodchiefsThirstOracle = "4236851b-5366-43a1-bde4-f525b4fbcbce"
	TearAsunderOracle       = "610af0f7-b5e3-43fb-9d02-7c59bd99034c"
	ExpelTheUnworthyOracle  = "4af2e62f-150e-4fd0-98b0-c6e72f5f9a51"
	BloodBeckoningOracle    = "67a48e3f-2388-42a9-a8b1-97c08761f807"
	DivineResilienceOracle  = "4f6e2e47-34df-4bf3-a546-e06b42840167"
)

// wpCreature seeds a creature with a printed mana cost on the
// battlefield, so its mana value is what the clause reads.
func WpCreature(g *game.Game, owner uuid.UUID, name, manaCost string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: manaCost,
		Power: 3, Toughness: 3, Owner: owner, Controller: owner,
	})
}
