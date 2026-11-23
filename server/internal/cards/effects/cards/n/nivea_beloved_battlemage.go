package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nivea, Beloved Battlemage — Legendary Creature — Human Wizard Warrior
// {2}{W}, 3/1 (Mystery Booster Commander, #5948):
//
//	"Whenever a creature you control becomes blocked, you may untap it and
//	 remove it from combat.
//	 {5}{W}{W}{W}, Sacrifice Nivea: Create an Akroma, Angel of Wrath token.
//	 (She's a {5}{W}{W}{W} legendary 6/6 Angel creature with flying, first
//	 strike, vigilance, trample, haste, and protection from black and from
//	 red.)"
//
// The trigger is one occurrence per blocked attacker (EventBecomesBlocked,
// CR 506.4), General Marhault's watch. The "may" is the trigger's own prompt;
// untapping and removing from combat (CR 506.4) use the attacker's object
// reference, so a creature that has left does nothing. The sacrifice is a cost,
// so the token is made even though Nivea is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a5376aa-8996-476e-9ee8-1426b0944091",
		Name:         "Nivea, Beloved Battlemage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCreature() && c.Controller == source.Controller
				}, "Nivea, Beloved Battlemage — untap that creature and remove it from combat", niveaUntapAndRemove),
				"Nivea, Beloved Battlemage — untap that creature and remove it from combat?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{W}{W}{W}, Sacrifice Nivea: Create an Akroma, Angel of Wrath token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}{W}{W}{W}"), effects.SacrificeThis()),
			Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("6/6 white Legendary Angel named Akroma, Angel of Wrath with flying, " +
				"first strike, vigilance, trample, haste and protection from black and from red"), N: 1}),
		}},
	})
}

// niveaUntapAndRemove untaps the creature that became blocked, then removes it
// from combat. A creature that has left the battlefield or is no longer in
// combat is skipped.
func niveaUntapAndRemove(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left {
		return nil
	}
	id := ctx.Trigger().Object.Ref().ID
	if id == uuid.Nil || !effects.OnBattlefield(g, id) {
		return nil
	}
	if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	if err := g.RemoveFromCombatForEffect(id); err != nil && err != game.ErrNotInCombat {
		return err
	}
	return nil
}
