package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dread Defiler — {6}{B} Creature — Eldrazi 6/8 (Oath of the Gatewatch, S82
// OLD b296):
//
//	"Devoid (This card has no color.)
//	 {3}{C}, Exile a creature card from your graveyard: Target opponent loses
//	 life equal to the exiled card's power."
//
// Corpse Explosion's read of the exiled card: the cost names the creature card
// at announcement (CR 602.2b) and the effect reads its power from exile as it
// resolves. The {C} is paid with colorless mana only. A target gone by
// resolution loses nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ebdd0027-3340-459e-a598-dd416e76744f",
		Name:            "Dread Defiler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{C}, Exile a creature card from your graveyard: Target opponent loses life equal to the exiled card's power.",
			Cost:    effects.Plus(effects.ManaCost("{3}{C}"), effects.ExileFromGraveyard(1, "a creature card", effects.IsCreatureCard)),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.TargetedPlayerOf(ctx)
				if id == uuid.Nil {
					return nil
				}
				power := 0
				for _, cid := range ctx.Exiled() {
					if c, ok := g.LookupCardForEffect(cid); ok {
						power = c.CurrentPower()
					}
				}
				if power <= 0 {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, id, -power)
			},
		}},
	})
}
