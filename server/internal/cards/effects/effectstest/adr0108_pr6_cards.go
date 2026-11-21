package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr6_cards_test.go — ADR 0108 PR 6 (#1904): the cards the
// not-one-use source shield (preventFromSource), Dark Sphere's half
// shield and the CR 615.7 division unblock, each against the damage it
// must stop and the damage it must not.

const (
	Pr6PayNoHeed        = "71da5acb-78bc-453a-ae28-d22faa5a4e43"
	Pr6AuriokReplica    = "783a62c0-3b06-4250-a1cb-5cf4042611aa"
	Pr6ForgeTender      = "70bb275b-3458-4690-a50f-b231fbf0bccb"
	Pr6Prahv            = "37ff5ba6-0763-4c73-85bf-66856e67b8f3"
	Pr6RithsCharm       = "9dbccdc5-61f7-4a7d-bbbc-5cab393e24f7"
	Pr6DarkSphere       = "397f53f7-f801-4442-a778-2f26ac246b62"
	Pr6HealingGrace     = "e3b22777-4d5b-4f6c-a339-24d204bd007e"
	Pr6SamiteMinistry   = "af669df1-3526-46ca-b90c-19b112b7ba44"
	Pr6ShieldmageAdv    = "57188b3d-567c-4dff-8b93-7cc7a47894be"
	Pr6ProtectiveSphere = "2037c4bf-bf0d-46ab-8b4b-61d9a50c431e"
	Pr6MournersShield   = "34cadd9d-b8e6-422c-b0ec-a2420cd983ad"
)

// pr6Damage deals `n` from `source` to the permanent `to` as one
// instruction, then runs the priority boundary.
func Pr6Damage(t *testing.T, g *game.Game, source, to uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(source, to, n); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
}

func Pr6Marked(g *game.Game, id uuid.UUID) int {
	c := FindBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.DamageMarked
}
