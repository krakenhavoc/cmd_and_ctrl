package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// theFantasticFourLabel is the trigger's stack label, and with it the
// key its "hasn't been chosen this turn" memory is kept under.
const theFantasticFourLabel = "The Fantastic Four — enters or a spell with a 4"

// The Fantastic Four — Legendary Creature — Human Hero {R}{G}{W}{U},
// 4/4:
//
//	"When The Fantastic Four enter and whenever you cast a spell with
//	 power, toughness, or mana value 4, choose one that hasn't been
//	 chosen this turn.
//	 • Create a 0/4 colorless Wall creature token with defender.
//	 • The Fantastic Four deal 3 damage to each opponent.
//	 • Put two +1/+1 counters on target creature.
//	 • Draw a card."
//
// One printed ability with two trigger conditions, so it is ONE
// triggered ability watching two event kinds (OnAny) — which matters
// here, because both conditions share one "hasn't been chosen this
// turn" memory (ADR 0097). The entry and three more fours in a turn
// use all four bullets; a fifth trigger that turn is removed with no
// effect.
//
// "A spell with power, toughness, or mana value 4" reads the spell as
// it is on the stack: its mana value counts X (CR 202.3e), and power
// and toughness are read for a creature or Vehicle spell, the spells
// that have them. The Fantastic Four is a 4/4 with mana value 4, but
// casting it does not trigger it — it is not on the battlefield yet;
// its entry does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2281c99-3f45-441b-8e78-7f1f29bf1dcd",
		Name:         "The Fantastic Four",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			theFantasticFourTrigger(),
		},
	})
}

func theFantasticFourTrigger() game.TriggeredAbility {
	castAFour := YouCast(spellWithAFour())
	t := OnAny([]game.EventKind{game.EventETB, game.EventCast},
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventETB {
				return ev.CardID == source.InstanceID
			}
			return castAFour(ev, source, lki, g)
		}, theFantasticFourLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosenThisTurn(
		ModeDoing("Create a 0/4 colorless Wall creature token with defender.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("0/4 colorless Wall with defender"), N: 1}.Apply(ctx)
			}),
		ModeDoing("The Fantastic Four deal 3 damage to each opponent.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return damageToEachOpponent(ctx.Game, item, 3)
			}),
		ModeDoing("Put two +1/+1 counters on target creature.", TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}),
		ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
	)
	return t
}

// spellWithAFour is "a spell with power, toughness, or mana value 4":
// the mana value as the spell has it on the stack, and power and
// toughness for a spell that has them (a creature or a Vehicle).
func spellWithAFour() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		if mv, ok := g.ManaValueForEffect(c); ok && mv == 4 {
			return true
		}
		if !c.IsCreature() && !c.HasSubtype("Vehicle") {
			return false
		}
		eff := c.Effective()
		return eff.Power == 4 || eff.Toughness == 4
	}
}
