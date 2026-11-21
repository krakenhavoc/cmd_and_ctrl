package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Time Lord Regeneration — Instant {U} (Doctor Who):
//
//	"Until end of turn, target Time Lord you control gains 'When this
//	 creature dies, reveal cards from the top of your library until you
//	 reveal a Time Lord creature card. Put that card onto the battlefield and
//	 the rest on the bottom of your library in a random order.'"
//
// The granted ability is a Spec.Grants bundle (Perigee Beckoner's shape) given
// to the target until end of turn; "your library" is the creature's
// controller's when it dies. The run reveals every card down to the first
// Time Lord creature card (Audacious Reshapers' reveal-and-put), which enters
// under that controller, and the rest go to the bottom in a random order.
//
// No simplification.
func init() {
	const grant = "time-lord-regeneration/dies-reveal-time-lord"
	Register(Spec{
		OracleID:     "a9a17d40-1675-49d9-9cbd-f6fe1b075d0c",
		Name:         "Time Lord Regeneration",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target Time Lord you control", YouControl(), gapB94TimeLordCard()),
		Grants: []AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Time Lord Regeneration — reveal until a Time Lord creature card; put it onto the battlefield and the rest on the bottom in a random order",
					timeLordRegenerationReveal),
			},
			Text: "When this creature dies, reveal cards from the top of your library until you reveal a Time Lord creature card. Put that card onto the battlefield and the rest on the bottom of your library in a random order.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return GrantAbilitiesFor{
					Target: t.ID,
					Keys:   []string{grant},
					Label:  "Time Lord Regeneration — gains \"When this creature dies, reveal until a Time Lord creature card\"",
				}.Apply(ctx)
			}
			return nil
		},
	})
}

func timeLordRegenerationReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() && gapB94IsTimeLord(c) },
		"Time Lord Regeneration — revealed until a Time Lord creature card")
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:      true,
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}
