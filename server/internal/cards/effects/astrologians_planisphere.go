package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Astrologian's Planisphere — Artifact — Equipment {1}{U}:
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature is a Wizard in addition to its other types and
//	 has 'Whenever you cast a noncreature spell and whenever you draw
//	 your third card each turn, put a +1/+1 counter on this creature.'
//	 Diana — Equip {2}"
//
// The granted ability belongs to the equipped creature (ADR 0093), so
// "this creature" is the host and its controller is "you". It is one
// ability with two triggers: a noncreature spell cast, and the third
// draw of the turn (Detective of the Month's tally read: the count is
// bumped before the harvester runs, so the third draw reports exactly
// three).
//
// No simplification.
func init() {
	const counterKey = "astrologians-planisphere/counter"
	grow := OnAny([]game.EventKind{game.EventCast, game.EventDrawCard},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventCast {
				return b10NoncreatureSpellCastByYou(ev, source, lki, g)
			}
			return ev.Actor == source.Controller && g.TurnTallyFor(source.Controller).CardsDrawn == 3
		},
		"Astrologian's Planisphere — put a +1/+1 counter on this creature",
		func(g *game.Game, item *game.StackItem) error {
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
		})
	Register(Spec{
		OracleID:     "4dcd0583-cee0-46a6-ba43-93bca5eb0428",
		Name:         "Astrologian's Planisphere",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       counterKey,
			Triggered: []game.TriggeredAbility{grow},
			Text:      "Whenever you cast a noncreature spell and whenever you draw your third card each turn, put a +1/+1 counter on this creature.",
		}},
		Static: []game.StaticAbility{
			AttachedIsAlso("Wizard"),
			GrantAbilitiesToAttached(counterKey),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Astrologian's Planisphere")},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
