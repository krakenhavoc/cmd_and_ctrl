package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restless Spire — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U} or {R}.
//	 {U}{R}: Until end of turn, this land becomes a 2/1 blue and red
//	 Elemental creature with 'During your turn, this creature has first
//	 strike.' It's still a land.
//	 Whenever this land attacks, scry 1."
//
// A manland built like Raging Ravine. The quoted first strike is a GRANT
// (ADR 0093, #3808): the animation gives the land the bundle until end of
// turn, and the bundle's layer 6 static reads the turn and whether the
// land is a creature, so an idle land carries no badge and the keyword
// blinks off on an opponent's turn. Being a granted ability and not the
// land's own, an ability-removing effect that began before the activation
// (Humility) leaves it (CR 613.6, 613.7a). The scry trigger is the land's
// own printed ability.
//
// No simplification.
const restlessSpireFirstStrikeGrant = "restless-spire/first-strike"

func init() {
	Register(Spec{
		OracleID:     "0ca4e80e-c19c-4b74-b531-c5a4dc5a8ba9",
		Name:         "Restless Spire",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|R}",
			Label:    "Add {U} or {R}",
		}},
		Grants: []AbilityGrant{{
			Key:    restlessSpireFirstStrikeGrant,
			Static: []game.StaticAbility{rfReprintBFirstStrikeWhileCreatureOnYourTurn()},
			Text:   "During your turn, this creature has first strike.",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Restless Spire — scry 1",
				Do(Scry{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{U}{R}: Until end of turn, this land becomes a 2/1 blue and red Elemental creature with \"During your turn, this creature has first strike.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{U}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod("Elemental"),
						game.SetColorsMod("U", "R"),
						game.SetBasePowerMod(2),
						game.SetBaseToughnessMod(1),
						game.GrantAbilitiesMod(restlessSpireFirstStrikeGrant),
					},
					Duration: DurationUntilEndOfTurn(NewContext(g, item)),
					Label:    "Restless Spire — a 2/1 blue and red Elemental creature until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
