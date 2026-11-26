package teststo

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_realdump_manual_test.go — #1127, and the guard ADR 0078
// decision 8 asks for. Opt-in against the real Scryfall bulk dump,
// gated on CMDCTRL_SCRYFALL_DUMP so ordinary CI never touches it (the
// file is ~630MB and is not in the repo):
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/cards/effects/ -run RealDumpEveryTokenTemplate -v
//
// e2e-nightly.yml has the dump and runs it there.
//
// WHY A TOKEN TEMPLATE IS CHECKED AGAINST A PRINTED TOKEN CARD. A
// template's Name, type line, P/T and Colors are the token's
// CHARACTERISTICS, and the printed token card is the published record
// of what they are. #1127 is what happens without this test: 21 rows
// declared no colour where every printing of that token is coloured,
// which is a rules bug — a colourless Soldier dodges protection from
// white, is missed by a white lord, survives "destroy all nonwhite
// creatures" and counts wrongly anywhere `Colors` is read — and it sat
// in the table until somebody matched the whole table against the dump
// by hand.
//
// It is also the precondition for ADR 0078's art: a template that
// matches no printing can be given no picture, so this list and that
// one are the same list.
//
// The match is ADR 0078 decision 3's identity rule, minus the art
// half: pool filters (3a) then name / type line / P/T / colours (3b).
// Keywords are decision 3c's soft preference and are deliberately NOT
// compared — the template's Keywords are what the ENGINE grants and
// Scryfall's are what that piece of card prints, and the two disagree
// harmlessly (a Wurmcoil half carries deathtouch; the shared base row
// carries nothing). When #1115 builds the resolver, this test should
// call it instead of re-implementing the rule here.

