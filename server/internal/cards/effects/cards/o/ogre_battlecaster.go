package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ogre Battlecaster — Creature — Ogre Shaman {2}{R}, 3/3 (Commander
// Legends: Battle for Baldur's Gate, #5948 GAP2 b07):
//
//	"First strike
//	 Whenever this creature attacks, you may cast target instant or
//	 sorcery card from your graveyard by paying {R}{R} in addition to its
//	 other costs. If that spell would be put into a graveyard, exile it
//	 instead. When you cast that spell, this creature gets +X/+0 until end
//	 of turn, where X is that spell's mana value."
//
// The target is chosen as the trigger goes on the stack. The cast is made
// during its resolution (CR 608.2g), through the resolution_cast prompt,
// under a permission over that card that pays its mana cost plus {R}{R}
// as an additional cost (CR 601.2f; CastPermission.CostMoreMana, the
// #6091 seam), at any timing, cast only. The spell is exiled instead of
// going to a graveyard (ExileInsteadOfGraveyard, Torrential Gearhulk's
// clause). Declining leaves the card in the graveyard.
//
// "When you cast that spell" is a delayed trigger created by this
// resolution (CR 603.7a), made by the permission's follow-up, which runs
// only for a cast made through it
// (QueueDelayedTriggerForPermissionPlayForEffect). Its X is the spell's
// mana value as it was cast (with any X it was cast with, CR 202.3e), and
// it pumps this creature only while it is the same object (CR 400.7).
//
// No simplification.
func init() {
	const name = "Ogre Battlecaster"
	effects.Register(effects.Spec{
		OracleID:        "00ba0c24-a671-493e-ba46-13e45d1818f1",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks(name+" — you may cast target instant or sorcery card from your graveyard by paying {R}{R} more",
				ogreBattlecasterOffer),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn())),
		},
	})
}

func ogreBattlecasterOffer(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 || legal[0].Kind != game.TargetCard {
		return nil
	}
	ref, _ := g.SourceObjectForEffect(item)
	perm := game.CastPermission{
		Player:                  item.Controller,
		Zone:                    game.ZoneGraveyard,
		Duration:                g.UntilEndOfTurnDuration(),
		AltCostKey:              game.AltCostKeyGrantedCost,
		CostMoreMana:            "{R}{R}",
		Timing:                  game.TimingFlash,
		CastOnly:                true,
		ExileInsteadOfGraveyard: true,
		LapseOnPass:             game.LapseStaysInHand,
		FollowUp:                ogreBattlecasterCast,
		FollowUpObject:          &ref,
		Source:                  ref.ID,
		SourceName:              "Ogre Battlecaster",
		Label:                   "Ogre Battlecaster — cast it by paying {R}{R} in addition to its other costs",
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      item.Controller,
		Source:       ref.ID,
		Reason:       "Ogre Battlecaster",
		Cards:        []uuid.UUID{legal[0].ID},
		Permission:   &perm,
		Keyword:      game.MayCastKeywordPaidCast,
		AcceptLabel:  "Cast it, paying {R}{R} more",
		DeclineLabel: effects.DeclineLeaveInGraveyard,
	})
	return nil
}

// ogreBattlecasterCast creates "when you cast that spell" once the spell is
// cast through the permission. An on-disk key.
var ogreBattlecasterCast = game.RegisterCastFollowUp("ogre-battlecaster-cast",
	func(g *game.Game, f game.CastFollowUp) error {
		x := 0
		if c, ok := g.LookupCardForEffect(f.Spell); ok {
			if mv, ok := g.ManaValueForEffect(c); ok {
				x = mv
			}
		}
		g.QueueDelayedTriggerForPermissionPlayForEffect(f, f.Player, game.ReflexiveTrigger{
			Label:  "Ogre Battlecaster — gets +X/+0 until end of turn",
			Body:   ogreBattlecasterPumpBody,
			Params: game.EffectParams{Amount: x, Object: f.Object},
		})
		return nil
	})

// ogreBattlecasterPumpBody gives the Ogre +X/+0 until end of turn, if it
// is still the object that attacked. An on-disk key.
var ogreBattlecasterPumpBody = game.ReflexiveBody("ogre-battlecaster/pump",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Amount <= 0 || !effects.OnBattlefield(g, p.Object.ID) {
			return nil
		}
		if c, ok := g.LookupCardForEffect(p.Object.ID); !ok || c.ObjectEpoch != p.Object.Epoch {
			return nil
		}
		return effects.BoostUntilEOT{Target: p.Object.ID, Power: p.Amount,
			Label: "Ogre Battlecaster — +X/+0 until end of turn"}.Apply(effects.NewContext(g, item))
	}, nil)
