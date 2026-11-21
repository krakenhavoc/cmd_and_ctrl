package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Flood of Mars — Creature — Alien Zombie Horror {2}{U}{U}, 3/3 (Doctor
// Who, #5948):
//
//	"Islandwalk (This creature can't be blocked as long as defending player
//	 controls an Island.)
//	 Water Always Wins — Whenever this creature attacks, put a flood counter
//	 on another target creature or land. If it's a creature, it becomes a
//	 copy of this creature. If it's a land, it becomes an Island in addition
//	 to its other types."
//
// Islandwalk is a printed keyword. The trigger puts a flood counter on the
// target and then applies the two "if it's a ..." clauses to what it is as
// the ability resolves: a creature becomes a copy of The Flood of Mars (a
// copy effect with no duration, CR 611.2a, CR 707.2), a land becomes an
// Island in addition to its other types (CR 205.1b, no duration). A creature
// land (a Dryad Arbor) gets both. "Water Always Wins" is an ability word
// (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "677bd655-b240-4862-912c-29209fb295e1",
		Name:            "The Flood of Mars",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("The Flood of Mars — put a flood counter on another target creature or land; a creature becomes a copy of this creature, a land becomes an Island",
					floodOfMarsAttack),
				effects.Another(effects.TargetPermanent("another target creature or land", effects.Or(effects.Creature(), effects.Land())))),
		},
	})
}

func floodOfMarsAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: id, Kind: "flood", N: 1}).Apply(ctx); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || !effects.OnBattlefield(g, id) {
		return nil
	}
	if c.IsCreature() {
		if err := (effects.BecomeCopy{
			Targets:  []uuid.UUID{id},
			Of:       item.SourceCardID,
			Duration: effects.CopyIndefinite,
			Label:    "The Flood of Mars — becomes a copy of this creature",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if c.IsLand() {
		return effects.LandBecomes{
			Target:     id,
			Types:      []string{"Island"},
			InAddition: true,
			Duration:   g.PinnedTo(game.IndefiniteDuration(), id),
			Label:      "The Flood of Mars — becomes an Island in addition to its other types",
		}.Apply(ctx)
	}
	return nil
}
