package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shadow, Mysterious Assassin — Legendary Creature — Human Assassin
// {2}{B}, 3/3 (deck request #3655):
//
//	"Deathtouch
//	 Throw — Whenever Shadow deals combat damage to a player, you may
//	 sacrifice another nonland permanent. If you do, draw two cards and
//	 each opponent loses life equal to the mana value of the sacrificed
//	 permanent."
//
// The permanent is picked from the other nonland permanents its
// controller has (a "may": picking none declines). Its mana value is
// read before it goes, since the loss is its last-known value (CR
// 608.2h); a token or a permanent with no cost is worth 0. The draw and
// the loss happen only if the sacrifice really took place ("if you do").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13ca49a2-8a78-4be4-976a-b684086864bb",
		Name:            "Shadow, Mysterious Assassin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Shadow, Mysterious Assassin — you may sacrifice another nonland permanent to draw two cards and drain its mana value",
				shadowThrow),
		},
	})
}

func shadowThrow(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, id := range NonlandPermanentsControlledBy(g, item.Controller) {
		if id != item.SourceCardID && !g.CantBeSacrificedForEffect(id) {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Shadow, Mysterious Assassin — sacrifice another nonland permanent? (draw two cards; each opponent loses life equal to its mana value)",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			mv := 0
			if c, ok := g.LookupCardForEffect(picked[0]); ok {
				mv = c.ManaValue()
			}
			return g.SacrificeThenForEffect(uuid.Nil, picked[0], func(g *game.Game, sacrificed bool) error {
				if !sacrificed {
					return nil
				}
				if err := g.DrawNForEffect(item.Controller, 2); err != nil {
					return err
				}
				return eachOpponentLosesLife(g, item, mv)
			})
		},
	})
	return nil
}
