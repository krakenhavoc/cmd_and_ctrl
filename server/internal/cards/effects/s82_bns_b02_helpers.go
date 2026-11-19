package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hinderingLightTargetsYouOrYourPermanent is Hindering Light's clause:
// the candidate spell has at least one target that is the caster or a
// permanent the caster controls on the battlefield.
func hinderingLightTargetsYouOrYourPermanent(g *game.Game, caster uuid.UUID, c game.Card) bool {
	item := g.StackItemForEffect(c.InstanceID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		switch t.Kind {
		case game.TargetPlayer:
			if t.ID == caster {
				return true
			}
		case game.TargetCard:
			z := g.FindCardZoneForEffect(t.ID)
			if z == nil || z.Kind != game.ZoneBattlefield {
				continue
			}
			if tc, ok := g.LookupCardForEffect(t.ID); ok && tc.Controller == caster {
				return true
			}
		}
	}
	return false
}

// grindstoneMillPass is one pass of Grindstone's "mills two cards. If two
// cards that share a color were milled this way, repeat this process":
// the repeat is recursion through the mill's continuation. The guard
// is carried by value and bounds a pass chain that never ends.
func grindstoneMillPass(ctx *Context, victim uuid.UUID, guard int) error {
	if guard >= 200 {
		return nil
	}
	return MillToZone{
		Player:   victim,
		N:        2,
		ThenTail: withN(withCarry(ctx.ItemTail(grindstoneAfterMill), victim), guard),
	}.Apply(ctx)
}

// grindstoneAfterMill is the clause after one pass, an on-disk key: the
// pass repeats when exactly two cards were milled and they share a color.
var grindstoneAfterMill game.TailRef

func init() {
	grindstoneAfterMill = CardTail("s82-bns-b02/grindstone-after-mill", func(ctx *Context, o game.TailOutcome) error {
		if !s82BNSb02TwoCardsShareAColor(ctx.Game, o.IDs) {
			return nil
		}
		return grindstoneMillPass(ctx, tailCarry0(o), o.Tail.N+1)
	})
}

// s82BNSb02TwoCardsShareAColor reports whether the two milled cards have
// a color in common, read wherever they now are. A colorless card (a
// land, an artifact) shares none (CR 105.2c).
func s82BNSb02TwoCardsShareAColor(g *game.Game, milled []uuid.UUID) bool {
	if len(milled) != 2 {
		return false
	}
	a, okA := g.LookupCardForEffect(milled[0])
	b, okB := g.LookupCardForEffect(milled[1])
	if !okA || !okB {
		return false
	}
	for _, col := range a.EffectiveColors() {
		if b.HasColor(col) {
			return true
		}
	}
	return false
}

// printedBnsHumanRogueToken is Outlaws' Merriment's 1/2 red and white
// Human Rogue token with haste and "When this creature enters the
// battlefield, it deals 1 damage to any target." The slug is an on-disk
// identity.
func printedBnsHumanRogueToken() tokenTemplate {
	return tokenTemplate{
		Slug: "human-rogue-haste-enters-ping",
		Card: game.Card{
			Name:      "Human Rogue",
			TypeLine:  "Token Creature — Human Rogue",
			Power:     1,
			Toughness: 2,
			Colors:    []string{"R", "W"},
			Keywords:  []string{"haste"},
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Human Rogue — 1 damage to any target",
				sourceDealsDamageToEachLegalTarget(1)), TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
		Text: "Haste\nWhen this creature enters the battlefield, it deals 1 damage to any target.",
	}
}
