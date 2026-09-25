package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Edea, Possessed Sorceress — Legendary Creature — Human Warlock
// {2}{U}{B}{R}, 2/5 (Edea steal-and-sac deck, #1565):
//
//	"Ward {2}
//	 At the beginning of combat on your turn, gain control of target
//	 creature an opponent controls until end of turn. Untap that
//	 creature. It gains haste until end of turn.
//	 Whenever a creature you control but don't own dies, return it to
//	 the battlefield under its owner's control and you draw a card."
//
// The deck's commander, so each clause below is the printed one with
// nothing left out.
//
// **Ward {2}** is the shared Ward constructor (CR 702.21).
//
// **The combat steal** is Act of Treason's three primitives in
// printed order (GainControl until end of turn, UntapTarget, haste
// until end of turn), on a beginning-of-combat trigger that fires on
// its controller's turn only (StepBegan(StepBeginCombat, true)). The
// target is chosen as the trigger goes on the stack (CR 603.3d), and
// a table where no opponent controls a creature drops the trigger
// with no prompt. The theft is a layer-2 effect, so control goes back
// by itself at cleanup.
//
// **The dies trigger** has three rules to get right:
//
//   - "A creature you control but don't own" is judged on the dying
//     creature as it last existed on the battlefield (CR 603.10a).
//     diedCreature reads the card in its new zone, and the controller
//     it held on the battlefield is still on it there (the layer-2
//     controller is materialised onto Card.Controller and a zone
//     change does not reset it). So a creature this trigger stole and
//     that died before cleanup counts, and one of Edea's owner's own
//     creatures does not. A stolen TOKEN counts too: it dies, and the
//     return finds nothing because the token has ceased to exist
//     (CR 111.7), but the draw still happens.
//   - "Return IT" names that one object (CR 400.7). Build records the
//     card's graveyard object epoch, and the effect returns it only if
//     it is still that object in a graveyard when the trigger resolves.
//     A creature exiled from the graveyard in response (Bojuka Bog)
//     stays exiled, and so does one that left the graveyard and came
//     back.
//   - "Under its OWNER's control", not Edea's. The return hands the
//     creature back. The draw goes to Edea's controller, and it
//     happens whether or not the return did (CR 608.2c: the effect
//     does as much as it can).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     edeaPossessedSorceressOracleID,
		Name:         "Edea, Possessed Sorceress",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Edea, Possessed Sorceress — ward {2}"),
			Targeting(On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
				"Edea, Possessed Sorceress — gain control of target creature an opponent controls until end of turn",
				edeaStealEffect),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			{
				Watches:   []game.EventKind{game.EventLTB},
				AppliesTo: edeaCreatureYouControlButDontOwnDied,
				Key:       edeaReturnLabel,
				// The dead creature's post-move epoch is a board read
				// made at trigger time (ADR 0041 P9's fill-in Build):
				// it is what edeaReturnAndDraw checks against at
				// resolution to confirm "it" is still the same object
				// in a graveyard, and it is the epoch the object has
				// NOW, in its new zone — not item.Trigger.Object.Epoch,
				// which is CR 603.10's battlefield-exit snapshot and
				// answers the epoch it had THERE, before it left.
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, edeaReturnLabel)
					epoch := -1
					if c, ok := g.LookupCardForEffect(ev.CardID); ok {
						epoch = c.ObjectEpoch
					}
					item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: epoch}
					return item
				},
				Effect: edeaReturnAndDrawEffect,
			},
		},
	})
}

const (
	edeaPossessedSorceressOracleID = "7beb7505-b765-486d-a205-78325922a10f"
	edeaReturnLabel                = "Edea, Possessed Sorceress — return it to its owner and draw a card"
)

// edeaStealEffect is the combat trigger's resolution: gain control
// until end of turn, untap, haste until end of turn, in printed order.
func edeaStealEffect(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	target := item.Targets[0].ID
	if err := (GainControl{
		Target:   target,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Edea, Possessed Sorceress — gain control until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"haste"},
		Label:    "Edea, Possessed Sorceress — haste until end of turn",
	}.Apply(ctx)
}

// edeaCreatureYouControlButDontOwnDied is the dies trigger's
// condition: a creature died that Edea's controller controlled when it
// left the battlefield and does not own.
func edeaCreatureYouControlButDontOwnDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller && dead.Owner != source.Controller
}

// edeaReturnAndDrawEffect is the dies trigger's resolution (ADR 0041
// P9): the dead creature's identity rides item.Params.Object, stamped
// by the row's fill-in Build.
func edeaReturnAndDrawEffect(g *game.Game, item *game.StackItem) error {
	return edeaReturnAndDraw(g, item, item.Params.Object.ID, item.Params.Object.Epoch)
}

// edeaReturnAndDraw returns the dead creature to the battlefield under
// its owner's control if it is still the same object in a graveyard,
// then draws Edea's controller a card either way.
func edeaReturnAndDraw(g *game.Game, item *game.StackItem, dead uuid.UUID, epoch int) error {
	ctx := NewContext(g, item)
	if c, ok := g.LookupCardForEffect(dead); ok && c.ObjectEpoch == epoch {
		if z := g.FindCardZoneForEffect(dead); z != nil && z.Kind == game.ZoneGraveyard {
			if err := (ReturnFromGraveyard{
				Target:     dead,
				Dest:       game.ZoneBattlefield,
				Controller: c.Owner,
			}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}
