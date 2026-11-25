package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b233_helpers.go — Commander 2017, batch 233 (#5533).

// KickerTapUntapped is "Kicker—Tap an untapped <permanent> you control"
// (Blood Tribute, CR 702.33a, CR 601.2h): a kicker whose cost is tapping
// one untapped permanent, not the {T} symbol (CR 302.6). It is the
// optional slot's tap component, so the candidate walk, validator and payer
// are New Blood's (TapUntappedCost); the caster names the permanent on
// teamwork_ids and it is tapped with the spell on the stack. Kicked once at
// most (CR 702.33d).
func KickerTapUntapped(label string, preds ...CardPredicate) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.KickerKey,
		TapOthers: TapOthersPrice(1, label, preds...),
		Label:     "Kicker—Tap " + label,
	}
}

// DragonAttackTrigger is "Whenever a Dragon you control attacks, <effect>"
// (Silumgar, the Drifting Death; Kolaghan, the Storm's Fury): one trigger
// per attacking Dragon (CR 508.3a), the defending player read off the
// attack as it is declared, so a Dragon attacking a planeswalker names that
// planeswalker's controller. The effect is told the defending player. The
// key is the label: an on-disk identity.
func DragonAttackTrigger(label string, apply func(ctx *Context, defender uuid.UUID) error) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return B14DragonYouControlAttacked(ev, source, g)
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Player = B17DefendingPlayer(g, ev)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return apply(NewContext(g, item), item.Params.Player)
		},
	}
}

// S82OldB233ReturnWithHaste is "Return target creature card from your
// graveyard to the battlefield. That creature gains haste." (Apprentice
// Necromancer, Cauldron Dance). It reports the returned object (zero when it never arrived, CR 608.2b),
// and gives it haste with no duration, as printed (Nahiri, the Harbinger).
func S82OldB233ReturnWithHaste(ctx *Context, card uuid.UUID) (game.ObjectRef, bool, error) {
	if err := (ReturnFromGraveyard{Target: card, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return game.ObjectRef{}, false, err
	}
	c, ok := BattlefieldCardForEffect(ctx.Game, card)
	if !ok {
		return game.ObjectRef{}, false, nil
	}
	err := (ScopedEffectFor{
		Target:   card,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: game.IndefiniteDuration(),
		Label:    "gains haste",
	}).Apply(ctx)
	return game.ObjectRef{ID: card, Epoch: c.ObjectEpoch}, true, err
}
