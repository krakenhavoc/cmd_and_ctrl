package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temporal Anomaly — Land (Star Trek, #4973):
//
//	"{T}: Add {C}.
//	 {3}{G}, {T}, Sacrifice this land: Mill four cards. You may put a
//	 permanent card from among them into your hand. Create a tapped Planet
//	 land token named New Planet with "{T}: Add one mana of any color.""
//
// The activation is Something Worth Saving's mill-then-pick: the pick is
// over the cards that actually reached the graveyard (CR 400.7), and the
// New Planet is made once the pick is answered, in printed order, whether
// or not a card was taken. The New Planet is Genesis Effect's token, as
// printed (ttrk 27), entering tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "cf0d1859-699e-47d4-bc91-7c5ff61fabe9",
		Name:          "Temporal Anomaly",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.TapForColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}, {T}, Sacrifice this land: Mill four cards. You may put a permanent card from among them into your hand. Create a tapped Planet land token named New Planet.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    effects.Plus(effects.ManaCost("{3}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RfSpellBMillThenMayTakePermanent(effects.NewContext(g, item),
					"Temporal Anomaly — you may put a permanent card milled this way into your hand",
					temporalAnomalyMilled)
			},
		}},
	})
}

// temporalAnomalyMilled is the mill's keyed "then": the pick, then the
// tapped New Planet.
var temporalAnomalyMilled = effects.RfSpellBTakeFromMilledTail("mill-then-take/temporal-anomaly", func(c *effects.Context) error {
	return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TrekB16NewPlanetToken()).EntersTapped(), N: 1}.Apply(c)
})
