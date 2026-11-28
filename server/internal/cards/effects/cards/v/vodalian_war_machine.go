package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vodalian War Machine — Creature — Wall {1}{U}{U}, 0/4 (Fallen Empires, S82 OLD
// b854, #5637):
//
//	"Defender
//	 Tap an untapped Merfolk you control: This creature can attack this turn as
//	 though it didn't have defender.
//	 Tap an untapped Merfolk you control: This creature gets +2/+1 until end of
//	 turn.
//	 When this creature dies, destroy all Merfolk tapped this turn to pay for its
//	 abilities."
//
// The Merfolk are tapped as a cost, which is not the {T} symbol (CR 302.6), and
// the turn tally (TurnTally.TappedToPay, #6217) remembers which Merfolk paid for
// which object's abilities. The dies trigger reads it for the War Machine as it
// last existed (CR 603.10a, 608.2h); a Merfolk that has left, or come back as a
// new object, or is no longer a Merfolk is not destroyed.
//
// No simplification.
func init() {
	merfolk := func() game.AbilityCost {
		return effects.TapAnUntapped("an untapped Merfolk you control", effects.HasSubtype("Merfolk"))
	}
	effects.Register(effects.Spec{
		OracleID:        "f72511b0-9993-4998-b646-65a855cb8170",
		Name:            "Vodalian War Machine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Tap an untapped Merfolk you control: This creature can attack this turn as though it didn't have defender.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    merfolk(),
				Effect:  effects.CanAttackThisTurnAsThoughNoDefender("Vodalian War Machine — can attack this turn as though it didn't have defender"),
			},
			{
				Label:   "Tap an untapped Merfolk you control: This creature gets +2/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    merfolk(),
				Effect:  effects.ThisGetsUntilEndOfTurn(2, 1, "Vodalian War Machine — +2/+1 until end of turn"),
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Vodalian War Machine — destroy all Merfolk tapped this turn to pay for its abilities", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				self, ok := g.PermanentRefForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				for _, ref := range g.TappedToPaySourceThisTurnForEffect(self) {
					info, ok := g.PermanentForEffect(ref)
					if !ok || info.Left || !effects.S82OldB854HasSubtype(info, "Merfolk") {
						continue
					}
					if err := (effects.DestroyTarget{Target: ref.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
