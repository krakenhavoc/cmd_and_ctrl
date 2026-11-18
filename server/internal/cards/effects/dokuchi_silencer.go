package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dokuchi Silencer — Creature — Human Ninja {1}{B}, 2/1 (Neon Dynasty):
//
//	"Ninjutsu {1}{B}
//	 Whenever this creature deals combat damage to a player, you may discard
//	 a creature card. When you do, destroy target creature or planeswalker
//	 that player controls."
//
// The "you may discard a creature card" is the discard prompt with a floor of
// zero and a set rule that accepts only a lone creature card; the reflexive
// trigger (CR 603.12) is created only once a card was really discarded. The
// damaged player rides the trigger's Carry and the reflexive params, and the
// target is chosen as the trigger goes on the stack from what that player
// controls (re-checked as it resolves, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d377b28f-c887-4937-9f14-74484b612551",
		Name:         "Dokuchi Silencer",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Dokuchi Silencer — you may discard a creature card; when you do, destroy target creature or planeswalker that player controls",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						UpTo:     true,
						Question: "Dokuchi Silencer — you may discard a creature card to destroy a creature or planeswalker that player controls",
					}, dokuchiSilencerDiscarded, []uuid.UUID{item.Trigger.Event.Target})
				}),
		},
	})
}

var dokuchiSilencerDiscarded = game.RegisterDiscardThenRule("choose-cards/dokuchi-silencer-destroy",
	func(picked []game.Card) bool { return len(picked) == 1 && picked[0].IsCreature() },
	func(g *game.Game, r game.DiscardsDone) error {
		if len(r.Discarded.Cards()) == 0 || len(r.Carry) == 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Dokuchi Silencer — destroy target creature or planeswalker that player controls",
			Body:   dokuchiSilencerDestroyBody,
			Params: game.EffectParams{Player: r.Carry[0]},
		}.Apply(keyedContext(g, r.Player, r.Source))
	})

var dokuchiSilencerDestroyBody = game.ReflexiveBody("s82-neo-b06/dokuchi-silencer-destroy",
	simpleBody(destroyFirstLegalCardTarget),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetPermanent("target creature or planeswalker that player controls", Or(Creature(), Planeswalker()), ControlledBy(p.Player))
	})
