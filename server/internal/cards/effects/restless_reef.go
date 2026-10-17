package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restless Reef — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U} or {B}.
//	 {2}{U}{B}: Until end of turn, this land becomes a 4/4 blue and black
//	 Shark creature with deathtouch. It's still a land.
//	 Whenever this land attacks, target player mills four cards."
//
// Restless Spire's shape: the animation is one effect at one timestamp
// (creature type and Shark subtype in layer 4, blue and black in layer 5,
// deathtouch in layer 6, base 4/4 in layer 7b), pinned to this object
// (CR 611.2c). Deathtouch is part of the animation's text, not a printed
// keyword of the land, so the effect adds it and it ends with the effect.
// The attack trigger is the land's OWN printed ability: it is on the card
// whether or not the land is a creature, and it can only happen while it
// is one, since a land that is not a creature can't attack. It is
// targeted: the engine prompts for the player when the trigger goes on
// the stack, and the mill resolves against the target if it is still
// legal (CR 608.2b). The target may be the controller.
//
// No simplification.
func init() {
	attackTrigger := WheneverThisAttacks("Restless Reef — target player mills four cards",
		func(g *game.Game, item *game.StackItem) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return MillCards{Player: item.Targets[0].ID, N: 4}.Apply(NewContext(g, item))
		})
	attackTrigger.Targets = TargetPlayer("target player")
	Register(Spec{
		OracleID:     "df20f85b-5f81-4ee4-8487-55d90109ac36",
		Name:         "Restless Reef",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|B}",
			Label:    "Add {U} or {B}",
		}},
		Triggered: []game.TriggeredAbility{attackTrigger},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}{B}: Until end of turn, this land becomes a 4/4 blue and black Shark creature with deathtouch. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate | game.AnswerCombatGrant},
			Cost:    ManaCost("{2}{U}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Shark")...),
						game.SetColorsMod(ctx.TextColors("U", "B")...),
						game.AddKeywordsMod("deathtouch"),
					}, game.SetBasePTMods(4, 4)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Restless Reef — a 4/4 blue and black Shark creature with deathtouch until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
