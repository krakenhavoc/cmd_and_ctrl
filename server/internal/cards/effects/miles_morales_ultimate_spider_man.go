package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Miles Morales // Ultimate Spider-Man — a modal double-faced card whose
// front face transforms (Marvel's Spider-Man promos, #4460):
//
//	Miles Morales                       {1}{G}   1/2
//	Legendary Creature — Human Citizen Hero
//	  When Miles Morales enters, put a +1/+1 counter on each of up to
//	  two target creatures.
//	  {3}{R}{G}{W}: Transform Miles Morales. Activate only as a sorcery.
//
//	Ultimate Spider-Man                 {3}{R}{G}{W}   4/3
//	Legendary Creature — Spider Human Hero
//	  First strike, haste
//	  Camouflage — {2}: Put a +1/+1 counter on Ultimate Spider-Man. He
//	  gains hexproof and becomes colorless until end of turn.
//	  Whenever you attack, double the number of each kind of counter on
//	  each Spider and legendary creature you control.
//
// Gwen Stacy // Ghost-Spider's layout: the back face registers under the
// oracle ID plus "#1", and the transform is TransformThis (CR 712.9 lets
// an instruction transform a modal double-faced permanent).
//
// The enter trigger is Agent Bishop's "each of up to two target
// creatures". Camouflage is Bristling Hydra's counter-and-hexproof with
// colorless added to the same until-end-of-turn effect (layer 5 colour,
// layer 6 ability); it does nothing to a Spider-Man that has left. The
// attack trigger is once per declaration and reads the creatures as it
// resolves: each Spider (the word read through text-changing effects)
// or legendary creature you control then has each kind of counter on it
// doubled (b30DoubleEachCounterKindOn), Ultimate Spider-Man included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     milesMoralesOracle,
		Name:         "Miles Morales",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Miles Morales — put a +1/+1 counter on each of up to two target creatures",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{R}{G}{W}: Transform Miles Morales. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{R}{G}{W}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        milesMoralesOracle + "#1",
		Name:            "Ultimate Spider-Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "Camouflage — {2}: Put a +1/+1 counter on Ultimate Spider-Man. He gains hexproof and becomes colorless until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerProtect},
			Cost:    ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := plusOneCountersOnThis(1)(g, item); err != nil {
					return err
				}
				return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
					"Ultimate Spider-Man — hexproof and colorless until end of turn",
					game.AddKeywordsMod("hexproof"), game.SetColorsMod())
			},
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Ultimate Spider-Man — double each kind of counter on each Spider and legendary creature you control",
				Do(Instruction(ultimateSpiderManDoubleCounters)))),
		},
	})
}

// milesMoralesOracle is the card's oracle ID; the back face adds "#1".
const milesMoralesOracle = "be48c721-bdc6-4f2c-96d6-82d934b31615"

// ultimateSpiderManDoubleCounters doubles each kind of counter on each
// Spider and each legendary creature the controller controls now.
func ultimateSpiderManDoubleCounters(ctx *Context) error {
	spider := ctx.TextSubtype("Spider")
	var mine []game.Card
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == ctx.Controller() &&
			(c.HasSubtype(spider) || c.HasSupertype("Legendary")) {
			mine = append(mine, c)
		}
	}
	for _, c := range mine {
		if err := b30DoubleEachCounterKindOn(ctx, c.InstanceID); err != nil {
			return err
		}
	}
	return nil
}
