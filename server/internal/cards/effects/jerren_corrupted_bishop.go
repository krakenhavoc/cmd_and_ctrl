package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jerrenOracle is shared by both faces: the front registers under it, the
// back under jerrenOracle + "#1".
const jerrenOracle = "20d0df30-013c-47c0-b70b-d628d427d30b"

// Jerren, Corrupted Bishop // Ormendahl, the Corrupter — a transforming
// creature (Midnight Hunt):
//
//	Jerren, Corrupted Bishop {2}{B}, Legendary Creature — Human Cleric, 2/3
//	  "Whenever Jerren enters or another nontoken Human you control dies,
//	   you lose 1 life and create a 1/1 white Human creature token.
//	   {2}: Target Human you control gains lifelink until end of turn.
//	   At the beginning of your end step, if you have exactly 13 life, you
//	   may pay {4}{B}{B}. If you do, transform Jerren."
//
//	Ormendahl, the Corrupter, Legendary Creature — Demon, 6/6
//	  "Flying, trample, lifelink
//	   Sacrifice another creature: Draw a card."
//
// The first ability is one trigger on two events: Jerren's own entry, or
// another nontoken Human its controller controlled dying, read from the
// creature's last-known information (CR 603.10a). The end-step ability's
// "if" is an intervening one (CR 603.4): you have exactly 13 life as the
// step begins and again as it resolves, before you are asked about the
// payment. Paying is optional and the transform is the paid branch, so a
// Jerren that turned over or left in the meantime does nothing
// (CR 701.27f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     jerrenOracle,
		Name:         "Jerren, Corrupted Bishop",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventLTB}, jerrenEntersOrAnotherHumanDied,
				"Jerren — you lose 1 life and create a 1/1 white Human creature token",
				Do(GainLife{Amount: -1}, CreateToken{Template: TokenCard("1/1 white Human"), N: 1})),
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, lki, g) && jerrenHasExactly13Life(g, source.Controller)
			}, "Jerren — you may pay {4}{B}{B}; if you do, transform Jerren", jerrenEndStep),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}: Target Human you control gains lifelink until end of turn.",
			Cost:    ManaCost("{2}"),
			Targets: TargetCreature("target Human you control", YouControl(), Subtype("Human")),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{Target: target, Keywords: []string{"lifelink"},
					Label: "Jerren — lifelink until end of turn"}.Apply(ctx)
			},
		}},
	})
	Register(Spec{
		OracleID:        jerrenOracle + "#1",
		Name:            "Ormendahl, the Corrupter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another creature: Draw a card.",
			Cost:    SacrificeAnotherN(1, "another creature", Creature()),
			Purpose: game.Purpose{Draws: 1},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}

// jerrenEntersOrAnotherHumanDied is Jerren's first trigger condition: this
// creature entered, or another nontoken Human its controller controlled
// died.
func jerrenEntersOrAnotherHumanDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind == game.EventETB {
		return ev.CardID == source.InstanceID
	}
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == source.Controller &&
		leftAsSubtype(ev, dead, "Human")
}

// jerrenHasExactly13Life is the end-step ability's intervening "if".
func jerrenHasExactly13Life(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Life == 13
}

// jerrenEndStep re-checks the "if" as the ability resolves and offers the
// payment.
func jerrenEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !jerrenHasExactly13Life(g, item.Controller) {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{4}{B}{B}",
		Question: "Jerren — pay {4}{B}{B} to transform Jerren?",
		Then:     jerrenTransformPaid,
	}.Apply(ctx)
}

var jerrenTransformPaid = PayUnlessBranches("pay-unless/jerren-transform",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return TransformThis{}.Apply(ctx)
	}, nil)
