package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Righteous Indignation — Enchantment {2}{W} (Mercadian Masques, S82 OLD,
// #5619):
//
//	"Whenever a creature blocks a black or red creature, the blocking creature
//	 gets +1/+1 until end of turn."
//
// One trigger for each creature-and-attacker pair (CR 509.3b: a creature that
// blocks two attackers triggers once for each black or red one), whoever controls
// the blocker. The attacker's colour is read as the block is declared, after
// layers, so a colourless or green attacker gives nothing, and the bonus goes to
// the blocker only while it is still the permanent that blocked (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97409c60-537e-4eeb-9d63-be0fde7e8284",
		Name:         "Righteous Indignation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, righteousIndignationBlockedBlackOrRed,
				"Righteous Indignation — the blocking creature gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					blocker := ctx.Trigger().Event.CardID
					if !effects.OnBattlefield(g, blocker) {
						return nil
					}
					return effects.BoostUntilEOT{Target: blocker, Power: 1, Toughness: 1, Label: "Righteous Indignation — +1/+1 until end of turn"}.Apply(ctx)
				}),
		},
	})
}

// righteousIndignationBlockedBlackOrRed is "a creature blocks a black or red
// creature": the event's attacker is black or red.
func righteousIndignationBlockedBlackOrRed(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBlock {
		return false
	}
	attacker, ok := g.LookupCardForEffect(ev.Target)
	return ok && (attacker.HasColor(source.TextColor("B")) || attacker.HasColor(source.TextColor("R")))
}
