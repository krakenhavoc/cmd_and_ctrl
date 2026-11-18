package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lorcan, Warlock Collector — Legendary Creature — Devil {5}{B}{B}, 6/6
// (Adventures in the Forgotten Realms):
//
//	"Flying
//	 Whenever a creature card is put into an opponent's graveyard from
//	 anywhere, you may pay life equal to its mana value. If you do, put it
//	 onto the battlefield under your control. It's a Warlock in addition to
//	 its other types.
//	 If a Warlock you control would die, exile it instead."
//
// The trigger watches every way a card reaches a graveyard (Lazav's event
// list) and names the card, with the epoch it had there, on the item
// (CR 400.7). On resolution the controller is asked, with the life cost
// declared on the prompt; the answer re-checks the card is still that object
// in that graveyard, pays life equal to its mana value (CR 119.4: a player
// whose life total is below it can't), and puts it onto the battlefield under
// the controller's control as a Warlock in addition to its other types
// (EntersAs, CR 614.12). The life is paid even if the card has gone by then.
// The last line is a replacement: a Warlock the controller controls going
// from the battlefield to a graveyard goes to exile instead. No
// simplification.
func init() {
	Register(Spec{
		OracleID:        "6c87c261-00a8-46f2-92c8-42009a0a2faf",
		Name:            "Lorcan, Warlock Collector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventMove || ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == src.Controller && c.HasSubtypeWord(src.TextSubtype("Warlock"))
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:    "Lorcan, Warlock Collector: exile a Warlock instead of dying",
			RowLabel: "If a Warlock you control would die, exile it instead.",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard, game.EventMill, game.EventCounterSpell},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if _, ok := b06CardInOpponentsGraveyard(ev, source.Controller, g); !ok {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && !c.IsToken()
			},
			Key: lorcanLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, lorcanLabel)
				ref := game.ObjectRef{ID: ev.CardID}
				if c, ok := g.LookupCardForEffect(ev.CardID); ok {
					ref.Epoch = c.ObjectEpoch
				}
				item.Params.Object = ref
				return item
			},
			Effect: lorcanAsk,
		}},
	})
}

const lorcanLabel = "Lorcan, Warlock Collector — you may pay life equal to its mana value to put that creature card onto the battlefield under your control"

func lorcanAsk(g *game.Game, item *game.StackItem) error {
	ref := item.Params.Object
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok {
		return nil
	}
	mv := c.ManaValue()
	return MayChoice{
		Question: "Lorcan, Warlock Collector — pay " + strconv.Itoa(mv) + " life to put " + c.Name + " onto the battlefield under your control?",
		LifeCost: mv,
		Then:     lorcanPaid,
		Objects:  []game.ObjectRef{ref},
	}.Apply(NewContext(g, item))
}

var lorcanPaid = MayYes("confirm/lorcan-warlock-collector-pay-life", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Objects) != 1 {
		return nil
	}
	ref := r.Objects[0]
	c, ok := ctx.Game.LookupCardForEffect(ref.ID)
	if !ok {
		return nil
	}
	n := c.ManaValue()
	p := ctx.Game.PlayerByIDForEffect(r.Chooser)
	if p == nil || p.Eliminated || p.Life < n {
		return nil
	}
	if n > 0 {
		if err := ctx.Game.PayLifeForEffect(ctx.Source(), r.Chooser, n); err != nil {
			return err
		}
	}
	z := ctx.Game.FindCardZoneForEffect(ref.ID)
	if z == nil || z.Kind != game.ZoneGraveyard || c.ObjectEpoch != ref.Epoch {
		return nil
	}
	opts := game.ZoneEntryOptions{
		Controller: ctx.Controller(),
		EntersAs: &game.EntersAs{
			AddSubtypes: ctx.TextSubtypes("Warlock"),
			Label:       "Lorcan, Warlock Collector — a Warlock in addition to its other types",
		},
	}
	return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{ref.ID}, opts, game.TailRef{}, 0)
})
