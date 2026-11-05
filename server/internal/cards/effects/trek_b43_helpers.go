package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b43_helpers.go — Star Trek Commander, batch 43 (#4974). The three
// bodies below were each written once already; batch 43 prints them a
// second time, so they are named here and both cards call them.

// pumpBlockerTwoThenDraw is the whole resolution of "target blocking
// creature [you control] gets +2/+2 until end of turn. Draw a card."
// (Captain's Defense, Evasive Maneuvers). The target clause is the
// card's own; a target that is no longer legal (it stopped blocking, or
// left) makes the whole spell do nothing, the draw included (CR 608.2b).
func pumpBlockerTwoThenDraw(label string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		if err := (BoostUntilEOT{
			Target: item.Targets[0].ID, Power: 2, Toughness: 2,
			Label: label,
		}).Apply(ctx); err != nil {
			return err
		}
		return DrawCards{N: 1}.Apply(ctx)
	}
}

// whenThisEntersFourDamageToTappedOpponentCreature is "When this enters,
// it deals 4 damage to target tapped creature an opponent controls"
// (The Wasp, Janet Van Dyne; Shuttle Crew). With no legal target the
// trigger is removed from the stack (CR 603.3d); a target untapped or
// gone by resolution is dealt nothing (CR 608.2b).
func whenThisEntersFourDamageToTappedOpponentCreature(label string) game.TriggeredAbility {
	return TriggerWithPurpose(
		Targeting(
			WhenThisEnters(label, fourDamageToFirstLegalTarget),
			TargetCreature("target tapped creature an opponent controls", OpponentControls(), b751Tapped())),
		ForTargets(DamageToTarget(0, 4)))
}

func fourDamageToFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: ts[0].ID, Amount: 4}.Apply(ctx)
}

// whenThisHitsAnOpponentThatMuchToEachOther is "Whenever this deals
// combat damage to an opponent, it deals that much damage to each other
// opponent" (Hydra Omnivore; U.S.S. Prometheus's Multivector Assault
// Mode). How much landed and on whom are facts of the damage event,
// carried on the item; the damage to the others is one damage instance.
func whenThisHitsAnOpponentThatMuchToEachOther(key string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return b38ThisDealtCombatDamageToAnOpponent(ev, source, g)
		},
		Key:    key,
		Effect: thatMuchToEachOpponentButTheOneHit,
	}
}

func thatMuchToEachOpponentButTheOneHit(g *game.Game, item *game.StackItem) error {
	amount, hit := item.Trigger.Event.Amount, item.Trigger.Event.Target
	ctx := NewContext(g, item)
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if opp == hit {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: opp, Amount: amount}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// --- U.S.S. Enterprise-C, Time-Tossed ------------------------------------

// trekB43EnterpriseCDiscardHand is the yes to "you may discard your
// hand": the whole hand goes as one discard, and Forgotten Creation's
// tail draws one card for each card really discarded this way.
var trekB43EnterpriseCDiscardHand = MayYes("confirm/uss-enterprise-c-discard-hand", func(ctx *Context, r game.ConfirmAnswer) error {
	p := ctx.Game.PlayerByIDForEffect(r.Chooser)
	if p == nil {
		return nil
	}
	return ctx.Game.DiscardRandomKeyedForEffect(r.Chooser, p.Hand.Size(), ctx.ItemTail(forgottenCreationDraw))
})

// --- U.S.S. Farragut, Bold Explorer --------------------------------------

// trekB43FarragutHasteAndBottom is the rest of the landfall trigger once
// the put has settled: the creature that entered gains haste (no
// duration is printed, so for as long as it is that object on the
// battlefield, CR 611.2a), and a delayed trigger at the beginning of the
// next end step puts it on the bottom of its owner's library. A declined
// "you may", or no creature card in hand, does nothing more. An on-disk
// key.
var trekB43FarragutHasteAndBottom = PutFromHandTail("uss-farragut/haste-then-bottom-at-end-step", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	g := ctx.Game
	c, ok := g.LookupCardForEffect(res.Entered)
	if !ok || !onBattlefield(g, res.Entered) {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   res.Entered,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), res.Entered),
		Label:    "U.S.S. Farragut — that creature gains haste",
	}).Apply(ctx); err != nil {
		return err
	}
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   res.Player,
		SourceCardID: res.Source,
		Label:        "U.S.S. Farragut — put that creature on the bottom of your library; when you do, draw a card",
		At:           game.StepEnd,
		Cards:        []uuid.UUID{res.Entered},
		Body:         trekB43FarragutBottomBody,
		Params:       game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	})
	return nil
})

