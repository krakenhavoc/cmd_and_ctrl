package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jaheira, Friend of the Forest — Legendary Creature — Human Elf Druid
// {2}{G}, 2/3:
//
//	"Tokens you control have "{T}: Add {G}."
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// TOKENS, not creature tokens: a Treasure, a Clue and a Food tap for
// {G} as well as a Soldier does. ADR 0093's layer-6 grant.
//
// Declared simplification (weaker than printed): "Choose a
// Background" is a deckbuilding permission (CR 702.124) the deck
// importer does not support, so Jaheira is a single commander — the
// same gap Ganax, Astral Hunter declares.
const jaheiraGrant = "jaheira-friend-of-the-forest/tap-for-green"

func init() {
	Register(Spec{
		OracleID:     "0aeeb0d7-15a3-4722-ad67-8fc7fe89220d",
		Name:         "Jaheira, Friend of the Forest",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Choose a Background isn't supported — Jaheira can be your commander, but not alongside a Background."},
		Grants:       []AbilityGrant{TapForManaGrant(jaheiraGrant, "{G}", "Add {G}", "{T}: Add {G}.")},
		Static:       []game.StaticAbility{GrantAbilities(tokensYouControl, jaheiraGrant)},
	})
}

// tokensYouControl is the "tokens you control" AppliesTo.
func tokensYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsToken() && target.Controller == source.Controller
}

// creatureTokensYouControl is the "creature tokens you control"
// AppliesTo (Insidious Roots, Springleaf Parade).
func creatureTokensYouControl(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.IsCreature() && tokensYouControl(target, g, source)
}
