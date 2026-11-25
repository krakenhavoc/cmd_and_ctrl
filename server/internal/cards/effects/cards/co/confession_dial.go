package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Confession Dial — Artifact {3} (Doctor Who):
//
//	"When this artifact enters, surveil 3.
//	 {T}: Target legendary creature card in your graveyard gains escape
//	 until end of turn. The escape cost is equal to its mana cost plus
//	 exile three other cards from your graveyard. (You may cast it from
//	 your graveyard for its escape cost this turn.)"
//
// The grant is Desdemona, Freedom's Edge's: a cast permission over the one
// card, claimed under the shared "escape" key (CR 702.138a, ADR 0066), at
// its mana cost plus three other cards exiled from the graveyard. A
// permission an effect grants lasts until end of turn (CR 611.2a), and a
// card that leaves the graveyard is a new object it no longer names
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6316158c-2f66-4dc4-b78c-c6ff7c39bb77",
		Name:         "Confession Dial",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Confession Dial — surveil 3", effects.Do(effects.Surveil{N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Target legendary creature card in your graveyard gains escape until end of turn. " +
				"The escape cost is equal to its mana cost plus exile three other cards from your graveyard.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCardInGraveyard("target legendary creature card in your graveyard", effects.Creature(), effects.Legendary(), effects.YouOwn()),
			Effect:  confessionDialGrantEscape,
		}},
	})
}

func confessionDialGrantEscape(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player:                  ctx.Controller(),
		Zone:                    game.ZoneGraveyard,
		AltCostKey:              "escape",
		ExileOtherFromGraveyard: 3,
		Source:                  ctx.Source(),
		Label:                   "Escape — its mana cost, Exile three other cards from your graveyard (Confession Dial)",
	})
	return nil
}
