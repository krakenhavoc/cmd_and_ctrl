package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b18_helpers.go — shared bodies for Star Trek batch 18 (#4973).

// trekB18DiedDuringYourTurn is "when this creature dies during your
// turn" (Starfleet Security): the source went to a graveyard from the
// battlefield while the player who controlled it as it last existed
// there (CR 603.10a) is the active player. Part of the trigger event,
// read as the death happens; not rechecked.
func trekB18DiedDuringYourTurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ThisDied(ev, source, lki, g) && IsYourTurn(g, leftUnderControlOf(ev, *source))
}

// trekB18ThisAttackedThisTurn is the trigger-time half of "if this
// creature attacked this turn" (Targ Wrangler): the source object was
// declared as an attacker this turn, per object (CR 400.7).
func trekB18ThisAttackedThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.AttackedThisTurn(source.InstanceID)
}

// trekB18CondThisAttackedThisTurn is the resolution half of the same
// intervening "if" (CR 603.4): the object the ability came from attacked
// this turn, read from its last known information once it has left the
// battlefield (CR 608.2h). The turn's attack history only grows, and it
// is a fact about the source object, so it declares a source read.
type trekB18CondThisAttackedThisTurn struct{}

func (trekB18CondThisAttackedThisTurn) Holds(ctx *Context) bool {
	ref, ok := ctx.SourceRef()
	return ok && ctx.Game.ObjectAttackedThisTurn(ref)
}

func (trekB18CondThisAttackedThisTurn) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondSource, 0), true
}

// trekB18AltCostKeyBeamMeUp is beam me up's offer key: on the wire as
// `alternative_cost`, on StackItem.AltCost and on a permanent's
// CastProvenance. An on-disk identity: never renamed, never reused. The
// same key every Star Trek batch that prints the keyword uses.
const trekB18AltCostKeyBeamMeUp = "beam_me_up"

// trekB18BeamMeUp is "Beam me up {cost} (You may cast this card from your
// graveyard for {cost} if you also return a creature you control to its
// owner's hand. Then exile this spell.)" — Transporter Malfunction.
//
// Flashback's place and exile (FromZone graveyard, ExileOnLeavingStack,
// CR 702.34a's shape) with Daze's bounce as the card component of the
// price: the caster names one creature they control in alt_cost_ids and
// it returns to its owner's hand as the cost is paid, with the spell
// already on the stack (CR 601.2a, 601.2h). A caster who controls no
// creature cannot claim the offer. The card file still opens the zone:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{trekB18BeamMeUp("{3}{R}{R}{R}")},
func trekB18BeamMeUp(cost string) game.AlternativeCost {
	ac := Flashback(cost)
	ac.Key = trekB18AltCostKeyBeamMeUp
	ac.Label = "Beam me up " + cost
	ac.ReturnToHand = PermanentYouControl("a creature you control", Creature())
	ac.PayLabel = "a creature you control"
	return ac
}

// trekB18TokenCopyOfThisSacrificedAtNextEndStep is Tasha Yar's graveyard
// ability body: "Create a token that's a copy of this card. Sacrifice it
// at the beginning of the next end step." Encore's shape without the
// per-opponent loop, the attack requirement or the haste grant: the cost
// has just exiled the card, so the token copies the card as it is in
// exile (its copiable values, CR 707.2), and one delayed trigger
// sacrifices it at the beginning of the next end step (CR 603.7).
func trekB18TokenCopyOfThisSacrificedAtNextEndStep(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		cursor := b25LastEventSeq(g)
		if err := (CreateTokenCopy{
			Controller: item.Controller,
			Copy:       item.SourceCardID,
			N:          1,
		}).Apply(ctx); err != nil {
			return err
		}
		tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
		if len(tokens) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label: label,
			Cards: tokens,
			Body:  sacrificeListedCardsBody,
		}.Apply(ctx)
	}
}

// trekB18WillRikerReveal is Will Riker's combat trigger: "reveal cards
// from the top of your library until you reveal two land cards. Put those
// land cards onto the battlefield and the rest on the bottom of your
// library in a random order. Sacrifice those lands at the beginning of
// your next end step."
//
// Open the Way's reveal with X fixed at two and the lands untapped: the
// run stops on the second land card, and a library that runs out first
// reveals itself whole and puts every land it found (none, one). A token
// in the library is revealed but is not a land card (CR 108.2). The put
// is keyed, and its tail puts the rest on the bottom and schedules the
// sacrifice of exactly the lands that entered.
func trekB18WillRikerReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	run, _ := revealUntilCount(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsLand() }, 2,
		"Will Riker, First in the Field — revealed until two land cards")
	return PutFromLibraryOntoBattlefield{
		Player:   ctx.Controller(),
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() },
		All:      true,
		ThenTail: trekB18WillRikerLandsTail,
	}.Apply(ctx)
}

// trekB18WillRikerLandsTail is the rest of Will Riker's resolution, run
// once the lands have entered: the rest go to the bottom in a random
// order, then a delayed trigger is scheduled for the beginning of the
// controller's next end step (ControllerTurnOnly: "your"; it is
// controlled by the combat trigger's controller, CR 603.7e) over each
// land that entered, pinned to the object it is now (CR 400.7). On the
// controller's own turn that is this turn's end step. An on-disk key.
var trekB18WillRikerLandsTail = PutFromLibraryTail("will-riker/rest-on-bottom-then-sacrifice-those-lands",
	func(ctx *Context, res PutFromLibraryResult) error {
		if err := PutRestOnBottomInRandomOrder(ctx.Game, res); err != nil {
			return err
		}
		var ids []uuid.UUID
		var objects []game.ObjectRef
		for _, id := range res.Entered {
			ref, ok := ctx.Game.PermanentRefForEffect(id)
			if !ok || !onBattlefield(ctx.Game, id) {
				continue
			}
			ids = append(ids, id)
			objects = append(objects, ref)
		}
		if len(objects) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label:              "Will Riker, First in the Field — sacrifice those lands",
			Controller:         res.Player,
			ControllerTurnOnly: true,
			Cards:              ids,
			Body:               trekB18SacrificeListedObjectsBody,
			Params:             game.EffectParams{Objects: objects},
		}.Apply(ctx)
	})

// trekB18SacrificeListedObjectsBody is "sacrifice those <permanents>"
// over objects pinned by CR 400.7 epoch: each one is sacrificed only
// while it is still the object it was when the trigger was created, on
// the battlefield, and controlled by the trigger's controller — a player
// sacrifices only a permanent they control (CR 701.21a). A land that left
// the battlefield is no longer there to be sacrificed (CR 603.7c), and one
// that came back is a new object and stays. An on-disk key.
var trekB18SacrificeListedObjectsBody = game.DelayedBody("sacrifice/listed-objects-you-control", trekB18SacrificeListedObjects)

func trekB18SacrificeListedObjects(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := NewContext(g, item)
	for _, ref := range p.Objects {
		c, ok := g.LookupCardForEffect(ref.ID)
		if !ok || c.ObjectEpoch != ref.Epoch || !onBattlefield(g, ref.ID) || c.Controller != item.Controller {
			continue
		}
		if err := (SacrificePermanent{Target: ref.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
