package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b05_helpers.go — shared by Dominating Vampire and Midnight Arsonist
// (Innistrad: Crimson Vow batch 5, #5495): "the number of Vampires you
// control".

// vowB05VampiresControlledBy counts the Vampire permanents `player`
// controls, each read post-layer (a creature that is a Vampire as it is now).
// `vampire` is the word as the asking card's text now says it.
func vowB05VampiresControlledBy(g *game.Game, player uuid.UUID, vampire string) int {
	return countControlled(g, player, func(c game.Card) bool { return c.HasSubtype(vampire) })
}

// vowB05OdricKeywords are the twelve abilities Odric, Blood-Cursed counts.
var vowB05OdricKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "haste", "hexproof",
	"indestructible", "lifelink", "menace", "reach", "trample", "vigilance",
}

// vowB05ListedAbilitiesAmongCreatures is the number of Odric's listed
// abilities that at least one creature `player` controls has, each counted
// once (CR 702 keywords read through the layered view, HasKeyword).
func vowB05ListedAbilitiesAmongCreatures(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, kw := range vowB05OdricKeywords {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == player && c.IsCreature() && game.HasKeyword(&c, kw) {
				n++
				break
			}
		}
	}
	return n
}

// vowB05BloodTokensControlledBy counts the Blood tokens `player` controls.
func vowB05BloodTokensControlledBy(g *game.Game, player uuid.UUID) int {
	return countControlled(g, player, func(c game.Card) bool { return IsToken(c) && c.HasSubtype("Blood") })
}
