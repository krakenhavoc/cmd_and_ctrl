package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s58_mishra_cards_test.go — S58 PR 4, the artifact cards Mishra's
// deck asked for (#2033).

const (
	S58MishraOracle            = "d3438037-3efd-4ce0-88ec-6d48ab521992"
	S58ArcumDagssonOracle      = "0fdfe986-0216-468b-aac8-bbcf588fd894"
	S58ArtificerClassOracle    = "6dccf583-1045-4058-86cc-ebbcc8de080e"
	S58CopyArtifactOracle      = "80bc56a9-40e0-48da-ae86-190e39c8a4a3"
	S58CybermenSquadronOracle  = "8d035688-089f-4d5c-bcad-9140a6c1681b"
	S58DiplomaticImmunityOrcl  = "f1a3153c-0200-4ff2-b7d5-48d23920bb3c"
	S58ForsakenMonumentOracle  = "7777fab1-df3f-467f-b9e2-46dd2bd2166e"
	S58KnightPaladinOracle     = "e2ab0f84-8402-472d-be66-9661b85e08fc"
	S58NautiloidShipOracle     = "613a8774-165e-4cf6-ad43-124f9ffc9980"
	S58PortalToPhyrexiaOracle  = "301d1d8e-d3fc-4010-8b29-4724fa0b31cd"
	S58MycosynthGardensOracle  = "03f5c566-825c-4c46-9c01-a2f9b1e70a13"
	S58ArtificerClassTypeLine  = "Enchantment — Class"
	S58MycosynthGardensTypes   = "Land — Sphere"
	S58MishraTypeLine          = "Legendary Creature — Human Artificer"
	S58NautiloidShipTypeLine   = "Artifact — Vehicle"
	S58PortalToPhyrexiaTypeLin = "Artifact"
)

// s58Rock seeds a plain (non-catalog) artifact with a mana cost.
func S58Rock(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// s58Named returns the battlefield cards named `name` that `controller`
// controls.
func S58Named(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

// s58Cast seeds a spell card into the active player's hand and casts it.
func S58Cast(t *testing.T, g *game.Game, c game.Card, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = active.ID, active.ID
	active.Hand.PushTop(c)
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, c.InstanceID, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", c.Name, err)
	}
	return c.InstanceID
}
