package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0107_pr6_damage_cards_test.go — ADR 0107 PR 6 (#1853): the cards
// "damage can't be prevented" unblocks, each against a prevention
// effect it must get through.

const (
	Pr6SkullcrackOracle     = "a6fc0b0c-9ba2-47f2-ab92-26fd20d0f86d"
	Pr6CallInOracle         = "ca54e14c-67ee-4de2-bc0f-19a5f9461d37"
	Pr6CombustOracle        = "f1ba56e3-3d09-4610-88e7-8681b5736105"
	Pr6PinpointOracle       = "d619cce4-036a-4da2-ae64-f43fc095ebe9"
	Pr6ArrowStormOracle     = "a0ce2f5a-9c0b-4801-9105-72f84d8a4a1f"
	Pr6UrzasRageOracle      = "363f8c66-fe0c-44b9-987d-1d160e3f9c54"
	Pr6LavaBurstOracle      = "f7be3da5-55b2-46f2-a5aa-277dee242b94"
	Pr6FearFireFoesOracle   = "35a0e085-df04-454d-b4d3-42db387bbad0"
	Pr6FlamesOracle         = "4e9df979-c1c2-4de1-944e-c5e2d782e66e"
	Pr6FlaringPainOracle    = "cad19ad2-3ea9-43f5-bb5a-b0793402fab0"
	Pr6ImpracticalOracle    = "c8da73ed-2834-4094-8723-a4ad75660290"
	Pr6WildSlashOracle      = "f76a0e5b-74c1-4ee8-b502-e8b988086de4"
	Pr6UnstableOracle       = "fb901095-2ea4-4b9e-ad5a-2892894dc4c3"
	Pr6ExcruciatorOracle    = "9911c2f8-fafb-4979-898e-e86b726e3538"
	Pr6MalignusOracle       = "f90174b3-effc-4c81-8cb0-856a0646516d"
	Pr6LeylineOracle        = "2608df54-dfbe-417d-aef5-49afbdfb03da"
	Pr6SunspineOracle       = "479adf64-f42c-4cb6-bf25-648c044bca8b"
	Pr6BalothOracle         = "6a4ef075-9254-4d36-9572-622833fae54e"
	Pr6GearhulkOracle       = "c4c6083a-b0b5-4c6b-9f11-9f4fe38417c2"
	Pr6QuestingBeastOracle  = "b685757b-521e-4353-a233-97052359723d"
	Pr6LightningSurgeOracle = "dc55e69e-e1b8-4129-902c-c71bcb952418"
)

// pr6ProRed puts a creature with protection from red on the
// battlefield for `owner`. Used only as an untargeted damage victim: a
// red spell can't target it (CR 702.16b).
func Pr6ProRed(g *game.Game, owner uuid.UUID, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pro-Red Knight", TypeLine: "Creature — Knight",
		Power: 2, Toughness: toughness, Colors: []string{"W"}, Keywords: []string{"protection from red"},
		Owner: owner, Controller: owner,
	})
}

func Pr6Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func Pr6Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
