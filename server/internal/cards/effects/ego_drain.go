package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ego Drain — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card. If you don't control a Faerie,
//	 exile a card from your hand."
//
// Thoughtseize's pick (ADR 0116), then the drawback. "A Faerie" is any
// permanent with the subtype, a Kindred one included, read as Ego Drain
// resolves (ControlsA, Snuff Out's Swamp check). Without one you choose
// a card from your own hand and exile it; with an empty hand there is
// nothing to exile (CR 609.3).
//
// The exile is printed after the discard and is queued on the next
// line, after the pick (ADR 0116 §6). It cannot change the pick: the
// revealed hand is an opponent's, and the exiled card comes out of
// yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4fc198f-3595-4880-ad12-96c18fbf9c33",
		Name:         "Ego Drain",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Nonland(),
				Label:  "nonland card",
			}).Apply(ctx); err != nil {
				return err
			}
			if ControlsA("Faerie")(ctx.Game, item.Controller) {
				return nil
			}
			return egoDrainExile(ctx.Game, item)
		},
	})
}

// egoDrainExile asks the caster which card of their own hand to exile.
func egoDrainExile(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	var hand []uuid.UUID
	if p := g.PlayerByIDForEffect(player); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
	}
	if len(hand) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Ego Drain — exile a card from your hand",
		Cards:    hand,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if err := g.ExileCardForEffect(id); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
}
