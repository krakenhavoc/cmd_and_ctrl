package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cait Sith, Fortune Teller — Legendary Artifact Creature — Cat Moogle
// {3}{R}, 3/3 (#3656):
//
//	"Lucky Slots — At the beginning of combat on your turn, scry 1, then
//	 exile the top card of your library. You may play that card this
//	 turn. When you exile a card this way, target creature you control
//	 gets +X/+0 until end of turn, where X is that card's mana value."
//
// Scry's own continuation runs the exile (the card on top is the one
// the player settled on), which grants the this-turn play permission
// (Reckless Impulse's impulse exile) and then creates the reflexive
// trigger (CR 603.12) carrying the exiled card's mana value. The
// trigger is targeted, so with no creature to choose it is removed, and
// an empty library exiles nothing and creates none. A land has mana
// value zero, so its trigger adds nothing.
//
// No simplification.
func init() {
	const label = "Cait Sith, Fortune Teller — Lucky Slots: scry 1, then exile the top card of your library; you may play it this turn"
	Register(Spec{
		OracleID:     "31851cb6-0caf-4575-aa5e-41aeadccafc1",
		Name:         "Cait Sith, Fortune Teller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return Scry{N: 1, ThenTail: ctx.ItemTail(caitSithImpulseExile)}.Apply(ctx)
			}),
		},
	})
}

// caitSithImpulseExile is "then exile the top card of your library; you may
// play it this turn", and the reflexive boost it earns. An on-disk key.
var caitSithImpulseExile = CardTail("cait-sith-fortune-teller/impulse-exile", func(ctx *Context, _ game.TailOutcome) error {
	item := ctx.Item
	return b12ImpulseExileForTurnThen(ctx.Game, item, 1, func(g *game.Game, exiled []uuid.UUID) error {
		if len(exiled) == 0 {
			return nil
		}
		card, ok := g.LookupCardForEffect(exiled[0])
		if !ok {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Cait Sith, Fortune Teller — target creature you control gets +X/+0 until end of turn",
			Cards:  exiled,
			Body:   caitSithBoostBody,
			Params: game.EffectParams{Amount: card.ManaValue()},
		}.Apply(NewContext(g, item))
	})
})

// caitSithBoost is the reflexive half: the chosen creature gets
// +X/+0, X the exiled card's mana value (Params.Amount).
func caitSithBoost(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok || p.Amount <= 0 {
		return nil
	}
	return BoostUntilEOT{Target: target, Power: p.Amount, Label: "Cait Sith, Fortune Teller — +X/+0"}.Apply(ctx)
}
