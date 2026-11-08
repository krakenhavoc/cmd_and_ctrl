package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Salt Road Skirmish — Sorcery {3}{B} (Tarkir: Dragonstorm, #4463):
//
//	"Destroy target creature. Create two 1/1 red Warrior creature
//	 tokens. They gain haste until end of turn. Sacrifice them at the
//	 beginning of the next end step."
//
// The destroy is Doom Blade's. The Warriors are mobilize's 1/1 red
// Warrior, read through the card's text (CR 612.2), and the rest is
// Old Hob, Alleycat Blues' keyed tail over the tokens the creation
// really made: haste that ends with the turn (CR 611.2a) and one delayed
// trigger that sacrifices exactly those tokens at the beginning of the
// next end step (CR 603.7). The tokens are made whether or not the
// creature was destroyed (an indestructible target), since nothing in the
// text depends on it. With its only target gone the spell does not
// resolve, so nothing is made (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6e72f4e-d6e0-4537-93c1-7045385c8401",
		Name:         "Salt Road Skirmish",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: item.Controller,
				Source:     item.SourceCardID,
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("1/1 red Warrior")), Count: 2}},
			}, ctx.ItemTail(saltRoadSkirmishWarriors))
		},
	})
}

// saltRoadSkirmishWarriors is "They gain haste until end of turn.
// Sacrifice them at the beginning of the next end step." over the tokens
// the creation made. An on-disk key: never renamed, never reused.
var saltRoadSkirmishWarriors = CardTail("salt-road-skirmish/warriors-haste-and-sacrifice", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
			Label: "Salt Road Skirmish — haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEnd,
		Label: "Salt Road Skirmish — sacrifice the Warriors",
		Cards: o.IDs,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
