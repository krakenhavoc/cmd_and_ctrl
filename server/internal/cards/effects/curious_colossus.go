package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curious Colossus — Creature — Giant Warrior {5}{W}{W}, 7/7 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, each creature target opponent controls
//	 loses all abilities, becomes a Coward in addition to its other types,
//	 and has base power and toughness 1/1."
//
// One continuous effect on a set fixed as the ability resolves (CR 611.2c):
// the creatures the target opponent controls then. It lasts indefinitely
// (it does not end at the cleanup step or when the Colossus leaves), and a
// creature that enters later, or leaves and returns as a new object (CR
// 400.7), is not affected. Its three parts are Mods in their own layers, all
// at one timestamp: losing abilities (layer 6), the added subtype (layer 4)
// and the base 1/1 (layer 7b). Counters, +N/+N effects and a P/T switch
// still apply on top, and a later set-P/T effect overwrites the 1/1; an
// ability gained afterwards is kept. "Coward" is a word of the card's text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8de2e1b-f2fa-4b6a-a3c5-9811c7f5f051",
		Name:         "Curious Colossus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Curious Colossus — each creature target opponent controls loses all abilities, becomes a Coward in addition to its other types, and has base power and toughness 1/1",
					curiousColossusShrink),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

func curiousColossusShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
			break
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	var those []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == opp && c.IsCreature() {
			those = append(those, c.InstanceID)
		}
	}
	if len(those) == 0 {
		return nil
	}
	return ScopedEffectFor{
		Match: b10OneOf(those),
		Mods: []game.Mod{
			game.LoseAllAbilitiesMod(),
			game.AddSubtypesMod(ctx.TextSubtypes("Coward")...),
			game.SetBasePowerMod(1),
			game.SetBaseToughnessMod(1),
		},
		Duration: game.IndefiniteDuration(),
		Label:    "Curious Colossus — loses all abilities, is a Coward, base power and toughness 1/1",
	}.Apply(ctx)
}
