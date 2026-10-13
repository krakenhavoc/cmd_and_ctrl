package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goldbug, Humanity's Ally // Goldbug, Scrappy Scout (#1856, ADR 0152) —
// the proof card for More Than Meets the Eye, convert and living metal.
//
// Front face, Legendary Artifact Creature — Robot {1}{W}{U}, 3/3:
//
//	"More Than Meets the Eye {W}{U} (You may cast this card converted
//	 for {W}{U}.)
//	 Prevent all combat damage that would be dealt to attacking Humans
//	 you control.
//	 Whenever you cast your second spell each turn, convert Goldbug."
//
// Back face, Legendary Artifact — Vehicle, 1/3:
//
//	"Living metal (During your turn, this Vehicle is also a creature.)
//	 Human spells you control can't be countered.
//	 Whenever Goldbug and at least one Human attack, draw a card and
//	 convert Goldbug."
//
// More Than Meets the Eye is the front face's offer (CastsFace 1, no
// zone binding): a converted Goldbug is cast back face up from the hand
// or the command zone, with the commander tax, and enters as the
// Vehicle. Convert is ConvertThis, which ignores a second "convert
// Goldbug" once Goldbug has already turned over since that ability went
// on the stack (CR 701.28e). Living metal is the back face's layer-4
// static. The prevention is a CR 615 replacement read off the board, the
// shape Losheel, Clockwork Scholar uses. "Human spells you control can't
// be countered" is the counter gate ADR 0106 PR 2 built.
//
// "Goldbug and at least one Human attack": Goldbug's own attack event,
// with at least one attacking Human you control. Every attacker of the
// declaration is stamped before any EventAttack is emitted, so the
// Humans attacking beside it are on the board when the trigger asks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         goldbugOracleID,
		Name:             "Goldbug, Humanity's Ally",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{W}{U}")},
		Replacements: []game.ReplacementEffect{
			preventCombatDamageToYourAttackers(
				"Goldbug, Humanity's Ally: prevent all combat damage that would be dealt to attacking Humans you control",
				func(c game.Card) bool { return c.HasSubtype("Human") }),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Goldbug, Humanity's Ally — convert Goldbug", Do(ConvertThis{})),
		},
	})
	Register(Spec{
		OracleID:     goldbugOracleID + "#1",
		Name:         "Goldbug, Scrappy Scout",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{LivingMetal()},
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Human spells you control can't be countered.", Subtype("Human")),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, goldbugAndAHumanAttack,
				"Goldbug, Scrappy Scout — draw a card and convert Goldbug",
				Do(DrawCards{N: 1}, ConvertThis{})),
		},
	})
}

const goldbugOracleID = "d1960749-c11e-4ac4-97cb-f37d7875d479"

// goldbugAndAHumanAttack is "Whenever Goldbug and at least one Human
// attack": this permanent's own attack, while a Human you control is
// attacking too.
func goldbugAndAHumanAttack(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	if !ThisAttacked(ev, source, ch, g) || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != source.InstanceID && c.Controller == source.Controller &&
			c.AttackingTarget != uuid.Nil && c.IsCreature() && c.HasSubtype("Human") {
			return true
		}
	}
	return false
}

// preventCombatDamageToYourAttackers is "Prevent all combat damage that
// would be dealt to attacking <kind> you control" (Goldbug, Humanity's
// Ally): a CR 615 prevention over combat damage events, read off the
// board while its source is on the battlefield. `kind` judges the
// attacking creature.
func preventCombatDamageToYourAttackers(label string, kind func(game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || !ev.IsCombatDamage {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && c.Controller == src.Controller && c.AttackingTarget != uuid.Nil &&
				c.IsCreature() && kind(c)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
