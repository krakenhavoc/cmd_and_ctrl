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

	// Grants are the ability bundles the token gives ANOTHER object
	// (ADR 0093, amended 2026-10-08): a Role token's "Enchanted
	// creature has 'Whenever this creature attacks, scry 1.'" is a
	// layer-6 grant in Static (GrantAbilitiesToAttached) naming a
	// bundle declared here. The bundle's abilities belong to the
	// RECIPIENT, so "this creature" is the enchanted creature, not the
	// token. Registered under game.GrantKey(Key) in the same defs map a
	// card's Spec.Grants file into, with the same boot-time checks, so
	// the keys are catalog-wide: namespace them with the token's slug.
	Grants []AbilityGrant

	// Back is the token's OTHER FACE, which makes the token a
	// double-faced token (CR 111.10i: the Incubator, whose back face is
	// the 0/0 Phyrexian artifact creature it transforms into; ADR 0185).
	// Nil for every single-faced token, which is every other one.
	//
	// The back is a template of its own so that a face can carry the
	// abilities the printed face does: its Mana, Activated, Triggered
	// and Static slots are filed under the back face's catalog key,
	// game.TokenKey(Slug) + "#1" — the same suffix a double-faced
	// CARD's back face files under, and the one game.CatalogKey asks for
	// once the token is turned over. Its Slug is not read (the front's
	// names the pair), and a back that prints nothing registers nothing:
	// a vanilla face needs no entry, exactly as a vanilla token does not.
	//
	// What this does NOT hold: the card-level fields. tokenFromCatalog
	// builds Layout, Faces and ActiveFace on the front's Card from the
	// two templates, so a template never spells them by hand and the
	// "flat printed fields equal Faces[ActiveFace]" invariant cannot be
	// written wrong.
	Back *tokenTemplate

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
	printedGingerbruteToken,
	printedClueToken,
	printedMapToken,
	printedBloodToken,
	printedPowerstoneToken,
	printedHeartwoodToken,
	printedLotusToken,
	printedEldraziSpawnToken,
	printedB28EldraziScionToken,
	printedB16LanderToken,
	printedB14ReplicatedRingToken,

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
	printedUrzaConstructToken,

	// #750: the first token whose printed text is a block rule.
	printedKurukSpiritToken,

	// ADR 0103: Spiked Corridor's Devil.
	printedDevilToken,

	// ADR 0114 PR 5: There and Back Again's Smaug.
	printedSmaugToken,

	// Weapons Manufacturing's Munitions.
	printedMunitionsToken,

	// #1945: the Role tokens.
	printedMonsterRoleToken,
	printedCursedRoleToken,
	printedRoyalRoleToken,
	printedWickedRoleToken,
	printedSorcererRoleToken,
	printedYoungHeroRoleToken,
	printedVirtuousRoleToken,
	printedChefRoleToken,

	// Twisted Sewer-Witch's Rat (#1945).
	printedCantBlockRatToken,

	// Reality Fracture: Skrelv's Hive and White Sun's Twilight's Mite.
	printedPhyrexianMiteToken,

	// ADR 0139 (#2796): the Jace planeswalker token Empower Jace makes,
	// the first token whose abilities are loyalty abilities.
	printedJaceToken,

	// Reality Fracture: Ajani Resolute's Ajani's Pridemate.
	printedAjanisPridemateToken,

	// Reality Fracture: Vraska, Soul of Stone's Sculpture Treasure.
	printedSculptureTreasureToken,

	// Hallowed Haunting's Spirit Cleric: a */* token that sizes itself.
	printedSpiritClericToken,

	// ADR 0185 (#2417): the Incubator, the first double-faced token.
	printedIncubatorToken,
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
	if why := checkTokenBack(t); why != "" {
		return why
	}
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

