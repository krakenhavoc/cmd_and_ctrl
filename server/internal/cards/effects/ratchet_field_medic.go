package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ratchet, Field Medic // Ratchet, Rescue Racer (#3195, ADR 0152) — a
// More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {2}{W}, 2/4:
//
//	"More Than Meets the Eye {1}{W}
//	 Lifelink
//	 Whenever you gain life, you may convert Ratchet. When you do,
//	 return target artifact card with mana value less than or equal to
//	 the amount of life you gained this turn from your graveyard to the
//	 battlefield tapped."
//
// Back face, Legendary Artifact — Vehicle, 1/4:
//
//	"Living metal
//	 Lifelink
//	 Whenever one or more nontoken artifacts you control are put into a
//	 graveyard from the battlefield, convert Ratchet. This ability
//	 triggers only once each turn."
//
// The "you may" is asked as the ability resolves (CR 608.2c), and "when
// you do" is a CR 603.12 reflexive trigger created only if Ratchet
// actually converted (CR 701.28e: an already-turned-over Ratchet does
// not, so nothing is returned). The reflexive trigger chooses its target
// after the convert, with the life gained this turn counted at that
// point and again as it resolves (CR 608.2b). The back face's "one or
// more" is a batch (CR 603.2c), and "only once each turn" counts the
// triggers put on the stack, a countered one included.
//
// No simplification.
const ratchetOracleID = "4f95664f-6113-41e6-a99e-12bfc9d1710a"

const ratchetBackLabel = "Ratchet, Rescue Racer — convert Ratchet"

var ratchetReturnBody = game.ReflexiveBody("ratchet-field-medic/return", simpleBody(returnFirstLegalGraveyardTargetToBattlefieldTapped),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard(
			"target artifact card in your graveyard with mana value at most the life you gained this turn",
			YouOwn(), Artifact(), ManaValueAtMostLifeGainedThisTurn())
	}))

func init() {
	Register(Spec{
		OracleID:         ratchetOracleID,
		Name:             "Ratchet, Field Medic",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Ratchet, Field Medic — you may convert Ratchet; when you do, return an artifact card", ratchetMayConvert),
		},
	})
	Register(Spec{
		OracleID:        ratchetOracleID + "#1",
		Name:            "Ratchet, Rescue Racer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static:          []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !permanentYouControlWasPutIntoAGraveyard(ev, source, g, "artifact") {
					return false
				}
				if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.IsToken() {
					return false
				}
				return !b11TriggeredThisTurn(g, source.InstanceID, ratchetBackLabel)
			}, ratchetBackLabel, Do(ConvertThis{}))),
		},
	})
}

func ratchetMayConvert(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Ratchet, Field Medic — convert Ratchet? (When you do, return an artifact card from your graveyard to the battlefield tapped.)",
		OnYes: func(ctx *Context) error {
			converted, err := convertThisReporting(ctx)
			if err != nil || !converted {
				return err
			}
			return WhenYouDo("Ratchet, Field Medic — return target artifact card from your graveyard to the battlefield tapped",
				ratchetReturnBody).Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
