package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "cb905eb0-bb20-41c0-9695-0f5c71972561",
		Name:         "Chainer, Dementia Master",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{TribalAnthem(TribeFilter{Tribes: []string{"Nightmare"}}, 1, 1)},
		Activated: []ActivatedAbility{{
			Label:   "{B}{B}{B}, Pay 3 life: Put target creature card from a graveyard onto the battlefield under your control. That creature is black and is a Nightmare in addition to its other creature types.",
			Cost:    Plus(ManaCost("{B}{B}{B}"), PayLife(3)),
			Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			Effect:  chainerReanimate,
		}},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				WhenThisLeaves("Chainer, Dementia Master — exile all Nightmares",
					Do(ExileAllMatching{Match: HasSubtype("Nightmare")})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}}),
		},
	})
}

func chainerReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    ids,
		Controller: ctx.Controller(),
		ThenTail:   chainerNightmares,
	}.Apply(ctx)
}

// chainerNightmares is "that creature is black and is a Nightmare in addition
// to its other creature types", pinned to the object that entered. An on-disk
// key: never renamed, never reused.
var chainerNightmares = TogetherTail("chainer-dementia-master/black-nightmares",
	func(ctx *Context, entered []uuid.UUID, _ int) error {
		for _, id := range entered {
			if err := (ScopedEffectFor{
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
