package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silver Surfer, Galactus's Herald — {5} Legendary Creature — Alien Hero 4/5
// (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 When Silver Surfer enters, you may search your library for a card named
//	 Galactus, Devourer of Worlds, reveal it, put it into your hand, then
//	 shuffle.
//	 Whenever Silver Surfer deals combat damage to a player, until the end of
//	 your next turn, target creature attacks that player each combat if able."
//
// The search is Cloud, Midgar Mercenary's, optional and narrowed to the name
// (a search with a stated quality may find nothing, CR 701.23b). The attack
// requirement is Encore's ModAttackPlayerIfAble on the target, naming the
// damaged player as the event had it, lasting through the end of the
// controller's next turn (CR 611.2): a target that is gone, or a damaged player
// who has left, gets nothing, and a creature that cannot attack that player
// is owed nothing (CR 508.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be02005c-0e37-4dba-ac8e-3c69d06bbec2",
		Name:            "Silver Surfer, Galactus's Herald",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Silver Surfer, Galactus's Herald — you may search for a card named Galactus, Devourer of Worlds", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return game.CardNameMatches(c, "Galactus, Devourer of Worlds")
					},
					Dest:     game.ZoneHand,
					Limit:    1,
					Reveal:   true,
					Shuffle:  true,
					Optional: true,
					Reason:   "Silver Surfer, Galactus's Herald — a card named Galactus, Devourer of Worlds, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
			Targeting(
				WheneverThisDealsCombatDamageToAPlayer(
					"Silver Surfer, Galactus's Herald — until the end of your next turn, target creature attacks that player each combat if able",
					silverSurferHeraldForce),
				TargetCreature("target creature")),
		},
	})
}

// silverSurferHeraldForce writes the requirement on the target.
func silverSurferHeraldForce(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := item.Trigger.Event.Target
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.AttackPlayerIfAbleMod(victim)},
			Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
			Label:    "Silver Surfer, Galactus's Herald — attacks that player each combat if able",
		}.Apply(ctx)
	}
	return nil
}