// trekB43FarragutBottomBody is the delayed trigger: the creature, if it
// is still the object that entered and still on the battlefield
// (CR 603.7c), goes to the bottom of its owner's library. An on-disk key.
var trekB43FarragutBottomBody = game.DelayedBody("uss-farragut/bottom-of-library", trekB43FarragutBottom)

func trekB43FarragutBottom(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !onBattlefield(g, id) {
		return nil
	}
	ctx := NewContext(g, item)
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{ToBottom: true},
		ctx.ItemTail(trekB43FarragutWhenYouDo))
}

// trekB43FarragutWhenYouDo is "When you do, draw a card": a reflexive
// trigger (CR 603.12) only when the creature really reached a library —
// not when it had gone, or a commander's owner sent it to the command
// zone instead (CR 903.9). An on-disk key.
var trekB43FarragutWhenYouDo = CardTail("uss-farragut/when-you-do-draw", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return WhenYouDo("U.S.S. Farragut — draw a card", drawOneBody).Apply(ctx)
})

// --- U.S.S. Stargazer, Here and There ------------------------------------

// trekB43StargazerIllusion is the Picard Maneuver: a 4/3 Spacecraft
// Illusion with flying, tapped and attacking what its controller chooses
// among what they could attack (CR 508.4), offered first what the
// Stargazer attacks. It was never declared, so it fires no attack
// trigger (CR 508.4).
func trekB43StargazerIllusion(g *game.Game, item *game.StackItem) error {
	return CreateTokensAttackingYourChoice{
		Template: TokenCard("4/3 colorless Spacecraft Illusion artifact with flying"),
		N:        1,
		Tapped:   true,
		Prefer:   triggerAttackTarget(item),
		Label:    "U.S.S. Stargazer",
		Tail:     game.KeyedTail{Ref: trekB43StargazerExile},
	}.Apply(NewContext(g, item))
}

// trekB43StargazerExile is "Exile that token at end of combat": a
// delayed trigger at the beginning of this combat's end of combat step
// (CR 511.2) naming the token that landed. An on-disk key.
var trekB43StargazerExile = AttackingTokensThen("attacking-tokens/uss-stargazer-exile-the-illusion",
	func(ctx *Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: "U.S.S. Stargazer — exile the Spacecraft Illusion",
			Cards: created,
			Body:  exileListedCardsBody,
		}.Apply(ctx)
	})

// --- Shore Leave Planet --------------------------------------------------

// trekB43ShoreLeaveCopy is "This land becomes a copy of target creature
// you control until end of turn, except it isn't legendary and it's a
// Robot artifact in addition to its other types." The exceptions are
// part of the copiable values (CR 707.9b). A target gone by resolution
// makes the ability do nothing (CR 608.2b).
func trekB43ShoreLeaveCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return BecomeCopy{
		Targets: []uuid.UUID{ctx.Source()},
		Of:      targets[0].ID,
		Except: func(v *game.PrintedValues) {
			v.RemoveSupertype("Legendary")
			v.AddCardType("Artifact")
			v.AddSubtype("Robot")
		},
		Label: "Shore Leave Planet — becomes a nonlegendary Robot artifact copy until end of turn",
	}.Apply(ctx)
}
