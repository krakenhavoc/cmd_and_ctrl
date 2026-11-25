package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dire Fleet Poisoner — Creature — Human Pirate {1}{B}, 2/2:
//
//	"Flash
//	 Deathtouch
//	 When this creature enters, target attacking Pirate you control gets +1/+1
//	 and gains deathtouch until end of turn."
//
// The target is an attacking creature you control with the Pirate subtype
// (a changeling counts), chosen as the trigger goes on the stack (CR 603.3d);
// the pump and the keyword are one until-end-of-turn effect on it. The Poisoner
// itself may be the target when it is an attacking Pirate (it flashes in
// attacking only through other effects).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8150cc84-3f9f-470d-a0f9-ee6d9dab13da",
		Name:            "Dire Fleet Poisoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Dire Fleet Poisoner — target attacking Pirate you control gets +1/+1 and gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, id := range effects.LegalTargetCardIDs(ctx) {
						if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Dire Fleet Poisoner — +1/+1"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Dire Fleet Poisoner — deathtouch"}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetCreature("target attacking Pirate you control", effects.AttackingCreature(), effects.YouControl(), effects.Subtype("Pirate"))),
		},
	})
}
