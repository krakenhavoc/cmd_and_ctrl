package draft

import (
	"regexp"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// facts.go is where a card's Facts come from (ADR 0222 §7, #5439): the
// printing in the Scryfall index, and what the catalog declares the card
// does. Nothing else: no ratings file, no outside draft data.

// CatalogFacts is a FactsFunc over the server's card index and the
// catalog. It reads, for a card of a draft:
//
//   - from the printing: types, power and toughness, keywords, and the
//     oracle text, for the three plain phrases the catalog does not
//     always declare (see textRemoval);
//   - from the catalog entry: the Purpose the card declares (ADR 0126),
//     on the card, on each mode and alternative cost, and on its
//     triggered and activated abilities, and whether it has a mana
//     ability.
//
// A card the index does not hold gets CardFacts. It never blocks: both
// lookups are in-memory maps.
func CatalogFacts(idx *cards.Index) FactsFunc {
	return func(c Card) Facts {
		f := CardFacts(c)
		if idx == nil {
			return f
		}
		full, ok := idx.Get(c.ScryfallID)
		if !ok {
			return f
		}
		text := full.OracleText
		power, toughness := full.Power, full.Toughness
		if len(full.CardFaces) > 0 {
			front := full.CardFaces[0]
			if text == "" {
				text = front.OracleText
			}
			if power == "" && toughness == "" {
				power, toughness = front.Power, front.Toughness
			}
		}
		if f.Creature && (power != "" || toughness != "") {
			f.Power, f.Toughness = statOf(power), statOf(toughness)
		}
		for _, k := range full.Keywords {
			addKeyword(&f, k)
		}
		text = strings.ToLower(text)
		if spec, ok := effects.Lookup(c.OracleID.String()); ok {
			for _, k := range spec.PrintedKeywords {
				addKeyword(&f, k)
			}
			readSpec(&f, spec, text)
		}
		readText(&f, text)
		return f
	}
}

func addKeyword(f *Facts, k string) {
	k = strings.ToLower(strings.TrimSpace(k))
	if k != "" && !slices.Contains(f.Keywords, k) {
		f.Keywords = append(f.Keywords, k)
	}
}

// statOf reads a printed power or toughness. "*" and "1+*" read as 0
// and 1: the printed number, without the part counted in play.
func statOf(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// readSpec fills in what the catalog entry declares.
func readSpec(f *Facts, spec effects.Spec, text string) {
	readPurpose(f, spec.Purpose, text)
	if spec.Purpose.Kicked != nil {
		readPurpose(f, *spec.Purpose.Kicked, text)
	}
	if spec.Modes != nil {
		for _, o := range spec.Modes.Options {
			readPurpose(f, o.Purpose, text)
		}
	}
	for _, a := range spec.AlternativeCosts {
		readPurpose(f, a.Purpose, text)
	}
	for _, t := range spec.Triggered {
		readPurpose(f, t.Purpose, text)
	}
	for _, a := range spec.Activated {
		readPurpose(f, a.Purpose, text)
	}
	if len(spec.ManaAbilities) > 0 && !f.Land {
		f.Ramp = max(f.Ramp, 1)
	}
}

// readPurpose folds one declared Purpose into the facts. Amounts keep
// the largest any one statement declares: a modal spell does one of its
// modes, not all of them.
func readPurpose(f *Facts, p game.Purpose, text string) {
	if !p.Sweep.IsZero() {
		f.Sweep = true
	}
	draws := p.Draws
	f.Ramp = max(f.Ramp, p.Lands)
	tokens := p.CreatureTokens
	if tokens == 0 && p.Answers&game.AnswerMakesBlocker != 0 {
		tokens = p.Tokens
	}
	f.Tokens = max(f.Tokens, tokens)
	if p.DamageToCreature > 0 {
		f.Damage = max(f.Damage, p.DamageToCreature)
	}
	for _, t := range p.Targets.List() {
		draws = max(draws, t.Draws)
		switch {
		case t.DamageIsX:
			f.Damage = max(f.Damage, DamageForX)
		case t.Damage > 0:
			f.Damage = max(f.Damage, t.Damage)
		}
		if !t.Returns.IsZero() {
			// Swords to Plowshares: a removal that gives something back
			// is still a removal.
			f.Removal = true
		}
		switch t.UntilEndOfTurn {
		case game.TargetBenefit:
			f.Trick = true
		case game.TargetHarm:
			if t.Pump != nil && t.Pump.Toughness < 0 {
				f.Damage = max(f.Damage, -t.Pump.Toughness)
			}
		}
	}
	f.Draws = max(f.Draws, draws)
	if p.Aura == game.AuraHarm {
		f.Removal = true
	}
	if p.Answers&game.AnswerRemove != 0 && p.Sweep.IsZero() && f.Damage == 0 {
		// AnswerRemove is declared by a kill spell and by a counterspell
		// alike (#4734). The printed text says which.
		if strings.Contains(text, "counter target") {
			f.Counterspell = true
		} else if !textNarrowRemoval.MatchString(text) || textRemoval.MatchString(text) {
			f.Removal = true
		}
	}
}

// textRemoval is the plain phrase of a creature removal spell: "destroy
// target creature", "exile target nonland permanent", "destroy target
// artifact or creature". The catalog declares a Purpose on the cards
// someone has curated; a card with none still prints this.
var textRemoval = regexp.MustCompile(`(destroy|exile) (another |up to one )?(other )?target [a-z ,\-]*(creature|permanent)`)

// textNarrowRemoval is removal that does not reach creatures: Naturalize.
var textNarrowRemoval = regexp.MustCompile(`(destroy|exile) target (artifact|enchantment|land)( or (artifact|enchantment|land))?[ .]`)

// textSteal is taking a creature for good: Mind Control. "Until end of
// turn" (Act of Treason) is not.
var textSteal = regexp.MustCompile(`you control enchanted creature|gain control of target creature[^.]*\.`)

// textDamage is a burn spell's "deals 3 damage" or "deals X damage".
var textDamage = regexp.MustCompile(`deals (\d+|x) damage`)

// readText fills in the things the printed text says plainly and the
// catalog entry may not declare: a kill spell, a burn spell, a
// counterspell and a plain draw. It only ever adds; a declared Purpose
// is not overruled.
func readText(f *Facts, text string) {
	if text == "" {
		return
	}
	if !f.Removal && !f.Sweep && textRemoval.MatchString(text) && !strings.Contains(text, "you control") {
		f.Removal = true
	}
	if m := textSteal.FindString(text); m != "" && !f.Removal && !strings.Contains(m, "until end of turn") {
		f.Removal = true
	}
	if m := textDamage.FindStringSubmatch(text); m != nil && f.Damage == 0 && !f.Sweep &&
		strings.Contains(text, "target") && !strings.Contains(text, "damage to each") {
		if m[1] == "x" {
			f.Damage = DamageForX
		} else {
			f.Damage = statOf(m[1])
		}
	}
	if !f.Counterspell && strings.Contains(text, "counter target") && (f.Instant || f.Creature) {
		f.Counterspell = true
	}
	if f.Draws == 0 {
		switch {
		case strings.Contains(text, "draw three cards"):
			f.Draws = 3
		case strings.Contains(text, "draw two cards"):
			f.Draws = 2
		case strings.Contains(text, "draw a card"):
			f.Draws = 1
		}
	}
}

// cardFacts is the FactsFunc the package's own entry points use when
// their caller gives none: BuildDeck, and the Picker a draft is made
// with by default. The server sets it once at start (UseCards).
var cardFacts atomic.Pointer[FactsFunc]

// UseCards tells the draft bots where card facts come from: the server's
// card index, read with the catalog (CatalogFacts). The server calls it
// once at start. Before it is called, and in a test that never calls it,
// the bots rate a card from the draft Card's own fields (CardFacts):
// weaker, never wrong.
func UseCards(idx *cards.Index) {
	if idx == nil {
		cardFacts.Store(nil)
		return
	}
	f := CatalogFacts(idx)
	cardFacts.Store(&f)
}

// defaultRater is the Rater over whatever UseCards and UseRatings last
// set.
func defaultRater() Rater {
	r := Rater{}
	if f := cardFacts.Load(); f != nil {
		r.Facts = *f
	}
	if b := ratingBook.Load(); b != nil {
		r.Overrides = BookOverrides(b)
	}
	return r
}
