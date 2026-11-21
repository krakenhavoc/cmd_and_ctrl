package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancestral Communion — {1}{G} Sorcery (Marvel Super Heroes Commander,
// #4455):
//
//	"When you cast this spell while you control your commander, copy
//	 this spell. You may choose a new target for the copy.
//	 Return target permanent card from your graveyard to your hand."
//
// A cast trigger of the spell itself (WhenYouCastThisSpell). "While you
// control your commander" is part of what triggers it; the wording has no
// "if", so CR 603.4's intervening-if recheck does not apply. The
// commander is checked once, as the spell is cast, and not again when the
// trigger resolves: a commander that left in between still gets the
// copy. The copy is CopySpell (Fin Fang Foom) on the spell itself, which
// asks for a new target. The spell's own ability is the shared
// graveyard-to-hand return.
//
// No simplification.
func init() {
	copyOnCast := effects.WhenYouCastThisSpell("Ancestral Communion — copy this spell; you may choose a new target for the copy",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			return effects.CopySpell{StackID: item.SourceCardID, ChooseNewTargets: true}.Apply(ctx)
		})
	copyOnCast.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return effects.Self(ev, source, ch, g) && effects.B18ControlsYourCommander(g, ev.Actor)
	}
	effects.Register(effects.Spec{
		OracleID:     "a45bb088-16fd-4d91-8fd6-3c52f1c24748",
		Name:         "Ancestral Communion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target permanent card in your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
		Triggered: []game.TriggeredAbility{copyOnCast},
	})
}
