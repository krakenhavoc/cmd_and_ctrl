package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fra_prepare_a_test.go — Reality Fracture slice fra-prepare-a: the
// preparation cards and the cards that make or cast prepared creatures.
// One test per behaviour, through the catalog and the real cast path.

const (
	FraPhytomedicOracle = "7737ccdc-49f2-463e-8ffd-43ab98d11f45"
	FraBlossomOracle    = "cd61c26b-8d3c-4730-92f2-0d94d961bafd"
	FraKonstrariOracle  = "bad2e3b2-ad47-4e17-b35e-0fcc13e82937"
	FraCrafterOracle    = "c8b3a070-408e-4a8f-8597-6476eeec0a5f"
	FraFateholdOracle   = "1063822f-47d3-42e9-8a21-f62b12609fe1"
	FraDivinerOracle    = "e7b78acd-5288-4528-97a8-a5b43de90bf5"
	FraHecklerOracle    = "d1af8018-35b5-4b6f-93c8-1ed6bd8c92b1"
	FraBloodlineOracle  = "c8e9a7e1-28ae-40e3-8715-1bb0aa2057c7"
	FraCultivatorOracle = "db655c5a-06f6-42bc-8180-d9ca1c1d939a"
	FraCodieOracle      = "a17e414c-8de0-47fd-bcbc-540c7fec7642"
	FraArenaOracle      = "2cc075da-f79b-4050-84eb-2c1b9a30260c"
	FraCourseworkOracle = "666b34b5-a0c2-43e3-befd-7c3979313682"
)

func FraCard(owner uuid.UUID, oracle string, keywords []string, front, spell game.Face) game.Card {
	return PreparationCard(owner, oracle, keywords, front, spell)
}

func FraBloodline(owner uuid.UUID) game.Card {
	return FraCard(owner, FraBloodlineOracle, nil,
		game.Face{Name: "Bloodline Recollector", TypeLine: "Creature — Vampire Warlock", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: 2, Toughness: 2},
		game.Face{Name: "Ancestral Craving", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"}})
}

func FraCastCopy(t *testing.T, g *game.Game, me *game.Player, cp game.Card, targets ...game.TargetRef) error {
	t.Helper()
	return g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: targets})
}

func FraHasCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterPlusOne]
			}
		}
	})
	return n
}
