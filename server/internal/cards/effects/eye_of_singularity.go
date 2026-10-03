package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eye of Singularity — World Enchantment {3}{W}:
//
//	"When this enchantment enters, destroy each permanent with the same
//	 name as another permanent, except for basic lands. They can't be
//	 regenerated.
//	 Whenever a permanent other than a basic land enters, destroy all
//	 other permanents with that name. They can't be regenerated."
//
// Two destructions, each one simultaneous sweep with regeneration
// ignored (CR 701.19c), through DestroyAllMatching, so indestructible
// permanents survive (CR 702.12b).
//
// "The same name" is the face-up name (CR 201.2): a face-down
// permanent has none (CR 708.2a), so it shares a name with nothing and
// its entry destroys nothing. The second trigger fires for the Eye's
// own entry too — it is a permanent other than a basic land — and
// reads the entering permanent's name as it resolves, from wherever
// the permanent is by then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e13edf92-cdba-4c24-b81b-088b1554fde6",
		Name:         "Eye of Singularity",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eye of Singularity — destroy each permanent that shares a name with another, except basic lands",
				Do(DestroyAllMatching{Match: eyeSharesANameWithAnother, CantBeRegenerated: true})),
			On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !IsBasicLand(c)
			}, "Eye of Singularity — destroy all other permanents with that name", eyeOfSingularityOnEntry),
		},
	})
}

// eyeSharesANameWithAnother is the entry sweep's predicate: a face-up
// permanent other than a basic land that another face-up permanent
// shares its name with.
func eyeSharesANameWithAnother(g *game.Game, _ uuid.UUID, c game.Card) bool {
	if IsBasicLand(c) || c.FaceDownIsPermanent() {
		return false
	}
	for _, other := range g.BattlefieldCardsForEffect() {
		if other.InstanceID != c.InstanceID && !other.FaceDownIsPermanent() && sameFaceUpName(other, c.Name) {
			return true
		}
	}
	return false
}

func eyeOfSingularityOnEntry(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	entered := ctx.Trigger().Event.CardID
	c, ok := g.LookupCardForEffect(entered)
	if !ok || c.FaceDownIsPermanent() || c.FaceDown {
		return nil
	}
	name := c.Name
	return DestroyAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, other game.Card) bool {
			return other.InstanceID != entered && !other.FaceDownIsPermanent() && sameFaceUpName(other, name)
		},
		CantBeRegenerated: true,
	}.Apply(ctx)
}
