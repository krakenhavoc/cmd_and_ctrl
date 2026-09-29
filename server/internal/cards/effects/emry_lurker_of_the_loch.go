package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emry, Lurker of the Loch — Legendary Creature — Merfolk Wizard
// {2}{U}, 1/2:
//
//	"Affinity for artifacts (This spell costs {1} less to cast for
//	 each artifact you control.)
//	 When Emry enters, mill four cards.
//	 {T}: Choose target artifact card in your graveyard. You may
//	 cast that card this turn. (You still pay its costs. Timing
//	 rules still apply.)"
//
// Three shapes, each already in the catalog's vocabulary:
//
//   - Affinity is Spec.SelfCostModifiers (#746, CR 702.41a) — the
//     same AffinityFor Thoughtcast and Myr Enforcer use.
//   - The ETB mill is MillCards on the ordinary WhenThisEnters
//     trigger.
//   - The tap ability targets a card in the graveyard and grants a
//     plain CR 611.2c cast permission over it — no alternative cost,
//     no exile, "you still pay its costs" is the printed reminder for
//     exactly that. Duration left zero reads as "until end of turn",
//     which is what "this turn" means. Snapcaster Mage's Flashback
//     grant (granted_permissions.go) is the same shape with an
//     AltCostKey and an exile clause bolted on; Emry's ability is the
//     bare permission underneath it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da3e7d3d-2ca0-40c3-9602-fca37c92f507",
		Name:         "Emry, Lurker of the Loch",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Emry, Lurker of the Loch — mill four cards", Do(MillCards{N: 4})),
		},
		Activated: []ActivatedAbility{{
			Label: "{T}: Choose target artifact card in your graveyard. You may cast that card this turn.",
			Cost:  TapCost(),
			Targets: TargetCardInGraveyard("target artifact card in your graveyard",
				YouOwn(), Artifact()),
			Effect: emryGrantCastFromGraveyard,
		}},
	})
}

// emryGrantCastFromGraveyard grants the controller permission to cast
// the chosen graveyard card this turn, paying its own costs. The
// target is re-read out of ctx.LegalTargets() rather than off
// item.Targets, so a card that left the graveyard in response (or was
// never there — an empty graveyard leaves nothing to target) grants
// nothing rather than a permission over a card that is not there.
func emryGrantCastFromGraveyard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	g.GrantCastPermissionOverCardForEffect(ts[0].ID, game.CastPermission{
		Player: item.Controller,
		Zone:   game.ZoneGraveyard,
		Source: ctx.Source(),
		Label:  "Emry, Lurker of the Loch — you may cast that card this turn",
	})
	return nil
}
