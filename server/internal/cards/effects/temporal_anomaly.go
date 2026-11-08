package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "cf0d1859-699e-47d4-bc91-7c5ff61fabe9",
		Name:          "Temporal Anomaly",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{tapForColorless()},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, {T}, Sacrifice this land: Mill four cards. You may put a permanent card from among them into your hand. Create a tapped Planet land token named New Planet.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
			Cost:    Plus(ManaCost("{3}{G}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return rfSpellBMillThenMayTakePermanent(NewContext(g, item),
					"Temporal Anomaly — you may put a permanent card milled this way into your hand",
					temporalAnomalyMilled)
			},
		}},
	})
}

// temporalAnomalyMilled is the mill's keyed "then": the pick, then the
// tapped New Planet.
var temporalAnomalyMilled = rfSpellBTakeFromMilledTail("mill-then-take/temporal-anomaly", func(c *Context) error {
	return CreateTokenAdvanced{Spec: Token(trekB16NewPlanetToken()).EntersTapped(), N: 1}.Apply(c)
})
