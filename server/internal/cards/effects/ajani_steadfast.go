package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani Steadfast — Legendary Planeswalker — Ajani {3}{W}, loyalty 4:
//
//	"+1: Until end of turn, up to one target creature gets +1/+1 and
//	     gains first strike, vigilance, and lifelink.
//	 −2: Put a +1/+1 counter on each creature you control and a loyalty
//	     counter on each other planeswalker you control.
//	 −7: You get an emblem with "If a source would deal damage to you or
//	     a planeswalker you control, prevent all but 1 of that damage.""
//
// +1 is a pump and a keyword grant until end of turn on the one target,
// or nothing at all with no target picked (Min 0). A target gone in
// response gets nothing (CR 608.2b).
//
// −2 reads the permanents you control as it resolves. "Each other
// planeswalker" leaves out Ajani himself, by object. A creature that is
// also another planeswalker gets both counters.
//
// −7 IS NOT OFFERED. Its emblem's ability is a prevention effect, and an
// emblem can only carry statics, triggers and rule gates today
// (EmblemSpec has no Replacements slot), so the emblem would be a chip
// that prevents nothing. ADR 0032 and the Emblems section of
// docs/adding-cards.md say to leave such an ultimate out rather than
// charge seven loyalty for nothing; the seam is #4192.
//
// Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "90249526-1875-450a-882d-bf6f6bfd37a6",
		Name:         "Ajani Steadfast",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The −7 emblem isn't implemented yet, so that ability can't be activated."},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Until end of turn, up to one target creature gets +1/+1 and gains first strike, vigilance, and lifelink.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard || t.ID == uuid.Nil {
							continue
						}
						if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1,
							Label: "Ajani Steadfast — +1/+1 until end of turn"}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike", "vigilance", "lifelink"},
							Label: "Ajani Steadfast — first strike, vigilance and lifelink"}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:  "−2: Put a +1/+1 counter on each creature you control and a loyalty counter on each other planeswalker you control.",
				Cost:   LoyaltyCost(-2),
				Effect: ajaniSteadfastCounters,
			},
		},
	})
}

// ajaniSteadfastCounters puts a +1/+1 counter on each creature the
// activator controls and a loyalty counter on each planeswalker they
// control other than this Ajani, the set read once as it resolves.
func ajaniSteadfastCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item).asGroupMember()
	you := item.Controller
	var creatures, walkers []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you {
			continue
		}
		if c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
		if c.IsPlaneswalker() && c.InstanceID != item.SourceCardID {
			walkers = append(walkers, c.InstanceID)
		}
	}
	for _, id := range creatures {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range walkers {
		if err := (AddCounter{Target: id, Kind: game.CounterLoyalty, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
