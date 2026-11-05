package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exploratory Mission — {2}{R}{R} Enchantment (Star Trek, #4973):
//
//	"Landfall — Whenever a land you control enters, put a quest counter on
//	 this enchantment. Then exile cards from the top of your library until
//	 you exile a nonland card. You may play that card this turn. If this
//	 enchantment has five or more quest counters on it, add {R}{R}."
//
// One landfall trigger, three steps in printed order. The counter goes on
// the enchantment only while it is still the same object (CounterOnThis).
// The exile is Darkchild's walk (MillToZone until a nonland card), and its
// keyed tail grants a play permission for this turn over "that card", the
// nonland card, alone: the lands exiled on the way stay in exile. The
// mana is added last, through AddMana (a triggered ability that is not a
// mana ability, CR 605.1b, so it resolves on the stack), when the
// enchantment has five or more quest counters as the ability resolves,
// or had them as it last existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fdba8ee-6dfe-480d-b84a-43922229d9b4",
		Name:         "Exploratory Mission",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Exploratory Mission — landfall: a quest counter, exile until a nonland card you may play this turn, then {R}{R} at five quest counters",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CounterOnThis{Kind: "quest", N: 1}).Apply(ctx); err != nil {
						return err
					}
					return MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    UntilNonland,
						ThenTail: ctx.ItemTail(exploratoryMissionExiled),
					}.Apply(ctx)
				}),
		},
	})
}

// exploratoryMissionExiled is the rest of the trigger after the walk: the
// play permission over the nonland card still in exile, then the {R}{R}.
// An on-disk key.
var exploratoryMissionExiled = CardTail("exploratory-mission/play-nonland-then-mana", func(ctx *Context, o game.TailOutcome) error {
	var nonland uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		z := ctx.Game.FindCardZoneForEffect(id)
		if ok && !c.IsLand() && z != nil && z.Kind == game.ZoneExile {
			nonland = id
		}
	}
	if nonland != uuid.Nil {
		c, _ := ctx.Game.LookupCardForEffect(nonland)
		perm := ExileTopWithPermission{}.permission()
		perm.Player = ctx.Controller()
		perm.Zone = game.ZoneExile
		ctx.Game.GrantCastPermissionToCardsForEffect(perm, []game.Card{c})
	}
	if CountersOnThisAsItResolves("quest")(ctx) < 5 {
		return nil
	}
	return AddMana{Produced: "{R}{R}"}.Apply(ctx)
})
