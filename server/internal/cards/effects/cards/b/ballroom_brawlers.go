package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ballroom Brawlers — Creature — Human Warrior {3}{W}{W}, 3/5 (Streets of
// New Capenna, #5491):
//
//	"Whenever this creature attacks, this creature and up to one other
//	 target creature you control both gain your choice of first strike or
//	 lifelink until end of turn."
//
// One trigger, one choice: the controller picks first strike or lifelink
// once as the trigger resolves, and both creatures gain that keyword
// (Practiced Offense's PickOption shape). "Up to one" means the trigger
// still goes on the stack with no target and then affects only the
// Brawlers. A target that has left the battlefield or become illegal is
// skipped alone (CR 608.2b); the Brawlers still gain the keyword if they
// are on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c87dff29-fa4c-44f4-9de6-d85a347c1720",
		Name:         "Ballroom Brawlers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Ballroom Brawlers — this creature and up to one other target creature you control both gain your choice of first strike or lifelink until end of turn",
					ballroomBrawlersAsk),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature you control", effects.YouControl())),
			),
		},
	})
}

func ballroomBrawlersAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	carry := []uuid.UUID{item.SourceCardID}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			carry = append(carry, t.ID)
		}
	}
	return effects.PickOption{
		Question: "Ballroom Brawlers — both creatures gain which until end of turn?",
		Options:  []game.ChoiceOption{{Label: "First strike"}, {Label: "Lifelink"}},
		ThenKey:  ballroomBrawlersKeyword,
		Carry:    carry,
	}.Apply(ctx)
}

// ballroomBrawlersKeyword is the keyword chosen, keyed (#4253): the two
// creatures ride the prompt as data. An on-disk key: never renamed,
// never reused.
var ballroomBrawlersKeyword = effects.OptionPickThen("option-pick/ballroom-brawlers-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	keywords := []string{"first strike", "lifelink"}
	if r.Index < 0 || r.Index >= len(keywords) {
		return nil
	}
	for _, id := range r.Carry {
		if id == uuid.Nil || !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{keywords[r.Index]},
			Label:    "Ballroom Brawlers — " + keywords[r.Index],
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
