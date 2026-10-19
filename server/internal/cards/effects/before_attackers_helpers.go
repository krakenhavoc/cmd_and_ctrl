package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// before_attackers_helpers.go - the small Effect bodies shared by the
// "Activate only during your turn, before attackers are declared" cards
// (#4062): Diaochan, Artful Beauty and the Portal Three Kingdoms
// creatures. Each is a body in the (g, item) shape an ActivatedAbility
// takes, reading its target at resolution the way CR 608.2b asks (a
// target that left is skipped, not an error).

// boostTargetUntilEOT is "target creature gets +P/+T until end of turn".
func boostTargetUntilEOT(label string, power, toughness int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return BoostUntilEOT{Target: id, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
	}
}

// targetCantBeBlockedThisTurn is "target creature can't be blocked this
// turn" (Zhuge Jin), Access Tunnel's body.
func targetCantBeBlockedThisTurn(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return RestrictUntilEOT{
			Target:       id,
			Restrictions: game.CantBeBlocked,
			Label:        name + " — can't be blocked",
		}.Apply(ctx)
	}
}

// gainOneLifeEffect is "you gain 1 life" (Temple Elder, Shu Farmer).
func gainOneLifeEffect(g *game.Game, item *game.StackItem) error {
	return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
}

// drawOneCardEffect is "draw a card" (Talas Researcher, Lu Su).
func drawOneCardEffect(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}

// targetOpponentDiscards is "target opponent discards N cards" (Cao
// Cao's two, Cunning Advisor's one): the opponent picks, a hand smaller
// than N gives up what it has (CR 701.9a).
func targetOpponentDiscards(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: t.ID,
				Source: item.SourceCardID,
				N:      n,
			})
			return nil
		}
		return nil
	}
}

// reanimateChosenUnderYourControl is "return target <creature> card from
// your graveyard to the battlefield" (Loyal Retainers): the announced
// card, if it is still in the graveyard, comes back under the
// ability's controller's control.
func reanimateChosenUnderYourControl(_ string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
		}
		return nil
	}
}

// norwoodPriestessPutsOne is "you may put a green creature card from
// your hand onto the battlefield".
func norwoodPriestessPutsOne(g *game.Game, item *game.StackItem) error {
	return PutFromHandOntoBattlefield{
		Match:    And(Creature(), OfColor("G")),
		Optional: true,
		Label:    "Norwood Priestess — you may put a green creature card from your hand onto the battlefield",
	}.Apply(NewContext(g, item))
}

// bounceThisAndAnotherTarget is "return this creature and another target
// creature to their owners' hands" (Coastal Wizard, Lady Sun). The
// source is returned only while it is the same object that activated
// the ability (BounceToHand's #1432 guard), and the target only if it
// is still legal; each half stands alone when the other is gone.
func bounceThisAndAnotherTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (BounceToHand{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil || t.ID == item.SourceCardID {
			continue
		}
		return BounceToHand{Target: t.ID}.Apply(ctx)
	}
	return nil
}

// diaochanDestroysTwo is "destroy target creature of your choice, then
// destroy target creature of an opponent's choice" (Diaochan, Artful
// Beauty): the two clauses in printed order, each skipped when its
// target is gone by resolution (CR 608.2b). The second may be the
// creature the first just destroyed, which is then not there to destroy.
func diaochanDestroysTwo(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for slot := 0; slot < 2; slot++ {
		t, ok := ctx.ClauseTarget(slot)
		if !ok || t.Kind != game.TargetCard {
			continue
		}
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
