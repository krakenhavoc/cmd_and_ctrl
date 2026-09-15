package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aftermath Analyst — Creature — Elf Detective {1}{G}, 1/3 (EDHREC
// rank 631):
//
//	"When this creature enters, mill three cards.
//	 {3}{G}, Sacrifice this creature: Return all land cards from your
//	 graveyard to the battlefield tapped."
//
// Self-mill on the way in, Splendid Reclamation on the way out. The
// ETB is an ordinary trigger; the cash-in is a CR 602 activated
// ability whose cost is mana plus sacrificing the source (Commander's
// Sphere's shape with a mana component), and its effect walks the
// controller's graveyard for land cards — snapshotted first, since
// each return removes a card from the pile being ranged over — and
// returns each one.
//
// Sandbox simplification: the lands enter untapped and are tapped a
// beat later (ReturnFromGraveyard has no tapped flag — Victimize's
// posture), so anything counting tap events sees one per land.
// Landfall triggers fire once per land, as printed.
func init() {
	Register(Spec{
		OracleID:     "374e54d8-8e73-4268-9dde-c28c77bbbf32",
		Name:         "Aftermath Analyst",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The returned lands come back untapped and are tapped a moment later, rather than entering tapped."},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aftermath Analyst — mill three cards", Do(MillCards{N: 3})),
		},
		Activated: []ActivatedAbility{{
			Label: "{3}{G}, Sacrifice this creature: Return all land cards from your graveyard to the battlefield tapped.",
			Cost:  Plus(ManaCost("{3}{G}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil || p.Graveyard == nil {
					return nil
				}
				var lands []uuid.UUID
				for _, c := range p.Graveyard.Cards {
					if c.IsLand() {
						lands = append(lands, c.InstanceID)
					}
				}
				for _, id := range lands {
					if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
						return err
					}
					if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
