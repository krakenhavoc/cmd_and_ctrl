package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sisay, Weatherlight Captain — Legendary Creature — Human Soldier {2}{W},
// 2/2 (Dominaria United Commander reprint, #5908):
//
//	"Sisay gets +1/+1 for each color among other legendary permanents you
//	 control.
//	 {W}{U}{B}{R}{G}: Search your library for a legendary permanent card
//	 with mana value less than Sisay's power, put that card onto the
//	 battlefield, then shuffle."
//
// The pump is a layer-7c modification counted live off the board, one point
// per distinct colour among the legendary permanents other than Sisay
// (CR 105.1), so a multicoloured legend counts each of its colours once
// and a colourless one counts none. The search reads Sisay's power as the
// ability resolves, from her last-known information if she has gone
// (CR 608.2h): a legend must have a mana value strictly less than it.
// Finding nothing is allowed (CR 701.23b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb777610-7562-4d9d-8497-62c562e6d7cb",
		Name:         "Sisay, Weatherlight Captain",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Gets +1/+1 for each color among other legendary permanents you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := sisayColors(g, source)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}: Search your library for a legendary permanent card with mana value less than Sisay's power, put that card onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    ManaCost("{W}{U}{B}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				power := 0
				if info, ok := ctx.SourcePermanent(); ok {
					power = info.Power
				}
				return SearchLibrary{
					Player: ctx.Controller(),
					Predicate: func(c game.Card) bool {
						return c.IsLegendary() && c.IsPermanent() && c.ManaValue() < power
					},
					Dest:    game.ZoneBattlefield,
					Limit:   1,
					Shuffle: true,
					Source:  item.SourceCardID,
					Reason:  "Sisay, Weatherlight Captain — search your library for a legendary permanent card with mana value less than her power",
				}.Apply(ctx)
			},
		}},
	})
}

// sisayColors is the number of distinct colours among the other legendary
// permanents Sisay's controller controls.
func sisayColors(g *game.Game, sisay *game.Card) int {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != sisay.Controller || c.InstanceID == sisay.InstanceID || !c.IsLegendary() {
			continue
		}
		for _, col := range c.EffectiveColors() {
			seen[strings.ToUpper(col)] = true
		}
	}
	return len(seen)
}
