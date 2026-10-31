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

// predefinedTokens are the predefined tokens of CR 111.10 that the
// catalog has a template for, keyed by the name the rule gives each
// one ("A Treasure token is a colorless Treasure artifact token with
// …"). The spawner offers them under these keys.
//
// They are apart from behaviourTokens because a predefined token is
// defined by the rules, not by a printing: a Treasure is the same
// object whichever Scryfall card pictures it. That is what lets
// PredefinedToken turn a token PRINTING into the template (#4738).
//
// CR 111.10 defines a few more. Walker (111.10d) is a vanilla 2/2, so
// its printing is already the token. Shard (111.10e), Junk (111.10t),
// Mutagen (111.10v) and the Role tokens (111.10j–r) have no template
// here, so a printing of one stays a printing, and the table is told
// the engine does not run its text.
var predefinedTokens = map[string]func() game.Card{
	"Treasure":   TreasureToken,   // CR 111.10a
	"Food":       FoodToken,       // CR 111.10b
	"Gold":       GoldToken,       // CR 111.10c
	"Clue":       ClueToken,       // CR 111.10f
	"Blood":      BloodToken,      // CR 111.10g
	"Powerstone": PowerstoneToken, // CR 111.10h
	"Incubator":  IncubatorToken,  // CR 111.10i
	"Map":        MapToken,        // CR 111.10s
	"Lander":     b16LanderToken,  // CR 111.10u
	"Vibranium":  VibraniumToken,  // CR 111.10w
	"Heartwood":  HeartwoodToken,  // CR 111.10x
}

// behaviourTokens are the other templates that are constructors rather
// than table rows: every one of them carries an ability, which is code
// and cannot be a row.
//
// Keyed in the table's stat-first style, because a player identifies
// them by their body. A key here must not collide with a tokenTable or
// a predefinedTokens key; TestSpawnableTokenKeysAreUnique holds that.
var behaviourTokens = map[string]func() game.Card{
	"0/1 colorless Eldrazi Spawn":                           EldraziSpawnToken,
	"2/2 blue Shapeshifter with changeling":                 BlueShapeshifterToken,
	"2/2 colorless Shapeshifter with changeling":            ColorlessShapeshifterToken,
	"3/3 colorless Phyrexian Wurm artifact with deathtouch": PhyrexianWurmDeathtouchToken,
	"3/3 colorless Phyrexian Wurm artifact with lifelink":   PhyrexianWurmLifelinkToken,
}

// TokenKeys lists every spawnable template key, sorted.
func (TokenLibrary) TokenKeys() []string {
	keys := make([]string, 0, len(tokenTable)+len(predefinedTokens)+len(behaviourTokens))
	for k := range tokenTable {
		keys = append(keys, k)
	}
	for k := range predefinedTokens {
		keys = append(keys, k)
	}
	for k := range behaviourTokens {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TokenTemplate returns a FRESH copy of the named template, or false
// for a key that is in none of the tables.
//
// Fresh matters: a caller that appends to Keywords or Colors on a
// shared template would leak into every later spawn of the same
// token. TokenCard already copies both slices; the behaviour
// constructors build a new Card each call.
func (l TokenLibrary) TokenTemplate(key string) (game.Card, bool) {
	if t, ok := l.PredefinedToken(key); ok {
		return t, true
	}
	if build, ok := behaviourTokens[key]; ok {
		return build(), true
	}
	if _, ok := tokenTable[key]; ok {
		return TokenCard(key), true
	}
	return game.Card{}, false
}

// PredefinedToken returns a fresh copy of the template for the
// predefined token (CR 111.10) with this name, or false when the name
// is not one the catalog has a template for. The name is the token's
// own, as a Scryfall printing of it spells it: "Treasure", "Incubator".
func (TokenLibrary) PredefinedToken(name string) (game.Card, bool) {
	build, ok := predefinedTokens[name]
	if !ok {
		return game.Card{}, false
	}
	return build(), true
}
