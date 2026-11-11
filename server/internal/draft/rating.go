package draft

import (
	"slices"
	"strings"
)

// rating.go is the draft bot's card rating (ADR 0222 §7, #5439): one
// number for how much a bot wants a card in a 40-card deck, computed
// from what the server already knows about it. No outside draft data.
//
// The rating is a pure function of Facts, so it can be tabulated and
// tested, and every weight it uses is a named constant in this file.
// Where the facts come from is facts.go: the printing's own fields and
// what the catalog declares the card does (ADR 0126's Purpose).
//
// The idea is the heuristic game bot's (aiseat/heuristic): a body is
// worth its power and toughness, a spell is worth what it declares it
// does, and both pay for their mana. The numbers are this file's own;
// the two never share a weight, because one prices a move on a board and
// the other a card in a pack.

// Facts is what the rating reads about a card. Every field is printed on
// the card or declared by the catalog; none is a judgement.
type Facts struct {
	ManaValue int
	// The card types on the front face's type line.
	Land, Creature, Instant, Sorcery    bool
	Artifact, Enchantment, Planeswalker bool
	// Aura and Equipment are the two subtypes that need a creature to
	// sit on.
	Aura, Equipment bool
	// Power and Toughness are the printed numbers; a "*" reads as 0.
	Power, Toughness int
	// Keywords are the printed keyword abilities, lowercase.
	Keywords []string
	// Rarity is Scryfall's spelling: common, uncommon, rare, mythic.
	Rarity string
	// Colors is the card's colours (W, U, B, R, G), empty for a
	// colourless card.
	Colors []string

	// What the catalog declares the card does (ADR 0126). Printed
	// amounts, as there. A creature's are what it does on entering or
	// from the battlefield.

	// Removal: it destroys, exiles or neutralises a creature or another
	// permanent an opponent controls.
	Removal bool
	// Damage is the damage it deals to a target. An X spell reads as
	// DamageForX.
	Damage int
	// Sweep: it is a board wipe.
	Sweep bool
	// Counterspell: it counters a spell.
	Counterspell bool
	// Draws is the cards it draws its controller.
	Draws int
	// Ramp is the lands it puts onto the battlefield, plus one for a
	// nonland permanent with a mana ability.
	Ramp int
	// Tokens is the creature tokens it makes.
	Tokens int
	// Trick: it pumps or protects a creature until end of turn.
	Trick bool
}

// The weights of the rating, all of them. A rating is in "points": a
// two-mana 2/2 is about RatingFiller, a card a deck is glad to play is
// above RatingPlayable, a premium removal spell is about 4 and a bomb is
// above 5.
const (
	// --- creatures ---

	// CreatureBase is what a creature is worth for being a body at all.
	CreatureBase = 2.0
	// PerStatPoint is the worth of one point of power or toughness.
	PerStatPoint = 0.45
	// PerManaCreature is what each point of mana value costs a
	// creature: a 3/3 for four is worse than a 2/2 for two.
	PerManaCreature = 0.6
	// DefenderPenalty is taken from a creature that cannot attack, and
	// NoPowerPenalty from one that deals no damage when it does.
	DefenderPenalty = 1.0
	NoPowerPenalty  = 1.0
	// FullKeywordPower is the power at which a combat keyword is worth
	// its whole weight: flying on a 0/2 is worth nothing, on a 1/1 half.
	FullKeywordPower = 2

	// --- keywords, added to a creature ---

	KeywordFlying         = 0.9
	KeywordDoubleStrike   = 1.0
	KeywordIndestructible = 0.8
	KeywordDeathtouch     = 0.6
	KeywordLifelink       = 0.5
	KeywordFirstStrike    = 0.4
	KeywordMenace         = 0.4
	KeywordHexproof       = 0.4
	KeywordTrample        = 0.3
	KeywordHaste          = 0.3
	KeywordVigilance      = 0.2
	KeywordReach          = 0.2
	KeywordFlash          = 0.2
	KeywordWard           = 0.3

	// --- what a card declares it does ---

	// SpellBase is a noncreature spell that declares nothing the rating
	// reads: a card the bot knows only the cost of. A spell that draws,
	// ramps or makes tokens is worth this and those.
	SpellBase = 1.5
	// PerManaSpell is what each point of mana value costs a spell whose
	// effect is priced below.
	PerManaSpell = 0.25
	// RemovalValue is a spell that removes a creature, before its cost.
	RemovalValue = 4.8
	// DamageBase and PerDamage price a burn spell: a little for being
	// one, and more for each point it deals to its target.
	DamageBase = 2.2
	PerDamage  = 0.6
	// DamageForX is what an X damage spell is read as dealing.
	DamageForX = 3
	// SweepValue is a board wipe.
	SweepValue = 4.4
	// CounterspellValue is a counterspell: an answer, but one that has
	// to be held.
	CounterspellValue = 2.6
	// PerDraw is one card drawn.
	PerDraw = 0.8
	// PerRamp is one land put onto the battlefield, or a mana ability.
	PerRamp = 0.9
	// PerToken is one creature token.
	PerToken = 0.8
	// TrickValue is a combat trick.
	TrickValue = 1.9
	// AuraPenalty is taken from an Aura or Equipment that is not
	// removal: it is a card that needs a creature first.
	AuraPenalty = 0.4
	// OnABody scales what a creature declares it does: the effect comes
	// with a body that is already priced.
	OnABody = 0.75
	// PlaneswalkerValue is a planeswalker, before its cost.
	PlaneswalkerValue = 5.5

	// --- rarity ---

	// The bonus a rarer card gets. A set's rares are, on the whole,
	// stronger than their printed numbers say, and the rating has no
	// other way to know it.
	RarityUncommon = 0.3
	RarityRare     = 0.9
	RarityMythic   = 1.4

	// --- lands ---

	// BasicLandRating is a basic land: free in deckbuilding, so worth
	// nothing in a pack.
	BasicLandRating = 0.0
	// NonbasicLandRating is any other land.
	NonbasicLandRating = 1.0

	// --- the scale ---

	// RatingFloor is the least a nonland card rates.
	RatingFloor = 0.2
	// RatingFiller is about a two-mana 2/2.
	RatingFiller = 2.7
	// RatingPlayable is the least rating BuildDeck counts as a card a
	// colour is glad to have, when it chooses the deck's colours.
	RatingPlayable = 2.2
)

