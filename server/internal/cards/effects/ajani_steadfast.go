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
// −7 makes an emblem whose one ability is a prevention effect
// (EmblemSpec.Replacements, #4192): the replacement gather walks every
// seat's emblems with the emblem as the source and its owner as "you"
// (CR 114.2). Damage to the owner, or to a planeswalker the owner
// controls, from any source is cut to 1 when it is above 1 (CR 615.1a);
// damage of 1 or 0 is untouched. The emblem is ordered against other
// prevention and replacement effects by the affected player (CR 616.1),
// and "that damage can't be prevented" leaves it prevented-nothing like
// any other prevention (CR 615.12, Prevention: true).
//
// Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "90249526-1875-450a-882d-bf6f6bfd37a6",
		Name:         "Ajani Steadfast",
		Completeness: CompletenessFull,
		Emblem: &EmblemSpec{
			Label:        "Ajani Steadfast emblem",
			Text:         "If a source would deal damage to you or a planeswalker you control, prevent all but 1 of that damage.",
			Replacements: []game.ReplacementEffect{ajaniSteadfastEmblemPrevention()},
		},
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
			{
				Label: "−7: You get an emblem with \"If a source would deal damage to you or a planeswalker you control, prevent all but 1 of that damage.\"",
				Cost:  LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// ajaniSteadfastEmblemPrevention is the emblem's "prevent all but 1 of
// that damage": for damage to the emblem's owner or to a planeswalker
// they control, the event's amount drops to 1. `src` is the emblem; its
// controller is its owner and never changes (CR 114.2).
func ajaniSteadfastEmblemPrevention() game.ReplacementEffect {
	return preventAllButOneOfTheDamage(
		"Ajani Steadfast emblem — prevent all but 1 of that damage",
		"Prevent all but 1 of the damage dealt to you or a planeswalker you control.",
		func(target uuid.UUID, g *game.Game, src *game.Card) bool {
			if target == src.Controller {
				return true
			}
			c, ok := g.LookupCardForEffect(target)
			return ok && g.Battlefield.Contains(target) && c.IsPlaneswalker() && c.Controller == src.Controller
		})
}

// preventAllButOneOfTheDamage is the CR 615 "prevent all but 1 of that
// damage" replacement, for any recipient `to` accepts (Ajani Steadfast's
// emblem, Hyperion, Supreme Hero): damage to a recipient it names drops to
// 1, and damage of 1 or less is left alone.
func preventAllButOneOfTheDamage(label, rowLabel string, to func(target uuid.UUID, g *game.Game, src *game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a — "prevent"; CR 615.12 reads it
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 1 {
				return false
			}
			return to(ev.DamageTarget, g, src)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount = 1
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
		RowLabel:   rowLabel,
	}
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
