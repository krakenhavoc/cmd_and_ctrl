package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dino DNA — Artifact {1} (Jurassic World Collection, #5948):
//
//	"Imprint — {1}, {T}: Exile target creature card from a graveyard.
//	 Activate only as a sorcery.
//	 {6}: Create a token that's a copy of target creature card exiled with
//	 this artifact, except it's a 6/6 green Dinosaur creature with trample.
//	 Activate only as a sorcery."
//
// The imprint exile is a linked exile (ExileTarget.Linked, CR 607.2a), so the
// second ability can name "creature card exiled with this artifact" as a
// target judged against this object (ExiledWithThis); a second Dino DNA's
// exiles are not offered. The token copies the card's copiable values as it
// sits in exile (CR 707.12), except its power and toughness are 6/6 (CR
// 707.9a), its colour is green only, its creature types are Dinosaur alone
// and it also has trample (CR 707.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26e988da-52cf-4382-8646-a27c5ad1b92d",
		Name:         "Dino DNA",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:        "Imprint — {1}, {T}: Exile target creature card from a graveyard. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				SorcerySpeed: true,
				Targets:      effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
				Effect:       effects.AvB7ExileTargetLinked, // the shared linked-exile body (avatar_b7_helpers.go)
			},
			{
				Label:        "{6}: Create a token that's a copy of target creature card exiled with this artifact, except it's a 6/6 green Dinosaur creature with trample. Activate only as a sorcery.",
				Cost:         effects.ManaCost("{6}"),
				SorcerySpeed: true,
				Targets:      effects.ExiledWithThis(effects.TargetCardInExile("target creature card exiled with Dino DNA", effects.Creature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.CreateTokenCopy{
							Controller: ctx.Controller(),
							Copy:       t.ID,
							N:          1,
							Except: func(tok *game.Card) {
								tok.Power, tok.Toughness = 6, 6
								tok.VariableToughness = false
								tok.PrintedPTKnown = true
								tok.SetCopyExceptionColors("G")
								tok.TypeLine = effects.S82MidB11TypeLineWithCreatureType(tok.TypeLine, "Dinosaur")
								tok.Keywords = append(tok.Keywords, "trample")
							},
						}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}