// keywordWeights is the keyword table Rate reads.
var keywordWeights = map[string]float64{
	"flying":         KeywordFlying,
	"double strike":  KeywordDoubleStrike,
	"indestructible": KeywordIndestructible,
	"deathtouch":     KeywordDeathtouch,
	"lifelink":       KeywordLifelink,
	"first strike":   KeywordFirstStrike,
	"menace":         KeywordMenace,
	"hexproof":       KeywordHexproof,
	"trample":        KeywordTrample,
	"haste":          KeywordHaste,
	"vigilance":      KeywordVigilance,
	"reach":          KeywordReach,
	"flash":          KeywordFlash,
	"ward":           KeywordWard,
}

// Rate is the rating of a card with these facts. It is a pure function:
// the same facts always give the same number.
func Rate(f Facts) float64 {
	if f.Land {
		if !f.Creature {
			// A basic land is told apart by the caller (Card.Basic); a
			// land's facts alone say only that it is one.
			return NonbasicLandRating
		}
	}
	var r float64
	switch {
	case f.Creature:
		r = CreatureBase + PerStatPoint*float64(max(f.Power, 0)+max(f.Toughness, 0)) -
			PerManaCreature*float64(f.ManaValue)
		scale := float64(min(max(f.Power, 0), FullKeywordPower)) / FullKeywordPower
		for _, k := range f.Keywords {
			if combatKeywords[k] {
				r += scale * keywordWeights[k]
			} else {
				r += keywordWeights[k]
			}
		}
		if slices.Contains(f.Keywords, "defender") {
			r -= DefenderPenalty
		} else if f.Power <= 0 {
			r -= NoPowerPenalty
		}
		r += OnABody * effectValue(f)
	case f.Planeswalker:
		r = PlaneswalkerValue - PerManaSpell*float64(f.ManaValue)
	default:
		r = SpellBase
		if v := effectValue(f); v > 0 {
			r = v - PerManaSpell*float64(f.ManaValue)
		}
		if (f.Aura || f.Equipment) && !f.Removal {
			r -= AuraPenalty
		}
	}
	r += rarityBonus(f.Rarity)
	return max(r, RatingFloor)
}

// combatKeywords are the keywords that only matter in proportion to the
// damage the creature deals.
var combatKeywords = map[string]bool{
	"flying": true, "double strike": true, "lifelink": true, "first strike": true,
	"menace": true, "trample": true, "haste": true, "vigilance": true,
}

