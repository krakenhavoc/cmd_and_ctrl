package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	BeastWithinOracle       = "7735eeba-693b-47e2-bd51-414379cf1016"
	GenerousGiftOracle      = "fae37e28-e137-4177-b973-fa8b4dd8f409"
	KrosanGripOracle        = "3e39224c-72ce-4ecc-aa17-12c071ea1f3e"
	NaturesClaimOracle      = "6d4e558e-9109-4918-a082-fdcbaffd516b"
	AnguishedUnmakingOracle = "ad09b3c3-c8e7-481c-8c45-e7f234935117"
	BlasphemousActOracle    = "7a2484a9-04fd-41a0-8224-610c1c07ed10"
	RampantGrowthOracle     = "8539f295-5d58-4436-a73a-b9277c4c7795"
	FarseekOracle           = "495e52e6-4c2b-4574-9474-eadbdcc8b4ac"
	NaturesLoreOracle       = "78826359-fe63-44ad-adc4-a17ffcd710e4"
	KodamasReachOracle      = "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90"
	WoodElvesOracle         = "8973bd99-20f8-4867-90ef-50392147ee1b"
	ThranDynamoOracle       = "a699c663-8131-4045-9265-a83e86609374"
	WornPowerstoneOracle    = "b166b670-febc-4821-855e-f8d465644c03"
	CommandTowerOracle      = "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"
	MindStoneOracle         = "c97361b5-af16-4a7b-af85-a429dbaf4ad2"
	NightsWhisperOracle     = "7ffae8f8-3006-4969-a339-6d30678f87ea"
)

// seedPermanentFor puts a permanent on the battlefield under owner's
// control and returns its instance ID.
func SeedPermanentFor(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// countBattlefieldNamed counts battlefield cards named `name` under
// controller. Tokens have no stable ID to assert against, so the
// token-producing cards check by name + controller instead.
func CountBattlefieldNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			n++
		}
	}
	return n
}

// battlefieldCard finds a battlefield card by instance ID.
func BattlefieldCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}
