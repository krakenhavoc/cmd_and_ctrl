package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sanguine Bond — Enchantment {3}{B}{B} (EDHREC rank 505):
//
//	"Whenever you gain life, target opponent loses that much life."
//
// Every point of lifegain becomes a targeted drain. The trigger
// watches EventChangeLife for a positive delta on its controller —
// which every lifegain path in the engine emits, lifelink included
// (combat lifelink, the damage-assignment resume and
// ChangePlayerLifeForEffect all emit it) — and captures the amount
// by value in Build, so a life total that changes again before the
// trigger resolves does not change the drain. The drain is life LOSS,
// not damage, so damage prevention and damage doublers never see it —
// but since #482 it does run the CR 614 life window, so a life-loss
// replacement (Bloodletter of Aclazotz) would.
//
// With Exquisite Blood on the same side of the table the two
// triggers feed each other until the targeted opponent is dead; each
// Bond trigger asks for its target, so the loop is click-through in
// this engine exactly as it is announced in paper, and it ends when
// the eliminated player is no longer a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73089a39-a2f6-4aa2-a058-e6551475153d",
		Name:         "Sanguine Bond",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target == source.Controller && ev.Amount > 0
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Sanguine Bond — target opponent loses that much life",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Sanguine Bond — target opponent loses that much life")
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Targets[0].ID, -item.Params.Amount)
			},
		}},
	})
}
