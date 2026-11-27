package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liar's Pendulum — Artifact {1} (Mirrodin, S82 OLD, #5606):
//
//	"{2}, {T}: Choose a card name. Target opponent guesses whether a card
//	 with that name is in your hand. You may reveal your hand. If you do
//	 and your opponent guessed wrong, draw a card."
//
// Three answers in turn, each a keyed prompt so a table waiting on one is a
// restore point: the controller names a card (not validated against any
// list), the target opponent answers yes or no to "is a card with that name
// in their hand?" (a target that has left makes the ability do nothing,
// CR 608.2b), then the controller is asked whether to reveal their hand. The reveal happens on a yes whatever the guess was, and only a
// wrong guess draws. Whether a card is named is read once, when the name is
// given, from the hand as it stood: nothing can change it before the
// answers (CR 608.2c). The guess is the pile-or-guess tally's (#6175, CR
// 700.3), noted once as the guess is asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f88ddfa-5478-426a-a978-dfc64c687db1",
		Name:         "Liar's Pendulum",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Choose a card name. Target opponent guesses whether a card with that name is in your hand. You may reveal your hand. If you do and your opponent guessed wrong, draw a card.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					g.QueueCardNameChoiceKeyedForEffect(item.Controller, item.SourceCardID,
						"Liar's Pendulum — choose a card name",
						effects.WithCarry(ctx.ItemTail(liarsPendulumNamed), t.ID))
					return nil
				}
				return nil
			},
		}},
	})
}

// liarsPendulumNamed asks the target opponent to guess once the name (the
// tail's Word) is given. The carry is the opponent, and the second ID the
// controller's own when a card with that name is in their hand, Nil when
// not. An on-disk key.
var liarsPendulumNamed = effects.CardTail("then/s82-old-b602-liars-pendulum-named", func(ctx *effects.Context, o game.TailOutcome) error {
	opp, name := effects.TailCarry0(o), o.Tail.Word
	inHand := uuid.Nil
	if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			if game.CardNameMatches(c, name) {
				inHand = ctx.Controller()
				break
			}
		}
	}
	if opp == uuid.Nil || ctx.Game.PlayerByIDForEffect(opp) == nil {
		return liarsPendulumAskReveal(ctx, false)
	}
	effects.NotePileOrGuess(ctx)
	return effects.MayChoice{
		Player:   opp,
		Question: "Liar's Pendulum — is a card named " + name + " in their hand?",
		YesLabel: "Yes, it is in their hand",
		NoLabel:  "No, it is not in their hand",
		Then:     liarsPendulumGuessed,
		Carry:    []uuid.UUID{inHand},
	}.Apply(ctx)
})

// liarsPendulumGuessed judges the guess (Accepted is "it is in their
// hand"), then asks the controller about the reveal. An on-disk key.
var liarsPendulumGuessed = effects.MayAnswer("confirm/s82-old-b602-liars-pendulum-guess", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	inHand := uuid.Nil
	if len(r.Carry) > 0 {
		inHand = r.Carry[0]
	}
	wrong := r.Accepted != (inHand != uuid.Nil)
	return liarsPendulumAskReveal(ctx, wrong)
})

// liarsPendulumAskReveal asks "you may reveal your hand"; the answer's word
// says whether the guess was wrong.
func liarsPendulumAskReveal(ctx *effects.Context, wrong bool) error {
	word := "right"
	if wrong {
		word = "wrong"
	}
	return effects.MayChoice{
		Question: "Liar's Pendulum — reveal your hand?",
		YesLabel: "Reveal your hand",
		NoLabel:  "Don't reveal",
		Then:     liarsPendulumReveal,
		Word:     word,
	}.Apply(ctx)
}

// liarsPendulumReveal reveals the hand on a yes and draws if the guess was
// wrong. An on-disk key.
var liarsPendulumReveal = effects.MayYes("confirm/s82-old-b602-liars-pendulum-reveal", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	ctx.Game.RevealHandForEffect(ctx.Controller())
	if r.Word != "wrong" {
		return nil
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
})
