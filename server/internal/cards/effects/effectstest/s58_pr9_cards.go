package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s58_pr9_cards_test.go — S58 PR 9, "Drawn To Death (2): wheels,
// tutors and the rest": ten cards through the real catalog.

const (
	S58ArchfiendOfIfnir     = "535e2af0-7b08-4026-b1bf-87626407dd42"
	S58AwakenErstwhile      = "06fa6719-aa4f-4857-8029-e42d01232645"
	S58BeholdTheBeyond      = "b2402ac0-8ab8-413c-bac4-9f4ba2163c58"
	S58CurseOfFoolsWisdom   = "93b2a685-8267-4cbd-ac55-db6ff95fe98d"
	S58DemonicCollusion     = "53cbc12d-e182-4cde-8967-9b2664daa817"
	S58DoomsdayExcruciator  = "ad0db433-a406-4ef0-8ffa-416af610c4e7"
	S58IllGottenGains       = "c88eff74-def5-42ea-873c-b3b479f6fe18"
	S58SchemingSilvertongue = "44443716-8356-40cd-a879-35264b29108c"
	S58StarvingRevenant     = "2ca969eb-3d79-4d1f-8d9d-7b8204ad166a"
	S58MasterOfLakeTown     = "89c3c81b-8960-4f79-b2ec-f13560058071"
)

// s58HandCards puts n filler cards in p's hand and returns their IDs.
func S58HandCards(p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: ids[i], Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	return ids
}

// s58Library stacks named cards on top of p's library; the LAST name is
// the top card.
func S58Library(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i, n := range names {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{InstanceID: ids[i], Name: n, TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	return ids
}

func S58CastRevenant(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	AdvanceToMain(t, g)
	CastFromHandForTest(t, g, me, "Starving Revenant", "Creature — Spirit Horror", "", S58StarvingRevenant, game.CastSpellParams{})
	PassPriorityAroundTable(t, g)
}
