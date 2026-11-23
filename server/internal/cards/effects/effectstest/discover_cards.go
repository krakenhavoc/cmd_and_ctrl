package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discover_cards_test.go — the discover cards (ADR 0099). Each test
// proves the card discovers the right N: the library is a card with
// mana value exactly N on the bottom and one with mana value N+1 on
// top, so the prompt offers the bottom card only when N is right.

const (
	HiddenCataractOracle    = "927979d7-9b5c-4448-aef0-baf2907a89f1"
	HiddenCourtyardOracle   = "e19d5071-4ea1-4883-b067-a21e553f96e0"
	HiddenNecropolisOracle  = "f780ee53-62b0-4c32-b5b7-047651f48e5f"
	HiddenNurseryOracle     = "1a26e2d6-6bfc-4cdc-9bd6-8b37a9be2961"
	HiddenVolcanoOracle     = "a1c7cd7a-0795-4135-b787-effeb981d95b"
	DaringDiscoveryOracle   = "a9564593-b5a6-4a83-be1e-2af8caf647d7"
	WalkAncestorsOracle     = "ce657a2e-461b-45fd-ab98-35b5c8512777"
	HitMotherLodeOracle     = "a49900b3-cc34-428f-806d-71861dc8059d"
	HurlIntoHistoryOracle   = "016718df-2eb5-499f-adea-18231c0375b3"
	ZoyowasJusticeOracle    = "17e11f0c-4ac7-4060-9e0b-27cd5c151e50"
	ContestOfClawsOracle    = "e30d5114-40d7-4dc5-8067-9c5e9625e097"
	TrumpetingCarnoOracle   = "f2ef8bda-373d-4387-900d-0f1b6ccf72e9"
	EtalisFavorOracle       = "df98b93b-0aa6-4bea-9d8d-798ebc795b0b"
	GeologicalApprOracle    = "a0a1cd2b-ed7c-4cce-8497-4873e1e5e3df"
	ChimilOracle            = "119d9671-61ba-4629-89b1-f94bdec5cb74"
	FranklinRichardsOracle  = "af20f452-5cb9-47cc-9a81-71f23204ddd0"
	MonstrousVortexOracle   = "6e27956e-ba2f-41a9-8e39-ec54424591b6"
	CuratorOracle           = "0e636c98-67c7-4f02-b909-5d43e276fa41"
	PantlazaOracle          = "b828ba28-e98d-498b-9ab4-d4aa7143e407"
	AloyOracle              = "f0554a8f-32de-4069-9f47-5e06ceb3f09d"
	DinosaurEggOracle       = "ac4d5a97-6177-4eac-b0f2-cf10531ad879"
	CaparoctiOracle         = "7f87a770-af47-4345-8dc1-0a1e8bae75c5"
	EllieAndAlanOracle      = "0f4d513e-100d-4cbb-b919-09249cef5fad"
	LongRangeSensorOracle   = "358297cd-4a9f-48e2-aaa5-131b0849f8dc"
	BuriedTreasureOracle    = "56f5ed8d-ef53-4766-9ef3-1e24a10e267b"
	QuintoriusKandOracle    = "91babc5f-ebeb-4894-8c35-b0f60249cb80"
	SwashbucklersWhipOracle = "5cdb69e5-9395-4ef6-be18-71e3dd70df66"
	ZoeticGlyphOracle       = "1c21efcf-1007-45bb-ba21-ac33c2a8d751"
)

// discoverLibrary replaces p's library with a mana-value-n card under a
// mana-value-(n+1) card and returns the first: the card a discover n
// finds, and the one a discover of anything else does not.
func DiscoverLibrary(p *game.Player, n int) uuid.UUID {
	exact := DiscoverLibraryCard(p.ID, "Exact Hit", "Artifact", GenericCost(n))
	over := DiscoverLibraryCard(p.ID, "Over By One", "Artifact", GenericCost(n+1))
	p.Library.Cards = []game.Card{exact, over}
	return exact.InstanceID
}

func GenericCost(n int) string {
	if n == 0 {
		return "{0}"
	}
	out := ""
	for i := 0; i < n; i++ {
		out += "{1}"
	}
	return out
}

// discoverPromptFor is the open discover prompt (a resolution_cast, #4243)
// addressed to chooser, or nil.
func DiscoverPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// expectDiscovered fails unless a discover prompt for chooser offers
// want, then declines it (the card goes to hand).
func ExpectDiscovered(t *testing.T, g *game.Game, chooser, want uuid.UUID) {
	t.Helper()
	c := DiscoverPromptFor(g, chooser)
	if c == nil {
		t.Fatalf("no discover prompt for the discoverer")
	}
	if c.MayCastKeyword != game.MayCastKeywordDiscover {
		t.Fatalf("the prompt is a %q cast, not a discover", c.MayCastKeyword)
	}
	if KeywordCastCard(c) != want {
		t.Fatalf("discover offered the wrong card — the N is wrong")
	}
	AnswerMayCastPrompt(t, g, chooser, false)
}
