package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grapple with the Past — Instant {1}{G} (EDHREC rank 2156):
//
//	"Mill three cards, then you may return a creature or land card
//	 from your graveyard to your hand. (To mill three cards, put the
//	 top three cards of your library into your graveyard.)"
//
// The self-mill regrowth. The mill is MillCards; the return is a
// creature or land card the caster owns, back to hand.
//
// Sandbox simplification, declared (the Rise of the Witch-king
// posture): "then you may return" is a resolution-time choice made
// AFTER the mill, and the engine has no pick-from-graveyard prompt
// with a continuation for a spell. So the card to return is picked
// when the spell is cast, as an optional target ("up to one"), from
// the graveyard as it stands BEFORE the mill, and it comes back after
// the three cards are milled. Three consequences, all weaker than
// printed: the three milled cards cannot be the one returned;
// opponents see the pick before the spell resolves; and if the
// picked card left the graveyard in response the spell is countered
// by game rules (CR 608.2b) and the mill does not happen either. With
// no target chosen the spell simply mills three.
func init() {
	Register(Spec{
		OracleID:     "d14c633d-977a-4f78-85c2-88582b3c3670",
		Name:         "Grapple with the Past",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The card to return is picked when you cast the spell, before the three cards are milled — so a card milled by the spell itself can't be the one that comes back, and opponents can respond to the choice.",
		},
		Targets: TargetCardInGraveyard("up to one creature or land card in your graveyard to return",
			YouOwn(), Or(Creature(), Land())).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (MillCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
				return err
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
			}
			return nil
		},
	})
}
