package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gríma, Saruman's Footman — Legendary Creature — Human Advisor {2}{U}{B},
// 1/4 (The Hobbit Commander):
//
//	"Gríma can't be blocked.
//	 Whenever Gríma deals combat damage to a player, that player exiles
//	 cards from the top of their library until they exile an instant or
//	 sorcery card. You may cast that card without paying its mana cost.
//	 Then that player puts the exiled cards that weren't cast this way on
//	 the bottom of their library in a random order."
//
// Tibalt's Trickery's walk with two different players: the damaged player's
// library is exiled from (and their cards are the ones that go back under),
// while Gríma's controller is the one offered the free cast
// (Game.ExileUntilCastFreeByForEffect). The cast is made during this
// resolution (CR 608.2g) through the resolution_cast prompt; declining, or
// a hit that cannot be cast, sends it to the bottom with the rest of the
// pile in one random order, and a library with no instant or sorcery card
// exiles all of itself and bottoms it again. The hit is "an instant or
// sorcery card": a token in a library is never one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9df60bb7-2dd4-4d22-b814-f52145356424",
		Name:         "Gríma, Saruman's Footman",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Gríma, Saruman's Footman — that player exiles cards from the top of their library until an instant or sorcery card; you may cast it without paying its mana cost",
				grimaSarumansFootmanExile),
		},
	})
}

// grimaSarumansFootmanExile is the trigger's body.
func grimaSarumansFootmanExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return g.ExileUntilCastFreeByForEffect(item.Trigger.Event.Target, item.Controller, item.SourceCardID,
		func(c game.Card) bool { return !c.IsToken() && (c.IsInstant() || c.IsSorcery()) })
}
