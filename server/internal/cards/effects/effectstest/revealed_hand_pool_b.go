package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0116 pool B (#2078): the revealed-hand pick on modal spells,
// alternative and variable costs, and spells with text after the
// discard. One test for each card's own clause; the pick itself is
// pinned in revealed_hand_discard_test.go.

const (
	AuntiesSentenceOracle      = "406e6db7-6e31-44b9-a793-57ee10688f71"
	CerebralConfiscationOracle = "1ea71ea2-65ca-4907-b857-dbd1ba8efacd"
	SplittingHeadacheOracle    = "3a5937de-5957-47a9-9bd4-0dc7f8b833ef"
	PoisonTheWatersOracle      = "2c9bf40b-ddcb-4f46-a81e-56b91dfc784f"
	MarduCharmOracle           = "147eabab-9d93-4cf4-811b-0efa3b84c5b5"
	DaiLiIndoctrinationOracle  = "5591efbc-1b2c-489a-b4f3-935c9b6bd71a"
	DrillBitOracle             = "0dbd6e47-a8b4-4268-ba44-8924cd4963a6"
	DreadFugueOracle           = "efe4194a-b3c5-4308-abfe-553e035f5e64"
	VenarianGlimmerOracle      = "336cc3ab-c1c7-4bf1-b1d7-dfc9f3c556ab"
	GruesomeDiscoveryOracle    = "ecdae60b-c594-4e70-909b-83483104a42c"
	RiversGraspOracle          = "f671583f-bd22-4d3e-bf19-fb740c607f8f"
	DevourIntellectOracle      = "abf8264e-d020-4939-8092-83469b1a2244"
	HumiliateOracle            = "cf18a1ca-1fb5-4c90-b20a-6ae2fb54a9bb"
	RenderSpeechlessOracle     = "6fdddab4-fcc9-4622-9c17-c69f9d4bca08"
	EgoDrainOracle             = "b4fc198f-3595-4880-ad12-96c18fbf9c33"
	MemoryTheftOracle          = "1d60e2a7-2059-48cb-a6ab-36ba25b80b3a"
)

func PbBear() game.Card {
	return game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2}
}

func PbDivination() game.Card {
	return game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}", Colors: []string{"U"}}
}

// pbCast puts a card in the active seat's hand, moves to a main phase
// and casts it with params. Returns the caster.
func pbCast(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, params game.CastSpellParams) *game.Player {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	ToMainForCost(t, g)
	id := HandCardFull(me, name, typeLine, manaCost, oracle, nil)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return me
}

// pbModal casts a "choose one" spell for mode `mode` with targets.
func PbModal(t *testing.T, g *game.Game, name, typeLine, oracle string, mode int, targets ...game.TargetRef) {
	t.Helper()
	pbCast(t, g, name, typeLine, oracle, "", game.CastSpellParams{Modes: []int{mode}, Targets: targets})
	PassPriorityAroundTable(t, g)
}

func PbPlayer(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }
func PbCard(id uuid.UUID) game.TargetRef   { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func ContainsID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
