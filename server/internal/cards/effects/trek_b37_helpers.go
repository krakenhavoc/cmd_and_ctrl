package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b37_helpers.go — shared pieces for Star Trek Commander batch 37
// (#4974). Names are prefixed trekB37 so they cannot collide with another
// batch's.

// --- Battlefront -------------------------------------------------------------

// trekB37OpponentsAttackedThisTurn is battlefront's count, "the number of
// opponents you attacked this turn" (Galvanizing Tenor, Own the Day).
// Battlefront is an ability word with no rules meaning of its own.
//
// It reads this turn's attack records the way melee reads this combat's
// (game.OpponentsAttackedThisCombat), over every combat phase of the turn:
//
//   - each opponent counts once, however many creatures attacked them
//     and in however many combats;
//   - an attack on a planeswalker or a battle is an attack on that
//     permanent and counts no player (CR 506.3);
//   - a creature that attacked and has since left combat or the
//     battlefield still counts, because the records are history;
//   - only the active player declares attackers (CR 508.1), so a player
//     who is not the active player attacked no one this turn.
//
// Caller must hold g.mu.
func trekB37OpponentsAttackedThisTurn(g *game.Game, you uuid.UUID) int {
	if g == nil || you == uuid.Nil || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return 0
	}
	if ap := g.Seats[g.Turn.ActiveSeat]; ap == nil || ap.ID != you {
		return 0
	}
	seen := map[uuid.UUID]bool{}
	count := func(defender uuid.UUID) {
		if defender == you || seen[defender] || g.PlayerByIDForEffect(defender) == nil {
			return
		}
		seen[defender] = true
	}
	for _, a := range g.TurnTally.Attacks {
		count(a.Defender)
	}
	// The declaration in progress: an attack trigger's targets are
	// chosen as its declaration's EventAttacks are still being recorded
	// one by one (CR 508.1m, 508.2), so every creature attacking right
	// now counts too. Only declared attackers carry an attack target
	// during the declaration; one put onto the battlefield attacking
	// arrives later in combat, and the end of combat clears them all.
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == you && c.AttackingTarget != uuid.Nil {
			count(c.AttackingTarget)
		}
	}
	return len(seen)
}

// --- Klingons and Warriors -----------------------------------------------------

// trekB37KlingonOrWarriorSpell is the cost half of "Klingon spells and
// Warrior spells you cast" (Gagh Slinger): SpellOfEitherCreatureType, so a
// Klingon Warrior spell is one spell and gets the discount once.
func trekB37KlingonOrWarriorSpell() CostPredicate {
	return SpellOfEitherCreatureType("Klingon", "Warrior")
}

// trekB37AnotherKlingonOrWarriorEntered is "whenever another Klingon or
// Warrior you control enters" (Gagh Slinger): a permanent other than the
// source with either type entered under the source's controller. The
// types are read off the entered permanent as it is on the battlefield
// (a Klingon Warrior token counts once, as one event).
func trekB37AnotherKlingonOrWarriorEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && (c.HasSubtype(source.TextSubtype("Klingon")) || c.HasSubtype(source.TextSubtype("Warrior")))
}

// trekB37KlingonsAndWarriorsYouControl is "Klingons and Warriors you
// control" (Gnarling Targ) as the set a resolving effect locks in (CR
// 611.2c).
func trekB37KlingonsAndWarriorsYouControl() CardPredicate {
	return And(YouControl(), AnySubtype("Klingon", "Warrior"))
}

// --- goad --------------------------------------------------------------------

// trekB37TribblesAreGoaded is the goad half of Koloth, Tribble Hunter's
// "All Tribbles get +1/+0 and are goaded." — a STATIC goad, written the
// way GoadAttached writes Shiny Impetus's: goad's two requirements (CR
// 701.15b) for Koloth's controller on every Tribble creature on the
// battlefield, whoever controls it, for as long as Koloth is there. It
// has no "until your next turn" end, because it is a continuous effect
// and not the act of goading (CR 701.15a). The type is a word of the
// source's text (CR 612.2).
func trekB37TribblesAreGoaded() game.StaticAbility {
	return game.StaticAbility{
		Label: "All Tribbles are goaded.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.HasSubtype(source.TextSubtype("Tribble"))
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			c.AttackRequirements = append(c.AttackRequirements, game.GoadRequirements(source, source.Controller)...)
		},
	}
}

// trekB37GoadedCreatureAttacked is the per-event half of "whenever one or
// more goaded creatures attack" (Kang, Suspicious Captain): a declared
// attacker, whoever controls it, that is goaded in any way — a resolved
// goad or a continuous one (Card.Goaded) — as it is declared. The row is
// OncePerBatch, so a declaration with several goaded attackers triggers
// it once.
func trekB37GoadedCreatureAttacked(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Goaded()
}
