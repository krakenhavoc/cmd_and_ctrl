package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jerrenOracle is shared by both faces: the front registers under it, the
// back under jerrenOracle + "#1".
const JerrenOracle = "20d0df30-013c-47c0-b70b-d628d427d30b"

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
	effects.Register(effects.Spec{
		OracleID:     JerrenOracle,
		Name:         "Jerren, Corrupted Bishop",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventLTB}, jerrenEntersOrAnotherHumanDied,
				"Jerren — you lose 1 life and create a 1/1 white Human creature token",
				effects.Do(effects.GainLife{Amount: -1}, effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1})),
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && jerrenHasExactly13Life(g, source.Controller)
			}, "Jerren — you may pay {4}{B}{B}; if you do, transform Jerren", jerrenEndStep),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target Human you control gains lifelink until end of turn.",
			Cost:    effects.ManaCost("{2}"),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink")),
			Targets: effects.TargetCreature("target Human you control", effects.YouControl(), effects.Subtype("Human")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"lifelink"},
					Label: "Jerren — lifelink until end of turn"}.Apply(ctx)
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        JerrenOracle + "#1",
		Name:            "Ormendahl, the Corrupter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: Draw a card.",
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
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
	dead, ok := effects.DiedCreature(ev, g)
	return ok && !effects.IsToken(dead) && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
		effects.LeftAsSubtype(ev, dead, source.TextSubtype("Human"))
}

// jerrenHasExactly13Life is the end-step ability's intervening "if".
func jerrenHasExactly13Life(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Life == 13
}

// jerrenEndStep re-checks the "if" as the ability resolves and offers the
// payment.
func jerrenEndStep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !jerrenHasExactly13Life(g, item.Controller) {
		return nil
	}
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{4}{B}{B}",
		Question: "Jerren — pay {4}{B}{B} to transform Jerren?",
		Then:     jerrenTransformPaid,
	}.Apply(ctx)
}

var jerrenTransformPaid = effects.PayUnlessBranches("pay-unless/jerren-transform",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.TransformThis{}.Apply(ctx)
	}, nil)
