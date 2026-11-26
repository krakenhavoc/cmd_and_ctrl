package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warden of the First Tree — {G} Creature — Human 1/1 (Fate Reforged):
//
//	"{1}{W/B}: This creature becomes a Human Warrior with base power and
//	 toughness 3/3.
//	 {2}{W/B}{W/B}: If this creature is a Warrior, it becomes a Human Spirit
//	 Warrior with trample and lifelink.
//	 {3}{W/B}{W/B}{W/B}: If this creature is a Spirit, put five +1/+1
//	 counters on it."
//
// Frodo, Sauron's Bane's ladder: each "becomes" is a ScopedEffect with no
// stated duration, so it lasts as long as this is the same object (CR 611.2a,
// 400.7), and its creature types are a set (CR 205.1a), so the Spirit Warrior
// is still a Warrior and the 3/3 base stays. The "if" is checked as the ability
// resolves (CR 608.2c), against the types the creature has then.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0068a341-92a5-470e-a565-9ace69eb04a8",
		Name:         "Warden of the First Tree",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{W/B}: This creature becomes a Human Warrior with base power and toughness 3/3.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
				Cost:    effects.ManaCost("{1}{W/B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					self := item.SourceCardID
					return effects.ScopedEffectFor{
						Target: self,
						Mods: []game.Mod{
							game.SetCreatureTypesMod(g.TextReaderForEffect(item).TextSubtypes("Human", "Warrior")...),
							game.SetBasePowerMod(3),
							game.SetBaseToughnessMod(3),
						},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    "Warden of the First Tree — a 3/3 Human Warrior",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{2}{W/B}{W/B}: If this creature is a Warrior, it becomes a Human Spirit Warrior with trample and lifelink.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{2}{W/B}{W/B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanentWithTextSubtype(g, item, "Warrior") {
						return nil
					}
					self := item.SourceCardID
					return effects.ScopedEffectFor{
						Target: self,
						Mods: []game.Mod{
							game.SetCreatureTypesMod(g.TextReaderForEffect(item).TextSubtypes("Human", "Spirit", "Warrior")...),
							game.AddKeywordsMod("trample", "lifelink"),
						},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    "Warden of the First Tree — a Human Spirit Warrior with trample and lifelink",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{3}{W/B}{W/B}{W/B}: If this creature is a Spirit, put five +1/+1 counters on it.",
				Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 5},
				Cost:    effects.ManaCost("{3}{W/B}{W/B}{W/B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanentWithTextSubtype(g, item, "Spirit") {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 5}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
