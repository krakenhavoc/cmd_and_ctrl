package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tidus, Yuna's Guardian — Legendary Creature — Human Warrior {G}{W}{U},
// 3/3:
//
//	"At the beginning of combat on your turn, you may move a counter
//	 from target creature you control onto a second target creature you
//	 control.
//	 Cheer — Whenever one or more creatures you control with counters on
//	 them deal combat damage to a player, you may draw a card and
//	 proliferate. Do this only once each turn."
//
// The move is Nesting Grounds' body over two creature clauses, the
// second Distinct: one counter, of the kind the controller picks when
// the first creature carries several, put on the second and then taken
// off the first only if it landed (CR 122.5). The "you may" is the
// trigger's own prompt.
//
// Cheer is one trigger per player your countered creatures connect with
// (CR 603.2c), whether the creature had a counter read as the damage is
// dealt. "Do this only once each turn" is Nykthos Paragon's tally
// (b28DoneThisTurn): the first Cheer the controller says "yes" to as it
// resolves uses up the turn, and declining one does not. The draw comes
// first, then the proliferate, which the controller
// chooses (#2525).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8fbcafd-3617-4cd7-b55d-457fb7b60162",
		Name:         "Tidus, Yuna's Guardian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(AtBeginningOfYourCombat(
				"Tidus, Yuna's Guardian — move a counter from target creature you control onto a second target creature you control",
				nestingGroundsMove),
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					Distinct(TargetCreature("a second target creature you control", YouControl())),
				)),
				"Tidus, Yuna's Guardian — move a counter between two creatures you control?"),
			tidusCheer(),
		},
	})
}

// tidusCheerLabel is Cheer's stack label.
const tidusCheerLabel = "Cheer — draw a card and proliferate"

// tidusCheerDone is Cheer's once-each-turn tally label (b28DoneThisTurn).
const tidusCheerDone = "Cheer — drew and proliferated this turn"

func tidusCheer() game.TriggeredAbility {
	t := WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hasCounterOfKind(c, "") },
		tidusCheerLabel,
		// The proliferate waits for a prompt the draw asks (CR 608.2c, #4155). Once
		// each turn is checked as the ability begins, and the turn's tally is marked
		// before the draw.
		Do(If{
			Cond: CondBoard{Name: "Cheer has not happened this turn", Test: func(ctx *Context) bool {
				return !b28DoneThisTurn(ctx.Game, ctx.Item.SourceCardID, tidusCheerDone)
			}},
			Then: []Applier{
				Instruction(func(ctx *Context) error {
					b28NoteDoneThisTurn(ctx.Game, ctx.Item.SourceCardID, tidusCheerDone)
					return nil
				}),
				DrawCards{N: 1},
				Proliferate{},
			},
		}))
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return when(ev, source, lki, g) &&
			!b28DoneThisTurn(g, source.InstanceID, tidusCheerDone)
	}
	t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Cheer — draw a card and proliferate?"}
	return t
}
