package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hourglass of the Lost — Artifact {2}{W}:
//
//	"{T}: Add {W}. Put a time counter on this artifact.
//	 {T}, Remove X time counters from this artifact and exile it: Return
//	 each nonland permanent card with mana value X from your graveyard to
//	 the battlefield. Activate only as a sorcery."
//
// The mana ability is Gemstone Mine's shape with its rider the other way
// round: the {W} is added and a time counter goes on the artifact as part
// of the same mana ability (CR 605.3b, the Rider). The second is an
// activated ability whose X is the number of counters taken off as a cost
// (RemoveCountersXFromThis, X can be 0), and the source is exiled as a cost
// too. It returns every nonland permanent card with mana value exactly X
// from its controller's graveyard together, as one entry (CR 603.6a), each
// under that player's control. Mana value X reads through the engine's
// mana value read (a card whose cost it cannot read never matches).
func init() {
	Register(Spec{
		OracleID:     "bf89bd6d-aa6c-4fd8-9278-151529ab0e62",
		Name:         "Hourglass of the Lost",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "{T}: Add {W}. Put a time counter on this artifact.",
			Rider:    hourglassOfTheLostCounter,
		}},
		Activated: []ActivatedAbility{{
			Label:        "{T}, Remove X time counters from this artifact and exile it: Return each nonland permanent card with mana value X from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         Plus(TapCost(), RemoveCountersXFromThis(game.CounterTime, 0), ExileThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.CountersRemoved()
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil {
					return nil
				}
				var ids []uuid.UUID
				for _, c := range p.Graveyard.Cards {
					if c.IsToken() || !c.IsPermanent() || c.IsLand() {
						continue
					}
					if mv, ok := g.ManaValueForEffect(c); ok && mv == x {
						ids = append(ids, c.InstanceID)
					}
				}
				if len(ids) == 0 {
					return nil
				}
				return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, ids, game.ZoneEntryOptions{Controller: item.Controller}, game.TailRef{}, 0)
			},
		}},
	})
}

// hourglassOfTheLostCounter is "Put a time counter on this artifact", the
// rider of the mana ability.
func hourglassOfTheLostCounter(g *game.Game, _, source uuid.UUID) error {
	return g.AddCounterByKeyedForEffect(uuid.Nil, source, game.CounterTime, 1, game.KeyedTail{})
}
