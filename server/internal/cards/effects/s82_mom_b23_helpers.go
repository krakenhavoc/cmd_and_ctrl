package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b23_helpers.go — shared bodies for the March of the Machine
// Sieges of S82 batch 23.

// lookAtOpponentsHandExileTaxed is the enters-trigger body shared by Zero,
// Incorporeal Sensate and Invasion of Gobakhan: "look at target opponent's
// hand. You may exile a nonland card from it. For as long as that card
// remains exiled, its owner may play it. A spell cast this way costs {2}
// more to cast." The target is the trigger's; the pick continues in
// zeroIncorporealSensateExile (an on-disk key, kept).
func lookAtOpponentsHandExileTaxed(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ChooseFromRevealedHand{
		Player:   TargetedPlayer(ctx),
		Filter:   Nonland(),
		Label:    "nonland card",
		Look:     true,
		Optional: true,
		Keep:     true,
		Then:     zeroIncorporealSensateExile,
	}.Apply(ctx)
}

// putPlusOneCounterOnEach puts one +1/+1 counter on each permanent that
// matches, the set taken before the first counter lands.
func PutPlusOneCounterOnEach(ctx *Context, match CardPredicate) error {
	var ids = make([]game.Card, 0)
	ids = append(ids, MatchingBattlefield(ctx, match)...)
	for _, c := range ids {
		if err := (AddCounter{Target: c.InstanceID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ExactlyTwoColors passes a card (or spell) that is exactly two colors,
// read through the layers (CR 105.2): Invasion of Ravnica's "isn't exactly
// two colors" and Guildpact Paragon's "a spell that's exactly two colors".
func ExactlyTwoColors() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return len(c.EffectiveColors()) == 2 }
}

// WhenThisEntersWithCastX is "When this <permanent> enters, <effect>" for a
// permanent cast with {X}: the effect reads X from item.Params.Amount, the
// value announced for the spell (CR 107.3m) read off the permanent as the
// trigger is built, so it is exact even if the source has left by the time
// the trigger resolves. A copy or a cheat onto the battlefield has X = 0.
func WhenThisEntersWithCastX(label string, effect Effect) game.TriggeredAbility {
	t := WhenThisEnters(label, effect)
	t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Amount = source.CastX()
		return item
	}
	return t
}

// YouDiscardedALandCard is the When for "whenever you discard a land card"
// (Doctor Doom, King of Latveria; Pyre of the World Tree): a discard by the
// source's controller of a card that is a land card, read where it landed.
func YouDiscardedALandCard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !DiscardedByYou(ev, source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand()
}
