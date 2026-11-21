package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Spear of Leonidas — Legendary Artifact — Equipment {2}{R} (Assassin's
// Creed, #5948):
//
//	"Whenever equipped creature attacks, choose one —
//	 • Bull Rush — It gains double strike until end of turn.
//	 • Summon — Create Phobos, a legendary 3/2 red Horse creature token.
//	 • Revelation — Discard two cards, then draw two cards.
//	 Equip {2}"
//
// A modal triggered ability on the Equipment, keyed on the creature it is
// attached to (CR 700.2, 603.3c): the mode is chosen as it goes on the stack.
// "It" in the first mode is the attacker, read off the triggering event, not
// whatever the Spear is attached to by resolution. The token is the printed
// Phobos, legendary, so a second one is put in the graveyard by the legend
// rule. Revelation's discard is the controller's choice of two cards (all of a
// smaller hand, CR 701.9a); the draw follows once every discard has settled
// and is not conditional on the discards.
//
// No simplification.
func init() {
	t := effects.On(game.EventAttack,
		func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return effects.AttachedCreatureAttacked(ev, source)
		},
		"The Spear of Leonidas — choose one", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeDoing("Bull Rush — It gains double strike until end of turn.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				id := ctx.Trigger().Event.CardID
				if !effects.OnBattlefield(ctx.Game, id) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "The Spear of Leonidas — double strike"}.Apply(ctx)
			}),
		effects.ModeSteps("Summon — Create Phobos, a legendary 3/2 red Horse creature token.", nil,
			effects.CreateToken{Template: effects.TokenCard("3/2 red Legendary Horse named Phobos"), N: 1}),
		effects.ModeDoing("Revelation — Discard two cards, then draw two cards.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   item.Controller,
					Source:   item.SourceCardID,
					N:        2,
					Question: "The Spear of Leonidas — discard two cards, then draw two cards",
				}, spearOfLeonidasDraw, nil)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "dac3a263-cc02-446a-88ec-882a21096b6c",
		Name:         "The Spear of Leonidas",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// spearOfLeonidasDraw is "then draw two cards", run once the discards have
// settled. An on-disk key: never renamed, never reused.
var spearOfLeonidasDraw = game.RegisterDiscardThen("choose-cards/spear-of-leonidas-draw", func(g *game.Game, r game.DiscardsDone) error {
	return g.DrawNForEffect(r.Player, 2)
})
