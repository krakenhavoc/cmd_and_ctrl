package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wormfang Crab — Creature — Nightmare Crab {3}{U}, 3/6 (Judgment):
//
//	"This creature can't be blocked.
//	 When this creature enters, an opponent chooses a permanent you control
//	 other than this creature and exiles it.
//	 When this creature leaves the battlefield, return the exiled card to the
//	 battlefield under its owner's control."
//
// Mist-Cloaked Herald's can't-be-blocked restriction. The opponent is picked
// by the Crab's controller (Burning-Rune Demon's ChoosePlayer), then that
// opponent picks the permanent, which is exiled linked to this object and
// comes back through the shared linked return (Wormfang Turtle's pair).
//
// No simplification.
func init() {
	const name = "Wormfang Crab"
	effects.Register(effects.Spec{
		OracleID:     "4216f9ce-bdd9-4213-8c3f-e2ef4cd848f9",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — an opponent chooses a permanent you control other than this creature and exiles it", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(ctx.Opponents()) == 0 {
					return nil
				}
				return effects.ChoosePlayer{
					Among:    effects.Opponents,
					Question: name + " — choose an opponent; they choose a permanent you control other than " + name + " and exile it",
					ThenKey:  wormfangCrabOpponent,
				}.Apply(ctx)
			}),
			effects.S82OldB626ReturnLinkedExiledCards(name+" — return the exiled card to the battlefield under its owner's control", false),
		},
	})
}

// wormfangCrabOpponent asks the chosen opponent for the permanent. An on-disk
// key: never renamed, never reused.
var wormfangCrabOpponent = effects.ChoosePlayerThen("option-pick/s82-old-b626-wormfang-crab-opponent", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.InstanceID != ctx.Source() {
			cands = append(cands, c.InstanceID)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chosen,
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Wormfang Crab — choose a permanent the Crab's controller controls other than the Crab: it is exiled",
		Cards:      cands,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    wormfangCrabPicked,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
})

// wormfangCrabPicked exiles the chosen permanent linked to the Crab. An
// on-disk key: never renamed, never reused.
var wormfangCrabPicked = effects.AskedPick("choose-cards/s82-old-b626-wormfang-crab-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 || !effects.OnBattlefield(ctx.Game, r.Picked[0]) {
		return nil
	}
	return effects.ExileTarget{Target: r.Picked[0], Linked: true}.Apply(ctx)
})
