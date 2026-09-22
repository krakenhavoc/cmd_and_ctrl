package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_restriction.go — #1210, ADR 0073's amendment of
// 2026-09-22: constructors for Spec.ActivationRestrictions, the
// board-wide "activated abilities of … can't be activated" statics
// (CR 602.5a, CR 101.2).
//
// The twin of cast_restriction.go and written the same way: one
// constructor per printed SHAPE rather than a generic builder,
// because the shape carries the clause the client is shown along with
// the predicate, and a card file that wrote the predicate and forgot
// the label ships a refusal nobody can explain. Register refuses both
// halves apart.
//
// THE MANA EXEMPTION IS AN ARGUMENT, NOT A DEFAULT. Half these cards
// print "unless they're mana abilities" (Pithing Needle, Null Rod)
// and half do not (Cursed Totem, Linvala, Collector Ouphe), and the
// difference is the whole of what separates a Cursed Totem from a
// blank against an elf-ball deck. So every constructor here takes
// `exemptMana` explicitly and none of them defaults it: a card file
// has to say which card it is holding.
//
// The predicate runs under g.mu inside the activation path.
// Read-only: *ForEffect accessors and plain field reads, never a
// locking mutator.

// SourcesCantActivate is the plain board-wide shape: "Activated
// abilities of <things> can't be activated [unless they're mana
// abilities]", binding EVERY player including the source's own
// controller.
//
// Cursed Totem ("of creatures", no exemption), Collector Ouphe ("of
// artifacts", no exemption), Null Rod and Stony Silence ("of
// artifacts", exempt). `match` picks out which sources the clause
// covers; nil covers every source, which is the Karn, the Great
// Creator shape.
//
// The match runs against the OBJECT whose ability is being activated,
// as it stands right now — so it reads the layered type line
// (Card.IsCreature is layer 4's output since ADR 0039), and a
// Darksteel Mutation'd Sol Ring really is a creature for Cursed
// Totem's purposes.
func SourcesCantActivate(label string, match CardPredicate, exemptMana bool) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			if exemptMana && q.Ability.Mana {
				return false
			}
			return matchActivationSource(q, match)
		},
	}
}

// OpponentsSourcesCantActivate is the same clause bound to every
// player EXCEPT the source's controller: Linvala, Keeper of Silence's
// "Activated abilities of creatures YOUR OPPONENTS CONTROL can't be
// activated".
//
// "Your opponents" is measured against the RESTRICTING permanent's
// controller (q.Source.Controller), never against the seat asking —
// which is why ActivationQuery carries both. A Linvala an opponent
// has stolen silences the original controller's creatures, and that
// is the printed behaviour.
//
// The seat compared is the object's CONTROLLER rather than the
// activator, because the clause says "creatures your opponents
// control" and CR 602.5a asks about the ability's source. On the
// battlefield the two are the same player; off it (a cycling ability,
// CR 108.4) there is no controller to be an opponent of, and such a
// card is not on the battlefield for "creatures ... control" to reach
// anyway.
func OpponentsSourcesCantActivate(label string, match CardPredicate, exemptMana bool) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			if exemptMana && q.Ability.Mana {
				return false
			}
			if q.FromZone != game.ZoneBattlefield {
				return false
			}
			if q.Card.Controller == q.Source.Controller {
				return false
			}
			return matchActivationSource(q, match)
		},
	}
}

// OpponentsSourcesCantActivateDuringYourTurn is Grand Abolisher's
// activation half: "During your turn, your opponents can't ...
// activate abilities of artifacts, creatures, or enchantments."
//
// OpponentsSourcesCantActivate with the turn added, the same way
// OpponentsCantCastDuringYourTurn (cast_restriction.go) adds it to
// OpponentsCantCast. Both halves are read off the SOURCE's
// controller: the "your" in "your opponents" and the "your" in
// "your turn", so a stolen Grand Abolisher locks the table out on
// its new controller's turn instead.
func OpponentsSourcesCantActivateDuringYourTurn(label string, match CardPredicate, exemptMana bool) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			if exemptMana && q.Ability.Mana {
				return false
			}
			if q.FromZone != game.ZoneBattlefield {
				return false
			}
			if q.Card.Controller == q.Source.Controller {
				return false
			}
			if !isActivePlayer(q.Game, q.Source.Controller) {
				return false
			}
			return matchActivationSource(q, match)
		},
	}
}

// ChosenNameCantActivate is Pithing Needle, Phyrexian Revoker and
// Sorcerous Spyglass: "Activated abilities of sources with the chosen
// name can't be activated unless they're mana abilities."
//
// SOURCES, not permanents — the clause reaches a card in any zone
// whose ability functions from there, which since #660 is a real case
// (a Ketria Triome's cycling from hand). So no zone test, unlike
// OpponentsSourcesCantActivate above, whose clause says "creatures
// your opponents control".
//
// The name is read LIVE off the restricting permanent, never captured
// at Register time: two Needles name two different cards, and the
// field is per instance.
//
// An unchosen name — the window between the permanent entering and
// its controller answering the prompt — matches nothing, which is the
// safe direction and the same one an empty NamedTribe takes.
//
// The comparison is game.CardNameMatches, which asks every FACE (CR
// 201.2b) case-insensitively. A card file must never compare
// q.Card.Name itself: Card.Name is the ACTIVE face's name, so a
// transformed permanent would stop matching the name it was named by.
func ChosenNameCantActivate(label string, exemptMana bool) game.ActivationRestriction {
	return game.ActivationRestriction{
		Label: label,
		Forbids: func(q game.ActivationQuery) bool {
			if exemptMana && q.Ability.Mana {
				return false
			}
			named := q.Source.ChosenName
			if named == "" {
				return false
			}
			return game.CardNameMatches(q.Card, named)
		},
	}
}

// ChooseCardNameAsEnters builds the `Spec.AsEnters` for a permanent
// whose text opens "As this permanent enters, choose a card name"
// (CR 614.12) — Pithing Needle, Phyrexian Revoker, Sorcerous
// Spyglass, Meddling Mage, Nevermore.
//
// `label` is the prompt header, normally the card's name.
//
// The prompt is queued rather than resolved: nothing happens until
// the controller answers, and until then the permanent's ChosenName
// is empty and every restriction that reads it applies to nothing.
// See game/choose_card_name.go for why this is an enter hook and not
// a CR 614 replacement, and for what that costs.
//
// It lives here rather than beside ChooseCreatureTypeAsEnters in
// tribal.go because a chosen NAME has nothing to do with tribes and
// everything to do with the restriction above it — the one clause in
// the catalog that reads the answer.
func ChooseCardNameAsEnters(label string) func(*game.Card, *Context) error {
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueCardNameChoiceForEffect(card.Controller, card.InstanceID, label)
		return nil
	}
}

// matchActivationSource runs a card predicate against the OBJECT
// whose ability is being activated.
//
// The predicate's `caster` argument is the player ACTIVATING, so a
// clause could say "creatures you control"; none does today, and the
// two seat-scoped shapes above answer that question themselves rather
// than leaving it to a predicate a card file could get backwards.
func matchActivationSource(q game.ActivationQuery, match CardPredicate) bool {
	if match == nil {
		return true
	}
	return match(q.Game, q.Controller, q.Card)
}
