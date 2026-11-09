package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ketramose, the New Dawn — Legendary Creature — God {1}{W}{B}, 4/4
// (Aetherdrift, #4464):
//
//	"Menace, lifelink, indestructible
//	 Ketramose can't attack or block unless there are seven or more cards
//	 in exile.
//	 Whenever one or more cards are put into exile from graveyards and/or
//	 the battlefield during your turn, you draw a card and lose 1 life."
//
// The restriction is Hazoret's self-only layer-6 static, read on every
// recompute against the total number of cards in exile, whoever owns them
// and however they were exiled (the printed ruling); once Ketramose has been
// declared an attacker or blocker it stays one if the count falls. The
// trigger is Laelia's batched exile watcher over the graveyard and
// battlefield origins of any player, on your turn only: the engine emits one
// event per card, so OncePerBatch collapses a mass exile into one trigger.
// A token exiled from the battlefield is not a card and does not count.
//
// No simplification.
func init() {
	const label = "Ketramose, the New Dawn — you draw a card and lose 1 life"
	Register(Spec{
		OracleID:        "de41ad94-b9d7-4c2b-bfa6-394d18737228",
		Name:            "Ketramose, the New Dawn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "lifelink", "indestructible"},
		Static: []game.StaticAbility{{
			Label: "Ketramose can't attack or block unless there are seven or more cards in exile.",
			Layer: game.Layer6Ability,
			// The count changes without anything on the battlefield moving:
			// a card arriving in or leaving exile from a hand or a library
			// re-reads the layers (game.StaticAbility.DependsOnExile).
			DependsOnExile: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && stdB15CardsInExile(g) < 7
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventLTB}, stdB15CardExiledOnYourTurn,
				label, Do(DrawCards{N: 1}, GainLife{Amount: -1}))),
		},
	})
}

// stdB15CardsInExile counts every card in exile, face down or up. It runs
// inside a layer pass, so it walks the zone and takes no lock.
func stdB15CardsInExile(g *game.Game) int {
	if g == nil || g.Exile == nil {
		return 0
	}
	return len(g.Exile.Cards)
}

// stdB15CardExiledOnYourTurn is "one or more cards are put into exile from
// graveyards and/or the battlefield during your turn": a card, not a token,
// moved into exile from a graveyard (EventZoneMove) or the battlefield
// (EventLTB) while it is the source controller's turn.
func stdB15CardExiledOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch {
	case ev.Kind == game.EventZoneMove && ev.OldZone == game.ZoneGraveyard && ev.NewZone == game.ZoneExile:
	case ev.Kind == game.EventLTB && ev.NewZone == game.ZoneExile:
	default:
		return false
	}
	if !IsYourTurn(g, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}
