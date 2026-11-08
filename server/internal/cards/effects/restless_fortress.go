package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restless Fortress — Land (Wilds of Eldraine, #4471):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {B}.
//	 {2}{W}{B}: This land becomes a 1/4 white and black Nightmare creature
//	 until end of turn. It's still a land.
//	 Whenever this land attacks, defending player loses 2 life and you gain
//	 2 life."
//
// Restless Ridgeline's shape: the animation is one effect at one
// timestamp (creature type and Nightmare subtype in layer 4, white and
// black in layer 5, base 1/4 in layer 7b), pinned to this object (CR
// 611.2c). The attack trigger is Agate-Blade Assassin's drain at 2
// (attackDrainNTriggerWhen): the defending player is read off the attack
// as it is declared, so attacking a planeswalker or a battle drains the
// player behind it, and one who has left the game loses nothing while you
// still gain 2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b3726f1-20b8-42ec-8f9b-b361515c3f05",
		Name:         "Restless Fortress",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|B}",
			Label:    "Add {W} or {B}",
		}},
		Triggered: []game.TriggeredAbility{attackDrainNTriggerWhen("Restless Fortress", 2, ThisAttacked)},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}{B}: This land becomes a 1/4 white and black Nightmare creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}{W}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Nightmare")...),
						game.SetColorsMod(ctx.TextColors("W", "B")...),
					}, game.SetBasePTMods(1, 4)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Restless Fortress — a 1/4 white and black Nightmare creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
