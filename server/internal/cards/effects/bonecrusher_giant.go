package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bonecrusher Giant // Stomp — the catalog's first TARGETED adventure
// card (CR 715), oracle d6d72f5f…:
//
//	Bonecrusher Giant — Creature — Giant {2}{R}, 4/3
//	  "Whenever this creature becomes the target of a spell, this
//	   creature deals 2 damage to that spell's controller."
//	Stomp — Instant — Adventure {1}{R}
//	  "Damage can't be prevented this turn. Stomp deals 2 damage to
//	   any target. (Then exile this card. You may cast the creature
//	   later from exile.)"
//
// # Why this card, and why it could not ship with #719
//
// #719 shipped the adventure LIFECYCLE and Foulmire Knight // Profane
// Insight proved it: the cheapest printed Adventure half that needs no
// announce-time choice at all. Every other candidate the issue named —
// Stomp, Petty Theft, Swift End — is a TARGETED instant, and a
// targeted Adventure half was blocked on something that was not in the
// engine at all. The view published `target_mode` and `legal_targets`
// for the face that was UP, so a human client asked to cast face 1 had
// no picker to open; the bot enumerator was already face-correct, and
// so was CastSpell. #992 put the announce data on the wire per
// castable face and taught `cardAsFace` to consume it, and this card
// is what proves it end to end.
//
// # Two keys, one card
//
// One oracle_id per CARD, so game.CatalogKey makes it composite: the
// creature keeps the bare oracle ID and Stomp takes "<oracle_id>#1".
// Both faces register, as Foulmire Knight's do — the creature half is
// a real castable object whose coverage the catalog page should state
// rather than leave blank.
//
// # ONE DECLARED SIMPLIFICATION, in Stomp's first sentence
//
// "DAMAGE CAN'T BE PREVENTED THIS TURN" IS NOT MODELLED. Unpreventable
// damage has no shape in the engine: the CR 615 prevention shield is a
// replacement effect and nothing suppresses one (builtin_replacements.go
// says so at CR 615.12), and the clause is a turn-scoped continuous
// effect rather than a rider on this spell's own damage — it applies to
// every source for the rest of the turn, which is why the printed card
// is a fog-breaker rather than a Shock. Banefire ships the identical
// caveat for the identical reason.
//
// Weaker than printed, which is the direction a simplification must
// take (#259): a fog or a protective shield stops a Stomp here. The
// damage half is live and it is the whole reason the card is being
// cast.
//
// The creature half ships FULL. "Becomes the target of A SPELL" is
// narrower than the "spell or ability" wording its family usually
// prints — EventBecomesTarget is emitted for both and carries no
// discriminator, so the spell half is read off the stack
// (SelfTargetedByASpell, the same predicate Departed Deckhand and
// Gargos use). Firing on abilities too would be a strictly better
// Bonecrusher Giant, which is the one direction this catalog does not
// ship.
//
// The trigger fires at ANNOUNCE (CR 115.7), so it goes on the stack
// ABOVE the spell that targeted and resolves first. That is exactly
// why the card is played: the two damage is dealt before the removal
// spell resolves, and a Giant that survives is still there when it
// does.

// bonecrusherGiantOracleID is shared by both faces of the card.
const bonecrusherGiantOracleID = "d6d72f5f-8f5d-4180-b514-f22ff5482902"

func init() {
	// Face 0 — the creature.
	Register(Spec{
		OracleID:     bonecrusherGiantOracleID,
		Name:         "Bonecrusher Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventBecomesTarget},
			AppliesTo: SelfTargetedByASpell,
			// "THAT SPELL'S CONTROLLER" is Event.Actor, captured in
			// Build rather than looked up in the effect. The trigger
			// is harvested at announce (CR 115.7) and resolves later,
			// by which time the spell may have been countered and its
			// stack item gone — and the ability names the player it
			// named when it triggered, not whoever is left. Actor is
			// set by emitBecameTargetLocked at every one of the six
			// sites that finish choosing targets, so this reads the
			// same answer however the spell was announced.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				caster := ev.Actor
				return game.NewTriggeredItem(source,
					"Bonecrusher Giant — 2 damage to that spell's controller",
					func(g *game.Game, item *game.StackItem) error {
						if caster == uuid.Nil {
							return nil
						}
						return DealDamage{
							Source: item.SourceCardID,
							Target: caster,
							Amount: 2,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})

	// Face 1 — the Adventure. An instant while it is on the stack
	// (CR 715.3a); what happens after OnResolve returns is the
	// engine's CR 715.3d branch (game/adventure.go), and this file
	// says nothing about it.
	Register(Spec{
		OracleID:     bonecrusherGiantOracleID + "#1",
		Name:         "Stomp",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Damage prevention still stops it — the \"damage can't be prevented this turn\" clause isn't implemented.",
		},
		Targets: TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
