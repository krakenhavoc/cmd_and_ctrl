package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hellkite Courser — Creature — Dragon {4}{R}{R}, 6/5 (Tarkir:
// Dragonstorm Commander):
//
//	"Flying
//	 When this creature enters, you may put a commander you own from the
//	 command zone onto the battlefield. It gains haste. Return it to the
//	 command zone at the beginning of the next end step."
//
// "You may" is a choose_cards prompt with a floor of zero over the
// commanders you own in your command zone (nothing is asked with none).
// The commander is PUT onto the battlefield, which is not casting it: no
// commander tax (the printed ruling), through the same door commander
// ninjutsu uses. Haste is a layer-6 grant pinned to that permanent with no
// stated duration. The return is a delayed trigger at the next end step,
// which moves it to the command zone only if it is still on the battlefield
// (game.ReturnToCommandZoneForEffect); one that went elsewhere stays there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c5abd58f-3e00-447b-a978-64cb88ad501b",
		Name:            "Hellkite Courser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hellkite Courser — you may put a commander you own from the command zone onto the battlefield; it gains haste; return it to the command zone at the beginning of the next end step",
				hellkiteCourserAsk),
		},
	})
}

// hellkiteCourserAsk offers the commanders the controller owns in their
// command zone.
func hellkiteCourserAsk(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Command == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Command.Cards {
		if c.IsCommander && c.Owner == item.Controller && !c.IsToken() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Hellkite Courser — you may put a commander you own from the command zone onto the battlefield",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneCommand,
		ThenKey:    hellkiteCourserPicked,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// hellkiteCourserPicked puts the chosen commander onto the battlefield,
// gives it haste and schedules its return. An on-disk key.
var hellkiteCourserPicked = askedPickRule("choose-cards/hellkite-courser-commander", nil,
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		entered, err := ctx.Game.PutFromCommandZoneOntoBattlefieldForEffect(r.Picked[0], game.ZoneEntryOptions{Controller: ctx.Controller()})
		if err != nil || entered == uuid.Nil {
			return err
		}
		if err := (ScopedEffectFor{
			Target:   entered,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), entered),
			Label:    "Hellkite Courser — it gains haste",
		}).Apply(ctx); err != nil {
			return err
		}
		return ScheduleDelayedTrigger{
			Label: "Hellkite Courser — return the commander to the command zone",
			Cards: []uuid.UUID{entered},
			Body:  cmdrB19ReturnListedToCommandZoneBody,
		}.Apply(ctx)
	})
