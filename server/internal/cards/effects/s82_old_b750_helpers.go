package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b750_helpers.go — shared vocabulary for the S82 OLD batch 750 cards
// (Stronghold, #5627).

// GrantedAltCostDreamHalls is the key of Dream Halls' granted offer. Like the
// other granted keys it is an on-disk identity: never renamed, never reused.
const GrantedAltCostDreamHalls = "granted-dream-halls"

// DiscardAColorSharingCardForEverySpell is Dream Halls' static (#6214, CR
// 118.9): "Rather than pay the mana cost for a spell, its controller may
// discard a card that shares a color with that spell." Every player's spells,
// not only the grantor's (EveryPlayer), from hand, and the discard is one card
// that shares a colour with the spell being cast (DiscardSharesColorWithSpell),
// so a colourless spell can't be cast this way.
func DiscardAColorSharingCardForEverySpell() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:                         GrantedAltCostDreamHalls,
			Label:                       "Discard a card that shares a color with that spell",
			DiscardFromHand:             &game.TargetSpec{},
			DiscardSharesColorWithSpell: true,
		},
		Zones:       []game.ZoneKind{game.ZoneHand},
		EveryPlayer: true,
	}
}

// OldB750UntapAllCreaturesAndLands is "untap all creatures and lands"
// (Awakening): every tapped creature and land on the battlefield, whoever
// controls it, read now as the ability resolves (CR 611.2c).
func OldB750UntapAllCreaturesAndLands(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Tapped && (c.IsCreature() || c.IsLand()) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (UntapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// OldB750AnOpponentControlsTwoOrMoreCreatures is Walking Dream's "if an
// opponent controls two or more creatures": some seat other than the
// permanent's controller, still in the game, controls at least two creatures.
// Read through the layered view, so an animated land counts.
func OldB750AnOpponentControlsTwoOrMoreCreatures(g *game.Game, controller uuid.UUID) bool {
	counts := map[uuid.UUID]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller && c.IsCreature() {
			counts[c.Controller]++
		}
	}
	for p, n := range counts {
		if n < 2 {
			continue
		}
		if pl := g.PlayerByIDForEffect(p); pl != nil && !pl.Eliminated {
			return true
		}
	}
	return false
}
