package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_n_cards_test.go — the cards #747's sacrifice-N cost
// enabled: five caveats removed (Savvy Hunter, Samwise Gamgee, Sai,
// both Magdas) and eight cards added. The engine rules are pinned in
// sacrifice_n_cost_test.go; these assert each card pays its printed
// count and does what it prints.

const (
	PriestOfForgottenGodsOracle = "2ad8ff62-d090-4835-9274-3b755ba0f8e6"
	KuldothaForgemasterOracle   = "b0f99367-4313-4bb1-a9fd-7711bc4ce40e"
	TamiyosJournalOracle        = "20591a83-7138-4876-b088-035bd3788be3"
	TeysaOrzhovScionOracle      = "8191342b-b25e-4c4d-8f69-aee662148ff4"
	HedronDetonatorOracle       = "9221b29b-9351-4f70-9af4-8b9a33b43138"
	OliviaOpulentOutlawOracle   = "4e5e0b22-ee5a-4c19-be95-2f9bd50641bc"
	BreyaEtheriumShaperOracle   = "d460a9e2-5a7d-4562-880e-45174be19a9d"
	WhisperBloodLiturgistOracle = "496ab82e-24a9-4a74-ba4b-992c7309b44a"
)

func PushTokens(g *game.Game, controller uuid.UUID, tmpl game.Card, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = PushToken(g, controller, tmpl)
	}
	return ids
}
