package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rendmaw, Creaking Nest — Legendary Artifact Creature — Scarecrow
// {3}{B}{G}, 5/5:
//
//	"Reach, menace
//	 When Rendmaw enters and whenever you play a card with two or more card
//	 types, each player creates a tapped 2/2 black Bird creature token with
//	 flying. The tokens are goaded for the rest of the game. (They attack
//	 each combat if able and attack a player other than you if able.)"
//
// One printed ability with three conditions (OnAny): Rendmaw entering, a
// spell its controller casts, and a land its controller plays (CR 305.1,
// the settled entry's Played marker). The card has two or more card types
// when its effective types number two or more (an artifact creature, an
// enchantment creature, a Dryad Arbor's land and creature): read off the
// spell on the stack or the land as it entered. Subtypes and supertypes are
// not card types. A land put onto the battlefield without being played does
// not count (CR 305.4).
//
// Each player creates their own Bird, tapped, in turn order from the active
// player; each is goaded for the rest of the game by Rendmaw's controller,
// which includes that player's own Bird. See
// cmdr2S34EachPlayerCreatesGoadedBird for the chain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2207fb3f-45a1-401b-99e9-f75723495c3f",
		Name:            "Rendmaw, Creaking Nest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "menace"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventCast, game.EventZoneMove}, rendmawTriggers,
				"Rendmaw, Creaking Nest — each player creates a tapped 2/2 Bird, goaded for the rest of the game",
				func(g *game.Game, item *game.StackItem) error {
					return cmdr2S34EachPlayerCreatesGoadedBird(NewContext(g, item))
				}),
		},
	})
}

// rendmawTriggers is "When Rendmaw enters" or "whenever you play a card
// with two or more card types".
func rendmawTriggers(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil {
		return false
	}
	switch ev.Kind {
	case game.EventETB:
		return ev.CardID == source.InstanceID
	case game.EventCast:
		return ev.Actor == source.Controller && rendmawTwoTypes(g, ev)
	case game.EventZoneMove:
		return ev.Actor == source.Controller && b20LandPlayed(ev, g) && rendmawTwoTypes(g, ev)
	}
	return false
}

func rendmawTwoTypes(g *game.Game, ev game.Event) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && len(c.Effective().Types) >= 2
}
