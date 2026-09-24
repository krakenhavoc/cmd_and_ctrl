package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_catalog.go is the catalog half of #521, widened to every
// ability slot by #1248 / ADR 0083: a token template's printed
// abilities become a catalog entry instead of a bundle of Go closures
// riding on every token minted from it.
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
// route (ManaAbilitiesForCard, ActivatedAbilitiesForCard,
// TriggersForCard, StaticAbilitiesForCard), which is the route
// printed cards and token copies have always taken.
//
// They go in `defs` and NOT in `registry`: a token is not a card, and
// All() — which the coverage census and the deck builders read — must
// keep counting cards.
//
// # Four slots, not two (ADR 0083)
//
// #521 shipped the mana and activated slots, which were the ones that
// blocked a restore point. A token's own TRIGGER ("when this token
// dies, create a 2/2 red Dragon") and its own STATIC had nowhere to
// be written at all, because the builder returned a `game.Card` and
// `game.Card` has no triggered or static field — deliberately, since
// closures on the instance are the thing #521 deleted. So the builder
// returns a `tokenTemplate` instead: the printed characteristics, the
// four ability slots beside each other, and the printed text.
//
// Nothing in `internal/game` needed a new dispatch path for it.
// `game.CatalogKey` has answered with `Card.TokenKey` since #521, so
// the trigger harvest, the LTB harvest, the layer pass and the
// activation path all found a token's entry the moment the entry had
// something in the slot.
//
// # What did not change
//
// A token COPY (CR 707.2) carries the copied card's real oracle ID
// and keeps resolving through it; game.CatalogKey prefers an oracle
// ID and reads the token key only when there is none. And a vanilla
// token — a 1/1 Soldier, a 2/2 Shapeshifter with changeling — needs
// no entry here at all: printed keywords and stats are plain data on
// the template already and always survived a round trip.

// tokenTemplate is one token as PRINTED: its identity, what it is,
// what it does, and what the words on it say.
//
// It is the single source of truth for one token. `init` registers
// the four ability slots under the token's catalog key, and
// `tokenFromCatalog` hands out the `Card` with that key stamped on
// and no closures attached.
type tokenTemplate struct {
	// Slug is the token's catalog identity — the last segment of
	// game.TokenKey(Slug), the token's printed name lowercased and
	// hyphenated.
	//
	// It is a stable ON-DISK identity: a snapshot carries the key and
	// a restore looks it up, so renaming a slug orphans the tokens
	// already written to disk under the old one. Rename with the same
	// care a database column gets.
	//
	// A name that alone would be ambiguous (two different Dragons)
	// takes a qualifier from what makes it distinct —
	// "dragon-firebending", "dragon-nesting" — because the slug is an
	// identity and not a label. Two cards that print the SAME token
	// share one slug: Beledros Witherbloom and Sedgemoor Witch both
	// make `pest`.
	//
	// It lives on the template rather than in the registry's key so
	// that the constructors below need no map lookup. That is not
	// taste: a token template may name ANOTHER token (the Goblin
	// Shaman makes a Treasure), and a registry keyed by slug would
	// then be a package-level variable whose initialiser reaches back
	// into itself — an initialization cycle Go refuses to compile.
	Slug string

	// Card is the token's printed CHARACTERISTICS — name, type line,
	// P/T, colours, printed keywords. Plain data, and the only part
	// of this struct that reaches the battlefield.
	//
	// It must not carry ManaAbilities or ActivatedAbilities: those
	// are closures on the instance, which is what the token key
	// exists to avoid. Declare them in the slots below.
	Card game.Card

	// Mana, Activated, Triggered and Static are the token's printed
	// ABILITIES, written exactly as a printed card's are — the same
	// shapes, the same triggers_common.go constructors, the same
	// game.StaticAbility. They are registered in the catalog under
	// game.TokenKey(Slug) and are never copied onto the instance.
	Mana      []game.ManaAbilityShape
	Activated []game.ActivatedAbilityShape
	Triggered []game.TriggeredAbility
	Static    []game.StaticAbility

	// BlockRules are the token's printed CR 509.1b block restrictions
	// with a parameter — Avatar Kuruk's Spirit's "This token can't
	// block or be blocked by non-Spirit creatures" (#750). Built with
	// the block_rules.go constructors, exactly as a card's
	// Spec.BlockRules are, and registered under the same token key.
	BlockRules []game.BlockRule

	// Text is the token's printed ability text, verbatim, as it
	// appears on the printed token card ("When this token dies,
	// create a 2/2 red Dragon creature token with flying.").
	//
	// Required for every template here, and that is the point: a
	// token has no printing behind it for the client to fetch oracle
	// text from, and a trigger's Label is a log line, not card text.
	// Without this the player would watch a token trigger fire with
	// no way to know it was coming. It reaches the board through
	// game.TokenTextForCard and CardView.TokenText.
	Text string
}

// tokenTemplateBuilder returns a token template with its printed
// abilities still attached. Re-run per call rather than a cached
// value being handed out, so each template is fresh and its slices
// are its own — the invariant the constructors have always had, and
// the one mintTokenLocked's Keywords append depends on.
type tokenTemplateBuilder func() tokenTemplate

