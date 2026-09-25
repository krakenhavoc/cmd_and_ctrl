package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Obliterator — Creature — Phyrexian Horror {B}{B}{B}{B},
// 5/5 (EDHREC rank 1871):
//
//	"Trample
//	 Whenever a source deals damage to this creature, that source's
//	 controller sacrifices that many permanents of their choice."
//
// The creature nobody wants to block or burn. One trigger per
// damage event — the engine emits EventDealDamage per source, with
// the damaged card in Target and the amount in Amount — so two
// blockers are two triggers, each for its own damage, as printed.
// The source's controller is read as the trigger fires: the combat
// paths stamp it on the event, a spell's damage does not, and the
// spell is still on the stack at that moment with its caster as
// controller. The sacrifice is that player's
// own choice: N sacrifice prompts over every permanent they control
// (Archfiend of Depravity's shape — the engine's prompt picks one
// permanent, so N are asked as N prompts, the option lists trimmed
// after every answer), and a player with fewer permanents than owed
// sacrifices what they have. A permanent chosen by the source's
// controller may be the source itself, as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "41820f91-27cf-41c0-bb5e-9adf6845a6a4",
		Name:            "Phyrexian Obliterator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b17SourceDealtDamageToSelf(ev, source)
			},
			Key: "Phyrexian Obliterator — the source's controller sacrifices that many permanents",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Phyrexian Obliterator — the source's controller sacrifices that many permanents", nil)
				item.Params.Player = b17DamageSourceController(ev, g)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				b17PlayerSacrificesN(g, item.SourceCardID, item.Params.Player, item.Params.Amount,
					"Phyrexian Obliterator — sacrifice a permanent")
				return nil
			},
		}},
	})
}
