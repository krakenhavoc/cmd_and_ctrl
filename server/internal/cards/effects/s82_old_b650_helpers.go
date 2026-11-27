package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b650_helpers.go — Invasion, S82 older sets batch 650 (#5616).

// OldB650KickedKeywordGrant is "If this creature was kicked, it enters with
// <keyword>": a layer 6 grant on the permanent itself, held while its own
// kicker record says it was kicked (game.CardKickedTimes, CR 400.7d). A
// permanent that was never cast (reanimated, flickered) has none.
func OldB650KickedKeywordGrant(keyword string) game.StaticAbility {
	grant := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID == source.InstanceID && game.CardKickedTimes(*target) > 0
	}, keyword)
	grant.Label = "If this creature was kicked, it has " + keyword + "."
	return grant
}
