package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disturbing Mirth — Enchantment {B}{R}:
//
//	"When this enchantment enters, you may sacrifice another
//	 enchantment or creature. If you do, draw two cards.
//	 When you sacrifice this enchantment, manifest dread."
//
// The sacrifice is chosen as the trigger resolves, with a floor of zero
// for the "you may", and the draw follows only from a permanent that
// actually went (a sacrificed commander still pays out, because the
// continuation is told what was sacrificed). The second trigger
// watches EventSacrifice for this enchantment, which is emitted before
// the zone move, so it triggers on the sacrifice itself; any way this
// enchantment dies other than a sacrifice does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7010bfbb-8aea-473b-a9a5-27488de13cc5",
		Name:         "Disturbing Mirth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Disturbing Mirth — you may sacrifice another enchantment or creature; if you do, draw two cards",
				disturbingMirthSacrifice),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor == source.Controller
			}, "Disturbing Mirth — manifest dread", Do(ManifestDread{})),
		},
	})
}

func disturbingMirthSacrifice(g *game.Game, item *game.StackItem) error {
	me, src := item.Controller, item.SourceCardID
	candidates := permanentsControlledByMatching(g, me, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.InstanceID != src && (c.IsCreature() || c.IsEnchantment())
	})
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   src,
		Question: "Disturbing Mirth — you may sacrifice another enchantment or creature to draw two cards",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return g.SacrificeAllThenForEffect(uuid.Nil, picked, func(g *game.Game, sacrificed []uuid.UUID) error {
				if len(sacrificed) == 0 {
					return nil
				}
				return g.DrawNForEffect(me, 2)
			})
		},
	})
	return nil
}
