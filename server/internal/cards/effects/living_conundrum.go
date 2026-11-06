package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Conundrum — Creature — Elemental {4}{U}, 2/5 (Murders at
// Karlov Manor, #4469):
//
//	"Hexproof
//	 If you would draw a card while your library has no cards in it,
//	 skip that draw instead.
//	 As long as there are no cards in your library, this creature has
//	 base power and toughness 10/10 and has flying and vigilance."
//
// The draw is SkipThatDrawFromAnEmptyLibrary (#5246), a CR 614 draw
// replacement that cancels the draw and does nothing instead: no card is
// drawn, "whenever you draw" does not trigger, and the player does not
// lose for drawing from an empty library (CR 704.5b). Each draw of an
// instruction to draw several is its own event (CR 121.2), so only the
// draws the library cannot cover are skipped (the 2024-02-02 ruling).
//
// The third sentence is three statics on the source, gated on its
// controller's library being empty: a layer 7b SET (CR 613.4b), so
// counters and pumps still apply on top of the 10/10, and two layer 6
// keyword grants. An empty library is the library's top card turning to
// nothing, so they declare DependsOnLibraryTop: the layer engine watches
// every library's top card and recomputes the moment one changes,
// including the draw that empties the library and a card put back into
// it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "378ae023-04d3-44cc-9248-3d787796ed6c",
		Name:            "Living Conundrum",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Replacements: []game.ReplacementEffect{
			SkipThatDrawFromAnEmptyLibrary("Living Conundrum"),
		},
		Static: livingConundrumStatics(),
	})
}

// livingConundrumStatics is "as long as there are no cards in your
// library, this creature has base power and toughness 10/10 and has
// flying and vigilance".
func livingConundrumStatics() []game.StaticAbility {
	base := game.StaticAbility{
		Label:     "As long as there are no cards in your library, this has base power and toughness 10/10.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7B_Set,
		AppliesTo: livingConundrumLibraryEmpty,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power, c.Toughness = 10, 10
		},
	}
	out := []game.StaticAbility{
		base,
		KeywordGrant(livingConundrumLibraryEmpty, "flying"),
		KeywordGrant(livingConundrumLibraryEmpty, "vigilance"),
	}
	for i := range out {
		out[i].DependsOnLibraryTop = true
	}
	return out
}

// livingConundrumLibraryEmpty is the statics' condition: the target is
// the source itself, and its controller's library has no cards in it.
func livingConundrumLibraryEmpty(target *game.Card, g *game.Game, source *game.Card) bool {
	if !selfOnly(target, g, source) {
		return false
	}
	p := g.PlayerByIDForEffect(source.Controller)
	return p != nil && p.Library != nil && p.Library.Size() == 0
}
