package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// role_tokens_part3_test.go — #1945 slice roles-c: the remaining cards
// that create a Role and need no new machinery.

const (
	GylwainOracle        = "e7dfe47f-41fc-4071-9ad3-96cd8d9be52e"
	CurseWerefoxOracle   = "64b41b54-ca22-41df-8bd3-dc4f7832fbb7"
	AsinineAnticsOracle  = "0a0a051c-59ed-4286-b842-a34da7ef15d5"
	DiminisherOracle     = "c2283655-f6b8-4399-ba1f-68a7107f9485"
	ErietteOracle        = "27ca3201-a374-4a9d-b0ba-93b391f36844"
	ShatterOathOracle    = "f545870d-b628-4acb-88c7-37c2d72a5a85"
	TwistedFealtyOracle  = "a5ae75de-0dff-44bf-9dc6-5895ab3ae731"
	SewerWitchOracle     = "99c9920a-f84b-4626-9fe5-d94cc62cd277"
	WitchsMarkOracle     = "00807825-88e5-4b88-8527-a49f553a91be"
	SkittersOracle       = "d5bd40fc-6138-4752-b3ff-b1e5fa269214"
	WitchsVanityOracle   = "ad2793f3-a36f-4a09-a38f-eb98ce0a2b79"
	GadwicksDuelOracle   = "322e3bc1-2dfa-4d5f-848f-a82d9ce02a67"
	GiantInheritOracle   = "6be21185-f85c-49e8-988a-a76b6613fdaa"
	NotDeadOracle        = "380c367f-73ea-485f-a37b-5af0ba160893"
	ScoundrelOracle      = "9eec304d-ffcb-4287-b877-280ac1e4f496"
	BecomeBrutesOracle   = "789e034c-7f4a-4fd5-9dd8-48313742126e"
	BesottedKnightOracle = "c7b6d2cd-9105-404b-88f0-a67037fb2120"
	ConceitedOracle      = "1c751201-24ba-4e5c-bf44-71c3e4693a92"
	WerefoxOracle        = "fef67393-9b7b-495d-8b03-865537386d42"
	TransmuterOracle     = "89ae3475-9063-4700-a7bd-47d5dcd43a3b"
)

func WantRole(t *testing.T, g *game.Game, host uuid.UUID, name string) {
	t.Helper()
	roles := RolesOn(g, host)
	if len(roles) != 1 || roles[0].Name != name {
		t.Fatalf("roles = %+v, want one %s", roles, name)
	}
}

func WantNoRole(t *testing.T, g *game.Game, host uuid.UUID) {
	t.Helper()
	if roles := RolesOn(g, host); len(roles) != 0 {
		t.Fatalf("roles = %+v, want none", roles)
	}
}

// --- Adventures -----------------------------------------------------

type AdventureCase struct {
	Name, Adv, Oracle, Role string
	Faces                   [2]game.Face
}

func AdventureCard(owner uuid.UUID, c AdventureCase) game.Card {
	card := game.Card{
		InstanceID: uuid.New(), OracleID: c.Oracle, Layout: game.LayoutAdventure,
		Owner: owner, Controller: owner,
		Faces: []game.Face{c.Faces[0], c.Faces[1]},
	}
	card.SetFace(0)
	return card
}
