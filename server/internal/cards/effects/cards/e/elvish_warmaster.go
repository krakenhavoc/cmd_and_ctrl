package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Warmaster — Creature — Elf Warrior {1}{G}, 2/2 (Tales of Middle-earth
// Commander, #5948):
//
//	"Whenever one or more other Elves you control enter, create a 1/1 green
//	 Elf Warrior creature token. This ability triggers only once each turn.
//	 {5}{G}{G}: Elves you control get +2/+2 and gain deathtouch until end of
//	 turn."
//
// "One or more" is one trigger per batch of entries (OncePerBatch), and
// TriggersOnlyOnceEachTurn spends the turn's single use the first time it goes
// on the stack. The Elves that enter are others (not the Warmaster), under
// your control. The pump is Overrun's two layers over the Elves you control as
// it resolves (CR 611.2c), the Warmaster included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bda83ef4-e345-495d-878c-4da171f997ba",
		Name:         "Elvish Warmaster",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
					return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Elf"))
				},
				"Elvish Warmaster — create a 1/1 green Elf Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1})))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}{G}: Elves you control get +2/+2 and gain deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				elves := effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype(ctx.TextSubtype("Elf")))
				if err := (effects.BoostUntilEOT{Match: elves, Power: 2, Toughness: 2, Label: "Elvish Warmaster — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: elves, Keywords: []string{"deathtouch"}, Label: "Elvish Warmaster — deathtouch"}.Apply(ctx)
			},
		}},
	})
}
