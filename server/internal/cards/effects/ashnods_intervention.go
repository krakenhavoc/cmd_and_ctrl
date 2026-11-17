package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const ashnodsInterventionReturn = "ashnods-intervention/return" // an on-disk identity: never renamed

// Ashnod's Intervention — Instant {B} (The Brothers' War):
//
//	"Until end of turn, target creature gets +2/+0 and gains "When this
//	 creature dies or is put into exile from the battlefield, return it to
//	 its owner's hand.""
//
// Undying Malice's duration grant, with God-Eternal Oketra's "dies or is put
// into exile" trigger condition (a bounce or a tuck does not count). "It" is
// the object that left: if the card has gone from the graveyard or exile
// since, and perhaps come back, nothing is found (CR 400.7), and a token has
// ceased to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1a7d108-8073-42a5-ade0-a0643d078268",
		Name:         "Ashnod's Intervention",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Grants: []AbilityGrant{{
			Key: ashnodsInterventionReturn,
			Triggered: []game.TriggeredAbility{
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b22SelfDiedOrWasExiledFromBattlefield(ev, source)
				}, "Ashnod's Intervention — return it to its owner's hand", ashnodsInterventionBack),
			},
			Text: "When this creature dies or is put into exile from the battlefield, return it to its owner's hand.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{
				Target: target,
				Power:  2,
				Label:  "Ashnod's Intervention — +2/+0 until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{ashnodsInterventionReturn},
				Label:  "Ashnod's Intervention — until end of turn, it returns when it dies or is exiled",
			}.Apply(ctx)
		},
	})
}

func ashnodsInterventionBack(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || (z.Kind != game.ZoneGraveyard && z.Kind != game.ZoneExile) || !sameGraveyardObject(z, item) {
		return nil
	}
	_, err := g.ReturnFromAnyZoneToHandForEffect(item.Controller, item.SourceCardID, item.SourceCardID)
	return err
}