// checkTokenBack is the rules a double-faced token's second face must
// satisfy, "" when the template has none or the face is sound. It is
// checkTokenTemplate's first question because a malformed back would
// otherwise surface as a token that cannot transform.
func checkTokenBack(t tokenTemplate) string {
	b := t.Back
	if b == nil {
		return ""
	}
	switch {
	case b.Back != nil:
		// A third face is not a thing CR 712 describes: a transforming
		// permanent has exactly two.
		return "has a back face that has a back face of its own"
	case t.Card.TypeLine == "" || b.Card.TypeLine == "":
		// Both faces read their type line into the flat card on
		// SetFace; an empty one would turn the token into a blank.
		return "has a face with no type line"
	case b.Card.OracleID != "" || b.Card.TokenKey != "":
		return "has a back face that carries a catalog identity — the key is the front's, with a face suffix"
	case len(b.Card.ManaAbilities) > 0 || len(b.Card.ActivatedAbilities) > 0:
		return "has a back face that carries abilities on the instance — declare them in the slots"
	case b.hasAbilities() && b.Text == "":
		return "has a back face with abilities but no printed text"
	case t.Card.Layout != "" || len(t.Card.Faces) > 0 || t.Card.ActiveFace != 0:
		// tokenFromCatalog builds these from the two templates.
		return "spells its own Layout/Faces — tokenFromCatalog builds them from Back"
	}
	return ""
}

// hasAbilities reports whether the template declares any ability slot.
func (t tokenTemplate) hasAbilities() bool {
	return len(t.Mana) > 0 || len(t.Activated) > 0 || len(t.Triggered) > 0 ||
		len(t.Static) > 0 || len(t.BlockRules) > 0
}

// tokenBackKey is the catalog key of a double-faced token's back face:
// the token key with the "#<face>" suffix game.CatalogKey appends for a
// turned-over token (ADR 0185).
func tokenBackKey(slug string) string { return game.TokenKey(slug) + "#1" }

// tokenFace is one face of a double-faced token as game.Face data: the
// printed characteristics SetFace materialises when that face is up,
// and the printed text the hover overlay shows for a face with no
// printing behind it.
func tokenFace(t tokenTemplate) game.Face {
	return game.Face{
		Name:              t.Card.Name,
		TypeLine:          t.Card.TypeLine,
		ManaCost:          t.Card.ManaCost,
		Colors:            append([]string(nil), t.Card.Colors...),
		Power:             t.Card.Power,
		Toughness:         t.Card.Toughness,
		VariableToughness: t.Card.VariableToughness,
		OracleText:        t.Text,
		Keywords:          append([]string(nil), t.Card.Keywords...),
	}
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
		for i, ab := range t.Activated {
			checkActivatedAnswers(t.Slug+" token", fmt.Sprintf("activated ability %d", i), ab)
		}
		for i, m := range t.Mana {
			checkManaAnswers(t.Slug+" token", fmt.Sprintf("mana ability %d", i), m)
		}
		fileDef(key, buildTokenDef(t))
		if b := t.Back; b != nil && b.hasAbilities() {
			for i, ab := range b.Activated {
				checkActivatedAnswers(t.Slug+" token back face", fmt.Sprintf("activated ability %d", i), ab)
			}
			for i, m := range b.Mana {
				checkManaAnswers(t.Slug+" token back face", fmt.Sprintf("mana ability %d", i), m)
			}
			backKey := tokenBackKey(t.Slug)
			if _, taken := defs[backKey]; taken {
				panic(fmt.Sprintf("effects: two token templates claim %q", backKey))
			}
			fileDef(backKey, buildTokenDef(*b))
		}
		checkGrants(t.Slug, t.Grants)
		for _, gr := range t.Grants {
			fileDef(game.GrantKey(gr.Key), buildGrantDef(gr))
		}
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
	if t.Back != nil {
		// CR 111.10i, CR 712: a double-faced token has a front and a
		// back face and enters with the front up (ADR 0185). The flat
		// fields above ARE the front, so Faces[0] and the card agree
		// by construction.
		c.Layout = game.LayoutTransform
		c.Faces = []game.Face{tokenFace(t), tokenFace(*t.Back)}
		c.ActiveFace = 0
	}
	return c
}
