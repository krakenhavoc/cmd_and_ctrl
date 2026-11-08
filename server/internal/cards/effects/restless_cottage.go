package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Restless Cottage — Land (Wilds of Eldraine, #4471):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}.
//	 {2}{B}{G}: This land becomes a 4/4 black and green Horror creature
//	 until end of turn. It's still a land.
//	 Whenever this land attacks, create a Food token and exile up to one
//	 target card from a graveyard."
//
// Restless Ridgeline's shape: the animation is one effect at one
// timestamp (creature type and Horror subtype in layer 4, black and green
// in layer 5, base 4/4 in layer 7b), pinned to this object (CR 611.2c).
// The attack trigger is the land's own printed ability: the Food comes
// first, then Afterthought Sentry's "up to one target card from a
// graveyard" is exiled. With no card chosen it still makes the Food; a
// chosen card that left its graveyard by resolution is not exiled, and
// since it was the only target the trigger does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e16595f-bdeb-422e-b99a-bfc0ed52e9f8",
		Name:         "Restless Cottage",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|G}",
			Label:    "Add {B} or {G}",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Restless Cottage — create a Food token and exile up to one target card from a graveyard",
				Do(CreateToken{Template: FoodToken(), N: 1}, ExileChosenTarget{})),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}{G}: This land becomes a 4/4 black and green Horror creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}{B}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Horror")...),
						game.SetColorsMod(ctx.TextColors("B", "G")...),
					}, game.SetBasePTMods(4, 4)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Restless Cottage — a 4/4 black and green Horror creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
