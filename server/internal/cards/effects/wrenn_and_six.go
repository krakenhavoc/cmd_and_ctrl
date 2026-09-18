package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wrenn and Six — Legendary Planeswalker — Wrenn for {R}{G},
// starting loyalty 3:
//
//	"+1: Return up to one target land card from your graveyard to
//	     your hand.
//	 −1: Wrenn and Six deals 1 damage to any target.
//	 −7: You get an emblem with 'Instant and sorcery cards in your
//	     graveyard have retrace.'"
//
// The two abilities that make the card are wired in full, and the −1
// is the reason S27's #406 fix had to land first: before it, a
// Wrenn's −1 aimed at an opposing planeswalker incremented a number
// nothing read.
//
// "Up to one target" is a Min 0 clause, so the +1 is activatable with
// an empty graveyard and still gains the loyalty. That is the printed
// card and it is most of why Wrenn is played — the plus is a
// loyalty engine first and a land recursion second.
//
// The ultimate is STILL NOT REGISTERED, and since S40 (#623) emblems
// are no longer the reason. RETRACE is: the emblem's entire text is
// "Instant and sorcery cards in your graveyard have retrace", and
// retrace needs two things that do not exist — the alternative cost
// itself ("cast this from your graveyard by discarding a land"), and
// a granted cast permission a player can hold over a set of cards
// they did not print, which is seam #652.
//
// Registering it now would offer a −7 that kills Wrenn, puts a real
// emblem chip on the board reading "…have retrace", and then never
// offers a retrace cast. That is the lie ADR 0032 refuses, moved one
// step later, and a player who paid seven loyalty for it would get no
// signal in game at all. The day #652 lands this card gains an
// `Emblem` slot with one static in it and nothing else changes.
func init() {
	Register(Spec{
		OracleID:     "108ae90a-50fa-4cfd-b751-d630e41425fe",
		Name:         "Wrenn and Six",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The -7 ultimate isn't offered — its emblem grants retrace, which the engine can't cast yet (#652)."},
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Return up to one target land card from your graveyard to your hand.",
				Cost:    LoyaltyCost(1),
				Targets: upToOneLandInYourGraveyard(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.Targets() {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−1: Wrenn and Six deals 1 damage to any target.",
				Cost:    LoyaltyCost(-1),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if len(item.Targets) == 0 {
						return nil
					}
					return DealDamage{
						Source: item.SourceCardID,
						Target: item.Targets[0].ID,
						Amount: 1,
					}.Apply(ctx)
				},
			},
		},
	})
}

// upToOneLandInYourGraveyard is Wrenn's +1 clause. Min 0 is the "up
// to one": the picker's Done button is live with nothing selected, so
// the loyalty gain is available on an empty graveyard.
func upToOneLandInYourGraveyard() *game.TargetSpec {
	spec := TargetCardInGraveyard("up to one target land card in your graveyard", YouOwn(), Land())
	spec.Min = 0
	return spec
}
