package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// attachments_batch3.go is the third batch of S24 catalog work on the
// attachment relation (ADR 0036), and like attachments.go it holds
// only what more than one card in the batch needs.
//
// Nothing here is a new engine primitive. Every helper is a
// composition of the five things #374 and #379 built — the relation,
// the equip ability, the attachment-scoped static, the
// attachment-scoped trigger condition and the enchant clause — and
// the reason they are in a shared file rather than copy-pasted is
// that the next Equipment to want "+1/+1 for each X" should get the
// same layer, the same sub-layer and the same re-read-every-recompute
// behaviour as the last one.

// --- attachment-scoped trigger conditions ------------------------

// attachedCreatureAttacked is "Whenever equipped creature attacks" /
// "Whenever enchanted creature attacks" — the third attachment
// trigger condition, after Skullclamp's death and the Swords' combat
// damage.
//
// EventAttack carries the attacking creature in CardID (Target is the
// player being attacked), and it fires once per attacking creature,
// only on that creature's FIRST declaration. So a single attachment
// on a single creature gets exactly one trigger per combat with no
// batching guard — unlike Curse of Opulence, which watches the
// DEFENDER and therefore sees one event per attacker.
func attachedCreatureAttacked(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventAttack && source.IsAttachedTo(ev.CardID)
}

// attachedCreatureDealtDamageToOpponent is Curiosity's condition:
// "Whenever enchanted creature deals damage to an OPPONENT".
//
// Three differences from attachedCreatureDealtCombatDamageToPlayer,
// all of them printed on the card:
//
//   - Any damage, not only combat damage. A Curiosity'd creature that
//     pings with an activated ability draws too.
//   - An OPPONENT, not any player. Damage dealt to the Aura's own
//     controller — a goaded creature forced to swing at you, a
//     creature stolen back — does not draw.
//   - "Opponent" is measured against the AURA's controller, which is
//     who the "you" in "you may draw a card" refers to (CR 109.5).
//     A stolen Curiosity draws for whoever holds the Aura now.
func attachedCreatureDealtDamageToOpponent(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 {
		return false
	}
	if !source.IsAttachedTo(ev.Source) {
		return false
	}
	p := g.PlayerByIDForEffect(ev.Target)
	return p != nil && p.ID != source.Controller
}

// --- attachment-scoped statics -----------------------------------

// PumpAttachedPer is "Equipped creature gets +P/+T for each <thing>"
// — Blackblade Reforged's lands, Cranial Plating's artifacts, All
// That Glitters's artifacts and enchantments.
//
// `count` is evaluated on EVERY layer recompute, not captured at
// attach time, which is what makes the bonus track a board that
// changes underneath it: play a land and the Blackblade grows in the
// same beat, with no invalidation bookkeeping. That is the same
// property AttachedToSource has, for the same reason.
//
// It runs inside a recompute pass where the caller may hold only the
// read lock, so `count` must walk g.Battlefield.Cards directly rather
// than through BattlefieldCardsForEffect — see sharedCreatureTypeCount
// for the full note.
func PumpAttachedPer(power, toughness int, count func(g *game.Game, source *game.Card) int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := count(g, source)
			c.Power += power * n
			c.Toughness += toughness * n
		},
	}
}

// GrantToAttachedWhile is "As long as equipped creature is <X>, it
// has <keyword>" — Champion's Helm's hexproof-while-legendary.
//
// The condition is re-read per recompute like everything else here,
// so a creature that BECOMES legendary under the Helm gains hexproof
// without the Helm moving, and one that stops being legendary loses
// it. The `host` passed to the condition is the attached permanent as
// the earlier layers have left it this pass.
//
// Same keyword rule as GrantToAttached: canonical lowercase tokens
// the engine honours, or the grant is a badge with nothing behind it.
func GrantToAttachedWhile(cond func(host *game.Card, g *game.Game, source *game.Card) bool, keywords ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return AttachedToSource(target, g, source) && cond(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range keywords {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
			}
		},
	}
}

// --- counting helpers for PumpAttachedPer ------------------------

// permanentsYouControl counts the battlefield permanents controlled
// by `source`'s controller that satisfy `match`.
//
// "You" is the ATTACHMENT's controller, not the host creature's. That
// is CR 109.5 again and it is observable: a Blackblade Reforged whose
// creature has been stolen still counts ITS controller's lands.
func permanentsYouControl(g *game.Game, source *game.Card, match func(c *game.Card) bool) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		other := &g.Battlefield.Cards[i]
		if other.Controller != source.Controller {
			continue
		}
		if match(other) {
			n++
		}
	}
	return n
}

// artifactOrEnchantmentsControlledBy counts "artifact and/or
// enchantment you control" — All That Glitters. A permanent that is
// both counts ONCE, which is what the "and/or" means and what the
// `||` gives.
//
// Reads post-layer types through Effective(), the way
// artifactsControlledBy does: an animated or type-changed permanent
// counts, because by the time layer 7 runs layer 4 has already
// written.
func artifactOrEnchantmentsControlledBy(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool {
		for _, t := range c.Effective().Types {
			if t == "Artifact" || t == "Enchantment" {
				return true
			}
		}
		return false
	})
}

// landsControlledBy counts "land you control" — Blackblade Reforged.
func landsControlledBy(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsLand() })
}

// forestsControlledBy counts "Forest you control" — Blanchwood Armor.
// The card says the land TYPE, not the name, so a Stomping Ground and
// a Dryad Arbor both count and a Forest that has been turned into an
// Island does not. HasSubtype is the accessor that answers that
// correctly (and post-layer, so a Urborg-style type grant lands).
func forestsControlledBy(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype("Forest") })
}

// otherEnchantmentsOnTheBattlefield counts "each OTHER enchantment on
// the battlefield" — Ancestral Mask, which is symmetrical and counts
// every player's enchantments, itself excluded.
func otherEnchantmentsOnTheBattlefield(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		other := &g.Battlefield.Cards[i]
		if other.InstanceID == source.InstanceID {
			continue
		}
		for _, t := range other.Effective().Types {
			if t == "Enchantment" {
				n++
				break
			}
		}
	}
	return n
}

// --- ward on the host --------------------------------------------

// WardAttached is "Equipped creature has ward {N}" — Lavaspur Boots.
//
// The narrow case of WardGranted (effects/ward.go), whose doc carries
// the design: ward is a TRIGGERED ability (CR 702.21a) with a cost,
// so a granted ward cannot ride GrantToAttached or any other layer-6
// keyword grant — the grant is the Equipment carrying the ward
// trigger and watching for its host becoming a target.
//
// The predicate is the only thing this shape adds: the warded
// permanent is whichever one the source is attached to. Everything
// else — the battlefield gate, "an opponent controls" measured
// against the HOST's controller rather than the Equipment's, and the
// payer being the spell's controller — is WardGranted's and is
// documented there.
func WardAttached(cost WardCost, label string) game.TriggeredAbility {
	return WardGranted(cost, label, func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source.IsAttachedTo(target.InstanceID)
	})
}
