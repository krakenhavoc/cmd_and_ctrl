package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_catalog.go is the catalog half of #521: a token template's
// printed abilities become a catalog entry instead of a bundle of Go
// closures riding on every token minted from it.
//
// # Why
//
// A token has no printing and so no oracle ID, and every catalog hook
// keys on one. So Treasure, Food, Clue, Blood, Gold, the Eldrazi
// Spawn and the rest carried their abilities on the template as
// closures, and the snapshot could not write them down. ADR 0041's
// rule is that a snapshot holding an unserialisable continuation is
// not a restore point, so ONE Treasure on the battlefield meant no
// restore point was written at all, for as long as it sat there — a
// deploy rewound the table past every turn the Treasure survived.
// Food, Clue and Blood also lost their activated abilities on the way
// to the client, because the wire projection skipped any card with no
// oracle ID: they reached the table inert.
//
// # How
//
// Exactly the way emblem.go and ability_grant.go already do it. The
// template's abilities are built once at init and filed in the same
// `defs` map cards are filed in, under a synthetic key from
// game.TokenKey — "token:treasure", "token:food". The template itself
// keeps no closures: it carries the KEY, and every reader that used
// to take the instance-ability route now takes the ordinary catalog
// route (ManaAbilitiesForCard, ActivatedAbilitiesForCard), which is
// the route printed cards and token copies have always taken.
//
// They go in `defs` and NOT in `registry`: a token is not a card, and
// All() — which the coverage census and the deck builders read — must
// keep counting cards.
//
// # What did not change
//
// A token COPY (CR 707.2) carries the copied card's real oracle ID
// and keeps resolving through it; game.CatalogKey prefers an oracle
// ID and reads the token key only when there is none. And a vanilla
// token — a 1/1 Soldier, a 2/2 Shapeshifter with changeling — needs
// no entry here at all: printed keywords and stats are plain data on
// the template already and always survived a round trip.

// tokenTemplateBuilder returns a token template with its printed
// abilities still attached. It is the single source of truth for one
// token: init registers the abilities it declares, and
// tokenFromCatalog hands out the rest of it.
type tokenTemplateBuilder func() game.Card

// tokenTemplates maps a token's catalog slug to its builder. Every
// token whose template declares a mana or an activated ability must
// be in here — one that is not would still carry closures on the
// instance and would still block a restore point.
//
// The slug is the token's printed name, lowercased and hyphenated.
// It is a stable on-disk identity: a snapshot carries it and a
// restore looks it up, so renaming a slug orphans the tokens already
// written to disk under the old one. Rename with the same care a
// database column gets.
var tokenTemplates = map[string]tokenTemplateBuilder{
	"treasure":                printedTreasureToken,
	"gold":                    printedGoldToken,
	"food":                    printedFoodToken,
	"clue":                    printedClueToken,
	"blood":                   printedBloodToken,
	"powerstone":              printedPowerstoneToken,
	"eldrazi-spawn":           printedEldraziSpawnToken,
	"eldrazi-scion":           printedB28EldraziScionToken,
	"lander":                  printedB16LanderToken,
	"replicated-ring":         printedB14ReplicatedRingToken,
	"springleaf-shapeshifter": printedB18SpringleafShapeshifterToken,
}

func init() {
	for slug, build := range tokenTemplates {
		key := game.TokenKey(slug)
		if _, taken := defs[key]; taken {
			panic(fmt.Sprintf("effects: two token templates claim %q", key))
		}
		c := build()
		if len(c.ManaAbilities) == 0 && len(c.ActivatedAbilities) == 0 {
			// A template with nothing to register would hand out a
			// key that resolves to an empty entry, which reads as
			// "this token's abilities were removed" everywhere
			// downstream. A vanilla token belongs in tokens_table.go,
			// not here.
			panic(fmt.Sprintf("effects: token template %q declares no abilities — it needs no catalog key", slug))
		}
		if c.OracleID != "" {
			// Both identities on one object is the one shape
			// game.CatalogKey cannot express: it would silently
			// prefer the oracle ID and the token's own abilities
			// would never be found.
			panic(fmt.Sprintf("effects: token template %q declares an oracle ID as well as a token key", slug))
		}
		defs[key] = &game.CardDef{
			ManaAbilities: c.ManaAbilities,
			Activated:     c.ActivatedAbilities,
		}
	}
}

// tokenFromCatalog returns the template for a registered token: the
// printed fields, with the abilities left in the catalog and the key
// that finds them stamped on.
//
// The builder is re-run per call rather than a cached Card being
// handed out, so each token template is a fresh value with its own
// slices — the invariant the constructors have always had, and the
// one mintTokenLocked's Keywords append depends on.
//
// Panics on an unregistered slug: it is a typo in a card file, at
// boot or on the first token of its kind, and a silently abilityless
// Treasure is far worse than a loud one.
func tokenFromCatalog(slug string) game.Card {
	build, ok := tokenTemplates[slug]
	if !ok {
		panic(fmt.Sprintf("effects: no token template registered for %q", slug))
	}
	c := build()
	c.ManaAbilities = nil
	c.ActivatedAbilities = nil
	c.TokenKey = game.TokenKey(slug)
	return c
}