// knownUnmatchedTokens are the templates that deliberately match no
// printed token card, each with the reason. The test fails on a set
// difference in EITHER direction: a template that stops matching is a
// regression, and a template on this list that starts matching is a
// stale exemption to delete.
var knownUnmatchedTokens = map[string]string{
	// Court of Embereth (Wilds of Eldraine Commander) creates "a 3/1 red
	// Knight creature token". The only Knight token cards in the dump are 2/2
	// white and 3/3 white ones; no 3/1 red Knight was printed. The template
	// follows the oracle text.
	"3/1 red Knight": "no 3/1 red Knight token card was printed for Court of Embereth",

	// Spawning Pit is a 2004 artifact and its token — "create a 2/2
	// colorless Spawn artifact creature token" — was never printed as
	// a token card; token cards start in 2007. The template is right
	// and the dump simply has no such record. (The dump's "Spawn" is
	// a 3/3 red one from a different card, and "Eldrazi Spawn" is its
	// own token this table already carries.)
	"2/2 colorless Spawn artifact": "no token card was ever printed for Spawning Pit's Spawn",

	// The Eagles Are Coming! (The Hobbit) creates "a 4/4 white Bird
	// Soldier creature token with flying". The dump's only 4/4 Bird
	// Soldier token card (thob) lists no colour, but the token is what
	// the card's oracle text makes it: white, which matters for
	// protection and "destroy target white creature". The template
	// follows the oracle text, as the table's header says it must.
	"4/4 white Bird Soldier with flying": "the printed Bird Soldier token card carries no colour in the dump; the oracle text of The Eagles Are Coming! says white",

	// Carrion (Visions, 1997) creates "X 0/1 black Insect creature
	// tokens". No 0/1 Insect token card of any colour exists in the
	// dump (a Scryfall search for t:token t:insect pow=0 tou=1 finds
	// none), for the same reason as Spawning Pit: the card predates
	// token cards. The template follows the oracle text (#2096).
	"0/1 black Insect": "no token card was ever printed for Carrion's Insects",

	// Tombstone Stairwell (Mirage, 1996) creates "a 2/2 black Zombie
	// creature token with haste named Tombspawn" (#1997). It predates
	// token cards, and no Tombspawn token card was ever printed. The
	// template follows the oracle text.
	"2/2 black Zombie with haste named Tombspawn": "no token card was ever printed for Tombstone Stairwell's Tombspawn",

	// Nested Ghoul (Mirrodin, 2003) creates "a 2/2 black Phyrexian
	// Zombie creature token". It predates token cards, and the dump has
	// no 2/2 black Phyrexian Zombie token card (the Phyrexian subtype
	// came with the 2021 oracle pass). The template follows the oracle
	// text.
	"2/2 black Phyrexian Zombie": "no Phyrexian Zombie token card was ever printed for Nested Ghoul",

	// Wurmcoil Engine's current oracle says "3/3 colorless Phyrexian
	// Wurm artifact creature token"; the printed token card, from
	// 2010, is named "Wurm" — it predates the 2021 Phyrexian subtype
	// pass. Decision (#1127): keep "Phyrexian Wurm". The engine
	// follows the current oracle because that is what the RULES read
	// — "Phyrexian" and "Wurm" are both subtypes a lord or a removal
	// spell can name — and a token card is a prop, not a rules text.
	// The cost is that this row keeps the text fallback until someone
	// pins an id through ADR 0078 decision 6's override seam.
	"3/3 colorless Phyrexian Wurm artifact": "printed token card says \"Wurm\"; the engine follows the current oracle",

	// Shuri, Vibranium Technologist (Marvel Super Heroes Commander) creates
	// "a 1/1 colorless Robot Hero artifact creature token with flying". The
	// only printed Robot Hero token card (tmsc) is 2/1, so no printing
	// matches a 1/1; the template follows the card's oracle text (#4455).
	"1/1 colorless Robot Hero artifact with flying": "the only printed Robot Hero token card is 2/1; Shuri, Vibranium Technologist's oracle text says 1/1",

	// Blessed Sanctuary (Jumpstart) creates "a 2/2 white Unicorn creature
	// token". The dump has no Unicorn token card of that size and colour,
	// so no printing matches; the template follows the card's oracle text
	// (#4893).
	"2/2 white Unicorn": "no 2/2 white Unicorn token card was ever printed; Blessed Sanctuary's oracle text says 2/2 white",
	// Wurmquake (Phyrexia: All Will Be One, #5479) creates "an X/X green
	// Phyrexian Wurm creature token with trample and toxic 1". Its printed
	// token card (ONC) is */*, which a fixed size can never match; the
	// template is 0/0 and the card sets X as it creates the token.
	"0/0 green Phyrexian Wurm with trample and toxic 1": "the printed token card is */*; Wurmquake sets X/X as it creates it",
	// Kura, the Boundless Sky (Kamigawa: Neon Dynasty) creates "an X/X green Spirit
	// creature token, where X is the number of lands you control". The printed token
	// card is */*; the card sets X/X as it creates the token.
	"0/0 green Spirit": "the printed token card is */*; Kura, the Boundless Sky sets X/X as it creates it",
	// Stitcher Geralf creates "an X/X blue Zombie creature token"; the printed token card is */*.
	"0/0 blue Zombie": "the printed token card is */*; Stitcher Geralf sets X/X as it creates it",
	// Corpseweft creates "a tapped X/X black Zombie Horror creature token"; the printed token card is */*.
	"0/0 black Zombie Horror": "the printed token card is */*; Corpseweft sets X/X as it creates it",
	// Soul Separator creates a black Zombie with power and toughness equal to the exiled card's; the printed token card is */*.
	"0/0 black Zombie": "the printed token card is */*; Soul Separator sets its power and toughness as it creates it",
	// Cynette, Jelly Drover creates "a 1/1 blue Jellyfish creature token with flying"; Scryfall lists no token card for her.
	"1/1 blue Jellyfish with flying": "no Jellyfish token card was printed for Cynette, Jelly Drover",
	// Psemilla, Meletian Poet (Foundations Jumpstart) creates "a 2/2 white Nymph
	// enchantment creature token". No Nymph token card was ever printed, so the
	// template follows the card's oracle text (as Shuri's Robot Hero does).
	"2/2 white Nymph enchantment creature": "no Nymph token card was ever printed; Psemilla, Meletian Poet's oracle text says a 2/2 white Nymph enchantment creature",
	// Balefang the Unslayable creates "a tapped Baneslayer Angel token"; Scryfall lists no token card for it.
	"5/5 white Angel named Baneslayer Angel": "no Baneslayer Angel token card was printed for Balefang the Unslayable",
	// Nivea, Beloved Battlemage (Mystery Booster Commander) creates an Akroma, Angel of Wrath
	// token; no such token card is in the dump, so the template follows the card's oracle text.
	"6/6 white Legendary Angel named Akroma, Angel of Wrath with flying, first strike, vigilance, trample, haste and protection from black and from red": "no Akroma token card is in the dump; Nivea's oracle text defines it",

	// Jeskai Baller (Mystery Booster 2) creates "a 1/1 white Athlete creature
	// token". The Athlete token was never printed as a token card; the
	// template follows the oracle text.
	"1/1 white Athlete": "no Athlete token card was printed for Jeskai Baller",

	// Worzel, the Protector (Mystery Booster Commander) creates "ten Scryb Sprites tokens". Scryb Sprites
	// is printed only as a card ("Creature — Faerie"), never as a token card; the template follows the oracle text.
	"1/1 green Faerie with flying named Scryb Sprites": "Scryb Sprites was never printed as a token card; Worzel's oracle text says a green 1/1 Faerie with flying",

	// Monkey Cage (Jumpstart 2022, Mercadian Masques) creates "X 2/2 green Monkey
	// creature tokens". No Monkey token card was printed for it; the template follows
	// the oracle text.
	"2/2 green Monkey": "no Monkey token card was printed for Monkey Cage",
	// Questing Phelddagrif (Secret Lair Drop) makes "a 1/1 green Hippo creature token"; the only Hippo token card in the dump is a 3/3.
	"1/1 green Hippo": "no 1/1 green Hippo token card in the dump; Questing Phelddagrif prints the token in its text",
	// Marath, Will of the Wild (Commander Legends) creates "an X/X green Elemental creature
	// token". The printed token card is */*, which a fixed size can never match; the template is
	// 0/0 and the card sets X/X as it creates the token.
	"0/0 green Elemental": "the printed token card is */*; Marath, Will of the Wild sets X/X as it creates it",
	// Profane Transfusion (Commander Legends) creates "an X/X colorless Horror artifact creature
	// token". The printed token card is */*, which a fixed size can never match; the template is
	// 0/0 and the card sets X/X as it creates the token.
	"0/0 colorless Horror artifact": "the printed token card is */*; Profane Transfusion sets X/X as it creates it",
	// Grakmaw, Skyclave Ravager creates "an X/X black and green Hydra creature token"; the printed
	// token card is */* (and lists no colours), so a fixed size never matches.
	"0/0 black and green Hydra": "the printed token card is */*; Grakmaw, Skyclave Ravager sets X/X as it creates it",
	// Experimental Overload (Core Set 2021) creates "an X/X blue and red Weird creature token". The printed
	// token card is */*, which a fixed size can never match; the template is 0/0 and the card sets X/X as it
	// creates the token.
	"0/0 blue and red Weird": "the printed token card is */*; Experimental Overload sets X/X as it creates it",
	// Chainer's Torment (Dominaria) chapter III creates "an X/X black Nightmare Horror creature token". The
	// printed token card is */*, which a fixed size can never match; the template is 0/0 and the card sets
	// X/X as it creates the token.
	"0/0 black Nightmare Horror": "the printed token card is */*; Chainer's Torment sets X/X as it creates it",
	// Kin-Tree Invocation (Khans of Tarkir) creates "an X/X black and green Spirit Warrior creature token".
	// The printed token card is */*, which a fixed size can never match; the template is 0/0 and the card
	// sets X/X as it creates the token.
	"0/0 black and green Spirit Warrior": "the printed token card is */*; Kin-Tree Invocation sets X/X as it creates it",

	// Thunderheads (Weatherlight) creates "a 3/3 blue Weird creature token with defender and flying". The
	// only printed token is the Weird face of the Weird // Goblin double-faced token (GK1), and the match
	// reads a token card's top-level characteristics, not its faces.
	"3/3 blue Weird with defender and flying": "printed only as a face of a double-faced token (Thunderheads)",
}

