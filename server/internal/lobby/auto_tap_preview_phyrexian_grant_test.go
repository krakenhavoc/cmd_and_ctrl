package lobby

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_phyrexian_grant_test.go — #1589. The auto-tap
// preview reads the one pricer, so under an any-colour or any-type
// grant a {1}{B/P}{B/P} keeps its Phyrexian symbols: `?phyrexian=2`
// plans one Mountain, and CastSpell{AutoTap} takes the same cast. With
// no life claimed, one Mountain is short — and the engine agrees.
func TestAutoTapPreviewKeepsPhyrexianLifeUnderSpendGrants(t *testing.T) {
	for _, tc := range []struct {
		name string
		perm game.CastPermission
	}{
		{"any color", game.CastPermission{AnyColor: true, CastOnly: true, Duration: game.WhileInZoneDuration()}},
		{"any type", game.CastPermission{AnyType: true, CastOnly: true, Duration: game.WhileInZoneDuration()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newPreviewFixture(t)
			f.mainPhase()
			card := f.spawn(f.alice, game.ZoneExile, game.Card{
				Name: "Preview Dismember", TypeLine: "Sorcery", ManaCost: "{1}{B/P}{B/P}",
			}, 1)[0]
			f.grant(card, tc.perm)
			f.mountains(1)

			if short := f.previewWith(card, "&from_zone=exile"); short.OK {
				t.Errorf("no life claimed, one Mountain: preview ok; {1}{B/P}{B/P} needs three mana")
			}
			got := f.previewWith(card, "&from_zone=exile&phyrexian=2")
			if !got.OK || len(got.Plan) != 1 {
				t.Fatalf("both symbols by life: ok=%v plan=%v missing=%v, want ok with one Mountain", got.OK, got.Plan, got.Missing)
			}
			if !f.castsWithAutoTap(card, game.CastSpellParams{FromZone: "exile", PhyrexianLife: 2}) {
				t.Error("the preview said ok and CastSpell refused the same cast")
			}
		})
	}
}