// tokenTemplates is every token whose template declares a mana,
// activated, triggered or static ability, or a block rule. A token
// that declares none is a row in tokens_table.go and needs no entry
// here; one that is not here could not declare an ability at all.
//
// A LIST and not a map, because the slug lives on the template — see
// tokenTemplate.Slug for the initialization cycle that shape avoids.
// init builds the by-slug index from it and refuses a duplicate.
var tokenTemplates = []tokenTemplateBuilder{
	printedTreasureToken,
	printedGoldToken,
	printedFoodToken,
	printedClueToken,
	printedBloodToken,
	printedPowerstoneToken,
	printedEldraziSpawnToken,
	printedB28EldraziScionToken,
	printedB16LanderToken,
	printedB14ReplicatedRingToken,
	printedB18SpringleafShapeshifterToken,

	// ADR 0083's first tokens with an ability of their own that is
	// not a mana or activated ability.
	printedPestToken,
	printedNoncreatureCastWizardToken,
	printedFableGoblinShamanToken,
	printedFirebendingDragonToken,
	printedReefWormFishToken,
	printedReefWormWhaleToken,
	printedDragonEggToken,
	printedNestingDragonDragonToken,

	// #750: the first token whose printed text is a block rule.
	printedKurukSpiritToken,
}

// tokenTemplatesBySlug indexes the list above. Written once by init
// and read-only afterwards; the tests walk it, and nothing in the
// production path needs it (a constructor holds its own builder).
var tokenTemplatesBySlug = map[string]tokenTemplateBuilder{}

// buildTokenDef projects one template into the engine's view of it.
// Split out of init so the registration rule can be exercised on a
// template that is not in the list.
func buildTokenDef(t tokenTemplate) *game.CardDef {
	return &game.CardDef{
		ManaAbilities: t.Mana,
		Activated:     t.Activated,
		Triggered:     t.Triggered,
		Static:        t.Static,
		BlockRules:    t.BlockRules,
		TokenText:     t.Text,
	}
}

// checkTokenTemplate is every rule a template must satisfy to be
// registrable, as one function so a test can state the rules rather
// than provoke them at boot. Returns the reason it is not, or "".
func checkTokenTemplate(t tokenTemplate) string {
	switch {
	case t.Slug == "":
		return "declares no slug — the slug is its catalog identity"
	case len(t.Mana) == 0 && len(t.Activated) == 0 && len(t.Triggered) == 0 && len(t.Static) == 0 && len(t.BlockRules) == 0:
		// A template with nothing to register would hand out a key
		// that resolves to an empty entry, which reads as "this
		// token's abilities were removed" everywhere downstream. A
		// vanilla token belongs in tokens_table.go, not here.
		return "declares no abilities — it needs no catalog key"
	case t.Text == "":
		// A token that DOES something the player cannot read is the
		// failure ADR 0083 decision 6 is about: a token has no
		// printing to fetch oracle text from, so if the template does
		// not say what it prints, nothing can.
		return "declares abilities but no printed text"
	case t.Card.OracleID != "":
		// Both identities on one object is the one shape
		// game.CatalogKey cannot express: it would silently prefer
		// the oracle ID and the token's own abilities would never be
		// found.
		return "declares an oracle ID as well as a token key"
	case t.Card.TokenKey != "":
		// The key is stamped by tokenFromCatalog, from the slug. A
		// second spelling on the Card is a second source of truth.
		return "stamps its own TokenKey — tokenFromCatalog does that, from the slug"
	case len(t.Card.ManaAbilities) > 0 || len(t.Card.ActivatedAbilities) > 0:
		// The whole point of the key: abilities belong in the slots,
		// never on the instance, or the snapshot cannot write the
		// board down (#521).
		return "carries abilities on the instance — declare them in the slots"
	}
	return ""
}

func init() {
	for _, build := range tokenTemplates {
		t := build()
		if why := checkTokenTemplate(t); why != "" {
			panic(fmt.Sprintf("effects: token template %q %s", t.Slug, why))
		}
		key := game.TokenKey(t.Slug)
		if _, taken := defs[key]; taken {
			panic(fmt.Sprintf("effects: two token templates claim %q", key))
		}
		defs[key] = buildTokenDef(t)
		tokenTemplatesBySlug[t.Slug] = build
	}
}

// tokenFromCatalog returns a registered token's template: the printed
// fields, with the abilities left in the catalog and the key that
// finds them stamped on.
//
// It takes the BUILDER rather than a slug, which is what keeps a
// token template free to name another token (the Goblin Shaman makes
// a Treasure) without the package-variable cycle tokenTemplate.Slug
// describes. The builder is the same value the list registers, so
// there is no second spelling of the slug to get wrong.
func tokenFromCatalog(build tokenTemplateBuilder) game.Card {
	t := build()
	c := t.Card
	c.TokenKey = game.TokenKey(t.Slug)
	return c
}
