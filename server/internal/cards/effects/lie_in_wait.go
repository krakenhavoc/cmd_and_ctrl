package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lie in Wait — Sorcery {B}{G}{U} (Tarkir: Dragonstorm, #4463):
//
//	"Return target creature card from your graveyard to your hand. Lie in
//	 Wait deals damage equal to that card's power to target creature."
//
// Two target clauses (slot 0 the card in your graveyard, slot 1 the
// creature on the battlefield), so the spell can't be cast without both
// (the rulings). The card returns first if it is still a legal target;
// the damage is the power of that card, read once it is in your hand
// (printed power, as a card in hand has no counters or effects), or from
// where it last was if it left the graveyard in response (CR 608.2h). The
// damage is dealt only while the creature is still a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4dc0831a-0eb1-4f7d-9591-6e670e180116",
		Name:         "Lie in Wait",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
			TargetCreature("target creature"),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			card, haveCard := ctx.ClauseTarget(0)
			power := 0
			if haveCard && card.Kind == game.TargetCard {
				if err := (ReturnFromGraveyard{Target: card.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			if card.Kind == game.TargetCard {
				if c, ok := ctx.Game.LookupCardForEffect(card.ID); ok {
					power = c.CurrentPower()
				}
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard || power <= 0 {
				return nil
			}
			return DealDamage{Source: item.SourceCardID, Target: victim.ID, Amount: power}.Apply(ctx)
		},
	})
}
