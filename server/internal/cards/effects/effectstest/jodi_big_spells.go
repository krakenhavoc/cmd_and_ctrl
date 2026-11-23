package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jodi_big_spells_test.go — S58 PR 5, the Jodi deck's big spells and
// the Eldrazi that need no annihilator. Each card is driven through the
// real cast, trigger or loyalty path and its result read off the game.

const (
	BeaconOfTomorrowsOracle   = "85909caa-d2ad-4487-b9b7-6ac60a14b833"
	BringerBlackDawnOracle    = "a9f98ad5-e8d4-4142-a19f-3368b85d1c95"
	BringerBlueDawnOracle     = "7d4676fa-4bfe-43b4-9241-c2a33214852a"
	ConfluxOracle             = "32ebb029-9f03-49c9-b9bb-cb1954e2a324"
	DamiaOracle               = "eb7ae3e8-0489-4f3a-8343-7e7ec2fe3f01"
	EmergentUltimatumOracle   = "0eecdfb3-3b05-4051-a660-060ff6df80ef"
	NexusOfFateOracle         = "6c1d22d4-f28e-4041-a9b6-1575e8929b61"
	NicolBolasOracle          = "cb33e07b-9599-4979-991f-df0c43ddac31"
	RiseOfTheEldraziOracle    = "982f70af-077f-40df-b2fe-7c80ebcc7712"
	SheoldredWhisperingOracle = "9218b56d-aaec-482f-99e9-d95d227bfe25"
	SunbirdsInvocationOracle  = "10d482fd-e034-4187-b148-51fe15885360"
	UginsBindingOracle        = "ed622e71-f348-4e46-8fb9-05aae430983a"
	UlamogCeaselessOracle     = "0bfa4512-e35a-4c93-b324-80ec659f5a97"
	VorinclexVoiceOracle      = "dbf0ad03-ab31-49d2-89b1-05b45948a61f"
	ZendikarResurgentOracle   = "a0af158b-f42a-4022-a438-e0a2e1fd6d9e"
)

// libraryHolds reports whether `id` is somewhere in `p`'s library.
func LibraryHolds(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// libraryIndex is the card's position in the library (the top is the
// last element), or -1.
func LibraryIndex(p *game.Player, id uuid.UUID) int {
	for i, c := range p.Library.Cards {
		if c.InstanceID == id {
			return i
		}
	}
	return -1
}
