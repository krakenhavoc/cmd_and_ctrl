package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Domain (#5666). "Domain" is an ability word (CR 207.2c): it has no
// rules meaning of its own, and the count behind it is "the number of
// basic land types among lands you control" (CR 305.6 names the five
// basic land types). Each type is counted once however many lands carry
// it, read from the layered land so a land that gained a type counts (a
// Dryad of the Ilysian Grove adds all five) and one that lost its types
// to a Blood Moon does not. Five at most.
//
// One count, three shapes:
//
//   - a cost reduction on a spell or on an activated ability:
//     DomainCostReduction (Leyline Binding, Yavimaya Sojourner, Jodah's
//     Codex);
//   - a characteristic-defining power and toughness: DomainPT
//     (Territorial Maro, Nishoba Brawler, Fblthp, Knows the Way);
//   - a count read as an ability resolves: DomainCount (Sunbathing
//     Rootwalla).

// DomainCount is the number of basic land types among the lands
// `controller` controls, now.
func DomainCount(g *game.Game, controller uuid.UUID) int {
	if g == nil {
		return 0
	}
	lands := g.BattlefieldCardsForEffect()
	n := 0
	for _, t := range game.BasicLandTypes {
		for _, c := range lands {
			if c.Controller == controller && c.IsLand() && c.HasSubtype(t) {
				n++
				break
			}
		}
	}
	return n
}

// DomainCostAmount is DomainCount for a CostQuery: "you" is the player
// casting the spell or activating the ability.
func DomainCostAmount(q game.CostQuery) int {
	return DomainCount(q.Game, q.Controller)
}

// DomainCostReduction is "This spell costs {1} less to cast for each
// basic land type among lands you control" and its activated-ability
// twin, as a cost modifier. It reduces generic mana only (CR 601.2f).
func DomainCostReduction(label string) game.CostModifier {
	return CostsLessEach(DomainCostAmount, label)
}

// DomainPT is a characteristic-defining power and toughness (CR 604.3,
// layer 7a) of `powerPer` and `toughnessPer` times the domain count of
// the source's controller; a zero factor leaves that stat alone. The
// label is the card's printed line.
func DomainPT(label string, powerPer, toughnessPer int) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := DomainCount(g, source.Controller)
			if powerPer != 0 {
				c.Power = powerPer * n
			}
			if toughnessPer != 0 {
				c.Toughness = toughnessPer * n
			}
		},
	}
}
