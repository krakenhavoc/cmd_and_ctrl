package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mari, the Killing Quill — Legendary Creature — Vampire Assassin
// {1}{B}{B}, 3/2 (EDHREC rank 2543):
//
//	"Whenever a creature an opponent controls dies, exile it with a
//	 hit counter on it.
//	 Assassins, Mercenaries, and Rogues you control have deathtouch
//	 and "Whenever this creature deals combat damage to a player, you
//	 may remove a hit counter from a card that player owns in exile.
//	 If you do, draw a card and create two Treasure tokens.""
//
// The bounty-hunter commander. Three pieces:
//
//   - The first trigger is Sangromancer's condition
//     (b18OpponentsCreatureDied); on resolution the dead card is
//     exiled from the graveyard and a hit counter is put on it there
//     — counters live in every zone. A card that left the graveyard
//     in response is not exiled (CR 400.7), and a token is gone by
//     then and gets nothing.
//   - The deathtouch half of the grant is a Layer 6 TribalKeywordGrant
//     over the three types, "you control", Mari included (she is an
//     Assassin); a changeling counts.
//   - The granted trigger has no slot of its own in the catalog (a
//     static cannot add a triggered ability to another permanent),
//     so it is modelled the way Dionus models his: as Mari's own
//     trigger, watching combat damage to a player by any Assassin,
//     Mercenary or Rogue her controller controls
//     (b24CombatDamageToPlayerByYourCreatureOfSubtypes). Observably
//     the same — it exists while Mari is on the battlefield, applies
//     to creatures under HER controller's control, and fires once per
//     creature that connects. "You may" is the controller's yes/no as
//     the trigger goes on the stack; the damaged player is captured
//     in Build; on resolution a hit counter comes off a card that
//     player owns in exile, and if one did, the controller draws a
//     card and makes two Treasures. With no such card the ability
//     does nothing — "if you do" is the gate.
//
// Which exiled card loses its counter is not a choice the engine
// asks: every hit-countered card that player owns is interchangeable
// for the effect, so the first in exile order is taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f1fdc23-9af0-41a1-aeba-7288f9642734",
		Name:         "Mari, the Killing Quill",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Assassin", "Mercenary", "Rogue"}, YoursOnly: true}, "deathtouch"),
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b18OpponentsCreatureDied(ev, source, g)
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					dead := ev.CardID
					return game.NewTriggeredItem(source, "Mari, the Killing Quill — exile it with a hit counter on it",
						func(g *game.Game, item *game.StackItem) error {
							if z := g.FindCardZoneForEffect(dead); z == nil || z.Kind != game.ZoneGraveyard {
								return nil
							}
							ctx := NewContext(g, item)
							if err := (ExileTarget{Target: dead}).Apply(ctx); err != nil {
								return err
							}
							return AddCounter{Target: dead, Kind: "hit", N: 1}.Apply(ctx)
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b24CombatDamageToPlayerByYourCreatureOfSubtypes(ev, source, g, "Assassin", "Mercenary", "Rogue")
				},
				HasLegalTarget: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := b24ExiledCardWithCounterOwnedBy(g, ev.Target, "hit")
					return ok
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					victim := ev.Target
					return game.NewTriggeredItem(source, "Mari, the Killing Quill — remove a hit counter, draw a card and create two Treasures",
						func(g *game.Game, item *game.StackItem) error {
							return b24RemoveHitCounterDrawAndTreasures(g, item, victim)
						})
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Mari, the Killing Quill — remove a hit counter from a card that player owns in exile?"},
			},
		},
	})
}
