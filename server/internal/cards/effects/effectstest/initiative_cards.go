package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// initiative_cards_test.go pins the catalog cards built on the
// initiative (ADR 0203, #4096): one test per card, plus From the
// Catacombs' escape and its exile-instead clause.

const (
	FromTheCatacombsOracle        = "cc21b29f-2d1a-4769-9b65-f92bd2ca1cc8"
	AarakocraSneakOracle          = "c1882cb4-f69e-441e-8871-743ed636cad0"
	AvengingHunterOracle          = "e0aa5e3d-17c0-4ac7-9d76-9ce579d1eb2b"
	GoliathPaladinOracle          = "269a02ca-1425-4b32-81f3-440e1caadb17"
	UnderdarkExplorerOracle       = "7647682d-d491-4cb1-add7-b4d745994e62"
	StirringBardOracle            = "2583836d-7fd3-47b3-9b69-6814efe54a90"
	ViciousBattleragerOracle      = "b8942c72-16ce-4f2f-875f-97a56e85a03e"
	PassagewaySeerOracle          = "2d8ebb48-5f37-4b7e-b8ec-3296d9a039c7"
	FeywildCaretakerOracle        = "f319b671-0388-48e8-ba98-04161a3e7f1a"
	ImoenOracle                   = "280e05c5-7ab4-45bf-aa15-e594cec49ef3"
	SafanaOracle                  = "afc1c443-93e3-4ef4-a404-1d9fc13e0b02"
	BloodboilSorcererOracle       = "acdcffa0-87e5-48a2-93d2-572febfc9d52"
	UndermountainAdventurerOracle = "2e5e0abd-38a6-46ca-b922-556db52f1332"
	DungeoneersPackOracle         = "c66976d3-57f1-4c48-ada0-d74081f1699a"
	LootDisputeOracle             = "7c885698-69f2-44b5-abfe-9a2833822794"
	ExploreTheUnderdarkOracle     = "13bad22a-6ba3-4197-8658-e47049a1cf1f"
	WhitePlumeAdventurerOracle    = "51f091a7-9b0b-4362-8e9c-c174752f369b"
	TrailblazersTorchOracle       = "1cdac498-3c94-48d6-b94c-1d98697370ea"
)

// initZone is the zone the card with this ID is in, "" when none.
func InitZone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var k game.ZoneKind
	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			k = z.Kind
		}
	})
	return k
}

// initCompleteADungeon records a completed Undercity for `p`, as a
// state a test starts from.
func InitCompleteADungeon(p *game.Player) {
	p.CompletedDungeons = append(p.CompletedDungeons, game.UndercityKey)
}
