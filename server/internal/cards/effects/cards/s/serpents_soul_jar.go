package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serpent's Soul-Jar — Artifact {2}{B} (Kaldheim Commander, #5506):
//
//	"Whenever an Elf you control dies, exile it.
//	 {T}, Pay 2 life: Until end of turn, you may cast a creature spell from
//	 among cards exiled with this artifact."
//
// The trigger exiles the dead Elf card from the graveyard LINKED to this
// artifact (CR 607.2a, Ian Malcolm's link), if it is still there; a token
// is gone already. The ability grants, until end of turn, a standing
// permission over the cards linked to THIS object, creature cards only,
// cast only (CR 601), paid for normally; so a Soul-Jar that left and came
// back has no claim on its old cards (CR 400.7), and a creature exiled later
// the same turn is castable too. The life is a cost (CR 119.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d451958-25f3-4e70-8be2-c832f60f39e9",
		Name:         "Serpent's Soul-Jar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.S82OldB074ElfYouControlDied(ev, source, g)
			}, "Serpent's Soul-Jar — exile it", serpentsSoulJarExile),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay 2 life: Until end of turn, you may cast a creature spell from among cards exiled with this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.PayLife(2)),
			Effect:  serpentsSoulJarGrant,
		}},
	})
}

// serpentsSoulJarExile exiles the Elf that died, linked to this Soul-Jar.
func serpentsSoulJarExile(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	src, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	return g.ExileLinkedForEffect(src, id)
}

// serpentsSoulJarGrant is the activated ability: a creature-spell cast
// permission until end of turn over the cards exiled with this artifact.
func serpentsSoulJarGrant(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	g.GrantCastPermissionForEffect(game.CastPermission{
		Player:   item.Controller,
		Zone:     game.ZoneExile,
		Scope:    game.ScopeStanding,
		Source:   item.SourceCardID,
		Filter:   game.PermissionFilter{ExiledWithSource: true, ExiledWith: game.PermissionCardRef(ref), CreatureOnly: true},
		CastOnly: true,
		Label:    "Serpent's Soul-Jar — cast a creature spell from among cards exiled with it",
	})
	return nil
}
