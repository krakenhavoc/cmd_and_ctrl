package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helvault — Legendary Artifact {3} (Innistrad Remastered, #5948):
//
//	"{1}, {T}: Exile target creature you control.
//	 {7}, {T}: Exile target creature you don't control.
//	 When Helvault is put into a graveyard from the battlefield, return all cards
//	 exiled with it to the battlefield under their owners' control."
//
// Both exiles are Fear of Abduction's linked exile (effects.FearOfAbductionExile): linked to the Helvault as an object (ExileTarget.Linked, CR
// 607.2a), so the link is its battlefield life's: the same card in a later life
// has no claim (CR 400.7). The return is one simultaneous entry of every card
// still in exile with that link, read from the Helvault's last-known object as
// the trigger resolves, under its owner's control each; a token never comes back
// (CR 111.8). It triggers only on going to a graveyard from the battlefield:
// bounced, exiled or put into a library, the Helvault returns nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa3d8a6f-dc5b-4890-a806-f785a41660c7",
		Name:         "Helvault",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Exile target creature you control.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  effects.FearOfAbductionExile,
			},
			{
				Label:   "{7}, {T}: Exile target creature you don't control.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{7}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				Effect:  effects.FearOfAbductionExile,
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ev.NewZone == game.ZoneGraveyard && effects.Self(ev, source, ch, g)
			}, "Helvault — return all cards exiled with it to the battlefield under their owners' control", gapB07HelvaultReturn),
		},
	})
}

// gapB07HelvaultReturn returns every card exiled with the Helvault as it last
// existed, all at once, to the battlefield under their owners' control.
func gapB07HelvaultReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{Targets: ids}.Apply(ctx)
}
