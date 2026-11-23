package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch17_test.go — card-level coverage for the card-coverage
// roadmap's batch 17 (#310, `edhrec_rank` 1830–1929): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play, attack or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b17-prefixed.

const (
	B17JaradOracle                = "87e65e36-9483-49fe-b644-2caca092107f"
	B17TormodsCryptOracle         = "1573f7f9-672c-421a-b1ac-3d0d8aea59ca"
	B17BorosGuildgateOracle       = "73c423b7-cab8-4e69-8070-9edbf96a6c2c"
	B17BattleOfBywaterOracle      = "a94c191d-a938-458e-bc1b-2f44fd8873a3"
	B17DeepglowSkateOracle        = "debce64a-18bd-42f5-9e85-158c2242e9e9"
	B17ClaimJumperOracle          = "18f0cd0b-3e4f-4637-a62e-75dd1b2f3fce"
	B17UndeadAugurOracle          = "c26887e1-27f4-4550-921b-53460e43c079"
	B17MultaniOracle              = "4b8bf64b-4800-45ff-81c6-2857f34999b5"
	B17ThreefoldThunderhulkOracle = "b8020a8b-557b-465d-865d-b59fecd7abc1"
	B17FiremaneCommandoOracle     = "ac675899-25bb-4f06-9c8c-bf188024495f"
	B17ContainmentConstructOracle = "54d48e54-be4a-4a78-a778-b28e74ef7134"
	B17RazakethOracle             = "136c9ecb-59b4-4ef6-bcb2-7b8d3df3ee75"
	B17CordialVampireOracle       = "d61fb9e8-d05a-481a-a90f-5def300c9abb"
	B17ArwenOracle                = "c0891157-f3dd-4866-866f-fd13e9be9633"
	B17DictateOfKruphixOracle     = "be74a4c1-d569-4203-b28a-3e2ac6a82990"
	B17KazuulOracle               = "f8bf3d91-cb50-48ce-88c6-cdcb37b64b57"
	B17PhyrexianObliteratorOracle = "41820f91-27cf-41c0-bb5e-9adf6845a6a4"
	B17BreenaOracle               = "d11e627b-8a48-411d-a261-2c9a02a758ba"
	B17NighthawkScavengerOracle   = "379ed4ec-8f2e-448b-9ed2-a3181e2f877e"
	B17HarvesterOfSoulsOracle     = "5987ce77-10ad-4871-900a-5a005fcf4955"
	B17TempleBellOracle           = "fc8032e9-c83a-4cf9-92e1-b1d2d9642695"
	B17ColossalGraveReaverOracle  = "df8e0d1b-b47c-4807-9c9b-84dcec835254"
	B17JacesArchivistOracle       = "b6c8ac69-daa7-4e2e-a1d9-439731a81870"
	B17InsidiousFungusOracle      = "0a8d0217-ff24-4177-b6be-707eb2b6b9e9"
	B17VileEntomberOracle         = "556dd333-4066-41f7-98f6-41794754de71"
	B17RefuteOracle               = "eb1dfa29-7371-4cb6-bfa2-16f7820b69be"
	B17LeafCrownedVisionaryOracle = "ef693b9f-11e3-49bf-8387-b8f480b9007a"
	B17TesharOracle               = "d4193f50-33b5-4201-ba03-eb135eced7ff"
	B17RiseOfTheWitchKingOracle   = "3c86541c-3601-4a38-8872-39705e41303a"
)

// b17Life reads every seat's life.
func B17Life(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Life)
	}
	return out
}

// b17GraveyardCard seeds a card with a full type line into a
// player's graveyard.
func B17GraveyardCard(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// b17PickCards answers a multi-slot pick_target prompt.
func B17PickCards(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

// b17PickCardsDivided answers the newest pick_target prompt with
// `ids` and the division announced with them (#1563, CR 601.2d).
func B17PickCardsDivided(t *testing.T, g *game.Game, chooser uuid.UUID, dist map[uuid.UUID]int, ids ...uuid.UUID) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	refs := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargetsDivided(p.ID, chooser, refs, dist); err != nil {
		t.Fatalf("ResolvePickTargetsDivided: %v", err)
	}
}
