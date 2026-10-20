package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blitzball Stadium — Artifact {X}{U}:
//
//	"When this artifact enters, support X. (Put a +1/+1 counter on each
//	 of up to X target creatures.)
//	 Go for the Goal! — {3}, {T}: Until end of turn, target creature
//	 gains "Whenever this creature deals combat damage to a player, draw
//	 a card for each kind of counter on it" and it can't be blocked this
//	 turn."
//
// The support's count is the X paid to cast the Stadium (Card.CastX,
// CR 107.3m), read as the trigger is built, so its clause is "up to X
// target creatures" (TargetsFrom). The artifact is not a creature, so
// the reminder text's "target creatures" are any (CR 701.41a). With X of
// zero there is nothing to target and nothing happens.
//
// Go for the Goal! grants the trigger for the turn (an ADR 0093 duration
// grant) and makes the creature unblockable for the turn (Distortion
// Strike's restriction). The granted trigger counts the KINDS of counter
// on the creature as it resolves, or as it last existed if it has left:
// a +1/+1 and a shield counter draw two, three +1/+1 counters draw one.
//
// No simplification.
const blitzballStadiumGrant = "blitzball-stadium/draw-per-counter-kind"

func init() {
	support := WhenThisEnters("Blitzball Stadium — support X", SupportEffect)
	support.TargetsFrom = UpToCastXCreatureTargets
	Register(Spec{
		OracleID:     "a4b66fb3-16ea-4bf1-bce3-6df872e8883b",
		Name:         "Blitzball Stadium",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: blitzballStadiumGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Go for the Goal! — draw a card for each kind of counter on this creature",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						self, ok := ctx.SourcePermanent()
						if !ok {
							return nil
						}
						kinds := 0
						for _, n := range self.Counters {
							if n > 0 {
								kinds++
							}
						}
						if kinds == 0 {
							return nil
						}
						return DrawCards{N: kinds}.Apply(ctx)
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card for each kind of counter on it.",
		}},
		Triggered: []game.TriggeredAbility{support},
		Activated: []ActivatedAbility{{
			Label:   "Go for the Goal! — {3}, {T}: Until end of turn, target creature gains \"Whenever this creature deals combat damage to a player, draw a card for each kind of counter on it\" and it can't be blocked this turn.",
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (GrantAbilitiesFor{
					Target: id,
					Keys:   []string{blitzballStadiumGrant},
					Label:  "Go for the Goal! — draws a card for each kind of counter on it when it deals combat damage to a player",
				}).Apply(ctx); err != nil {
					return err
				}
				return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Go for the Goal! — can't be blocked"}.Apply(ctx)
			},
		}},
	})
}
