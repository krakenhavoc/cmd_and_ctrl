package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Meathook Massacre — Legendary Enchantment {X}{B}{B}:
//
//	"When The Meathook Massacre enters, each creature gets -X/-X until
//	 end of turn.
//	 Whenever a creature you control dies, each opponent loses 1 life.
//	 Whenever a creature an opponent controls dies, you gain 1 life."
//
// X is the X the spell was announced with (CR 107.3m, Card.CastX), read
// once in Build and carried on Params.Amount — the enters ability is a
// real trigger on the stack, not a side effect of the spell resolving.
// A Massacre that entered without being cast (flickered, reanimated)
// has X = 0 and shrinks nothing. The -X/-X hits every creature, the
// caster's included, evaluated once at resolution (CR 611.2c); the
// creatures it kills die as ONE state-based event, so the Massacre's
// own death triggers (already on the battlefield) see all of them.
//
// "You" and "an opponent" are relative to the Massacre's controller at
// the time of death, read off the controller the exit stamped on the
// event (leftUnderControlOf), as Massacre Wurm does. Both payouts are
// life loss / life gain, not damage.
//
// X is not flagged XMatters: at X=0 the death triggers still make the
// card a real play.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "127de52b-df75-4342-95a0-20d84c5bf916",
		Name:         "The Meathook Massacre",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "The Meathook Massacre — each creature gets -X/-X",
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "The Meathook Massacre — each creature gets -X/-X")
					item.Params.Amount = source.CastX()
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					x := item.Params.Amount
					if x <= 0 {
						return nil
					}
					return BoostUntilEOT{
						Match:     Creature(),
						Power:     -x,
						Toughness: -x,
						Label:     "The Meathook Massacre — -X/-X",
					}.Apply(NewContext(g, item))
				},
			},
			TriggerWithPurpose(
				WheneverACreatureYouControlDies("The Meathook Massacre — each opponent loses 1 life",
					Do(EachOpponentLosesLife{N: 1})),
				game.Purpose{DeathPayoff: true}),
			TriggerWithPurpose(
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					dead, ok := diedCreature(ev, g)
					return ok && leftUnderControlOf(ev, dead) != source.Controller
				}, "The Meathook Massacre — you gain 1 life", Do(GainLife{Amount: 1})),
				game.Purpose{DeathPayoff: true}),
		},
	})
}
