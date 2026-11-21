package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Higure, the Still Wind — Legendary Creature — Human Ninja {3}{U}{U},
// 3/4 (Betrayers of Kamigawa, #5602):
//
//	"Ninjutsu {2}{U}{U} ({2}{U}{U}, Return an unblocked attacker you
//	 control to hand: Put this card onto the battlefield from your hand
//	 tapped and attacking.)
//	 Whenever Higure deals combat damage to a player, you may search your
//	 library for a Ninja card, reveal it, put it into your hand, then
//	 shuffle.
//	 {2}: Target Ninja creature can't be blocked this turn."
//
// The keyword is Ninjutsu("{2}{U}{U}") (ninjutsu.go). The search is the
// printed "you may search" (SearchLibrary.Optional), so the controller
// may decline both the card and the shuffle. The activated ability is
// Rogue's Passage's RestrictUntilEOT with CantBeBlocked on one Ninja
// creature, any player's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01197456-62f1-4bc0-9f2f-127f97f9b05d",
		Name:         "Higure, the Still Wind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Higure, the Still Wind — you may search your library for a Ninja card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Ninja") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Higure, the Still Wind — a Ninja card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.Ninjutsu("{2}{U}{U}"),
			{
				Label:   "{2}: Target Ninja creature can't be blocked this turn.",
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    effects.ManaCost("{2}"),
				Targets: effects.TargetCreature("target Ninja creature", effects.Subtype("Ninja")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked,
						Label: "Higure, the Still Wind — can't be blocked this turn"}.Apply(ctx)
				},
			},
		},
	})
}
