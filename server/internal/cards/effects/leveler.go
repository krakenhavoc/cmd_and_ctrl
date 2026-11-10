package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// leveler.go — leveler cards (CR 702.87, CR 711; #5426).
//
// A leveler card prints one activated ability and two level symbols:
//
//	CR 702.87a: "Level up [cost]" means "[Cost]: Put a level counter on
//	this permanent. Activate only as a sorcery."
//	CR 711.2a: "{LEVEL N1-N2} [Abilities] [P/T]" means "As long as this
//	creature has at least N1 level counters on it, but no more than N2
//	level counters on it, it has base power and toughness [P/T] and has
//	[abilities]."
//	CR 711.2b: "{LEVEL N3+} [Abilities] [P/T]" means "As long as this
//	creature has N3 or more level counters on it, it has base power and
//	toughness [P/T] and has [abilities]."
//
// Nothing here is a new engine rule. A level symbol "is a keyword
// ability that represents a static ability" (CR 711.2), so a band is the
// ordinary statics a card with that sentence printed in full would
// carry (Living Conundrum's "as long as …, this creature has base power
// and toughness 10/10 and has flying"): a layer 7b set (CR 613.4b), a
// layer 6 grant per keyword, and the band's other abilities gated on the
// same count. Everything else follows from their being statics of the
// permanent:
//
//   - below N1 no band applies and the creature has its printed
//     (uppermost) power and toughness (CR 711.5), which is also what the
//     card has in every other zone (CR 711.6);
//   - a creature that loses its abilities loses its level symbols with
//     them, and is its printed size however many counters it has;
//   - the set is base power and toughness, so counters and pumps apply
//     on top of it, and a later effect that sets base power and
//     toughness wins by timestamp (CR 613.7);
//   - a copy of the creature has the same symbols and reads its own
//     level counters.
//
// The level counter is game.CounterLevel. Class levels are a
// designation, not counters, and do not meet this file (CR 702.87c).

// LevelUpKeyword is "Level up [cost]" (CR 702.87a). It is one of the
// permanent's ordinary abilities, so it can be activated however many
// level counters the permanent has (CR 711.4). It is not
// designations.go's LevelUp, which is a Class's level ability (CR 716).
func LevelUpKeyword(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:        "Level up " + cost,
		Cost:         ManaCost(cost),
		SorcerySpeed: true,
		Effect:       Do(CounterOnThis{Kind: game.CounterLevel, N: 1}),
	}
}

// LevelBand is one level symbol with its striation (CR 711.2): the
// range of level counters, the power/toughness box, and the abilities
// printed beside them.
type LevelBand struct {
	// From and To are N1 and N2 of "LEVEL N1-N2". To == 0 is the open
	// band "LEVEL N3+", with From as N3.
	From, To int
	// Power and Toughness are the band's power/toughness box.
	Power, Toughness int
	// Keywords are the band's keyword abilities, lower case.
	Keywords []string
	// Statics are the band's other static abilities, written as they
	// would be for a card that always had them. LevelBands gates each
	// one on the source's level counters.
	Statics []game.StaticAbility
}

// holds reports whether a permanent with n level counters is in the
// band.
func (b LevelBand) holds(n int) bool {
	return n >= b.From && (b.To == 0 || n <= b.To)
}

// symbol is the level symbol as a player reads it: "Level 2-3",
// "Level 4+".
func (b LevelBand) symbol() string {
	if b.To == 0 {
		return fmt.Sprintf("Level %d+", b.From)
	}
	return fmt.Sprintf("Level %d-%d", b.From, b.To)
}

// LevelBands returns the static abilities of a leveler card's level
// symbols, for Spec.Static.
func LevelBands(bands ...LevelBand) []game.StaticAbility {
	var out []game.StaticAbility
	for _, b := range bands {
		b := b
		inBand := func(source *game.Card) bool {
			return source != nil && b.holds(source.Counters[game.CounterLevel])
		}
		self := func(target *game.Card, g *game.Game, source *game.Card) bool {
			return selfOnly(target, g, source) && inBand(source)
		}
		out = append(out, game.StaticAbility{
			Label:     fmt.Sprintf("%s: this creature has base power and toughness %d/%d.", b.symbol(), b.Power, b.Toughness),
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: self,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power, c.Toughness = b.Power, b.Toughness
			},
		})
		for _, kw := range b.Keywords {
			out = append(out, Labelled(KeywordGrant(self, kw), fmt.Sprintf("%s: this creature has %s.", b.symbol(), kw)))
		}
		for _, s := range b.Statics {
			inner := s.AppliesTo
			s.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
				return inBand(source) && inner(target, g, source)
			}
			if s.Label != "" {
				s.Label = b.symbol() + ": " + s.Label
			}
			out = append(out, s)
		}
	}
	return out
}
