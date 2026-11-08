package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Restless Vinestalk — Land (Wilds of Eldraine, #4471):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}.
//	 {3}{G}{U}: Until end of turn, this land becomes a 5/5 green and blue
//	 Plant creature with trample. It's still a land.
//	 Whenever this land attacks, up to one other target creature has base
//	 power and toughness 3/3 until end of turn."
//
// Restless Vents' shape: the animation is one effect at one timestamp
// (creature type and Plant subtype in layer 4, green and blue in layer 5,
// trample in layer 6, base 5/5 in layer 7b), pinned to this object (CR
// 611.2c). Trample is part of the animation's text, so it ends with the
// effect. The attack trigger is the land's own printed ability: "up to
// one other target creature", anyone's but the Vinestalk's, gets Atomic
// Microsizer's layer 7b base power and toughness until end of turn, so
// its counters and pumps still apply on top (CR 613.4b, 613.4c). A pick
// gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Restless Vinestalk — up to one other target creature has base power and toughness 3/3 until end of turn"
	attack := WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return untilEndOfTurn(ctx, target, nil, label, game.SetBasePTMods(3, 3)...)
	})
	attack.Targets = Another(UpToOneTargetCreature("up to one other target creature"))
	Register(Spec{
		OracleID:     "0935faa2-fb90-48db-8a92-906ba0f374c7",
		Name:         "Restless Vinestalk",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
		Triggered: []game.TriggeredAbility{attack},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}{U}: Until end of turn, this land becomes a 5/5 green and blue Plant creature with trample. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{3}{G}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Plant")...),
						game.SetColorsMod(ctx.TextColors("G", "U")...),
						game.AddKeywordsMod("trample"),
					}, game.SetBasePTMods(5, 5)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Restless Vinestalk — a 5/5 green and blue Plant creature with trample until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
