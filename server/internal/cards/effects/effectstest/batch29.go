package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch29_test.go — card-level coverage for the card-coverage
// roadmap's batch 29 (#391, `edhrec_rank` 3041–3144): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, damage, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b29-prefixed.

const (
	B29BlackWaltzNo3Oracle        = "9c84c681-40c1-4ff3-86e6-48b4ab7491e5"
	B29HeartlessHidetsuguOracle   = "b68e965a-96f4-447a-99be-35dbd27846e9"
	B29SettleTheWreckageOracle    = "2a2ea189-b663-4f4a-bb23-ff7a4af25f71"
	B29LegionLieutenantOracle     = "5fa1b2f0-3ba4-49db-9cb5-b6130e4c255e"
	B29SazhsChocoboOracle         = "c2b082b4-5d9c-4125-b353-d1cf4f334720"
	B29YouthfulValkyrieOracle     = "6d37ba4b-ff56-4eec-9dc2-2d7f357dc9c9"
	B29ElvishPromenadeOracle      = "7df3e379-c217-416e-a1c8-46338608c49e"
	B29WorldMapOracle             = "51d4f95e-aad2-4e5f-9f15-a68bbdf9ab3e"
	B29DazzlingAngelOracle        = "7de464e3-fae3-44cc-8233-776fc727c00a"
	B29HammerOfPurphorosOracle    = "212d058b-69c6-4dc2-8c93-bdfe26dc2ffe"
	B29CorneredByBlackMagesOracle = "68a58e0b-1506-492e-8ca3-016e520c10ba"
	B29EmpyreanEagleOracle        = "270d14b2-07bc-46bc-918f-658102265ccf"
	B29MassHysteriaOracle         = "4500131b-7417-4f30-a1b0-97d51b2e6458"
	B29BlowflyInfestationOracle   = "3859873f-2a0f-463b-8ed9-7f2ab5ed393a"
	B29ZombieApocalypseOracle     = "8241277d-654f-4985-9d49-a22c1e59eec2"
	B29LightPawsOracle            = "1718a442-b878-4690-b608-a013de3d79fc"
	B29ZiatoraOracle              = "d46c3fb6-f1c7-4a96-ae42-5bce17fc7c1d"
	B29WitchsOvenOracle           = "8fa0fe02-2452-4386-8e0c-165757b0f0a3"
	B29VedalkenArchmageOracle     = "568cf486-0261-4634-ac36-a6507101b2d0"
	B29ScytheclawRaptorOracle     = "8c653a0d-e35a-4596-bed6-b5156192955b"
	B29NayaCharmOracle            = "20cc0f3e-de69-4e5f-88b3-ae7e6c6b5996"
	B29HallowedHauntingOracle     = "e310ab58-180e-4840-bc8d-9f9b06f1b478"
	B29MountainValleyOracle       = "0b7393aa-d563-45bc-9946-8e7d1729d498"
	B29YargleAndMultaniOracle     = "980e1721-3717-4425-95a4-938dbf2ac661"
	B29ChitterspitterOracle       = "4a338863-d599-46e7-9c30-e11b898ae1b0"
	B29GimliOracle                = "f6747295-d21d-40d7-bc70-baa4a37ae668"
)

// b29Push seeds a permanent with a mana cost, colours and P/T on the
// battlefield with a layer timestamp and no summoning sickness.
func B29Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Colors: colors, Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b29Lives reads every seat's life total.
func B29Lives(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Life)
	}
	return out
}

// b29CountNamed counts battlefield permanents `controller` controls
// with the given name.
func B29CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// b29CastAs has `caster` cast a card from hand right now, at
// whatever step the game is at — for an instant cast during combat
// or on another player's turn.
func B29CastAs(t *testing.T, g *game.Game, caster *game.Player, name, typeLine, oracle, manaCost string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// pushGraveyardCardTyped seeds a card with a type line into a
// player's graveyard.
func PushGraveyardCardTyped(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}
