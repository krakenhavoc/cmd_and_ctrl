package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch18_test.go — card-level coverage for the card-coverage
// roadmap's batch 18 (#311, `edhrec_rank` 1930–2031): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play, attack, block or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b18-prefixed.

const (
	B18DreamstoneHedronOracle   = "0e2575be-c596-4c8c-bf07-7941ca065721"
	B18SpringleafParadeOracle   = "b1305916-53cc-4021-897e-bbefc65dce78"
	B18StormfistCrusaderOracle  = "0d3bebc6-662a-4cd4-ad1e-aeed0c5e04f6"
	B18ChartACourseOracle       = "05878e49-93ad-4144-9c50-a0bb86126c2e"
	B18WoundReflectionOracle    = "b09206a4-8b73-4125-8b79-53f6fd511b16"
	B18SpectralSailorOracle     = "a8fdbcdf-479d-4582-9ad5-9fbd4c740c29"
	B18GrazilaxxOracle          = "d22ff377-d282-4a28-9dce-96f25913dc96"
	B18KorSpiritdancerOracle    = "2bfc7467-1316-43d3-9ca6-ff36cc2de607"
	B18ManyPartingsOracle       = "92ea750e-62b7-4422-a527-ddf435759d08"
	B18ObNixilisTheFallenOracle = "154104b7-4125-4276-8bc7-9a6edfe48cb9"
	B18MikaeusTheLunarchOracle  = "82f3faa8-39fa-450b-843f-d60a4c36d8f7"
	B18OgreSlumlordOracle       = "0a5e3748-2e58-4e53-9653-8af4e21cf223"
	B18HelpfulHunterOracle      = "c0864adb-e9aa-40b6-91ff-a0646193e887"
	B18OrzhovGuildgateOracle    = "57b37df5-fee4-4720-931f-f0cb0a8b338c"
	B18JunkDiverOracle          = "08c595da-9305-42a9-b72f-5ccc546edc01"
	B18VenserShaperSavantOracle = "0f41cefc-d6ff-4db7-ba35-502b7e081de1"
	B18SplitUpOracle            = "2e82520a-9da3-49ae-b5c8-37e7ac8853fe"
	B18HerosDownfallOracle      = "03df6a57-37c9-46d3-83b3-4a6240100714"
	B18CycleOfRenewalOracle     = "77eda626-5cef-496d-b291-cfa1ee6fc1fb"
	B18EurekaMomentOracle       = "0e2c11b2-d95f-4402-9a4a-afd3f7ffb8be"
	B18FalkenrathNobleOracle    = "3739b179-bc81-4737-8376-66a57e16b942"
	B18BitterbloomBearerOracle  = "5effb651-f9e0-4c14-8192-bd0e132b8d5d"
	B18EmeriaAngelOracle        = "ea6616e4-db8d-4905-80f8-cb0162906850"
	B18KederektParasiteOracle   = "fc7b46af-6c07-455a-99c7-f1bccaa2a5f4"
	B18WindsOfAbandonOracle     = "499a4715-6776-461b-bee4-22ed172d455e"
	B18MagdaTheHoardmasterOrcl  = "7fd2beb8-f823-4723-beec-e59b62127490"
	B18SangromancerOracle       = "920445ab-0ac2-4de7-bc1c-f5e58eb4424c"
	B18ThunderfootBalothOracle  = "f55334d9-1ec2-4667-bf28-5a0384d6f053"
	B18ArixmethesOracle         = "caeb39ee-f0bb-4305-9e8b-b30ba0a74c78"
	B18RakdosGuildgateOracle    = "361f534b-39d1-4421-b5a8-d3813c62f86d"
	B18GarruksPackleaderOracle  = "13279222-422d-4447-9451-2463b0c714c6"
	B18PerpetualTimepieceOracle = "17b4778b-82b1-4845-ad08-00f3ff66877b"
	B18WoodedBastionOracle      = "61b85077-64aa-4bcc-890d-2d88da9543c0"
	B18ThirstForKnowledgeOracle = "939e6f71-185e-41f2-9d54-72cce06f1dce"
	B18RighteousValkyrieOracle  = "891d2690-7144-4f87-b6ef-96f2469780a9"
	B18KrenkosCommandOracle     = "9cfe86ae-eebe-44aa-a956-4b3e9e621105"
)

// b18Counters reads a battlefield card's counters of one kind.
func B18Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[kind]
}

// b18Kill destroys a battlefield permanent through the single-target
// verb, under the lock, and settles what it triggered up to any
// prompt.
func B18Kill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	PassPriorityAroundTable(t, g)
}

// b18CastFromHand puts a card in a player's hand and casts it with
// the given params from a main phase of the active seat.
func B18CastFromHand(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b18TriggerPromptCount counts the open yes/no trigger prompts
// addressed to a player.
func B18TriggerPromptCount(g *game.Game, chooser uuid.UUID) int {
	return OwedMays(g, chooser)
}
