package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Restless Ridgeline — Land (The Lost Caverns of Ixalan, #4470):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}.
//	 {2}{R}{G}: This land becomes a 3/4 red and green Dinosaur creature
//	 until end of turn. It's still a land.
//	 Whenever this land attacks, another target attacking creature gets
//	 +2/+0 until end of turn. Untap that creature."
//
// Restless Reef's shape: the animation is one effect at one timestamp
// (creature type and Dinosaur subtype in layer 4, red and green in layer
// 5, base 3/4 in layer 7b), pinned to this object (CR 611.2c). The attack
// trigger is the land's own printed ability. Its target is another
// attacking creature, chosen as the trigger goes on the stack (CR
// 603.3d, Helitrooper's clause); the land itself can't be chosen. The
// +2/+0 is a layer 7c change until the cleanup step, and the untap
// follows it; a target that is gone by resolution gets neither (CR
// 608.2b).
//
// No simplification.
func init() {
	const label = "Restless Ridgeline — another target attacking creature gets +2/+0 until end of turn; untap it"
	attack := WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		if err := (BoostUntilEOT{Target: target, Power: 2, Label: label}).Apply(ctx); err != nil {
			return err
		}
		return UntapTarget{Target: target}.Apply(ctx)
	})
	attack.Targets = Another(TargetCreature("another target attacking creature", AttackingCreature()))
	Register(Spec{
		OracleID:     "4c0f4a63-586a-4dde-9621-b0dd9118b2e5",
		Name:         "Restless Ridgeline",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Triggered: []game.TriggeredAbility{attack},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}{G}: This land becomes a 3/4 red and green Dinosaur creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Dinosaur")...),
						game.SetColorsMod(ctx.TextColors("R", "G")...),
					}, game.SetBasePTMods(3, 4)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Restless Ridgeline — a 3/4 red and green Dinosaur creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
