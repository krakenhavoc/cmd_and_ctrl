package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sliver of Hope — Creature — Sliver {3}{W}, 3/3:
//
//	"Slivers you control have hope. (Prevent all damage that would be dealt
//	 to attacking creatures with hope.)"
//
// Hope is a keyword the card's own reminder text defines: it is granted as a
// layer-6 keyword to the Slivers its controller controls (a lord's grant, so
// the Sliver of Hope itself has it, CR 613.1f), and the damage prevention is
// a static replacement (CR 615.1a) that cancels every instance of damage, from
// any source and of either kind, dealt to an ATTACKING creature that has hope
// at that moment. A Sliver that is not attacking, or that lost the keyword
// when Sliver of Hope left, is hit normally.
func init() {
	Register(Spec{
		OracleID:     "4d57d6a0-2c4d-466c-8545-f97058a8f1c5",
		Name:         "Sliver of Hope",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, sliverHopeKeyword),
		},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true,
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.DamageTarget)
				return ok && c.AttackingTarget != uuid.Nil && game.HasKeyword(&c, sliverHopeKeyword)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Sliver of Hope — prevent all damage dealt to attacking creatures with hope",
		}},
	})
}

const sliverHopeKeyword = "hope"
