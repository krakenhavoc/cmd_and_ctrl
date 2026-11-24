package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Efreet Flamepainter — Creature — Efreet Shaman {3}{R}, 1/4 (Strixhaven:
// School of Mages, #5502):
//
//	"Double strike
//	 Whenever this creature deals combat damage to a player, you may cast
//	 target instant or sorcery card from your graveyard without paying its mana
//	 cost. If that spell would be put into your graveyard, exile it instead."
//
// The target card is chosen as the trigger goes on the stack (CR 603.3d) and
// checked again as it resolves (CR 608.2b). The cast is made during the
// resolution (CR 608.2g), ignoring timing, through the shared free graveyard
// cast (Cmdr2S42FreeGraveyardCast: a {0}, cast-only permission and the
// resolution_cast prompt, "you may" being the prompt's decline). The spell is
// exiled instead of going to a graveyard, as it resolves, is countered or
// fizzles. Double strike means the trigger can fire twice in one combat, once
// for each damage step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb5e8b8c-bad2-45bb-abb4-2ed454525749",
		Name:            "Efreet Flamepainter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Efreet Flamepainter — you may cast target instant or sorcery card from your graveyard without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					effects.Cmdr2S42FreeGraveyardCast(ctx, "Efreet Flamepainter", effects.FreeInstantOrSorcery(), true, []uuid.UUID{id})
					return nil
				}),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()))),
		},
	})
}
