package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mishra's Foundry — Land (The Brothers' War, #5482):
//
//	"{T}: Add {C}.
//	 {2}: This land becomes a 2/2 Assembly-Worker artifact creature until
//	 end of turn. It's still a land.
//	 {1}, {T}: Target attacking Assembly-Worker gets +2/+2 until end of
//	 turn."
//
// A manland built like Mutavault: the animation adds the Artifact and
// Creature types and the Assembly-Worker subtype (layer 4) and sets the
// base power and toughness (layer 7b) until end of turn (CR 611.2a). The
// land stays a land with its mana ability. The pump may target the
// Foundry itself once it is an attacking Assembly-Worker, tapping it as
// a cost (an attacking creature that taps stays attacking, CR 506.4).
// The target is checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b43e9772-6ad4-49c7-9557-b18ee1e4587d",
		Name:         "Mishra's Foundry",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}: This land becomes a 2/2 Assembly-Worker artifact creature until end of turn. It's still a land.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					return ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: append([]game.Mod{
							game.AddTypesMod("Artifact", "Creature"),
							game.AddSubtypesMod(ctx.TextSubtypes("Assembly-Worker")...),
						}, game.SetBasePTMods(2, 2)...),
						Duration: DurationUntilEndOfTurn(ctx),
						Label:    "Mishra's Foundry — a 2/2 Assembly-Worker artifact creature until end of turn",
					}.Apply(ctx)
				},
			},
			{
				Label:   "{1}, {T}: Target attacking Assembly-Worker gets +2/+2 until end of turn.",
				Purpose: withAnswers(ForTargets(BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
				Cost:    game.AbilityCost{Mana: "{1}", Tap: true},
				Targets: TargetCreature("target attacking Assembly-Worker", AttackingCreature(), Subtype("Assembly-Worker")),
				Effect:  boostTargetUntilEOT("Mishra's Foundry — +2/+2", 2, 2),
			},
		},
	})
}
