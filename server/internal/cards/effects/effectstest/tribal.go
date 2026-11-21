package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tribal_test.go — S26 (#78). The lords, the named-tribe permanents,
// Coat of Arms and the changeling interactions.

const (
	CavernOfSoulsOracle     = "89ca686a-7c72-4d8f-9290-e89635624a83"
	DoorOfDestiniesOracle   = "9b6d3dcf-aa5a-4516-bd48-a0723e86bfd1"
	VanquishersBannerOracle = "8cf38025-5821-45a0-9483-266353b7e82d"
	AdaptiveAutomatonOracle = "53c730c6-2f8c-4af8-b400-b9d573a71e60"
	CoatOfArmsOracle        = "5f7f133e-58ea-41ab-b1be-be4b400fac4c"
	MaskwoodNexusOracle     = "9b2cdbed-c733-409b-b0e4-2c8960c25111"
	ElvishArchdruidOracle   = "6e2c2423-d854-4478-99e6-64f29851f026"
	ElvishChampionOracle    = "7e40f37c-9a0c-40e0-b195-7ea94b12f798"
	GoblinChieftainOracle   = "368b4052-174e-4458-a6e6-eaf8093aa0fe"
	DeathBaronOracle        = "99024aa8-5687-4d38-8a4b-feef42d6c1ff"
	IrregularCohortOracle   = "c0636d16-671c-4e80-af8c-67d80d2cd979"
	ShieldsOfVelisVelOracle = "7ad6be4e-5c3c-4633-a641-beb06e4129b9"
)

// pushTribalCreature seeds a creature on the battlefield under
// `owner` with the layer listener fired, and returns its ID.
func PushTribalCreature(g *game.Game, owner uuid.UUID, name, typeLine string, p, t int, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		Power:      p,
		Toughness:  t,
		Keywords:   keywords,
		Owner:      owner,
		Controller: owner,
	})
}

// pushNamedTribePermanent seeds a catalog permanent and answers its
// "choose a creature type" prompt with `tribe`.
func PushNamedTribePermanent(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracleID, tribe string) uuid.UUID {
	t.Helper()
	id := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      owner,
		Controller: owner,
	})
	// The OnETB hook runs off the real resolution path; seeding the
	// battlefield directly skips it, so the choice is stamped the
	// same way the resolve path would.
	var choiceID uuid.UUID
	g.WithWriteLock(func() {
		choiceID = g.QueueCreatureTypeChoiceForEffect(owner, id, name)
	})
	if err := g.ResolveCreatureTypeChoice(choiceID, owner, tribe); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice(%s, %s): %v", name, tribe, err)
	}
	return id
}

func HasAbility(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
