package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_target_restrictions.go — the card-facing half of ADR 0106 §2
// (#1794): a creature that "can't attack its owner". The engine half —
// the predicate every declaration-time caller shares, the enumerator's
// per-attacker target list, the exemptions for CR 508.7b and 508.4c —
// is game/attack_target_restrictions.go. A card file only says which
// form it prints:
//
//	Static: []game.StaticAbility{AttacksEachCombat(), CantAttackItsOwnerOrItsOwnersPlaneswalkers()}, // Xantcha, Sleeper Agent
//	Static: []game.StaticAbility{AttacksEachCombat(), CantAttackItsOwner()},                         // Alexios, Deimos of Kosmos
//
// Both are the creature's own ability, so they go with its abilities:
// a creature that loses all abilities may attack its owner (CR 613.1f —
// the catalog stops being asked once CatalogAbilityKey answers empty).
// The owner is read live by the engine, so a Clone copying Xantcha may
// not attack the Clone's owner rather than Xantcha's.
//
// ADR 0107 §2 (#1879) adds the third form, "This creature can't attack
// unless defending player controls <a permanent>", as data the engine
// asks of each target's defending player (CR 508.5):
//
//	Static: []game.StaticAbility{CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island"))}, // Sea Serpent
//
// "This creature can't attack its owner" GRANTED by a resolved effect
// (Elrond of the White Council) is not this: it needs a ScopedEffect
// mod kind that writes the same restriction, which ADR 0106 leaves out
// of scope.

// CantAttackItsOwner is "~ can't attack its owner".
func CantAttackItsOwner() game.StaticAbility {
	return cantAttackOwnerStatic(false)
}

// CantAttackItsOwnerOrItsOwnersPlaneswalkers is "~ can't attack its
// owner or planeswalkers its owner controls". A battle the owner
// protects is still a legal target.
func CantAttackItsOwnerOrItsOwnersPlaneswalkers() game.StaticAbility {
	return cantAttackOwnerStatic(true)
}

func cantAttackOwnerStatic(planeswalkers bool) game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{NotOwner: true, NotOwnersPlaneswalkers: planeswalkers})
}

// CantAttackUnlessDefendingPlayerControls is "This creature can't attack
// unless defending player controls <a permanent>" (ADR 0107 §2, #1879,
// CR 508.1c). The queries are any-of: Godhunter Octopus's "an enchantment
// or an enchanted permanent" is two of them.
//
//	Static: []game.StaticAbility{CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island"))}, // Sea Serpent
//
// The engine works out the defending player of each target (CR 508.5: the
// player, a planeswalker's controller, a battle's protector), so in
// Commander the creature may attack the opponents who control an Island
// and nobody else. Like the owner clauses it is the creature's own ability
// and goes with its abilities (CR 613.1f).
func CantAttackUnlessDefendingPlayerControls(queries ...game.PermanentQuery) game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{DefenderMustControl: queries})
}

// CantAttackUnlessDefendingPlayerIsPoisoned is "This creature can't
// attack unless defending player is poisoned" (Chained Throatseeker,
// #1879): one or more poison counters (CR 122.1f), asked of each
// target's defending player.
func CantAttackUnlessDefendingPlayerIsPoisoned() game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{DefenderMustBePoisoned: true})
}

// CantAttackUnlessDefendingPlayerIsTheMonarch is "This creature can't
// attack unless defending player is the monarch" (Crown-Hunter
// Hireling, #1879, CR 725.1). With no monarch it can't attack at all.
func CantAttackUnlessDefendingPlayerIsTheMonarch() game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{DefenderMustBeMonarch: true})
}

// CantAttackUnlessDefendingPlayerHasCardsInGraveyard is "This creature
// can't attack unless defending player has N or more cards in their
// graveyard" (Vantress Gargoyle's seven, #1879).
func CantAttackUnlessDefendingPlayerHasCardsInGraveyard(n int) game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{DefenderGraveyardAtLeast: n})
}

