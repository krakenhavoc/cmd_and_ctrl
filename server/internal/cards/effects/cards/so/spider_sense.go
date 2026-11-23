package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spider-Sense — Instant {1}{U}:
//
//	"Web-slinging {U} (You may cast this spell for {U} if you also
//	 return a tapped creature you control to its owner's hand.)
//	 Counter target instant spell, sorcery spell, or triggered
//	 ability."
//
// Web-slinging is Daze's return-to-hand alternative cost (ReturnInstead
// / AlternativeCost.ReturnToHand) with a real mana component alongside
// it rather than in place of it — the {U} still has to be paid, the
// bounce is the discount. Built as a literal game.AlternativeCost
// rather than through ReturnInstead, because that constructor hardcodes
// ManaCost to "" for Daze's all-or-nothing pitch and this keyword pays
// both halves.
//
// The counter target is TargetSpellOrAbility narrowed to an instant or
// sorcery SPELL on one side (ASpellItem) and a TRIGGERED ability item
// on the other (StackItemKind == StackItemTriggered) — Disallow's
// three-way clause (#1211) with the activated-ability third dropped,
// since this card never offers one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4485bf9-e7ff-48b5-a983-02900939ee9d",
		Name:         "Spider-Sense",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetSpellOrAbility("target instant spell, sorcery spell, or triggered ability",
			effects.AnyStackItem(effects.ASpellItem(effects.Or(effects.Instant(), effects.Sorcery())), spiderSenseTriggeredAbilityItem)),
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {U}",
			ManaCost:     "{U}",
			ReturnToHand: effects.PermanentYouControl("a tapped creature you control", effects.Creature(), effects.B751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
		OnResolve: effects.CounterTheTargetSpell,
	})
}

// spiderSenseTriggeredAbilityItem passes for a TRIGGERED ability item
// on the stack — not a spell, not an activated ability.
func spiderSenseTriggeredAbilityItem(_ *game.Game, _ uuid.UUID, item *game.StackItem) bool {
	return item.Kind == game.StackItemTriggered
}
