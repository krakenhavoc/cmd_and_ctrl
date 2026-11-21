package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// znr_mdfc_fronts_test.go — #4170: the front faces of the Zendikar Rising
// modal double-faced cards whose land backs were already automated
// (mdfc_lands.go). Each front registers under the bare oracle ID, the same
// as Bridgeworks Battle's, so the entry is Full when both faces are.

const (
	AkoumWarriorOracle         = "afedce7b-0e18-40ad-a26a-1933fddb560d"
	BeyeenVeilOracle           = "b03de49d-246f-44e2-9487-9e4e43ec7be4"
	BlackbloomRogueOracle      = "34320ebf-da97-44a4-bbeb-a9da06548289"
	BloodsoakedInsightOracle   = "c52fc8a1-43c6-41f8-b010-03be7c89ef1d"
	DrownerOfTruthOracle       = "db19a27a-ee22-4931-ae3c-0ce21f456ea6"
	EmeriasCallOracle          = "6ec2a242-9068-4ee2-8ac8-8341cc570f56"
	GlasswingGraceOracle       = "3a3e8c9b-e458-4661-980d-0a84a4c2452b"
	JwariDisruptionOracle      = "941a4b14-ea2a-4bd0-8cc2-d609f80df32c"
	KabiraTakedownOracle       = "0bb73c07-0220-4ba9-8d85-3c357c223833"
	KhalniAmbushOracle         = "37a55560-6e32-4f54-b9a8-fd157aea6eb5"
	LegionLeadershipOracle     = "ad225ec2-ff3a-48f6-81a7-dfdd1b75e1f7"
	MakindiStampedeOracle      = "342e08f9-d4d0-4408-8621-66e087058616"
	RazorgrassAmbushOracle     = "5da954fa-9001-4557-825c-1462035d21ed"
	SilundiVisionOracle        = "b0182ca0-f353-4012-9121-6f4ac9f7a046"
	SkyclaveClericOracle       = "da9e3910-9a1c-43a9-9138-ca971b2bccae"
	SongMadTreacheryOracle     = "81b61770-2ed5-4a50-84d0-97790002fc5a"
	StrengthOfTheHarvestOracle = "1a8c996d-ca93-4c17-ace5-66ecd6b99317"
	TangledFlorahedronOracle   = "53542c79-a62a-4d6a-97db-5296e9c68302"
	TurntimberSymbiosisOracle  = "403b59f3-7ade-4bc2-a3e6-de0c3c700f18"
	UmaraWizardOracle          = "6bc668f4-8fc7-4aaf-891b-277d8328b376"
	VastwoodFortifOracle       = "ce148a0c-6c63-49d5-a156-99efae4e367a"
	ZofConsumptionOracle       = "d9f11985-e460-425d-b083-9cb0edf1983a"
)

// --- shared board helpers ---------------------------------------------

// znrStackLibrary puts the named cards on top of p's library, the first
// name on top, and returns their IDs in that order.
func ZnrStackLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

func ZnrZoneOf(g *game.Game, id uuid.UUID) game.ZoneKind {
	var kind game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			kind = z.Kind
		}
	})
	return kind
}

func ZnrBattlefieldCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c, true
				return
			}
		}
	})
	return out, found
}

// --- Turntimber Symbiosis -----------------------------------------------

func ZnrTurntimberBoard(t *testing.T) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[0]
	top := ZnrStackLibrary(me,
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Big Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}", Power: 5, Toughness: 5},
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Three Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	CastCatalogSpell(t, g, "Turntimber Symbiosis", "Sorcery", TurntimberSymbiosisOracle, nil)
	PassPriorityAroundTable(t, g)
	return g, me, top
}
