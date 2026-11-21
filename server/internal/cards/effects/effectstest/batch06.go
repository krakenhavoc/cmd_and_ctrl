package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch06_test.go — card-level coverage for the card-coverage
// roadmap's batch 06 (#299, `edhrec_rank` 694–797): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or step. Helpers from
// batch01_test.go / batch04_test.go are reused by name; new ones are
// b06-prefixed.

const (
	B06IntoTheFloodMawOracle   = "8cb36a67-9206-4665-a03f-64f52ba559c4"
	B06GrimTutorOracle         = "e62f8d69-a559-4f13-a5c9-5fb750b4af2c"
	B06CastleArdenvaleOracle   = "f8f4fc60-725d-46d8-8e8f-e68e00d20589"
	B06WhipOfErebosOracle      = "53987a39-c18c-4c13-b1ea-fd1b2a369f9e"
	B06FireDiamondOracle       = "97b477d8-2e05-475e-8ed6-7d680cb21cd9"
	B06RiveteersOverlookOracle = "5548ff43-e5f6-4a63-8562-a2b1de06d6f5"
	B06RisingOfTheDayOracle    = "434a20f8-3f87-4004-9155-0f196fc2257e"
	B06VernalFenOracle         = "40544d12-0391-4a61-af95-9b8ec01ed8fc"
	B06WarstormSurgeOracle     = "42fb1a1c-ab3d-4cdc-a6ff-a591f7481583"
	B06ForbiddenOrchardOracle  = "cfd60d1f-9832-4408-b84e-0fd3018b015b"
	B06OranRiefOracle          = "e88027a6-24cc-4a8b-86db-734f26149ea8"
	B06ElvesOfDeepShadowOracle = "20347559-95a9-4689-bb79-c5bb3809b719"
	B06SpelunkingOracle        = "2962fe4c-bf48-454b-8a6b-0f8253352ae8"
	// #732's two fetch routes: a sorcery's search and an ETB
	// trigger's. Neither card is a batch-06 card; the IDs live here
	// because the Spelunking tests below are the only callers.
	B06CultivateOracle           = "8b755881-a72d-4e21-a369-d2924eb4585a"
	B06SolemnSimulacrumOracle    = "00c0543c-2a1f-4425-8283-4062d74a1637"
	B06ScouredBarrensOracle      = "d37f858e-03c8-4594-9b92-cd03699a1591"
	B06StartingTownOracle        = "d04e0975-f401-41b8-a9db-9bcf9cbbce66"
	B06IdyllicTutorOracle        = "57c9ff89-dd30-467b-bb3e-499eeea8cb94"
	B06BloodchiefAscensionOracle = "d5ea905b-4bb6-48d0-9082-c472703db550"
	B06ReturnToNatureOracle      = "777b8ec4-a783-4297-96b7-4f200d0eb734"
	B06DryadArborOracle          = "e996cd67-739c-40f4-b276-0042acf26c71"
	B06MossbornHydraOracle       = "84122d81-9634-4d3f-85a5-f6cf4303691a"
	B06EladamrisCallOracle       = "4acb6612-54e8-428d-acb6-c7259a5ad6a8"
	B06OrnithopterOracle         = "a3a98bc9-caa0-49b7-951c-fe4e4f54e4ba"
	B06SpringbloomDruidOracle    = "788a4896-3414-40bf-b391-d2efea2fb5f9"
	B06JungleHollowOracle        = "6de714e1-446d-4fb9-9e3d-bcd3ec6af9ca"
	B06TalrandOracle             = "ea1eb902-a23c-44ff-9169-19baf71de238"
	B06CastleLocthwainOracle     = "be811e70-aaaa-41f3-bf9e-5d3f9f719b49"
	B06MeteorGolemOracle         = "d9f11aa1-9219-42a8-85a9-a8f204160706"
	B06SkyDiamondOracle          = "2224b6e0-c5ff-45d0-84e3-83758c5fc99f"
	B06GildedGooseOracle         = "f2f09757-1931-47c0-a5f0-39280445489d"
)

// b06AddMana drops the listed colours straight into a player's pool
// so an activated ability's mana component can be paid.
func B06AddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// b06TriggerPromptsFor counts the yes/no trigger prompts addressed to
// chooser — the "exactly one trigger per card" canary Bloodchief
// Ascension needs.
func B06TriggerPromptsFor(g *game.Game, chooser uuid.UUID) int {
	return OwedMays(g, chooser)
}

// b06AnswerAllTriggerPrompts answers every trigger prompt addressed
// to chooser the same way, then settles the stack.
func B06AnswerAllTriggerPrompts(t *testing.T, g *game.Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	for i := 0; i < 16 && OwedMays(g, chooser) > 0; i++ {
		AnswerLatestTriggerPrompt(t, g, chooser, apply)
	}
	PassPriorityAroundTable(t, g)
}

// b06GreenCreature pushes a green 2/2 onto the battlefield through the
// real token path, so it carries an EventETB (Oran-Rief reads it).
func B06GreenCreature(t *testing.T, g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(controller, game.Card{
			Name: name, TypeLine: "Token Creature — Elf", Power: 2, Toughness: 2, Colors: []string{"G"},
		}, 1)
	})
	PassPriorityAroundTable(t, g)
	return BattlefieldIDNamed(g, controller, name)
}

// b06ReplacementOrderFor returns the open CR 616 ordering prompt
// addressed to chooser, or nil.
func B06ReplacementOrderFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