// printedToken is the slice of a Scryfall record this match needs.
type printedToken struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Lang        string            `json:"lang"`
	Layout      string            `json:"layout"`
	TypeLine    string            `json:"type_line"`
	Power       string            `json:"power"`
	Toughness   string            `json:"toughness"`
	Colors      []string          `json:"colors"`
	BorderColor string            `json:"border_color"`
	SetType     string            `json:"set_type"`
	ImageURIs   map[string]string `json:"image_uris"`
}

func TestRealDumpEveryTokenTemplateMatchesAPrintedToken(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to match the token table against the real dump")
	}
	pool := loadPrintedTokenPool(t, path)
	t.Logf("eligible printed token records: %d", len(pool))
	if len(pool) < 1000 {
		t.Fatalf("only %d eligible token printings — the pool filter or the dump is wrong", len(pool))
	}
	byName := make(map[string][]printedToken, len(pool))
	for _, r := range pool {
		n := normalizeTokenName(r.Name)
		byName[n] = append(byName[n], r)
	}

	templates := tokenTemplatesUnderTest()
	unmatched := map[string]string{}
	display := map[string]string{}
	for _, label := range templates {
		display[label.key] = label.display
		if matchPrintedToken(byName, label.card) == 0 {
			unmatched[label.key] = describeTemplate(label.card)
		}
	}

	for key, why := range unmatched {
		if _, ok := knownUnmatchedTokens[key]; !ok {
			t.Errorf("%s matches no printed token card (%s).\n"+
				"\tEither the template's characteristics are wrong — check the oracle text of the card "+
				"that makes it — or it is a deliberate exemption and belongs in knownUnmatchedTokens with a reason.",
				display[key], why)
		}
	}
	for key := range knownUnmatchedTokens {
		if _, ok := unmatched[key]; !ok {
			t.Errorf("%q now matches a printed token card: remove it from knownUnmatchedTokens", key)
		}
	}
	t.Logf("%d templates checked, %d unmatched (all expected)", len(templates), len(unmatched))
}

