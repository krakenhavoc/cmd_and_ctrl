package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oakhollow Village — Land (Bloomburrow, #4467):
//
//	"{T}: Add {C}.
//	 {T}: Add {G}. Spend this mana only to cast a creature spell.
//	 {G}, {T}: Put a +1/+1 counter on each Frog, Rabbit, Raccoon, or
//	 Squirrel you control that entered the battlefield this turn."
//
// Bloomburrow's green creature-type land, built like Lilypad Village.
// The counter ability picks its permanents as it resolves: each one you
// control with one of the four subtypes whose current object entered
// this turn (Game.EnteredThisTurn, the per-object entry tally). So a
// creature that entered during another player's turn does not count,
// and one that entered, left and came back counts as the new object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "177b7fe4-8565-4631-b4d9-8b2b4282f3ac",
		Name:         "Oakhollow Village",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G} (creature spells only)",
				Restrictions: []string{
					effects.ManaRestrictCast,
					effects.ManaRestrictType("Creature"),
				},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Put a +1/+1 counter on each Frog, Rabbit, Raccoon, or Squirrel you control that entered the battlefield this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Effect:  oakhollowVillageCounters,
		}},
	})
}

// oakhollowVillageCounters puts one +1/+1 counter on each matching
// permanent, the set chosen once as the ability resolves.
func oakhollowVillageCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	match := effects.And(effects.AnySubtype("Frog", "Rabbit", "Raccoon", "Squirrel"), effects.YouControl(),
		func(g *game.Game, _ uuid.UUID, c game.Card) bool { return g.EnteredThisTurn(c.InstanceID) })
	var ids []uuid.UUID
	for _, c := range effects.MatchingBattlefield(ctx, match) {
		ids = append(ids, c.InstanceID)
	}
	for _, id := range ids {
		if !effects.OnBattlefield(g, id) {
			continue
		}
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