// CantAttackUnlessYouControlMoreThanDefendingPlayer is "This creature
// can't attack unless you control more <permanents> than defending
// player" (#1879): Goblin Goon's and Mogg Toady's creatures, Monstrous
// Hound's lands. "You" is the creature's controller; the counts are of
// effective characteristics, read at declaration.
//
//	Static: []game.StaticAbility{CantAttackUnlessYouControlMoreThanDefendingPlayer(QueryType("creature"))}, // Goblin Goon
func CantAttackUnlessYouControlMoreThanDefendingPlayer(queries ...game.PermanentQuery) game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{ControllerMustControlMore: queries})
}

// CantAttackAPlayerItAlreadyAttackedThisTurn is "This creature can't
// attack a player it has already attacked this turn" (Bloodthirster,
// #2171, CR 508.1c). The engine reads the turn's attack record for this
// object, so a creature that left and returned is a new object (CR
// 400.7) with a clean slate.
func CantAttackAPlayerItAlreadyAttackedThisTurn() game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{NotAlreadyAttackedThisTurn: true})
}

// QuerySubtype is "a permanent with this subtype": an Island, a Mountain.
func QuerySubtype(subtype string) game.PermanentQuery {
	return game.PermanentQuery{Subtypes: []string{subtype}}
}

// CreaturesCantAttackPlayersWhoDidNothingLastTurn is Arboria's
// "Creatures can't attack a player unless that player cast a spell or
// put a nontoken permanent onto the battlefield during their last turn"
// (#1997, ADR 0109 amendment 2026-10-23, CR 508.1c): every creature, any
// controller's, gets the restriction, with this permanent as its source,
// so a creature that loses all abilities keeps it (it is not the
// creature's own ability). Players only: planeswalkers and battles may
// still be attacked. The engine reads the player's last turn live at
// declaration (game.Game.CastOrPutDuringLastTurnForEffect).
func CreaturesCantAttackPlayersWhoDidNothingLastTurn() game.StaticAbility {
	r := game.AttackTargetRestriction{PlayerMustHaveActedLastTurn: true}
	return game.StaticAbility{
		Label: "Creatures can't attack a player unless that player cast a spell or put a nontoken permanent onto the battlefield during their last turn.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := r
			if source != nil {
				r.Source, r.SourceName = source.InstanceID, source.Name
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, r)
		},
	}
}

// CreaturesMatchingCantAttackYou is "<creatures matching the queries>
// can't attack you" (CR 508.1c, #4601): the PLAYER alone, not their
// planeswalkers or battles, so a matching creature may still attack
// those. Storm, Windrider's "Creatures with flying can't attack you":
//
//	Static: []game.StaticAbility{CreaturesMatchingCantAttackYou(
//	    "Creatures with flying can't attack you.", game.PermanentQuery{Types: []string{"creature"}, Keyword: "flying"})},
//
// "You" is the controller of the permanent that prints it, read when the
// layer pass runs, so it follows a change of control. Every creature, any
// controller's, gets the restriction with this permanent as its source
// (it is not the creature's own ability, so it survives a creature losing
// all abilities), and the engine decides at declaration whether the
// attacker matches the queries (AttackTargetRestriction.OnlyAttackersMatching),
// on its effective characteristics at that moment: a creature that
// gains or loses flying before attackers are declared is judged as it
// then stands. Any-of across queries.
func CreaturesMatchingCantAttackYou(label string, attackers ...game.PermanentQuery) game.StaticAbility {
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
				NotPlayerOnly:         true,
				OnlyAttackersMatching: append([]game.PermanentQuery(nil), attackers...),
			})
		},
	}
}

// selfAttackTargetRestriction writes r onto the creature whose ability it
// is, naming that creature as the source.
func selfAttackTargetRestriction(r game.AttackTargetRestriction) game.StaticAbility {
	return game.StaticAbility{
		Label: attackTargetLabel(r),
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && selfOnly(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			r := r
			if source != nil {
				r.Source, r.SourceName = source.InstanceID, source.Name
				// "Unless defending player controls an Island" names
				// its land type in the source's text (CR 612.2).
				r.DefenderMustControl = game.PermanentQueriesThroughTextOf(r.DefenderMustControl, source)
				r.ControllerMustControlMore = game.PermanentQueriesThroughTextOf(r.ControllerMustControlMore, source)
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, r)
		},
	}
}
