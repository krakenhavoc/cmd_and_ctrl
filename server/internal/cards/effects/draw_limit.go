package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// draw_limit.go — the card-side vocabulary for Spec.DrawLimits (#3927,
// CR 121.2b). The engine half, and why a limited draw is not performed
// rather than replaced, is game/draw_limit.go.

// EachOpponentCantDrawMoreThanOneCardEachTurn is "Each opponent can't
// draw more than one card each turn." (Narset, Parter of Veils; Leovold,
// Emissary of Trest).
func EachOpponentCantDrawMoreThanOneCardEachTurn() []game.DrawLimitStatic {
	return []game.DrawLimitStatic{{
		Label: "Each opponent can't draw more than one card each turn.",
		Whose: game.DrawLimitOpponents,
		Max:   1,
	}}
}

// EachPlayerCantDrawMoreThanOneCardEachTurn is "Each player can't draw
// more than one card each turn." (Spirit of the Labyrinth).
func EachPlayerCantDrawMoreThanOneCardEachTurn() []game.DrawLimitStatic {
	return []game.DrawLimitStatic{{
		Label: "Each player can't draw more than one card each turn.",
		Whose: game.DrawLimitEveryone,
		Max:   1,
	}}
}

// PlayersCantDrawCards is "Players can't draw cards." (Maralen of the
// Mornsong, Mornsong Aria): the same limit at zero.
func PlayersCantDrawCards() []game.DrawLimitStatic {
	return []game.DrawLimitStatic{{
		Label: "Players can't draw cards.",
		Whose: game.DrawLimitEveryone,
		Max:   0,
	}}
}

// mornsongDrawStep is the second line Maralen of the Mornsong and
// Mornsong Aria share: "At the beginning of each player's draw step,
// that player loses 3 life, searches their library for a card, puts it
// into their hand, then shuffles." "That player" is the active player
// whose draw step began (the event's actor); the search is theirs, and
// it is not a draw, so their "can't draw cards" doesn't stop it (CR
// 121.5). The life loss comes first and is not a cost: a player at 3
// life or less still searches, and loses as a state-based action after.
func mornsongDrawStep(name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventBeginDrawStep},
		Key:     name + " — that player loses 3 life and searches for a card",
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			player := item.Trigger.Event.Actor
			if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
				return nil
			}
			if err := g.ChangePlayerLifeForEffect(item.SourceCardID, player, -3); err != nil {
				return err
			}
			return SearchLibrary{
				Player:  player,
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  name + " — search your library for a card",
			}.Apply(NewContext(g, item))
		},
	}
}
