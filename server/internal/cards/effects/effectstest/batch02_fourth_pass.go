package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch02_fourth_pass_test.go — the second slice of roadmap batch 02's
// re-triage (#295): the six cards #1241's triage confirmed ready and
// left for the next PR. The Ozolith shipped early as #1244's own proof
// card (the_ozolith_test.go), so this file covers the other five.
//
// Naming is `b02d`-prefixed to stay clear of batch02_test.go's `b02`
// helpers, batch02_second_pass_test.go's `b02b` helpers and
// batch02_third_pass_test.go's `b02c` helpers.

const (
	B02dBraidsOracle              = "e0445c80-fa53-4c3e-881e-940e9fce7f57"
	B02dHullbreakerHorrorOracle   = "d4a84e78-d9b9-4c67-8a4b-4329e65f0f15"
	B02dFinaleOfDevastationOracle = "69872a9a-fe54-4e58-940c-89395af71acd"
	B02dKindredDiscoveryOracle    = "005ee549-1bf5-478f-bc3f-3e791bd7eecf"
	B02dDawnsTruceOracle          = "37c06f89-db36-4937-9404-2b07cd22e1a6"
)

// --- Braids, Arisen Nightmare ----------------------------------------

// b02dCastInstantAs pushes a costless instant into player's hand and
// casts it directly, bypassing castCatalogSpell's "advance to main
// phase" loop — an instant has no sorcery-speed gate, and advancing
// steps drains the stack (#914 / CR 117.4), which would resolve away
// whatever this leaves open for the next cast.
func B02dCastInstantAs(t *testing.T, g *game.Game, player *game.Player, name string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	player.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		Owner: player.ID, Controller: player.ID,
	})
	if err := g.CastSpell(player.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// hasKeyword reports whether list contains kw. Only used to read a
// Spec's PrintedKeywords back in a test.
func HasKeyword(list []string, kw string) bool {
	for _, k := range list {
		if k == kw {
			return true
		}
	}
	return false
}
