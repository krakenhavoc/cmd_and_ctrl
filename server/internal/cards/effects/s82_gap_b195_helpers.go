package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b195_helpers.go — shared pieces of S82 GAP batch 195 (#5948).

// GapB195InfectionCounter is the infection counter Genestealer Patriarch
// puts on creatures. Counter kinds are free-form strings (game/counter_types.go).
const GapB195InfectionCounter = "infection"

// GapB195CopyTheNextInstantOrSorcery is Tzaangor Shaman's effect: "copy the
// next instant or sorcery spell you cast this turn when you cast it. You may
// choose new targets for the copy." It is Doublecast's delayed trigger.
func GapB195CopyTheNextInstantOrSorcery(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return WhenYouNextCast(label,
			game.CastFilter{Types: []string{"Instant", "Sorcery"}}, CopyTheSpellBody).Apply(NewContext(g, item))
	}
}

// GapB195OtherCreaturesWithCounters is the scope of "other creatures you
// control with counters on them" (Winged Hive Tyrant).
func GapB195OtherCreaturesWithCounters(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
		target.IsCreature() && B64HasAnyCounter()(g, uuid.Nil, *target)
}

// GapB195AnAttackerAttackedYou is "whenever a creature attacks you": an attack
// declaration whose defending player is the source's controller (an attack on
// a planeswalker or battle of theirs is not an attack on them).
func GapB195AnAttackerAttackedYou(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil && ev.Target == source.Controller
}

// GapB195AnAttackerAttackedOneOfYourOpponents is "whenever a creature attacks
// one of your opponents": an attack declaration at a player other than the
// source's controller.
func GapB195AnAttackerAttackedOneOfYourOpponents(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventAttack && ev.CardID != uuid.Nil && ev.Target != source.Controller &&
		g.PlayerByIDForEffect(ev.Target) != nil
}

// GapB195AttackerGetsPumpUntilEOT is the effect of a Genestealer Locus
// trigger: the attacking creature gets +P/+T until end of turn.
func GapB195AttackerGetsPumpUntilEOT(power, toughness int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return PumpTriggeringCreatureUntilEOT(NewContext(g, item), power, toughness, label)
	}
}

// GapB195InfectTarget is Genestealer Patriarch's attack effect: an infection
// counter on the target creature the defending player controls.
func GapB195InfectTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	return AddCounter{Target: id, Kind: GapB195InfectionCounter, N: 1}.Apply(ctx)
}

// GapB195AnInfectedCreatureDied is "whenever a creature with an infection
// counter on it dies": any creature, anyone's, that last had one (CR 603.10a).
func GapB195AnInfectedCreatureDied(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if _, ok := DiedCreature(ev, g); !ok {
		return false
	}
	return g.LastKnownCountersForEffect(ev.CardID)[GapB195InfectionCounter] > 0
}

// GapB195CopyTheInfectedCreature is "you create a token that's a copy of that
// creature, except it's a Tyranid in addition to its other types." The copy
// takes the creature's copiable values as it last existed (CR 707.2, 608.2h).
func GapB195CopyTheInfectedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	dead := ctx.Trigger().Event.CardID
	if dead == uuid.Nil {
		return nil
	}
	tyranid := ctx.TextSubtype("Tyranid")
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       dead,
		N:          1,
		Except: func(t *game.Card) {
			t.TypeLine = AddedSubtypeTypeLine(t.TypeLine, tyranid)
		},
	}.Apply(ctx)
}

// GapB195CounterTheTargetItem is Sister of Silence's effect: counter the
// target instant spell, sorcery spell, activated ability or triggered ability.
func GapB195CounterTheTargetItem(g *game.Game, item *game.StackItem) error {
	return CounterTheChosenSpell(g, item)
}

// GapB195SisterTarget is "target instant spell, sorcery spell, activated
// ability, or triggered ability".
func GapB195SisterTarget() *game.TargetSpec {
	return TargetSpellOrAbility("target instant spell, sorcery spell, activated ability, or triggered ability",
		AnyStackItem(ASpellItem(Or(Instant(), Sorcery())), AnActivatedAbilityItem(), BnsB06TriggeredAbilityItem()))
}

// GapB195MasterOfMachines is Belisarius Cawl's second ability: look at the top
// X cards of your library, you may reveal an artifact card from among them and
// put it into your hand, and put the rest on the bottom in a random order. X is
// the number of creatures tapped for the cost.
func GapB195MasterOfMachines(g *game.Game, item *game.StackItem) error {
	x := NewContext(g, item).X()
	if x <= 0 {
		return nil
	}
	return LookAtTopThenMayTakeToHand(x, Artifact(), 1,
		"Belisarius Cawl — you may reveal an artifact card and put it into your hand")(g, item)
}

// GapB195LibrarianFreeCast is "you may cast a spell with mana value X or less
// from your hand without paying its mana cost" for Epistolary Librarian.
var GapB195LibrarianFreeCast = RegisterFreeHandCastOffer("choose-cards/epistolary-librarian-cast", "Epistolary Librarian")

// GapB195VeilOfTime is the body of Epistolary Librarian's attack trigger: X is
// the number of attacking creatures as the trigger resolves.
func GapB195VeilOfTime(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	OfferFreeHandCast(ctx, ctx.Controller(), GapB195LibrarianFreeCast,
		"a spell with mana value X or less", SpellWithManaValueAtMost(len(AttackingCreatures(g))))
	return nil
}
