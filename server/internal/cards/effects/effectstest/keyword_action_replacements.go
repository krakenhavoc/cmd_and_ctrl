package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyword_action_replacements_test.go — #976 from the card side: the
// CR 614 window on a keyword action, watched through the ordinary
// Spec.Replacements slot.
//
// Tekuthal, Inquiry Dominus is the printed proliferate half and it is
// tested through its own catalog entry. The scry half has no printed
// card in the catalog yet — "if you would scry, scry that many plus
// one instead" is the Crystal Ball shape and nothing in the tree
// prints it — so the two probes below stand in for it. They declare
// the SAME helpers a card file would (ScryPlusOne, and the generic
// KeywordActionBecomes for the doubler), which is the point: the
// helper is what a card would reach for, so the helper is what is
// under test.

const (
	ScryPlusOneProbeOracle    = "test-scry-plus-one-probe"
	ScryDoubledProbeOracle    = "test-scry-doubled-probe"
	ProliferateProbeOracle    = "test-proliferate-twice-probe"
	SurveilPlusOneProbeOracle = "test-surveil-plus-one-probe"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     ScryPlusOneProbeOracle,
		Name:         "Scrying Probe",
		Replacements: []game.ReplacementEffect{effects.ScryPlusOne("Scrying Probe — scry one more")},
	})
	effects.Register(effects.Spec{
		OracleID: ScryDoubledProbeOracle,
		Name:     "Doubling Probe",
		Replacements: []game.ReplacementEffect{
			effects.KeywordActionBecomes(game.KeywordActionScry, func(n int) int { return n * 2 },
				"Doubling Probe — scry twice as many"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     ProliferateProbeOracle,
		Name:         "Proliferating Probe",
		Replacements: []game.ReplacementEffect{effects.ProliferateTwice("Proliferating Probe — proliferate twice")},
	})
	effects.Register(effects.Spec{
		OracleID:     SurveilPlusOneProbeOracle,
		Name:         "Surveilling Probe",
		Replacements: []game.ReplacementEffect{effects.SurveilPlusOne("Surveilling Probe — surveil one more")},
	})
}

// replacementIDsByLabel maps a prompt's effect IDs onto the labels the
// test names, in the order it names them.
func ReplacementIDsByLabel(t *testing.T, g *game.Game, ids []game.ReplacementEffectID, labels []string) []game.ReplacementEffectID {
	t.Helper()
	byLabel := make(map[string]game.ReplacementEffectID, len(ids))
	g.ReadSnapshot(func() {
		for _, id := range ids {
			label, _ := g.ReplacementOptionMetaForEffect(id)
			byLabel[label] = id
		}
	})
	out := make([]game.ReplacementEffectID, 0, len(labels))
	for _, want := range labels {
		id, ok := byLabel[want]
		if !ok {
			t.Fatalf("the prompt has no option labelled %q (has %v)", want, byLabel)
		}
		out = append(out, id)
	}
	return out
}
