package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Underdark Rift — Land (The List, S82):
//
//	"{T}: Add {C}.
//	 {5}, {T}, Exile this land: Roll a d10. Put target artifact, creature,
//	 or planeswalker into its owner's library just beneath the top X cards
//	 of that library, where X is the result. Activate only as a sorcery."
//
// The exile is part of the cost. The d10 goes through the one dice door and
// the target is carried to the tail as data; a target that has left the
// battlefield (or is no longer legal) by the time the ability resolves
// stops the whole ability, the roll included (CR 608.2b). "Just beneath the
// top X cards" is X + 1 cards down (TuckOptions.Depth), and a library of
// fewer than X cards takes it on the bottom (CR 401.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "199c7604-6b4b-4cec-b6c2-b6b2918b11c8",
		Name:         "Underdark Rift",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}, {T}, Exile this land: Roll a d10. Put target artifact, creature, or planeswalker into its owner's library just beneath the top X cards of that library. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.ExileThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target artifact, creature, or planeswalker", effects.Or(effects.Artifact(), effects.Creature(), effects.Planeswalker())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.RollDice{Sides: 10, N: 1, Then: underdarkRiftRolled, TailCarry: []uuid.UUID{t.ID}}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}

// underdarkRiftRolled is the rest of the ability, after the d10. An on-disk key.
var underdarkRiftRolled = effects.CardTail("dice/underdark-rift-rolled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	id := o.Tail.Carry[0]
	if !effects.OnBattlefield(ctx.Game, id) {
		return nil
	}
	return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{Depth: o.Rolls[0] + 1}, game.KeyedTail{})
})
