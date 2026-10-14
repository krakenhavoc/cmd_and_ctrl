package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuna, Grand Summoner — Legendary Creature — Human Cleric {1}{G}{W}{U},
// 1/5:
//
//	"Grand Summon — {T}: Add one mana of any color. When you next cast a
//	 creature spell this turn, that creature enters with two additional
//	 +1/+1 counters on it.
//	 Whenever another permanent you control is put into a graveyard from
//	 the battlefield, if it had one or more counters on it, you may put
//	 that number of +1/+1 counters on target creature."
//
// Grand Summon is a mana ability (CR 605.1a) whose rider sets up the
// next creature spell: Savage Summoning's one-use promise
// (game.NextSpellPromise) carrying two +1/+1 counters, spent by the
// first creature spell you cast this turn and marked on it as it becomes
// cast, so the creature enters with them (CR 614.1c). Two activations
// before casting set up two promises, and the next creature takes both,
// as two delayed triggers would. It ends at cleanup. The delayed trigger
// is not a stack object here, so it can't itself be countered.
//
// The graveyard trigger is Resourceful Defense's: "another" permanent you
// controlled, put into a graveyard, with counters on it as it last
// existed (an intervening if, CR 603.4, asked as it triggers and as it
// resolves). "That number" is every counter it had, of every kind. The
// "you may" is the trigger's prompt; the target is chosen with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba72d4a9-6f36-4796-bd71-2342fa9b4a30",
		Name:         "Yuna, Grand Summoner",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Grand Summon — Add one mana of any color",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				g.GrantNextSpellPromiseForEffect(controller, game.NextSpellPromise{
					Filter:      game.PermissionFilter{CreatureOnly: true},
					CounterKind: game.CounterPlusOne,
					Counters:    2,
					Text:        "The next creature spell you cast this turn enters with two additional +1/+1 counters.",
				}, "Yuna, Grand Summoner", source, g.UntilEndOfTurnDuration())
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(On(game.EventLTB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.NewZone == game.ZoneGraveyard && b14PermanentYouControlLeft(ev, source, g)
				},
				"Yuna, Grand Summoner — put that many +1/+1 counters on target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					left, ok := ctx.TriggeringPermanent()
					if !ok {
						return nil
					}
					n := 0
					for _, k := range left.Counters {
						if k > 0 {
							n += k
						}
					}
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok || n == 0 {
						return nil
					}
					return nothingIfGone(g.AddCounterByForEffect(item.Controller, id, game.CounterPlusOne, n))
				}),
				TargetCreature("target creature")),
				"Yuna, Grand Summoner — put that many +1/+1 counters on target creature?"),
		},
	})
}
