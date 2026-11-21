package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_additional_cost_cards_test.go — #3868: "As an additional cost to
// cast this spell, tap an untapped Vampire you control" (CR 601.2h), the
// activated abilities' TapOthersCost on a cast's mandatory slot. The
// permanents ride cast_spell's teamwork_ids; they are untapped, yours,
// matched without targeting, with no summoning-sickness check (CR 302.6),
// and tapped with the spell on the stack.

const (
	NewBloodOracle        = "021d654e-b731-40d0-a18e-eaff43f9fdb9"
	SharedDiscoveryOracle = "dd8e9f28-39da-4336-9259-0e34358749d9"
	GazeOfJusticeOracle   = "07c7e4d4-6452-437b-9f17-b0ec1fd9f702"
	SwallowWholeOracle    = "54d4c632-1e7c-42f1-938c-9b4bd9b9a9b9"
	GuardianGreatDoor     = "18ba355d-a25f-4a5d-b734-11739b0b86af"
	WarlordsEliteOracle   = "ee2c8714-6e85-4eb5-9bd6-057b6abfb136"
	FearOfExposureOracle  = "abec56eb-7b3d-4bd8-93c7-bd7049c859c3"
	DisruptionProtocol    = "4ca3306b-06f7-4a09-b904-f5540d11ddf1"
)

// tacVampire puts an untapped Vampire onto `owner`'s battlefield.
func TacVampire(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Vampire", Power: 2, Toughness: 2, Colors: []string{"B"}})
}

func TacCast(g *game.Game, p *game.Player, card uuid.UUID, targets []game.TargetRef, taps ...uuid.UUID) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{Targets: targets, TeamworkIDs: taps, Strict: true})
}
