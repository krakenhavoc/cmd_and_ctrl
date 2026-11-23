package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sami_c_test.go — the Sami Whammy deck's lands and removal artifacts
// (slice sami-c, tracker #2190).

const (
	SamiConduitPylonsOracle    = "37f924e1-7c25-4f06-88bb-054693a21e5a"
	SamiCrystalGrottoOracle    = "f15fb0cc-8e96-4f03-94d0-b51410415afd"
	SamiEiganjoOracle          = "7edb3d15-4f70-4ebe-8c5e-caf6a225076d"
	SamiFomoriVaultOracle      = "622a53bd-d894-422c-a606-7126041afa02"
	SamiMonumentalHengeOracle  = "c48df45c-3513-4d56-aed6-30c2f3a759cd"
	SamiNeedlevergePathOracle  = "a9b8d020-4d72-4934-8942-df29ef19fc1d"
	SamiSokenzanOracle         = "c5ee72d5-3a9e-4fe5-8802-3286ee612055"
	SamiSpireOfIndustryOracle  = "55a519b4-61cb-448a-875b-4d6dbe00580f"
	SamiSunbillowVergeOracle   = "a202276b-1f1b-4277-95ee-26877a204f5e"
	SamiBlitzballOracle        = "75719f87-ce1a-40d6-ac5a-5ebf0c936c84"
	SamiBoommobileOracle       = "cb3edb86-ddd5-461c-98b5-ffb4df437a39"
	SamiDragonsparkOracle      = "a8acb0b0-7265-42f6-8143-62d0c02d5a28"
	SamiMemorialVaultOracle    = "10dda10f-2549-47e5-9a8b-fb414cf38d3f"
	samiTestLibraryCardTypeFmt = "Sorcery"
)

// samiLibraryTopFirst seeds a player's library so the first entry is
// the top card, with the given type lines.
func SamiLibraryTopFirst(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := range cards {
		ids[i] = uuid.New()
	}
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = ids[i]
		c.Owner, c.Controller = p.ID, p.ID
		if c.TypeLine == "" {
			c.TypeLine = samiTestLibraryCardTypeFmt
		}
		p.Library.PushTop(c)
	}
	return ids
}

func SamiMainPhase(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}
