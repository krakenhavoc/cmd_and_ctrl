package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Knight Rampager — Artifact Creature — Knight {4}{R}, 6/5:
//
//	"Trample
//	 Frenzied Rampage — At the beginning of combat on your turn, choose
//	 an opponent at random. This creature attacks that player this
//	 combat if able.
//	 When this creature dies, it deals 4 damage to target opponent
//	 chosen at random."
//
// Frenzied Rampage chooses as it resolves, not as a target: the engine
// draws one of the controller's opponents still in the game through the
// keyed random stream and logs it, then gives the Knight "attacks that
// player this combat if able" (game.ModAttackPlayerIfAble, the encore
// requirement, #3172) until end of combat. It is one requirement among
// the others CR 508.1d maximises, so a goaded Knight must attack but
// may attack anyone when the goader is the chosen player (the
// 2022-10-07 ruling). A Knight that left the battlefield is a new
// object the requirement does not follow.
//
// The death trigger's target is ChosenAtRandom (#1815): drawn from the
// opponents as the trigger goes on the stack, and the damage is dealt
// by the Knight as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2fa681b2-ca94-4f0c-83a4-98f00b8e1af8",
		Name:            "Knight Rampager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Knight Rampager — Frenzied Rampage: choose an opponent at random; this creature attacks that player this combat if able", knightRampagerRampage),
			Targeting(WhenThisDies("Knight Rampager — 4 damage to target opponent chosen at random", sourceDealsDamageToEachLegalTarget(4)),
				ChosenAtRandom(TargetPlayer("target opponent chosen at random", Opponent()))),
		},
	})
}

// knightRampagerRampage is Frenzied Rampage's body.
func knightRampagerRampage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), ctx.Opponents(), 1)
	if len(pick) != 1 {
		return nil
	}
	opp := pick[0]
	name := "an opponent"
	if p := g.PlayerByIDForEffect(opp); p != nil && p.Name != "" {
		name = p.Name
	}
	g.EmitEvent(game.Event{
		Kind:   game.EventOptionChosen,
		Actor:  item.Controller,
		CardID: item.SourceCardID,
		Label:  "Frenzied Rampage: " + name + " (at random)",
	})
	if sourceIsNewObject(g, item) {
		return nil
	}
	d, ok := DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     []game.Mod{game.AttackPlayerIfAbleMod(opp)},
		Duration: d,
		Label:    "Knight Rampager — attacks that player this combat if able",
	}.Apply(ctx)
}
