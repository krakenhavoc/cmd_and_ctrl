package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sea of Sand — Plane — Rabiah (Planechase Anthology, #4870):
//
//	"Players reveal each card they draw.
//	 Whenever a player draws a land card, that player gains 3 life.
//	 Whenever a player draws a nonland card, that player loses 3 life.
//	 Whenever chaos ensues, put target permanent on top of its owner's
//	 library."
//
// A plane (ADR 0216 §6). The reveal is a draw replacement that changes
// nothing: as each individual card draw is performed (CR 121.2) the top
// card of the drawer's library is revealed to every seat, then drawn. The
// two life abilities are triggers on each drawn card, whoever draws it,
// judged by the card's land-ness. The chaos ability reuses Set Adrift's
// move.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edd21d18-98d2-4e3f-8da1-1ba645b8f69c",
		Name:         "Sea of Sand",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && g.PlayerByIDForEffect(ev.DrawPlayer) != nil
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				g.RevealTopOfLibraryForEffect(ev.DrawPlayer, src.InstanceID, 1, "Sea of Sand — the card being drawn")
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Sea of Sand: players reveal each card they draw",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, drewACardOfKind(true), "Sea of Sand — that player gains 3 life", seaOfSandLife(3)),
			On(game.EventDrawCard, drewACardOfKind(false), "Sea of Sand — that player loses 3 life", seaOfSandLife(-3)),
			Targeting(WheneverChaosEnsues("Sea of Sand — put target permanent on top of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					return putTargetOnTopOfOwnersLibrary(item, NewContext(g, item))
				}),
				TargetPermanent("target permanent")),
		},
	})
}

// drewACardOfKind is "a player draws a land card" (land true) or "a nonland
// card", judged on the card drawn.
func drewACardOfKind(land bool) When {
	return func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsLand() == land
	}
}

func seaOfSandLife(delta int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		who := item.Trigger.Event.Actor
		if who == uuid.Nil {
			return nil
		}
		return g.ChangePlayerLifeForEffect(item.SourceCardID, who, delta)
	}
}
