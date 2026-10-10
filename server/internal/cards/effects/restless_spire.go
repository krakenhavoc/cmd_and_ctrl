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
// A manland built like Raging Ravine. The quoted first strike is
// declared as a layer 6 static on the land that reads the turn and
// whether the land is a creature, so an idle land carries no badge and
// the keyword blinks off on an opponent's turn. The scry trigger is the
// land's own printed ability.
//
// No simplification.
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
		Static: []game.StaticAbility{rfReprintBFirstStrikeWhileCreatureOnYourTurn()},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Restless Spire — scry 1",
				func(g *game.Game, item *game.StackItem) error {
					return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{U}{R}: Until end of turn, this land becomes a 2/1 blue and red Elemental creature with \"During your turn, this creature has first strike.\" It's still a land.",
			Cost:  ManaCost("{U}{R}"),
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
					},
					Duration: DurationUntilEndOfTurn(NewContext(g, item)),
					Label:    "Restless Spire — a 2/1 blue and red Elemental creature until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
