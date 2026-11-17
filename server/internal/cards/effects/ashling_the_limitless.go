package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashling, the Limitless — Legendary Creature — Elemental Sorcerer
// {2}{R}, 2/3 (Lorwyn Eclipsed Commander, #5428):
//
//	"Elemental permanent spells you cast from your hand gain evoke {4}
//	 as you cast them. (If you cast a spell for its evoke cost, it's
//	 sacrificed when it enters.)
//	 Whenever you sacrifice a nontoken Elemental, create a token that's a
//	 copy of it. The token gains haste until end of turn. At the
//	 beginning of your next end step, sacrifice it unless you pay
//	 {W}{U}{B}{R}{G}."
//
// The grant is EvokeForSpellsYouCastFromHand (granted_evoke.go): each
// Elemental permanent spell its controller casts from their hand may be
// cast for {4} rather than its mana cost, and the permanent it becomes
// is then sacrificed as it enters (CR 702.74a). The spell gained evoke as
// it was cast (CR 610.5), so it is sacrificed even if Ashling has left by
// then, and an Elemental that prints its own evoke keeps that cost
// beside this one (CR 118.9a: one alternative cost per spell). "Elemental"
// is read off the card being cast, so a kindred Elemental enchantment and
// a changeling are reached.
//
// The trigger sees every nontoken Elemental its controller sacrifices,
// Ashling included (a leaves-the-battlefield ability looks back in time),
// read as the permanent stood when it was sacrificed. The token copies
// the permanent as it last existed on the battlefield (CR 707.2, CR
// 608.2h), haste is a gain until end of turn and not part of the copy,
// and one delayed trigger (CR 603.7) asks at the beginning of its
// controller's next end step: pay {W}{U}{B}{R}{G}, or sacrifice the
// token. A token that has left by then asks nothing, and one its
// controller no longer controls cannot be sacrificed by them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77a9e7cf-76e9-4323-8504-a5b831de00be",
		Name:         "Ashling, the Limitless",
		Completeness: CompletenessFull,
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{
			EvokeForSpellsYouCastFromHand("{4}", "Elemental permanent spells", game.PermissionFilter{PermanentOnly: true, CreatureType: "Elemental"}),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, ashlingYouSacrificedANontokenElemental,
				"Ashling, the Limitless — create a token that's a copy of it; it gains haste until end of turn; at the beginning of your next end step, sacrifice it unless you pay {W}{U}{B}{R}{G}",
				ashlingCopyTheSacrificedElemental),
		},
	})
}

// ashlingYouSacrificedANontokenElemental is "you sacrifice a nontoken
// Elemental": the sacrifice is announced while the permanent is still on
// the battlefield, so it is read as it stands there, through the layered
// view and Ashling's own text.
func ashlingYouSacrificedANontokenElemental(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil || source == nil || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken() && c.HasSubtype(source.TextSubtype("Elemental"))
}

// ashlingCopyTheSacrificedElemental creates the token copy of the
// sacrificed permanent as it last existed, then hands the token to the
// rest of the ability.
func ashlingCopyTheSacrificedElemental(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ref game.ObjectRef
	if obj := ctx.Trigger().Object; obj != nil {
		ref = obj.Ref()
	} else if r, ok := g.PermanentRefForEffect(ctx.Trigger().Event.CardID); ok {
		ref = r
	} else {
		return nil
	}
	tmpl, ok := g.TokenCopyTemplateForObjectForEffect(ref)
	if !ok {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: copiedToken(tmpl), Count: 1}},
	}, ctx.ItemTail(ashlingHasteAndUpkeep))
}

// ashlingHasteAndUpkeep is "The token gains haste until end of turn. At
// the beginning of your next end step, sacrifice it unless you pay
// {W}{U}{B}{R}{G}." An on-disk key.
var ashlingHasteAndUpkeep = CardTail("ashling-the-limitless/haste-and-pay-or-sacrifice", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"},
			Label: "Ashling, the Limitless — the token gains haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepEnd,
		Controller:         ctx.Item.Controller,
		ControllerTurnOnly: true,
		Label:              "Ashling, the Limitless — sacrifice the token unless you pay {W}{U}{B}{R}{G}",
		Cards:              o.IDs,
		Body:               ashlingPayOrSacrificeBody,
	}.Apply(ctx)
})

// ashlingPayOrSacrificeBody is the delayed trigger: for the token it
// names, if its controller still controls it, pay {W}{U}{B}{R}{G} or
// sacrifice it. An on-disk key.
//
// The question is UpkeepPayUnless, not PayUnless (#5803). It is the
// #997 shape that ADR 0018's amendment names for an end step too: the
// controller is asked, during their own end step, about their own
// token. The payment is made as the trigger resolves (CR 118.12), so
// nobody gets priority (CR 117.3b) and the end step cannot end (CR
// 500.2) until they answer, and a decline sacrifices the token in that
// end step. The decline keeps its key, so no new key is needed.
var ashlingPayOrSacrificeBody = game.SimpleDelayedBody("ashling-the-limitless/pay-wubrg-or-sacrifice", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || !onBattlefield(g, t.ID) || c.Controller != item.Controller {
			continue
		}
		if err := (UpkeepPayUnless{
			Chooser:  item.Controller,
			Cost:     "{W}{U}{B}{R}{G}",
			Question: "Ashling, the Limitless — pay {W}{U}{B}{R}{G}, or sacrifice " + c.Name + "?",
			Else:     ashlingSacrificeTheToken,
			Carry:    []uuid.UUID{t.ID},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// ashlingSacrificeTheToken is the "or else": sacrifice the token the
// prompt carried, if its controller still controls it. An on-disk key.
var ashlingSacrificeTheToken = PayUnlessElse("pay-unless/ashling-the-limitless-sacrifice-token",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		for _, id := range r.Carry {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !onBattlefield(ctx.Game, id) || c.Controller != r.Chooser {
				continue
			}
			if err := (SacrificePermanent{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
