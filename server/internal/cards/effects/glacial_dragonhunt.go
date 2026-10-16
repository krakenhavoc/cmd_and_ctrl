package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glacial Dragonhunt — Sorcery {U}{R} (#3931):
//
//	"Draw a card, then you may discard a card. When you discard a nonland
//	 card this way, Glacial Dragonhunt deals 3 damage to target creature.
//	 Harmonize {4}{U}{R}"
//
// The loot is the draw-then-discard machinery with an "up to one"
// discard, so the choice opens once the card is drawn and the drawn card
// is a legal discard. The second sentence is a reflexive trigger (CR
// 603.12) created only when the discarded card was nonland; its target
// is chosen as it goes on the stack, and the damage comes from Glacial
// Dragonhunt as it last existed (Foray of Orcs' shape: the spell has
// left the stack by then, and was exiled if it was cast with
// harmonize). Harmonize is CR 702.180.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "8f976399-29a9-4eb7-844d-3d3f2fdf6ea7",
		Name:             "Glacial Dragonhunt",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{4}{U}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
				Ref: glacialDragonhuntDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
			})
		},
	})
}

var (
	// The loot's discard half: "then you may discard a card", run once
	// the draw has landed. A nonland discard creates the reflexive
	// trigger. The key is an on-disk identity.
	glacialDragonhuntDiscard = game.RegisterDrawThen("glacial-dragonhunt/may-discard", func(g *game.Game, d game.DrawThen) error {
		parent := &game.StackItem{Controller: d.Player, SourceCardID: d.Source, Label: "Glacial Dragonhunt"}
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        1,
			UpTo:     true,
			Question: "Glacial Dragonhunt — you may discard a card",
			Then: func(g *game.Game, _ uuid.UUID, discarded []uuid.UUID) error {
				nonland := false
				for _, id := range discarded {
					if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
						nonland = true
					}
				}
				if !nonland {
					return nil
				}
				t := WhenYouDo("Glacial Dragonhunt — 3 damage to target creature", glacialDragonhuntDamageBody)
				t.Params = game.EffectParams{Amount: 3}
				return t.Apply(NewContext(g, parent))
			},
		})
		return nil
	})

	// "When you discard a nonland card this way, Glacial Dragonhunt deals
	// 3 damage to target creature." Breeches' damage body, the 3 in
	// Params.Amount.
	glacialDragonhuntDamageBody = game.ReflexiveBody("glacial-dragonhunt/damage", breechesBlastEffect,
		constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))
)