// effectValue is what the card's declared effect is worth before its
// mana: the best single thing it does, plus the smaller things beside
// it. Zero for a card that declares nothing the rating reads.
//
// On a creature it is only the effect; on a spell it is the whole card,
// so an effect that is not an answer (a draw, a ramp, tokens) stands on
// SpellBase.
func effectValue(f Facts) float64 {
	best := 0.0
	if f.Removal {
		best = max(best, RemovalValue)
	}
	if f.Damage > 0 {
		best = max(best, DamageBase+PerDamage*float64(f.Damage))
	}
	if f.Sweep {
		best = max(best, SweepValue)
	}
	if f.Counterspell {
		best = max(best, CounterspellValue)
	}
	if f.Trick {
		best = max(best, TrickValue)
	}
	extra := PerDraw*float64(f.Draws) + PerRamp*float64(f.Ramp) + PerToken*float64(f.Tokens)
	if best == 0 {
		if extra == 0 || f.Creature {
			return extra
		}
		return SpellBase + extra
	}
	// A removal spell that also draws a card is better than one that
	// does not, but not by a whole Divination.
	return best + extra/2
}

func rarityBonus(rarity string) float64 {
	switch strings.ToLower(rarity) {
	case "uncommon":
		return RarityUncommon
	case "rare":
		return RarityRare
	case "mythic":
		return RarityMythic
	}
	return 0
}

// FactsFunc answers what is known about a card. It must not block: the
// picker calls it with the draft's lock held.
type FactsFunc func(Card) Facts

// RatingOverrides is the seam for a per-set ratings file (ADR 0222 §7):
// a rating for a card, by set and oracle ID, that replaces the computed
// one. It is empty today. No such file ships, and whether outside draft
// data may fill one is an owner and licence question the ADR leaves
// open (#5439's comment on 17Lands).
type RatingOverrides interface {
	// Rating is the override for a card of a set, and false when the set
	// has none for it.
	Rating(set string, c Card) (float64, bool)
}

// NoOverrides is the RatingOverrides every draft has today: none.
type NoOverrides struct{}

// Rating answers that there is no override.
func (NoOverrides) Rating(string, Card) (float64, bool) { return 0, false }

// Rater rates the cards of a draft: an override where one exists, else
// Rate over the card's facts. The zero Rater rates from the card's own
// fields alone (CardFacts).
type Rater struct {
	// Facts is where a card's facts come from. Nil is CardFacts.
	Facts FactsFunc
	// Overrides is the per-set ratings seam. Nil is NoOverrides.
	Overrides RatingOverrides
}

// FactsOf is the facts the rater reads for a card.
func (r Rater) FactsOf(c Card) Facts {
	if r.Facts != nil {
		return r.Facts(c)
	}
	return CardFacts(c)
}

// Rating is the rating of a card.
func (r Rater) Rating(c Card) float64 {
	if c.Basic {
		return BasicLandRating
	}
	if r.Overrides != nil {
		if v, ok := r.Overrides.Rating(c.Set, c); ok {
			return v
		}
	}
	return Rate(r.FactsOf(c))
}

// CardFacts is the facts a draft Card carries by itself: mana value,
// types, rarity and colours. It is what a rater falls back to when it
// has no index to ask, and the base CatalogFacts fills in.
func CardFacts(c Card) Facts {
	f := Facts{
		ManaValue: c.ManaValue,
		Rarity:    strings.ToLower(c.Rarity),
		Colors:    append([]string(nil), c.Colors...),
	}
	setTypes(&f, c.TypeLine)
	if f.Creature {
		// With no printed stats to read, a creature is taken to be the
		// size its cost usually buys.
		f.Power, f.Toughness = max(c.ManaValue, 1), max(c.ManaValue, 1)
	}
	return f
}

// setTypes reads the card types and the two subtypes the rating knows
// from a type line. A two-faced card's line is its front face's.
func setTypes(f *Facts, typeLine string) {
	front, _, _ := strings.Cut(typeLine, " // ")
	types, subtypes, _ := strings.Cut(front, "—")
	for _, w := range strings.Fields(types) {
		switch w {
		case "Land":
			f.Land = true
		case "Creature":
			f.Creature = true
		case "Instant":
			f.Instant = true
		case "Sorcery":
			f.Sorcery = true
		case "Artifact":
			f.Artifact = true
		case "Enchantment":
			f.Enchantment = true
		case "Planeswalker":
			f.Planeswalker = true
		}
	}
	for _, w := range strings.Fields(subtypes) {
		switch w {
		case "Aura":
			f.Aura = true
		case "Equipment":
			f.Equipment = true
		}
	}
}