type labelledTemplate struct {
	key     string // the table key, or the constructor name
	display string // how the key reads in a failure message
	card    game.Card
}

// tokenTemplatesUnderTest is every plain row plus the behaviour
// templates that live in tokens.go — the ones with a mana or activated
// ability, which are still printed token cards and still have to be
// right.
func tokenTemplatesUnderTest() []labelledTemplate {
	out := make([]labelledTemplate, 0, len(effects.TokenTable)+12)
	for _, k := range effects.TokenKeys() {
		out = append(out, labelledTemplate{key: k, display: strconv.Quote(k), card: effects.TokenCard(k)})
	}
	for _, bt := range []struct {
		name string
		card game.Card
	}{
		{"TreasureToken()", effects.TreasureToken()},
		{"GoldToken()", effects.GoldToken()},
		{"FoodToken()", effects.FoodToken()},
		{"ClueToken()", effects.ClueToken()},
		{"BloodToken()", effects.BloodToken()},
		{"PowerstoneToken()", effects.PowerstoneToken()},
		{"HeartwoodToken()", effects.HeartwoodToken()},
		{"MutagenToken()", effects.MutagenToken()},
		{"EldraziSpawnToken()", effects.EldraziSpawnToken()},
		{"BlueShapeshifterToken()", effects.BlueShapeshifterToken()},
		{"ColorlessShapeshifterToken()", effects.ColorlessShapeshifterToken()},
		{"SmaugToken()", effects.SmaugToken()},
	} {
		out = append(out, labelledTemplate{key: bt.name, display: bt.name, card: bt.card})
	}
	// #1236: the Army token is DERIVED from the amass keyword's
	// subtype rather than tabled (ADR 0087 decision 2), so the dump is
	// the only thing that can say the derivation is right. Every
	// species printed today, which is the whole set a card can ask
	// for — a ninth from a future set becomes a row here on the day
	// the card that amasses it is registered.
	for _, subtype := range []string{"Zombie", "Orc", "Goblin", "Sliver"} {
		name := "ArmyToken(" + strconv.Quote(subtype) + ")"
		out = append(out, labelledTemplate{key: name, display: name, card: effects.ArmyToken(subtype)})
	}
	return out
}

// matchPrintedToken counts the pool records whose identity equals the
// template's: name, (card types, subtypes), printed P/T, colours.
func matchPrintedToken(byName map[string][]printedToken, tmpl game.Card) int {
	_, wantTypes, wantSubs := game.ParseTypeLine(tmpl.TypeLine)
	wantP, wantT := "", ""
	if effectstest.ContainsString(wantTypes, "Creature") {
		wantP, wantT = strconv.Itoa(tmpl.Power), strconv.Itoa(tmpl.Toughness)
	}
	n := 0
	for _, r := range byName[normalizeTokenName(tmpl.Name)] {
		_, gotTypes, gotSubs := game.ParseTypeLine(r.TypeLine)
		if !sameStringSet(gotTypes, wantTypes) || !sameStringSet(gotSubs, wantSubs) {
			continue
		}
		if r.Power != wantP || r.Toughness != wantT {
			continue
		}
		if !sameStringSet(r.Colors, tmpl.Colors) {
			continue
		}
		n++
	}
	return n
}

// loadPrintedTokenPool streams the dump and keeps the records ADR
// 0078 decision 3a admits: an English, black-bordered `layout: token`
// printing with art, from a set whose type is not memorabilia,
// minigame or funny.
func loadPrintedTokenPool(t *testing.T, path string) []printedToken {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	dec := json.NewDecoder(f)
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("read opening token: %v", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		t.Fatalf("expected a top-level JSON array, got %v", tok)
	}
	var pool []printedToken
	for dec.More() {
		var r printedToken
		if err := dec.Decode(&r); err != nil {
			t.Fatalf("decode record: %v", err)
		}
		switch {
		case r.Layout != "token",
			r.Lang != "en",
			len(r.ImageURIs) == 0,
			r.BorderColor != "black",
			r.SetType == "memorabilia", r.SetType == "minigame", r.SetType == "funny":
			continue
		}
		pool = append(pool, r)
	}
	return pool
}

// normalizeTokenName is index.go's normalizeName: trim, collapse
// internal whitespace, lowercase.
func normalizeTokenName(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as, bs := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(as)
	sort.Strings(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

func describeTemplate(c game.Card) string {
	colors := "colourless"
	if len(c.Colors) > 0 {
		colors = strings.Join(c.Colors, "")
	}
	return fmt.Sprintf("name %q, type line %q, %d/%d, colours %s",
		c.Name, c.TypeLine, c.Power, c.Toughness, colors)
}
