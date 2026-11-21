package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boiling Rock Rioter — Creature — Human Rogue Ally {2}{B}, 3/3:
//
//	"Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 Tap an untapped Ally you control: Exile target card from a
//	 graveyard.
//	 Whenever this creature attacks, you may cast an Ally spell from among
//	 cards you own exiled with this creature."
//
// Firebending is Avatar Roku's (`firebending`). The exile's cost is
// Honor-Worn Shaku's `TapAnUntapped`: not the {T} symbol, so the Rioter
// itself or an Ally that arrived this turn may pay it (the ruling). The
// exile is linked to this object (CR 406.6, 607.2a).
//
// The second attack trigger is Summon: Esper Valigarmanda's paid cast
// during resolution (CR 608.2g): one cast, from among the cards exiled
// with THIS object that you own and that are Ally cards, paying the
// spell's costs (timing is no bar, since it is cast as the ability
// resolves). Declining leaves the cards in exile, still linked, for the
// next attack. A Rioter that left and came back is a new object with no
// exiled cards (CR 400.7).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5193172c-9024-409e-bd9e-0387971d65fe",
		Name:         "Boiling Rock Rioter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Boiling Rock Rioter — firebending 1", effects.Do(effects.NewFirebending(1))),
			effects.WheneverThisAttacks("Boiling Rock Rioter — you may cast an Ally spell from among cards you own exiled with it",
				boilingRockRioterCast),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped Ally you control: Exile target card from a graveyard.",
			Cost:    effects.TapAnUntapped("an untapped Ally you control", effects.OfCreatureType("Ally")),
			Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
			Effect:  effects.AvB7ExileTargetLinked,
		}},
	})
}

const boilingRockRioterName = "Boiling Rock Rioter"

func boilingRockRioterCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	ally := ctx.TextSubtype("Ally")
	var cards []game.Card
	ids := []uuid.UUID{}
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.Owner != ctx.Controller() || c.IsLand() || !c.HasSubtype(ally) {
			continue
		}
		cards = append(cards, *c)
		ids = append(ids, c.InstanceID)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneExile,
		Timing:      game.TimingFlash,
		CastOnly:    true,
		CastsLeft:   1,
		Duration:    g.UntilEndOfTurnDuration(),
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  boilingRockRioterName,
		Label:       boilingRockRioterName + " — cast an Ally spell from among cards you own exiled with it",
	}, cards)
	effects.OfferResolutionCastOver(g, ctx.Controller(), ctx.Source(), boilingRockRioterName,
		game.MayCastKeywordPaidCast, effects.DeclineLeaveInExile, ids...)
	return nil
}
