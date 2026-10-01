package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Questing Beast — Legendary Creature — Beast {2}{G}{G}, 4/4:
//
//	"Vigilance, deathtouch, haste
//	 Questing Beast can't be blocked by creatures with power 2 or less.
//	 Combat damage that would be dealt by creatures you control can't be
//	 prevented.
//	 Whenever Questing Beast deals combat damage to an opponent, it
//	 deals that much damage to target planeswalker that player
//	 controls."
//
// The block restriction is the shared CantBeBlockedBy (#750). The third
// line is ADR 0107 §5's battlefield static, judged against each combat
// damage source as it was when the damage was dealt (CR 608.2h). The
// trigger's target clause is built from the trigger's own event
// (TargetsFrom, CR 603.3d): a planeswalker controlled by the opponent it
// damaged. With none, the trigger has no target and is removed. "That
// much" is the damage actually dealt, off the event. The damage to the
// planeswalker is not combat damage, so the static does not cover it.
//
// No simplifications.
func init() {
	t := On(game.EventDealDamage, questingBeastHitAnOpponent,
		"Questing Beast — that much damage to target planeswalker that player controls", questingBeastRedirect)
	t.TargetsFrom = questingBeastPlaneswalkerClause
	Register(Spec{
		OracleID:              "b685757b-521e-4353-a233-97052359723d",
		Name:                  "Questing Beast",
		Completeness:          CompletenessFull,
		PrintedKeywords:       []string{"vigilance", "deathtouch", "haste"},
		BlockRules:            []game.BlockRule{CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less")},
		DamageCantBePrevented: CombatDamageByYourCreaturesCantBePrevented(),
		Triggered:             []game.TriggeredAbility{t},
	})
}

func questingBeastHitAnOpponent(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ThisDealtCombatDamageToAPlayer(ev, source, lki, g) && ev.Target != source.Controller
}

// questingBeastPlaneswalkerClause reads only the trigger's event, so a
// restore rebuilds the same clause.
func questingBeastPlaneswalkerClause(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	player := tc.Event.Target
	if player == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "permanent", Label: "target planeswalker that player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsPlaneswalker() && c.Controller == player
		},
		Min: 1, Max: 1,
	}
}

func questingBeastRedirect(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: item.Trigger.Event.Amount}.Apply(ctx)
}
