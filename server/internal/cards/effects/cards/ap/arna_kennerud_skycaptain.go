package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arna Kennerüd, Skycaptain — Legendary Creature — Human Knight
// {2}{W}{U}{B}, 4/4:
//
//	"Flying, lifelink
//	 Ward—Discard a card.
//	 Whenever a modified creature you control attacks, double the number
//	 of each kind of counter on it. Then for each nontoken permanent
//	 attached to it, create a token that's a copy of that permanent
//	 attached to that creature."
//
// "Modified" (CR 700.15) is judged as the creature is declared as an
// attacker. As the ability resolves, each kind of counter on the attacker
// is doubled (Vorel's B30DoubleEachCounterKindOn), then each nontoken
// permanent attached to it then is copied (CR 707.2) and the copy enters
// attached to it, unless that permanent could not be attached to it
// (CR 303.4j). An attacker that has left the battlefield gets nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "436e7548-f8d5-43af-947d-b382da0d19b3",
		Name:            "Arna Kennerüd, Skycaptain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardDiscard(1), "Arna Kennerüd — ward, discard a card"),
			effects.On(game.EventAttack, arnaModifiedCreatureYouControlAttacked,
				"Arna Kennerüd — double each kind of counter on it, then copy each nontoken permanent attached to it",
				arnaDoubleAndCopy),
		},
	})
}

func arnaModifiedCreatureYouControlAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && effects.B07IsModified(g, c)
}

func arnaDoubleAndCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	attacker := ctx.Trigger().Event.CardID
	if !effects.OnBattlefield(g, attacker) {
		return nil
	}
	if err := effects.B30DoubleEachCounterKindOn(ctx, attacker); err != nil {
		return err
	}
	var attached []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsAttachedTo(attacker) && !effects.IsToken(c) {
			attached = append(attached, c.InstanceID)
		}
	}
	for _, id := range attached {
		ref, ok := g.PermanentRefForEffect(id)
		if !ok {
			continue
		}
		tmpl, ok := g.TokenCopyTemplateForObjectForEffect(ref)
		if !ok {
			continue
		}
		perm, _ := g.LookupCardForEffect(id)
		if !effects.OnBattlefield(g, attacker) || !effects.CanBeAttachedTo(g, ctx.Controller(), perm, attacker) {
			continue
		}
		if _, err := g.CreateTokensForEffect(ctx.Controller(), effects.CopiedToken(tmpl), 1, game.TokenEntryOptions{
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: attacker},
		}); err != nil {
			return err
		}
	}
	return nil
}
