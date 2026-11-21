package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chainer, Dementia Master — Legendary Creature — Human Minion {3}{B}{B},
// 3/3 (Commander Masters):
//
//	"All Nightmares get +1/+1.
//	 {B}{B}{B}, Pay 3 life: Put target creature card from a graveyard onto
//	 the battlefield under your control. That creature is black and is a
//	 Nightmare in addition to its other creature types.
//	 When Chainer leaves the battlefield, exile all Nightmares."
//
// The anthem is every Nightmare's, any player's (TribalAnthem without
// YoursOnly). The reanimation is Liliana, Death's Majesty's: the card enters
// under the activator's control (any graveyard) and a tail pins a layer-5
// effect to the new object with no end, which sets its colour to black (not
// "in addition", the printed text) and adds Nightmare. The trigger exiles
// every Nightmare permanent as Chainer leaves, itself not being one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb905eb0-bb20-41c0-9695-0f5c71972561",
		Name:         "Chainer, Dementia Master",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Nightmare"}}, 1, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}{B}, Pay 3 life: Put target creature card from a graveyard onto the battlefield under your control. That creature is black and is a Nightmare in addition to its other creature types.",
			Cost:    effects.Plus(effects.ManaCost("{B}{B}{B}"), effects.PayLife(3)),
			Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			Effect:  chainerReanimate,
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenThisLeaves("Chainer, Dementia Master — exile all Nightmares",
					effects.Do(effects.ExileAllMatching{Match: effects.HasSubtype("Nightmare")})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}}),
		},
	})
}

func chainerReanimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ids := effects.LegalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyardTogether{
		Targets:    ids,
		Controller: ctx.Controller(),
		ThenTail:   chainerNightmares,
	}.Apply(ctx)
}

// chainerNightmares is "that creature is black and is a Nightmare in addition
// to its other creature types", pinned to the object that entered. An on-disk
// key: never renamed, never reused.
var chainerNightmares = effects.TogetherTail("chainer-dementia-master/black-nightmares",
	func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
		for _, id := range entered {
			if err := (effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Nightmare")...), game.SetColorsMod(ctx.TextColors("B")...)},
				Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
				Label:    "Chainer, Dementia Master — it is black and is a Nightmare in addition to its other creature types",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
