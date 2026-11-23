package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_snc_b16_helpers.go — shared pieces for Streets of New Capenna, S82
// batch 16.

// sncB16CardTypes are the card types of CR 205.2a, lowercase, the way
// Card.HasCardType reads them.
var sncB16CardTypes = []string{
	"artifact", "battle", "conspiracy", "creature", "dungeon", "enchantment",
	"instant", "kindred", "land", "phenomenon", "plane", "planeswalker",
	"scheme", "sorcery", "vanguard",
}

// sncB16OneCardOfEachType is the set rule of "up to one card of each card
// type": the picked cards can each stand for a different card type. A card
// with two types (an artifact creature) may stand for either, never both,
// so two artifact creature cards are a legal pick only because one can be
// the artifact and the other the creature. It is a bipartite matching of
// cards to types, tried by augmenting paths. A function of the cards alone,
// capturing nothing (game.SetRuleFunc).
func SncB16OneCardOfEachType(picked []game.Card) bool {
	owner := map[string]int{} // type → index of the picked card standing for it
	var place func(i int, seen map[string]bool) bool
	place = func(i int, seen map[string]bool) bool {
		for _, ty := range sncB16CardTypes {
			if !picked[i].HasCardType(ty) || seen[ty] {
				continue
			}
			seen[ty] = true
			if j, taken := owner[ty]; !taken || place(j, seen) {
				owner[ty] = i
				return true
			}
		}
		return false
	}
	for i := range picked {
		if !place(i, map[string]bool{}) {
			return false
		}
	}
	return true
}
