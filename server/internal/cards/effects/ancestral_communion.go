package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancestral Communion — {1}{G} Sorcery (Marvel Super Heroes Commander,
// #4455):
//
//	"When you cast this spell while you control your commander, copy
//	 this spell. You may choose a new target for the copy.
//	 Return target permanent card from your graveyard to your hand."
//
// A cast trigger of the spell itself (WhenYouCastThisSpell) with an
// intervening if (CR 603.4): the commander is checked as the spell is
// cast, so the trigger is never put on the stack otherwise, and again
// as it resolves. The copy is CopySpell (Fin Fang Foom) on the spell
// itself, which asks for a new target. The spell's own ability is the
// shared graveyard-to-hand return.
//
// No simplification.
func init() {
	copyOnCast := WhenYouCastThisSpell("Ancestral Communion — copy this spell; you may choose a new target for the copy",
		func(g *game.Game, item *game.StackItem) error {
			if !b18ControlsYourCommander(g, item.Controller) {
				return nil
			}
			ctx := NewContext(g, item)
			return CopySpell{StackID: item.SourceCardID, ChooseNewTargets: true}.Apply(ctx)
		})
	copyOnCast.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return Self(ev, source, ch, g) && b18ControlsYourCommander(g, ev.Actor)
	}
	Register(Spec{
		OracleID:     "a45bb088-16fd-4d91-8fd6-3c52f1c24748",
		Name:         "Ancestral Communion",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card in your graveyard", YouOwn(), Permanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
		Triggered: []game.TriggeredAbility{copyOnCast},
	})
}
