package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zul Ashur, Lich Lord — Legendary Creature — Zombie Warlock {1}{B},
// 2/2:
//
//	"Ward—Pay 2 life.
//	 {T}: You may cast target Zombie creature card from your graveyard
//	 this turn."
//
// The ward is the life-paying ward (WardLife). The activation is
// Havengul Lich's permission without the borrow: a cast-only grant on
// the one card, for this turn, under the activator (CR 601.2). The
// spell's own costs are paid as usual; nothing exiles it afterwards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7c6c9b3-91f2-4ecf-b2b0-b0edd96974d9",
		Name:         "Zul Ashur, Lich Lord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardLife(2), "Zul Ashur, Lich Lord — ward, pay 2 life"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: You may cast target Zombie creature card from your graveyard this turn.",
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target Zombie creature card from your graveyard", Creature(), HasSubtype("Zombie"), YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					card, ok := g.LookupCardForEffect(t.ID)
					if !ok {
						return nil
					}
					g.GrantCastPermissionToCardsForEffect(game.CastPermission{
						Player:     item.Controller,
						Zone:       game.ZoneGraveyard,
						CastOnly:   true,
						Source:     item.SourceCardID,
						SourceName: "Zul Ashur, Lich Lord",
						Label:      "Zul Ashur, Lich Lord — cast " + card.Name + " this turn",
					}, []game.Card{card})
					return nil
				}
				return nil
			},
		}},
	})
}
