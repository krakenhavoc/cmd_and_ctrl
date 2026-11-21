package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mausoleum Wanderer — Creature — Spirit {U}, 1/1 (Innistrad Remastered,
// S82 GAP, #5948):
//
//	"Flying
//	 Whenever another Spirit you control enters, this creature gets +1/+1
//	 until end of turn.
//	 Sacrifice this creature: Counter target instant or sorcery spell
//	 unless its controller pays {X}, where X is this creature's power."
//
// The pump is a Layer 7c effect on the Wanderer for each Spirit that enters
// under your control (a changeling counts). The sacrifice is a cost, so X is
// the Wanderer's power as it last existed on the battlefield, pump included
// (CR 608.2h); at power 0 the tax is {0} and nothing is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1803850e-9499-4bf7-a4d9-6bf655cbe87d",
		Name:            "Mausoleum Wanderer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.HasSubtype(source.TextSubtype("Spirit"))
			}, "Mausoleum Wanderer — this creature gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Mausoleum Wanderer — +1/+1 until end of turn")),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Counter target instant or sorcery spell unless its controller pays {X}, where X is this creature's power.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    SacrificeThis(),
			Targets: TargetSpell("target instant or sorcery spell", Or(Instant(), Sorcery())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok || info.Power <= 0 || len(item.Targets) == 0 {
					return nil
				}
				cost := "{" + strconv.Itoa(info.Power) + "}"
				return CounterUnlessPaid{
					StackID:  item.Targets[0].ID,
					Cost:     cost,
					Question: "Mausoleum Wanderer — pay " + cost + " or your spell is countered",
				}.Apply(ctx)
			},
		}},
	})
}
