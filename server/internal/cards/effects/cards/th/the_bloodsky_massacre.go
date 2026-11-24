package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Bloodsky Massacre — Enchantment — Saga {1}{B}{R} (Kaldheim Promos,
// #5505):
//
//	"I — Create a 2/3 red Demon Berserker creature token with menace.
//	 II — Whenever a Berserker attacks this turn, you draw a card and you
//	 lose 1 life.
//	 III — Add {R} for each Berserker you control. Until end of turn, you
//	 don't lose this mana as steps and phases end."
//
// Chapter II is a repeating delayed trigger (CR 603.7b) on any Berserker its
// controller attacks with this turn, one trigger per attacker (CR 603.2c). It
// draws, then loses 1 life, as one resolution. Chapter III counts the
// Berserkers you control as it resolves (a changeling is one) and adds that
// much {R} carrying the keep-until-end-of-turn rider.
//
// No simplification.
func init() {
	const name = "The Bloodsky Massacre"
	effects.Register(effects.Spec{
		OracleID:     "fd9cd2b0-3f43-41a6-b963-c781d0f285e8",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 2/3 red Demon Berserker creature token with menace"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/3 red Demon Berserker with menace"), N: 1})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "whenever a Berserker attacks this turn, you draw a card and you lose 1 life"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DelayedOnEvent{
						Label:      name + " — you draw a card and you lose 1 life",
						On:         []game.EventKind{game.EventAttack},
						Condition:  effects.GapB41AssassinYouControlAttacks,
						CondParams: game.EffectParams{Name: ctx.TextSubtype("Berserker")},
						Body:       bloodskyMassacreDrawBody,
						Repeats:    true,
					}.Apply(ctx)
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "add {R} for each Berserker you control; you don't lose this mana as steps and phases end until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					word := ctx.TextSubtype("Berserker")
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.HasSubtype(word) {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.AddMana{
						Amount: effects.ManaN(n, "R"),
						Riders: []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()},
					}.Apply(ctx)
				}),
		},
	})
}

// bloodskyMassacreDrawBody is chapter II's delayed trigger: draw, then lose 1
// life. An on-disk key.
var bloodskyMassacreDrawBody = game.SimpleDelayedBody("the-bloodsky-massacre/draw-and-lose-life", effects.B36DrawAndLoseOne)
