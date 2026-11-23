package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b06_helpers.go — the shared shapes of Neon Dynasty's ninja batch
// (S82, #5493).

// s82NeoB06RemoveKeywordCounterThen is the parent half of "Whenever this
// creature deals combat damage to a player, you may remove a <kind> counter
// from it. When you do, <body>" (Biting-Palm Ninja, Kappa Tech-Wrecker). The
// "you may" is the trigger's optional prompt; removing the counter is the
// doing, and the reflexive trigger (CR 603.12) exists only if a counter was
// really removed, so a source that left or lost its counter gives nothing.
// The damaged player rides the reflexive trigger's params, for "that
// player".
func S82NeoB06RemoveKeywordCounterThen(kind, label string, body game.BodyRef) func(*game.Game, *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		ctx := NewContext(g, item)
		src, ok := ctx.SourcePermanent()
		if !ok || src.Left || src.Counters[kind] <= 0 {
			return nil
		}
		if err := (AddCounter{Target: item.SourceCardID, Kind: kind, N: -1}).Apply(ctx); err != nil {
			return err
		}
		return ReflexiveTrigger{
			Label:  label,
			Body:   body,
			Params: game.EffectParams{Player: item.Trigger.Event.Target},
		}.Apply(ctx)
	}
}

// s82NeoB06ThatPlayerRevealsHandExile is Biting-Palm Ninja's reflexive body:
// "that player reveals their hand and you choose a nonland card from it.
// Exile that card."
var S82NeoB06RevealHandExileBody = game.ReflexiveBody("s82-neo-b06/biting-palm-ninja-reveal-exile",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Player == uuid.Nil {
			return nil
		}
		return ChooseFromRevealedHand{
			Player: p.Player,
			Filter: Nonland(),
			Label:  "nonland card",
			Exile:  true,
		}.Apply(NewContext(g, item))
	}, nil)
