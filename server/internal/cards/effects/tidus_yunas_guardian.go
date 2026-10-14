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
// dealt. "Do this only once each turn" is Nykthos Paragon's pair of
// checks: a Cheer already put on the stack this turn, or one still
// asking, stops another, and declining the "you may" does not use up the
// turn. The draw comes first, then the proliferate, which the controller
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

// tidusCheerLabel is Cheer's stack label, which the once-each-turn tally
// matches on.
const tidusCheerLabel = "Cheer — draw a card and proliferate"

func tidusCheer() game.TriggeredAbility {
	t := WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hasCounterOfKind(c, "") },
		tidusCheerLabel,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
				return err
			}
			return Proliferate{}.Apply(ctx)
		})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return when(ev, source, lki, g) &&
			!b11TriggeredThisTurn(g, source.InstanceID, tidusCheerLabel) &&
			!b28TriggerPromptPendingFrom(g, source)
	}
	t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Cheer — draw a card and proliferate?"}
	return t
}
