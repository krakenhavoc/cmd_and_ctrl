package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Robot Domination — Enchantment — Plan {3}{B} (Marvel Super Heroes, #4455):
//
//	"Whenever one or more creature cards are put into your graveyard from
//	 anywhere, you draw a card, lose 1 life, and put a plan counter on this
//	 enchantment.
//	 When the third plan counter is put on this enchantment, sacrifice it and
//	 create three 2/2 colorless Robot Villain artifact creature tokens."
//
// Dennick's "creature card put into a graveyard from anywhere" watch (a
// death, a discard, a mill, a countered spell), narrowed to the source
// controller's own graveyard and collapsed to once per batch. The third
// counter's trigger is one instruction after another, not a reflexive: the
// tokens come even when the enchantment is already gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f69d2f9-c16a-4752-9026-794a29d73af8",
		Name:         "Robot Domination",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(
				[]game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if !creatureCardPutIntoAGraveyard(ev, g) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.Owner == source.Controller
				},
				"Robot Domination — you draw a card, lose 1 life, and put a plan counter on this enchantment",
				mshB7RobotDominationPlan)),
			On(game.EventCounterPlaced, mshB2PlanCounterReached(3),
				"Robot Domination — sacrifice it and create three 2/2 colorless Robot Villain artifact creature tokens",
				mshB7RobotDominationFinish),
		},
	})
}

func mshB7RobotDominationPlan(g *game.Game, item *game.StackItem) error {
	if err := b36DrawAndLoseOne(g, item); err != nil {
		return err
	}
	return AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(NewContext(g, item))
}

func mshB7RobotDominationFinish(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if onBattlefield(g, item.SourceCardID) && !ctx.isNewSourceObjectAsThis(item.SourceCardID) {
		if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
	}
	return CreateToken{Template: TokenCard("2/2 colorless Robot Villain artifact"), N: 3}.Apply(ctx)
}
