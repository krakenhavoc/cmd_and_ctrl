package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deepfathom Skulker — Creature — Eldrazi {5}{U}, 4/4 (Modern Horizons 3
// Commander, #5948):
//
//	"Devoid (This card has no color.)
//	 Whenever a creature you control deals combat damage to a player, you
//	 may draw a card.
//	 {3}{C}: Target creature can't be blocked this turn."
//
// The draw is Bident of Thassa's: one event per creature per player, so
// three unblocked attackers queue three "you may draw" prompts. The
// activated ability's {C} is colorless mana only (CR 107.4b): a coloured
// or generic payment does not pay it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a09a4556-8913-495c-8768-e05413c48f0b",
		Name:            "Deepfathom Skulker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Deepfathom Skulker — draw a card", Do(DrawCards{N: 1})), "Deepfathom Skulker — draw a card?", 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{C}: Target creature can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    ManaCost("{3}{C}"),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Deepfathom Skulker — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
