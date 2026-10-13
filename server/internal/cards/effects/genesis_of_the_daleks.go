package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Genesis of the Daleks — Enchantment — Saga {4}{B}{B}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III — Create a 3/3 black Dalek artifact creature token with
//	 menace for each lore counter on Genesis of the Daleks.
//	 IV — Target opponent faces a villainous choice — Destroy all Dalek
//	 creatures and each of your opponents loses life equal to the total
//	 power of Daleks that died this turn, or destroy all non-Dalek
//	 creatures."
//
// Chapters I to III count the lore counters as they resolve, read off
// the Saga as it is then (or as it last was, if it has gone, CR 608.2h),
// so a chapter that was skipped over still makes one Dalek per counter.
//
// Chapter IV is a villainous choice (CR 701.55, #3143) faced by the
// target opponent. Option A is two instructions in the printed order:
// the sweep first, then the life loss, so the Daleks it destroyed are
// among "the Daleks that died this turn" (Game.DiedPowerWithSubtypeThisTurn,
// each as it last was on the battlefield, tokens and counters included).
// The loss is from EACH of the controller's opponents, not only the one
// who chose, and is read after the sweep has settled (a commander
// answering a command-zone prompt delays it). "Dalek creatures" is
// OfCreatureType, which reads Card.HasSubtype, so a changeling is a
// Dalek. Option B is the sweep of everything else. Neither option
// targets, so hexproof and shroud do not matter; indestructible
// survives both (CR 701.8).
//
// No simplifications.
var genesisChoice = NewVillainousChoice("genesis-of-the-daleks",
	"Genesis of the Daleks — destroy all Daleks and drain by their power, or destroy all non-Daleks?",
	VillainousOption{
		Label: "Destroy all Dalek creatures; each opponent of the controller loses life equal to the total power of Daleks that died this turn",
		Harm:  2,
		Body: func(ctx *Context, _ Facing) error {
			return DestroyAllMatching{
				Match: genesisDalekCreature(),
				Then: func(ctx *Context, _ []game.Card, _ int) error {
					n := ctx.Game.DiedPowerWithSubtypeThisTurn("Dalek")
					if n <= 0 {
						return nil
					}
					return ctx.Game.LoseLifeEachThenForEffect(ctx.Source(), ctx.Opponents(), n, nil)
				},
			}.Apply(ctx)
		},
	},
	VillainousOption{
		Label: "Destroy all non-Dalek creatures",
		Harm:  3,
		Body: func(ctx *Context, _ Facing) error {
			return DestroyAllMatching{Match: And(Creature(), Not(genesisDalekCreature()))}.Apply(ctx)
		},
	})

// genesisDalekCreature is "Dalek creatures".
func genesisDalekCreature() CardPredicate { return OfCreatureType("Dalek") }

func init() {
	const name = "Genesis of the Daleks"
	daleksForLore := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		lore := 0
		if info, ok := ctx.SourcePermanent(); ok {
			lore = info.Counters[game.CounterLore]
		}
		return CreateToken{
			Controller: item.Controller,
			Template:   TokenCard("3/3 black Dalek artifact with menace"),
			N:          lore,
		}.Apply(ctx)
	}
	Register(Spec{
		OracleID:     "71ff4603-9f7a-459e-b5b3-9b39583f1dc7",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create a Dalek for each lore counter"), daleksForLore),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "create a Dalek for each lore counter"), daleksForLore),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "create a Dalek for each lore counter"), daleksForLore),
			TriggerWithPurpose(ChapterTriggerTargeting(4, SagaChapterLabel(name, 4, "target opponent faces a villainous choice"),
				TargetPlayer("target opponent", Opponent()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var facing []uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							facing = append(facing, t.ID)
						}
					}
					return FaceVillainousChoice{Players: facing, Choice: genesisChoice}.Apply(ctx)
				}),
				// Either option sweeps creatures and neither sweeps all of them
				// (ADR 0126 §6, owner decision 2).
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		},
	})
}
