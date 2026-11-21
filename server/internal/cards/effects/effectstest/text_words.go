package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// text_words_test.go — changing one word of an object's text (CR 612.2,
// ADR 0160's 2026-10-15 amendment, #2779): Sleight of Mind, Glamerdye,
// Alter Reality, Magical Hack, Mind Bend, Crystal Spray, Whim of Volrath
// and Artificial Evolution, and the readers that consult the change.

const (
	TwSleightOracle    = "99dba614-40d3-41c1-a3b2-edc8777b010f"
	TwGlamerdyeOracle  = "570a5e60-ed0f-4c14-8cc3-af819648030b"
	TwAlterOracle      = "054c00b4-c940-455a-ad88-b8aa35931ecf"
	TwHackOracle       = "cba229fa-9035-405b-b091-3798898a37ee"
	TwMindBendOracle   = "d8377869-f8f7-4b4b-93ad-3ad3c4826b1a"
	TwSprayOracle      = "fe371e3a-976f-41d8-ad58-4b3b61415d71"
	TwWhimOracle       = "51a20c9f-68d2-471b-8622-67d481b53b92"
	TwEvolutionOracle  = "c5ab6dc9-34af-434f-ae8e-b2389ee6c5ae"
	TwDoomBladeOracle  = "59e7f2ae-4535-4191-98be-3e65b6b2befa"
	TwGoblinKingOracle = "d236b3fc-0d3f-4d99-875d-e32a33fe5767"
)

// The option indices the two prompts offer: colours WUBRG, then the
// basic land types in CR 305.6 order. The second prompt leaves the
// first answer out.
const (
	TwWhite, TwBlue, TwBlack, TwRed, TwGreen = 0, 1, 2, 3, 4
)

// twPut puts a vanilla permanent onto the battlefield.
func TwPut(g *game.Game, owner uuid.UUID, name, typeLine string, colors []string, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Colors: colors, Keywords: keywords,
		Owner: owner, Controller: owner,
	})
}

func TwTargetRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// twChange casts a text-changing spell on `target`, resolves it and
// answers its two prompts with the given indices.
func TwChange(t *testing.T, g *game.Game, name, oracle string, target uuid.UUID, from, to int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	CastCatalogSpell(t, g, name, "Instant", oracle, TwTargetRef(target))
	PassPriorityAroundTable(t, g)
	AnswerOptionPick(t, g, me.ID, from)
	AnswerOptionPick(t, g, me.ID, to)
}

func TwTextChanges(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	c := GaLayered(t, g, id)
	return c.TextChanges()
}
