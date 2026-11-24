package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malik, Grim Manipulator — Legendary Creature — Human Rogue {2}{R}{W}{B},
// 4/3 (Secret Lair Drop):
//
//	"When Malik, Grim Manipulator enters, you and target opponent each
//	 secretly choose a creature that player controls. Then those choices are
//	 revealed, and that player sacrifices those creatures.
//	 Whenever an opponent sacrifices a creature, you create a Treasure token."
//
// The secret choice is Prisoner's Dilemma's machinery (Vote with Choice and
// Secret, one hidden prompt per chooser, one reveal): the options are the
// target opponent's creatures when the ability resolves, and the voters are
// the controller and that opponent. A creature both picked is one creature
// (one sacrifice); two different ones are both sacrificed. The opponent
// sacrifices them, so the second ability sees an opponent's sacrifice (Malik
// himself is a creature that player may have picked, and the Treasure is
// made for each creature sacrificed).
func init() {
	enters := effects.WhenThisEnters("Malik, Grim Manipulator — you and target opponent each secretly choose a creature that player controls; that player sacrifices those creatures",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			var victim uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					victim = t.ID
				}
			}
			if victim == uuid.Nil {
				return nil
			}
			var ids []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == victim && c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			return effects.Vote{
				Question: "Malik, Grim Manipulator — secretly choose a creature the target opponent controls",
				Options:  effects.VoteForPermanents(g, ids),
				Voters:   []uuid.UUID{item.Controller, victim},
				Choice:   true,
				Secret:   true,
				Carry:    []uuid.UUID{victim},
				Then:     malikChosen,
			}.Apply(ctx)
		})
	enters.Targets = effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:     "416608a1-89a4-44ce-8fd0-fab6b023d3d4",
		Name:         "Malik, Grim Manipulator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			enters,
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventSacrifice || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
					return false
				}
				c, _ := g.LookupCardForEffect(ev.CardID)
				return effects.LeftAsType(ev, c, "creature")
			}, "Malik, Grim Manipulator — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}

// malikChosen is "those choices are revealed, and that player sacrifices
// those creatures": every option that got a ballot, once each, still on the
// battlefield under that player's control.
var malikChosen = effects.VoteResultThen("vote/malik-grim-manipulator", func(ctx *effects.Context, r game.VoteResult) error {
	var victim uuid.UUID
	if len(r.Carry) > 0 {
		victim = r.Carry[0]
	}
	for i := range r.Options {
		if r.Votes(i) == 0 {
			continue
		}
		id := effects.VotedPermanent(r, i)
		if id == uuid.Nil {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Controller != victim || !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := ctx.Game.SacrificePermanentForEffect(id); err != nil {
			return err
		}
	}
	return nil
})
