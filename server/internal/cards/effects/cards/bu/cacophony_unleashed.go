package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cacophony Unleashed — Enchantment {5}{B}{B} (Commander Masters,
// #5948):
//
//	"When this enchantment enters, if you cast it, destroy all
//	 nonenchantment creatures.
//	 Whenever this enchantment or another enchantment you control enters,
//	 until end of turn, this enchantment becomes a legendary 6/6 Nightmare
//	 God creature with menace and deathtouch. It's still an enchantment."
//
// The wipe is an intervening "if you cast it" (CR 603.4): checked as the
// enchantment enters and again as the trigger resolves, so a copy or a
// reanimated one does not wipe the board. The animation is a layer-4
// type change plus a layer-7b base size (CR 613.1d, CR 613.4b) that
// lasts until end of turn, so it can attack once the enchantment has been
// under its controller's control since the turn began (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f6056be-67c8-40a9-86ef-1606961644eb",
		Name:         "Cacophony Unleashed",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: every creature except enchantment creatures, and only when cast.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.B16EnteredFromStack(g, source.InstanceID)
			}, "Cacophony Unleashed — destroy all nonenchantment creatures", func(g *game.Game, item *game.StackItem) error {
				if !effects.B16EnteredFromStack(g, item.SourceCardID) {
					return nil
				}
				return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.Not(effects.Enchantment()))}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventETB {
					return false
				}
				if ev.CardID == source.InstanceID {
					return true
				}
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.HasCardType("enchantment")
			}, "Cacophony Unleashed — until end of turn, it becomes a legendary 6/6 Nightmare God creature with menace and deathtouch",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: append([]game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSupertypesMod("Legendary"),
							game.AddSubtypesMod(ctx.TextSubtypes("Nightmare", "God")...),
							game.AddKeywordsMod("menace", "deathtouch"),
						}, game.SetBasePTMods(6, 6)...),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Cacophony Unleashed — a legendary 6/6 Nightmare God creature with menace and deathtouch until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
