package effects

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_spawn.go publishes the catalog's token templates to the
// table spawner (ADR 0075 §2.4). Tokens are the reason the owner
// wanted spawning in production at all: the dev spawner resolves a
// Scryfall printing, and a token has none, so a table that needed a
// Treasure or a Clue had no way to get one.
//
// It is a SEPARATE surface from tokenTable rather than an accessor on
// it, because the two answer different questions. tokenTable is what
// a CARD may create — data, keyed by the printed phrase. This is what
// a PERSON may ask for, which is the table plus the behaviour tokens
// in tokens.go, whose templates are constructors because the ability
// on them is code.
//
// The lobby reaches this through a tiny interface on lobby.Config
// (lobby.TokenTemplates), wired in main. That keeps internal/cards/effects
// where it has always been — imported by main and by internal/catalog and
// by nothing else — so the session layer does not grow a dependency on
// the whole card catalog to offer a Treasure.

// TokenLibrary is the spawner's view of the token templates. Zero
// value is ready to use; it holds nothing.
type TokenLibrary struct{}

// Tokens returns the token library. main hands it to lobby.Config.
func Tokens() TokenLibrary { return TokenLibrary{} }

// behaviourTokens are the templates that are constructors rather than
// table rows: every one of them carries a mana ability or an
// activated ability, which is code and cannot be a row.
//
// Keyed the way the card that makes one prints it — "Treasure",
// "Food" — for the plain artifacts, and in the table's stat-first
// style for the ones a player identifies by their body. A key here
// must not collide with a tokenTable key; TestSpawnableTokenKeysAreUnique
// holds that.
var behaviourTokens = map[string]func() game.Card{
	"Treasure":   TreasureToken,
	"Gold":       GoldToken,
	"Food":       FoodToken,
	"Clue":       ClueToken,
	"Blood":      BloodToken,
	"Powerstone": PowerstoneToken,

	"0/1 colorless Eldrazi Spawn":                           EldraziSpawnToken,
	"2/2 blue Shapeshifter with changeling":                 BlueShapeshifterToken,
	"2/2 colorless Shapeshifter with changeling":            ColorlessShapeshifterToken,
	"3/3 colorless Phyrexian Wurm artifact with deathtouch": PhyrexianWurmDeathtouchToken,
	"3/3 colorless Phyrexian Wurm artifact with lifelink":   PhyrexianWurmLifelinkToken,
}

// TokenKeys lists every spawnable template key, sorted.
func (TokenLibrary) TokenKeys() []string {
	keys := make([]string, 0, len(tokenTable)+len(behaviourTokens))
	for k := range tokenTable {
		keys = append(keys, k)
	}
	for k := range behaviourTokens {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TokenTemplate returns a FRESH copy of the named template, or false
// for a key that is not in either table.
//
// Fresh matters: a caller that appends to Keywords or Colors on a
// shared template would leak into every later spawn of the same
// token. TokenCard already copies both slices; the behaviour
// constructors build a new Card each call.
func (TokenLibrary) TokenTemplate(key string) (game.Card, bool) {
	if build, ok := behaviourTokens[key]; ok {
		return build(), true
	}
	if _, ok := tokenTable[key]; ok {
		return TokenCard(key), true
	}
	return game.Card{}, false
}
