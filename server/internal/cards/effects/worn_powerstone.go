package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Worn Powerstone — Artifact {3}:
//
//	"Worn Powerstone enters tapped. {T}: Add {C}{C}."
//
// The enters-tapped clause is the drawback that makes a two-mana
// rock cost three, so it is implemented rather than deferred: OnETB
// taps the permanent as it lands. That is a beat later than the real
// replacement effect (it enters untapped and is then tapped, rather
// than entering tapped), which is observable only by something
// watching for an untapped-to-tapped transition on ETB. Nothing in
// the catalog does, and a true enters-tapped needs a CR 614
// replacement on a self-referential event the engine builds after
// the card is already placed.
func init() {
	Register(Spec{
		OracleID: "b166b670-febc-4821-855e-f8d465644c03",
		Name:     "Worn Powerstone",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		OnETB: func(card *game.Card, ctx *Context) error {
			return TapTarget{Target: card.InstanceID}.Apply(ctx)
		},
	})
}
