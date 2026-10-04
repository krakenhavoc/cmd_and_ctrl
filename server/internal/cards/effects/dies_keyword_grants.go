package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// dies_keyword_grants.go — shared vocabulary for the cards that give
// undying or persist (#2075, ADR 0113 §4) and the Scarecrows that have
// a keyword while you control a creature of a colour. Append-only.

// grantEachLegalTargetUntilEOT gives every card target still legal at
// resolution (CR 608.2b) the keyword until end of turn: Undying Evil,
// Cauldron Haze, Cauldron of Souls, Antler Skulkin and Rhys, the
// Evermore. The grant is a layer-6 record pinned to the object, so a
// creature that dies wearing it has the keyword as it last existed
// (CR 603.10a) and its dies trigger fires.
func grantEachLegalTargetUntilEOT(ctx *Context, keyword, label string) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{keyword}, Label: label}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// selfWhileYouControlA is a self-only KeywordGrant condition: "this
// creature has <keyword> as long as you control a <pred> creature"
// (Rattleblaze Scarecrow, Wingrattle Scarecrow). The source itself
// counts if it matches.
func selfWhileYouControlA(pred CardPredicate) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID &&
			len(permanentsControlledByMatching(g, source.Controller, pred)) > 0
	}
}
