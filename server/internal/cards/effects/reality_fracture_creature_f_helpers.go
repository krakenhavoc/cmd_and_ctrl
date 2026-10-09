package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Helpers for Reality Fracture slice fra-creature-f. Every name carries
// the rfCreatureF prefix so a parallel slice cannot collide with it.

// SculptureTreasureToken is Vraska, Soul of Stone's token: a 1/1
// colourless Sculpture Treasure artifact creature with "{T}, Sacrifice
// this token: Add one mana of any color."
func SculptureTreasureToken() game.Card { return tokenFromCatalog(printedSculptureTreasureToken) }

// printedSculptureTreasureToken is that token as PRINTED, mana ability
// included. It is a creature, so (unlike a plain Treasure) its tap cost
// waits out summoning sickness (CR 302.6) and it is an artifact creature
// for every effect that counts either.
func printedSculptureTreasureToken() tokenTemplate {
	return tokenTemplate{
		Slug: "sculpture-treasure",
		Card: game.Card{
			Name:      "Sculpture Treasure",
			TypeLine:  "Token Artifact Creature — Sculpture Treasure",
			Power:     1,
			Toughness: 1,
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:       true,
			SacrificeCost: true,
			Produced:      "{W|U|B|R|G}",
			Label:         "{T}, Sacrifice: Add one mana of any color",
		}},
		Text: "{T}, Sacrifice this token: Add one mana of any color.",
	}
}

// rfCreatureFCastNoncreatureSpellThisTurn is "if you've cast a
// noncreature spell this turn" as a cost predicate. The tally is bumped
// after a cast succeeds, so the spell being priced is not in it.
func rfCreatureFCastNoncreatureSpellThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.CastTallyFor(q.Controller).Noncreature > 0
	}
}

// rfCreatureFArtifactOrCreature matches the spells Traxos, Scourge
// Eternal untaps for.
func rfCreatureFArtifactOrCreature() CardPredicate {
	return Or(Artifact(), Creature())
}

// rfCreatureFCountersOnTarget is "put N +1/+1 counters on target
// creature" as a targeted trigger's effect: the pick is re-read as legal
// at resolution (CR 608.2b).
func rfCreatureFCountersOnTarget(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
		}
		return nil
	}
}

// rfCreatureFControlsSixLands is Vraska, the Cutting Glare's
// "if you control six or more lands".
func rfCreatureFControlsSixLands(g *game.Game, controller uuid.UUID) bool {
	return b02CountLandsControlledBy(g, controller) >= 6
}

// rfCreatureFEntersWithSixLands is the trigger half of that
// intervening if (CR 603.4): it only triggers when the condition holds.
func rfCreatureFEntersWithSixLands(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID == source.InstanceID && rfCreatureFControlsSixLands(g, source.Controller)
}

// rfCreatureFDestroyThenTreasure is Vraska's body: re-check the
// condition (CR 603.4), destroy the target, and its controller — read
// before the destruction — creates a Treasure whether or not it died.
func rfCreatureFDestroyThenTreasure(g *game.Game, item *game.StackItem) error {
	if !rfCreatureFControlsSixLands(g, item.Controller) {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		// A legal target is on the battlefield, so its controller is
		// always found; the Treasure is theirs whether or not it died.
		victim, _ := controllerOfTarget(ctx, t.ID)
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return CreateToken{Controller: victim, Template: TreasureToken(), N: 1}.Apply(ctx)
	}
	return nil
}

// rfCreatureFCombat is true in any of the five combat steps.
func rfCreatureFCombat(g *game.Game) bool {
	return g != nil && game.PhaseOf(g.Turn.Step) == game.PhaseCombat
}

// rfCreatureFAttackedAPlayerAlone is Yuriko's trigger condition: a
// creature its controller controls was declared as an attacker, it is
// the only attacking creature (CR 506.5), and what it attacks is a
// player rather than a planeswalker or battle.
func rfCreatureFAttackedAPlayerAlone(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.Actor != source.Controller {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n == 1
}

// rfCreatureFCopyTargetSpellTwice is Venser's first mode.
func rfCreatureFCopyTargetSpellTwice(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return CopySpell{
		StackID:          t.ID,
		Controller:       item.Controller,
		Count:            2,
		ChooseNewTargets: true,
	}.Apply(ctx)
}

// rfCreatureFTwoHastyCopiesSacrificedAtEnd is Venser's second mode: two
// token copies with haste, sacrificed at the beginning of the next end
// step (CR 603.7).
func rfCreatureFTwoHastyCopiesSacrificedAtEnd(item *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	cursor := b25LastEventSeq(ctx.Game)
	if err := (CreateTokenCopy{
		Controller: item.Controller,
		Copy:       t.ID,
		N:          2,
		Except:     TokenCopyGainsHaste,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(ctx.Game, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Venser — sacrifice the tokens",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}

// rfCreatureFCounterOnTargetCreature is the {6} counter-giving
// activations of the two Yoshimarus.
func rfCreatureFCounterOnTargetCreature(g *game.Game, item *game.StackItem) error {
	return rfCreatureFCountersOnTarget(1)(g, item)
}
