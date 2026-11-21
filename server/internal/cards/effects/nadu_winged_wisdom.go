package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nadu, Winged Wisdom — Legendary Creature — Bird Wizard {1}{G}{U}, 3/4 (Modern
// Horizons 3):
//
//	"Flying
//	 Creatures you control have "Whenever this creature becomes the target of a
//	 spell or ability, reveal the top card of your library. If it's a land card,
//	 put it onto the battlefield. Otherwise, put it into your hand. This ability
//	 triggers only twice each turn.""
//
// The granted trigger belongs to each creature (ADR 0093), and its own label is
// counted per object, so each creature may trigger twice a turn.
func init() {
	const (
		key   = "nadu-winged-wisdom/reveal-top"
		label = "Nadu, Winged Wisdom — reveal the top card of your library; a land card goes onto the battlefield, any other into your hand"
		text  = "Whenever this creature becomes the target of a spell or ability, reveal the top card of your library. If it's a land card, put it onto the battlefield. Otherwise, put it into your hand. This ability triggers only twice each turn."
	)
	Register(Spec{
		OracleID:        "0ff11a61-6fe3-4c70-aba8-417023a5e54e",
		Name:            "Nadu, Winged Wisdom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key: key,
			Triggered: []game.TriggeredAbility{
				On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return Self(ev, source, lki, g) && g.TriggeredThisTurn(source.InstanceID, label) < 2
				}, label, gapB52NaduReveal),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller
		}, key), "Creatures you control have \""+text+"\"")},
	})
}

// gapB52NaduReveal reveals the top card of the controller's library: a land
// card enters the battlefield (untapped), anything else goes to the hand.
func gapB52NaduReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 1, "Nadu, Winged Wisdom — reveal the top card")
	if len(revealed) == 0 {
		return nil
	}
	id := revealed[0]
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if c.IsLand() {
		return putCardsFromZoneTogether(ctx, game.ZoneLibrary, []uuid.UUID{id}, item.Controller, false)
	}
	return g.BounceToHandForEffect(id)
}
