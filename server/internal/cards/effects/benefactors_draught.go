package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benefactor's Draught — Instant {1}{G}:
//
//	"Untap all creatures. Until end of turn, whenever a creature an
//	 opponent controls blocks, draw a card.
//	 Draw a card."
//
// The second sentence is CR 603.7b's repeating delayed trigger (#2169),
// keyed on EventBlock: "blocks" is per blocker (CR 509.3a), so a creature
// blocking two attackers draws once, and the trigger fires for each
// opponent creature that blocks until cleanup. "An opponent" is read
// against the caster at the time the creature blocks. The untap is every
// creature on the battlefield, whoever controls it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85113e73-26f7-4938-b7a7-607e704af0ca",
		Name:         "Benefactor's Draught",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var all []game.Card
			all = ctx.Game.BattlefieldCardsForEffect()
			for _, c := range all {
				if !c.IsCreature() {
					continue
				}
				if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			if err := (DelayedOnEvent{
				Label:     "Benefactor's Draught — draw a card",
				On:        []game.EventKind{game.EventBlock},
				Condition: anOpponentsCreatureBlocksCondition,
				Body:      drawOneBody,
				Repeats:   true,
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}

// "Whenever a creature an opponent controls blocks": an EventBlock whose
// blocking creature's controller (Actor) is not the trigger's controller,
// once per blocker however many attackers it blocks (CR 509.3a).
var anOpponentsCreatureBlocksCondition = game.DelayedCondition("block/an-opponents-creature-blocks", anOpponentsCreatureBlocks)

func anOpponentsCreatureBlocks(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return ev.Kind == game.EventBlock && ev.Amount <= 1 && ev.Actor != dt.Controller
}
