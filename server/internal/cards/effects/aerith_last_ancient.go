package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerith, Last Ancient — Legendary Creature — Human Cleric Druid
// {2}{G}{W}, 3/5 (#3656):
//
//	"Lifelink
//	 Raise — At the beginning of your end step, if you gained life this
//	 turn, return target creature card from your graveyard to your
//	 hand. If you gained 7 or more life this turn, return that card to
//	 the battlefield instead."
//
// A targeted end-step trigger with an intervening if (CR 603.4): the
// life gained comes off the turn tally, checked as the trigger would go
// on the stack and again as it resolves. The 7-or-more branch is read
// at resolution, so life gained in response counts, and the card
// returns under its owner's control.
//
// No simplification.
func init() {
	const label = "Aerith, Last Ancient — Raise: return target creature card from your graveyard to your hand, or to the battlefield if you gained 7 or more life"
	Register(Spec{
		OracleID:        "d6cba688-faff-4d9f-8148-37d389fa5fb3",
		Name:            "Aerith, Last Ancient",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, lki, g) && b15LifeGainedThisTurn(g, source.Controller) >= 1
			},
			Key:     label,
			Targets: TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				gained := b15LifeGainedThisTurn(g, item.Controller)
				if gained < 1 || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				dest := game.ZoneHand
				if gained >= 7 {
					dest = game.ZoneBattlefield
				}
				return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: dest}.Apply(NewContext(g, item))
			},
		}},
	})
}
