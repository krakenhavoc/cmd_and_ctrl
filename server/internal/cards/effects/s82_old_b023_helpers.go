package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b023_helpers.go — Secret Lair Drop, removal (S82 older sets, batch
// 023, #5462): the shared shapes of Fracturing Gust, Force of Despair, Kogla,
// the Titan Ape and Leveler.

// S82OldB023GainTwoLifeForEach is Fracturing Gust's "You gain 2 life for each
// permanent destroyed this way": a registered tail told the permanents the
// sweep actually destroyed (one that was indestructible or regenerated is not
// counted).
var S82OldB023GainTwoLifeForEach = MassTail("mass/s82-old-b023-gain-two-life-for-each",
	func(ctx *Context, _ []game.Card, destroyed int) error {
		if destroyed <= 0 {
			return nil
		}
		return GainLife{Amount: 2 * destroyed}.Apply(ctx)
	})

// S82OldB023DestroyArtifactsAndEnchantments is "Destroy all artifacts and
// enchantments. You gain 2 life for each permanent destroyed this way."
func S82OldB023DestroyArtifactsAndEnchantments(_ *game.StackItem, ctx *Context) error {
	return DestroyAllMatching{
		Match:    Or(Artifact(), Enchantment()),
		ThenTail: ctx.Tail(S82OldB023GainTwoLifeForEach),
	}.Apply(ctx)
}

// S82OldB023CreatureEnteredThisTurn is the clause "creatures that entered this
// turn": a creature whose own entry (this object's, CR 400.7) happened during
// the current turn, not summoning sickness.
func S82OldB023CreatureEnteredThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && B06EnteredThisTurn(g, c.InstanceID)
	}
}

// TargetArtifactOrEnchantmentDefendingPlayerControls is the TargetsFrom clause
// for "target artifact or enchantment defending player controls" (Frenzied
// Trapbreaker, Kogla). It reads only the source's own attack, so a restore
// rebuilds the same clause; with no defending player to read it has no target
// and the trigger is removed from the stack (CR 603.3d).
func TargetArtifactOrEnchantmentDefendingPlayerControls(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "permanent", Label: "target artifact or enchantment defending player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return (c.IsArtifact() || c.IsEnchantment()) && c.Controller == defender
		},
		Min: 1, Max: 1,
	}
}

// S82OldB023ReturnHumanThenIndestructible is Kogla's third ability: "Return
// target Human you control to its owner's hand. Kogla gains indestructible
// until end of turn." The target must be legal as the ability resolves
// (CR 608.2b). The second sentence is not gated on the return (a commander
// that went to the command zone, or a move a replacement redirected, still
// gives Kogla the indestructible), so it is the bounce's tail, run once the
// move has settled, and only while Kogla is still the same permanent.
func S82OldB023ReturnHumanThenIndestructible(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BounceToHand{Target: id, ThenTail: ctx.ItemTail(s82OldB023KoglaIndestructible)}.Apply(ctx)
}

// s82OldB023KoglaIndestructible is the tail of Kogla's third ability. An
// on-disk key: never renamed, never reused.
var s82OldB023KoglaIndestructible = CardTail("then/s82-old-b023-kogla-indestructible",
	func(ctx *Context, _ game.TailOutcome) error {
		src := ctx.Source()
		if !OnBattlefield(ctx.Game, src) || ctx.IsNewSourceObject(src) {
			return nil
		}
		return GrantKeywordUntilEOT{
			Target:   src,
			Keywords: []string{"indestructible"},
			Label:    "Kogla, the Titan Ape — indestructible",
		}.Apply(ctx)
	})

// S82OldB023ExileYourLibrary is Leveler's enters trigger: "exile all cards
// from your library". Nothing is looked at or kept, so it is one exile of
// every card (an empty library exiles nothing, and is not a loss: nobody
// drew).
func S82OldB023ExileYourLibrary(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Library == nil || p.Library.Size() == 0 {
		return nil
	}
	return g.ExileTopKeyedForEffect(item.Controller, item.Controller, p.Library.Size(), game.KeyedTail{})
}
