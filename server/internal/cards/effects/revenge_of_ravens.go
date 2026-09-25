package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revenge of Ravens — Enchantment {3}{B} (EDHREC rank 2848):
//
//	"Whenever a creature attacks you or a planeswalker you control,
//	 that creature's controller loses 1 life and you gain 1 life."
//
// The pillow-fort drain. One trigger per attacking creature aimed at
// the controller — b17OpponentsCreatureAttackedYou resolves the
// defending player behind a planeswalker attack — with the
// attacker's controller carried into the item BY VALUE (a copied
// UUID, clone-safe) so the drain lands on the right player even
// after the attacker has left. The loss is first and the gain
// second, as printed; the gain is a real life-gain event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd96f145-73eb-4a6d-bcfd-5d6313fac1f9",
		Name:         "Revenge of Ravens",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17OpponentsCreatureAttackedYou(ev, source, g)
			},
			Key: "Revenge of Ravens — the attacker's controller loses 1 life, you gain 1 life",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Revenge of Ravens — the attacker's controller loses 1 life, you gain 1 life")
				item.Params.Player = ev.Actor
				return item
			},
			Effect: b27LoseOneAndYouGainOne,
		}},
	})
}
