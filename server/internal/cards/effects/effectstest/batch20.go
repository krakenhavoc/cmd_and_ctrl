package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch20_test.go — card-level coverage for the card-coverage
// roadmap's batch 20 (#313, `edhrec_rank` 2133–2234): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b20-prefixed.

const (
	B20HornOfGreedOracle          = "b8181d53-1954-4f46-8670-8696440208e8"
	B20FierceEmpathOracle         = "5104053a-d394-4b00-82a4-60fd9f051a6e"
	B20TradingPostOracle          = "63788566-e25a-44bb-bb55-197e1b93b3e8"
	B20TributeMageOracle          = "c2b51295-8820-425f-a517-55231123c7de"
	B20GruulGuildgateOracle       = "d38476e9-2e47-4c0c-8129-483c0bd09ec0"
	B20KwainOracle                = "ce6a85b5-e249-4582-a6be-ce200da5fd53"
	B20GrappleWithThePastOracle   = "d14c633d-977a-4f78-85c2-88582b3c3670"
	B20FontOfMythosOracle         = "7194a262-5e7a-4c12-b271-2bc5e0799477"
	B20MasterOfEtheriumOracle     = "181e6b3e-c33e-49e7-acb8-67473289a856"
	B20ContaminatedAquiferOracle  = "c27b771d-b5ec-459a-a101-f078cb8d0184"
	B20GreedOracle                = "1ff62220-be95-4901-b8d8-812b9a1a1b0a"
	B20DesolateLighthouseOracle   = "aa6dbdf2-2379-4ff5-8a6c-70258784dc35"
	B20VerdantSunsAvatarOracle    = "6ee71432-a44d-495e-86ea-d495d891c331"
	B20GrindingStationOracle      = "0fcd476f-4db8-4293-9388-1678a0043c9e"
	B20PrimevalBountyOracle       = "f633c16d-d943-421c-ad18-af35db9ec9fc"
	B20CemeteryReaperOracle       = "70fa209f-5e79-44de-81ce-1d8d0d8c1006"
	B20ProsperOracle              = "1e9b0fd6-5aae-401a-8df5-d94ce6442696"
	B20TimelessLotusOracle        = "fec77e0a-433e-4d32-9260-e26a37f6ad23"
	B20CircuitMenderOracle        = "1665ca9f-176d-40f1-a4e9-42da4f1236e9"
	B20StarfallInvocationOracle   = "7024532b-f99b-43a7-b0ed-5b3e7ec7592b"
	B20MercilessExecutionerOracle = "c3c45d50-9038-41df-bb2f-9bc40071845b"
	B20TriplicateTitanOracle      = "56e05b17-aa1d-483d-857c-4cab4f12aa8a"
	B20AssembleTheLegionOracle    = "6f81bef3-6ca0-4cf4-aa99-7b2813eeef04"
	B20JetmirOracle               = "da72a4bc-ce6f-4b72-bc66-2ee33cfa87df"
	B20WallOfBlossomsOracle       = "ef4d5fb3-70a3-433d-a9d3-18b2beb8d79f"
	B20ManabarbsOracle            = "0f1afedd-c60f-454f-b84a-c8117aec0128"
	B20BrokenBondOracle           = "858e12e9-3eaa-40cf-9e22-f9ccdfe485b3"
)

// b20Lives snapshots every seat's life total.
func B20Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b20Hands snapshots every seat's hand size.
func B20Hands(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Hand.Size()
	}
	return out
}

// b20HandCard puts a card in a player's hand.
func B20HandCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// b20CastCreature puts a creature with a printed P/T in a player's
// hand and casts it from a main phase of their turn.
func B20CastCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b20Tapped reads a battlefield card's tapped state, failing when it
// is not there.
func B20Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}
