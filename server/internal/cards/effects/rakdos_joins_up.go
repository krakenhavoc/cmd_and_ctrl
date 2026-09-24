package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rakdos Joins Up — Legendary Enchantment {3}{B}{R} (EDHREC rank
// 3508):
//
//	"When Rakdos Joins Up enters, return target creature card from
//	 your graveyard to the battlefield with two additional +1/+1
//	 counters on it.
//	 Whenever a legendary creature you control dies, Rakdos Joins Up
//	 deals damage equal to that creature's power to target opponent."
//
// A reanimation spell that stays to punish every legend's death.
// The entry trigger targets a creature card in the controller's
// graveyard and brings it back under the controller's control, then
// puts the two counters on it. The dies trigger is diedCreature
// narrowed to a legendary creature the controller controlled, with
// the opponent picked at the prompt and the dead legend's power read
// as the trigger resolves.
//
// One shape note, not a caveat: the counters land a beat AFTER the
// entry rather than as part of it (the reanimation path carries no
// counter option). Nothing in the catalog reads a creature's counters
// between its arrival and the next event, and Doubling Season doubles
// them either way.
//
// The dead legend's power — previously a declared gap — is closed by
// #1379's resolution-time LKI: ctx.TriggeringPermanent() reads
// PermanentInfo.Power, PowerForComparison as the legend last existed
// on the battlefield, layers (an anthem's bonus) and counters both
// included. See b33DamageChosenOpponentByDeadCreaturesPower.
func init() {
	Register(Spec{
		OracleID:     "6a47865c-8fa2-4cb2-aebf-8009c065395f",
		Name:         "Rakdos Joins Up",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Rakdos Joins Up — return the creature card to the battlefield with two +1/+1 counters",
						b33ReanimateChosenWithCounters(2))
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b33LegendaryCreatureYouControlDied(ev, source, g)
				},
				Targets: TargetPlayer("target opponent", Opponent()),
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Rakdos Joins Up — deal damage equal to the legend's power to target opponent",
						b33DamageChosenOpponentByDeadCreaturesPower(ev.CardID))
				},
			},
		},
	})
}
