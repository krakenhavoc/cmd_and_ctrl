package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b3_helpers.go — Secrets of Strixhaven Commander, batch 3 (#4834):
// the pieces two clauses or a card's own test share.

// socB3CreatureTokenYouControlDealtCombatDamageToAPlayer is "whenever a
// creature token you control deals combat damage to a player" (Curiosity
// Crafter): the damage source is a creature token its controller
// controls as the damage is dealt.
func socB3CreatureTokenYouControlDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.IsToken()
}

// socB3CommanderCastsFromCommandZone is how many times `player` has cast
// a commander from the command zone this game: Player.CommanderCasts,
// which only a cast from the command zone bumps, summed over every
// commander they cast.
func socB3CommanderCastsFromCommandZone(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return 0
	}
	n := 0
	for _, casts := range p.CommanderCasts {
		n += casts
	}
	return n
}

// socB3SubtypeCantAttackYouOrYourPlaneswalkers is "<creatures of a
// subtype> can't attack you or planeswalkers you control" (Combat
// Calligrapher's "Inklings can't …", CR 508.1c): CreaturesMatchingCantAttackYou
// without its NotPlayerOnly narrowing, so the planeswalkers the printing
// permanent's controller controls are closed to those attackers too. A
// battle that player protects is not covered: the text does not say it.
//
// The subtype is a word of the printing permanent's text (CR 612.2), and
// whether an attacker has it is judged at declaration on its effective
// types (AttackTargetRestriction.OnlyAttackersMatching), so a changeling
// is an Inkling and a creature that stops being one may attack.
func socB3SubtypeCantAttackYouOrYourPlaneswalkers(label, subtype string) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:                source.InstanceID,
				SourceName:            source.Name,
				NotPlayer:             source.Controller,
				OnlyAttackersMatching: []game.PermanentQuery{QuerySubtype(source.TextSubtype(subtype))},
			})
		},
	}
}

// socB3APlayerAttackedYourOpponent is "whenever a player attacks one of
// your opponents" read by CR 508.3e: a creature was declared attacking
// that PLAYER (an attack on their planeswalker or battle is not one), and
// the player is not this permanent's controller. The attacker may be
// anyone, the controller included. Batched per attacked player
// (OncePerBatchPerPlayer), so one declaration at one opponent is one
// trigger however many creatures it names.
func socB3APlayerAttackedYourOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventAttack || ev.Target == uuid.Nil || ev.Target == source.Controller {
		return false
	}
	p := g.PlayerByIDForEffect(ev.Target)
	return p != nil && !p.Eliminated
}

// KeywordDemonstrate is the printed keyword Demonstrate's trigger row
// carries.
const KeywordDemonstrate = "demonstrate"

// Demonstrate is the keyword's cast trigger (CR 702.144a): "When you cast
// this spell, you may copy it and you may choose new targets for the
// copy. If you copy the spell, choose an opponent. That player copies the
// spell and may choose new targets for that copy."
//
// The "you may" is asked as the trigger resolves. The copy is made from
// last-known information when the spell has already left the stack: the
// trigger names the spell without targeting it, as storm and replicate
// do. Only once the controller's copy exists is an opponent chosen, and
// that opponent creates and controls the second copy, so it resolves
// first and its "you" is that opponent. A copy is not cast, so neither
// copy demonstrates again.
func Demonstrate() game.TriggeredAbility {
	t := On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.CardID == source.InstanceID
	}, "Demonstrate — you may copy this spell; if you do, an opponent of your choice also copies it", demonstrateEffect)
	t.Keyword = KeywordDemonstrate
	t.FromStack = true
	return Optional(t, "Demonstrate — copy this spell? If you do, choose an opponent to also copy it.")
}

func demonstrateEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spell := item.SourceCardID
	if err := (CopySpell{StackID: spell, ChooseNewTargets: true, FromLastKnown: true}).Apply(ctx); err != nil {
		return err
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Demonstrate — choose an opponent to also copy the spell",
		ThenKey:  demonstrateOpponentCopies,
		Carry:    []uuid.UUID{spell},
	}.Apply(ctx)
}

// demonstrateOpponentCopies is "that player copies the spell and may
// choose new targets for that copy", keyed: the spell rides the prompt.
var demonstrateOpponentCopies = ChoosePlayerThen("option-pick/demonstrate-opponent-copies", func(ctx *Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if chosen == uuid.Nil || len(carry) == 0 {
		return nil
	}
	return CopySpell{StackID: carry[0], Controller: chosen, ChooseNewTargets: true, FromLastKnown: true}.Apply(ctx)
})
