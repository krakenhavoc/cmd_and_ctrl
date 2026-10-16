package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leovold, Emissary of Trest — Legendary Creature — Elf Advisor {B}{G}{U},
// 3/3:
//
//	"Each opponent can't draw more than one card each turn.
//	 Whenever you or a permanent you control becomes the target of a
//	 spell or ability an opponent controls, you may draw a card."
//
// #3927. The first line is Narset's static (Spec.DrawLimits, CR 121.2b).
// The second watches EventBecomesTarget, which the engine emits once per
// target, so a spell targeting both Leovold's controller and one of their
// permanents, or two of their permanents, triggers it twice (the
// 2018-12-07 ruling). "A spell or ability an opponent controls" is the
// event's actor, the controller of the targeting object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5d91377-fd66-4dbe-a092-07f2ea379ca7",
		Name:         "Leovold, Emissary of Trest",
		Completeness: CompletenessFull,
		DrawLimits:   EachOpponentCantDrawMoreThanOneCardEachTurn(),
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventBecomesTarget, leovoldTargetedByAnOpponent,
				"Leovold, Emissary of Trest — you may draw a card", Do(DrawCards{N: 1})),
				"Leovold, Emissary of Trest — draw a card?", 1),
		},
	})
}

// leovoldTargetedByAnOpponent is "you or a permanent you control becomes
// the target of a spell or ability an opponent controls".
func leovoldTargetedByAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil {
		return false
	}
	you := source.Controller
	if you == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == you {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.Actor); p == nil {
		return false
	}
	if ev.CardID == uuid.Nil {
		return ev.Target == you
	}
	if !onBattlefield(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == you
}
