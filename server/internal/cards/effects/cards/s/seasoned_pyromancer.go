package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seasoned Pyromancer — Creature — Human Shaman {1}{R}{R}, 2/2 (Modern
// Horizons 3):
//
//	"When this creature enters, discard two cards, then draw two cards. For
//	 each nonland card discarded this way, create a 1/1 red Elemental
//	 creature token.
//	 {3}{R}{R}, Exile this card from your graveyard: Create two 1/1 red
//	 Elemental creature tokens."
//
// You choose the two cards (CR 701.9b; a hand of one or none discards what
// it has, CR 701.9a, and still draws two). The discard is a keyed
// continuation (Mind Roots'): once the cards have really reached the
// graveyard it draws two, then counts the discarded cards that are not lands
// as they now sit in the graveyard (a card a replacement sent elsewhere is
// not counted) and makes that many Elementals. The second ability is Mother
// Bear's graveyard activation: the exile is part of the cost, and it works
// only from the graveyard. It has no sorcery-speed restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02238355-9c84-4ae0-b850-269f460144a8",
		Name:         "Seasoned Pyromancer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Seasoned Pyromancer — discard two cards, then draw two cards; create an Elemental for each nonland card discarded",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        2,
						Question: "Seasoned Pyromancer — discard two cards",
						Asked:    game.NewAskedItem(item),
					}, seasonedPyromancerDiscarded, nil)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}{R}, Exile this card from your graveyard: Create two 1/1 red Elemental creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 2, CreatureTokens: 2},
			Cost:    effects.Plus(effects.ManaCost("{3}{R}{R}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Elemental"), N: 2}),
		}},
	})
}

// seasonedPyromancerDiscarded is the clause after the discard: draw two, then
// one Elemental per nonland card discarded this way. An on-disk key: never
// renamed, never reused.
var seasonedPyromancerDiscarded = game.RegisterDiscardThen("choose-cards/seasoned-pyromancer-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil {
		return nil
	}
	nonland := 0
	for _, id := range r.Discarded.Cards() {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			nonland++
		}
	}
	ctx := effects.NewContext(g, r.Item)
	if err := (effects.DrawCards{N: 2}).Apply(ctx); err != nil {
		return err
	}
	if nonland == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("1/1 red Elemental"), N: nonland}.Apply(ctx)
})
